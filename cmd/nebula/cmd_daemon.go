package main

import (
	"database/sql"
	"strconv"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
)

// DaemonCommand contains the daemon sub-command configuration.
var DaemonCommand = &cli.Command{
	Name:   "daemon",
	Usage:  "Start a long running process that crawls and monitors the DHT network.",
	Action: DaemonAction,
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:        "crawl-workers",
			Usage:       "How many concurrent workers should dial and crawl peers.",
			EnvVars:     []string{"NEBULA_CRAWL_WORKER_COUNT"},
			DefaultText: strconv.Itoa(config.DefaultConfig.CrawlWorkerCount),
			Value:       config.DefaultConfig.CrawlWorkerCount,
		},
		&cli.IntFlag{
			Name:        "monitor-workers",
			Usage:       "How many concurrent workers should dial and ping peers.",
			EnvVars:     []string{"NEBULA_MONITOR_WORKER_COUNT"},
			DefaultText: strconv.Itoa(config.DefaultConfig.MonitorWorkerCount),
			Value:       config.DefaultConfig.MonitorWorkerCount,
		},
	},
}

// DaemonAction is the function that is called when running nebula daemon.
func DaemonAction(c *cli.Context) error {
	log.Infoln("Starting Nebula daemon...")

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
	if err = metrics.RegisterCrawlMetrics(); err != nil {
		return err
	}
	if err = metrics.RegisterMonitorMetrics(); err != nil {
		return err
	}
	if err = metrics.ListenAndServe(conf.PrometheusHost, conf.PrometheusPort); err != nil {
		return errors.Wrap(err, "initialize metrics")
	}
	_ = dbh

	return errors.New("this command isn't implemented yet :/")
}
