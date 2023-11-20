package main

import (
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"
)

// HealthCommand is used as a health check by, e.g., the docker runtime to test
// whether Nebula is running fine. It just calls the /health endpoint.
var HealthCommand = &cli.Command{
	Name:   "health",
	Usage:  "Probes Nebula's health endpoint",
	Hidden: true,
	Action: HealthAction,
}

// HealthAction is the function that is called when running `nebula resolve`.
func HealthAction(c *cli.Context) error {
	endpoint := fmt.Sprintf("http://%s:%d/health", rootConfig.TelemetryHost, rootConfig.TelemetryPort)
	resp, err := http.Get(endpoint)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("unhealthy")
}
