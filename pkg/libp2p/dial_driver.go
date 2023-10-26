package libp2p

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	log "github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"

	"github.com/dennis-tra/nebula-crawler/pkg/core"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
)

type DialDriverConfig struct {
	Version string
}

type DialDriver struct {
	cfg         *DialDriverConfig
	host        host.Host
	dbc         *db.DBClient
	taskQueue   chan PeerInfo
	start       chan struct{}
	shutdown    chan struct{}
	done        chan struct{}
	dialerCount int
	writerCount int
}

var _ core.Driver[PeerInfo, core.DialResult[PeerInfo]] = (*DialDriver)(nil)

func NewDialDriver(dbc *db.DBClient, cfg *DialDriverConfig) (*DialDriver, error) {
	// Configure the resource manager to not limit anything
	limiter := rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits)
	rm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		return nil, fmt.Errorf("new resource manager: %w", err)
	}

	// Initialize a single libp2p node that's shared between all dialers.
	h, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.ResourceManager(rm),
		libp2p.UserAgent("nebula/"+cfg.Version),
	)
	if err != nil {
		return nil, fmt.Errorf("new libp2p host: %w", err)
	}

	d := &DialDriver{
		cfg:       cfg,
		host:      h,
		dbc:       dbc,
		taskQueue: make(chan PeerInfo),
		start:     make(chan struct{}),
		shutdown:  make(chan struct{}),
		done:      make(chan struct{}),
	}

	go d.monitorDatabase()

	return d, nil
}

func (d *DialDriver) NewWorker() (core.Worker[PeerInfo, core.DialResult[PeerInfo]], error) {
	dialer := &Dialer{
		id:   fmt.Sprintf("dialer-%02d", d.dialerCount),
		host: d.host,
	}

	d.dialerCount += 1

	return dialer, nil
}

func (d *DialDriver) NewWriter() (core.Worker[core.DialResult[PeerInfo], core.WriteResult], error) {
	id := fmt.Sprintf("writer-%02d", d.writerCount)
	w := core.NewDialWriter[PeerInfo](id, d.dbc)
	d.writerCount += 1
	return w, nil
}

func (d *DialDriver) Tasks() <-chan PeerInfo {
	close(d.start)
	return d.taskQueue
}

func (d *DialDriver) Close() {
	close(d.shutdown)
	<-d.done
	close(d.taskQueue)

	if err := d.host.Close(); err != nil {
		log.WithError(err).Warnln("Error closing libp2p host")
	}
}

// monitorDatabase checks every 10 seconds if there are peer sessions that are due to be renewed.
func (d *DialDriver) monitorDatabase() {
	defer close(d.done)

	select {
	case <-d.start:
	case <-d.shutdown:
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-d.shutdown
		cancel()
	}()

	for {
		log.Infof("Looking for sessions to check...")
		sessions, err := d.dbc.FetchDueOpenSessions(ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.WithError(err).Warnln("Could not fetch sessions")
			goto TICK
		}

		for _, session := range sessions {
			peerID, err := peer.Decode(session.R.Peer.MultiHash)
			if err != nil {
				log.WithField("mhash", session.R.Peer.MultiHash).
					WithError(err).
					Warnln("Could not parse multi address")
				continue
			}
			logEntry := log.WithField("peerID", peerID.ShortString())

			// Parse multi addresses from database
			addrInfo := peer.AddrInfo{ID: peerID}
			for _, maddrStr := range session.R.Peer.R.MultiAddresses {
				maddr, err := ma.NewMultiaddr(maddrStr.Maddr)
				if err != nil {
					logEntry.WithError(err).Warnln("Could not parse multi address")
					continue
				}
				addrInfo.Addrs = append(addrInfo.Addrs, maddr)
			}

			select {
			case d.taskQueue <- PeerInfo{AddrInfo: addrInfo}:
				continue
			case <-ctx.Done():
				// fallthrough
			}
			break
		}
		// log.Infof("In dial queue %d peers", s.inDialQueueCount.Load())

	TICK:
		select {
		case <-time.Tick(10 * time.Second):
			continue
		case <-ctx.Done():
			return
		}
	}
}
