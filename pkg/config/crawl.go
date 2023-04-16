package config

import (
	"encoding/json"
	"fmt"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/urfave/cli/v2"
)

// Crawl contains general user configuration.
type Crawl struct {
	Root *Root

	// The list of multi addresses that will make up the entry points to the network.
	BootstrapPeers *cli.StringSlice

	// The list of protocols that this crawler should look for.
	Protocols *cli.StringSlice

	// How many parallel workers should crawl the network.
	CrawlWorkerCount int

	// Only crawl the specified amount of peers
	CrawlLimit int

	// Whether to persist all k-bucket entries
	PersistNeighbors bool

	// Whether to check if the Kubo API is exposed
	CheckExposed bool

	// File path to the udger datbase
	FilePathUdgerDB string

	// The network to crawl
	Network string
	DryRun  bool
}

// ReachedCrawlLimit returns true if the crawl limit is configured (aka != 0) and the crawled peers exceed this limit.
func (c *Crawl) ReachedCrawlLimit(crawled int) bool {
	return c.CrawlLimit > 0 && crawled >= c.CrawlLimit
}

// String prints the configuration as a json string
func (c *Crawl) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return fmt.Sprintf("%s", data)
}

// BootstrapAddrInfos parses the configured multi address strings to proper multi addresses.
func (c *Crawl) BootstrapAddrInfos() ([]peer.AddrInfo, error) {
	peerAddrs := map[peer.ID][]ma.Multiaddr{}
	for _, maddrStr := range c.BootstrapPeers.Value() {

		maddr, err := ma.NewMultiaddr(maddrStr)
		if err != nil {
			return nil, err
		}

		pi, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return nil, err
		}

		_, found := peerAddrs[pi.ID]
		if found {
			peerAddrs[pi.ID] = append(peerAddrs[pi.ID], pi.Addrs...)
		} else {
			peerAddrs[pi.ID] = pi.Addrs
		}
	}

	var pis []peer.AddrInfo
	for pid, addrs := range peerAddrs {
		pi := peer.AddrInfo{
			ID:    pid,
			Addrs: addrs,
		}
		pis = append(pis, pi)
	}

	return pis, nil
}

func (c *Crawl) ConfigureNetwork() error {
	switch Network(c.Network) {
	case NetworkFilecoin:
		c.BootstrapPeers = cli.NewStringSlice(BootstrapPeersFilecoin...)
		c.Protocols = cli.NewStringSlice("/fil/kad/testnetnet/kad/1.0.0")
	case NetworkKusama:
		c.BootstrapPeers = cli.NewStringSlice(BootstrapPeersKusama...)
		c.Protocols = cli.NewStringSlice("/ksmcc3/kad")
	case NetworkPolkadot:
		c.BootstrapPeers = cli.NewStringSlice(BootstrapPeersPolkadot...)
		c.Protocols = cli.NewStringSlice("/dot/kad")
	case NetworkRococo:
		c.BootstrapPeers = cli.NewStringSlice(BootstrapPeersRococo...)
		c.Protocols = cli.NewStringSlice("/rococo/kad")
	case NetworkWestend:
		c.BootstrapPeers = cli.NewStringSlice(BootstrapPeersWestend...)
		c.Protocols = cli.NewStringSlice("/wnd2/kad")
	case NetworkArabica:
		c.BootstrapPeers = cli.NewStringSlice(BootstrapPeersArabica...)
		c.Protocols = cli.NewStringSlice("/celestia/arabica-6/kad/1.0.0")
	case NetworkMocha:
		c.BootstrapPeers = cli.NewStringSlice(BootstrapPeersMocha...)
		c.Protocols = cli.NewStringSlice("/celestia/mocha/kad/1.0.0")
	case NetworkBlockRa:
		c.BootstrapPeers = cli.NewStringSlice(BootstrapPeersBlockspaceRace...)
		c.Protocols = cli.NewStringSlice("/celestia/blockspacerace-0/kad/1.0.0")
	case NetworkIPFS:
		bps := []string{}
		for _, maddr := range dht.DefaultBootstrapPeers {
			bps = append(bps, maddr.String())
		}
		c.BootstrapPeers = cli.NewStringSlice(bps...)
		c.Protocols = cli.NewStringSlice("/ipfs/kad/1.0.0")
	default:
		return fmt.Errorf("unknown network identifier: %s", c.Network)
	}

	return nil
}
