package monitor

import (
	"context"
	"time"

	"github.com/dennis-tra/nebula-crawler/pkg/db"
	log "github.com/sirupsen/logrus"
)

func Start(ctx context.Context, dbc *db.Client) {
	for {
		select {
		case <-time.Tick(time.Minute):
			go CheckExpired(ctx, dbc)
		case <-ctx.Done():
			return
		}
	}
}

func CheckExpired(ctx context.Context, dbc *db.Client) {
	peers, err := dbc.ExpiredPeers()
	if err != nil {
		log.WithError(err).Warnln("Error querying expired peers")
		return
	}
	_ = peers
}
