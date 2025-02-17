package db

import (
	"context"
	"crypto/tls"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/golang-migrate/migrate/v4"
	mch "github.com/golang-migrate/migrate/v4/database/clickhouse"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

//go:embed migrations/ch
var clickhouseMigrations embed.FS

type ClickHouseClientConfig struct {
	DatabaseHost          string
	DatabasePort          int
	DatabaseName          string
	DatabaseUser          string
	DatabasePassword      string
	DatabaseSSL           bool
	ClusterName           string        // TODO: plumb
	MigrationsTableEngine string        // TODO: plumb
	ApplyMigrations       bool          // TODO: plumb
	BatchSize             int           // TODO: plumb
	BatchTimeout          time.Duration // TODO: plumb
	NetworkID             string        // TODO: plumb

	// MeterProvider is the meter provider to use when initialising metric instruments.
	MeterProvider metric.MeterProvider

	// TracerProvider is the tracer provider to use when initialising tracing
	TracerProvider trace.TracerProvider
}

func (cfg *ClickHouseClientConfig) Options() *clickhouse.Options {
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

	return options
}

type ClickHouseClient struct {
	conn driver.Conn
	cfg  *ClickHouseClientConfig

	crawlMu sync.Mutex
	crawl   *ClickHouseCrawl

	// this channel will receive all new visits information
	visitsChan chan *VisitArgs

	// this channel can be used to trigger a flush of the prepared batches
	flushChan chan struct{}

	// this channel is closed when the flusher has exited
	flusherDone chan struct{}

	// this cancel func can be used to forcefully stop the flusher.
	flusherCancel context.CancelFunc
}

func NewClickHouseClient(ctx context.Context, cfg *ClickHouseClientConfig) (*ClickHouseClient, error) {
	conn, err := clickhouse.Open(cfg.Options())
	if err != nil {
		return nil, fmt.Errorf("open clickhouse database: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping clickhouse database: %w", err)
	}

	flusherCtx, flusherCancel := context.WithCancel(context.Background())

	client := &ClickHouseClient{
		conn:          conn,
		cfg:           cfg,
		flusherCancel: flusherCancel,
		visitsChan:    make(chan *VisitArgs),
		flushChan:     make(chan struct{}),
		flusherDone:   make(chan struct{}),
	}

	if cfg.ApplyMigrations {
		if err = client.applyMigration(); err != nil {
			return nil, fmt.Errorf("apply migrations: %w", err)
		}
	}

	go client.startFlusher(flusherCtx)

	return client, nil
}

func (c *ClickHouseClient) applyMigration() error {
	// load clickhouse migrations files
	migrationsDir, err := iofs.New(clickhouseMigrations, "migrations/ch")
	if err != nil {
		return fmt.Errorf("create iofs migrations source: %w", err)
	}

	db := clickhouse.OpenDB(c.cfg.Options())

	migrationsDriver, err := mch.WithInstance(db, &mch.Config{
		DatabaseName:          c.cfg.DatabaseName,
		ClusterName:           c.cfg.ClusterName,
		MigrationsTableEngine: c.cfg.MigrationsTableEngine,
	})
	if err != nil {
		return fmt.Errorf("create migrate driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", migrationsDir, c.cfg.DatabaseName, migrationsDriver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate database: %w", err)
	}

	return nil
}

func (c *ClickHouseClient) startFlusher(ctx context.Context) {
	defer close(c.flusherDone)

	// convenience function to send a batch to clickhouse with error logging
	send := func(batch driver.Batch, table string) {
		if err := batch.Send(); err != nil {
			log.WithError(err).WithField("rows", batch.Rows()).Errorln("Failed to send " + table + " batch")
		}
	}

	// convenience function to prepare a batch to be sent to clickhouse with
	// error logging
	prepare := func(table string) driver.Batch {
		batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO "+table)
		if err != nil {
			log.WithError(err).Errorln("Failed to prepare " + table + " batch")
		}
		return batch
	}

	// prepare visits and neighbors batches
	visitsBatch := prepare("visits")
	neighborsBatch := prepare("neighbors")

	// create a ticker that triggers a write of both batches to clickhouse
	ticker := time.NewTicker(c.cfg.BatchTimeout)
	defer ticker.Stop()

	for {

		// prepare new batches if they were sent in the previous iteration
		if visitsBatch.IsSent() {
			visitsBatch = prepare("visits")
		}

		if neighborsBatch.IsSent() {
			neighborsBatch = prepare("neighbors")
		}

		// if any of the above batch preparations failed they will be null.
		// Because the flusher here works asynchronously we can't easily signal
		// back to the main routine that there was an error. Therefore, we just
		// log an error here and consume and discard all events until done.
		// This if-statement is part of the for-loop because later batch
		// preparations could also fail.
		if visitsBatch == nil || neighborsBatch == nil {
			log.Errorln("Failed to prepare visits or neighbors batch. Discarding all events until done.")
			select {
			case <-ctx.Done():
				return
			case _, done := <-c.visitsChan:
				if done {
					return
				}
				continue
			}
		}

		select {
		case <-ctx.Done():
			// don't send anything as the context was canceled. The context is
			// part of the batches. Sending them wouldn't work because of that.
			return
		case <-ticker.C:
			// sending batches to clickhouse because of a timeout
			send(visitsBatch, "visits")
			send(neighborsBatch, "neighbors")

		case <-c.flushChan:
			// sending batches to clickhouse because the user asked for it
			send(visitsBatch, "visits")
			send(neighborsBatch, "neighbors")

		case visitArgs, more := <-c.visitsChan:
			if !more {
				// we won't receive any more visits
				send(visitsBatch, "visits")
				send(neighborsBatch, "neighbors")
				return
			}

			// the crawl can be null if it's a visit from the monitoring task
			var crawlID *string
			if c.crawl != nil {
				id := c.crawl.ID.String()
				crawlID = &id
			}

			if err := visitsBatch.Append(
				crawlID,
				visitArgs.PeerID,
				visitArgs.AgentVersion,
				visitArgs.Protocols,
				visitArgs.VisitType,
				visitArgs.Maddrs,
				visitArgs.ConnectMaddr,
				[]string{},
				visitArgs.CrawlErrorStr,
				visitArgs.VisitStartedAt,
				visitArgs.VisitEndedAt,
			); err != nil {
				log.WithError(err).Errorln("Failed to append visits batch")
			}

			// check if we have enough data to submit the batch
			if visitsBatch.Rows() >= c.cfg.BatchSize {
				send(visitsBatch, "visits")
			}

			// neighbors are only set during a crawl, so if this is a visit
			// from the monitoring task continue here.
			if crawlID == nil {
				continue
			}

			// append neighbors to batch
			for _, neighbor := range visitArgs.Neighbors {
				if err := neighborsBatch.Append(crawlID, visitArgs.PeerID, neighbor, visitArgs.ErrorBits); err != nil {
					log.WithError(err).Errorln("Failed to append visits batch")
				}
			}

			// check if we have enough data to submit the batch
			if neighborsBatch.Rows() >= c.cfg.BatchSize {
				send(neighborsBatch, "neighbors")
			}
		}
	}
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

	// Use Batch because of the convenience of AppendStruct
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
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.visitsChan <- args:
		return nil
	}
}

func (c *ClickHouseClient) InsertCrawlProperties(ctx context.Context, properties map[string]map[string]int) error {
	return nil
}

func (c *ClickHouseClient) InsertNeighbors(ctx context.Context, peerID peer.ID, neighbors []peer.ID, errorBits uint16) error {
	c.crawlMu.Lock()
	defer c.crawlMu.Unlock()
	if c.crawl == nil {
		return fmt.Errorf("crawl not initialized")
	}

	query := "INSERT INTO neighbors (crawl_id, peer_id, neighbor, error_bits) VALUES (?, ?, ?, ?)"

	return c.conn.Exec(ctx, query, c.crawl.ID, peerID, neighbors, errorBits)
}

func (c *ClickHouseClient) SelectPeersToProbe(ctx context.Context) ([]peer.AddrInfo, error) {
	// TODO implement me
	panic("implement me")
}

func (c *ClickHouseClient) Flush(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.flushChan <- struct{}{}:
		return nil
	}
}

func (c *ClickHouseClient) Close() error {
	close(c.visitsChan)

	select {
	case <-c.flusherDone:
	case <-time.After(5 * time.Second):
		log.Warnln("Flusher did not finish in time, forcing shutdown")
		c.flusherCancel()
		<-c.flusherDone
	}

	close(c.flushChan)

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
