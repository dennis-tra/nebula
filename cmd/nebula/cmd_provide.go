package main

import (
	"strconv"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/provide"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// ProvideCommand contains the provide sub-command configuration.
var ProvideCommand = &cli.Command{
	Name:   "provide",
	Usage:  "Starts a DHT measurement experiment by providing and requesting random content.",
	Action: ProvideAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "dry-run",
			Usage:   "Don't persist anything to a database (you don't need a running DB)",
			EnvVars: []string{"NEBULA_PROVIDE_DRY_RUN"},
		},
		&cli.BoolFlag{
			Name:        "routing-table",
			Usage:       "Whether or not Nebula should wait until the provider's routing table was refreshed",
			EnvVars:     []string{"NEBULA_PROVIDE_ROUTING_TABLE"},
			DefaultText: strconv.FormatBool(config.DefaultConfig.RefreshRoutingTable),
		},
	},
}

// ProvideAction is the function that is called when running `nebula resolve`.
func ProvideAction(c *cli.Context) error {
	log.Infoln("Starting Nebula DHT measurement...")

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
	if err = metrics.ListenAndServe(conf.PrometheusHost, conf.PrometheusPort); err != nil {
		return errors.Wrap(err, "initialize metrics")
	}

	s, err := provide.NewScheduler(conf, dbc)
	if err != nil {
		return errors.Wrap(err, "creating new scheduler")
	}

	go func() {
		// Nebula was asked to stop (e.g. SIGINT) -> tell the scheduler to stop
		<-c.Context.Done()
		s.Shutdown()
	}()

	return s.StartExperiment()
}
