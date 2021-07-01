package crawl

import (
	"context"
	"fmt"
	"sync"
	"time"

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

var runningWorkers = atomic.NewInt32(0)

// CrawlResult captures data that is gathered from crawling a single peer.
type CrawlResult struct {
	WorkerID string

	// The crawled peer
	Peer peer.AddrInfo

	// The neighbors of the crawled peer
	Neighbors []peer.AddrInfo

	// The agent version of the crawled peer
	Agent string

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

func NewWorker(h host.Host, conf *config.Config) (*Worker, error) {
	ms := &msgSender{
		h:         h,
		protocols: ProtocolStrings,
		timeout:   time.Second * 30,
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

func millisSince(start time.Time) float64 {
	return float64(time.Since(start)) / float64(time.Millisecond)
}

func (w *Worker) Shutdown() {
	defer w.Service.Shutdown()
	log.Debugf("Worker %s crawled %d peers\n", w.Identifier(), w.crawledPeers)
}

func (w *Worker) StartCrawling(crawlQueue chan peer.AddrInfo, resultsQueue chan CrawlResult) {
	w.ServiceStarted()
	defer w.ServiceStopped()

	ctx := w.ServiceContext()
	for pi := range crawlQueue {
		start := time.Now()
		logEntry := log.WithField("targetID", pi.ID.Pretty()[:16]).WithField("workerID", w.Identifier())
		logEntry.Debugln("Crawling peer ", pi.ID.Pretty()[:16])
		stats.Record(ctx, metrics.WorkersWorkingCount.M(float64(runningWorkers.Inc())))

		cr := CrawlResult{
			WorkerID: w.Identifier(),
			Peer:     pi,
		}

		cr.Error = w.connect(ctx, pi)
		if cr.Error == nil {
			// Extract agent
			if agent, err := w.host.Peerstore().Get(pi.ID, "AgentVersion"); err == nil {
				cr.Agent = agent.(string)
			}
			// Fetch all neighbors
			cr.Neighbors, cr.Error = w.fetchNeighbors(ctx, pi)
		}

		go func(cpi peer.AddrInfo) {
			if err := w.host.Network().ClosePeer(cpi.ID); err != nil {
				log.WithError(err).WithField("targetID", cpi.ID.Pretty()[:16]).Warnln("Could not close connection to peer")
			}
		}(pi)

		w.crawledPeers++
		select {
		case resultsQueue <- cr:
		case <-w.SigShutdown():
			return
		}

		stats.Record(ctx,
			metrics.WorkersWorkingCount.M(float64(runningWorkers.Dec())),
			metrics.PeerCrawlDuration.M(millisSince(start)),
		)

		select {
		case <-w.SigShutdown():
			return
		default:
		}
		logEntry.Debugln("Crawled peer ", pi.ID.Pretty()[:16])
	}
}

// connect strips all private multi addresses in `pi` and establishes a connection to the given peer.
// It also handles metric capturing.
func (w *Worker) connect(ctx context.Context, pi peer.AddrInfo) error {
	start := time.Now()
	stats.Record(ctx, metrics.ConnectsCount.M(1))

	pi = filterPublicMaddrs(pi)
	if len(pi.Addrs) == 0 {
		stats.Record(ctx, metrics.ConnectErrors.M(1))
		return fmt.Errorf("skipping node as it has no public IP address")
	}

	ctx, cancel := context.WithTimeout(ctx, w.config.DialTimeout)
	defer cancel()

	if err := w.host.Connect(ctx, pi); err != nil {
		stats.Record(ctx, metrics.ConnectErrors.M(1))
		return err
	}

	ps := w.host.Peerstore()
	as, _ := ps.Get(pi.ID, "AgentVersion")
	_ = as
	x, _ := ps.GetProtocols(pi.ID)
	_ = x

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
			for _, n := range neighbors {
				allNeighbors = append(allNeighbors, *n)
			}
			allNeighborsLk.Unlock()
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

// filterPublicMaddrs returns only the publicly reachable multiaddrs from the given peer address information.
func filterPublicMaddrs(pi peer.AddrInfo) peer.AddrInfo {
	filtered := peer.AddrInfo{
		ID:    pi.ID,
		Addrs: []ma.Multiaddr{},
	}

	// Just keep public multi addresses
	for _, maddr := range pi.Addrs {
		if !manet.IsPublicAddr(maddr) { // TODO: Strip relays?
			continue
		}
		filtered.Addrs = append(filtered.Addrs, maddr)
	}

	return filtered
}
