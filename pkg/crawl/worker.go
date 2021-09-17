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
func (w *Worker) StartCrawling(crawlQueue <-chan peer.AddrInfo, resultsQueue chan<- Result) {
	w.ServiceStarted()
	defer w.ServiceStopped()

	ctx := w.ServiceContext()
	logEntry := log.WithField("workerID", w.Identifier())
	for pi := range crawlQueue {
		logEntry = logEntry.WithField("targetID", pi.ID.Pretty()[:16]).WithField("crawlCount", w.crawledPeers)
		logEntry.Debugln("Crawling peer")

		cr := w.crawlPeer(ctx, pi)

		if cr.Error == nil && w.config.MeasureLatencies {
			cr.Latencies = w.measureLatency(ctx, pi)
		}

		select {
		case resultsQueue <- cr:
		case <-w.SigShutdown():
			return
		}

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
	// Only consider publicly reachable multi-addresses
	pi = filterPrivateMaddrs(pi)

	// The following loops extract addresses from the AddrInfo multi-addresses.
	// TODO: To which address should the ping messages be sent? Currently it's to all available IPv4 IPv6 addresses found.
	addrs := map[string]string{}
	for _, maddr := range pi.Addrs {
		for _, pr := range []int{ma.P_IP4, ma.P_IP6} {
			if addr, err := maddr.ValueForProtocol(pr); err == nil {
				addrs[addr] = addr
				break
			}
		}
	}

	// Exit early if there is no address
	if len(addrs) == 0 {
		log.Debugln("No good address")
		return nil
	}

	// Start sending ping messages to all IP addresses in parallel.
	var wg sync.WaitGroup
	results := make(chan interface{})
	for addr := range addrs {
		wg.Add(1)

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
				PeerID:          pi.ID.Pretty(),
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
