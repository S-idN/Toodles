package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"p2pChat/config"
	"p2pChat/models"
)

func main() {
	isLogged := false

	if len(os.Args) < 2 {
		fmt.Println("Usage: main.exe <listen-port>")
		return
	}

	//Port select
	listenPortStr := os.Args[1]
	listenPort, err := strconv.Atoi(listenPortStr)
	if err != nil {
		fmt.Println("Invalid listening port:", err)
		return
	}

	if isLogged {
		fmt.Println("Random bullshit go")
	}
	//Create node if new node
	newNode := models.Node{}
	config.SetupNewNode(&newNode, listenPort)
	defer config.NodeShutdown(&newNode)

	fmt.Println("Node is broadcasting via mDNS.")
	fmt.Println("Press enter to start scanning the local network for peers...")

	discardReader := bufio.NewReader(os.Stdin)
	_, _ = discardReader.ReadString('\n')

	config.StartPeerOps(&newNode)
}
