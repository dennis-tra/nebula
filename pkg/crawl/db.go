package crawl

import (
	"context"
	"fmt"
	"time"

	"github.com/dennis-tra/nebula-crawler/pkg/utils"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

// updateCrawl writes crawl statistics to the database
func (s *Scheduler) updateCrawl(ctx context.Context, crawlerCtx context.Context, success bool) error {
	log.Infoln("Persisting crawl result...")

	s.crawl.StartedAt = s.crawlStart
	s.crawl.FinishedAt = null.TimeFrom(time.Now())
	s.crawl.CrawledPeers = null.IntFrom(len(s.crawled))
	s.crawl.DialablePeers = null.IntFrom(len(s.crawled) - s.TotalErrors())
	s.crawl.UndialablePeers = null.IntFrom(s.TotalErrors())

	if success {
		s.crawl.State = models.CrawlStateSucceeded
	} else if errors.Is(crawlerCtx.Err(), context.Canceled) {
		s.crawl.State = models.CrawlStateCancelled
	} else {
		s.crawl.State = models.CrawlStateFailed
	}

	return s.dbc.UpdateCrawl(ctx, s.crawl)
}

// persistCrawlProperties writes crawl property statistics to the database.
func (s *Scheduler) persistCrawlProperties(ctx context.Context) error {
	log.Infoln("Persisting crawl properties...")

	// Extract full and core agent versions. Core agent versions are just strings like 0.8.0 or 0.5.0
	// The full agent versions have much more information e.g., /go-ipfs/0.4.21-dev/789dab3
	avFull := map[string]int{}
	for version, count := range s.agentVersion {
		avFull[version] += count
	}
	pps := map[string]map[string]int{
		"agent_version": avFull,
		"protocol":      s.protocols,
		"error":         s.errors,
	}

	return s.dbc.PersistCrawlProperties(ctx, s.crawl, pps)
}

// persistNeighbors fills the neighbors table with topology information
func (s *Scheduler) persistNeighbors() {
	if !s.config.PersistNeighbors {
		return
	}

	log.Infoln("Persisting neighbor information...")

	start := time.Now()
	neighborsCount := 0
	i := 0
	for p, routingTable := range s.routingTables {
		if i%100 == 0 && i > 0 {
			log.Infof("Persisted %d peers and their neighbors", i)
		}
		i++
		neighborsCount += len(routingTable.Neighbors)
		if err := s.dbc.PersistNeighbors(s.crawl, p, routingTable.ErrorBits, routingTable.PeerIDs()); err != nil {
			log.WithError(err).WithField("peerID", utils.FmtPeerID(p)).Warnln("Could not persist neighbors")
		}
	}
	log.WithFields(log.Fields{
		"duration":       time.Since(start),
		"avg":            fmt.Sprintf("%.2fms", time.Since(start).Seconds()/float64(len(s.routingTables))*1000),
		"peers":          len(s.routingTables),
		"totalNeighbors": neighborsCount,
	}).Infoln("Finished persisting neighbor information")
}
