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
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	pgmodels "github.com/dennis-tra/nebula-crawler/db/models/pg"
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

	// the transport of a successful connection
	Transport string
}

func (c *Crawler) crawlP2P(ctx context.Context, pi PeerInfo) <-chan P2PResult {
	resultCh := make(chan P2PResult)

	go func() {
		result := P2PResult{
			RoutingTable: &core.RoutingTable[PeerInfo]{PeerID: pi.ID()},
		}

		var conn network.Conn
		result.ConnectStartTime = time.Now()
		conn, result.ConnectError = c.connect(ctx, pi.AddrInfo) // use filtered addr list
		result.ConnectEndTime = time.Now()

		// If we could successfully connect to the peer we actually crawl it.
		if result.ConnectError == nil {

			// keep track of the transport of the open connection
			result.Transport = conn.ConnState().Transport

			// Fetch all neighbors
			result.RoutingTable, result.CrawlError = c.drainBuckets(ctx, pi.AddrInfo)
			if result.CrawlError != nil {
				result.CrawlErrorStr = db.NetError(result.CrawlError)
			}

			// wait for the Identify exchange to complete (no-op if already done)
			// the internal timeout is set to 30 s. When crawling we only allow 5s.
			timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			select {
			case <-timeoutCtx.Done():
				// identification timed out.
			case <-c.host.IDService().IdentifyWait(conn):
				// identification may have succeeded.
			}

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
func (c *Crawler) connect(ctx context.Context, pi peer.AddrInfo) (network.Conn, error) {
	if len(pi.Addrs) == 0 {
		return nil, fmt.Errorf("skipping node as it has no public IP address") // change knownErrs map if changing this msg
	}

	// init an exponential backoff
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = time.Second
	bo.MaxInterval = 10 * time.Second
	bo.MaxElapsedTime = time.Minute
	bo.Clock = c.cfg.Clock

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

		// save addresses into the peer store temporarily
		c.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.TempAddrTTL)

		timeoutCtx, cancel := context.WithTimeout(ctx, c.cfg.DialTimeout)
		conn, err := c.host.Network().DialPeer(timeoutCtx, pi.ID)
		cancel()

		// yay, it worked! Or has it? The caller checks the connectedness again.
		if err == nil {
			return conn, nil
		}

		switch true {
		case strings.Contains(err.Error(), db.ErrorStr[pgmodels.NetErrorConnectionRefused]):
			// Might be transient because the remote doesn't want us to connect. Try again!
		case strings.Contains(err.Error(), db.ErrorStr[pgmodels.NetErrorConnectionGated]):
			// Hints at a configuration issue and should not happen, but if it
			// does it could be transient. Try again anyway, but at least log a warning.
			logEntry.WithError(err).Warnln("Connection gated!")
		case strings.Contains(err.Error(), db.ErrorStr[pgmodels.NetErrorCantAssignRequestedAddress]):
			// Transient error due to local UDP issues. Try again!
		case strings.Contains(err.Error(), "dial backoff"):
			// should not happen because we disabled backoff checks with our
			// go-libp2p fork. Try again anyway, but at least log a warning.
			logEntry.WithError(err).Warnln("Dial backoff!")
		case strings.Contains(err.Error(), "RESOURCE_LIMIT_EXCEEDED (201)"): // thrown by a circuit relay
			// We already have too many open connections over a relay. Try again!
		default:
			logEntry.WithError(err).Debugln("Failed connecting to peer", pi.ID.ShortString())
			return nil, err
		}

		sleepDur := bo.NextBackOff()
		if sleepDur == backoff.Stop {
			logEntry.WithError(err).Debugln("Exceeded retries connecting to peer", pi.ID.ShortString())
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(sleepDur):
			retry += 1
			continue
		}

	}
}

// drainBuckets sends RPC messages to the given peer and asks for its closest peers to an artificial set
// of 15 random peer IDs with increasing common prefix lengths (CPL).
func (c *Crawler) drainBuckets(ctx context.Context, pi peer.AddrInfo) (*core.RoutingTable[PeerInfo], error) {
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
			neighbors, err := c.drainBucket(ctx, rt, pi.ID, count)
			if err != nil {
				errorBits.Add(1 << count)
				return err
			}

			allNeighborsLk.Lock()
			for _, n := range neighbors {
				allNeighbors[n.ID] = *n
			}
			allNeighborsLk.Unlock()

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

func (c *Crawler) drainBucket(ctx context.Context, rt *kbucket.RoutingTable, pid peer.ID, bucket uint) ([]*peer.AddrInfo, error) {
	// Generate a peer with the given common prefix length
	rpi, err := rt.GenRandPeerID(bucket)
	if err != nil {
		log.WithError(err).WithField("enr", pid.ShortString()).WithField("bucket", bucket).Warnln("Failed generating random peer ID")
		return nil, fmt.Errorf("generating random peer ID with CPL %d: %w", bucket, err)
	}

	var neighbors []*peer.AddrInfo
	for retry := 0; retry < 2; retry++ {
		neighbors, err = c.pm.GetClosestPeers(ctx, pid, rpi)
		if err == nil {
			// getting closest peers was successful!
			return neighbors, nil
		}

		var sleepDur time.Duration
		switch true {
		case strings.HasSuffix(err.Error(), network.ErrResourceLimitExceeded.Error()):
			// the remote has responded with a resource limit exceeded error. Try again soon!
			sleepDur = time.Second * time.Duration(3*(retry+1))
		case strings.Contains(err.Error(), "connection failed"):
			// This error happens in: https://github.com/libp2p/go-libp2p/blob/851f49d5edc46a24131a11f06df648602cd5968c/p2p/host/basic/basic_host.go#L648
			// we were connected to the remote but couldn't open a stream because
			// we lost the connection. Try again immediately! GetClosestPeers
			// internally calls NewStream on the basichost.Host which attempts
			// to connect to the peer again.
			sleepDur = 0
		default:
			// this is an unhandled error and we won't try again.
			return nil, fmt.Errorf("getting closest peer with CPL %d: %w", bucket, err)
		}

		// the other node has indicated that it's out of resources. Wait a bit and try again.
		select {
		case <-time.After(sleepDur): // may add jitter here
			continue
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("getting closest peer with CPL %d: %w", bucket, err)
}
