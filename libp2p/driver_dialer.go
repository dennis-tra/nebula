package libp2p

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/host"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
)

type DialDriverConfig struct {
	Version     string
	DialTimeout time.Duration
}

type DialDriver struct {
	cfg         *DialDriverConfig
	host        host.Host
	dbc         db.Client
	taskQueue   chan PeerInfo
	start       chan struct{}
	shutdown    chan struct{}
	done        chan struct{}
	dialerCount int
	writerCount int
}

var _ core.Driver[PeerInfo, core.DialResult[PeerInfo]] = (*DialDriver)(nil)

func NewDialDriver(dbc db.Client, cfg *DialDriverConfig) (*DialDriver, error) {
	// Configure the resource manager to not limit anything
	limiter := rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits)
	rm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		return nil, fmt.Errorf("new resource manager: %w", err)
	}

	// Don't use a connection manager that could potentially
	// prune any connections. We _theoretically_ clean up after
	// ourselves.
	cm := connmgr.NullConnMgr{}

	// Initialize a single libp2p node that's shared between all dialers.
	h, err := libp2p.New(
		libp2p.UserAgent("nebula/"+cfg.Version),
		libp2p.ResourceManager(rm),
		libp2p.ConnectionManager(cm),
		libp2p.DisableMetrics(),
		libp2p.EnableRelay(), // enable the relay transport
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
		id:      fmt.Sprintf("dialer-%02d", d.dialerCount),
		host:    d.host,
		timeout: d.cfg.DialTimeout,
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
		log.Infof("Looking for peers to probe...")
		addrInfos, err := d.dbc.SelectPeersToProbe(ctx)
		if errors.Is(err, sql.ErrNoRows) || len(addrInfos) == 0 {
			log.Infoln("No peers due to be probed")
		} else if err != nil {
			log.WithError(err).Warnln("Could not fetch sessions")
			goto TICK
		}

		for _, addrInfo := range addrInfos {
			select {
			case d.taskQueue <- PeerInfo{AddrInfo: addrInfo}:
				continue
			case <-ctx.Done():
				// fallthrough
			}
			break
		}

	TICK:
		select {
		case <-time.Tick(10 * time.Second):
			continue
		case <-ctx.Done():
			return
		}
	}
}
