package db

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/volatiletech/null/v8"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

type Client interface {
	io.Closer
	InitCrawl(ctx context.Context) (*models.Crawl, error)
	UpdateCrawl(ctx context.Context, crawl *models.Crawl) error
	PersistCrawlProperties(ctx context.Context, crawl *models.Crawl, properties map[string]map[string]int) error
	PersistCrawlVisit(ctx context.Context, crawlID int, peerID peer.ID, maddrs []ma.Multiaddr, protocols []string, agentVersion string, connectDuration time.Duration, crawlDuration time.Duration, visitStartedAt time.Time, visitEndedAt time.Time, connectErrorStr string, crawlErrorStr string, properties null.JSON) (*InsertVisitResult, error)
	PersistNeighbors(ctx context.Context, crawl *models.Crawl, dbPeerID *int, peerID peer.ID, errorBits uint16, dbNeighborsIDs []int, neighbors []peer.ID) error
}

// NewClient will initialize the right database client based on the given
// configuration. This can either be a Postgres, JSON, or noop client. The noop
// client is a dummy implementation of the [Client] interface that does nothing
// when the methods are called. That's the one used if the user specifies
// `--dry-run` on the command line. The JSON client is used when the user
// specifies a JSON output directory. Then JSON files with crawl information
// are written to that directory. In any other case, the Postgres client is
// used.
func NewClient(ctx context.Context, cfg *config.Database) (Client, error) {
	var (
		dbc Client
		err error
	)

	// dry run has presedence. Then, if a JSON output directory is given, use
	// the JSON client. In any other case, use the Postgres database client.
	if cfg.DryRun {
		dbc = InitNoopClient()
	} else if cfg.JSONOut != "" {
		dbc, err = InitJSONClient(cfg.JSONOut)
	} else {
		dbc, err = InitDBClient(ctx, cfg)
	}
	if err != nil {
		return nil, fmt.Errorf("init db client: %w", err)
	}

	return dbc, nil
}
