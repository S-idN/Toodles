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

	knownMu    sync.Mutex
	KnownPeers map[string]string
}

func (n *Node) RememberPeer(addr, nodeID string) {
	n.knownMu.Lock()
	defer n.knownMu.Unlock()
	if n.KnownPeers == nil {
		n.KnownPeers = make(map[string]string)
	}
	n.KnownPeers[addr] = nodeID
}

func (n *Node) ExpectedIDFor(addr string) (string, bool) {
	n.knownMu.Lock()
	defer n.knownMu.Unlock()
	id, ok := n.KnownPeers[addr]
	return id, ok
}

func (n *Node) ForgetPeer(addr string) {
	n.knownMu.Lock()
	defer n.knownMu.Unlock()
	delete(n.KnownPeers, addr)
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

func (n *Node) PeerAddrs() []string {
	n.peersMu.Lock()
	defer n.peersMu.Unlock()

	addrs := make([]string, 0, len(n.Peers))
	for addr := range n.Peers {
		addrs = append(addrs, addr)
	}
	return addrs
}
