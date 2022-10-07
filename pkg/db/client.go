package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"go.opencensus.io/stats"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
)

//go:embed migrations
var migrations embed.FS

type Client struct {
	ctx context.Context

	// Database handler
	dbh *sql.DB

	// protocols cache
	protocols *lru.Cache

	// protocols cache
	agentVersions *lru.Cache
}

func InitClient(ctx context.Context, conf *config.Config) (*Client, error) {
	log.WithFields(log.Fields{
		"host": conf.DatabaseHost,
		"port": conf.DatabasePort,
		"name": conf.DatabaseName,
		"user": conf.DatabaseUser,
		"ssl":  conf.DatabaseSSLMode,
	}).Infoln("Initializing database client")

	driverName, err := ocsql.Register("postgres")
	if err != nil {
		return nil, errors.Wrap(err, "register ocsql")
	}

	// Open database handle
	dbh, err := sql.Open(driverName, conf.DatabaseSourceName())
	if err != nil {
		return nil, errors.Wrap(err, "opening database")
	}

	// Ping database to verify connection.
	if err = dbh.Ping(); err != nil {
		return nil, errors.Wrap(err, "pinging database")
	}

	client := &Client{ctx: ctx, dbh: dbh}
	client.applyMigrations(conf, dbh)

	client.protocols, err = lru.New(conf.ProtocolsCacheSize)
	if err != nil {
		return nil, errors.Wrap(err, "new protocol lru cache")
	}

	client.agentVersions, err = lru.New(conf.AgentVersionsCacheSize)
	if err != nil {
		return nil, errors.Wrap(err, "new agent versions lru cache")
	}

	if err = client.fillProtocolsCache(ctx); err != nil {
		return nil, errors.Wrap(err, "fill protocols cache")
	}

	if err = client.fillAgentVersionsCache(ctx); err != nil {
		return nil, errors.Wrap(err, "fill agent versions cache")
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

func (c *Client) Handle() *sql.DB {
	return c.dbh
}

func (c *Client) Close() error {
	return c.dbh.Close()
}

func (c *Client) applyMigrations(conf *config.Config, dbh *sql.DB) {
	tmpDir, err := os.MkdirTemp("", "nebula-"+conf.Version)
	if err != nil {
		log.WithError(err).WithField("pattern", "nebula-"+conf.Version).Warnln("Could not create tmp directory for migrations")
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
			return os.MkdirAll(join, 0755)
		}

		data, err := migrations.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "read file")
		}

		err = os.WriteFile(join, data, 0644)
		return err
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

	m, err := migrate.NewWithDatabaseInstance("file://"+filepath.Join(tmpDir, "migrations"), conf.DatabaseName, driver)
	if err != nil {
		log.WithError(err).Warnln("Could not create migrate instance")
		return
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		log.WithError(err).Warnln("Couldn't apply migrations")
		return
	}
}

func (c *Client) ensurePartitions(ctx context.Context, baseDate time.Time) {
	lowerBound := time.Date(baseDate.Year(), baseDate.Month(), 1, 0, 0, 0, 0, baseDate.Location())
	upperBound := lowerBound.AddDate(0, 1, 0)
	lowerBound.Format("2006-01-02")

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS visits_%s_%s PARTITION OF visits FOR VALUES FROM ('%s') TO ('%s')",
		lowerBound.Format("2006"),
		lowerBound.Format("01"),
		lowerBound.Format("2006-01-02"),
		upperBound.Format("2006-01-02"),
	)
	_, err := c.dbh.ExecContext(ctx, query)
	if err != nil {
		log.WithError(err).WithField("query", query).Warnln("could not create visits partition")
	}

	query = fmt.Sprintf("CREATE TABLE IF NOT EXISTS sessions_closed_%s_%s PARTITION OF sessions_closed FOR VALUES FROM ('%s') TO ('%s')",
		lowerBound.Format("2006"),
		lowerBound.Format("01"),
		lowerBound.Format("2006-01-02"),
		upperBound.Format("2006-01-02"),
	)
	_, err = c.dbh.ExecContext(ctx, query)
	if err != nil {
		log.WithError(err).WithField("query", query).Warnln("could not create sessions closed partition")
	}
}

// InitCrawl inserts a crawl instance into the database in the state `started`.
// This is done to receive a database ID that all subsequent database entities can be linked to.
func (c *Client) InitCrawl(ctx context.Context) (*models.Crawl, error) {
	crawl := &models.Crawl{
		State:     models.CrawlStateStarted,
		StartedAt: time.Now(),
	}
	return crawl, crawl.Insert(ctx, c.dbh, boil.Infer())
}

// UpdateCrawl takes the crawl model an updates it in the database.
func (c *Client) UpdateCrawl(ctx context.Context, crawl *models.Crawl) error {
	_, err := crawl.Update(ctx, c.dbh, boil.Infer())
	return err
}

func (c *Client) PersistCrawlVisit(
	crawlID int,
	peerID peer.ID,
	maddrs []ma.Multiaddr,
	protocolStrs types.StringArray,
	agentVersion string,
	connectDuration time.Duration,
	crawlDuration time.Duration,
	visitStartedAt time.Time,
	visitEndedAt time.Time,
	errorStr string,
) error {
	maddrStrs := utils.MaddrsToAddrs(maddrs)

	protocolStrs, protocolIDs := c.parseProtocols(protocolStrs)
	agentVersionID, found := c.AgentVersionID(agentVersion)

	dbAgentVersionID := null.NewInt(agentVersionID, found)
	dbAgentVersion := null.NewString(agentVersion, !dbAgentVersionID.Valid)
	dbConnectDuration := null.NewString(fmt.Sprintf("%f seconds", connectDuration.Seconds()), connectDuration != 0)
	dbCrawlDuration := null.NewString(fmt.Sprintf("%f seconds", crawlDuration.Seconds()), crawlDuration != 0)
	dbErrorStr := null.NewString(errorStr, errorStr != "")

	rows, err := queries.Raw("SELECT insert_visit($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
		crawlID,
		peerID.String(),
		types.StringArray(maddrStrs),
		protocolStrs,
		protocolIDs,
		dbAgentVersion,
		dbAgentVersionID,
		nil,
		dbConnectDuration,
		dbCrawlDuration,
		visitStartedAt,
		visitEndedAt,
		models.VisitTypeCrawl,
		dbErrorStr,
	).Query(c.dbh)
	if err != nil {
		return err
	}
	return rows.Close()
}

func (c *Client) GetOrCreateProtocol(ctx context.Context, exec boil.ContextExecutor, protocol string) (int, error) {
	if id, found := c.ProtocolID(protocol); found {
		return id, nil
	}

	row := exec.QueryRowContext(ctx, "SELECT upsert_protocol($1)", protocol)
	if row.Err() != nil {
		return 0, errors.Wrap(row.Err(), "unable to upsert protocol")
	}

	var protocolID *int
	if err := row.Scan(&protocolID); err != nil {
		return 0, errors.Wrap(err, "unable to scan result from upsert protocol")
	}

	if protocolID == nil {
		return 0, fmt.Errorf("protocol not created")
	}

	c.protocols.Add(protocol, *protocolID)

	return *protocolID, nil
}

// ProtocolID returns the protocol database id for the given string
func (c *Client) ProtocolID(protocol string) (int, bool) {
	val, ok := c.protocols.Get(protocol)
	if !ok {
		stats.Record(c.ctx, metrics.ProtocolCacheMissCount.M(1))
		return 0, false
	}
	stats.Record(c.ctx, metrics.ProtocolCacheHitCount.M(1))
	return val.(int), true
}

// fillProtocolsCache fetches all rows until protocol cache size from the protocols table and
// initializes the DB clients protocols cache.
func (c *Client) fillProtocolsCache(ctx context.Context) error {
	if c.protocols.Len() == 0 {
		return nil
	}

	prots, err := models.Protocols(qm.Limit(c.protocols.Len())).All(ctx, c.dbh)
	if err != nil {
		return err
	}

	for _, p := range prots {
		c.protocols.Add(p.Protocol, p.ID)
	}

	return nil
}

func (c *Client) parseProtocols(protocols []string) (types.StringArray, types.Int64Array) {
	var protocolStrs []string
	var protocolIDs []int64
	for _, protocol := range protocols {
		if protocolID, found := c.ProtocolID(protocol); found {
			protocolIDs = append(protocolIDs, int64(protocolID))
		} else {
			protocolStrs = append(protocolStrs, protocol)
		}
	}
	return protocolStrs, protocolIDs
}

func (c *Client) GetOrCreateAgentVersion(ctx context.Context, exec boil.ContextExecutor, agentVersion string) (int, error) {
	if id, found := c.AgentVersionID(agentVersion); found {
		return id, nil
	}

	row := exec.QueryRowContext(ctx, "SELECT upsert_agent_version($1)", agentVersion)
	if row.Err() != nil {
		return 0, errors.Wrap(row.Err(), "unable to upsert agent version")
	}

	var agentVersionID *int
	if err := row.Scan(&agentVersionID); err != nil {
		return 0, errors.Wrap(err, "unable to scan result from upsert agent version")
	}

	if agentVersionID == nil {
		return 0, fmt.Errorf("agentVersion not created")
	}

	c.agentVersions.Add(agentVersion, *agentVersionID)

	return *agentVersionID, nil
}

// AgentVersionID returns the agent version database id for the given string
func (c *Client) AgentVersionID(agentVersion string) (int, bool) {
	val, ok := c.agentVersions.Get(agentVersion)
	if !ok {
		stats.Record(c.ctx, metrics.AgentVersionCacheMissCount.M(1))
		return 0, false
	}
	stats.Record(c.ctx, metrics.AgentVersionCacheHitCount.M(1))
	return val.(int), true
}

// fillAgentVersionsCache fetches all rows until agent version cache size from the agent_versions table and
// initializes the DB clients agent version cache.
func (c *Client) fillAgentVersionsCache(ctx context.Context) error {
	if c.agentVersions.Len() == 0 {
		return nil
	}

	avs, err := models.AgentVersions(qm.Limit(c.agentVersions.Len())).All(ctx, c.dbh)
	if err != nil {
		return err
	}

	for _, av := range avs {
		c.agentVersions.Add(av.AgentVersion, av.ID)
	}

	return nil
}

func (c *Client) PersistDialVisit(
	peerID peer.ID,
	maddrs []ma.Multiaddr,
	dialDuration time.Duration,
	visitStartedAt time.Time,
	visitEndedAt time.Time,
	errorStr string,
) error {
	maddrStrs := utils.MaddrsToAddrs(maddrs)

	dbDialDuration := null.NewString(fmt.Sprintf("%f seconds", dialDuration.Seconds()), dialDuration != 0)
	dbErrorStr := null.NewString(errorStr, errorStr != "")

	rows, err := queries.Raw("SELECT insert_visit($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
		nil,
		peerID.String(),
		types.StringArray(maddrStrs),
		nil,
		nil,
		nil,
		nil,
		dbDialDuration,
		nil,
		nil,
		visitStartedAt,
		visitEndedAt,
		models.VisitTypeDial,
		dbErrorStr,
	).Query(c.dbh)
	if err != nil {
		return err
	}
	return rows.Close()
}

func (c *Client) PersistNeighbors(crawl *models.Crawl, peerID peer.ID, errorBits uint16, neighbors []peer.ID) error {
	neighborMHashes := make([]string, len(neighbors))
	for i, neighbor := range neighbors {
		neighborMHashes[i] = neighbor.String()
	}
	// postgres does not support unsigned integers. So we interpret the uint16 as an int16
	bitMask := *(*int16)(unsafe.Pointer(&errorBits))
	rows, err := queries.Raw("SELECT insert_neighbors($1, $2, $3, $4)", crawl.ID, peerID.String(), fmt.Sprintf("{%s}", strings.Join(neighborMHashes, ",")), bitMask).Query(c.dbh)
	if err != nil {
		return err
	}
	return rows.Close()
}

func (c *Client) PersistCrawlProperties(ctx context.Context, crawl *models.Crawl, properties map[string]map[string]int) error {
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

			if property == "protocol" {
				protocolID, err := c.GetOrCreateProtocol(ctx, txn, value)
				if err != nil {
					continue
				}
				cp.ProtocolID = null.IntFrom(protocolID)
			} else if property == "agent_version" {
				agentVersionID, err := c.GetOrCreateAgentVersion(ctx, txn, value)
				if err != nil {
					continue
				}
				cp.AgentVersionID = null.IntFrom(agentVersionID)
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

func (c *Client) QueryBootstrapPeers(ctx context.Context, limit int) ([]peer.AddrInfo, error) {
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

func (c *Client) QueryPeers(ctx context.Context, pis []peer.AddrInfo) (models.PeerSlice, error) {
	mhs := make([]interface{}, len(pis))
	for i, pi := range pis {
		mhs[i] = pi.ID.String()
	}
	return models.Peers(qm.WhereIn(models.PeerColumns.MultiHash+" in ?", mhs...)).All(ctx, c.dbh)
}

//func (c *Client) InsertLatencies(ctx context.Context, peer *models.Peer, latencies []*models.Latency) error {
//	txn, err := c.dbh.BeginTx(ctx, nil)
//	if err != nil {
//		return errors.Wrap(err, "create latencies txn")
//	}
//
//	for _, latency := range latencies {
//		if err := latency.SetPeer(ctx, c.dbh, false, peer); err != nil {
//			return errors.Wrap(err, "associating peer with latency")
//		}
//
//		if err := latency.Insert(ctx, txn, boil.Infer()); err != nil {
//			return errors.Wrap(err, "insert latency measurement")
//		}
//	}
//
//	if err = txn.Commit(); err != nil {
//		return errors.Wrap(err, "commit latencies txn")
//	}
//	return nil
//}

//func (c *Client) FetchDueSessions(ctx context.Context) (models.SessionSlice, error) {
//	return models.Sessions(
//		qm.Where("next_dial_attempt - NOW() < '10s'::interval"),
//		qm.Load(models.SessionRels.Peer),
//		qm.Load(qm.Rels(models.SessionRels.Peer, models.PeerRels.MultiAddresses)),
//	).All(ctx, c.dbh)
//}
//
//func (c *Client) FetchUnresolvedMultiAddresses(ctx context.Context, limit int) (models.MultiAddressSlice, error) {
//	return models.MultiAddresses(
//		qm.Where(models.MultiAddressColumns.ID+" > coalesce((SELECT max("+models.MultiAddressesXIPAddressColumns.MultiAddressID+") FROM "+models.TableNames.MultiAddressesXIPAddresses+"), 0)"),
//		qm.OrderBy(models.MultiAddressColumns.ID),
//		qm.Limit(limit),
//	).All(ctx, c.dbh)
//}
//
//func ToAddrInfo(p *models.Peer) (peer.AddrInfo, error) {
//	pi := peer.AddrInfo{
//		Addrs: []ma.Multiaddr{},
//	}
//	peerID, err := peer.Decode(p.MultiHash)
//	if err != nil {
//		return pi, err
//	}
//	pi.ID = peerID
//
//	for _, dbmaddr := range p.R.MultiAddresses {
//		maddr, err := ma.NewMultiaddr(dbmaddr.Maddr)
//		if err != nil {
//			return pi, err
//		}
//		pi.Addrs = append(pi.Addrs, maddr)
//	}
//	return pi, nil
//}

// Rollback calls rollback on the given transaction and logs the potential error.
func Rollback(txn *sql.Tx) {
	if err := txn.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		log.WithError(err).Warnln("An error occurred when rolling back transaction")
	}
}
