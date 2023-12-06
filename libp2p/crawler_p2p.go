package libp2p

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/db/models"
	"github.com/dennis-tra/nebula-crawler/utils"
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

	// When have we established a successful connection
	ConnectEndTime time.Time

	// All connections that the remote peer claims to listen on
	// this can be different from the ones that we received from another peer
	// e.g., they could miss quic-v1 addresses if the reporting peer doesn't
	// know about that protocol.
	ListenAddrs []ma.Multiaddr

	// If the connection was closed immediately
	ConnClosedImmediately bool
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

			// check if we're actually connected
			if c.host.Network().Connectedness(pi.ID()) == network.NotConnected {
				// this is a weird behavior I was obesrving. Libp2p reports a
				// successful connection establishment but isn't connected right
				// after the call returned. This is not a big problem at this
				// point because fetchNeighbors will open the connection again.
				// This works more often than not but is still weird. At least
				// keep track of this issue - just in case.
				result.ConnClosedImmediately = true
			}

			// Fetch all neighbors
			result.RoutingTable, result.CrawlError = c.fetchNeighbors(ctx, pi.AddrInfo)
			if result.CrawlError != nil {
				result.CrawlErrorStr = db.NetError(result.CrawlError)
			}

			// wait for the Identify exchange to complete (no-op if already done)
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
	if len(pi.Addrs) == 0 {
		return fmt.Errorf("skipping node as it has no public IP address") // change knownErrs map if changing this msg
	}

	// init an exponential backoff
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = time.Second
	bo.MaxInterval = 10 * time.Second
	bo.MaxElapsedTime = time.Minute

	// keep track of retries for debug logging
	retry := 0

	for {
		logEntry := log.WithFields(log.Fields{
			"timeout":  c.cfg.DialTimeout.String(),
			"remoteID": pi.ID.String(),
			"retry":    retry,
			"maddrs":   pi.Addrs,
		})
		logEntry.Debugln("Connecting to peer", pi.ID.ShortString())

		timeoutCtx, cancel := context.WithTimeout(ctx, c.cfg.DialTimeout)
		err := c.host.Connect(timeoutCtx, pi)
		cancel()

		// yay, it worked! Or has it? The caller checks the connectedness again.
		if err == nil {
			return nil
		}

		switch true {
		case strings.Contains(err.Error(), db.ErrorStr[models.NetErrorConnectionRefused]):
			// Might be transient because the remote doesn't want us to connect. Try again!
		case strings.Contains(err.Error(), db.ErrorStr[models.NetErrorConnectionGated]):
			// Hints at a configuration issue and should not happen, but if it
			// does it could be transient. Try again anyway, but at least log a warning.
			logEntry.WithError(err).Warnln("Connection gated!")
		case strings.Contains(err.Error(), db.ErrorStr[models.NetErrorCantAssignRequestedAddress]):
			// Transient error due to local UDP issues. Try again!
		case strings.Contains(err.Error(), "dial backoff"):
			// should not happen because we disabled backoff checks with our
			// go-libp2p fork. Try again anyway, but at least log a warning.
			logEntry.WithError(err).Warnln("Dial backoff!")
		case strings.Contains(err.Error(), "RESOURCE_LIMIT_EXCEEDED (201)"): // thrown by a circuit relay
			// We already have too many open connections over a relay. Try again!
		default:
			logEntry.WithError(err).Debugln("Failed connecting to peer", pi.ID.ShortString())
			return err
		}

		sleepDur := bo.NextBackOff()
		if sleepDur == backoff.Stop {
			logEntry.WithError(err).Debugln("Exceeded retries connecting to peer", pi.ID.ShortString())
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleepDur):
			retry += 1
			continue
		}

	}
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
				log.WithError(err).WithField("enr", pi.ID.ShortString()).WithField("cpl", count).Warnln("Failed generating random peer ID")
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
					// a stream but then failed to do so. Try to reconnect as
					// the peer appears to be online

					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(sleepDur): // may add jitter here
						// fall through
					}

					ctx, cancel := context.WithTimeout(ctx, c.cfg.DialTimeout)
					if err := c.host.Connect(ctx, pi); err != nil {
						cancel()
						return err
					}
					cancel()

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

			if err != nil {
				errorBits.Add(1 << count)
				return err
			}

			return nil
		})
	}
	err = errg.Wait()

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
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
				if c.isIdentified(pi.ID) {
					cancel()
					return
				}
			}
		}()
	}

	wg.Wait()
}

// isIdentified returns true if the given peer.ID was successfully identified.
// Just because IdentifyWait returns doesn't mean the peer was identified.
func (c *Crawler) isIdentified(pid peer.ID) bool {
	agent, err := c.host.Peerstore().Get(pid, "AgentVersion")
	return err == nil && agent.(string) != ""
}
