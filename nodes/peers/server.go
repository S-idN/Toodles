package peer

import (
	"crypto/tls"
	"p2pChat/identity"
	"p2pChat/models"
	"strconv"
)

func StartServer(port int, newNode *models.Node) {
	cert, err := identity.GenerateSelfSignedCert(newNode.PrivateKey, newNode.Name)
	if err != nil {
		return
	}
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		ClientAuth:         tls.RequireAnyClientCert,
		InsecureSkipVerify: true,
		VerifyPeerCertificate: identity.VerifyPeerID("", func(observedID string) {
			// handled per-connection below, id stashed via closure
		}),
	}
	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(port), tlsConfig)
	if err != nil {
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		addr := conn.RemoteAddr().String()

		tlsConn := conn.(*tls.Conn)
		if err := tlsConn.Handshake(); err != nil {
			conn.Close()
			continue
		}
		observedID := ""
		if state := tlsConn.ConnectionState(); len(state.PeerCertificates) > 0 {
			observedID, _ = identity.FingerprintFromCert(state.PeerCertificates[0])
		}
		newNode.AddPending(addr, conn, observedID)
	}
}
