package config

import (
	"encoding/json"
	"fmt"

	dht "github.com/libp2p/go-libp2p-kad-dht"
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

	// How many parallel workers should write crawl results to the database
	WriteWorkerCount int

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
}

// String prints the configuration as a json string
func (c *Crawl) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
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
	case NetworkEthereum:
		c.BootstrapPeers = cli.NewStringSlice(BootstrapPeersEthereum...)
		c.Protocols = cli.NewStringSlice("discv5") // TODO
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
