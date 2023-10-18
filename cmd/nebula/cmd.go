package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
)

const (
	flagCategoryDatabase  = "Database Configuration:"
	flagCategoryDebugging = "Debugging Configuration:"
	flagCategoryCache     = "Cache Configuration:"
)

var (
	// RawVersion and build tag of the
	// Nebula command line tool.
	RawVersion = "dev"
)

var rootConfig = &config.Root{
	RawVersion:             RawVersion,
	Debug:                  false,
	LogLevel:               4,
	DialTimeout:            time.Minute,
	TelemetryHost:          "0.0.0.0",
	TelemetryPort:          6666,
	DatabaseHost:           "localhost",
	DatabasePort:           5432,
	DatabaseName:           "nebula",
	DatabasePassword:       "password",
	DatabaseUser:           "nebula",
	DatabaseSSLMode:        "disable",
	AgentVersionsCacheSize: 200,
	ProtocolsCacheSize:     100,
	ProtocolsSetCacheSize:  200,
}

func main() {

	app := &cli.App{
		Name:      "nebula",
		Usage:     "A libp2p DHT crawler, monitor, and measurement tool that exposes timely information about DHT networks.",
		UsageText: "nebula [global options] command [command options] [arguments...]",
		Authors: []*cli.Author{
			{
				Name:  "Dennis Trautwein",
				Email: "nebula@dtrautwein.eu",
			},
		},
		Version: rootConfig.Version(),
		Before:  Before,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "Set this flag to enable debug logging",
				EnvVars:     []string{"NEBULA_DEBUG"},
				Value:       rootConfig.Debug,
				Destination: &rootConfig.Debug,
				Category:    flagCategoryDebugging,
			},
			&cli.IntFlag{
				Name:        "log-level",
				Usage:       "Set this flag to a value from 0 (least verbose) to 6 (most verbose). Overrides the --debug flag",
				EnvVars:     []string{"NEBULA_LOG_LEVEL"},
				Value:       rootConfig.LogLevel,
				Destination: &rootConfig.LogLevel,
				Category:    flagCategoryDebugging,
			},
			&cli.StringFlag{
				Name:        "telemetry-host",
				Usage:       "To which network address should the telemetry (prometheus, pprof) server bind",
				EnvVars:     []string{"NEBULA_TELEMETRY_HOST"},
				Value:       rootConfig.TelemetryHost,
				Destination: &rootConfig.TelemetryHost,
				Category:    flagCategoryDebugging,
			},
			&cli.IntFlag{
				Name:        "telemetry-port",
				Usage:       "On which port should the telemetry (prometheus, pprof) server listen",
				EnvVars:     []string{"NEBULA_TELEMETRY_PORT"},
				Value:       rootConfig.TelemetryPort,
				Destination: &rootConfig.TelemetryPort,
				Category:    flagCategoryDebugging,
			},
			&cli.StringFlag{
				Name:        "db-host",
				Usage:       "On which host address can nebula reach the database",
				EnvVars:     []string{"NEBULA_DATABASE_HOST"},
				Value:       rootConfig.DatabaseHost,
				Destination: &rootConfig.DatabaseHost,
				Category:    flagCategoryDatabase,
			},
			&cli.IntFlag{
				Name:        "db-port",
				Usage:       "On which port can nebula reach the database",
				EnvVars:     []string{"NEBULA_DATABASE_PORT"},
				Value:       rootConfig.DatabasePort,
				Destination: &rootConfig.DatabasePort,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "db-name",
				Usage:       "The name of the database to use",
				EnvVars:     []string{"NEBULA_DATABASE_NAME"},
				Value:       rootConfig.DatabaseName,
				Destination: &rootConfig.DatabaseName,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "db-password",
				Usage:       "The password for the database to use",
				EnvVars:     []string{"NEBULA_DATABASE_PASSWORD"},
				Value:       rootConfig.DatabasePassword,
				Destination: &rootConfig.DatabasePassword,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "db-user",
				Usage:       "The user with which to access the database to use",
				EnvVars:     []string{"NEBULA_DATABASE_USER"},
				Value:       rootConfig.DatabaseUser,
				Destination: &rootConfig.DatabaseUser,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "db-sslmode",
				Usage:       "The sslmode to use when connecting the the database",
				EnvVars:     []string{"NEBULA_DATABASE_SSL_MODE"},
				Value:       rootConfig.DatabaseSSLMode,
				Destination: &rootConfig.DatabaseSSLMode,
				Category:    flagCategoryDatabase,
			},
			&cli.IntFlag{
				Name:        "agent-versions-cache-size",
				Usage:       "The cache size to hold agent versions in memory",
				EnvVars:     []string{"NEBULA_AGENT_VERSIONS_CACHE_SIZE"},
				Value:       rootConfig.AgentVersionsCacheSize,
				Destination: &rootConfig.AgentVersionsCacheSize,
				Category:    flagCategoryCache,
			},
			&cli.IntFlag{
				Name:        "protocols-cache-size",
				Usage:       "The cache size to hold protocols in memory",
				EnvVars:     []string{"NEBULA_PROTOCOLS_CACHE_SIZE"},
				Value:       rootConfig.ProtocolsCacheSize,
				Destination: &rootConfig.ProtocolsCacheSize,
				Category:    flagCategoryCache,
			},
			&cli.IntFlag{
				Name:        "protocols-set-cache-size",
				Usage:       "The cache size to hold sets of protocols in memory",
				EnvVars:     []string{"NEBULA_PROTOCOLS_SET_CACHE_SIZE"},
				Value:       rootConfig.ProtocolsSetCacheSize,
				Destination: &rootConfig.ProtocolsSetCacheSize,
				Category:    flagCategoryCache,
			},
		},
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			CrawlCommand,
			MonitorCommand,
			ResolveCommand,
			EthCommand,
		},
	}

	sigs := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	go func() {
		sig := <-sigs
		log.Printf("Received %s signal - Stopping...\n", sig.String())
		signal.Stop(sigs)
		cancel()
	}()

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Errorf("error: %v\n", err)
		os.Exit(1)
	}
}

// Before is executed before any subcommands are run, but after the context is ready
// If a non-nil error is returned, no subcommands are run.
func Before(c *cli.Context) error {
	if rootConfig.Debug {
		log.SetLevel(log.DebugLevel)
	}

	if c.IsSet("log-level") {
		ll := c.Int("log-level")
		log.SetLevel(log.Level(ll))
		if ll == int(log.TraceLevel) {
			boil.DebugMode = true
		}
	}

	// Start prometheus metrics endpoint
	go metrics.ListenAndServe(rootConfig.TelemetryHost, rootConfig.TelemetryPort)

	return nil
}
