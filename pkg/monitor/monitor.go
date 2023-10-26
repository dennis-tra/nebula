package monitor

import (
	"context"
	"database/sql"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
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
	switch m.cfg.Network {
	case string(config.NetworkEthereum):
	default:
	}
	return nil
}

// monitorDatabase checks every 10 seconds if there are peer sessions that are due to be renewed.
func (m *Monitor) monitorDatabase(ctx context.Context) {
	for {
		log.Infof("Looking for sessions to check...")
		sessions, err := m.dbc.FetchDueOpenSessions(ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.WithError(err).Warnln("Could not fetch sessions")
			goto TICK
		}

		// For every session schedule that it gets pushed into the dialQueue
		for _, session := range sessions {
			if err = s.scheduleDial(ctx, session); err != nil {
				log.WithError(err).Warnln("Could not schedule dial")
			}
		}
		log.Infof("In dial queue %d peers", s.inDialQueueCount.Load())

	TICK:
		select {
		case <-time.Tick(10 * time.Second):
			continue
		case <-ctx.Done():
			return
		}
	}
}
