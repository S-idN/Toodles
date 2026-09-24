package config

import (
	"bufio"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"p2pChat/P2pMDNS"
	"p2pChat/identity"
	"p2pChat/models"
	peer "p2pChat/peers"

	"github.com/google/uuid"
)

func GenerateBaseId() string {
	id := uuid.New()
	hashedId := sha512.Sum512_256(id[:])
	return hex.EncodeToString(hashedId[:])
}

func GenerateFinalId() string {
	return GenerateBaseId()
}

// NewNode sets up identity only. Assign hooks after this returns, then call
// StartServices - this ordering ensures no discovery/connection event fires
// before the GUI is listening for it.
func NewNode(name string, listenPort int) (*models.Node, error) {
	newNode := &models.Node{}

	key, err := identity.LoadOrCreateKey(".")
	if err != nil {
		return nil, fmt.Errorf("identity error: %w", err)
	}
	newNode.PrivateKey = key
	newNode.NodeId = identity.NodeIDFromKey(key)
	newNode.Name = name
	newNode.ListenPort = listenPort

	return newNode, nil
}

// StartServices starts the TLS listener, mDNS advertising, and continuous
// background discovery. Call only after hooks are assigned on the node.
func StartServices(newNode *models.Node) {
	go peer.StartServer(newNode.ListenPort, newNode)
	newNode.Server = P2pMDNS.SetupMDNS(newNode.ListenPort, newNode)
	P2pMDNS.StartContinuousScan(newNode, 5*time.Second)
}

func PromptNameFromStdin() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter a node name:")
	inputName, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return inputName[:len(inputName)-1], nil
}

func NodeShutdown(newNode *models.Node) {
	if newNode.Server != nil {
		newNode.Server.Shutdown()
	}
}
