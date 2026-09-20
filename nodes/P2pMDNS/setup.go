package P2pMDNS

import (
	"fmt"
	"os"
	"p2pChat/models"

	"github.com/hashicorp/mdns"
)

func SetupMDNS(listenPort int, newNode *models.Node) *mdns.Server {
	host, errHost := os.Hostname()

	serviceName := "_toodles__p2pchat._tcp"

	if errHost != nil {
		fmt.Println("Host error")
	}

	// Node ID goes in the TXT record so peers can verify who they're
	// connecting to before trusting the connection.
	info := []string{"nodeid=" + newNode.NodeId}
	service, errService := mdns.NewMDNSService(host, serviceName, "", "", listenPort, nil, info)

	if errService != nil {
		fmt.Println("Service error")
	}

	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		fmt.Println("Server error:", err)
		return nil
	}

	return server
}
