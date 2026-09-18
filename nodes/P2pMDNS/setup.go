package P2pMDNS

import (
	"fmt"
	"os"

	"github.com/hashicorp/mdns"
)

func SetupMDNS(listenPort int) *mdns.Server {
	host, errHost := os.Hostname()

	// id := uuid.New().String()

	// serviceName := "_" + host + id + "._tcp"
	serviceName := "_foobar._tcp"

	if errHost != nil {
		fmt.Println("Host error")
	}
	info := []string{"My awesome service"}
	service, errService := mdns.NewMDNSService(host, serviceName, "", "", listenPort, nil, info)

	if errService != nil {
		fmt.Println("Service error")
	}
	// Create the mDNS server, defer shutdown
	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	// defer server.Shutdown()
	if err != nil {
		fmt.Println("Server error:", err)
		return nil
	}

	fmt.Println("HEREE", service.Service)
	return server
}
