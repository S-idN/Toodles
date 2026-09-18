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

	//NODE : Conn
	// newNode.Conn = peer.StartServer(listenPort)

	newNode.Server = P2pMDNS.SetupMDNS(listenPort)

	StartPeerOps(newNode)

	fmt.Println("STRUCT IS THIS:", newNode)

	peerConn, err := peer.ConnectToPeer("localhost:" + newNode.PeerList[0])
	newNode.PeerConn = peerConn

	return newNode
}

func StartPeerOps(newNode *models.Node) {
	for {
		fmt.Println("Scanning local network...")
		peerPort, err := P2pMDNS.LookupMDNS(newNode)

		if err != nil {
			fmt.Println("Scan error or no peers found:", err)
			time.Sleep(10 * time.Second)
			continue
		}

		fmt.Println("Struct is:", newNode)

		peerConn, err := peer.ConnectToPeer("localhost:" + peerPort)
		newNode.PeerConn = peerConn

		if err != nil {
			fmt.Println("Connect error:", err)
			time.Sleep(10 * time.Second)
			continue
		}

		peer.WriteToPeer(newNode.PeerConn)
		time.Sleep(10 * time.Second)
	}
}

func NodeShutdown(newNode *models.Node) {
	defer newNode.Server.Shutdown()
}
