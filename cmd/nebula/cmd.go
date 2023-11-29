package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/dennis-tra/nebula-crawler/tele"
)

const (
	flagCategoryDatabase  = "Database Configuration:"
	flagCategoryDebugging = "Debugging Configuration:"
	flagCategoryCache     = "Cache Configuration:"
	flagCategoryNetwork   = "Network Specific Configuration:"
)

// RawVersion and build tag of the Nebula command line tool.
var RawVersion = "dev"

var rootConfig = &config.Root{
	RawVersion:      RawVersion,
	Debug:           false,
	LogLevel:        4,
	LogFormat:       "text",
	LogDisableColor: false,
	DialTimeout:     10 * time.Second,
	MetricsHost:     "0.0.0.0",
	MetricsPort:     6666,
	TracesHost:      "", // disabled
	TracesPort:      0,  // disabled
	Database: &config.Database{
		DryRun:                 false,
		JSONOut:                "",
		DatabaseHost:           "localhost",
		DatabasePort:           5432,
		DatabaseName:           "nebula",
		DatabasePassword:       "password",
		DatabaseUser:           "nebula",
		DatabaseSSLMode:        "disable",
		AgentVersionsCacheSize: 200,
		ProtocolsCacheSize:     100,
		ProtocolsSetCacheSize:  200,
	},
}

func main() {
	app := &cli.App{
		Name:      "nebula",
		Usage:     "A DHT crawler, monitor, and measurement tool that exposes timely information about DHT networks.",
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
				Name:        "log-format",
				Usage:       "Define the formatting of the log output (values: text, json)",
				EnvVars:     []string{"NEBULA_LOG_FORMAT"},
				Value:       rootConfig.LogFormat,
				Destination: &rootConfig.LogFormat,
				Category:    flagCategoryDebugging,
			},
			&cli.BoolFlag{
				Name:        "log-disable-color",
				Usage:       "Whether to have colorized log output (only text log format)",
				EnvVars:     []string{"NEBULA_LOG_COLOR"},
				Value:       rootConfig.LogDisableColor,
				Destination: &rootConfig.LogDisableColor,
				Category:    flagCategoryDebugging,
			},
			&cli.DurationFlag{
				Name:        "dial-timeout",
				Usage:       "Global timeout when trying to connect to or dial another peer in the network.",
				EnvVars:     []string{"NEBULA_DIAL_TIMEOUT"},
				Value:       rootConfig.DialTimeout,
				Destination: &rootConfig.DialTimeout,
			},
			&cli.StringFlag{
				Name:        "metrics-host",
				Usage:       "To which network address should the metrics (prometheus, pprof) server bind",
				EnvVars:     []string{"NEBULA_METRICS_HOST"},
				Value:       rootConfig.MetricsHost,
				Destination: &rootConfig.MetricsHost,
				Category:    flagCategoryDebugging,
			},
			&cli.IntFlag{
				Name:        "metrics-port",
				Usage:       "On which port should the metrics (prometheus, pprof) server listen",
				EnvVars:     []string{"NEBULA_METRICS_PORT"},
				Value:       rootConfig.MetricsPort,
				Destination: &rootConfig.MetricsPort,
				Category:    flagCategoryDebugging,
			},
			&cli.StringFlag{
				Name:        "traces-host",
				Usage:       "To which host should traces be sent",
				EnvVars:     []string{"NEBULA_TRACES_HOST"},
				Value:       rootConfig.TracesHost,
				Destination: &rootConfig.TracesHost,
				Category:    flagCategoryDebugging,
			},
			&cli.IntFlag{
				Name:        "traces-port",
				Usage:       "On which port does the trace collector listen",
				EnvVars:     []string{"NEBULA_TRACES_PORT"},
				Value:       rootConfig.TracesPort,
				Destination: &rootConfig.TracesPort,
				Category:    flagCategoryDebugging,
			},
			&cli.BoolFlag{
				Name:        "dry-run",
				Usage:       "Don't write anything to disk",
				EnvVars:     []string{"NEBULA_DRY_RUN", "NEBULA_CRAWL_DRY_RUN" /*<-legacy*/},
				Value:       rootConfig.Database.DryRun,
				Destination: &rootConfig.Database.DryRun,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "json-out",
				Usage:       "If set, stores results as JSON documents at `DIR` (takes precedence over database settings).",
				EnvVars:     []string{"NEBULA_JSON_OUT", "NEBULA_CRAWL_JSON_OUT" /*<-legacy*/},
				Value:       rootConfig.Database.JSONOut,
				Destination: &rootConfig.Database.JSONOut,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "db-host",
				Usage:       "On which host address can nebula reach the database",
				EnvVars:     []string{"NEBULA_DATABASE_HOST"},
				Value:       rootConfig.Database.DatabaseHost,
				Destination: &rootConfig.Database.DatabaseHost,
				Category:    flagCategoryDatabase,
			},
			&cli.IntFlag{
				Name:        "db-port",
				Usage:       "On which port can nebula reach the database",
				EnvVars:     []string{"NEBULA_DATABASE_PORT"},
				Value:       rootConfig.Database.DatabasePort,
				Destination: &rootConfig.Database.DatabasePort,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "db-name",
				Usage:       "The name of the database to use",
				EnvVars:     []string{"NEBULA_DATABASE_NAME"},
				Value:       rootConfig.Database.DatabaseName,
				Destination: &rootConfig.Database.DatabaseName,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "db-password",
				Usage:       "The password for the database to use",
				EnvVars:     []string{"NEBULA_DATABASE_PASSWORD"},
				Value:       rootConfig.Database.DatabasePassword,
				Destination: &rootConfig.Database.DatabasePassword,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "db-user",
				Usage:       "The user with which to access the database to use",
				EnvVars:     []string{"NEBULA_DATABASE_USER"},
				Value:       rootConfig.Database.DatabaseUser,
				Destination: &rootConfig.Database.DatabaseUser,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "db-sslmode",
				Usage:       "The sslmode to use when connecting the the database",
				EnvVars:     []string{"NEBULA_DATABASE_SSL_MODE"},
				Value:       rootConfig.Database.DatabaseSSLMode,
				Destination: &rootConfig.Database.DatabaseSSLMode,
				Category:    flagCategoryDatabase,
			},
			&cli.IntFlag{
				Name:        "agent-versions-cache-size",
				Usage:       "The cache size to hold agent versions in memory",
				EnvVars:     []string{"NEBULA_AGENT_VERSIONS_CACHE_SIZE"},
				Value:       rootConfig.Database.AgentVersionsCacheSize,
				Destination: &rootConfig.Database.AgentVersionsCacheSize,
				Category:    flagCategoryCache,
			},
			&cli.IntFlag{
				Name:        "protocols-cache-size",
				Usage:       "The cache size to hold protocols in memory",
				EnvVars:     []string{"NEBULA_PROTOCOLS_CACHE_SIZE"},
				Value:       rootConfig.Database.ProtocolsCacheSize,
				Destination: &rootConfig.Database.ProtocolsCacheSize,
				Category:    flagCategoryCache,
			},
			&cli.IntFlag{
				Name:        "protocols-set-cache-size",
				Usage:       "The cache size to hold sets of protocols in memory",
				EnvVars:     []string{"NEBULA_PROTOCOLS_SET_CACHE_SIZE"},
				Value:       rootConfig.Database.ProtocolsSetCacheSize,
				Destination: &rootConfig.Database.ProtocolsSetCacheSize,
				Category:    flagCategoryCache,
			},
		},
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			CrawlCommand,
			MonitorCommand,
			ResolveCommand,
			NetworksCommand,
			HealthCommand,
		},
	}

	sigs := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	go func() {
		sig := <-sigs
		log.Infof("Received %s signal - Stopping...\n", sig.String())
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

	switch strings.ToLower(c.String("log-format")) {
	case "text":
		log.SetFormatter(&log.TextFormatter{
			DisableColors: c.Bool("log-disable-color"),
		})
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		return fmt.Errorf("unknown log format: %q", c.String("log-format"))
	}

	meterProvider, err := tele.NewMeterProvider()
	if err != nil {
		return fmt.Errorf("new meter provider: %w", err)
	}
	rootConfig.MeterProvider = meterProvider
	rootConfig.Database.MeterProvider = meterProvider

	tracerProvider, err := tele.NewTracerProvider(c.Context, rootConfig.TracesHost, rootConfig.TracesPort)
	if err != nil {
		return fmt.Errorf("new tracer provider: %w", err)
	}
	rootConfig.TracerProvider = tracerProvider
	rootConfig.Database.TracerProvider = tracerProvider

	// Start prometheus metrics endpoint (but only if it's not the health command)
	if c.Args().Get(0) != HealthCommand.Name {
		go tele.ListenAndServe(rootConfig.MetricsHost, rootConfig.MetricsPort)
	}

	return nil
}
