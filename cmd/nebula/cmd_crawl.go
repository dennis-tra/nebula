package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/dennis-tra/nebula-crawler/pkg/db"

	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/crawl"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
)

// CrawlCommand contains the crawl sub-command configuration.
var CrawlCommand = &cli.Command{
	Name:   "crawl",
	Usage:  "Crawls the entire network based on a set of bootstrap nodes.",
	Action: CrawlAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "dry-run",
			Usage:   "Don't persist results but just crawl the network.",
			EnvVars: []string{"NEBULA_CRAWL_DRY_RUN"},
		},
		&cli.IntFlag{
			Name:        "workers",
			Usage:       "How many concurrent workers should dial and crawl peers.",
			EnvVars:     []string{"NEBULA_CRAWL_WORKER_COUNT"},
			DefaultText: strconv.Itoa(config.DefaultConfig.WorkerCount),
			Value:       config.DefaultConfig.WorkerCount,
		},
		&cli.DurationFlag{
			Name:        "dial-timeout",
			Usage:       "How long should be waited before a dial is considered unsuccessful.",
			EnvVars:     []string{"NEBULA_CRAWL_DIAL_TIMEOUT"},
			DefaultText: config.DefaultConfig.DialTimeout.String(),
			Value:       config.DefaultConfig.DialTimeout,
		},
	},
}

// CrawlAction is the function that is called when running nebula crawl.
func CrawlAction(c *cli.Context) error {
	log.Infoln("Starting Nebula crawler...")

	// Load configuration file
	c, err := config.FillContext(c)
	if err != nil {
		return errors.Wrap(err, "filling context with configuration")
	}

	// Initialize new database client
	var dbc *db.Client
	if !c.Bool("dry-run") {
		if dbc, err = db.NewClient(); err != nil {
			return errors.Wrap(err, "initialize db")
		}
	}

	// Start prometheus metrics endpoint
	if err = metrics.RegisterListenAndServe(c.Context); err != nil {
		return errors.Wrap(err, "initialize metrics")
	}

	o, _ := crawl.NewOrchestrator(c.Context, dbc)
	go o.CrawlNetwork([]peer.AddrInfo{*pi})

	select {
	case <-c.Context.Done():
		o.Shutdown()
	case <-o.SigDone():
	}

	f, _ := os.Create("errors.txt")
	o.Errors.Range(func(errorStr, value interface{}) bool {
		fmt.Fprintf(f, "%s\n", errorStr)
		return true
	})
	f.Close()

	return nil
}
