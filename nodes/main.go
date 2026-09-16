package main

import (
	"bufio"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"
	"uuid"

	"p2pChat/P2pMDNS"
	peer "p2pChat/peers"
)

func GenerateBaseId() string {
	id := uuid.New() // Updated to match standard google/uuid package usage
	hashedId := sha512.Sum512_256(id[:])
	nodeID := hex.EncodeToString(hashedId[:])

	return nodeID
}

func GenerateFinalId() string {
	bootstrapNodes := []string{"ip_here"}
	hashId_portion := GenerateBaseId()

	for i := 0; i < len(bootstrapNodes); i++ {
		hashId_portion += peer.ReturnHashIdPortion()
	}

	return hashId_portion
}

func main() {
	isLogged := false

	if !isLogged {
		node_id := GenerateFinalId()
		fmt.Println("Your Node ID:", node_id)
	}

	// We only need 1 argument now because peer-port is discovered automatically!
	if len(os.Args) < 2 {
		fmt.Println("Usage: main.exe <listen-port>")
		return
	}

	listenPortStr := os.Args[1]
	listenPort, err := strconv.Atoi(listenPortStr)
	if err != nil {
		fmt.Println("Invalid listening port:", err)
		return
	}

	go peer.StartServer(listenPortStr)

	mdnsServer, err := P2pMDNS.SetupMDNS(listenPort)
	if err != nil {
		fmt.Println("Error starting mDNS broadcaster:", err)
		return
	}
	defer mdnsServer.Shutdown()

	fmt.Println("Node is broadcasting via mDNS.")
	fmt.Println("Press enter to start scanning the local network for peers...")

	discardReader := bufio.NewReader(os.Stdin)
	_, _ = discardReader.ReadString('\n')

	for {
		fmt.Println("Scanning local network...")
		P2pMDNS.LookupMDNS()

		time.Sleep(10 * time.Second)
	}
}
