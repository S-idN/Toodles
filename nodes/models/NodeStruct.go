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

type DiscoveredPeer struct {
	Addr string
	Name string
	ID   string
}

type Node struct {
	Name       string
	NodeId     string
	ListenPort int
	Server     *mdns.Server
	PrivateKey *ecdsa.PrivateKey

	peersMu sync.Mutex
	Peers   map[string]net.Conn

	knownMu    sync.Mutex
	KnownPeers map[string]string

	discoveredMu sync.Mutex
	Discovered   map[string]DiscoveredPeer // keyed by addr

	pendingMu sync.Mutex
	Pending   map[string]net.Conn // incoming connections awaiting GUI accept/reject, keyed by addr

	// GUI hooks - set these before starting the node. Called from
	// background goroutines, so the GUI side must marshal onto its own
	// main thread (e.g. Fyne's binding system) before touching widgets.
	OnPeerDiscovered    func(DiscoveredPeer)
	OnConnectionRequest func(addr string, observedID string)
	OnPeerConnected     func(addr string)
	OnPeerDisconnected  func(addr string)
	OnMessage           func(addr string, msg Message)
}

func (n *Node) AddDiscovered(p DiscoveredPeer) {
	n.discoveredMu.Lock()
	if n.Discovered == nil {
		n.Discovered = make(map[string]DiscoveredPeer)
	}
	_, existed := n.Discovered[p.Addr]
	n.Discovered[p.Addr] = p
	n.discoveredMu.Unlock()

	if !existed && n.OnPeerDiscovered != nil {
		n.OnPeerDiscovered(p)
	}
}

func (n *Node) AddPeer(addr string, conn net.Conn) bool {
	n.peersMu.Lock()
	if n.Peers == nil {
		n.Peers = make(map[string]net.Conn)
	}
	if _, exists := n.Peers[addr]; exists {
		n.peersMu.Unlock()
		conn.Close()
		return false
	}
	n.Peers[addr] = conn
	n.peersMu.Unlock()

	if n.OnPeerConnected != nil {
		n.OnPeerConnected(addr)
	}
	return true
}

func (n *Node) RemovePeer(addr string) {
	n.peersMu.Lock()
	conn, exists := n.Peers[addr]
	if exists {
		conn.Close()
		delete(n.Peers, addr)
	}
	n.peersMu.Unlock()

	if exists && n.OnPeerDisconnected != nil {
		n.OnPeerDisconnected(addr)
	}
}

func (n *Node) PeerAddrs() []string {
	n.peersMu.Lock()
	defer n.peersMu.Unlock()
	addrs := make([]string, 0, len(n.Peers))
	for a := range n.Peers {
		addrs = append(addrs, a)
	}
	return addrs
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
	delete(n.KnownPeers, addr)
	n.knownMu.Unlock()
}

// AddPending stashes an incoming connection until the GUI accepts/rejects it.
func (n *Node) AddPending(addr string, conn net.Conn, observedID string) {
	n.pendingMu.Lock()
	if n.Pending == nil {
		n.Pending = make(map[string]net.Conn)
	}
	n.Pending[addr] = conn
	n.pendingMu.Unlock()

	if n.OnConnectionRequest != nil {
		n.OnConnectionRequest(addr, observedID)
	}
}

// AcceptPending moves a pending connection into the live Peers map.
func (n *Node) AcceptPending(addr string) (net.Conn, bool) {
	n.pendingMu.Lock()
	conn, ok := n.Pending[addr]
	if ok {
		delete(n.Pending, addr)
	}
	n.pendingMu.Unlock()
	if !ok {
		return nil, false
	}
	return conn, n.AddPeer(addr, conn)
}

func (n *Node) RejectPending(addr string) {
	n.pendingMu.Lock()
	conn, ok := n.Pending[addr]
	if ok {
		delete(n.Pending, addr)
	}
	n.pendingMu.Unlock()
	if ok {
		conn.Close()
	}
}

func (n *Node) SendTo(addr string, body string) error {
	n.peersMu.Lock()
	conn, ok := n.Peers[addr]
	n.peersMu.Unlock()
	if !ok {
		return fmt.Errorf("no connection to %s", addr)
	}

	msg := Message{ID: uuid.New().String(), From: n.Name, NodeID: n.NodeId, Timestamp: time.Now(), Body: body}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = conn.Write(data)
	return err
}
