package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/crawl"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
)

var crawlConfig = &config.Crawl{
	Root:             rootConfig,
	CrawlWorkerCount: 1000,
	CrawlLimit:       0,
	PersistNeighbors: false,
	CheckExposed:     false,
	FilePathUdgerDB:  "",
	Network:          string(config.NetworkIPFS),
	BootstrapPeers:   cli.NewStringSlice(),
	Protocols:        cli.NewStringSlice("/ipfs/kad/1.0.0"),
	DryRun:           false,
}

// CrawlCommand contains the crawl sub-command configuration.
var CrawlCommand = &cli.Command{
	Name:   "crawl",
	Usage:  "Crawls the entire network starting with a set of bootstrap nodes.",
	Action: CrawlAction,
	Before: func(c *cli.Context) error {
		if err := crawlConfig.ConfigureNetwork(); err != nil {
			return err
		}

		// Give CLI option precedence
		if c.IsSet("protocols") {
			crawlConfig.Protocols = cli.NewStringSlice(c.StringSlice("protocols")...)
		}

		if c.IsSet("bootstrap-peers") {
			crawlConfig.BootstrapPeers = cli.NewStringSlice(c.StringSlice("bootstrap-peers")...)
		}

		return nil
	},
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:        "bootstrap-peers",
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
			Name:        "limit",
			Usage:       "Only crawl the specified amount of peers (0 for unlimited)",
			EnvVars:     []string{"NEBULA_CRAWL_PEER_LIMIT"},
			Value:       crawlConfig.CrawlLimit,
			Destination: &crawlConfig.CrawlLimit,
		},
		&cli.BoolFlag{
			Name:        "dry-run",
			Usage:       "Don't persist anything to a database (you don't need a running DB)",
			EnvVars:     []string{"NEBULA_CRAWL_DRY_RUN"},
			Value:       crawlConfig.DryRun,
			Destination: &crawlConfig.DryRun,
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

	// Acquire database handle
	var (
		dbc *db.Client
		err error
	)
	if !c.Bool("dry-run") {
		if dbc, err = db.InitClient(c.Context, rootConfig); err != nil {
			return err
		}
	}

	// Parse bootstrap info
	pis, err := crawlConfig.BootstrapAddrInfos()
	if err != nil {
		return fmt.Errorf("parsing multi addresses to peer addresses: %w", err)
	}

	// Initialize scheduler that handles crawling the network.
	s, err := crawl.NewScheduler(crawlConfig, dbc)
	if err != nil {
		return fmt.Errorf("creating new scheduler: %w", err)
	}

	return s.CrawlNetwork(c.Context, pis)
}
