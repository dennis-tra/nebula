package db

import (
	"context"
	"crypto/tls"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net"
	"sort"
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

	"github.com/dennis-tra/nebula-crawler/utils"
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
	ClusterName           string
	MigrationsTableEngine string
	ApplyMigrations       bool
	BatchSize             int
	BatchTimeout          time.Duration
	NetworkID             string

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

	crawlMu sync.RWMutex
	crawl   *ClickHouseCrawl

	// this channel will receive all new visits
	visitsChan chan *ClickHouseVisit

	// this channel can be used to trigger a flush of the prepared batches
	// the channel in the channel will be closed when the flush is done.
	flushChan chan chan struct{}

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
		visitsChan:    make(chan *ClickHouseVisit),
		flushChan:     make(chan chan struct{}),
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
	migrationsDir, err := iofs.New(clickhouseMigrations, "migrations/ch")
	if err != nil {
		return fmt.Errorf("create iofs migrations source: %w", err)
	}

	db := clickhouse.OpenDB(c.cfg.Options())

	mdriver, err := mch.WithInstance(db, &mch.Config{
		DatabaseName:          c.cfg.DatabaseName,
		ClusterName:           c.cfg.ClusterName,
		MigrationsTableEngine: c.cfg.MigrationsTableEngine,
	})
	if err != nil {
		return fmt.Errorf("create migrate driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", migrationsDir, c.cfg.DatabaseName, mdriver)
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

		case doneChan := <-c.flushChan:
			// sending batches to clickhouse because the user asked for it
			send(visitsBatch, "visits")
			send(neighborsBatch, "neighbors")
			close(doneChan)

		case visit, more := <-c.visitsChan:
			if !more {
				// we won't receive any more visits
				send(visitsBatch, "visits")
				send(neighborsBatch, "neighbors")
				return
			}

			if err := visitsBatch.AppendStruct(visit); err != nil {
				log.WithError(err).Errorln("Failed to append visits batch")
			}

			// check if we have enough data to submit the batch
			if visitsBatch.Rows() >= c.cfg.BatchSize {
				send(visitsBatch, "visits")
			}

			// neighbors are only set during a crawl, so if this is a visit
			// from the monitoring task continue here.
			if visit.CrawlID == nil {
				continue
			}

			// append neighbors to batch
			for _, neighbor := range visit.neighbors {
				if err := neighborsBatch.AppendStruct(neighbor); err != nil {
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

type ClickHouseVisit struct {
	CrawlID        *uuid.UUID `ch:"crawl_id"`
	PeerID         string     `ch:"peer_id"`
	AgentVersion   *string    `ch:"agent_version"`
	Protocols      []string   `ch:"protocols"`
	VisitType      string     `ch:"type"`
	Maddrs         []string   `ch:"multi_addresses"`
	ConnectMaddr   *string    `ch:"connect_multi_address"`
	ConnectErrors  []string   `ch:"connect_errors"`
	CrawlError     *string    `ch:"crawl_error"`
	VisitStartedAt time.Time  `ch:"visit_started_at"`
	VisitEndedAt   time.Time  `ch:"visit_ended_at"`
	neighbors      []*ClickhouseNeighbor
}

type ClickhouseNeighbor struct {
	CrawlID   uuid.UUID `ch:"crawl_id"`
	PeerID    string    `ch:"peer_id"`
	Neighbor  string    `ch:"neighbor_id"`
	ErrorBits uint16    `ch:"error_bits"`
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
	query := `
		SELECT peer_id, multi_addresses
		FROM visits
		WHERE empty(connect_errors)
		  AND visit_started_at BETWEEN (now() - INTERVAL '24 hours') AND now()
		limit ?
	`
	rows, err := c.conn.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	var addrInfos []peer.AddrInfo
	for rows.Next() {
		var pidStr string
		var maddrStrs []string
		if err := rows.Scan(&pidStr, &maddrStrs); err != nil {
			return nil, err
		}

		maddrs, err := utils.AddrsToMaddrs(maddrStrs)
		if err != nil {
			log.WithError(err).WithField("maddrs", maddrStrs).Warnln("Could not parse bootstrap multi addresses from database")
			continue
		}

		pid, err := peer.Decode(pidStr)
		if err != nil {
			log.WithError(err).WithField("pid", pidStr).Warnln("Could not parse bootstrap peer ID from database")
			continue
		}
		addrInfos = append(addrInfos, peer.AddrInfo{
			ID:    pid,
			Addrs: maddrs,
		})
	}

	return addrInfos, nil
}

func (c *ClickHouseClient) InsertVisit(ctx context.Context, args *VisitArgs) error {
	// the crawl can be null if it's a visit from the monitoring task
	var crawlID *uuid.UUID
	if c.crawl != nil {
		crawlID = &c.crawl.ID
	}

	var av *string
	if args.AgentVersion != "" {
		av = &args.AgentVersion
	}

	var connMaddrStr *string
	if args.ConnectMaddr != nil {
		maddrStr := args.ConnectMaddr.String()
		connMaddrStr = &maddrStr
	}

	var crawlErrStr *string
	if args.CrawlErrorStr != "" {
		crawlErrStr = &args.CrawlErrorStr
	}

	visit := &ClickHouseVisit{
		CrawlID:        crawlID,
		PeerID:         args.PeerID.String(),
		AgentVersion:   av,
		Protocols:      args.Protocols,
		VisitType:      args.VisitType.String(),
		Maddrs:         utils.MaddrsToAddrs(args.Maddrs),
		ConnectMaddr:   connMaddrStr,
		ConnectErrors:  []string{},
		CrawlError:     crawlErrStr,
		VisitStartedAt: args.VisitStartedAt,
		VisitEndedAt:   args.VisitEndedAt,
	}

	sort.Strings(visit.Protocols)

	if crawlID != nil {
		visit.neighbors = make([]*ClickhouseNeighbor, len(args.Neighbors))
		for i, neighbor := range args.Neighbors {
			visit.neighbors[i] = &ClickhouseNeighbor{
				CrawlID:   *crawlID,
				PeerID:    args.PeerID.String(),
				Neighbor:  neighbor.String(),
				ErrorBits: args.ErrorBits,
			}
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.visitsChan <- visit:
		return nil
	}
}

func (c *ClickHouseClient) InsertCrawlProperties(ctx context.Context, properties map[string]map[string]int) error {
	// irrelevant for clickhouse
	return nil
}

func (c *ClickHouseClient) InsertNeighbors(ctx context.Context, peerID peer.ID, neighbors []peer.ID, errorBits uint16) error {
	c.crawlMu.RLock()
	defer c.crawlMu.RUnlock()
	if c.crawl == nil {
		return fmt.Errorf("crawl not initialized")
	}

	query := "INSERT INTO neighbors (crawl_id, peer_id, neighbor_id, error_bits) VALUES (?, ?, ?, ?)"

	return c.conn.Exec(ctx, query, c.crawl.ID, peerID, neighbors, errorBits)
}

func (c *ClickHouseClient) SelectPeersToProbe(ctx context.Context) ([]peer.AddrInfo, error) {
	return []peer.AddrInfo{}, nil // TODO: ...
}

func (c *ClickHouseClient) Flush(ctx context.Context) error {
	flushed := make(chan struct{})
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.flushChan <- flushed:
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-flushed:
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

func (c *ClickHouseClient) selectLatestVisit(ctx context.Context) (*ClickHouseVisit, error) {
	visit := &ClickHouseVisit{}
	err := c.conn.QueryRow(ctx, "SELECT * FROM visits ORDER BY visit_started_at desc LIMIT 1").ScanStruct(visit)
	return visit, err
}

func (c *ClickHouseClient) selectNeighbors(ctx context.Context, crawlID uuid.UUID) ([]ClickhouseNeighbor, error) {
	rows, err := c.conn.Query(ctx, "SELECT * FROM neighbors WHERE crawl_id = ?", crawlID)
	if err != nil {
		return nil, err
	}

	var neighbors []ClickhouseNeighbor
	for rows.Next() {
		neighbor := ClickhouseNeighbor{}
		if err := rows.ScanStruct(&neighbor); err != nil {
			return nil, err
		}
		neighbors = append(neighbors, neighbor)
	}
	return neighbors, err
}
