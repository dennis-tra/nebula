package db

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type ClickHouseClientConfig struct {
	DatabaseHost     string
	DatabasePort     int
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseSSL      string

	// MeterProvider is the meter provider to use when initialising metric instruments.
	MeterProvider metric.MeterProvider

	// TracerProvider is the tracer provider to use when initialising tracing
	TracerProvider trace.TracerProvider
}

type ClickHouseClient struct {
	conn driver.Conn
}

func NewClickHouseClient(ctx context.Context, cfg *ClickHouseClientConfig) (*ClickHouseClient, error) {
	options := &clickhouse.Options{
		Addr: []string{net.JoinHostPort(cfg.DatabaseHost, strconv.Itoa(cfg.DatabasePort))},
		Auth: clickhouse.Auth{
			Database: cfg.DatabaseName,
			Username: cfg.DatabaseUser,
			Password: cfg.DatabasePassword,
		},
	}

	switch strings.ToLower(cfg.DatabaseSSL) {
	case "yes", "true", "1":
		options.TLS = &tls.Config{
			// TODO: allow skipping CA verification step
		}
	}

	conn, err := clickhouse.Open(options)
	if err != nil {
		return nil, fmt.Errorf("open clickhouse database: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping clickhouse database: %w", err)
	}

	return &ClickHouseClient{
		conn: conn,
	}, nil
}

func (c *ClickHouseClient) InitCrawl(ctx context.Context, version string) error {
	// TODO implement me
	panic("implement me")
}

func (c *ClickHouseClient) SealCrawl(ctx context.Context, args *SealCrawlArgs) error {
	// TODO implement me
	panic("implement me")
}

func (c *ClickHouseClient) QueryBootstrapPeers(ctx context.Context, limit int) ([]peer.AddrInfo, error) {
	// TODO implement me
	panic("implement me")
}

func (c *ClickHouseClient) InsertVisit(ctx context.Context, args *VisitArgs) error {
	// TODO implement me
	panic("implement me")
}

func (c *ClickHouseClient) InsertCrawlProperties(ctx context.Context, properties map[string]map[string]int) error {
	// TODO implement me
	panic("implement me")
}

func (c *ClickHouseClient) InsertNeighbors(ctx context.Context, peerID peer.ID, neighbors []peer.ID, errorBits uint16) error {
	// TODO implement me
	panic("implement me")
}

func (c *ClickHouseClient) SelectPeersToProbe(ctx context.Context) ([]peer.AddrInfo, error) {
	// TODO implement me
	panic("implement me")
}

func (c *ClickHouseClient) Close() error {
	// TODO implement me
	panic("implement me")
}
