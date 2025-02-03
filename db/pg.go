package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	lru "github.com/hashicorp/golang-lru"
	_ "github.com/lib/pq"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	pgmodels "github.com/dennis-tra/nebula-crawler/db/models/pg"
	"github.com/dennis-tra/nebula-crawler/utils"
)

//go:embed migrations/pg
var migrations embed.FS

var (
	ErrEmptyAgentVersion = fmt.Errorf("empty agent version")
	ErrEmptyProtocol     = fmt.Errorf("empty protocol")
	ErrEmptyProtocolsSet = fmt.Errorf("empty protocols set")
)

type PostgresClientConfig struct {
	// Determines the host address of the database.
	DatabaseHost string

	// Determines the port of the database.
	DatabasePort int

	// Determines the name of the database that should be used.
	DatabaseName string

	// Determines the password with which we access the database.
	DatabasePassword string

	// Determines the username with which we access the database.
	DatabaseUser string

	// The database SSL configuration. For Postgres SSL mode should be
	// one of the supported values here: https://www.postgresql.org/docs/current/libpq-ssl.html)
	DatabaseSSL string

	// The cache size to hold agent versions in memory to skip database queries.
	AgentVersionsCacheSize int

	// The cache size to hold protocols in memory to skip database queries.
	ProtocolsCacheSize int

	// The cache size to hold sets of protocols in memory to skip database queries.
	ProtocolsSetCacheSize int

	// Set the maximum idle connections for the database handler.
	MaxIdleConns int

	// MeterProvider is the meter provider to use when initialising metric instruments.
	MeterProvider metric.MeterProvider

	// TracerProvider is the tracer provider to use when initialising tracing
	TracerProvider trace.TracerProvider
}

// DatabaseSourceName returns the data source name string to be put into the sql.Open method.
func (c *PostgresClientConfig) DatabaseSourceName() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.DatabaseHost,
		c.DatabasePort,
		c.DatabaseName,
		c.DatabaseUser,
		c.DatabasePassword,
		c.DatabaseSSL,
	)
}

type PostgresClient struct {
	ctx context.Context

	// Reference to the configuration
	cfg *PostgresClientConfig

	// Database handler
	dbh *sql.DB

	// protocols cache
	agentVersions *lru.Cache

	// protocols cache
	protocols *lru.Cache

	// protocols set cache
	protocolsSets *lru.Cache

	// A map that maps peer IDs to their database IDs. This speeds up the insertion of neighbor information as
	// the database does not need to look up every peer ID but only the ones not yet present in the database.
	// Speed up for ~11k peers: 5.5 min -> 30s
	peerMappings map[peer.ID]int

	// the crawl entity that was created in the database
	// we don't propagate this object through the rest of the code but instead
	// cache it here because clickhouse and postgres have different type for
	// database IDs (clickhouse strings, postgres ints).
	crawlMu sync.Mutex
	crawl   *pgmodels.Crawl

	// reference to all relevant db telemetry
	telemetry *telemetry
}

// NewClickHouseClient establishes a database connection with the provided configuration
// and applies any pending migrations.
func NewPostgresClient(ctx context.Context, cfg *PostgresClientConfig) (*PostgresClient, error) {
	log.WithFields(log.Fields{
		"host": cfg.DatabaseHost,
		"port": cfg.DatabasePort,
		"name": cfg.DatabaseName,
		"user": cfg.DatabaseUser,
		"ssl":  cfg.DatabaseSSL,
	}).Infoln("Initializing database client")

	dbh, err := otelsql.Open("postgres", cfg.DatabaseSourceName(),
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithMeterProvider(cfg.MeterProvider),
		otelsql.WithTracerProvider(cfg.TracerProvider),
	)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Set to match the writer worker
	dbh.SetMaxIdleConns(cfg.MaxIdleConns) // default is 2 which leads to many connection open/closings

	otelsql.ReportDBStatsMetrics(dbh, otelsql.WithMeterProvider(cfg.MeterProvider))

	// Ping database to verify connection.
	if err = dbh.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	telemetry, err := newTelemetry(cfg.TracerProvider, cfg.MeterProvider)
	if err != nil {
		return nil, fmt.Errorf("new telemetry: %w", err)
	}

	client := &PostgresClient{
		ctx:          ctx,
		cfg:          cfg,
		dbh:          dbh,
		peerMappings: make(map[peer.ID]int),
		telemetry:    telemetry,
	}
	client.applyMigrations(cfg, dbh)

	client.agentVersions, err = lru.New(cfg.AgentVersionsCacheSize)
	if err != nil {
		return nil, fmt.Errorf("new agent versions lru cache: %w", err)
	}

	client.protocols, err = lru.New(cfg.ProtocolsCacheSize)
	if err != nil {
		return nil, fmt.Errorf("new protocol lru cache: %w", err)
	}

	client.protocolsSets, err = lru.New(cfg.ProtocolsSetCacheSize)
	if err != nil {
		return nil, fmt.Errorf("new protocols set lru cache: %w", err)
	}

	if err = client.fillAgentVersionsCache(ctx); err != nil {
		return nil, fmt.Errorf("fill agent versions cache: %w", err)
	}

	if err = client.fillProtocolsCache(ctx); err != nil {
		return nil, fmt.Errorf("fill protocols cache: %w", err)
	}

	if err = client.fillProtocolsSetCache(ctx); err != nil {
		return nil, fmt.Errorf("fill protocols set cache: %w", err)
	}

	// Ensure all appropriate partitions exist
	client.ensurePartitions(ctx, time.Now())
	client.ensurePartitions(ctx, time.Now().Add(24*time.Hour))

	go func() {
		for range time.NewTicker(24 * time.Hour).C {
			client.ensurePartitions(ctx, time.Now().Add(12*time.Hour))
		}
	}()

	return client, nil
}

func (c *PostgresClient) Handle() *sql.DB {
	return c.dbh
}

func (c *PostgresClient) Close() error {
	return c.dbh.Close()
}

func (c *PostgresClient) applyMigrations(cfg *PostgresClientConfig, dbh *sql.DB) {
	tmpDir, err := os.MkdirTemp("", "nebula")
	if err != nil {
		log.WithError(err).WithField("pattern", "nebula").Warnln("Could not create tmp directory for migrations")
		return
	}
	defer func() {
		if err = os.RemoveAll(tmpDir); err != nil {
			log.WithError(err).WithField("tmpDir", tmpDir).Warnln("Could not clean up tmp directory")
		}
	}()
	log.WithField("dir", tmpDir).Debugln("Created temporary directory")

	err = fs.WalkDir(migrations, ".", func(path string, d fs.DirEntry, err error) error {
		join := filepath.Join(tmpDir, path)
		if d.IsDir() {
			return os.MkdirAll(join, 0o755)
		}

		data, err := migrations.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		return os.WriteFile(join, data, 0o644)
	})
	if err != nil {
		log.WithError(err).Warnln("Could not create migrations files")
		return
	}

	// Apply migrations
	driver, err := postgres.WithInstance(dbh, &postgres.Config{})
	if err != nil {
		log.WithError(err).Warnln("Could not create driver instance")
		return
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+filepath.Join(tmpDir, "migrations/pg"), cfg.DatabaseName, driver)
	if err != nil {
		log.WithError(err).Warnln("Could not create migrate instance")
		return
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.WithError(err).Warnln("Couldn't apply migrations")
		return
	}
}

func (c *PostgresClient) ensurePartitions(ctx context.Context, baseDate time.Time) {
	lowerBound := time.Date(baseDate.Year(), baseDate.Month(), 1, 0, 0, 0, 0, baseDate.Location())
	upperBound := lowerBound.AddDate(0, 1, 0)

	query := partitionQuery(pgmodels.TableNames.Visits, lowerBound, upperBound)
	if _, err := c.dbh.ExecContext(ctx, query); err != nil {
		log.WithError(err).WithField("query", query).Warnln("could not create visits partition")
	}

	query = partitionQuery(pgmodels.TableNames.SessionsClosed, lowerBound, upperBound)
	if _, err := c.dbh.ExecContext(ctx, query); err != nil {
		log.WithError(err).WithField("query", query).Warnln("could not create sessions closed partition")
	}

	query = partitionQuery(pgmodels.TableNames.PeerLogs, lowerBound, upperBound)
	if _, err := c.dbh.ExecContext(ctx, query); err != nil {
		log.WithError(err).WithField("query", query).Warnln("could not create peer_logs partition")
	}

	crawl, err := pgmodels.Crawls(qm.OrderBy(pgmodels.CrawlColumns.StartedAt+" DESC")).One(ctx, c.dbh)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithError(err).Warnln("could not load most recent crawl")
	}
	maxCrawlID := 0
	if crawl != nil {
		maxCrawlID = crawl.ID + 1
	}

	neighborsPartitionSize := 1000
	lower := (maxCrawlID / neighborsPartitionSize) * neighborsPartitionSize
	upper := lower + neighborsPartitionSize
	query = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s_%d_%d PARTITION OF %s FOR VALUES FROM (%d) TO (%d)",
		pgmodels.TableNames.Neighbors,
		lower,
		upper,
		pgmodels.TableNames.Neighbors,
		lower,
		upper,
	)
	if _, err := c.dbh.ExecContext(ctx, query); err != nil {
		log.WithError(err).WithField("query", query).Warnln("could not create neighbors partition")
	}
}

func partitionQuery(table string, lower time.Time, upper time.Time) string {
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s_%s_%s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s')",
		table,
		lower.Format("2006"),
		lower.Format("01"),
		table,
		lower.Format("2006-01-02"),
		upper.Format("2006-01-02"),
	)
}

// InitCrawl inserts a crawl instance into the database in the state `started`.
// This is done to receive a database ID that all subsequent database entities can be linked to.
func (c *PostgresClient) InitCrawl(ctx context.Context, version string) (err error) {
	c.crawlMu.Lock()
	defer c.crawlMu.Unlock()

	defer func() {
		if err != nil {
			c.crawl = nil
		}
	}()

	if c.crawl != nil {
		return fmt.Errorf("crawl already initialized")
	}

	c.crawl = &pgmodels.Crawl{
		State:     pgmodels.CrawlStateStarted,
		StartedAt: time.Now(),
		Version:   version,
	}

	return c.crawl.Insert(ctx, c.dbh, boil.Infer())
}

func (c *PostgresClient) SealCrawl(ctx context.Context, args *SealCrawlArgs) (err error) {
	c.crawlMu.Lock()
	defer c.crawlMu.Unlock()

	// TODO: does this perform a deep copy?
	original := *c.crawl
	defer func() {
		// roll back in case of an error
		if err != nil {
			c.crawl = &original
		}
	}()

	now := time.Now()

	c.crawl.UpdatedAt = now
	c.crawl.CrawledPeers = null.IntFrom(args.Crawled)
	c.crawl.DialablePeers = null.IntFrom(args.Dialable)
	c.crawl.UndialablePeers = null.IntFrom(args.Undialable)
	c.crawl.RemainingPeers = null.IntFrom(args.Remaining)
	c.crawl.State = string(args.State)
	c.crawl.FinishedAt = null.TimeFrom(now)

	_, err = c.crawl.Update(ctx, c.dbh, boil.Infer())
	return err
}

func (c *PostgresClient) GetOrCreateProtocol(ctx context.Context, exec boil.ContextExecutor, protocol string) (*int, error) {
	if protocol == "" {
		return nil, ErrEmptyProtocol
	}

	if id, found := c.protocols.Get(protocol); found {
		c.telemetry.cacheQueriesCount.Add(ctx, 1, metric.WithAttributes(
			attribute.String("entity", "protocol"),
			attribute.Bool("hit", true),
		))
		return id.(*int), nil
	}
	c.telemetry.cacheQueriesCount.Add(ctx, 1, metric.WithAttributes(
		attribute.String("entity", "protocol"),
		attribute.Bool("hit", false),
	))

	log.WithField("protocol", protocol).Infoln("Upsert protocol")
	row := exec.QueryRowContext(ctx, "SELECT upsert_protocol($1)", protocol)
	if row.Err() != nil {
		return nil, fmt.Errorf("unable to upsert protocol: %w", row.Err())
	}

	var protocolID *int
	if err := row.Scan(&protocolID); err != nil {
		return nil, fmt.Errorf("unable to scan result from upsert protocol: %w", err)
	}

	if protocolID == nil {
		return nil, fmt.Errorf("protocol not created")
	}

	c.protocols.Add(protocol, protocolID)

	return protocolID, nil
}

// fillProtocolsCache fetches all rows until protocol cache size from the protocols table and
// initializes the DB clients protocols cache.
func (c *PostgresClient) fillProtocolsCache(ctx context.Context) error {
	if c.cfg.ProtocolsCacheSize == 0 {
		return nil
	}

	prots, err := pgmodels.Protocols(qm.Limit(c.cfg.ProtocolsCacheSize)).All(ctx, c.dbh)
	if err != nil {
		return err
	}

	for _, p := range prots {
		c.protocols.Add(p.Protocol, &p.ID)
	}

	return nil
}

// fillProtocolsSetCache fetches all rows until protocolSet cache size from the protocolsSets table and
// initializes the DB clients protocolsSets cache.
func (c *PostgresClient) fillProtocolsSetCache(ctx context.Context) error {
	if c.cfg.ProtocolsSetCacheSize == 0 {
		return nil
	}

	protSets, err := pgmodels.ProtocolsSets(qm.Limit(c.cfg.ProtocolsSetCacheSize)).All(ctx, c.dbh)
	if err != nil {
		return err
	}

	for _, ps := range protSets {
		c.protocolsSets.Add(string(ps.Hash), &ps.ID)
	}

	return nil
}

// protocolsSetHash returns a unique hash digest for this set of protocol IDs as it's also generated by the database.
// It expects the list of protocolIDs to be sorted in ascending order.
func (c *PostgresClient) protocolsSetHash(protocolIDs []int64) string {
	protocolStrs := make([]string, len(protocolIDs))
	for i, id := range protocolIDs {
		protocolStrs[i] = strconv.Itoa(int(id)) // safe because protocol IDs are just integers in the database.
	}
	dat := []byte("{" + strings.Join(protocolStrs, ",") + "}")

	h := sha256.New()
	h.Write(dat)
	return string(h.Sum(nil))
}

func (c *PostgresClient) GetOrCreateProtocolsSetID(ctx context.Context, exec boil.ContextExecutor, protocols []string) (*int, error) {
	if len(protocols) == 0 {
		return nil, ErrEmptyProtocolsSet
	}

	protocolIDs := make([]int64, len(protocols))
	for i, protocol := range protocols {
		protocolID, err := c.GetOrCreateProtocol(ctx, exec, protocol)
		if errors.Is(err, ErrEmptyProtocol) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("get or create protocol: %w", err)
		}
		protocolIDs[i] = int64(*protocolID)
	}

	sort.Slice(protocolIDs, func(i, j int) bool { return protocolIDs[i] < protocolIDs[j] })

	key := c.protocolsSetHash(protocolIDs)
	if id, found := c.protocolsSets.Get(key); found {
		c.telemetry.cacheQueriesCount.Add(ctx, 1, metric.WithAttributes(
			attribute.String("entity", "protocol_set"),
			attribute.Bool("hit", true),
		))
		return id.(*int), nil
	}
	c.telemetry.cacheQueriesCount.Add(ctx, 1, metric.WithAttributes(
		attribute.String("entity", "protocol_set"),
		attribute.Bool("hit", false),
	))

	log.WithField("key", hex.EncodeToString([]byte(key))).Infoln("Upsert protocols set")
	row := exec.QueryRowContext(ctx, "SELECT upsert_protocol_set_id($1)", types.Int64Array(protocolIDs))
	if row.Err() != nil {
		return nil, fmt.Errorf("unable to upsert protocols set: %w", row.Err())
	}

	var protocolsSetID *int
	if err := row.Scan(&protocolsSetID); err != nil {
		return nil, fmt.Errorf("unable to scan result from upsert protocol set id: %w", err)
	}

	if protocolsSetID == nil {
		return nil, fmt.Errorf("protocols set not created")
	}

	c.protocolsSets.Add(key, protocolsSetID)

	return protocolsSetID, nil
}

func (c *PostgresClient) GetOrCreateAgentVersionID(ctx context.Context, exec boil.ContextExecutor, agentVersion string) (*int, error) {
	if agentVersion == "" {
		return nil, ErrEmptyAgentVersion
	}

	if id, found := c.agentVersions.Get(agentVersion); found {
		c.telemetry.cacheQueriesCount.Add(ctx, 1, metric.WithAttributes(
			attribute.String("entity", "agent_version"),
			attribute.Bool("hit", true),
		))
		return id.(*int), nil
	}
	c.telemetry.cacheQueriesCount.Add(ctx, 1, metric.WithAttributes(
		attribute.String("entity", "agent_version"),
		attribute.Bool("hit", false),
	))

	log.WithField("agentVersion", agentVersion).Infoln("Upsert agent version")
	row := exec.QueryRowContext(ctx, "SELECT upsert_agent_version($1)", agentVersion)
	if row.Err() != nil {
		return nil, fmt.Errorf("unable to upsert agent version: %w", row.Err())
	}

	var agentVersionID *int
	if err := row.Scan(&agentVersionID); err != nil {
		return nil, fmt.Errorf("unable to scan result from upsert agent version: %w", err)
	}

	if agentVersionID == nil {
		return nil, fmt.Errorf("agentVersion not created")
	}

	c.agentVersions.Add(agentVersion, agentVersionID)

	return agentVersionID, nil
}

// fillAgentVersionsCache fetches all rows until agent version cache size from the agent_versions table and
// initializes the DB clients agent version cache.
func (c *PostgresClient) fillAgentVersionsCache(ctx context.Context) error {
	if c.cfg.AgentVersionsCacheSize == 0 {
		return nil
	}

	avs, err := pgmodels.AgentVersions(qm.Limit(c.cfg.AgentVersionsCacheSize)).All(ctx, c.dbh)
	if err != nil {
		return err
	}

	for _, av := range avs {
		c.agentVersions.Add(av.AgentVersion, &av.ID)
	}

	return nil
}

func (c *PostgresClient) UpsertPeer(mh string, agentVersionID null.Int, protocolSetID null.Int, properties null.JSON) (int, error) {
	rows, err := queries.Raw("SELECT upsert_peer($1, $2, $3, $4)",
		mh, agentVersionID, protocolSetID, properties,
	).Query(c.dbh)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.WithError(err).Warnln("Could not close rows")
		}
	}()

	id := 0
	if !rows.Next() {
		return id, nil
	}

	if err = rows.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (c *PostgresClient) InsertVisit(ctx context.Context, args *VisitArgs) error {
	var agentVersionID, protocolsSetID *int
	var avidErr, psidErr error

	var wg sync.WaitGroup
	if args.AgentVersion != "" {
		wg.Add(1)
		go func() {
			agentVersionID, avidErr = c.GetOrCreateAgentVersionID(ctx, c.dbh, args.AgentVersion)
			if avidErr != nil && !errors.Is(avidErr, ErrEmptyAgentVersion) && !errors.Is(psidErr, context.Canceled) {
				log.WithError(avidErr).WithField("agentVersion", args.AgentVersion).Warnln("Error getting or creating agent version id")
			}
			wg.Done()
		}()
	}

	if args.Protocols != nil && len(args.Protocols) > 0 {

		wg.Add(1)
		go func() {
			protocolsSetID, psidErr = c.GetOrCreateProtocolsSetID(ctx, c.dbh, args.Protocols)
			if psidErr != nil && !errors.Is(psidErr, ErrEmptyProtocolsSet) && !errors.Is(psidErr, context.Canceled) {
				log.WithError(psidErr).WithField("protocols", args.Protocols).Warnln("Error getting or creating protocols set id")
			}
			wg.Done()
		}()
	}
	wg.Wait()

	var crawlID *int
	if c.crawl != nil {
		crawlID = &c.crawl.ID
	} else if args.CrawlDuration != nil {
		log.Warnln("Crawl duration provided but no crawl initialized.")
	}

	start := time.Now()
	rows, err := queries.Raw("SELECT insert_visit($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
		crawlID,
		args.PeerID.String(),
		types.StringArray(utils.MaddrsToAddrs(args.Maddrs)),
		agentVersionID,
		protocolsSetID,
		durationToInterval(args.DialDuration),
		durationToInterval(args.ConnectDuration),
		durationToInterval(args.CrawlDuration),
		args.VisitStartedAt,
		args.VisitEndedAt,
		args.VisitType,
		null.NewString(args.ConnectErrorStr, args.ConnectErrorStr != ""),
		null.NewString(args.CrawlErrorStr, args.CrawlErrorStr != ""),
		args.Properties,
	).QueryContext(ctx, c.dbh)
	c.telemetry.insertVisitHistogram.Record(ctx, time.Since(start).Milliseconds(), metric.WithAttributes(
		attribute.String("type", string(args.VisitType)),
		attribute.Bool("success", err == nil),
	))
	if err != nil {
		return err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.WithError(err).Warnln("Could not close rows")
		}
	}()

	ivr := insertVisitResult{
		PID: args.PeerID,
	}
	if !rows.Next() {
		return nil
	}

	if err = rows.Scan(&ivr); err != nil {
		return err
	}

	c.peerMappings[ivr.PID] = *ivr.PeerID

	return nil
}

type insertVisitResult struct {
	PID     peer.ID
	PeerID  *int
	VisitID *int
}

func (ivr *insertVisitResult) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("incompatible type %T", value)
	}

	parts := strings.Split(string(data[1:len(data)-1]), ",")
	if len(parts) != 3 {
		return fmt.Errorf("unexpected number of return values: %s", string(data))
	}

	if parts[0] != "" {
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("invalid db peer id %s", parts[0])
		}
		ivr.PeerID = &id
	}

	if parts[1] != "" {
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid db visit id %s", parts[1])
		}
		ivr.VisitID = &id
	}

	return nil
}

func durationToInterval(dur *time.Duration) *string {
	if dur == nil || *dur == 0 {
		return nil
	}
	s := fmt.Sprintf("%f seconds", dur.Seconds())
	return &s
}

func (c *PostgresClient) InsertNeighbors(ctx context.Context, peerID peer.ID, neighbors []peer.ID, errorBits uint16) error {
	var dbPeerID *int
	if value, ok := c.peerMappings[peerID]; ok {
		dbPeerID = &value
	}

	var dbNeighborIDs []int
	var neighborMHashes []string
	for _, n := range neighbors {
		if id, found := c.peerMappings[n]; found {
			dbNeighborIDs = append(dbNeighborIDs, id)
		} else {
			neighborMHashes = append(neighborMHashes, n.String())
		}
	}

	// postgres does not support unsigned integers. So we interpret the uint16 as an int16
	bitMask := *(*int16)(unsafe.Pointer(&errorBits))
	rows, err := queries.Raw("SELECT insert_neighbors($1, $2, $3, $4, $5, $6)",
		c.crawl.ID,
		dbPeerID,
		peerID.String(),
		fmt.Sprintf("{%s}", strings.Trim(strings.Join(strings.Split(fmt.Sprint(dbNeighborIDs), " "), ","), "[]")),
		fmt.Sprintf("{%s}", strings.Join(neighborMHashes, ",")),
		bitMask,
	).QueryContext(ctx, c.dbh)
	if err != nil {
		return err
	}
	return rows.Close()
}

func (c *PostgresClient) InsertCrawlProperties(ctx context.Context, properties map[string]map[string]int) error {
	txn, err := c.dbh.BeginTx(ctx, nil)
	if err != nil {
		return errors.New("start peer property txn")
	}
	defer Rollback(txn)

	for property, valuesMap := range properties {
		for value, count := range valuesMap {

			cp := &pgmodels.CrawlProperty{
				CrawlID: c.crawl.ID,
				Count:   count,
			}

			if value == "" {
				continue
			} else if property == "protocol" {
				protocolID, err := c.GetOrCreateProtocol(ctx, txn, value)
				if err != nil {
					continue
				}
				cp.ProtocolID = null.IntFromPtr(protocolID)
			} else if property == "agent_version" {
				agentVersionID, err := c.GetOrCreateAgentVersionID(ctx, txn, value)
				if err != nil {
					continue
				}
				cp.AgentVersionID = null.IntFromPtr(agentVersionID)
			} else if property == "error" {
				cp.Error = null.StringFrom(value)
			}

			if cp.ProtocolID.IsZero() && cp.AgentVersionID.IsZero() && cp.Error.IsZero() {
				log.WithFields(log.Fields{
					"property": property,
					"value":    value,
					"count":    count,
				}).Warnln("No crawl property set")
				continue
			}

			if err := cp.Insert(ctx, txn, boil.Infer()); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"crawlId":  c.crawl.ID,
					"property": property,
					"value":    value,
				}).Warnln("Could not insert peer property txn")
				continue
			}
		}
	}

	return txn.Commit()
}

func (c *PostgresClient) QueryBootstrapPeers(ctx context.Context, limit int) ([]peer.AddrInfo, error) {
	peers, err := pgmodels.Peers(
		qm.Load(pgmodels.PeerRels.MultiAddresses),
		qm.InnerJoin("sessions_open s on s.peer_id = peers.id"),
		qm.OrderBy("s.last_visited_at"),
		qm.Limit(limit),
	).All(ctx, c.dbh)
	if err != nil {
		return nil, err
	}

	var pis []peer.AddrInfo
	for _, p := range peers {
		id, err := peer.Decode(p.MultiHash)
		if err != nil {
			log.WithError(err).Warnln("Could not decode multi hash ", p.MultiHash)
			continue
		}
		var maddrs []ma.Multiaddr
		for _, maddrStr := range p.R.MultiAddresses {
			maddr, err := ma.NewMultiaddr(maddrStr.Maddr)
			if err != nil {
				log.WithError(err).Warnln("Could not decode multi addr ", maddrStr)
				continue
			}

			maddrs = append(maddrs, maddr)
		}
		pi := peer.AddrInfo{
			ID:    id,
			Addrs: maddrs,
		}
		pis = append(pis, pi)
	}
	return pis, nil
}

// SelectPeersToProbe fetches all open sessions from the database that are due
// to be dialed/probed.
func (c *PostgresClient) SelectPeersToProbe(ctx context.Context) ([]peer.AddrInfo, error) {
	openSessions, err := pgmodels.SessionsOpens(
		pgmodels.SessionsOpenWhere.NextVisitDueAt.LT(time.Now()),
		qm.Load(pgmodels.SessionsOpenRels.Peer),
		qm.Load(qm.Rels(pgmodels.SessionsOpenRels.Peer, pgmodels.PeerRels.MultiAddresses)),
	).All(ctx, c.dbh)
	if err != nil {
		return nil, err
	}

	addrInfos := make([]peer.AddrInfo, 0, len(openSessions))
	for _, session := range openSessions {
		// take multi hash and decode into PeerID
		peerID, err := peer.Decode(session.R.Peer.MultiHash)
		if err != nil {
			log.WithField("mhash", session.R.Peer.MultiHash).
				WithError(err).
				Warnln("Could not parse multi address")
			continue
		}
		logEntry := log.WithField("peerID", peerID.ShortString())

		// Parse multi addresses from database
		maddrs := make([]ma.Multiaddr, 0, len(session.R.Peer.R.MultiAddresses))
		for _, dbMaddr := range session.R.Peer.R.MultiAddresses {
			maddr, err := ma.NewMultiaddr(dbMaddr.Maddr)
			if err != nil {
				logEntry.WithError(err).WithField("maddr", dbMaddr.Maddr).Warnln("Could not parse multi address")
				continue
			}
			maddrs = append(maddrs, maddr)
		}
		addrInfos = append(addrInfos, peer.AddrInfo{
			ID:    peerID,
			Addrs: maddrs,
		})
	}

	return addrInfos, nil
}

// FetchUnresolvedMultiAddresses fetches all multi addresses that were not resolved yet.
func (c *PostgresClient) FetchUnresolvedMultiAddresses(ctx context.Context, limit int) (pgmodels.MultiAddressSlice, error) {
	return pgmodels.MultiAddresses(
		pgmodels.MultiAddressWhere.Resolved.EQ(false),
		qm.OrderBy(pgmodels.MultiAddressColumns.CreatedAt),
		qm.Limit(limit),
	).All(ctx, c.dbh)
}

// Rollback calls rollback on the given transaction and logs the potential error.
func Rollback(txn *sql.Tx) {
	if err := txn.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		log.WithError(err).Warnln("An error occurred when rolling back transaction")
	}
}
