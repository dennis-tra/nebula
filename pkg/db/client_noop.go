package db

import (
	"context"
	"time"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/volatiletech/null/v8"
)

type NoopClient struct{}

var _ Client = (*NoopClient)(nil)

func InitNoopClient() *NoopClient {
	return &NoopClient{}
}

func (n *NoopClient) InitCrawl(ctx context.Context) (*models.Crawl, error) {
	return &models.Crawl{
		ID:        1,
		StartedAt: time.Now(),
		State:     models.CrawlStateStarted,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}, nil
}

func (n *NoopClient) UpdateCrawl(ctx context.Context, crawl *models.Crawl) error {
	crawl.UpdatedAt = time.Now()
	return nil
}

func (n *NoopClient) PersistCrawlProperties(ctx context.Context, crawl *models.Crawl, properties map[string]map[string]int) error {
	return nil
}

func (n *NoopClient) PersistCrawlVisit(ctx context.Context, crawlID int, peerID peer.ID, maddrs []ma.Multiaddr, protocols []string, agentVersion string, connectDuration time.Duration, crawlDuration time.Duration, visitStartedAt time.Time, visitEndedAt time.Time, connectErrorStr string, crawlErrorStr string, properties null.JSON) (*InsertVisitResult, error) {
	return &InsertVisitResult{PID: peerID}, nil
}

func (n *NoopClient) PersistNeighbors(ctx context.Context, crawl *models.Crawl, dbPeerID *int, peerID peer.ID, errorBits uint16, dbNeighborsIDs []int, neighbors []peer.ID) error {
	return nil
}

func (n *NoopClient) Close() error {
	return nil
}
