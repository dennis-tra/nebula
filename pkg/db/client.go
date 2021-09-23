package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/libp2p/go-libp2p-core/peer"
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
)

type Client struct {
	// Database handler
	dbh *sql.DB
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

	return &Client{
		dbh: dbh,
	}, nil
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

func (c *Client) PersistPeerProperties(ctx context.Context, crawl *models.Crawl, properties map[string]map[string]int) error {
	txn, err := c.dbh.BeginTx(ctx, nil)
	if err != nil {
		return errors.New("start peer property txn")
	}

	for property, valuesMap := range properties {
		for value, count := range valuesMap {
			pp := &models.PeerProperty{
				CrawlID:  crawl.ID,
				Property: property,
				Value:    value,
				Count:    count,
			}

			if err := pp.Insert(ctx, txn, boil.Infer()); err != nil {
				log.WithError(err).WithFields(log.Fields{
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
		qm.Select(models.PeerColumns.MultiHash, models.PeerColumns.MultiAddresses),
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
		for _, maddrStr := range p.MultiAddresses {
			maddr, err := ma.NewMultiaddr(maddrStr)
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

func (c *Client) UpsertPeer(ctx context.Context, pi peer.AddrInfo, agent string) (*models.Peer, error) {
	maddrs := make(types.StringArray, len(pi.Addrs))
	for i, maddr := range pi.Addrs {
		maddrs[i] = maddr.String()
	}
	p := &models.Peer{
		MultiHash:      pi.ID.Pretty(),
		MultiAddresses: maddrs,
		AgentVersion:   null.StringFrom(agent),
	}
	return p, p.Upsert(ctx, c.dbh, true, []string{models.PeerColumns.MultiHash}, boil.Whitelist(models.PeerColumns.UpdatedAt), boil.Infer())
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

func (c *Client) UpsertSessionSuccess(peer *models.Peer) error {
	query := `
INSERT INTO sessions (
  peer_id,
  first_successful_dial,
  last_successful_dial,
  first_failed_dial,
  next_dial_attempt,
  successful_dials,
  finished,
  created_at,
  updated_at
) VALUES ($1, NOW(), NOW(), '1970-01-01', NOW() + $2::interval, 1, false, NOW(), NOW())
ON CONFLICT ON CONSTRAINT uq_peer_id_first_failed_dial DO UPDATE SET
  last_successful_dial = EXCLUDED.last_successful_dial,
  successful_dials     = sessions.successful_dials + 1,
  updated_at           = EXCLUDED.updated_at,
  next_dial_attempt    = 
   CASE
     WHEN $4 * (EXCLUDED.last_successful_dial - sessions.first_successful_dial) < $2::interval THEN
       EXCLUDED.last_successful_dial + $2::interval
     WHEN $4 * (EXCLUDED.last_successful_dial - sessions.first_successful_dial) > $3::interval THEN
       EXCLUDED.last_successful_dial + $3::interval
     ELSE
       EXCLUDED.last_successful_dial + $4 * (EXCLUDED.last_successful_dial - sessions.first_successful_dial)
   END;
`
	rows, err := queries.Raw(query, peer.ID, MinInterval.String(), MaxInterval.String(), IntervalMultiplier).Query(c.dbh)
	if err != nil {
		return err
	}

	return rows.Close()
}

func (c *Client) UpsertSessionError(peer *models.Peer, failedAt time.Time, reason string) error {
	query := `
UPDATE sessions SET
  first_failed_dial = $2,
  min_duration = last_successful_dial - first_successful_dial,
  max_duration = NOW() - first_successful_dial,
  finished = true,
  updated_at = NOW(),
  next_dial_attempt = null,
  finish_reason = $3
WHERE peer_id = $1 AND finished = false;
`
	rows, err := queries.Raw(query, peer.ID, failedAt.Format(time.RFC3339Nano), reason).Query(c.dbh)
	if err != nil {
		return err
	}

	return rows.Close()
}
