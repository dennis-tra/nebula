package crawl

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/protocol"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
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

	// The agent version of the crawled peer
	Agent string

	// The protocols the peer supports
	Protocols []string

	// Any error that has occurred during the crawl
	Error error
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
func (w *Worker) StartCrawling(crawlQueue chan peer.AddrInfo, resultsQueue chan Result) {
	w.ServiceStarted()
	defer w.ServiceStopped()
	defer log.Debugf("Worker %s crawled %d peers\n", w.Identifier(), w.crawledPeers)

	ctx := w.ServiceContext()
	for pi := range crawlQueue {
		logEntry := log.WithField("targetID", pi.ID.Pretty()[:16]).WithField("workerID", w.Identifier())
		logEntry.Debugln("Crawling peer ", pi.ID.Pretty()[:16])

		cr := w.crawlPeer(ctx, pi)

		select {
		case resultsQueue <- cr:
		case <-w.SigShutdown():
			return
		}

		logEntry.Debugln("Crawled peer ", pi.ID.Pretty()[:16])
	}
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

	if err := w.host.Network().ClosePeer(pi.ID); err != nil {
		log.WithError(err).WithField("targetID", pi.ID.Pretty()[:16]).Warnln("Could not close connection to peer")
	}

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
	stats.Record(ctx, metrics.ConnectsCount.M(1))

	pi = filterPrivateMaddrs(pi)
	if len(pi.Addrs) == 0 {
		stats.Record(ctx, metrics.ConnectErrorsCount.M(1))
		return fmt.Errorf("skipping node as it has no public IP address") // change knownErrs map if changing this msg
	}

	ctx, cancel := context.WithTimeout(ctx, w.config.DialTimeout)
	defer cancel()

	if err := w.host.Connect(ctx, pi); err != nil {
		stats.Record(ctx, metrics.ConnectErrorsCount.M(1))
		return err
	}

	stats.Record(w.ServiceContext(), metrics.ConnectDuration.M(millisSince(start)))
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
