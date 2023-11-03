package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
)

// NetworksCommand contains the networks sub-command configuration.
var NetworksCommand = &cli.Command{
	Name:      "networks",
	Usage:     "Lists the supported networks and their default configuration",
	ArgsUsage: "NETWORK (optional)",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "bootstrappers",
			Usage:   "Include information about bootstrappers in the output (true if network argument is given)",
			EnvVars: []string{"NEBULA_NETWORKS_BOOTSTRAPPERS"},
		},
		&cli.BoolFlag{
			Name:    "protocols",
			Usage:   "Include information about protocols in the output (true if network argument is given)",
			EnvVars: []string{"NEBULA_NETWORKS_PROTOCOLS"},
		},
	},
	Action: func(c *cli.Context) error {
		networks := config.Networks()

		networkArg := c.Args().First()
		if networkArg != "" {
			networks = []config.Network{config.Network(networkArg)}
		}

		log.Infoln("Supported Networks")
		for _, network := range networks {
			log.Infof("- %s", network)
			bootstrappers, protocols, err := config.ConfigureNetwork(string(network))
			if err != nil {
				return err
			}

			if c.Bool("protocols") || networkArg != "" {
				log.Infof("  Protocols: %s", protocols)
			}

			if c.Bool("bootstrappers") || networkArg != "" {
				log.Infof("  Bootstrapperes: %s", network)
				for i, b := range bootstrappers.Value() {
					log.Infof("    [%d] %s", i, b)
				}
			}
		}
		if networkArg == "" {
			log.Infoln("Specify a network with `nebula networks $NETWORK` to see additional information")
		}
		return nil
	},
}
