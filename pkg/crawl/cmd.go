package crawl

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// Command contains the receive sub-command configuration.
var Command = &cli.Command{
	Name:   "crawl",
	Action: Action,
}

// Action is the function that is called when running pcp receive.
func Action(c *cli.Context) error {
	fmt.Println("Crawling")
	return nil
}
