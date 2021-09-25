package main

import (
	"strconv"

	"github.com/friendsofgo/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/ping"
)

// PingCommand contains the monitor sub-command configuration.
var PingCommand = &cli.Command{
	Name:   "ping",
	Usage:  "Runs an ICMP latency measurement over the set of online peers of the most recent crawl",
	Action: PingAction,
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:        "workers",
			Usage:       "How many concurrent workers should ping peers.",
			EnvVars:     []string{"NEBULA_PING_WORKER_COUNT"},
			DefaultText: strconv.Itoa(config.DefaultConfig.PingWorkerCount),
			Value:       config.DefaultConfig.PingWorkerCount,
		},
		&cli.IntFlag{
			Name:        "limit",
			Usage:       "Only ping the specified amount of peers (0 for unlimited)",
			EnvVars:     []string{"NEBULA_PING_PEER_LIMIT"},
			DefaultText: strconv.Itoa(config.DefaultConfig.CrawlLimit),
			Value:       config.DefaultConfig.CrawlLimit,
		},
	},
}

// PingAction is the function that is called when running `nebula ping`.
func PingAction(c *cli.Context) error {
	log.Infoln("Starting Nebula ping latency measurement...")

	// Load configuration file
	ctx, conf, err := config.FillContext(c)
	if err != nil {
		return errors.Wrap(err, "filling context with configuration")
	}
	c.Context = ctx

	dbc, err := db.InitClient(c.Context)
	if err != nil {
		return err
	}

	// Start prometheus metrics endpoint
	if err = metrics.RegisterCrawlMetrics(); err != nil {
		return err
	}
	if err = metrics.ListenAndServe(conf.PrometheusHost, conf.PrometheusPort); err != nil {
		return errors.Wrap(err, "initialize metrics")
	}

	// Initialize scheduler that handles crawling the network.
	s, err := ping.NewScheduler(c.Context, dbc)
	if err != nil {
		return errors.Wrap(err, "creating new scheduler")
	}

	go func() {
		// Nebula was asked to stop (e.g. SIGINT) -> tell the scheduler to stop
		<-c.Context.Done()
		s.Shutdown()
	}()

	return s.PingNetwork()
}
