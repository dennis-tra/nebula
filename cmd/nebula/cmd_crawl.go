package main

import (
	"context"
	"database/sql"
	"os"
	"runtime/pprof"
	"strconv"
	"time"

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
			Name:    "dry-run",
			Usage:   "Don't persist anything to a database (you don't need a running DB)",
			EnvVars: []string{"NEBULA_CRAWL_DRY_RUN"},
		},
		&cli.BoolFlag{
			Name:    "save-neighbours",
			Usage:   "Save the neighbours raletions in this crawl",
			EnvVars: []string{"NEBULA_SAVE_NEIGHBOURS"},
		},
		&cli.BoolFlag{
			Name:    "not-truncate-neighbours",
			Usage:   "Do not truncate the neighbours raletions in this crawl",
			EnvVars: []string{"NEBULA_NOT_TRUNCATE_NEIGHBOURS"},
		},
		&cli.BoolFlag{
			Name:    "save-connections",
			Usage:   "Save the connections raletions in this crawl",
			EnvVars: []string{"NEBULA_SAVE_CONNECTIONS"},
		},
		&cli.BoolFlag{
			Name:    "not-truncate-connections",
			Usage:   "Do not truncate the connections raletions in this crawl",
			EnvVars: []string{"NEBULA_NOT_TRUNCATE_CONNECTIONS"},
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
	var dbh *sql.DB
	if !c.Bool("dry-run") {
		if dbh, err = db.Open(c.Context); err != nil {
			return err
		}
	}
	saveNeighbour := false
	if c.Bool("save-neighbours") {
		saveNeighbour = true
	}
	notTruncateNeighbour := false
	if c.Bool("not-truncate-neighbours") {
		notTruncateNeighbour = true
	}
	saveConnection := false
	if c.Bool("save-connections") {
		saveConnection = true
	}
	notTruncateConnection := false
	if c.Bool("not-truncate-connections") {
		notTruncateConnection = true
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
	s, err := crawl.NewScheduler(c.Context, dbh, saveNeighbour, notTruncateNeighbour, saveConnection, notTruncateConnection)
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
