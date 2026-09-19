package peer

import (
	"fmt"
	"net"
	"p2pChat/models"
	"strconv"
)

func StartServer(port int, newNode *models.Node) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
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
