package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/dennis-tra/nebula-crawler/pkg/db"

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
		&cli.IntFlag{
			Name:        "limit",
			Usage:       "Only crawl the specified amount of peers (0 for unlimited)",
			EnvVars:     []string{"NEBULA_CRAWL_PEER_LIMIT"},
			DefaultText: strconv.Itoa(config.DefaultConfig.CrawlLimit),
			Value:       config.DefaultConfig.CrawlLimit,
		},
	},
}

// CrawlAction is the function that is called when running nebula crawl.
func CrawlAction(c *cli.Context) error {
	log.Infoln("Starting Nebula crawler...")

	// Load configuration file
	ctx, conf, err := config.FillContext(c)
	if err != nil {
		return errors.Wrap(err, "filling context with configuration")
	}
	c.Context = ctx

	// Acquire database handle
	var dbh *sql.DB
	if !c.Bool("dry-run") {
		if dbh, err = db.Open(c.Context); err != nil {
			return err
		}
	}

	// Start prometheus metrics endpoint
	if err = metrics.RegisterListenAndServe(conf.PrometheusHost, conf.PrometheusPort, "crawl"); err != nil {
		return errors.Wrap(err, "initialize metrics")
	}

	// Parse bootstrap info
	pis, err := conf.BootstrapAddrInfos()
	if err != nil {
		return errors.Wrap(err, "parsing multi addresses to peer addresses")
	}

	// Initialize orchestrator that handles crawling the network.
	o, _ := crawl.NewOrchestrator(c.Context, dbh)
	go o.CrawlNetwork(pis)

	select {
	case <-c.Context.Done():
		// Nebula was asked to stop (e.g. SIGINT) -> tell the orchestrator to stop
		o.Shutdown()
	case <-o.SigDone():
		// the orchestrator finished autonomously
	}

	// Temporary code: save all errors that were encountered
	f, _ := os.Create("errors.txt")
	o.Errors.Range(func(errorStr, value interface{}) bool {
		fmt.Fprintf(f, "%s\n", errorStr)
		return true
	})
	f.Close()
	// Temporary code: save all errors that were encountered

	return nil
}
