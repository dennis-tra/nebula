package db

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type ClickHouseClientConfig struct {
	DatabaseHost     string
	DatabasePort     int
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseSSL      bool

	// TODO: Populate accordingly
	NetworkID string

	// MeterProvider is the meter provider to use when initialising metric instruments.
	MeterProvider metric.MeterProvider

	// TracerProvider is the tracer provider to use when initialising tracing
	TracerProvider trace.TracerProvider
}

type ClickHouseClient struct {
	conn driver.Conn
	cfg  *ClickHouseClientConfig

	crawlMu sync.Mutex
	crawl   *ClickHouseCrawl
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

	if cfg.DatabaseSSL {
		// TODO: allow skipping CA verification step
		options.TLS = &tls.Config{}
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
		cfg:  cfg,
	}, nil
}

type ClickHouseCrawl struct {
	ID              uuid.UUID  `ch:"id"`
	State           string     `ch:"state"`
	FinishedAt      *time.Time `ch:"finished_at"`
	UpdatedAt       time.Time  `ch:"updated_at"`
	CreatedAt       time.Time  `ch:"created_at"`
	CrawledPeers    *int32     `ch:"crawled_peers"`
	DialablePeers   *int32     `ch:"dialable_peers"`
	UndialablePeers *int32     `ch:"undialable_peers"`
	RemainingPeers  *int32     `ch:"remaining_peers"`
	Version         string     `ch:"version"`
	NetworkID       string     `ch:"network_id"`
}

func (c *ClickHouseClient) InitCrawl(ctx context.Context, version string) error {
	c.crawlMu.Lock()
	defer c.crawlMu.Unlock()

	if c.crawl != nil {
		return fmt.Errorf("crawl already initialized")
	}

	latestCrawl, err := c.selectLatestCrawl(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		// ok
	} else if err != nil {
		return fmt.Errorf("select latest crawl: %w", err)
	} else if latestCrawl.NetworkID != c.cfg.NetworkID {
		return fmt.Errorf("network id mismatch (expected %s, got %s)", c.cfg.NetworkID, latestCrawl.NetworkID)
	} else if latestCrawl.State != string(CrawlStateStarted) {
		log.WithField("id", latestCrawl.ID).Warnln("Another crawl is already running")
	}

	uuidv7, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("new uuid v7: %w", err)
	}

	now := time.Now()

	crawl := &ClickHouseCrawl{
		ID:              uuidv7,
		State:           string(CrawlStateStarted),
		FinishedAt:      nil,
		UpdatedAt:       now,
		CreatedAt:       now,
		CrawledPeers:    nil,
		DialablePeers:   nil,
		UndialablePeers: nil,
		RemainingPeers:  nil,
		Version:         version,
		NetworkID:       c.cfg.NetworkID,
	}

	batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO crawls")
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	if err := batch.AppendStruct(crawl); err != nil {
		return fmt.Errorf("append crawl struct: %w", err)
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("insert crawl: %w", err)
	}

	c.crawl = crawl

	return nil
}

func (c *ClickHouseClient) SealCrawl(ctx context.Context, args *SealCrawlArgs) (err error) {
	c.crawlMu.Lock()
	defer c.crawlMu.Unlock()

	if c.crawl == nil {
		return fmt.Errorf("crawl not initialized")
	}

	// TODO: does this perform a deep copy?
	original := *c.crawl
	defer func() {
		// roll back in case of an error
		if err != nil {
			c.crawl = &original
		}
	}()

	toPtr := func(val int) *int32 {
		cast := int32(val)
		return &cast
	}

	now := time.Now()
	c.crawl.UpdatedAt = now
	c.crawl.CrawledPeers = toPtr(args.Crawled)
	c.crawl.DialablePeers = toPtr(args.Dialable)
	c.crawl.UndialablePeers = toPtr(args.Undialable)
	c.crawl.RemainingPeers = toPtr(args.Remaining)
	c.crawl.State = string(args.State)
	c.crawl.FinishedAt = &now

	batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO crawls")
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	if err := batch.AppendStruct(c.crawl); err != nil {
		return fmt.Errorf("append crawl struct: %w", err)
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("insert crawl: %w", err)
	}

	return nil
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

func (c *ClickHouseClient) Flush(ctx context.Context) error {
	// TODO implement me
	panic("implement me")
}

func (c *ClickHouseClient) Close() error {
	return c.conn.Close()
}

func (c *ClickHouseClient) selectCrawl(ctx context.Context, id uuid.UUID) (*ClickHouseCrawl, error) {
	crawl := &ClickHouseCrawl{}
	err := c.conn.QueryRow(ctx, "SELECT * FROM crawls FINAL WHERE id = ? LIMIT 1", id).ScanStruct(crawl)
	return crawl, err
}

func (c *ClickHouseClient) selectLatestCrawl(ctx context.Context) (*ClickHouseCrawl, error) {
	crawl := &ClickHouseCrawl{}
	err := c.conn.QueryRow(ctx, "SELECT * FROM crawls FINAL ORDER BY created_at desc LIMIT 1").ScanStruct(crawl)
	return crawl, err
}
