package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// ProvideCommand contains the provide sub-command configuration.
var ProvideCommand = &cli.Command{
	Name:   "provide",
	Usage:  "Starts a DHT measurement experiment by providing and requesting random content.",
	Action: ProvideAction,
}

// ProvideAction is the function that is called when running `nebula resolve`.
func ProvideAction(c *cli.Context) error {
	log.Infoln("Starting Nebula DHT measurement...")
	return nil
}
