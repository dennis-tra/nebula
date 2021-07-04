package monitor

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/stats"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
)

var workerID = atomic.NewInt32(0)

// Result captures data that is gathered from pinging a single peer.
type Result struct {
	WorkerID string

	// The pinged peer
	Peer peer.AddrInfo

	// Whether the pinged peer is alive
	Alive bool
}

// Worker encapsulates a libp2p host that crawls the network.
type Worker struct {
	*service.Service

	host   host.Host
	dbh    *sql.DB
	config *config.Config
}

func NewWorker(dbh *sql.DB, h host.Host, conf *config.Config) (*Worker, error) {
	c := &Worker{
		Service: service.New(fmt.Sprintf("worker-%02d", workerID.Load())),
		host:    h,
		dbh:     dbh,
		config:  conf,
	}
	workerID.Inc()

	return c, nil
}

func millisSince(start time.Time) float64 {
	return float64(time.Since(start)) / float64(time.Millisecond)
}

func (w *Worker) StartConnecting(connectQueue chan peer.AddrInfo, resultsQueue chan Result) {
	w.ServiceStarted()
	defer w.ServiceStopped()

	ctx := w.ServiceContext()
	for pi := range connectQueue {
		logEntry := log.WithField("targetID", pi.ID.Pretty()[:16]).WithField("workerID", w.Identifier())
		logEntry.Debugln("Connecting to peer ", pi.ID.Pretty()[:16])

		pr := Result{
			WorkerID: w.Identifier(),
			Peer:     pi,
		}

		if err := w.connect(ctx, pi); err == nil {
			pr.Alive = true
			if err := w.host.Network().ClosePeer(pi.ID); err != nil {
				log.WithError(err).WithField("targetID", pi.ID.Pretty()[:16]).Warnln("Could not close connection to peer")
			}
		}

		select {
		case resultsQueue <- pr:
		case <-w.SigShutdown():
			return
		}

		logEntry.Debugln("Connected to peer", pi.ID.Pretty()[:16])
	}
}

// connect strips all private multi addresses in `pi` and establishes a connection to the given peer.
// It also handles metric capturing.
func (w *Worker) connect(ctx context.Context, pi peer.AddrInfo) error {
	start := time.Now()
	stats.Record(ctx, metrics.MonitorConnectsCount.M(1))

	ctx, cancel := context.WithTimeout(ctx, w.config.DialTimeout)
	defer cancel()

	if err := w.host.Connect(ctx, pi); err != nil {
		stats.Record(ctx, metrics.MonitorConnectErrorsCount.M(1))
		return err
	}

	stats.Record(w.ServiceContext(), metrics.MonitorConnectDuration.M(millisSince(start)))
	return nil
}
