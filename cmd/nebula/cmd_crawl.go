package main

import (
	"context"
	"fmt"
	"os"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/crawl"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
)

// CrawlCommand contains the crawl sub-command configuration.
var CrawlCommand = &cli.Command{
	Name:   "crawl",
	Usage:  "Crawls the entire network starting with a set of bootstrap nodes.",
	Action: CrawlAction,
	Flags: []cli.Flag{
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
			Name:        "latencies",
			Usage:       "Measure latencies to peers by exchanging ICM ping packets (approx. doubles crawl duration)",
			EnvVars:     []string{"NEBULA_CRAWL_LATENCIES"},
			DefaultText: fmt.Sprintf("%v", config.DefaultConfig.MeasureLatencies),
			Value:       config.DefaultConfig.MeasureLatencies,
		},
		&cli.BoolFlag{
			Name:        "neighbors",
			Usage:       "Persist DHT peer neighbors",
			EnvVars:     []string{"NEBULA_CRAWL_NEIGHBORS"},
			DefaultText: fmt.Sprintf("%v", config.DefaultConfig.PersistNeighbors),
			Value:       config.DefaultConfig.PersistNeighbors,
		}, &cli.BoolFlag{
			Name:        "clear-neighbors",
			Usage:       "If the neighbors table should be cleared prior persisting neighbors",
			EnvVars:     []string{"NEBULA_CRAWL_CLEAR_NEIGHBORS"},
			DefaultText: fmt.Sprintf("%v", config.DefaultConfig.ClearNeighbors),
			Value:       config.DefaultConfig.ClearNeighbors,
		},
		&cli.BoolFlag{
			Name:    "dry-run",
			Usage:   "Don't persist anything to a database (you don't need a running DB)",
			EnvVars: []string{"NEBULA_CRAWL_DRY_RUN"},
		},
	},
}

// CrawlAction is the function that is called when running `nebula crawl`.
func CrawlAction(c *cli.Context) error {
	log.Infoln("Starting Nebula crawler...")

	// Load configuration file
	ctx, conf, err := config.FillContext(c)
	if err != nil {
		return errors.Wrap(err, "filling context with configuration")
	}
	c.Context = ctx

	// Acquire database handle
	var dbc *db.Client
	if !c.Bool("dry-run") {
		if dbc, err = db.InitClient(c.Context); err != nil {
			return err
		}
	}

	// Start prometheus metrics endpoint
	if err = metrics.RegisterCrawlMetrics(); err != nil {
		return err
	}
	if err = metrics.ListenAndServe(conf.PrometheusHost, conf.PrometheusPort); err != nil {
		return errors.Wrap(err, "initialize metrics")
	}

	// Parse bootstrap info
	pis, err := conf.BootstrapAddrInfos()
	if err != nil {
		return errors.Wrap(err, "parsing multi addresses to peer addresses")
	}

	// Initialize scheduler that handles crawling the network.
	s, err := crawl.NewScheduler(c.Context, dbc)
	if err != nil {
		return errors.Wrap(err, "creating new scheduler")
	}

	go dumpGoRoutines(c.Context)

	go func() {
		// Nebula was asked to stop (e.g. SIGINT) -> tell the scheduler to stop
		<-c.Context.Done()
		s.Shutdown()
	}()

	return s.CrawlNetwork(pis)
}

func dumpGoRoutines(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Minute):
			if err := pprof.Lookup("goroutine").WriteTo(os.Stdout, 1); err != nil {
				log.WithError(err).Warnln("Could not dump goroutines")
			}
		}
	}
}
