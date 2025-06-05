package discv4

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/netip"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/discover/v4wire"
	"github.com/ethereum/go-ethereum/p2p/enode"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	pgmodels "github.com/dennis-tra/nebula-crawler/db/models/pg"
)

type CrawlerConfig struct {
	DialTimeout  time.Duration
	AddrDialType config.AddrType
	MaxJitter    time.Duration
	LogErrors    bool
	KeepENR      bool
}

type Crawler struct {
	id           string
	cfg          *CrawlerConfig
	listener     *discover.UDPv4
	client       *Client
	crawledPeers int
	taskDoneChan chan time.Time
	done         chan struct{}
}

var _ core.Worker[PeerInfo, core.CrawlResult[PeerInfo]] = (*Crawler)(nil)

func (c *Crawler) Work(ctx context.Context, task PeerInfo) (core.CrawlResult[PeerInfo], error) {
	// indicate to the driver that we have handled a task
	defer func() { c.taskDoneChan <- time.Now() }()

	// add a startup jitter delay to prevent all workers to crawl at exactly the
	// same time and potentially overwhelm the machine that Nebula is running on
	if c.crawledPeers == 0 {
		jitter := time.Duration(0)
		if c.cfg.MaxJitter > 0 { // could be <= 0 if the worker count is 1
			jitter = time.Duration(rand.Int63n(int64(c.cfg.MaxJitter)))
		}
		select {
		case <-time.After(jitter):
		case <-ctx.Done():
		}
	}

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

	discV4Result := <-discv4ResultCh
	devp2pResult := <-devp2pResultCh

	properties := map[string]any{}

	// keep track of all unknown connection errors
	if devp2pResult.ConnectErrorStr == pgmodels.NetErrorUnknown && devp2pResult.ConnectError != nil {
		properties["connect_error"] = devp2pResult.ConnectError.Error()
	}

	// keep track of all unknown crawl errors
	if discV4Result.ErrorStr == pgmodels.NetErrorUnknown && discV4Result.Error != nil {
		properties["crawl_error"] = discV4Result.Error.Error()
	}

	// keep track of the strategy that we used to crawl that peer
	if discV4Result.Strategy != "" {
		properties["strategy"] = string(discV4Result.Strategy)
	}

	if devp2pResult.Status != nil {
		properties["network_id"] = devp2pResult.Status.NetworkID
		properties["fork_id"] = hex.EncodeToString(devp2pResult.Status.ForkID.Hash[:])
	}

	if c.cfg.KeepENR {
		properties["enr"] = task.enr // discV4Result.ENR.String() panics :/
	}

	// keep track of all unknown connection errors
	if devp2pResult.ConnectErrorStr == pgmodels.NetErrorUnknown && devp2pResult.ConnectError != nil {
		properties["connect_error"] = devp2pResult.ConnectError.Error()
	}

	// keep track of all unknown crawl errors
	if discV4Result.ErrorStr == pgmodels.NetErrorUnknown && discV4Result.Error != nil {
		properties["crawl_error"] = discV4Result.Error.Error()
	}

	data, err := json.Marshal(properties)
	if err != nil {
		logEntry.WithError(err).WithField("properties", properties).Warnln("Could not marshal peer properties")
	}

	connectMaddr := devp2pResult.ConnectMaddr
	if connectMaddr == nil && discV4Result.Error == nil {
		connectMaddr = discV4Result.ConnectMaddr
	}

	// construct a lookup map with maddrs we tried to dial
	dialMaddrsMap := make(map[string]struct{})
	for _, maddr := range devp2pResult.DialMaddrs {
		dialMaddrsMap[string(maddr.Bytes())] = struct{}{}
	}

	// construct a slice with multi addresses that we did NOT try to dial
	// We loop through all known addrs and check if we dialed it.
	// Further, build a lookup map with known multi addresses for the peer.
	// this map contains all maddrs we've found via the discovery protocol
	var (
		knownMaddrsMap = make(map[string]struct{}, len(task.maddrs))
		filteredMaddrs []ma.Multiaddr
	)
	for _, maddr := range task.maddrs {
		knownMaddrsMap[string(maddr.Bytes())] = struct{}{}
		if _, ok := dialMaddrsMap[string(maddr.Bytes())]; ok {
			continue
		}
		filteredMaddrs = append(filteredMaddrs, maddr)
	}

	cr := core.CrawlResult[PeerInfo]{
		CrawlerID:           c.id,
		Info:                task,
		CrawlStartTime:      crawlStart,
		RoutingTableFromAPI: false,
		RoutingTable:        discV4Result.RoutingTable,
		Agent:               devp2pResult.Agent,
		Protocols:           devp2pResult.Protocols,
		DialMaddrs:          devp2pResult.DialMaddrs,
		FilteredMaddrs:      filteredMaddrs,
		ExtraMaddrs:         []ma.Multiaddr{},
		ListenMaddrs:        []ma.Multiaddr{},
		ConnectMaddr:        connectMaddr,
		DialErrors:          db.MaddrErrors(devp2pResult.DialMaddrs, devp2pResult.ConnectError),
		ConnectError:        devp2pResult.ConnectError,
		ConnectErrorStr:     devp2pResult.ConnectErrorStr,
		CrawlError:          discV4Result.Error,
		CrawlErrorStr:       discV4Result.ErrorStr,
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

	// The multiaddress via which we have received a response
	ConnectMaddr ma.Multiaddr

	// The neighbors of the crawled peer
	RoutingTable *core.RoutingTable[PeerInfo]

	// The strategy used to crawl the peer
	Strategy CrawlStrategy

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
		defer close(resultCh)

		// the final result struct
		result := DiscV4Result{}

		enr, err := c.listener.RequestENR(pi.Node)
		if err != nil {
			result.ENR = pi.Node
			err = nil
		} else {
			result.ENR = enr
			now := time.Now()
			result.RespondedAt = &now
		}

		// the number of probes to issue against bucket 0
		probes := 3

		closestMap, closestSet, respondedAt, err := c.probeBucket0(pi, probes, result.RespondedAt != nil)

		if err == nil {
			// track the respondedAt timestamp if it wasn't already set
			if result.RespondedAt == nil && !respondedAt.IsZero() {
				result.RespondedAt = &respondedAt
				result.ConnectMaddr = pi.UDPMaddr()
			}

			result.Strategy = determineStrategy(closestSet)

			var remainingClosest map[string]PeerInfo
			switch result.Strategy {
			case crawlStrategySingleProbe:
				remainingClosest = c.crawlRemainingBucketsSequentially(pi.Node, pi.udpAddr, 1)
			case crawlStrategyMultiProbe:
				remainingClosest = c.crawlRemainingBucketsSequentially(pi.Node, pi.udpAddr, 3)
			case crawlStrategyRandomProbe:
				probesPerBucket := int(1.3333 * discover.BucketSize / (float32(len(closestMap)) / float32(probes)))
				remainingClosest = c.crawlRemainingBucketsSequentially(pi.Node, pi.udpAddr, probesPerBucket)
			default:
				panic("unexpected strategy: " + string(result.Strategy))
			}

			for k, v := range remainingClosest {
				closestMap[k] = v
			}
		}

		// track done timestamp and error
		result.DoneAt = time.Now()
		result.Error = err

		result.RoutingTable = &core.RoutingTable[PeerInfo]{
			PeerID:    pi.ID(),
			Neighbors: make([]PeerInfo, 0, len(closestMap)),
			ErrorBits: uint16(0),
			Error:     err,
		}

		for _, n := range closestMap {
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
	}()

	return resultCh
}

func (c *Crawler) probeBucket0(pi PeerInfo, probes int, returnedENR bool) (map[string]PeerInfo, []mapset.Set[string], time.Time, error) {
	var (
		respondedAt time.Time
		closestMap  = make(map[string]PeerInfo)
		closestSets []mapset.Set[string]
		errs        []error
	)

	// do it sequentially because if a remote peer returns `probes` responses
	// containing only three peers each (we've observed that) then these
	// will be mapped to a single response because of how the discv4
	// implementation works. This in turn means that the determineStrategy
	// won't work
	for i := 0; i < probes; i++ {
		// first, we generate a random key that falls into bucket 0
		targetKey, err := GenRandomPublicKey(pi.Node.ID(), 0)
		if err != nil {
			return nil, nil, time.Time{}, err
		}

		// second, we do the Find node request
		closest, err := c.listener.FindNode(pi.Node.ID(), pi.udpAddr, targetKey)
		if err != nil {
			// exit early if the node hasn't returned an ENR and the first probe
			// also timed out
			if !returnedENR && errors.Is(err, discover.ErrTimeout) {
				return nil, nil, time.Time{}, fmt.Errorf("failed to probe bucket 0: %w", discover.ErrTimeout)
			}

			errs = append(errs, err)
		} else if !respondedAt.IsZero() {
			respondedAt = time.Now()
		}

		// third, we parse the responses into our [PeerInfo] struct
		for _, c := range closest {
			pi, err := NewPeerInfo(c)
			if err != nil {
				log.WithError(err).Warnln("Failed parsing ethereum node neighbor")
				continue
			}

			closestMap[pi.DeduplicationKey()] = pi
		}

		closestSets = append(closestSets, mapset.NewThreadUnsafeSetFromMapKeys(closestMap))
	}

	if len(errs) == probes {
		return nil, nil, time.Time{}, fmt.Errorf("failed to probe bucket 0: %w", errors.Join(errs...))
	}

	return closestMap, closestSets, respondedAt, nil
}

type CrawlStrategy string

const (
	crawlStrategySingleProbe CrawlStrategy = "single-probe"
	crawlStrategyMultiProbe  CrawlStrategy = "multi-probe"
	crawlStrategyRandomProbe CrawlStrategy = "random-probe"
)

func determineStrategy(sets []mapset.Set[string]) CrawlStrategy {
	// Calculate the average difference between two responses. If the response
	// sizes are always 16, one new peer will result in a symmetric difference
	// of cardinality 2. One peer in the first set that's not in the second and one
	// peer in the second that's not in the first set. We consider that it's the
	// happy path if the average symmetric difference is less than 2.
	avgSymDiff := float32(0)
	diffCount := float32(0)
	allNodes := mapset.NewThreadUnsafeSet[string]()
	for i := 0; i < len(sets); i++ {
		allNodes = allNodes.Union(sets[i])
		for j := i + 1; j < len(sets); j++ {
			diffCount += 1
			avgSymDiff += float32(sets[i].SymmetricDifference(sets[j]).Cardinality())
		}
	}
	avgSymDiff /= diffCount

	switch {
	case avgSymDiff < 2:
		return crawlStrategySingleProbe
	case allNodes.Cardinality() > v4wire.MaxNeighbors:
		return crawlStrategyMultiProbe
	default:
		return crawlStrategyRandomProbe
	}
}

func (c *Crawler) crawlRemainingBucketsSequentially(node *enode.Node, udpAddr netip.AddrPort, probesPerBucket int) map[string]PeerInfo {
	timeouts := 0
	allNeighbors := map[string]PeerInfo{}

OUTER:
	for i := 1; i < 15; i++ { // although there are 17 buckets, GenRandomPublicKey only supports the first 16
		for j := 0; j < probesPerBucket; j++ {

			// first, we generate a random key that falls into bucket 0
			targetKey, err := GenRandomPublicKey(node.ID(), i)
			if err != nil {
				log.WithError(err).WithField("nodeID", node.ID().String()).Warnf("Failed generating random key for bucket %d", i)
				break
			}

			// second, we do the Find node request
			closest, err := c.listener.FindNode(node.ID(), udpAddr, targetKey)
			if errors.Is(err, discover.ErrTimeout) {
				timeouts += 1
				if timeouts > 3 {
					break OUTER
				}
				continue
			} else if err != nil {
				break OUTER
			}

			for _, c := range closest {
				pi, err := NewPeerInfo(c)
				if err != nil {
					log.WithError(err).Warnln("Failed parsing ethereum node neighbor")
					continue
				}
				allNeighbors[pi.DeduplicationKey()] = pi
			}
		}
	}

	return allNeighbors
}

type Devp2pResult struct {
	DialMaddrs       []ma.Multiaddr
	ConnectStartTime time.Time
	ConnectEndTime   time.Time
	IdentifyEndTime  time.Time
	ConnectError     error
	ConnectErrorStr  string
	ConnectMaddr     ma.Multiaddr
	Agent            string
	Protocols        []string
	Status           *eth.StatusPacket
}

func (c *Crawler) crawlDevp2p(ctx context.Context, pi PeerInfo) <-chan Devp2pResult {
	resultCh := make(chan Devp2pResult)
	go func() {
		var dialMaddrs []ma.Multiaddr
		if tcpMaddr := pi.TCPMaddr(); tcpMaddr != nil {
			dialMaddrs = []ma.Multiaddr{tcpMaddr}
		}

		// the final result struct
		result := Devp2pResult{
			DialMaddrs: dialMaddrs,
		}

		result.ConnectStartTime = time.Now()
		conn, err := c.client.Connect(ctx, pi)
		result.ConnectEndTime = time.Now()
		result.ConnectError = err

		if result.ConnectError == nil {
			result.ConnectMaddr = pi.TCPMaddr()

			// start another go routine to cancel the entire operation if it
			// times out. The context will be cancelled when this function
			// returns or the timeout is reached. In both cases, we close the
			// connection to the remote peer which will trigger that the call
			// to Identify below will return (if the context is canceled because
			// of a timeout and not function return).
			timeoutCtx, cancel := context.WithTimeout(ctx, c.cfg.DialTimeout)
			defer cancel()
			go func() {
				<-timeoutCtx.Done()
				// Free connection resources
				if err := conn.Close(); err != nil && !strings.Contains(err.Error(), errUseOfClosedNetworkConnectionStr) {
					log.WithError(err).WithField("remoteID", pi.ID().ShortString()).Warnln("Could not close connection to peer")
				}
			}()

			resp, status, err := conn.Identify()
			if err != nil && resp == nil && status == nil {
				result.ConnectError = err
			}
			result.IdentifyEndTime = time.Now()
			result.Status = status

			if resp != nil {
				result.Agent = resp.Name
				protocols := make([]string, len(resp.Caps))
				for i, c := range resp.Caps {
					protocols[i] = "/" + c.String()
				}
				result.Protocols = protocols
			}
		} else {
			// if there was a connection error, parse it to a known one
			result.ConnectErrorStr = db.NetError(result.ConnectError)
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
