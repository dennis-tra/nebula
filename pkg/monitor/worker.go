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

	// Tracks the timestamp of the first time we couldn't dial the remote peer.
	// Due to retries this could deviate significantly from the time when this
	// result is published.
	FirstFailedDial time.Time

	// If error is set the peer was not dialable
	Error error
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

func (w *Worker) StartDialing(dialQueue <-chan peer.AddrInfo, resultsQueue chan<- Result) {
	w.ServiceStarted()
	defer w.ServiceStopped()

	ctx := w.ServiceContext()
	for pi := range dialQueue {
		start := time.Now()

		// Creating log entry
		logEntry := log.WithFields(log.Fields{
			"workerID": w.Identifier(),
			"targetID": pi.ID.Pretty()[:16],
		})
		logEntry.Debugln("Connecting to peer ", pi.ID.Pretty()[:16])

		// Initialize dial result
		dr := Result{
			WorkerID: w.Identifier(),
			Peer:     pi,
		}

		// Add peer information to peer store so that DialPeer can pick it up from there
		w.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, time.Minute)

		// Try to dial the peer 3 times
		for i := 0; i < 3; i++ {

			// Update log entry
			logEntry = logEntry.WithField("retry", i)

			// Actually dial the peer
			if err := w.dial(ctx, pi.ID); err != nil {
				dr.FirstFailedDial = time.Now()
				dr.Error = err
				sleepDuration := time.Duration(float64(5*(i+1)) * float64(time.Second))
				logEntry.WithError(err).WithField("retry", i).Debugf("Dial failed, sleeping %s\n", sleepDuration)
				time.Sleep(sleepDuration)
				continue
			}

			// Dial was successful - reset error
			dr.Error = nil

			// Close established connection to prevent running out of FDs?
			if err := w.host.Network().ClosePeer(pi.ID); err != nil {
				logEntry.WithError(err).Warnln("Could not close connection to peer")
			}

			// Break out of for loop as the connection was successfully established
			break
		}

		select {
		case resultsQueue <- dr:
		case <-w.SigShutdown():
			return
		}

		logEntry.WithFields(log.Fields{
			"duration": time.Since(start),
			"alive":    dr.Error == nil,
		}).Debugln("Tried to connect to peer", pi.ID.Pretty()[:16])
	}
}

func (w *Worker) dial(ctx context.Context, peerID peer.ID) error {
	start := time.Now()
	stats.Record(ctx, metrics.MonitorDialCount.M(1))

	if _, err := w.host.Network().DialPeer(ctx, peerID); err != nil {
		stats.Record(ctx, metrics.MonitorDialErrorsCount.M(1))
		return err
	}

	stats.Record(ctx, metrics.MonitorDialDuration.M(millisSince(start)))
	return nil
}
