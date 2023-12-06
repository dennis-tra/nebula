package discv4

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/discvx"
)

type CrawlerConfig struct {
	DialTimeout  time.Duration
	AddrDialType config.AddrType
	LogErrors    bool
}

type Crawler struct {
	id           string
	cfg          *CrawlerConfig
	listener     *discvx.UDPv4
	crawledPeers int
	done         chan struct{}
}

var _ core.Worker[PeerInfo, core.CrawlResult[PeerInfo]] = (*Crawler)(nil)

func (c *Crawler) Work(ctx context.Context, task PeerInfo) (core.CrawlResult[PeerInfo], error) {
	logEntry := log.WithFields(log.Fields{
		"crawlerID":  c.id,
		"remoteID":   task.peerID.ShortString(),
		"crawlCount": c.crawledPeers,
	})
	logEntry.Debugln("Crawling peer")
	defer logEntry.Debugln("Crawled peer")

	crawlStart := time.Now()

	result := c.crawlDiscV4(ctx, task)

	cr := core.CrawlResult[PeerInfo]{
		CrawlerID:           c.id,
		Info:                task,
		CrawlStartTime:      crawlStart,
		RoutingTableFromAPI: false,
		RoutingTable:        result.RoutingTable,
		// Agent:               libp2pResult.Agent,
		// Protocols:           libp2pResult.Protocols,
		// ConnectError:        libp2pResult.ConnectError,
		// ConnectErrorStr:     libp2pResult.ConnectErrorStr,
		CrawlError:    result.Error,
		CrawlErrorStr: result.ErrorStr,
		CrawlEndTime:  time.Now(),
		// ConnectStartTime:    libp2pResult.ConnectStartTime,
		// ConnectEndTime:      libp2pResult.ConnectEndTime,
		// Properties: data,
		LogErrors: c.cfg.LogErrors,
	}

	// We've now crawled this peer, so increment
	c.crawledPeers++

	return cr, nil
}

type DiscV4Result struct {
	// The time we received the first successful response
	RespondedAt *time.Time

	// The updated ethereum node record
	ENR *enode.Node

	// The neighbors of the crawled peer
	RoutingTable *core.RoutingTable[PeerInfo]

	// The time the draining of bucket entries was finished
	DoneAt time.Time

	// The combined error of crawling the peer's buckets
	Error error

	// The above error mapped to a known string
	ErrorStr string
}

func (c *Crawler) crawlDiscV4(ctx context.Context, pi PeerInfo) DiscV4Result {
	// mutex to guard access to result and allNeighbors
	mu := sync.RWMutex{}

	// the final result struct
	result := DiscV4Result{}

	// all neighbors of pi. We're using a map to deduplicate.
	allNeighbors := map[string]PeerInfo{}

	// errorBits tracks at which CPL errors have occurred.
	// 0000 0000 0000 0000 - No error
	// 0000 0000 0000 0001 - An error has occurred at CPL 0
	// 1000 0000 0000 0001 - An error has occurred at CPL 0 and 15
	errorBits := atomic.NewUint32(0)

	enr, err := c.listener.RequestENR(pi.Node)
	if err != nil {
		result.ENR = pi.Node
	} else {
		result.ENR = enr
		now := time.Now()
		result.RespondedAt = &now
	}

	errg := errgroup.Group{}
	for i := 0; i <= 15; i++ { // 15 is maximum
		count := i // Copy value
		errg.Go(func() error {
			pubKey, err := discvx.GenRandomPublicKey(pi.Node.ID(), count)
			if err != nil {
				log.WithError(err).WithField("enr", pi.Node.String()).Warnln("Failed generating public key")
				errorBits.Add(1 << count)
				return fmt.Errorf("generating random public key with CPL %d: %w", count, err)
			}

			udpAddr := &net.UDPAddr{IP: pi.Node.IP(), Port: pi.Node.UDP()}

			var neighbors []*enode.Node
			for retry := 0; retry < 2; retry++ {
				neighbors, err = c.listener.FindNode(pi.Node.ID(), udpAddr, pubKey)
				if err == nil {
					break
				}

				errorBits.Add(1 << count)

				if errors.Is(err, discvx.ErrTimeout) {
					sleepDur := time.Second * time.Duration(3*(retry+1))
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(sleepDur): // may add jitter here
						continue
					}
				}

				errorBits.Add(1 << count)

				return fmt.Errorf("getting closest peer with CPL %d: %w", count, err)
			}

			mu.Lock()
			defer mu.Unlock()

			if result.RespondedAt == nil {
				now := time.Now()
				result.RespondedAt = &now
			}

			for _, n := range neighbors {
				npi, err := NewPeerInfo(n)
				if err != nil {
					log.WithError(err).Warnln("Failed parsing ethereum node neighbor")
					continue
				}
				allNeighbors[string(npi.peerID)] = npi
			}

			if err != nil {
				errorBits.Add(1 << count)
				return err
			}

			return nil
		})
	}

	// wait for go routines to finish
	err = errg.Wait()

	// track done timestamp and error
	result.DoneAt = time.Now()
	result.Error = err

	result.RoutingTable = &core.RoutingTable[PeerInfo]{
		PeerID:    pi.ID(),
		Neighbors: []PeerInfo{},
		ErrorBits: uint16(errorBits.Load()),
		Error:     err,
	}

	for _, n := range allNeighbors {
		result.RoutingTable.Neighbors = append(result.RoutingTable.Neighbors, n)
	}

	// if there was a connection error, parse it to a known one
	if result.Error != nil {
		result.ErrorStr = db.NetError(result.Error)
	}

	return result
}
