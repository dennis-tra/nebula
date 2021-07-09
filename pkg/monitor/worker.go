package monitor

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/stats"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
)

var workerID = atomic.NewInt32(0)

// Result captures data that is gathered from pinging a single peer.
type Result struct {
	WorkerID string

	// The pinged peer
	Peer peer.AddrInfo

	// If error is set the peer was not dialable
	Error error

	// As it can take some time to handle the result we track the timestamp explicitly
	// Tracks the timestamp of the first time we couldn't dial the remote peer.
	// Due to retries this could deviate significantly from the time when this
	// result is published.
	ErrorTime time.Time
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

		// Try to dial the peer 3 times
	retryLoop:
		for retry := 0; retry < 3; retry++ {

			// Update log entry
			logEntry = logEntry.WithField("retry", retry)

			// Add peer information to peer store so that DialPeer can pick it up from there
			// Do this in every retry due to the TTL of one minute
			w.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, time.Minute)

			// Actually dial the peer
			if err := w.dial(ctx, pi.ID); err != nil {
				dr.ErrorTime = time.Now()
				dr.Error = err

				if errors.Is(err, context.Canceled) {
					break retryLoop
				}

				sleepDur := time.Duration(float64(10*(retry+1)) * float64(time.Second))
				errMsg := fmt.Sprintf("Dial failed, sleeping %s", sleepDur)

				switch determineDialError(dr.Error) {
				case models.DialErrorPeerIDMismatch:
					logEntry.WithError(err).Debugln("Dial failed due peer ID mismatch - stopping retry")
					// TODO: properly connect to new peer and see if it is part of the DHT.
					break retryLoop
				case models.DialErrorNoPublicIP, models.DialErrorNoGoodAddresses:
					logEntry.WithError(err).Debugln("Dial failed due to no public ip - stopping retry")
					break retryLoop
				case models.DialErrorMaxDialAttemptsExceeded:
					sleepDur = 70 * time.Second
					errMsg = fmt.Sprintf("Max dial attempts exceeded, sleeping longer %s", sleepDur)
				case models.DialErrorConnectionRefused:
					// The monitoring task receives a lot of "connection refused" messages. I guess there is
					// a limit somewhere of how often a peer can connect. I could imagine that this rate limiting
					// is set to one minute. As the scheduler fetches all sessions that are due in the next 10
					// seconds I'll add that and another one just to be sure ¯\_(ツ)_/¯
					if retry >= 1 {
						logEntry.WithError(err).Debugf("Received 'connection refused' the second time - stopping retry")
						break retryLoop
					}
					sleepDur = 70 * time.Second
					errMsg = fmt.Sprintf("Connection refused, sleeping longer %s", sleepDur)
				default:
				}
				logEntry.WithError(err).Debugf(errMsg)
				time.Sleep(sleepDur)
				continue retryLoop
			}

			// Dial was successful - reset error
			dr.Error = nil
			dr.ErrorTime = time.Time{}

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
