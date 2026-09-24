package peer

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"p2pChat/identity"
	"p2pChat/models"
	"time"
)

func ConnectToPeer(address string, newNode *models.Node, expectedPeerID string) (net.Conn, error) {
	cert, err := identity.GenerateSelfSignedCert(newNode.PrivateKey, newNode.Name)
	if err != nil {
		return nil, fmt.Errorf("cert generation error: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates:          []tls.Certificate{cert},
		InsecureSkipVerify:    true,
		VerifyPeerCertificate: identity.VerifyPeerID(expectedPeerID, nil),
	}

	conn, err := tls.Dial("tcp", address, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("could not connect to peer: %w", err)
	}
	return conn, nil
}

// ReadFromPeer parses incoming messages and hands them to the GUI via
// newNode.OnMessage, instead of printing. No stdout coupling.
func ReadFromPeer(addr string, conn net.Conn, newNode *models.Node) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var msg models.Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue // silently skip malformed lines; GUI has no console to report to
		}
		if newNode.OnMessage != nil {
			newNode.OnMessage(addr, msg)
		}
	}
	newNode.RemovePeer(addr) // triggers OnPeerDisconnected internally

	if expectedID, known := newNode.ExpectedIDFor(addr); known {
		go reconnectWithBackoff(addr, expectedID, newNode)
	}
}

func reconnectWithBackoff(addr string, expectedID string, newNode *models.Node) {
	delay := 2 * time.Second
	const maxDelay = 30 * time.Second
	const maxAttempts = 6

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if _, stillKnown := newNode.ExpectedIDFor(addr); !stillKnown {
			return
		}

		time.Sleep(delay)

		conn, err := ConnectToPeer(addr, newNode, expectedID)
		if err != nil {
			if delay < maxDelay {
				delay *= 2
			}
			continue
		}

		if newNode.AddPeer(addr, conn) {
			go ReadFromPeer(addr, conn, newNode)
			return
		}
		return
	}

	newNode.ForgetPeer(addr)
}
