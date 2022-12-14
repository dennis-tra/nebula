package main

import (
	"strconv"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/crawl"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
)

// CrawlCommand contains the crawl sub-command configuration.
var CrawlCommand = &cli.Command{
	Name:   "crawl",
	Usage:  "Crawls the entire network starting with a set of bootstrap nodes.",
	Action: CrawlAction,
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "bootstrap-peers",
			Usage:   "Comma separated list of multi addresses of bootstrap peers",
			EnvVars: []string{"NEBULA_BOOTSTRAP_PEERS"},
		},
		&cli.IntFlag{
			Name:        "workers",
			Usage:       "How many concurrent workers should dial and crawl peers.",
			EnvVars:     []string{"NEBULA_CRAWL_WORKER_COUNT"},
			DefaultText: strconv.Itoa(config.DefaultConfig.CrawlWorkerCount),
			Value:       config.DefaultConfig.CrawlWorkerCount,
		},
		&cli.IntFlag{
			Name:        "limit",
			Usage:       "Only crawl the specified amount of peers (0 for unlimited)",
			EnvVars:     []string{"NEBULA_CRAWL_PEER_LIMIT"},
			DefaultText: strconv.Itoa(config.DefaultConfig.CrawlLimit),
			Value:       config.DefaultConfig.CrawlLimit,
		},
		&cli.BoolFlag{
			Name:    "dry-run",
			Usage:   "Don't persist anything to a database (you don't need a running DB)",
			EnvVars: []string{"NEBULA_CRAWL_DRY_RUN"},
		},
		&cli.BoolFlag{
			Name:        "neighbors",
			Usage:       "Whether to persist all k-bucket entries of a particular peer at the end of a crawl.",
			EnvVars:     []string{"NEBULA_CRAWL_NEIGHBORS"},
			DefaultText: strconv.FormatBool(config.DefaultConfig.PersistNeighbors),
			Value:       config.DefaultConfig.PersistNeighbors,
		},
		&cli.StringFlag{
			Name:        "network",
			Usage:       "Which network should be crawled (IPFS, FILECOIN, KUSAMA, POLKADOT). Presets default bootstrap peers.",
			EnvVars:     []string{"NEBULA_CRAWL_NETWORK"},
			DefaultText: string(config.DefaultConfig.Network),
			Value:       string(config.DefaultConfig.Network),
		},
	},
}

// CrawlAction is the function that is called when running `nebula crawl`.
func CrawlAction(c *cli.Context) error {
	log.Infoln("Starting Nebula crawler...")

	// Load configuration file
	conf, err := config.Init(c)
	if err != nil {
		return err
	}

	// Acquire database handle
	var dbc *db.Client
	if !c.Bool("dry-run") {
		if dbc, err = db.InitClient(c.Context, conf); err != nil {
			return err
		}
	}

	// Parse bootstrap info
	pis, err := conf.BootstrapAddrInfos()
	if err != nil {
		return errors.Wrap(err, "parsing multi addresses to peer addresses")
	}

	// Initialize scheduler that handles crawling the network.
	s, err := crawl.NewScheduler(conf, dbc)
	if err != nil {
		return errors.Wrap(err, "creating new scheduler")
	}

	return s.CrawlNetwork(c.Context, pis)
}
