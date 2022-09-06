package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/volatiletech/sqlboiler/v4/types"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
	_ "github.com/lib/pq"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Client struct {
	// Database handler
	dbh *sql.DB

	// Holds database properties entities for caching
	propLk sync.RWMutex
}

func InitClient(conf *config.Config) (*Client, error) {
	log.WithFields(log.Fields{
		"host": conf.DatabaseHost,
		"port": conf.DatabasePort,
		"name": conf.DatabaseName,
		"user": conf.DatabaseUser,
	}).Infoln("Initializing database client")

	driverName, err := ocsql.Register("postgres")
	if err != nil {
		return nil, errors.Wrap(err, "register ocsql")
	}

	// Open database handle
	srcName := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		conf.DatabaseHost,
		conf.DatabasePort,
		conf.DatabaseName,
		conf.DatabaseUser,
		conf.DatabasePassword,
		conf.DatabaseSSLMode,
	)
	dbh, err := sql.Open(driverName, srcName)
	if err != nil {
		return nil, errors.Wrap(err, "opening database")
	}

	// Ping database to verify connection.
	if err = dbh.Ping(); err != nil {
		return nil, errors.Wrap(err, "pinging database")
	}

	return &Client{
		dbh: dbh,
	}, nil
}

func (c *Client) Handle() *sql.DB {
	return c.dbh
}

// InitCrawl inserts a crawl instance in the database in the state `started`.
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
	protocolIDs types.Int64Array,
	agentVersion string,
	agentVersionID int,
	connectDuration time.Duration,
	crawlDuration time.Duration,
	visitStartedAt time.Time,
	visitEndedAt time.Time,
	errorStr string,
) error {
	maddrStrs := utils.MaddrsToAddrs(maddrs)

	dbAgentVersionID := null.NewInt(agentVersionID, agentVersionID != 0)
	dbAgentVersion := null.NewString(agentVersion, agentVersion != "" && !dbAgentVersionID.Valid)
	dbConnectDuration := null.NewString(fmt.Sprintf("%f seconds", connectDuration.Seconds()), connectDuration != 0)
	dbCrawlDuration := null.NewString(fmt.Sprintf("%f seconds", crawlDuration.Seconds()), crawlDuration != 0)
	dbErrorStr := null.NewString(errorStr, errorStr != "")

	rows, err := queries.Raw("SELECT insert_visit($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
		crawlID,
		peerID.Pretty(),
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
		peerID.Pretty(),
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

// GetAllAgentVersions fetches all rows from the agent_versions table and returns a map that's indexed
// by the agent version string.
func (c *Client) GetAllAgentVersions(ctx context.Context) (map[string]*models.AgentVersion, error) {
	avs, err := models.AgentVersions().All(ctx, c.dbh)
	if err != nil {
		return nil, err
	}

	avMap := map[string]*models.AgentVersion{}
	for _, av := range avs {
		avMap[av.AgentVersion] = av
	}

	return avMap, nil
}

// GetAllProtocols fetches all rows from the protocols table and returns a map that's indexed
// by the protocol string.
func (c *Client) GetAllProtocols(ctx context.Context) (map[string]*models.Protocol, error) {
	ps, err := models.Protocols().All(ctx, c.dbh)
	if err != nil {
		return nil, err
	}

	pMap := map[string]*models.Protocol{}
	for _, p := range ps {
		pMap[p.Protocol] = p
	}

	return pMap, nil
}

func (c *Client) GetOrCreateProtocol(ctx context.Context, exec boil.ContextExecutor, protocol string) (*models.Protocol, error) {
	c.propLk.Lock()
	defer c.propLk.Unlock()

	p, err := models.Protocols(qm.Where(models.ProtocolColumns.Protocol+" = ?", protocol)).One(ctx, c.dbh)
	if err == nil {
		return p, nil
	}

	p = &models.Protocol{
		Protocol: protocol,
	}
	return p, p.Upsert(ctx, exec, true,
		[]string{models.ProtocolColumns.Protocol},
		boil.Whitelist(models.ProtocolColumns.UpdatedAt),
		boil.Infer(),
	)
}

func (c *Client) GetOrCreateAgentVersion(ctx context.Context, exec boil.ContextExecutor, agentVersion string) (*models.AgentVersion, error) {
	c.propLk.Lock()
	defer c.propLk.Unlock()

	av, err := models.AgentVersions(qm.Where(models.AgentVersionColumns.AgentVersion+" = ?", agentVersion)).One(ctx, c.dbh)
	if err == nil {
		return av, nil
	}

	av = &models.AgentVersion{
		AgentVersion: agentVersion,
	}
	return av, av.Upsert(ctx, exec, true,
		[]string{models.AgentVersionColumns.AgentVersion},
		boil.Whitelist(models.AgentVersionColumns.UpdatedAt),
		boil.Infer(),
	)
}

func (c *Client) PersistNeighbors(crawl *models.Crawl, peerID peer.ID, errorBits uint16, neighbors []peer.ID) error {
	neighborMHashes := make([]string, len(neighbors))
	for i, neighbor := range neighbors {
		neighborMHashes[i] = neighbor.Pretty()
	}
	// postgres does not support unsigned integers. So we interpret the uint16 as an int16
	bitMask := *(*int16)(unsafe.Pointer(&errorBits))
	rows, err := queries.Raw("SELECT insert_neighbors($1, $2, $3, $4)", crawl.ID, peerID.Pretty(), fmt.Sprintf("{%s}", strings.Join(neighborMHashes, ",")), bitMask).Query(c.dbh)
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

	for property, valuesMap := range properties {
		for value, count := range valuesMap {

			cp := &models.CrawlProperty{
				CrawlID: crawl.ID,
				Count:   count,
			}

			if property == "protocol" {
				p, err := c.GetOrCreateProtocol(ctx, txn, value)
				if err != nil {
					continue
				}
				cp.ProtocolID = null.IntFrom(p.ID)
			} else if property == "agent_version" {
				av, err := c.GetOrCreateAgentVersion(ctx, txn, value)
				if err != nil {
					continue
				}
				cp.AgentVersionID = null.IntFrom(av.ID)
			} else if property == "error" {
				cp.Error = null.StringFrom(value)
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
		qm.InnerJoin("sessions s on s.peer_id = peers.id"),
		qm.Where("s.finished = false"),
		qm.OrderBy("s.updated_at"),
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
		mhs[i] = pi.ID.Pretty()
	}
	return models.Peers(qm.WhereIn(models.PeerColumns.MultiHash+" in ?", mhs...)).All(ctx, c.dbh)
}

func (c *Client) InsertLatencies(ctx context.Context, peer *models.Peer, latencies []*models.Latency) error {
	txn, err := c.dbh.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "create latencies txn")
	}

	for _, latency := range latencies {
		if err := latency.SetPeer(ctx, c.dbh, false, peer); err != nil {
			return errors.Wrap(err, "associating peer with latency")
		}

		if err := latency.Insert(ctx, txn, boil.Infer()); err != nil {
			return errors.Wrap(err, "insert latency measurement")
		}
	}

	if err = txn.Commit(); err != nil {
		return errors.Wrap(err, "commit latencies txn")
	}
	return nil
}

func (c *Client) FetchDueSessions(ctx context.Context) (models.SessionSlice, error) {
	return models.Sessions(
		qm.Where("next_dial_attempt - NOW() < '10s'::interval"),
		qm.Load(models.SessionRels.Peer),
		qm.Load(qm.Rels(models.SessionRels.Peer, models.PeerRels.MultiAddresses)),
	).All(ctx, c.dbh)
}

func (c *Client) FetchUnresolvedMultiAddresses(ctx context.Context, limit int) (models.MultiAddressSlice, error) {
	return models.MultiAddresses(
		qm.Where(models.MultiAddressColumns.ID+" > coalesce((SELECT max("+models.MultiAddressesXIPAddressColumns.MultiAddressID+") FROM "+models.TableNames.MultiAddressesXIPAddresses+"), 0)"),
		qm.OrderBy(models.MultiAddressColumns.ID),
		qm.Limit(limit),
	).All(ctx, c.dbh)
}

func ToAddrInfo(p *models.Peer) (peer.AddrInfo, error) {
	pi := peer.AddrInfo{
		Addrs: []ma.Multiaddr{},
	}
	peerID, err := peer.Decode(p.MultiHash)
	if err != nil {
		return pi, err
	}
	pi.ID = peerID

	for _, dbmaddr := range p.R.MultiAddresses {
		maddr, err := ma.NewMultiaddr(dbmaddr.Maddr)
		if err != nil {
			return pi, err
		}
		pi.Addrs = append(pi.Addrs, maddr)
	}
	return pi, nil
}
