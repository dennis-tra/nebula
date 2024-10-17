package db

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/dennis-tra/nebula-crawler/db/models"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/volatiletech/null/v8"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseClient struct{}

var _ Client = (*ClickHouseClient)(nil)

func InitClickHouseClient(ctx context.Context) (*ClickHouseClient, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"<CLICKHOUSE_SECURE_NATIVE_HOSTNAME>:9440"},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "<DEFAULT_USER_PASSWORD>",
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "an-example-go-client", Version: "0.1"},
			},
		},

		Debugf: func(format string, v ...interface{}) {
			fmt.Printf(format, v)
		},
		TLS: &tls.Config{
			InsecureSkipVerify: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}
}

func (c *ClickHouseClient) Close() error {
}

func (c *ClickHouseClient) InitCrawl(ctx context.Context, version string) (*models.Crawl, error) {
}

func (c *ClickHouseClient) UpdateCrawl(ctx context.Context, crawl *models.Crawl) error {
}

func (c *ClickHouseClient) QueryBootstrapPeers(ctx context.Context, limit int) ([]peer.AddrInfo, error) {
}

func (c *ClickHouseClient) PersistCrawlProperties(ctx context.Context, crawl *models.Crawl, properties map[string]map[string]int) error {
}

func (c *ClickHouseClient) PersistCrawlVisit(ctx context.Context, crawlID int, peerID peer.ID, maddrs []ma.Multiaddr, protocols []string, agentVersion string, connectDuration time.Duration, crawlDuration time.Duration, visitStartedAt time.Time, visitEndedAt time.Time, connectErrorStr string, crawlErrorStr string, properties null.JSON) (*InsertVisitResult, error) {
}

func (c *ClickHouseClient) PersistNeighbors(ctx context.Context, crawl *models.Crawl, dbPeerID *int, peerID peer.ID, errorBits uint16, dbNeighborsIDs []int, neighbors []peer.ID) error {
}
