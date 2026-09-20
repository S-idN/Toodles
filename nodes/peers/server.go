package peer

import (
	"crypto/tls"
	"fmt"
	"p2pChat/identity"
	"p2pChat/models"
	"strconv"
)

func StartServer(port int, newNode *models.Node) {
	cert, err := identity.GenerateSelfSignedCert(newNode.PrivateKey, newNode.Name)
	if err != nil {
		fmt.Println("Cert generation error:", err)
		return
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		ClientAuth:         tls.RequireAnyClientCert,
		InsecureSkipVerify: true,
		VerifyPeerCertificate: identity.VerifyPeerID("", func(observedID string) {
			fmt.Println("Incoming connection from peer ID:", observedID)
		}),
	}

	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(port), tlsConfig)
	if err != nil {
		fmt.Println("Failed to start listener:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Listening on Port:", port)

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}

		addr := conn.RemoteAddr().String()
		if newNode.AddPeer(addr, conn) {
			go ReadFromPeer(addr, conn, newNode)
		}
	}
}
