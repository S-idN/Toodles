package P2pMDNS

import (
	"fmt"
	"p2pChat/models"
	"strconv"
	"sync"

	"github.com/hashicorp/mdns"
)

func LookupMDNS(newNode *models.Node) (string, error) {
	var wg sync.WaitGroup
	var peerPort string

	entryList := []*mdns.ServiceEntry{}
	entriesCh := make(chan *mdns.ServiceEntry, 10)
	wg.Go(func() {
		for entry := range entriesCh {
			fmt.Printf("Got new entry: %v\n", entry)

			if peerPort == "" && entry.Port != 0 {
				peerPort = strconv.Itoa(entry.Port)
			}
			entryList = append(entryList, entry)
			// newNode.PeerList = entryList
		}
	})

	mdns.Lookup("_foobar._tcp", entriesCh)

	close(entriesCh)
	wg.Wait()

	if peerPort == "" {
		return "", fmt.Errorf("no valid peer port found")
	}

	//Maybe add an error later hopefully
	return peerPort, nil
}
