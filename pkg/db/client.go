package db

import (
	"context"
	"fmt"
	"time"

	"github.com/dennis-tra/nebula-crawler/pkg/config"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/volatiletech/null/v8"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

type Client interface {
	QueryBootstrapPeers(ctx context.Context, limit int) ([]peer.AddrInfo, error)
	InitCrawl(ctx context.Context) (*models.Crawl, error)
	UpdateCrawl(ctx context.Context, crawl *models.Crawl) error
	PersistCrawlProperties(ctx context.Context, crawl *models.Crawl, properties map[string]map[string]int) error
	PersistCrawlVisit(ctx context.Context, crawlID int, peerID peer.ID, maddrs []ma.Multiaddr, protocols []string, agentVersion string, connectDuration time.Duration, crawlDuration time.Duration, visitStartedAt time.Time, visitEndedAt time.Time, connectErrorStr string, crawlErrorStr string, isExposed null.Bool) (*InsertVisitResult, error)
	PersistNeighbors(ctx context.Context, crawl *models.Crawl, dbPeerID *int, peerID peer.ID, errorBits uint16, dbNeighborsIDs []int, neighbors []peer.ID) error
}

func NewClient(ctx context.Context, cfg *config.Database) (Client, error) {
	// Acquire database handle
	var (
		dbc Client
		err error
	)

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
