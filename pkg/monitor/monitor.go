package monitor

import (
	"context"
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/core"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/discv5"
	"github.com/dennis-tra/nebula-crawler/pkg/libp2p"
)

type Monitor struct {
	cfg *config.Monitor
	dbc *db.DBClient
}

func New(dbc *db.DBClient, cfg *config.Monitor) (*Monitor, error) {
	return &Monitor{
		cfg: cfg,
		dbc: dbc,
	}, nil
}

func (m *Monitor) MonitorNetwork(ctx context.Context) error {
	handlerCfg := &core.DialHandlerConfig{}

	engineCfg := &core.EngineConfig{
		WorkerCount:         m.cfg.MonitorWorkerCount,
		WriterCount:         m.cfg.WriteWorkerCount,
		Limit:               0,
		DuplicateProcessing: true,
	}

	switch m.cfg.Network {
	case string(config.NetworkEthCons):
		driverCfg := &discv5.DialDriverConfig{
			Version: m.cfg.Root.Version(),
		}

		driver, err := discv5.NewDialDriver(m.dbc, driverCfg)
		if err != nil {
			return fmt.Errorf("new driver: %w", err)
		}

		handler := core.NewDialHandler[discv5.PeerInfo](handlerCfg)
		eng, err := core.NewEngine[discv5.PeerInfo, core.DialResult[discv5.PeerInfo]](driver, handler, engineCfg)
		if err != nil {
			return fmt.Errorf("new engine: %w", err)
		}

		_, err = eng.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("running crawl engine: %w", err)
		}

	default:
		driverCfg := &libp2p.DialDriverConfig{
			Version: m.cfg.Root.Version(),
		}

		driver, err := libp2p.NewDialDriver(m.dbc, driverCfg)
		if err != nil {
			return fmt.Errorf("new driver: %w", err)
		}

		handler := core.NewDialHandler[libp2p.PeerInfo](handlerCfg)
		eng, err := core.NewEngine[libp2p.PeerInfo, core.DialResult[libp2p.PeerInfo]](driver, handler, engineCfg)
		if err != nil {
			return fmt.Errorf("new engine: %w", err)
		}

		// Set the timeout for dialing peers
		ctx = network.WithDialPeerTimeout(ctx, m.cfg.Root.DialTimeout)

		// Force direct dials will prevent swarm to run into dial backoff
		// errors. It also prevents proxied connections.
		ctx = network.WithForceDirectDial(ctx, "prevent backoff")

		_, err = eng.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("running crawl engine: %w", err)
		}
	}
	return nil
}
