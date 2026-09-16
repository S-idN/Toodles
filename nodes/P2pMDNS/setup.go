package P2pMDNS

import (
	"fmt"
	"os"

	"github.com/hashicorp/mdns"
)

func SetupMDNS(listenPort int) (*mdns.Server, error) {
	host, errHost := os.Hostname()

	if errHost != nil {
		fmt.Println("Host error")
	}
	info := []string{"My awesome service"}
	service, errService := mdns.NewMDNSService(host, "_foobar._tcp", "", "", listenPort, nil, info)

	if errService != nil {
		fmt.Println("Service error")
	}
	// Create the mDNS server, defer shutdown
	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	// defer server.Shutdown()
	if err != nil {
		return nil, err
	}

	fmt.Println("HEREE", service.Service)
	return server, nil
}
