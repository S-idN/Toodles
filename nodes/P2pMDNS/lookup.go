package P2pMDNS

import (
	"fmt"
	"sync"

	"github.com/hashicorp/mdns"
)

func LookupMDNS() {
	var wg sync.WaitGroup

	entriesCh := make(chan *mdns.ServiceEntry, 10)
	wg.Go(func() {
		for entry := range entriesCh {
			fmt.Printf("Got new entry: %v\n", entry)
		}
	})

	mdns.Lookup("_foobar._tcp", entriesCh)

	close(entriesCh)
	wg.Wait()
}
