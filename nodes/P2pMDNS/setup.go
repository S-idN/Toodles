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

	info := []string{"nodeid=" + newNode.NodeId, "name=" + newNode.Name}
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
