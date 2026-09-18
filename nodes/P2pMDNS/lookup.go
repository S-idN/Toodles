package P2pMDNS

import (
	"fmt"
	"p2pChat/models"
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
			if entry.Port != 0 {
				entryList = append(entryList, fmt.Sprintf("%s|%s:%d", entry.Name, entry.AddrV4, entry.Port))
			}
		}
	})

	mdns.Lookup("_foobar._tcp", entriesCh)

	close(entriesCh)
	wg.Wait()

	newNode.PeerList = entryList

	if len(entryList) == 0 {
		return nil, fmt.Errorf("no valid peers found")
	}

	//Maybe add an error later hopefully
	return entryList, nil
}
