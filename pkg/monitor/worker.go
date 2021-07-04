package monitor

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"

	"github.com/go-ping/ping"

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

// PingResult captures data that is gathered from pinging a single peer.
type PingResult struct {
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

func (w *Worker) StartPinging(pingQueue chan peer.AddrInfo, resultsQueue chan PingResult) {
	w.ServiceStarted()
	defer w.ServiceStopped()

	ctx := w.ServiceContext()
	for pi := range pingQueue {
		start := time.Now()
		logEntry := log.WithField("targetID", pi.ID.Pretty()[:16]).WithField("workerID", w.Identifier())
		logEntry.Debugln("Pinging peer ", pi.ID.Pretty()[:16])

		pr := PingResult{
			WorkerID: w.Identifier(),
			Peer:     pi,
		}

		// Parse addresses we can ping - using map as there can be multiple maddrs with same IP but different transports
		hostAddrs := map[string]ma.Multiaddr{}
		for _, maddr := range pi.Addrs {
			addr, err := w.getHostAddr(maddr)
			if err != nil {
				logEntry.WithError(err).WithField("maddr", maddr.String()).Debugln("Could not parse host address")
				continue
			}
			hostAddrs[addr] = maddr
		}

		// Loop through all addresses and try to ping them.
		for addr := range hostAddrs {
			pinger, err := ping.NewPinger(addr)
			if err != nil {
				logEntry.WithError(err).WithField("addr", addr).Debugln("Could not init pinger")
				continue
			}

			done := make(chan struct{})
			go func() {
				select {
				case <-done:
				case <-w.SigShutdown():
					pinger.Stop()
				}
			}()

			pinger.Timeout = w.config.DialTimeout
			pinger.Count = 1

			logEntry.WithField("addr", addr).Debugln("Pinging peer")
			if err = pinger.Run(); err != nil {
				logEntry.WithError(err).WithField("addr", addr).Debugln("Error running pinger")
				close(done)
				continue
			}
			close(done)

			if pinger.PacketsRecv > 0 {
				pr.Alive = true
				break
			}
			logEntry.WithField("addr", addr).Debugln("Peer not responding to ping, trying proper connect")

			ctx, cancel := context.WithTimeout(ctx, w.config.DialTimeout)
			if err = w.host.Connect(ctx, pi); err == nil {
				pr.Alive = true
				stats.Record(ctx, metrics.PingBlockedDialSuccessCount.M(1))
				go func(cpi peer.AddrInfo) {
					if err := w.host.Network().ClosePeer(cpi.ID); err != nil {
						log.WithError(err).WithField("targetID", cpi.ID.Pretty()[:16]).Warnln("Could not close connection to peer")
					}
				}(pi)
				cancel()
				break
			}
			cancel()
			logEntry.WithField("addr", addr).Debugln("Also cannot properly connect to peer")
		}

		select {
		case resultsQueue <- pr:
		case <-w.SigShutdown():
			return
		}

		stats.Record(ctx, metrics.PeerPingDuration.M(millisSince(start)))

		select {
		case <-w.SigShutdown():
			return
		default:
		}
		logEntry.Debugln("Pinged peer", pi.ID.Pretty()[:16])
	}
}

func (w *Worker) getHostAddr(maddr ma.Multiaddr) (string, error) {
	protocols := []int{
		ma.P_IP4,
		ma.P_IP6,
		ma.P_DNS,
		ma.P_DNS4,
		ma.P_DNS6,
		ma.P_DNSADDR,
	}

	for _, p := range protocols {
		if addr, err := maddr.ValueForProtocol(p); err == nil {
			return addr, nil
		}
	}

	return "", errors.New("could not parse host address")
}
