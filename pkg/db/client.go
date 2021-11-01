package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"contrib.go.opencensus.io/integrations/ocsql"
	_ "github.com/lib/pq"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
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
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		conf.DatabaseHost,
		conf.DatabasePort,
		conf.DatabaseName,
		conf.DatabaseUser,
		conf.DatabasePassword,
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

func (c *Client) InsertRawVisit(ctx context.Context, re *models.RawVisit) error {
	return re.Insert(ctx, c.dbh, boil.Infer())
}

func (c *Client) InitCrawl(ctx context.Context) (*models.Crawl, error) {
	crawl := &models.Crawl{
		State:     models.CrawlStateStarted,
		StartedAt: time.Now(),
	}
	return crawl, crawl.Insert(ctx, c.dbh, boil.Infer())
}

func (c *Client) UpdateCrawl(ctx context.Context, crawl *models.Crawl) error {
	_, err := crawl.Update(ctx, c.dbh, boil.Infer())
	return err
}

func (c *Client) GetOrCreateProtocol(ctx context.Context, exec boil.ContextExecutor, protocol string) (*models.Protocol, error) {
	c.propLk.Lock()
	defer c.propLk.Unlock()

	p := &models.Protocol{
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

	av := &models.AgentVersion{
		AgentVersion: agentVersion,
	}
	return av, av.Upsert(ctx, exec, true,
		[]string{models.AgentVersionColumns.AgentVersion},
		boil.Whitelist(models.AgentVersionColumns.UpdatedAt),
		boil.Infer(),
	)
}

func (c *Client) PersistNeighbors(crawl *models.Crawl, peerID peer.ID, neighbors []peer.ID) error {
	neighborMHashes := make([]string, len(neighbors))
	for i, neighbor := range neighbors {
		neighborMHashes[i] = neighbor.Pretty()
	}
	rows, err := queries.Raw("SELECT insert_neighbors($1, $2, $3)", crawl.ID, peerID.Pretty(), fmt.Sprintf("{%s}", strings.Join(neighborMHashes, ","))).Query(c.dbh)
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

func (c *Client) FetchMultiAddresses(ctx context.Context, offset int, limit int) (models.MultiAddressSlice, error) {
	return models.MultiAddresses(
		qm.Offset(offset),
		qm.Limit(limit),
	).All(ctx, c.dbh)
}

func (c *Client) FetchUnresolvedMultiAddresses(ctx context.Context, offset int, limit int) (models.MultiAddressSlice, error) {
	return models.MultiAddresses(
		qm.LeftOuterJoin(models.TableNames.MultiAddressesXIPAddresses+" maxia on maxia."+models.MultiAddressesXIPAddressColumns.MultiAddressID+" = "+models.TableNames.MultiAddresses+".id"),
		qm.Where("maxia."+models.MultiAddressesXIPAddressColumns.MultiAddressID+" IS NULL"),
		qm.Offset(offset),
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

func (c *Client) DeletePreviousRawVisits(ctx context.Context) error {
	// Get most recent successful crawl
	crawl, err := models.Crawls(
		qm.OrderBy(models.CrawlColumns.FinishedAt+" DESC"),
		qm.Limit(1),
	).One(ctx, c.dbh)
	if err != nil {
		return err
	}

	// Delete all raw visits of that crawl
	deleted, err := models.RawVisits(
		qm.Where(models.RawVisitColumns.CreatedAt+" < ?", crawl.StartedAt),
	).DeleteAll(ctx, c.dbh)
	log.Debugf("Deleted %d previous obsolete raw_visit rows", deleted)
	return err
}
