package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/maxmind"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/udger"
)

var resolveConfig = &config.Resolve{
	Root:            rootConfig,
	FilePathUdgerDB: "",
	BatchSize:       1000,
}

// ResolveCommand contains the monitor sub-command configuration.
var ResolveCommand = &cli.Command{
	Name:   "resolve",
	Usage:  "Resolves all multi addresses to their IP addresses and geo location information",
	Action: ResolveAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "udger-db",
			Usage:       "Location of the Udger database v3",
			EnvVars:     []string{"NEBULA_RESOLVE_UDGER_DB"},
			Destination: &resolveConfig.FilePathUdgerDB,
		},
		&cli.IntFlag{
			Name:        "batch-size",
			Usage:       "How many database entries should be fetched at each iteration",
			EnvVars:     []string{"NEBULA_RESOLVE_BATCH_SIZE"},
			Value:       resolveConfig.BatchSize,
			Destination: &resolveConfig.BatchSize,
		},
	},
}

// ResolveAction is the function that is called when running `nebula resolve`.
func ResolveAction(c *cli.Context) error {
	log.Infoln("Starting Nebula multi address resolver...")

	// Initialize the database client
	dbc, err := db.InitDBClient(c.Context, rootConfig.Database)
	if err != nil {
		return err
	}

	// can't bother extracting the limited functionality below to a separate
	// package, so we're operating on the handle directly
	dbh := dbc.Handle()

	// Initialize new maxmind client to interact with the country database.
	mmc, err := maxmind.NewClient()
	if err != nil {
		return err
	}

	var uclient *udger.Client
	if resolveConfig.FilePathUdgerDB != "" {
		uclient, err = udger.NewClient(resolveConfig.FilePathUdgerDB)
		if err != nil {
			return err
		}
	}

	limit := c.Int("batch-size")

	// Start the main loop
	for {
		log.Infoln("Fetching multi addresses...")
		dbmaddrs, err := dbc.FetchUnresolvedMultiAddresses(c.Context, limit)
		if errors.Is(err, context.Canceled) {
			return nil
		} else if err != nil {
			return fmt.Errorf("fetching multi addresses: %w", err)
		}
		log.Infof("Fetched %d multi addresses", len(dbmaddrs))
		if len(dbmaddrs) == 0 {
			return nil
		}

		if err = resolve(c.Context, dbh, mmc, uclient, dbmaddrs); err != nil && !errors.Is(err, context.Canceled) {
			log.WithError(err).Warnln("Error resolving multi addresses")
		}
	}
}

// Resolve save the resolved IP addresses + their countries in a transaction
func resolve(ctx context.Context, dbh *sql.DB, mmc *maxmind.Client, uclient *udger.Client, dbmaddrs models.MultiAddressSlice) error {
	log.WithField("size", len(dbmaddrs)).Infoln("Resolving batch of multi addresses...")

	for _, dbmaddr := range dbmaddrs {
		if err := resolveAddr(ctx, dbh, mmc, uclient, dbmaddr); err != nil {
			log.WithField("maddr", dbmaddr.Maddr).WithError(err).Warnln("Error resolving multi address")
		}
	}

	return nil
}

func resolveAddr(ctx context.Context, dbh *sql.DB, mmc *maxmind.Client, uclient *udger.Client, dbmaddr *models.MultiAddress) error {
	logEntry := log.WithField("maddr", dbmaddr.Maddr)
	txn, err := dbh.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin txn: %w", err)
	}
	defer db.Rollback(txn)

	maddr, err := ma.NewMultiaddr(dbmaddr.Maddr)
	if err != nil {
		logEntry.WithError(err).Warnln("Error parsing multi address - deleting row")
		if _, err = dbmaddr.Delete(ctx, txn); err != nil {
			logEntry.WithError(err).Warnln("Error deleting multi address")
		}
		return fmt.Errorf("parse multi address: %w", err)
	}

	dbmaddr.Resolved = true
	dbmaddr.IsPublic = null.BoolFrom(manet.IsPublicAddr(maddr))
	dbmaddr.IsRelay = null.BoolFrom(isRelayedMaddr(maddr))

	addrInfos, err := mmc.MaddrInfo(ctx, maddr)
	if err != nil {
		logEntry.WithError(err).Warnln("Error deriving address information from maddr ", maddr)
	}

	if len(addrInfos) == 0 {
		dbmaddr.HasManyAddrs = null.BoolFrom(false)
	} else if len(addrInfos) == 1 {
		dbmaddr.HasManyAddrs = null.BoolFrom(false)
		var addr string
		var addrInfo *maxmind.AddrInfo
		for k, v := range addrInfos {
			addr, addrInfo = k, v
			break
		}
		dbmaddr.Asn = null.NewInt(int(addrInfo.ASN), addrInfo.ASN != 0)
		dbmaddr.Country = null.NewString(addrInfo.Country, addrInfo.Country != "")
		dbmaddr.Continent = null.NewString(addrInfo.Continent, addrInfo.Continent != "")
		dbmaddr.Addr = null.NewString(addr, addr != "")

		if uclient != nil {
			datacenterID, err := uclient.Datacenter(addr)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				logEntry.WithError(err).WithField("addr", addr).Warnln("Error resolving ip address to datacenter")
			}
			dbmaddr.IsCloud = null.NewInt(datacenterID, datacenterID != 0)
		}
	} else if len(addrInfos) > 1 { // not "else" because the MaddrInfo could have failed and we still want to update the maddr
		dbmaddr.HasManyAddrs = null.BoolFrom(true)
		// Due to dnsaddr protocols each multi address can point to multiple
		// IP addresses each in a different country.
		for addr, addrInfo := range addrInfos {

			datacenterID := 0
			if uclient != nil {
				datacenterID, err = uclient.Datacenter(addr)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					logEntry.WithError(err).WithField("addr", addr).Warnln("Error resolving ip address to datacenter")
				} else if datacenterID > 0 {
					dbmaddr.IsCloud = null.IntFrom(datacenterID)
				}
			}

			// Save the IP address + country information + asn information
			ipaddr := &models.IPAddress{
				Asn:       null.NewInt(int(addrInfo.ASN), addrInfo.ASN != 0),
				IsCloud:   null.NewInt(datacenterID, datacenterID != 0),
				Country:   null.NewString(addrInfo.Country, addrInfo.Country != ""),
				Continent: null.NewString(addrInfo.Continent, addrInfo.Continent != ""),
				Address:   addr,
			}
			if err := dbmaddr.AddIPAddresses(ctx, txn, true, ipaddr); err != nil {
				logEntry.WithError(err).WithField("addr", ipaddr.Address).Warnln("Could not insert ip address")
				return fmt.Errorf("add ip addresses: %w", err)
			}
		}
	}
	if _, err = dbmaddr.Update(ctx, txn, boil.Infer()); err != nil {
		logEntry.WithError(err).Warnln("Could not update multi address")
		return fmt.Errorf("update multi address: %w", err)
	}

	return txn.Commit()
}

func isRelayedMaddr(maddr ma.Multiaddr) bool {
	_, err := maddr.ValueForProtocol(ma.P_CIRCUIT)
	if err == nil {
		return true
	} else if errors.Is(err, ma.ErrProtocolNotFound) {
		return false
	} else {
		log.WithError(err).WithField("maddr", maddr).Warnln("Unexpected error while parsing multi address")
		return false
	}
}
