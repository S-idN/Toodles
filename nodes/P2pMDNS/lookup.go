package P2pMDNS

import (
	"fmt"
	"p2pChat/models"
	"strings"
	"time"

	"github.com/hashicorp/mdns"
)

func StartContinuousScan(newNode *models.Node, interval time.Duration) {
	go func() {
		for {
			entriesCh := make(chan *mdns.ServiceEntry, 10)
			go func() {
				for entry := range entriesCh {
					if entry.Port == 0 || !strings.Contains(entry.Name, "_toodles__p2pchat._tcp") {
						continue
					}
					nodeID, peerName := "", ""
					for _, f := range entry.InfoFields {
						if strings.HasPrefix(f, "nodeid=") {
							nodeID = strings.TrimPrefix(f, "nodeid=")
						}
						if strings.HasPrefix(f, "name=") {
							peerName = strings.TrimPrefix(f, "name=")
						}
					}
					addr := fmt.Sprintf("%s:%d", entry.AddrV4, entry.Port)
					if nodeID != "" && nodeID != newNode.NodeId { // don't discover self
						newNode.AddDiscovered(models.DiscoveredPeer{Addr: addr, Name: peerName, ID: nodeID})
					}
				}
			}()
			mdns.Lookup("_toodles__p2pchat._tcp", entriesCh)
			close(entriesCh)
			time.Sleep(interval)
		}
	}()
}
