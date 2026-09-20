package peer

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"p2pChat/identity"
	"p2pChat/models"
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

func ReadFromPeer(addr string, conn net.Conn, newNode *models.Node) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var msg models.Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			fmt.Println("[system] Received unreadable message from", addr)
			continue
		}
		fmt.Printf("[%s] %s: %s\n", msg.Timestamp.Format("15:04:05"), msg.From, msg.Body)
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("[system] Error reading from peer", addr, ":", err)
	}
	fmt.Println("[system] Peer disconnected:", addr)
	newNode.RemovePeer(addr)
}

func WriteLoop(newNode *models.Node) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		message := scanner.Text()
		newNode.Broadcast(message)
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading stdin:", err)
	}
}

func SendMessage(conn net.Conn, message string) error {
	_, err := conn.Write([]byte(message + "\n"))
	return err
}
