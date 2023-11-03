package libp2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	kbucket "github.com/libp2p/go-libp2p-kbucket"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/dennis-tra/nebula-crawler/pkg/core"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
)

type P2PResult struct {
	RoutingTable *core.RoutingTable[PeerInfo]

	// The agent version of the crawled peer
	Agent string

	// The protocols the peer supports
	Protocols []string

	// Any error that has occurred when connecting to the peer
	ConnectError error

	// The above error transferred to a known error
	ConnectErrorStr string

	// Any error that has occurred during fetching neighbor information
	CrawlError error

	// The above error transferred to a known error
	CrawlErrorStr string

	// When was the connection attempt made
	ConnectStartTime time.Time

	// As it can take some time to handle the result we track the timestamp explicitly
	ConnectEndTime time.Time

	// All connections that the remote peer claims to listen on
	// this can be different from the ones that we received from another peer
	// e.g., they could miss quic-v1 addresses if the reporting peer doesn't
	// know about that protocol.
	ListenAddrs []ma.Multiaddr
}

func (c *Crawler) crawlP2P(ctx context.Context, pi PeerInfo) <-chan P2PResult {
	resultCh := make(chan P2PResult)

	go func() {
		result := P2PResult{
			RoutingTable: &core.RoutingTable[PeerInfo]{PeerID: pi.ID()},
		}

		result.ConnectStartTime = time.Now()
		result.ConnectError = c.connect(ctx, pi.AddrInfo) // use filtered addr list
		result.ConnectEndTime = time.Now()

		// If we could successfully connect to the peer we actually crawl it.
		if result.ConnectError == nil {

			// Fetch all neighbors
			result.RoutingTable, result.CrawlError = c.fetchNeighbors(ctx, pi.AddrInfo)
			if result.CrawlError != nil {
				result.CrawlErrorStr = db.NetError(result.CrawlError)
			}

			// wait for the Identify exchange to complete
			c.identifyWait(ctx, pi.AddrInfo)

			// Extract information from peer store
			ps := c.host.Peerstore()

			// Extract agent
			if agent, err := ps.Get(pi.ID(), "AgentVersion"); err == nil {
				result.Agent = agent.(string)
			}

			// Extract protocols
			if protocols, err := ps.GetProtocols(pi.ID()); err == nil {
				result.Protocols = make([]string, len(protocols))
				for i := range protocols {
					result.Protocols[i] = string(protocols[i])
				}
			}

			// Extract listen addresses
			result.ListenAddrs = ps.Addrs(pi.ID())
		}

		// if there was a connection error, parse it to a known one
		if result.ConnectError != nil {
			result.ConnectErrorStr = db.NetError(result.ConnectError)
		}

		// Free connection resources
		if err := c.host.Network().ClosePeer(pi.ID()); err != nil {
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

// connect establishes a connection to the given peer. It also handles metric capturing.
func (c *Crawler) connect(ctx context.Context, pi peer.AddrInfo) error {
	metrics.VisitCount.With(metrics.CrawlLabel).Inc()

	if len(pi.Addrs) == 0 {
		metrics.VisitErrorsCount.With(metrics.CrawlLabel).Inc()
		return fmt.Errorf("skipping node as it has no public IP address") // change knownErrs map if changing this msg
	}

	ctx, cancel := context.WithTimeout(ctx, c.cfg.DialTimeout)
	defer cancel()

	if err := c.host.Connect(ctx, pi); err != nil {
		metrics.VisitErrorsCount.With(metrics.CrawlLabel).Inc()
		return err
	}

	return nil
}

// fetchNeighbors sends RPC messages to the given peer and asks for its closest peers to an artificial set
// of 15 random peer IDs with increasing common prefix lengths (CPL).
func (c *Crawler) fetchNeighbors(ctx context.Context, pi peer.AddrInfo) (*core.RoutingTable[PeerInfo], error) {
	rt, err := kbucket.NewRoutingTable(20, kbucket.ConvertPeerID(pi.ID), time.Hour, nil, time.Hour, nil)
	if err != nil {
		return nil, err
	}

	allNeighborsLk := sync.RWMutex{}
	allNeighbors := map[peer.ID]peer.AddrInfo{}

	// errorBits tracks at which CPL errors have occurred.
	// 0000 0000 0000 0000 - No error
	// 0000 0000 0000 0001 - An error has occurred at CPL 0
	// 1000 0000 0000 0001 - An error has occurred at CPL 0 and 15
	errorBits := atomic.NewUint32(0)

	errg := errgroup.Group{}
	for i := uint(0); i <= 15; i++ { // 15 is maximum
		count := i // Copy value
		errg.Go(func() error {
			// Generate a peer with the given common prefix length
			rpi, err := rt.GenRandPeerID(count)
			if err != nil {
				errorBits.Add(1 << count)
				return fmt.Errorf("generating random peer ID with CPL %d: %w", count, err)
			}

			var neighbors []*peer.AddrInfo
			for retry := 0; retry < 2; retry++ {
				neighbors, err = c.pm.GetClosestPeers(ctx, pi.ID, rpi)
				if err == nil {
					break
				}

				sleepDur := time.Second * time.Duration(5*(retry+1))

				if utils.IsResourceLimitExceeded(err) {
					// the other node has indicated that it's out of resources. Wait a bit and try again.
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(sleepDur): // may add jitter here
						continue
					}
				}

				// This error happens in: https://github.com/libp2p/go-libp2p/blob/4e2a16dd3f4f980bf9429572b3d2aed885594ec4/p2p/host/basic/basic_host.go#L645
				if err.Error() == "connection failed" {
					// This means we were connected to the peer, tried to open
					// a stream but then failed to do so.

					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(sleepDur): // may add jitter here
						// fall through
					}

					ctx, cancel := context.WithTimeout(ctx, c.cfg.DialTimeout)
					defer cancel()

					if err := c.host.Connect(ctx, pi); err != nil {
						metrics.VisitErrorsCount.With(metrics.CrawlLabel).Inc()
						return err
					}

					continue
				}

				errorBits.Add(1 << count)

				return fmt.Errorf("getting closest peer with CPL %d: %w", count, err)
			}

			allNeighborsLk.Lock()
			defer allNeighborsLk.Unlock()
			for _, n := range neighbors {
				allNeighbors[n.ID] = *n
			}
			return nil
		})
	}
	err = errg.Wait()
	metrics.FetchedNeighborsCount.Observe(float64(len(allNeighbors)))

	routingTable := &core.RoutingTable[PeerInfo]{
		PeerID:    pi.ID,
		Neighbors: []PeerInfo{},
		ErrorBits: uint16(errorBits.Load()),
		Error:     err,
	}

	for _, n := range allNeighbors {
		routingTable.Neighbors = append(routingTable.Neighbors, PeerInfo{AddrInfo: n})
	}

	return routingTable, err
}

// identifyWait waits until any connection to a peer passed the Identify
// exchange successfully or all identification attempts have failed.
// The call to IdentifyWait returns immediately if the connection was
// identified in the past. We detect a successful identification if an
// AgentVersion is stored in the peer store
func (c *Crawler) identifyWait(ctx context.Context, pi peer.AddrInfo) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for _, conn := range c.host.Network().ConnsToPeer(pi.ID) {
		conn := conn

		wg.Add(1)
		go func() {
			defer wg.Done()

			select {
			case <-timeoutCtx.Done():
			case <-c.host.IDService().IdentifyWait(conn):

				// check if identification was successful by looking for
				// the AgentVersion key. If it exists, we cancel the
				// identification of the remaining connections.
				agent, err := c.host.Peerstore().Get(pi.ID, "AgentVersion")
				if err == nil && agent.(string) != "" {
					cancel()
					return
				}
			}
		}()
	}

	wg.Wait()
}
