package main

import (
	"bufio"
	"fmt"
	"os"
	peer "p2pChat/peers"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: main.exe <listen-port> <peer-port>")
		return
	}

	listenPort := os.Args[1]
	peerPort := os.Args[2]

	go peer.StartServer(listenPort)

	fmt.Println("Press enter to connect (wait for both peers first)")
	discardReader := bufio.NewReader(os.Stdin)
	_, _ = discardReader.ReadString('\n')

	conn, err := peer.ConnectToPeer("localhost:" + peerPort)
	if err != nil {
		fmt.Println("Connect error:", err)
		return
	}

	peer.WriteToPeer(conn)
	// peer.SendMessage(conn, "hello from peer on port "+listenPort)

	select {}
}
