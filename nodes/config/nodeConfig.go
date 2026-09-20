package config

import (
	"bufio"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"p2pChat/P2pMDNS"
	"p2pChat/identity"
	"p2pChat/models"
	peer "p2pChat/peers"
	"strconv"
	"strings"
	"uuid"
)

func GenerateBaseId() string {
	id := uuid.New()
	hashedId := sha512.Sum512_256(id[:])
	nodeID := hex.EncodeToString(hashedId[:])

	return nodeID
}

func GenerateFinalId() string {
	// bootstrapNodes := []string{"ip_here"}
	hashId_portion := GenerateBaseId()

	// for i := 0; i < len(bootstrapNodes); i++ {
	// 	hashId_portion += peer.ReturnHashIdPortion()
	// }

	return hashId_portion
}
func SetupNewNode(newNode *models.Node, listenPort int) (retNode *models.Node) {

	inputReader := bufio.NewReader(os.Stdin)

	key, err := identity.LoadOrCreateKey(".")
	if err != nil {
		fmt.Println("Identity error:", err)
		os.Exit(1)
	}
	newNode.PrivateKey = key
	newNode.NodeId = identity.NodeIDFromKey(key)

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

	newNode.Server = P2pMDNS.SetupMDNS(listenPort, newNode)

	StartPeerOps(newNode)

	fmt.Println("STRUCT IS THIS:", newNode)

	return newNode
}

func promptEntrySelection(entries []string, scanner *bufio.Scanner) (addr string, nodeID string) {
	for {
		fmt.Println("Discovered peers:")
		for i, e := range entries {
			parts := strings.SplitN(e, "|", 3)
			label := e
			if len(parts) == 3 {
				label = fmt.Sprintf("%s (%s)", parts[0], parts[1])
			}
			fmt.Printf("  [%d] %s\n", i, label)
		}
		fmt.Print("Select a peer by number: ")

		if !scanner.Scan() {
			fmt.Println("Read error or input closed, try again.")
			continue
		}
		raw := scanner.Text()

		choice, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || choice < 0 || choice >= len(entries) {
			fmt.Println("Invalid selection, try again.")
			continue
		}

		parts := strings.SplitN(entries[choice], "|", 3)
		if len(parts) == 3 {
			return parts[1], parts[2] // addr:port, nodeid
		}
		return entries[choice], ""
	}
}

func StartPeerOps(newNode *models.Node) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Type a message to broadcast to connected peers, or /scan to discover peers.")

	for scanner.Scan() {
		line := scanner.Text()

		if line == "/scan" {
			fmt.Println("Scanning local network...")
			entries, err := P2pMDNS.LookupMDNS(newNode)

			if err != nil {
				fmt.Println("Scan error or no peers found:", err)
				continue
			}

			addr, expectedPeerID := promptEntrySelection(entries, scanner)
			fmt.Println("Connecting to", addr, "expecting peer ID", expectedPeerID)

			conn, err := peer.ConnectToPeer(addr, newNode, expectedPeerID)
			if err != nil {
				fmt.Println("Connect error:", err)
				continue
			}

			if newNode.AddPeer(addr, conn) {
				go peer.ReadFromPeer(addr, conn, newNode)
			}
			continue
		}

		newNode.Broadcast(line)
	}
}

func NodeShutdown(newNode *models.Node) {
	defer newNode.Server.Shutdown()
}
