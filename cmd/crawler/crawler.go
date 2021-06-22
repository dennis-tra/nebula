package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dennis-tra/go-ipfs-crawler/pkg/crawl"
	"github.com/urfave/cli/v2"
)

var (
	// RawVersion and build tag of the
	// PCP command line tool. This is
	// replaced on build via e.g.:
	// -ldflags "-X main.RawVersion=${VERSION}"
	RawVersion  = "dev"
	ShortCommit = "5f3759df" // quake
)

func main() {
	// ShortCommit version tag
	verTag := fmt.Sprintf("v%s+%s", RawVersion, ShortCommit)

	app := &cli.App{
		Name: "pcp",
		Authors: []*cli.Author{
			{
				Name:  "Dennis Trautwein",
				Email: "ipfs-crawler@dtrautwein.eu",
			},
		},
		Version:              verTag,
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			crawl.Command,
		},
	}

	sigs := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	go func() {
		<-sigs
		log.Println("Stopping...")
		signal.Stop(sigs)
		cancel()
	}()

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
