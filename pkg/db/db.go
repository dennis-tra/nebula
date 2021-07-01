package db

import (
	"context"
	"database/sql"
	"fmt"

	"contrib.go.opencensus.io/integrations/ocsql"
	_ "github.com/lib/pq"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
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

	// Open handle to database
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

	boil.SetDB(db)

	return db, nil
}

func UpsertSessionSuccess(dbh *sql.DB, peerID string) error {
	query := `
INSERT INTO sessions
    (
		 peer_id,
		 first_successful_dial,
		 last_successful_dial,
		 first_failed_dial,
		 successful_dials,
		 finished,
		 created_at,
		 updated_at
     ) VALUES ($1, NOW(), NOW(), '1970-01-01', 1, false, NOW(), NOW())
     ON CONFLICT ON CONSTRAINT uq_peer_id_first_failed_dial
     DO UPDATE SET 
		last_successful_dial = EXCLUDED.last_successful_dial,
		successful_dials = sessions.successful_dials + 1,
		updated_at = EXCLUDED.updated_at;
`
	rows, err := queries.Raw(query, peerID).Query(dbh)
	if err != nil {
		return err
	}

	return rows.Close()
}

func UpsertSessionError(dbh *sql.DB, peerID string) error {
	query := `
UPDATE sessions SET
    first_failed_dial = NOW(),
    min_duration = last_successful_dial - first_successful_dial,
    max_duration = NOW() - first_successful_dial,
    finished = true,
    updated_at = NOW(),
    next_dial_attempt = null
WHERE peer_id = $1;
`
	rows, err := queries.Raw(query, peerID).Query(dbh)
	if err != nil {
		return err
	}
	return rows.Close()
}

func UpsertPeer(ctx context.Context, dbh *sql.DB, peerID string, maddrs []ma.Multiaddr) error {
	p := &models.Peer{
		ID:             peerID,
		MultiAddresses: make([]string, len(maddrs)),
	}
	for i, maddr := range maddrs {
		p.MultiAddresses[i] = maddr.String()
	}
	return p.Upsert(ctx, dbh, true, []string{"id"}, boil.Whitelist("updated_at"), boil.Infer())
}
