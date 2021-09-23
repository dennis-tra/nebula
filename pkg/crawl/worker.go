package crawl

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dennis-tra/nebula-crawler/pkg/queue"

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
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
)

var workerID = atomic.NewInt32(0)

// Result captures data that is gathered from crawling a single peer.
type Result struct {
	WorkerID string

	// The crawled peer
	Peer peer.AddrInfo

	// The neighbors of the crawled peer
	Neighbors []peer.AddrInfo

	// The latency to the particular peer as measured via ICM ping packets
	Latencies []*models.Latency

	// The agent version of the crawled peer
	Agent string

	// The protocols the peer supports
	Protocols []string

	// Any error that has occurred during the crawl
	Error error

	// As it can take some time to handle the result we track the timestamp explicitly
	ErrorTime time.Time
}

// Worker encapsulates a libp2p host that crawls the network.
type Worker struct {
	*service.Service

	host         host.Host
	config       *config.Config
	pm           *pb.ProtocolMessenger
	crawledPeers int
}

// NewWorker initializes a new worker based on the given configuration.
func NewWorker(h host.Host, conf *config.Config) (*Worker, error) {
	ms := &msgSender{
		h:         h,
		protocols: protocol.ConvertFromStrings(conf.Protocols),
		timeout:   conf.DialTimeout,
	}

	pm, err := pb.NewProtocolMessenger(ms)
	if err != nil {
		return nil, err
	}

	c := &Worker{
		Service: service.New(fmt.Sprintf("worker-%02d", workerID.Load())),
		host:    h,
		pm:      pm,
		config:  conf,
	}
	workerID.Inc()

	return c, nil
}

// StartCrawling reads from the given crawl queue and publishes the results on the results queue until interrupted.
func (w *Worker) StartCrawling(crawlQueue *queue.FIFO, resultsQueue *queue.FIFO) {
	w.ServiceStarted()
	defer w.ServiceStopped()

	ctx := w.ServiceContext()
	logEntry := log.WithField("workerID", w.Identifier())
	for elem := range crawlQueue.Consume() {
		pi := elem.(peer.AddrInfo)

		logEntry = logEntry.WithField("targetID", pi.ID.Pretty()[:16]).WithField("crawlCount", w.crawledPeers)
		logEntry.Debugln("Crawling peer")

		// Start crawling peers and measuring the latency in parallel. If we cannot connect to
		// the peer (in crawlPeer) we discard the measurement. This is done by cancelling the
		// latencyCtx. Crawling the peer will likely resolve earlier. Then we wait until
		// the measurement is done.
		var cr Result
		var wg sync.WaitGroup
		wg.Add(2)
		latencyCtx, cancel := context.WithCancel(ctx)

		// Start crawling the peer.
		go func() {
			defer wg.Done()
			cr = w.crawlPeer(ctx, pi)
			if cr.Error != nil {
				cancel()
			}
		}()

		// Start measuring the peer.
		var latencies []*models.Latency
		go func() {
			defer wg.Done()
			if w.config.MeasureLatencies {
				latencies = w.measureLatency(latencyCtx, pi)
			}
		}()

		// Wait until the crawl and measurement are done.
		wg.Wait()

		// In case there is no error in crawling the peer we still need to cancel the context to not leak it.
		cancel()

		cr.Latencies = latencies
		resultsQueue.Produce() <- cr
		logEntry.Debugln("Crawled peer")
	}

	logEntry.Debugf("Crawled %d peers\n", w.crawledPeers)
}

func (w *Worker) crawlPeer(ctx context.Context, pi peer.AddrInfo) Result {
	start := time.Now()
	defer stats.Record(ctx, metrics.PeerCrawlDuration.M(millisSince(start)))

	cr := Result{
		WorkerID: w.Identifier(),
		Peer:     filterPrivateMaddrs(pi),
		Agent:    "n.a.",
	}

	cr.Error = w.connect(ctx, pi)
	if cr.Error == nil {

		ps := w.host.Peerstore()

		// Extract agent
		if agent, err := ps.Get(pi.ID, "AgentVersion"); err == nil {
			cr.Agent = agent.(string)
		}

		// Extract protocols
		if protocols, err := ps.GetProtocols(pi.ID); err == nil {
			cr.Protocols = protocols
		}

		// Fetch all neighbors
		cr.Neighbors, cr.Error = w.fetchNeighbors(ctx, pi)
	}

	// If connection or neighbor fetching failed we track the timestamp
	if cr.Error != nil {
		cr.ErrorTime = time.Now()
	}

	// Free connection resources
	if err := w.host.Network().ClosePeer(pi.ID); err != nil {
		log.WithError(err).WithField("targetID", pi.ID.Pretty()[:16]).Warnln("Could not close connection to peer")
	}

	// We've now crawled this peer, so increment
	w.crawledPeers++

	return cr
}

// millisSince returns the number of milliseconds between now and the given time.
func millisSince(start time.Time) float64 {
	return float64(time.Since(start)) / float64(time.Millisecond)
}

// connect strips all private multi addresses in `pi` and establishes a connection to the given peer.
// It also handles metric capturing.
func (w *Worker) connect(ctx context.Context, pi peer.AddrInfo) error {
	start := time.Now()
	stats.Record(ctx, metrics.CrawlConnectsCount.M(1))

	pi = filterPrivateMaddrs(pi)
	if len(pi.Addrs) == 0 {
		stats.Record(ctx, metrics.CrawlConnectErrorsCount.M(1))
		return fmt.Errorf("skipping node as it has no public IP address") // change knownErrs map if changing this msg
	}

	ctx, cancel := context.WithTimeout(ctx, w.config.DialTimeout)
	defer cancel()

	if err := w.host.Connect(ctx, pi); err != nil {
		stats.Record(ctx, metrics.CrawlConnectErrorsCount.M(1))
		return err
	}

	stats.Record(w.ServiceContext(), metrics.CrawlConnectDuration.M(millisSince(start)))
	return nil
}

// fetchNeighbors sends RPC messages to the given peer and asks for its closest peers to an artificial set
// of 15 random peer IDs with increasing common prefix lengths (CPL). The returned peers are streamed
// to the results channel.
func (w *Worker) fetchNeighbors(ctx context.Context, pi peer.AddrInfo) ([]peer.AddrInfo, error) {
	start := time.Now()
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

			neighbors, err := w.pm.GetClosestPeers(ctx, pi.ID, rpi)
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
	stats.Record(w.ServiceContext(),
		metrics.FetchNeighborsDuration.M(millisSince(start)),
		metrics.FetchedNeighborsCount.M(float64(len(allNeighbors))),
	)
	return allNeighbors, err
}

// measureLatency measures the ICM ping latency to the given peer.
func (w *Worker) measureLatency(ctx context.Context, pi peer.AddrInfo) []*models.Latency {
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
