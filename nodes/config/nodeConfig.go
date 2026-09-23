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
	hashId_portion := GenerateBaseId()
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
	name := inputName[:len(inputName)-1]
	newNode.Name = name
	fmt.Println("Name is", newNode.Name)

	newNode.ListenPort = listenPort

	go peer.StartServer(listenPort, newNode)

	newNode.Server = P2pMDNS.SetupMDNS(listenPort, newNode)

	StartPeerOps(newNode)

	fmt.Println("STRUCT IS THIS:", newNode)

	return newNode
}

func promptEntrySelection(entries []string, scanner *bufio.Scanner) (addr string, nodeID string, peerName string) {
	for {
		fmt.Println("Discovered peers:")
		for i, e := range entries {
			parts := strings.SplitN(e, "|", 4)
			label := e
			if len(parts) == 4 {
				label = fmt.Sprintf("%s (%s)", parts[3], parts[1])
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

		parts := strings.SplitN(entries[choice], "|", 4)
		if len(parts) == 4 {
			return parts[1], parts[2], parts[3]
		}
		return entries[choice], "", ""
	}
}

func StartPeerOps(newNode *models.Node) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Type a message to broadcast to connected peers, /scan to discover peers, or /disconnect to drop a peer.")

	for scanner.Scan() {
		line := scanner.Text()

		if line == "/scan" {
			fmt.Println("Scanning local network...")
			entries, err := P2pMDNS.LookupMDNS(newNode)

			if err != nil {
				fmt.Println("Scan error or no peers found:", err)
				continue
			}

			addr, expectedPeerID, peerName := promptEntrySelection(entries, scanner)

			if expectedPeerID == "" {
				fmt.Println("[system] No fingerprint advertised by this peer - refusing to connect.")
				continue
			}

			fmt.Println()
			fmt.Println("=== Verify peer identity before connecting ===")
			fmt.Println("Your name:         ", newNode.Name)
			fmt.Println("Your fingerprint:  ", newNode.NodeId)
			fmt.Println("Peer name:         ", peerName)
			fmt.Println("Peer fingerprint:  ", expectedPeerID)
			fmt.Println("Confirm with the other person, out loud or via another channel,")
			fmt.Println("that their fingerprint matches what's shown above.")
			fmt.Print("Type 'yes' to trust and connect, anything else to cancel: ")

			if !scanner.Scan() {
				continue
			}
			if strings.TrimSpace(strings.ToLower(scanner.Text())) != "yes" {
				fmt.Println("[system] Connection cancelled.")
				continue
			}

			fmt.Println("Connecting to", addr, "expecting peer ID", expectedPeerID)

			conn, err := peer.ConnectToPeer(addr, newNode, expectedPeerID)
			if err != nil {
				fmt.Println("Connect error:", err)
				continue
			}

			if newNode.AddPeer(addr, conn) {
				newNode.RememberPeer(addr, expectedPeerID)
				go peer.ReadFromPeer(addr, conn, newNode)
			}
			continue
		}

		if line == "/disconnect" {
			addrs := newNode.PeerAddrs()
			if len(addrs) == 0 {
				fmt.Println("[system] No connected peers.")
				continue
			}

			fmt.Println("Connected peers:")
			for i, a := range addrs {
				fmt.Printf("  [%d] %s\n", i, a)
			}
			fmt.Print("Select a peer to disconnect: ")

			if !scanner.Scan() {
				continue
			}
			choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil || choice < 0 || choice >= len(addrs) {
				fmt.Println("[system] Invalid selection.")
				continue
			}

			addr := addrs[choice]
			newNode.ForgetPeer(addr)
			newNode.RemovePeer(addr)
			fmt.Println("[system] Disconnected from", addr)
			continue
		}

		newNode.Broadcast(line)
	}
}

func NodeShutdown(newNode *models.Node) {
	defer newNode.Server.Shutdown()
}
