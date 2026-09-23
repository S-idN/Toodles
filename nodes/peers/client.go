package peer

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"p2pChat/P2pMDNS"
	"p2pChat/identity"
	"p2pChat/models"
	"strconv"
	"strings"
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

func StartPeerOps(newNode *models.Node) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Type a message to broadcast to connected peers, /scan to discover peers, or /disconnect to drop a peer.")

	for scanner.Scan() {
		line := scanner.Text()

		if line == "/scan" {
			fmt.Println("Scanning local network...")
			entries, err := P2pMDNS.LookupMDNS(newNode)

			if err != nil {
				fmt.Println("Scan error or no peers found:", err)
				continue
			}

			addr, expectedPeerID := promptEntrySelection(entries, scanner)
			fmt.Println("Connecting to", addr, "expecting peer ID", expectedPeerID)

			conn, err := peer.ConnectToPeer(addr, newNode, expectedPeerID)
			if err != nil {
				fmt.Println("Connect error:", err)
				continue
			}

			if newNode.AddPeer(addr, conn) {
				newNode.RememberPeer(addr, expectedPeerID)
				go peer.ReadFromPeer(addr, conn, newNode)
			}
			continue
		}

		if line == "/disconnect" {
			addrs := newNode.PeerAddrs()
			if len(addrs) == 0 {
				fmt.Println("[system] No connected peers.")
				continue
			}

			fmt.Println("Connected peers:")
			for i, a := range addrs {
				fmt.Printf("  [%d] %s\n", i, a)
			}
			fmt.Print("Select a peer to disconnect: ")

			if !scanner.Scan() {
				continue
			}
			choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil || choice < 0 || choice >= len(addrs) {
				fmt.Println("[system] Invalid selection.")
				continue
			}

			addr := addrs[choice]
			newNode.ForgetPeer(addr) // do this FIRST so the reconnect loop doesn't kick in
			newNode.RemovePeer(addr) // closes the conn, triggering ReadFromPeer's exit + RemovePeer (idempotent)
			fmt.Println("[system] Disconnected from", addr)
			continue
		}

		newNode.Broadcast(line)
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
