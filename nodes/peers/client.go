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

func ReadFromPeer(addr string, conn net.Conn, newNode *models.Node) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var msg models.Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			fmt.Println("[system] Received unreadable message from", addr)
			continue
		}
		fmt.Print("[", msg.Timestamp.Format("15:04:05"), "]: ", msg.From, ": ", msg.Body, "\n")
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("[system] Error reading from peer", addr, ":", err)
	}
	fmt.Println("[system] Peer disconnected:", addr)
	newNode.RemovePeer(addr)

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
			return // peer was forgotten, stop retrying
		}

		fmt.Printf("[system] Reconnect attempt %d to %s...\n", attempt, addr)
		time.Sleep(delay)

		conn, err := ConnectToPeer(addr, newNode, expectedID)
		if err != nil {
			fmt.Println("[system] Reconnect failed:", err)
			if delay < maxDelay {
				delay *= 2
			}
			continue
		}

		if newNode.AddPeer(addr, conn) {
			fmt.Println("[system] Reconnected to", addr)
			go ReadFromPeer(addr, conn, newNode)
			return
		}
		return // AddPeer returned false (dupe) - someone else already reconnected
	}

	fmt.Println("[system] Giving up on reconnecting to", addr, "after", maxAttempts, "attempts")
	newNode.ForgetPeer(addr)
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
