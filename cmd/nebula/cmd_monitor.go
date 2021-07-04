package main

import (
	"strconv"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/monitor"
)

// MonitorCommand contains the receive sub-command configuration.
var MonitorCommand = &cli.Command{
	Name:   "monitor",
	Usage:  "Monitors the network by periodically dialing and pinging previously crawled peers.",
	Action: MonitorAction,
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:        "workers",
			Usage:       "How many concurrent workers should ping peers.",
			EnvVars:     []string{"NEBULA_MONITOR_WORKER_COUNT"},
			DefaultText: strconv.Itoa(config.DefaultConfig.MonitorWorkerCount),
			Value:       config.DefaultConfig.MonitorWorkerCount,
		},
	},
}

// MonitorAction is the function that is called when running `nebula monitor`.
func MonitorAction(c *cli.Context) error {
	log.Infoln("Starting Nebula monitor...")

	// Load configuration file
	ctx, conf, err := config.FillContext(c)
	if err != nil {
		return errors.Wrap(err, "filling context with configuration")
	}
	c.Context = ctx

	// Acquire database handle
	dbh, err := db.Open(c.Context)
	if err != nil {
		return err
	}

	// Start prometheus metrics endpoint
	if err = metrics.RegisterListenAndServe(conf.PrometheusHost, conf.PrometheusPort, "monitor"); err != nil {
		return errors.Wrap(err, "initialize metrics")
	}

	// Initialize orchestrator that handles crawling the network.
	m, _ := monitor.NewMonitor(c.Context, dbh)
	go m.StartMonitoring()

	select {
	case <-c.Context.Done():
		// Nebula was asked to stop (e.g. SIGINT) -> tell the monitor to stop
		m.Shutdown()
	case <-m.SigDone():
		// monitor finished autonomously
	}

	return nil
}
