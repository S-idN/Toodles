package config

import (
	"bufio"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"p2pChat/P2pMDNS"
	"p2pChat/models"
	peer "p2pChat/peers"
	"strconv"
	"strings"
	"time"
	"uuid"
)

func GenerateBaseId() string {
	id := uuid.New()
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

func SetupNewNode(newNode *models.Node, listenPort int) (retNode *models.Node) {

	inputReader := bufio.NewReader(os.Stdin)
	newNode.NodeId = GenerateFinalId()
	fmt.Println("Your Node ID:", newNode.NodeId)
	fmt.Println("Enter a node name:")
	inputName, err := inputReader.ReadString('\n')
	fmt.Print("Input name:", inputName)
	if err != nil || inputName == "nil" {
		fmt.Println("Invalid name", err)
		os.Exit(1)
	}
	//NODE : Name
	name := inputName[:len(inputName)-1]
	newNode.Name = name
	fmt.Println("Name is", newNode.Name)

	//NODE : Listen Port
	newNode.ListenPort = listenPort

	//NODE : Conn (listens for incoming connections in the background)
	go peer.StartServer(listenPort, newNode)

	newNode.Server = P2pMDNS.SetupMDNS(listenPort)

	StartPeerOps(newNode)

	fmt.Println("STRUCT IS THIS:", newNode)

	return newNode
}

// promptEntrySelection prints "name|addr:port" entries and returns the
// "addr:port" portion of the one the user picks.
func promptEntrySelection(entries []string) string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Discovered peers:")
		for i, e := range entries {
			parts := strings.SplitN(e, "|", 2)
			label := e
			if len(parts) == 2 {
				label = fmt.Sprintf("%s (%s)", parts[0], parts[1])
			}
			fmt.Printf("  [%d] %s\n", i, label)
		}
		fmt.Print("Select a peer by number: ")

		raw, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Read error, try again:", err)
			continue
		}

		choice, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || choice < 0 || choice >= len(entries) {
			fmt.Println("Invalid selection, try again.")
			continue
		}

		parts := strings.SplitN(entries[choice], "|", 2)
		if len(parts) == 2 {
			return parts[1] // addr:port
		}
		return entries[choice]
	}
}

func StartPeerOps(newNode *models.Node) {
	rounds := 0

	for {
		fmt.Println("Scanning local network...")
		entries, err := P2pMDNS.LookupMDNS(newNode)

		if err != nil {
			fmt.Println("Scan error or no peers found:", err)
			time.Sleep(10 * time.Second)
			continue
		}

		addr := promptEntrySelection(entries)
		fmt.Println("Connecting to", addr)

		peerConn, err := peer.ConnectToPeer(addr)
		newNode.PeerConn = peerConn

		if err != nil {
			fmt.Println("Connect error:", err)
			time.Sleep(10 * time.Second)
			continue
		}

		peer.WriteToPeer(newNode.PeerConn)
		time.Sleep(10 * time.Second)

		rounds += 1

		if rounds >= 3 {
			break
		}
	}
}

func NodeShutdown(newNode *models.Node) {
	defer newNode.Server.Shutdown()
}
