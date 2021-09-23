package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	_ "github.com/lib/pq"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

const (
	MinInterval        = 30 * time.Second
	MaxInterval        = 15 * time.Minute
	IntervalMultiplier = 0.5
)

func Open(ctx context.Context) (*sql.DB, error) {
	conf, err := config.FromContext(ctx)
	if err != nil {
		return nil, err
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
	db, err := sql.Open(driverName, srcName)
	if err != nil {
		return nil, err
	}

	// Ping database to verify connection.
	if err = db.Ping(); err != nil {
		return nil, errors.Wrap(err, "pinging database")
	}

	return db, nil
}

func FetchSession(ctx context.Context, db *sql.DB, peerID string) (*models.Session, error) {
	return models.Sessions(qm.Where("peer_id = ?", peerID)).One(ctx, db)
}

func UpsertSessionSuccess(dbh *sql.DB, peerID int) error {
	// TODO: use config for min interval and factor
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
	rows, err := queries.Raw(query, peerID, MinInterval.String(), MaxInterval.String(), IntervalMultiplier).Query(dbh)
	if err != nil {
		return err
	}

	return rows.Close()
}

func UpsertSessionError(dbh *sql.DB, peerID int, failedAt time.Time, reason string) error {
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
	rows, err := queries.Raw(query, peerID, failedAt.Format(time.RFC3339Nano), reason).Query(dbh)
	if err != nil {
		return err
	}
	return rows.Close()
}

func UpsertPeer(dbh *sql.DB, peerID string, maddrs []ma.Multiaddr) (int, types.StringArray, error) {
	maddrStrs := make(types.StringArray, len(maddrs))
	for i, maddr := range maddrs {
		maddrStrs[i] = maddr.String()
	}

	query := `
INSERT INTO peers (
  multi_hash,
  multi_addresses,
  created_at,
  updated_at
) VALUES ($1, $2, NOW(), NOW())
ON CONFLICT (multi_hash) DO UPDATE SET
  multi_addresses     = EXCLUDED.multi_addresses,
  old_multi_addresses = peers.multi_addresses,
  updated_at          = EXCLUDED.updated_at
RETURNING id, old_multi_addresses;
`
	rows, err := queries.Raw(query, peerID, maddrStrs).Query(dbh)
	if err != nil {
		return 0, nil, err
	}

	if ok := rows.Next(); !ok {
		return 0, nil, rows.Err()
	}

	var id int
	var oldMaddrs types.StringArray
	if err = rows.Scan(&id, &oldMaddrs); err != nil {
		return 0, nil, err
	}

	return id, oldMaddrs, rows.Close()
}

func FetchDueSessions(ctx context.Context, dbh *sql.DB) (models.SessionSlice, error) {
	return models.Sessions(
		qm.Where("next_dial_attempt - NOW() < '10s'::interval"),
		qm.Load(models.SessionRels.Peer),
		qm.Load(qm.Rels(models.SessionRels.Peer, models.PeerRels.MultiAddresses)),
	).All(ctx, dbh)
}
