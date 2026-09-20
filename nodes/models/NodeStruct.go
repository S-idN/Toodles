package models

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"sync"

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
		fmt.Println("Already connected to", addr, "- ignoring duplicate")
		conn.Close()
		return false
	}

	n.Peers[addr] = conn
	fmt.Println("Peer added:", addr, "| total peers:", len(n.Peers))
	return true
}

func (n *Node) RemovePeer(addr string) {
	n.peersMu.Lock()
	defer n.peersMu.Unlock()

	if conn, exists := n.Peers[addr]; exists {
		conn.Close()
		delete(n.Peers, addr)
		fmt.Println("Peer removed:", addr, "| total peers:", len(n.Peers))
	}
}

// Broadcast sends message to every currently connected peer.
func (n *Node) Broadcast(message string) {
	n.peersMu.Lock()
	defer n.peersMu.Unlock()

	for addr, conn := range n.Peers {
		_, err := conn.Write([]byte(message + "\n"))
		if err != nil {
			fmt.Println("Failed to send to", addr, ":", err)
		}
	}
}
