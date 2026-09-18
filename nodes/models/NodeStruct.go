package models

import (
	"net"

	"github.com/hashicorp/mdns"
)

type Node struct {
	Name       string
	Conn       net.Conn
	PeerConn   net.Conn
	NodeId     string
	ListenPort int
	Server     *mdns.Server
	PeerList   []string
}
