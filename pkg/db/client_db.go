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

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	lru "github.com/hashicorp/golang-lru"
	_ "github.com/lib/pq"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
)

//go:embed migrations
var migrations embed.FS

var (
	ErrEmptyAgentVersion = fmt.Errorf("empty agent version")
	ErrEmptyProtocol     = fmt.Errorf("empty protocol")
	ErrEmptyProtocolsSet = fmt.Errorf("empty protocols set")
)

type DBClient struct {
	ctx context.Context

	// Reference to the configuration
	cfg *config.Database

	// Database handler
	dbh *sql.DB

	// protocols cache
	agentVersions *lru.Cache

	// protocols cache
	protocols *lru.Cache

	// protocols set cache
	protocolsSets *lru.Cache
}

var _ Client = (*DBClient)(nil)

// InitDBClient establishes a database connection with the provided configuration
// and applies any pending migrations.
func InitDBClient(ctx context.Context, cfg *config.Database) (*DBClient, error) {
	log.WithFields(log.Fields{
		"host": cfg.DatabaseHost,
		"port": cfg.DatabasePort,
		"name": cfg.DatabaseName,
		"user": cfg.DatabaseUser,
		"ssl":  cfg.DatabaseSSLMode,
	}).Infoln("Initializing database client")

	driverName, err := ocsql.Register("postgres")
	if err != nil {
		return nil, fmt.Errorf("register ocsql: %w", err)
	}

	// Open database handle
	dbh, err := sql.Open(driverName, cfg.DatabaseSourceName())
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Ping database to verify connection.
	if err = dbh.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	client := &DBClient{ctx: ctx, cfg: cfg, dbh: dbh}
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

func (c *DBClient) Handle() *sql.DB {
	return c.dbh
}

func (c *DBClient) Close() error {
	return c.dbh.Close()
}

func (c *DBClient) applyMigrations(cfg *config.Database, dbh *sql.DB) {
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

	m, err := migrate.NewWithDatabaseInstance("file://"+filepath.Join(tmpDir, "migrations"), cfg.DatabaseName, driver)
	if err != nil {
		log.WithError(err).Warnln("Could not create migrate instance")
		return
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.WithError(err).Warnln("Couldn't apply migrations")
		return
	}
}

func (c *DBClient) ensurePartitions(ctx context.Context, baseDate time.Time) {
	lowerBound := time.Date(baseDate.Year(), baseDate.Month(), 1, 0, 0, 0, 0, baseDate.Location())
	upperBound := lowerBound.AddDate(0, 1, 0)

	query := partitionQuery(models.TableNames.Visits, lowerBound, upperBound)
	if _, err := c.dbh.ExecContext(ctx, query); err != nil {
		log.WithError(err).WithField("query", query).Warnln("could not create visits partition")
	}

	query = partitionQuery(models.TableNames.SessionsClosed, lowerBound, upperBound)
	if _, err := c.dbh.ExecContext(ctx, query); err != nil {
		log.WithError(err).WithField("query", query).Warnln("could not create sessions closed partition")
	}

	query = partitionQuery(models.TableNames.PeerLogs, lowerBound, upperBound)
	if _, err := c.dbh.ExecContext(ctx, query); err != nil {
		log.WithError(err).WithField("query", query).Warnln("could not create peer_logs partition")
	}

	crawl, err := models.Crawls(qm.OrderBy(models.CrawlColumns.StartedAt+" DESC")).One(ctx, c.dbh)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithError(err).Warnln("could not load most recent crawl")
	}
	maxCrawlID := 0
	if crawl != nil {
		maxCrawlID = crawl.ID
	}

	neighborsPartitionSize := 1000
	lower := (maxCrawlID / neighborsPartitionSize) * neighborsPartitionSize
	upper := lower + neighborsPartitionSize
	query = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s_%d_%d PARTITION OF %s FOR VALUES FROM (%d) TO (%d)",
		models.TableNames.Neighbors,
		lower,
		upper,
		models.TableNames.Neighbors,
		lower,
		upper,
	)
	if _, err := c.dbh.ExecContext(ctx, query); err != nil {
		log.WithError(err).WithField("query", query).Warnln("could not create peer_logs partition")
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
func (c *DBClient) InitCrawl(ctx context.Context) (*models.Crawl, error) {
	crawl := &models.Crawl{
		State:     models.CrawlStateStarted,
		StartedAt: time.Now(),
	}
	return crawl, crawl.Insert(ctx, c.dbh, boil.Infer())
}

// UpdateCrawl takes the crawl model an updates it in the database.
func (c *DBClient) UpdateCrawl(ctx context.Context, crawl *models.Crawl) error {
	_, err := crawl.Update(ctx, c.dbh, boil.Infer())
	return err
}

func (c *DBClient) PersistCrawlVisit(
	ctx context.Context,
	crawlID int,
	peerID peer.ID,
	maddrs []ma.Multiaddr,
	protocols []string,
	agentVersion string,
	connectDuration time.Duration,
	crawlDuration time.Duration,
	visitStartedAt time.Time,
	visitEndedAt time.Time,
	connectErrorStr string,
	crawlErrorStr string,
	properties null.JSON,
) (*InsertVisitResult, error) {
	var agentVersionID, protocolsSetID *int
	var avidErr, psidErr error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		agentVersionID, avidErr = c.GetOrCreateAgentVersionID(ctx, c.dbh, agentVersion)
		if avidErr != nil && !errors.Is(avidErr, ErrEmptyAgentVersion) && !errors.Is(psidErr, context.Canceled) {
			log.WithError(avidErr).WithField("agentVersion", agentVersion).Warnln("Error getting or creating agent version id")
		}
		wg.Done()
	}()
	go func() {
		protocolsSetID, psidErr = c.GetOrCreateProtocolsSetID(ctx, c.dbh, protocols)
		if psidErr != nil && !errors.Is(psidErr, ErrEmptyProtocolsSet) && !errors.Is(psidErr, context.Canceled) {
			log.WithError(psidErr).WithField("protocols", protocols).Warnln("Error getting or creating protocols set id")
		}
		wg.Done()
	}()
	wg.Wait()

	return c.insertVisit(
		&crawlID,
		peerID,
		maddrs,
		null.IntFromPtr(agentVersionID),
		null.IntFromPtr(protocolsSetID),
		nil,
		&connectDuration,
		&crawlDuration,
		visitStartedAt,
		visitEndedAt,
		models.VisitTypeCrawl,
		connectErrorStr,
		crawlErrorStr,
		properties,
	)
}

func (c *DBClient) GetOrCreateProtocol(ctx context.Context, exec boil.ContextExecutor, protocol string) (*int, error) {
	if protocol == "" {
		return nil, ErrEmptyProtocol
	}

	if id, found := c.protocols.Get(protocol); found {
		metrics.CacheQueriesCount.With(prometheus.Labels{"entity": "protocol", "outcome": "hit"}).Inc()
		return id.(*int), nil
	}
	metrics.CacheQueriesCount.With(prometheus.Labels{"entity": "protocol", "outcome": "miss"}).Inc()

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
func (c *DBClient) fillProtocolsCache(ctx context.Context) error {
	if c.cfg.ProtocolsCacheSize == 0 {
		return nil
	}

	prots, err := models.Protocols(qm.Limit(c.cfg.ProtocolsCacheSize)).All(ctx, c.dbh)
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
func (c *DBClient) fillProtocolsSetCache(ctx context.Context) error {
	if c.cfg.ProtocolsSetCacheSize == 0 {
		return nil
	}

	protSets, err := models.ProtocolsSets(qm.Limit(c.cfg.ProtocolsSetCacheSize)).All(ctx, c.dbh)
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
func (c *DBClient) protocolsSetHash(protocolIDs []int64) string {
	protocolStrs := make([]string, len(protocolIDs))
	for i, id := range protocolIDs {
		protocolStrs[i] = strconv.Itoa(int(id)) // safe because protocol IDs are just integers in the database.
	}
	dat := []byte("{" + strings.Join(protocolStrs, ",") + "}")

	h := sha256.New()
	h.Write(dat)
	return string(h.Sum(nil))
}

func (c *DBClient) GetOrCreateProtocolsSetID(ctx context.Context, exec boil.ContextExecutor, protocols []string) (*int, error) {
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
		metrics.CacheQueriesCount.With(prometheus.Labels{"entity": "protocol_set", "outcome": "hit"}).Inc()
		return id.(*int), nil
	}
	metrics.CacheQueriesCount.With(prometheus.Labels{"entity": "protocol_set", "outcome": "miss"}).Inc()

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

func (c *DBClient) GetOrCreateAgentVersionID(ctx context.Context, exec boil.ContextExecutor, agentVersion string) (*int, error) {
	if agentVersion == "" {
		return nil, ErrEmptyAgentVersion
	}

	if id, found := c.agentVersions.Get(agentVersion); found {
		metrics.CacheQueriesCount.With(prometheus.Labels{"entity": "agent_version", "outcome": "hit"}).Inc()
		return id.(*int), nil
	}
	metrics.CacheQueriesCount.With(prometheus.Labels{"entity": "agent_version", "outcome": "miss"}).Inc()

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
func (c *DBClient) fillAgentVersionsCache(ctx context.Context) error {
	if c.cfg.AgentVersionsCacheSize == 0 {
		return nil
	}

	avs, err := models.AgentVersions(qm.Limit(c.cfg.AgentVersionsCacheSize)).All(ctx, c.dbh)
	if err != nil {
		return err
	}

	for _, av := range avs {
		c.agentVersions.Add(av.AgentVersion, &av.ID)
	}

	return nil
}

func (c *DBClient) UpsertPeer(mh string, agentVersionID null.Int, protocolSetID null.Int, properties null.JSON) (int, error) {
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

func (c *DBClient) PersistDialVisit(
	peerID peer.ID,
	maddrs []ma.Multiaddr,
	dialDuration time.Duration,
	visitStartedAt time.Time,
	visitEndedAt time.Time,
	errorStr string,
) (*InsertVisitResult, error) {
	return c.insertVisit(
		nil,
		peerID,
		maddrs,
		null.IntFromPtr(nil),
		null.IntFromPtr(nil),
		&dialDuration,
		nil,
		nil,
		visitStartedAt,
		visitEndedAt,
		models.VisitTypeDial,
		errorStr,
		"",
		null.JSONFromPtr(nil),
	)
}

func (c *DBClient) insertVisit(
	crawlID *int,
	peerID peer.ID,
	maddrs []ma.Multiaddr,
	agentVersionID null.Int,
	protocolsSetID null.Int,
	dialDuration *time.Duration,
	connectDuration *time.Duration,
	crawlDuration *time.Duration,
	visitStartedAt time.Time,
	visitEndedAt time.Time,
	visitType string,
	connectErrorStr string,
	crawlErrorStr string,
	properties null.JSON,
) (*InsertVisitResult, error) {
	maddrStrs := utils.MaddrsToAddrs(maddrs)

	start := time.Now()
	rows, err := queries.Raw("SELECT insert_visit($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
		crawlID,
		peerID.String(),
		types.StringArray(maddrStrs),
		agentVersionID,
		protocolsSetID,
		durationToInterval(dialDuration),
		durationToInterval(connectDuration),
		durationToInterval(crawlDuration),
		visitStartedAt,
		visitEndedAt,
		visitType,
		null.NewString(connectErrorStr, connectErrorStr != ""),
		null.NewString(crawlErrorStr, crawlErrorStr != ""),
		properties,
	).Query(c.dbh)
	metrics.InsertVisitHistogram.With(prometheus.Labels{
		"type":    visitType,
		"success": strconv.FormatBool(err == nil),
	}).Observe(time.Since(start).Seconds())
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.WithError(err).Warnln("Could not close rows")
		}
	}()

	ivr := InsertVisitResult{
		PID: peerID,
	}
	if !rows.Next() {
		return &ivr, nil
	}

	if err = rows.Scan(&ivr); err != nil {
		return nil, err
	}

	return &ivr, nil
}

type InsertVisitResult struct {
	PID       peer.ID
	PeerID    *int
	VisitID   *int
	SessionID *int
}

func (ivr *InsertVisitResult) Scan(value interface{}) error {
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

	if parts[2] != "" {
		id, err := strconv.Atoi(parts[2])
		if err != nil {
			return fmt.Errorf("invalid db session id %s", parts[2])
		}
		ivr.SessionID = &id
	}

	return nil
}

func durationToInterval(dur *time.Duration) *string {
	if dur == nil {
		return nil
	}
	s := fmt.Sprintf("%f seconds", dur.Seconds())
	return &s
}

func (c *DBClient) PersistNeighbors(ctx context.Context, crawl *models.Crawl, dbPeerID *int, peerID peer.ID, errorBits uint16, dbNeighborsIDs []int, neighbors []peer.ID) error {
	neighborMHashes := make([]string, len(neighbors))
	for i, neighbor := range neighbors {
		neighborMHashes[i] = neighbor.String()
	}
	// postgres does not support unsigned integers. So we interpret the uint16 as an int16
	bitMask := *(*int16)(unsafe.Pointer(&errorBits))
	rows, err := queries.Raw("SELECT insert_neighbors($1, $2, $3, $4, $5, $6)",
		crawl.ID,
		dbPeerID,
		peerID.String(),
		fmt.Sprintf("{%s}", strings.Trim(strings.Join(strings.Split(fmt.Sprint(dbNeighborsIDs), " "), ","), "[]")),
		fmt.Sprintf("{%s}", strings.Join(neighborMHashes, ",")),
		bitMask,
	).QueryContext(ctx, c.dbh)
	if err != nil {
		return err
	}
	return rows.Close()
}

func (c *DBClient) PersistCrawlProperties(ctx context.Context, crawl *models.Crawl, properties map[string]map[string]int) error {
	txn, err := c.dbh.BeginTx(ctx, nil)
	if err != nil {
		return errors.New("start peer property txn")
	}
	defer Rollback(txn)

	for property, valuesMap := range properties {
		for value, count := range valuesMap {

			cp := &models.CrawlProperty{
				CrawlID: crawl.ID,
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
					"crawlId":  crawl.ID,
					"property": property,
					"value":    value,
				}).Warnln("Could not insert peer property txn")
				continue
			}
		}
	}

	return txn.Commit()
}

func (c *DBClient) QueryBootstrapPeers(ctx context.Context, limit int) ([]peer.AddrInfo, error) {
	peers, err := models.Peers(
		qm.Load(models.PeerRels.MultiAddresses),
		qm.InnerJoin("sessions_closed s on s.peer_id = peers.id"),
		qm.OrderBy("s.created_at"),
		qm.Limit(limit),
	).All(ctx, c.dbh)
	if err != nil {
		return nil, err
	}

	var pis []peer.AddrInfo
	for _, p := range peers {
		id, err := peer.Decode(p.MultiHash)
		if err != nil {
			log.Warnln("Could not decode multi hash ", p.MultiHash)
			continue
		}
		var maddrs []ma.Multiaddr
		for _, maddrStr := range p.R.MultiAddresses {
			maddr, err := ma.NewMultiaddr(maddrStr.Maddr)
			if err != nil {
				log.Warnln("Could not decode multi addr ", maddrStr)
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

// FetchDueOpenSessions fetches all open sessions from the database that are due.
func (c *DBClient) FetchDueOpenSessions(ctx context.Context) (models.SessionsOpenSlice, error) {
	return models.SessionsOpens(
		models.SessionsOpenWhere.NextVisitDueAt.LT(time.Now()),
		qm.Load(models.SessionsOpenRels.Peer),
		qm.Load(qm.Rels(models.SessionsOpenRels.Peer, models.PeerRels.MultiAddresses)),
	).All(ctx, c.dbh)
}

// FetchUnresolvedMultiAddresses fetches all multi addresses that were not resolved yet.
func (c *DBClient) FetchUnresolvedMultiAddresses(ctx context.Context, limit int) (models.MultiAddressSlice, error) {
	return models.MultiAddresses(
		models.MultiAddressWhere.Resolved.EQ(false),
		qm.OrderBy(models.MultiAddressColumns.CreatedAt),
		qm.Limit(limit),
	).All(ctx, c.dbh)
}

// Rollback calls rollback on the given transaction and logs the potential error.
func Rollback(txn *sql.Tx) {
	if err := txn.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		log.WithError(err).Warnln("An error occurred when rolling back transaction")
	}
}
