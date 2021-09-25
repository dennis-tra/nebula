package db

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/volatiletech/sqlboiler/v4/queries"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Client struct {
	// Database handler
	dbh *sql.DB

	// Holds database properties entities for caching
	propLk  sync.RWMutex
	propMap map[string]*models.Property
}

func InitClient(ctx context.Context) (*Client, error) {
	conf, err := config.FromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "config from context")
	}

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

	properties, err := models.Properties().All(ctx, dbh)
	if err != nil {
		return nil, err
	}

	propMap := map[string]*models.Property{}
	for _, p := range properties {
		propMap[p.Property+p.Value] = p
	}

	return &Client{
		dbh:     dbh,
		propMap: propMap,
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

func (c *Client) GetOrCreateProperty(ctx context.Context, exec boil.ContextExecutor, property string, value string) (*models.Property, error) {
	c.propLk.Lock()
	defer c.propLk.Unlock()

	if prop, ok := c.propMap[property+value]; ok {
		return prop, nil
	}

	p := &models.Property{
		Property: property,
		Value:    value,
	}
	return p, p.Upsert(ctx, exec, true,
		[]string{models.PropertyColumns.Property, models.PropertyColumns.Value},
		boil.Whitelist(models.PropertyColumns.UpdatedAt),
		boil.Infer(),
	)
}

func (c *Client) PersistCrawlProperties(ctx context.Context, crawl *models.Crawl, properties map[string]map[string]int) error {
	txn, err := c.dbh.BeginTx(ctx, nil)
	if err != nil {
		return errors.New("start peer property txn")
	}

	for property, valuesMap := range properties {
		for value, count := range valuesMap {

			p, err := c.GetOrCreateProperty(ctx, txn, property, value)
			if err != nil {
				continue
			}

			cp := &models.CrawlProperty{
				CrawlID:    crawl.ID,
				PropertyID: p.ID,
				Count:      count,
			}

			if err := cp.Insert(ctx, txn, boil.Infer()); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"crawlId":    crawl.ID,
					"propertyId": p.ID,
					"property":   property,
					"value":      value,
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

func (c *Client) UpsertPeer(ctx context.Context, pi peer.AddrInfo, agent string, protocols []string) (*models.Peer, error) {
	txn, err := c.dbh.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "begin txn")
	}

	p := &models.Peer{MultiHash: pi.ID.Pretty()}
	if err = p.Upsert(ctx, txn, true,
		[]string{models.PeerColumns.MultiHash},
		boil.Whitelist(models.PeerColumns.UpdatedAt),
		boil.Infer(),
	); err != nil {
		return nil, errors.Wrap(err, "upsert peer")
	}

	var props []*models.Property
	for _, protocol := range protocols {
		prop, err := c.GetOrCreateProperty(ctx, c.dbh, "protocol", protocol)
		if err != nil {
			return nil, errors.Wrap(err, "get or create protocol property")
		}
		props = append(props, prop)
	}

	if agent != "" {
		agentProp, err := c.GetOrCreateProperty(ctx, c.dbh, "agent_version", agent)
		if err != nil {
			return nil, errors.Wrap(err, "get or create agent version property")
		}
		props = append(props, agentProp)
	}

	if err = p.SetProperties(ctx, txn, false, props...); err != nil {
		return nil, errors.Wrap(err, "set peer properties")
	}

	// TODO: we need to sort the multi addresses for insertion. See:
	// https://stackoverflow.com/questions/59017059/postgres-sharelock-deadlock-on-transaction
	// I received the same error if the addresses were not sorted.
	maddrStrs := make([]string, len(pi.Addrs))
	for i, maddr := range pi.Addrs {
		maddrStrs[i] = maddr.String()
	}
	sort.Strings(maddrStrs)

	var maddrs []*models.MultiAddress
	for _, maddrStr := range maddrStrs {
		ma := &models.MultiAddress{Maddr: maddrStr}
		if err = ma.Upsert(ctx, txn, true,
			[]string{models.MultiAddressColumns.Maddr},
			boil.Whitelist(models.MultiAddressColumns.UpdatedAt), boil.Infer(),
		); err != nil {
			return nil, errors.Wrap(err, "upsert multi address")
		}
		maddrs = append(maddrs, ma)
	}

	if err = p.SetMultiAddresses(ctx, txn, false, maddrs...); err != nil {
		return nil, errors.Wrap(err, "set multi addresses for peer")
	}
	if err = txn.Commit(); err != nil {
		_ = txn.Rollback()
		return nil, err
	}

	return p, nil
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

func (c *Client) InsertNeighbors(ctx context.Context, crawl *models.Crawl, dbPeer *models.Peer, neighbors []peer.AddrInfo) error {
	txn, err := c.dbh.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin neighbors txn")
	}

	var mhashes []interface{}
	for _, neighbor := range neighbors {
		mhashes = append(mhashes, neighbor.ID.Pretty())
	}

	neighborPeers, err := models.Peers(qm.WhereIn("multi_hash in ?", mhashes...)).All(ctx, txn)
	if err != nil {
		return errors.Wrap(err, "getting neighbors")
	}

OUTER:
	for _, n := range neighbors {
		for _, n2 := range neighborPeers {
			if n.ID.Pretty() == n2.MultiHash {
				neighbor := &models.Neighbor{
					CrawlID:    crawl.ID,
					PeerID:     dbPeer.ID,
					NeighborID: n2.ID,
				}
				if err = neighbor.Insert(ctx, txn, boil.Infer()); err != nil {
					return errors.Wrap(err, "inserting neighbor")
				}
				continue OUTER
			}
		}
		pp, err := c.UpsertPeer(ctx, peer.AddrInfo{ID: n.ID}, "", []string{})
		if err != nil {
			return errors.Wrap(err, "upserting peer")
		}
		neighbor := &models.Neighbor{
			CrawlID:    crawl.ID,
			PeerID:     dbPeer.ID,
			NeighborID: pp.ID,
		}
		if err = neighbor.Insert(ctx, txn, boil.Infer()); err != nil {
			return errors.Wrap(err, "inserting neighbor")
		}
	}

	if err = txn.Commit(); err != nil {
		_ = txn.Rollback()
		return err
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
	b := models.MultiAddressSlice{}
	err := queries.Raw(`
SELECT *
FROM multi_addresses ma
WHERE NOT EXISTS(
        SELECT
        FROM multi_addresses_x_ip_addresses maxia
        WHERE ma.id = maxia.multi_address_id
    )
LIMIT $1 OFFSET $2`, limit, offset).Bind(ctx, c.dbh, &b)
	return b, err
}
