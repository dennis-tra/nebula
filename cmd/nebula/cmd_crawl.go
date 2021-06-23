package main

import (
	"fmt"
	"os"

	"github.com/dennis-tra/nebula-crawler/pkg/db"

	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/crawl"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
)

// CrawlCommand contains the receive sub-command configuration.
var CrawlCommand = &cli.Command{
	Name:   "crawl",
	Action: CrawlAction,
}

// CrawlAction is the function that is called when running pcp receive.
func CrawlAction(c *cli.Context) error {
	log.Infoln("Crawling")
	log.SetLevel(log.DebugLevel)

	dbc, err := db.NewClient()
	if err != nil {
		return errors.Wrap(err, "initialize db")
	}

	//err = metrics.RegisterDB(dbc)
	//if err != nil {
	//	return errors.Wrap(err, "metrics register db")
	//}
	metrics.Serve()

	c, err = config.FillContext(c)
	if err != nil {
		return errors.Wrap(err, "failed to load crawler configuration")
	}
	mmm, err := ma.NewMultiaddr("/ip4/159.69.43.228/tcp/4001/p2p/QmSKVUFAyCddg2wDUdZVCfvqG5YCwwJTWY1HRmorebXcKG")
	if err != nil {
		return err
	}
	pi, err := peer.AddrInfoFromP2pAddr(mmm)
	if err != nil {
		return err
	}

	o, _ := crawl.NewOrchestrator(c.Context, dbc)
	go o.CrawlNetwork([]peer.AddrInfo{*pi})

	select {
	case <-c.Context.Done():
		o.Shutdown()
	case <-o.SigDone():
	}

	f, _ := os.Create("errors.txt")
	o.Errors.Range(func(errorStr, value interface{}) bool {
		fmt.Fprintf(f, "%s\n", errorStr)
		return true
	})
	f.Close()

	return nil
}
