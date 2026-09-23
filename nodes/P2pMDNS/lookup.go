package P2pMDNS

import (
	"fmt"
	"p2pChat/models"
	"strings"
	"sync"

	"github.com/hashicorp/mdns"
)

func LookupMDNS(newNode *models.Node) ([]string, error) {
	var wg sync.WaitGroup

	entryList := []string{}
	entriesCh := make(chan *mdns.ServiceEntry, 10)
	wg.Go(func() {
		for entry := range entriesCh {
			fmt.Printf("Got new entry: %v\n", entry)

			if entry.Port == 0 || !strings.Contains(entry.Name, "_toodles__p2pchat._tcp") {
				continue
			}

			nodeID := ""
			peerName := ""
			for _, field := range entry.InfoFields {
				if strings.HasPrefix(field, "nodeid=") {
					nodeID = strings.TrimPrefix(field, "nodeid=")
				}
				if strings.HasPrefix(field, "name=") {
					peerName = strings.TrimPrefix(field, "name=")
				}
			}

			// format: name|addr:port|nodeid|peername
			entryList = append(entryList, fmt.Sprintf("%s|%s:%d|%s|%s", entry.Name, entry.AddrV4, entry.Port, nodeID, peerName))
		}
	})

	mdns.Lookup("_toodles__p2pchat._tcp", entriesCh)

	close(entriesCh)
	wg.Wait()

	newNode.PeerList = entryList

	if len(entryList) == 0 {
		return nil, fmt.Errorf("no valid peers found")
	}

	return entryList, nil
}
