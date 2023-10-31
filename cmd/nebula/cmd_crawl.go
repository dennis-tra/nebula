package main

import (
	"fmt"
	"strings"

	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/crawl"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
)

var crawlConfig = &config.Crawl{
	Root:             rootConfig,
	CrawlWorkerCount: 1000,
	WriteWorkerCount: 10,
	CrawlLimit:       0,
	PersistNeighbors: false,
	CheckExposed:     false,
	FilePathUdgerDB:  "",
	Network:          string(config.NetworkIPFS),
	BootstrapPeers:   cli.NewStringSlice(),
	Protocols:        cli.NewStringSlice(string(kaddht.ProtocolDHT)),
	AddrTrackTypeStr: "public",
	AddrDialTypeStr:  "public",
}

// CrawlCommand contains the crawl sub-command configuration.
var CrawlCommand = &cli.Command{
	Name:   "crawl",
	Usage:  "Crawls the entire network starting with a set of bootstrap nodes.",
	Action: CrawlAction,
	Before: func(c *cli.Context) error {
		// based on the network setting, return the default bootstrap peers and protocols
		bootstrapPeers, protocols, err := config.ConfigureNetwork(crawlConfig.Network)
		if err != nil {
			return err
		}

		// Give CLI option precedence
		if c.IsSet("protocols") {
			crawlConfig.Protocols = cli.NewStringSlice(c.StringSlice("protocols")...)
		} else {
			crawlConfig.Protocols = protocols
		}

		if c.IsSet("bootstrap-peers") {
			crawlConfig.BootstrapPeers = cli.NewStringSlice(c.StringSlice("bootstrap-peers")...)
		} else {
			crawlConfig.BootstrapPeers = bootstrapPeers
		}

		if log.GetLevel() >= log.DebugLevel {
			log.Debugln("Using the following configuration:")
			fmt.Println(crawlConfig.String())
		}

		switch config.AddrType(strings.ToLower(crawlConfig.AddrTrackTypeStr)) {
		case config.AddrTypePrivate:
			crawlConfig.AddrTrackTypeStr = string(config.AddrTypePrivate)
		case config.AddrTypePublic:
			crawlConfig.AddrTrackTypeStr = string(config.AddrTypePublic)
		case config.AddrTypeAny:
			crawlConfig.AddrTrackTypeStr = string(config.AddrTypeAny)
		default:
			return fmt.Errorf("unknown type of addresses to track: %s (supported values are private, public, any)", crawlConfig.AddrTrackTypeStr)
		}

		switch config.AddrType(strings.ToLower(crawlConfig.AddrDialTypeStr)) {
		case config.AddrTypePrivate:
			crawlConfig.AddrDialTypeStr = string(config.AddrTypePrivate)
		case config.AddrTypePublic:
			crawlConfig.AddrDialTypeStr = string(config.AddrTypePublic)
		case config.AddrTypeAny:
			crawlConfig.AddrDialTypeStr = string(config.AddrTypeAny)
		default:
			return fmt.Errorf("unknown type of addresses to dial: %s (supported values are private, public, any)", crawlConfig.AddrDialTypeStr)
		}

		return nil
	},
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:        "bootstrap-peers", // TODO: rename to bootstrappers
			Usage:       "Comma separated list of multi addresses of bootstrap peers",
			EnvVars:     []string{"NEBULA_CRAWL_BOOTSTRAP_PEERS", "NEBULA_BOOTSTRAP_PEERS" /* legacy */},
			Destination: crawlConfig.BootstrapPeers,
			DefaultText: "default " + crawlConfig.Network,
		},
		&cli.StringSliceFlag{
			Name:        "protocols",
			Usage:       "Comma separated list of protocols that this crawler should look for",
			EnvVars:     []string{"NEBULA_CRAWL_PROTOCOLS", "NEBULA_PROTOCOLS" /* legacy */},
			Value:       crawlConfig.Protocols,
			Destination: crawlConfig.Protocols,
		},
		&cli.IntFlag{
			Name:        "workers",
			Usage:       "How many concurrent workers should dial and crawl peers.",
			EnvVars:     []string{"NEBULA_CRAWL_WORKER_COUNT"},
			Value:       crawlConfig.CrawlWorkerCount,
			Destination: &crawlConfig.CrawlWorkerCount,
		},
		&cli.IntFlag{
			Name:        "write-workers",
			Usage:       "How many concurrent workers should write crawl results to the database.",
			EnvVars:     []string{"NEBULA_CRAWL_WRITE_WORKER_COUNT"},
			Value:       crawlConfig.WriteWorkerCount,
			Destination: &crawlConfig.WriteWorkerCount,
			Hidden:      true,
		},
		&cli.IntFlag{
			Name:        "limit",
			Usage:       "Only crawl the specified amount of peers (0 for unlimited)",
			EnvVars:     []string{"NEBULA_CRAWL_PEER_LIMIT"},
			Value:       crawlConfig.CrawlLimit,
			Destination: &crawlConfig.CrawlLimit,
		},
		&cli.BoolFlag{
			Name:        "neighbors",
			Usage:       "Whether to persist all k-bucket entries of a particular peer at the end of a crawl.",
			EnvVars:     []string{"NEBULA_CRAWL_NEIGHBORS"},
			Value:       crawlConfig.PersistNeighbors,
			Destination: &crawlConfig.PersistNeighbors,
		},
		&cli.BoolFlag{
			Name:        "check-exposed",
			Usage:       "Whether to check if the Kubo API is exposed. Checking also includes crawling the API.",
			EnvVars:     []string{"NEBULA_CRAWL_CHECK_EXPOSED"},
			Value:       crawlConfig.CheckExposed,
			Destination: &crawlConfig.CheckExposed,
			Category:    flagCategoryNetwork,
		},
		&cli.StringFlag{
			Name:        "addr-track-type",
			Usage:       "Which type addresses should be stored to the database (private, public, any)",
			EnvVars:     []string{"NEBULA_CRAWL_ADDR_TRACK_TYPE"},
			Value:       crawlConfig.AddrTrackTypeStr,
			Destination: &crawlConfig.AddrTrackTypeStr,
		},
		&cli.StringFlag{
			Name:        "addr-dial-type",
			Usage:       "Which type of addresses should Nebula try to dial (private, public, any)",
			EnvVars:     []string{"NEBULA_CRAWL_ADDR_DIAL_TYPE"},
			Value:       crawlConfig.AddrDialTypeStr,
			Destination: &crawlConfig.AddrDialTypeStr,
		},
		&cli.StringFlag{
			Name:        "network",
			Usage:       "Which network should be crawled (IPFS, FILECOIN, KUSAMA, POLKADOT). Presets default bootstrap peers and protocol.",
			EnvVars:     []string{"NEBULA_CRAWL_NETWORK"},
			Value:       crawlConfig.Network,
			Destination: &crawlConfig.Network,
		},
	},
}

// CrawlAction is the function that is called when running `nebula crawl`.
func CrawlAction(c *cli.Context) error {
	log.Infoln("Starting Nebula crawler...")
	defer log.Infoln("Stopped Nebula crawler.")

	// initialize a new database client based on the given configuration.
	// Options are Postgres, JSON, and noop (dry-run).
	dbc, err := db.NewClient(c.Context, rootConfig.Database)
	if err != nil {
		return fmt.Errorf("new database client: %w", err)
	}
	defer func() {
		if err := dbc.Close(); err != nil {
			log.WithError(err).Warnln("Failed closing database handle")
		}
	}()

	// initialize crawl instance that is responsible for setting up the internal
	// bits and pieces (engine, network stack, etc.).
	crawl, err := crawl.New(dbc, crawlConfig)
	if err != nil {
		return fmt.Errorf("new crawl: %w", err)
	}

	// instruct the crawl instance to, well, crawl the network.
	return crawl.CrawlNetwork(c.Context)
}
