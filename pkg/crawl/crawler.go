package crawl

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	ma "github.com/multiformats/go-multiaddr"
	madns "github.com/multiformats/go-multiaddr-dns"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/stats"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
)

var crawlerID = atomic.NewInt32(0)

// Result captures data that is gathered from crawling a single peer.
type Result struct {
	// The crawler that generated this result
	CrawlerID string

	// The crawled peer
	Peer peer.AddrInfo

	// The neighbors of the crawled peer
	Neighbors []peer.AddrInfo

	// The agent version of the crawled peer
	Agent string

	// The protocols the peer supports
	Protocols []string

	// Any error that has occurred during the crawl
	Error error

	// The above error transferred to a known error
	DialError string

	// When was the crawl started
	CrawlStartTime time.Time

	// When did this crawl end
	CrawlEndTime time.Time

	// When was the connection attempt made
	ConnectStartTime time.Time

	// As it can take some time to handle the result we track the timestamp explicitly
	ConnectEndTime time.Time

	// The latency to the particular peer as measured via ICM ping packets
	PingLatencies []*models.Latency
}

// CrawlDuration returns the time it took to crawl to the peer (connecting + fetching neighbors)
func (r *Result) CrawlDuration() time.Duration {
	return r.CrawlEndTime.Sub(r.CrawlStartTime)
}

// ConnectDuration returns the time it took to connect to the peer.
func (r *Result) ConnectDuration() time.Duration {
	return r.ConnectEndTime.Sub(r.ConnectStartTime)
}

// Crawler encapsulates a libp2p host that crawls the network.
type Crawler struct {
	*service.Service

	host         host.Host
	config       *config.Config
	pm           *pb.ProtocolMessenger
	crawledPeers int
}

// NewCrawler initializes a new crawler based on the given configuration.
func NewCrawler(h host.Host, conf *config.Config) (*Crawler, error) {
	ms := &msgSender{
		h:         h,
		protocols: protocol.ConvertFromStrings(conf.Protocols),
		timeout:   conf.DialTimeout,
	}

	pm, err := pb.NewProtocolMessenger(ms)
	if err != nil {
		return nil, err
	}

	c := &Crawler{
		Service: service.New(fmt.Sprintf("crawler-%02d", crawlerID.Load())),
		host:    h,
		pm:      pm,
		config:  conf,
	}
	crawlerID.Inc()

	return c, nil
}

// StartCrawling enters an endless loop and consumes crawl jobs from the crawl queue
// and publishes its result on the results queue until it is told to stop or the
// crawl queue was closed.
func (c *Crawler) StartCrawling(crawlQueue *queue.FIFO, resultsQueue *queue.FIFO) {
	c.ServiceStarted()
	defer c.ServiceStopped()
	ctx := c.ServiceContext()

	for {
		// Give the shutdown signal precedence
		select {
		case <-c.SigShutdown():
			return
		default:
		}

		select {
		case elem, ok := <-crawlQueue.Consume():
			if !ok {
				// The crawl queue was closed
				return
			}
			result := c.handleCrawlJob(ctx, elem.(peer.AddrInfo))
			resultsQueue.Push(result)
		case <-c.SigShutdown():
			return
		}
	}
}

// handleCrawlJob takes peer address information and crawls (connects and fetches neighbors).
func (c *Crawler) handleCrawlJob(ctx context.Context, pi peer.AddrInfo) Result {
	logEntry := log.WithFields(log.Fields{
		"crawlerID":  c.Identifier(),
		"targetID":   pi.ID.Pretty()[:16],
		"crawlCount": c.crawledPeers,
	})
	logEntry.Debugln("Crawling peer")
	defer logEntry.Debugln("Crawled peer")

	// Start crawling peers and measuring the latency in parallel. If we cannot connect to
	// the peer (in crawlPeer) we discard the measurement. This is done by cancelling the
	// latencyCtx. Crawling the peer will likely resolve earlier. Then we wait until
	// the measurement is done.
	var cr Result
	var wg sync.WaitGroup
	wg.Add(2)
	latencyCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start crawling the peer.
	go func() {
		defer wg.Done()
		cr = c.crawlPeer(ctx, pi)
		if cr.Error != nil {
			// stop latency measurement if we could not connect to the peer.
			cancel()
		}
	}()

	// Start measuring the peer.
	var latencies []*models.Latency // use separate var to avoid race condition
	go func() {
		defer wg.Done()
		if c.config.MeasureLatencies {
			latencies = c.measureLatency(latencyCtx, pi)
		}
	}()

	// Wait until the crawl and measurement are done.
	wg.Wait()

	cr.PingLatencies = latencies
	return cr
}

func (c *Crawler) crawlPeer(ctx context.Context, pi peer.AddrInfo) Result {
	cr := Result{
		CrawlerID:      c.Identifier(),
		Peer:           filterPrivateMaddrs(pi),
		CrawlStartTime: time.Now(),
	}

	cr.ConnectStartTime = time.Now()
	cr.Error = c.connect(ctx, pi)
	cr.ConnectEndTime = time.Now()

	// If we could successfully connect to the peer we actually crawl it.
	if cr.Error == nil {

		ps := c.host.Peerstore()

		// Extract agent
		if agent, err := ps.Get(pi.ID, "AgentVersion"); err == nil {
			cr.Agent = agent.(string)
		}

		// Extract protocols
		if protocols, err := ps.GetProtocols(pi.ID); err == nil {
			cr.Protocols = protocols
		}

		// Fetch all neighbors
		cr.Neighbors, cr.Error = c.fetchNeighbors(ctx, pi)
	}

	if cr.Error != nil {
		cr.DialError = db.DialError(cr.Error)
	}

	// Free connection resources
	if err := c.host.Network().ClosePeer(pi.ID); err != nil {
		log.WithError(err).WithField("targetID", pi.ID.Pretty()[:16]).Warnln("Could not close connection to peer")
	}

	// We've now crawled this peer, so increment
	c.crawledPeers++

	cr.CrawlEndTime = time.Now()

	return cr
}

// connect strips all private multi addresses in `pi` and establishes a connection to the given peer.
// It also handles metric capturing.
func (c *Crawler) connect(ctx context.Context, pi peer.AddrInfo) error {
	stats.Record(ctx, metrics.CrawlConnectsCount.M(1))

	if len(pi.Addrs) == 0 {
		stats.Record(ctx, metrics.CrawlConnectErrorsCount.M(1))
		return fmt.Errorf("skipping node as it has no public IP address") // change knownErrs map if changing this msg
	}

	ctx, cancel := context.WithTimeout(ctx, c.config.DialTimeout)
	defer cancel()

	if err := c.host.Connect(ctx, pi); err != nil {
		stats.Record(ctx, metrics.CrawlConnectErrorsCount.M(1))
		return err
	}

	return nil
}

// fetchNeighbors sends RPC messages to the given peer and asks for its closest peers to an artificial set
// of 15 random peer IDs with increasing common prefix lengths (CPL). The returned peers are streamed
// to the results channel.
func (c *Crawler) fetchNeighbors(ctx context.Context, pi peer.AddrInfo) ([]peer.AddrInfo, error) {
	var allNeighbors []peer.AddrInfo
	rt, err := kbucket.NewRoutingTable(20, kbucket.ConvertPeerID(pi.ID), time.Hour, nil, time.Hour, nil)
	if err != nil {
		return allNeighbors, err
	}

	allNeighborsLk := sync.RWMutex{}
	errg := errgroup.Group{}
	for i := uint(0); i <= 15; i++ { // 15 is maximum
		count := i
		errg.Go(func() error {
			// Generate a peer with the given common prefix length
			rpi, err := rt.GenRandPeerID(count)
			if err != nil {
				return errors.Wrapf(err, "generating random peer ID with CPL %d", count)
			}

			neighbors, err := c.pm.GetClosestPeers(ctx, pi.ID, rpi)
			if err != nil {
				return errors.Wrapf(err, "getting closest peer with CPL %d", count)
			}

			allNeighborsLk.Lock()
			defer allNeighborsLk.Unlock()
			for _, n := range neighbors {
				allNeighbors = append(allNeighbors, *n)
			}

			return nil
		})
	}
	err = errg.Wait()
	stats.Record(c.ServiceContext(),
		metrics.FetchedNeighborsCount.M(float64(len(allNeighbors))),
	)
	return allNeighbors, err
}

// measureLatency measures the ICM ping latency to the given peer.
func (c *Crawler) measureLatency(ctx context.Context, pi peer.AddrInfo) []*models.Latency {
	// TODO: The following three steps can probably be consolidated. In the current state it's quite messy.

	// Only consider publicly reachable multi-addresses
	pi = filterPrivateMaddrs(pi)

	// Resolve DNS multi addresses to IP addresses (especially maddrs with the dnsaddr protocol)
	pi.Addrs = resolveAddrs(ctx, pi)

	// The following loops extract addresses from the AddrInfo multi-addresses.
	// The set of multi addresses could contain multiple maddrs with the
	// same IPv4/IPv6 addresses. This loop de-duplicates that.
	// TODO: To which address should the ping messages be sent? Currently it's to all found addresses.
	// TODO: The deduplication can probably be implemented a little bit prettier
	addrsMap := map[string]string{}
	for _, maddr := range pi.Addrs {
		for _, pr := range []int{ma.P_IP4, ma.P_IP6} { // DNS protocols are stripped via resolveAddrs above
			if addr, err := maddr.ValueForProtocol(pr); err == nil {
				addrsMap[addr] = addr
				break
			}
		}
	}

	// Exit early if there is no address
	if len(addrsMap) == 0 {
		return nil
	}

	// Start sending ping messages to all IP addresses in parallel.
	var wg sync.WaitGroup
	results := make(chan interface{})
	for addr := range addrsMap {

		// Configure the new pinger instance
		pinger, err := ping.NewPinger(addr)
		if err != nil {
			log.WithError(errors.Wrap(err, "new pinger")).Warnln("Error instantiating new pinger")
			continue
		}

		pinger.Timeout = time.Minute
		pinger.Count = 10

		// This Go routine allows reacting to context cancellations (e.g., user presses ^C)
		// The done channel is necessary to not leak this go routine after the pinger has finished.
		done := make(chan struct{})
		go func() {
			select {
			case <-done:
			case <-ctx.Done():
				pinger.Stop()
			}
		}()

		// This Go routine starts sending ICM pings to the address configured above.
		// After it has terminated (successfully or erroneously) it sends the result
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(done)
			// Blocks until finished.
			if err := pinger.Run(); err != nil {
				results <- err
			} else {
				results <- pinger.Statistics()
			}
		}()
	}

	// Since we're ranging over the results channel below we need to
	// know when the pinger Go routines are done. This the case
	// after the wg.Wait() call returns. We close the channel and
	// break out of the for loop below.
	go func() {
		wg.Wait()
		close(results)
	}()

	var latencies []*models.Latency
	for result := range results {
		switch res := result.(type) {
		case error:
			log.WithError(errors.Wrap(res, "pinger run")).Warnln("Error pinging peer")
		case *ping.Statistics:
			latencies = append(latencies, &models.Latency{
				Address:         res.Addr,
				PingLatencySAvg: res.AvgRtt.Seconds(),
				PingLatencySSTD: res.StdDevRtt.Seconds(),
				PingLatencySMin: res.MinRtt.Seconds(),
				PingLatencySMax: res.MaxRtt.Seconds(),
				PingPacketsSent: res.PacketsSent,
				PingPacketsRecv: res.PacketsRecv,
				PingPacketsDupl: res.PacketsRecvDuplicates,
				PingPacketLoss:  res.PacketLoss,
			})
		}
	}

	return latencies
}

// filterPrivateMaddrs strips private multiaddrs from the given peer address information.
func filterPrivateMaddrs(pi peer.AddrInfo) peer.AddrInfo {
	filtered := peer.AddrInfo{
		ID:    pi.ID,
		Addrs: []ma.Multiaddr{},
	}

	// Just keep public multi addresses
	for _, maddr := range pi.Addrs {
		if manet.IsPrivateAddr(maddr) {
			continue
		}
		filtered.Addrs = append(filtered.Addrs, maddr) // TODO: Strip relays?
	}

	return filtered
}

// resolveAddrs loops through the multi addresses of the given peer and recursively resolves
// the various DNS protocols (especially the dnsaddr protocol). This implementation is
// taken from:
// https://github.com/libp2p/go-libp2p/blob/9d3fd8bc4675b9cebf3102bdf62e56204c67ce5b/p2p/host/basic/basic_host.go#L676
func resolveAddrs(ctx context.Context, pi peer.AddrInfo) []ma.Multiaddr {
	proto := ma.ProtocolWithCode(ma.P_P2P).Name
	p2paddr, err := ma.NewMultiaddr("/" + proto + "/" + pi.ID.Pretty())
	if err != nil {
		return []ma.Multiaddr{}
	}

	resolveSteps := 0

	// Recursively resolve all addrs.
	//
	// While the toResolve list is non-empty:
	// * Pop an address off.
	// * If the address is fully resolved, add it to the resolved list.
	// * Otherwise, resolve it and add the results to the "to resolve" list.
	toResolve := append(([]ma.Multiaddr)(nil), pi.Addrs...)
	resolved := make([]ma.Multiaddr, 0, len(pi.Addrs))
	for len(toResolve) > 0 {
		// pop the last addr off.
		addr := toResolve[len(toResolve)-1]
		toResolve = toResolve[:len(toResolve)-1]

		// if it's resolved, add it to the resolved list.
		if !madns.Matches(addr) {
			resolved = append(resolved, addr)
			continue
		}

		resolveSteps++

		// otherwise, resolve it
		reqaddr := addr.Encapsulate(p2paddr)
		resaddrs, err := madns.DefaultResolver.Resolve(ctx, reqaddr)
		if err != nil {
			log.Infof("error resolving %s: %s", reqaddr, err)
		}

		// add the results to the toResolve list.
		for _, res := range resaddrs {
			pi, err := peer.AddrInfoFromP2pAddr(res)
			if err != nil {
				log.Infof("error parsing %s: %s", res, err)
			}
			toResolve = append(toResolve, pi.Addrs...)
		}
	}

	return resolved
}
