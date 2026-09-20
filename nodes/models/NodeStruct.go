package models

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/mdns"
)

type Node struct {
	Name       string
	NodeId     string
	ListenPort int
	Server     *mdns.Server
	PeerList   []string
	PrivateKey *ecdsa.PrivateKey

	peersMu sync.Mutex
	Peers   map[string]net.Conn
}

func (n *Node) AddPeer(addr string, conn net.Conn) bool {
	n.peersMu.Lock()
	defer n.peersMu.Unlock()

	if n.Peers == nil {
		n.Peers = make(map[string]net.Conn)
	}

	if _, exists := n.Peers[addr]; exists {
		fmt.Println("[system] Already connected to", addr, "- ignoring duplicate")
		conn.Close()
		return false
	}

	n.Peers[addr] = conn
	fmt.Println("[system] Peer added:", addr, "| total peers:", len(n.Peers))
	return true
}

func (n *Node) RemovePeer(addr string) {
	n.peersMu.Lock()
	defer n.peersMu.Unlock()

	if conn, exists := n.Peers[addr]; exists {
		conn.Close()
		delete(n.Peers, addr)
		fmt.Println("[system] Peer removed:", addr, "| total peers:", len(n.Peers))
	}
}

// Broadcast wraps body in a structured Message (sender name, node ID,
// timestamp, message ID) and sends it as newline-delimited JSON to every
// connected peer.
func (n *Node) Broadcast(body string) {
	msg := Message{
		ID:        uuid.New().String(),
		From:      n.Name,
		NodeID:    n.NodeId,
		Timestamp: time.Now(),
		Body:      body,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("[system] Failed to encode message:", err)
		return
	}
	data = append(data, '\n')

	n.peersMu.Lock()
	defer n.peersMu.Unlock()

	for addr, conn := range n.Peers {
		if _, err := conn.Write(data); err != nil {
			fmt.Println("[system] Failed to send to", addr, ":", err)
		}
	}
}
