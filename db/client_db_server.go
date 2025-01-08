package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dennis-tra/nebula-crawler/db/models"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/dennis-tra/nebula-crawler/config"
)

type DBServerClient struct {
	ctx context.Context

	// Reference to the configuration
	cfg *config.Database

	// Database handler
	dbh *sql.DB

	// reference to all relevant db telemetry
	telemetry *telemetry
}

var _ ServerClient = (*DBServerClient)(nil)

// InitDBServerClient establishes a database connection with the provided
// configuration
func InitDBServerClient(ctx context.Context, cfg *config.Database) (*DBServerClient, error) {
	log.WithFields(log.Fields{
		"host": cfg.DatabaseHost,
		"port": cfg.DatabasePort,
		"name": cfg.DatabaseName,
		"user": cfg.DatabaseUser,
		"ssl":  cfg.DatabaseSSLMode,
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

	client := &DBServerClient{ctx: ctx, cfg: cfg, dbh: dbh, telemetry: telemetry}

	return client, nil
}

func (d *DBServerClient) Close() error {
	return d.dbh.Close()
}

func (d *DBServerClient) GetPeer(ctx context.Context, multiHash string) (*models.Peer, models.ProtocolSlice, error) {
	// write a hand-crafted query to avoid two DB round-trips

	dbPeer, err := models.Peers(
		models.PeerWhere.MultiHash.EQ(multiHash),
		qm.Load(models.PeerRels.AgentVersion),
		qm.Load(models.PeerRels.MultiAddresses),
		qm.Load(models.PeerRels.ProtocolsSet),
	).One(ctx, d.dbh)
	if err != nil {
		return nil, nil, fmt.Errorf("getting peer: %w", err)
	}

	if dbPeer.R.ProtocolsSet == nil {
		return dbPeer, nil, nil
	}

	protocolIDs := dbPeer.R.ProtocolsSet.ProtocolIds
	ids := make([]int, 0, len(protocolIDs))
	for _, id := range protocolIDs {
		ids = append(ids, int(id))
	}

	dbProtocols, err := models.Protocols(models.ProtocolWhere.ID.IN(ids)).All(ctx, d.dbh)
	if err != nil {
		return dbPeer, nil, fmt.Errorf("getting protocols: %w", err)
	}

	return dbPeer, dbProtocols, nil
}
