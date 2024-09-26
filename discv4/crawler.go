package discv4

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/db/models"
	"github.com/dennis-tra/nebula-crawler/devp2p"
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
	client       *devp2p.Client
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

	discv4ResultCh := c.crawlDiscV4(ctx, task)
	devp2pResultCh := c.crawlDevp2p(ctx, task)

	discv4Result := <-discv4ResultCh
	devp2pResult := <-devp2pResultCh

	properties := map[string]any{}

	// keep track of all unknown connection errors
	if devp2pResult.ConnectErrorStr == models.NetErrorUnknown && devp2pResult.ConnectError != nil {
		properties["connect_error"] = devp2pResult.ConnectError.Error()
	}

	// keep track of all unknown crawl errors
	if discv4Result.ErrorStr == models.NetErrorUnknown && discv4Result.Error != nil {
		properties["crawl_error"] = discv4Result.Error.Error()
	}

	data, err := json.Marshal(properties)
	if err != nil {
		log.WithError(err).WithField("properties", properties).Warnln("Could not marshal peer properties")
	}

	cr := core.CrawlResult[PeerInfo]{
		CrawlerID:           c.id,
		Info:                task,
		CrawlStartTime:      crawlStart,
		RoutingTableFromAPI: false,
		RoutingTable:        discv4Result.RoutingTable,
		Agent:               devp2pResult.Agent,
		Protocols:           devp2pResult.Protocols,
		ConnectError:        devp2pResult.ConnectError,
		ConnectErrorStr:     devp2pResult.ConnectErrorStr,
		CrawlError:          discv4Result.Error,
		CrawlErrorStr:       discv4Result.ErrorStr,
		CrawlEndTime:        time.Now(),
		ConnectStartTime:    devp2pResult.ConnectStartTime,
		ConnectEndTime:      devp2pResult.ConnectEndTime,
		Properties:          data,
		LogErrors:           c.cfg.LogErrors,
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

func (c *Crawler) crawlDiscV4(ctx context.Context, pi PeerInfo) <-chan DiscV4Result {
	resultCh := make(chan DiscV4Result)

	go func() {
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

				ipAddr, ok := netip.AddrFromSlice(pi.Node.IP())
				if !ok {
					return fmt.Errorf("failed to convert ip to netip.Addr")
				}
				udpAddr := netip.AddrPortFrom(ipAddr, uint16(pi.Node.UDP()))

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

		// send the result back and close channel
		select {
		case resultCh <- result:
		case <-ctx.Done():
		}

		close(resultCh)
	}()

	return resultCh
}

type Devp2pResult struct {
	ConnectStartTime time.Time
	ConnectEndTime   time.Time
	ConnectError     error
	ConnectErrorStr  string
	Agent            string
	Protocols        []string
}

func (c *Crawler) crawlDevp2p(ctx context.Context, pi PeerInfo) <-chan Devp2pResult {
	resultCh := make(chan Devp2pResult)
	go func() {
		// the final result struct
		result := Devp2pResult{}

		addrInfo := peer.AddrInfo{
			ID:    pi.ID(),
			Addrs: pi.Addrs(),
		}

		result.ConnectStartTime = time.Now()
		for retry := 0; retry < 3; retry++ {
			result.ConnectError = c.client.Connect(ctx, addrInfo)
			if result.ConnectError == nil {
				break
			}

			if strings.Contains(result.ConnectError.Error(), "handshake failed: EOF") {
				time.Sleep(time.Second)
				continue
			}
		}
		result.ConnectEndTime = time.Now()

		if result.ConnectError == nil {
			resp, err := c.client.Identify(pi.ID())
			if err == nil {
				result.Agent = resp.Name
				protocols := make([]string, len(resp.Caps))
				for i, c := range resp.Caps {
					protocols[i] = "/" + c.String()
				}
				result.Protocols = protocols
			} else {
				log.WithError(err).Debugln("Could not identify peer")
			}
		}
		// if there was a connection error, parse it to a known one
		if result.ConnectError != nil {
			result.ConnectErrorStr = db.NetError(result.ConnectError)
		}

		// Free connection resources
		if err := c.client.CloseConn(pi.ID()); err != nil {
			log.WithError(err).WithField("remoteID", pi.ID().ShortString()).Warnln("Could not close connection to peer")
		}

		// send the result back and close channel
		select {
		case resultCh <- result:
		case <-ctx.Done():
		}

		close(resultCh)
	}()
	return resultCh
}
