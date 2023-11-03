package discv5

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/bits"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/friendsofgo/errors"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/core"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	discover "github.com/dennis-tra/nebula-crawler/pkg/eth"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

type CrawlerConfig struct {
	DialTimeout  time.Duration
	AddrDialType config.AddrType
}

type Crawler struct {
	id           string
	cfg          *CrawlerConfig
	host         *basichost.BasicHost
	listener     *discover.UDPv5
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

	// start crawling both ways
	libp2pResultCh := c.crawlLibp2p(ctx, task)
	discV5ResultCh := c.crawlDiscV5(ctx, task)

	libp2pResult := <-libp2pResultCh
	discV5Result := <-discV5ResultCh

	cr := core.CrawlResult[PeerInfo]{
		CrawlerID:           c.id,
		Info:                task,
		CrawlStartTime:      crawlStart,
		RoutingTableFromAPI: false,
		RoutingTable:        discV5Result.RoutingTable,
		Agent:               libp2pResult.Agent,
		Protocols:           libp2pResult.Protocols,
		ConnectError:        libp2pResult.ConnectError,
		ConnectErrorStr:     libp2pResult.ConnectErrorStr,
		CrawlError:          discV5Result.Error,
		CrawlErrorStr:       discV5Result.ErrorStr,
		CrawlEndTime:        time.Now(),
		ConnectStartTime:    libp2pResult.ConnectStartTime,
		ConnectEndTime:      libp2pResult.ConnectEndTime,
		Properties:          c.PeerProperties(task.Node),
	}

	// We've now crawled this peer, so increment
	c.crawledPeers++

	return cr, nil
}

func (c *Crawler) PeerProperties(node *enode.Node) json.RawMessage {
	properties := map[string]any{}

	properties["seq"] = node.Record().Seq()
	properties["signature"] = node.Record().Signature()

	var enrEntryEth2 ENREntryEth2
	if err := node.Load(&enrEntryEth2); err == nil {
		if beaconData, err := enrEntryEth2.Data(); err == nil {
			properties["fork_digest"] = beaconData.ForkDigest.String()
			properties["next_fork_version"] = beaconData.NextForkVersion.String()
			properties["next_fork_epoch"] = beaconData.NextForkEpoch.String()
		}
	}

	var enrEntryAttnets ENREntryAttnets
	if err := node.Load(&enrEntryAttnets); err == nil {
		rawInt := binary.BigEndian.Uint64(enrEntryAttnets)
		properties["attnets_num"] = bits.OnesCount64(rawInt)
		properties["attnets"] = hex.EncodeToString(enrEntryAttnets)
	}

	var enrEntrySyncCommsSubnet ENREntrySyncCommsSubnet
	if err := node.Load(&enrEntrySyncCommsSubnet); err == nil {
		// check out https://github.com/prysmaticlabs/prysm/blob/203dc5f63b060821c2706f03a17d66b3813c860c/beacon-chain/p2p/subnets.go#L221
		properties["syncnets"] = hex.EncodeToString(enrEntrySyncCommsSubnet)
	}

	data, err := json.Marshal(properties)
	if err != nil {
		log.WithError(err).WithField("properties", properties).Warnln("Could not marshal peer properties")
		return nil
	}

	return data
}

type Libp2pResult struct {
	ConnectStartTime time.Time
	ConnectEndTime   time.Time
	ConnectError     error
	ConnectErrorStr  string
	Agent            string
	Protocols        []string
	ListenAddrs      []ma.Multiaddr
}

func (c *Crawler) crawlLibp2p(ctx context.Context, pi PeerInfo) chan Libp2pResult {
	resultCh := make(chan Libp2pResult)

	go func() {
		result := Libp2pResult{}

		addrInfo := peer.AddrInfo{
			ID:    pi.ID(),
			Addrs: pi.Addrs(),
		}

		result.ConnectStartTime = time.Now()
		result.ConnectError = c.connect(ctx, addrInfo) // use filtered addr list
		result.ConnectEndTime = time.Now()

		// If we could successfully connect to the peer we actually crawl it.
		if result.ConnectError == nil {

			// wait for the Identify exchange to complete
			c.identifyWait(ctx, addrInfo)

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

	retry := 0
	maxRetries := 1
	for {
		timeout := time.Duration(c.cfg.DialTimeout.Nanoseconds() / int64(retry+1))
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		err := c.host.Connect(timeoutCtx, pi)
		cancel()

		if err == nil {
			return nil
		}

		switch true {
		case strings.Contains(err.Error(), db.ErrorStr[models.NetErrorNegotiateSecurityProtocol]):
		case strings.Contains(err.Error(), db.ErrorStr[models.NetErrorConnectionRefused]):
		case strings.Contains(err.Error(), db.ErrorStr[models.NetErrorConnectionResetByPeer]):
		default:
			return err
		}

		if retry == maxRetries {
			metrics.VisitErrorsCount.With(metrics.CrawlLabel).Inc()
			return err
		}

		select {
		case <-ctx.Done():
			metrics.VisitErrorsCount.With(metrics.CrawlLabel).Inc()
			return ctx.Err()
		case <-time.After(time.Second * time.Duration(3*(retry+1))): // TODO: parameterize
			retry += 1
			continue
		}

	}
}

// identifyWait waits until any connection to a peer passed the Identify
// exchange successfully or all identification attempts have failed.
// The call to IdentifyWait returns immediately if the connection was
// identified in the past. We detect a successful identification if an
// AgentVersion is stored in the peer store
func (c *Crawler) identifyWait(ctx context.Context, pi peer.AddrInfo) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second) // TODO: parameterize
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

type DiscV5Result struct {
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

func (c *Crawler) crawlDiscV5(ctx context.Context, pi PeerInfo) chan DiscV5Result {
	resultCh := make(chan DiscV5Result)

	go func() {
		// all neighbors of pi. We're using a map to not deduplicate.
		allNeighbors := map[string]PeerInfo{}

		// errorBits tracks at which CPL errors have occurred.
		// 0000 0000 0000 0000 - No error
		// 0000 0000 0000 0001 - An error has occurred at CPL 0
		// 1000 0000 0000 0001 - An error has occurred at CPL 0 and 15
		errorBits := atomic.NewUint32(0)

		timeouts := 0
		enr, err := c.listener.RequestENR(pi.Node)
		if errors.Is(err, discover.ErrTimeout) {
			timeouts += 1
		}

		result := DiscV5Result{
			ENR: enr,
		}

		// loop through the buckets sequentially because discv5 is also doing that
		// internally, so we won't gain much by spawning multiple parallel go
		// routines here. Stop the process as soon as we have received a timeout and
		// don't let the following calls time out as well.
		for i := 0; i <= discover.NBuckets; i++ { // 15 is maximum
			var neighbors []*enode.Node
			neighbors, err = c.listener.FindNode(pi.Node, []uint{uint(discover.HashBits - i)})
			if err != nil {
				errorBits.Add(1 << i)

				if errors.Is(err, discover.ErrTimeout) {
					timeouts += 1
					if timeouts < 2 {
						continue
					}
				}

				err = fmt.Errorf("getting closest peer with CPL %d: %w", i, err)
				break
			}

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
		}
		result.DoneAt = time.Now()
		metrics.FetchedNeighborsCount.Observe(float64(len(allNeighbors)))

		result.RoutingTable = &core.RoutingTable[PeerInfo]{
			PeerID:    pi.ID(),
			Neighbors: []PeerInfo{},
			ErrorBits: uint16(errorBits.Load()),
			Error:     err,
		}

		for _, n := range allNeighbors {
			result.RoutingTable.Neighbors = append(result.RoutingTable.Neighbors, n)
		}

		result.DoneAt = time.Now()
		result.Error = err

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
