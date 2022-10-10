package main

import (
	"context"
	"database/sql"
	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/maxmind"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	ma "github.com/multiformats/go-multiaddr"

	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// ResolveCommand contains the monitor sub-command configuration.
var ResolveCommand = &cli.Command{
	Name:   "resolve",
	Usage:  "Resolves all multi addresses to their IP addresses and geo location information",
	Action: ResolveAction,
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:        "batch-size",
			Usage:       "How many database entries should be fetched at each iteration",
			EnvVars:     []string{"NEBULA_RESOLVE_BATCH_SIZE"},
			DefaultText: "100",
			Value:       100,
		},
	},
}

// ResolveAction is the function that is called when running `nebula resolve`.
func ResolveAction(c *cli.Context) error {
	log.Infoln("Starting Nebula multi address resolver...")

	// Load configuration file
	conf, err := config.Init(c)
	if err != nil {
		return err
	}

	// Initialize the database client
	dbc, err := db.InitClient(c.Context, conf)
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

	limit := c.Int("batch-size")

	//// lastID tracks the highest ID of the set of fetched multi addresses. It could be that the last multi address
	//// in dbmaddrs below can't be resolved to a proper IP address. This means we would not add an entry in the
	//// multi_address_x_ip_address table. The next round we fetch all multi addresses where the id is larger then
	//// the maximum id in the multi_address_x_ip_address. Since we could not resolve and hence have not inserted
	//// the last multi address we fetch it again and try to resolve it again... basically and endless loop. So we
	//// keep track of the last ID and if they are equal we break out.
	//lastID := 0

	// Start the main loop
	for {
		log.Infoln("Fetching multi addresses...")
		dbmaddrs, err := dbc.FetchUnresolvedMultiAddresses(c.Context, limit)
		if err != nil {
			return errors.Wrap(err, "fetching multi addresses")
		}
		log.Infof("Fetched %d multi addresses", len(dbmaddrs))

	}

	log.Infof("Done")
	return nil
}

// Resolve save the resolved IP addresses + their countries in a transaction
func resolve(ctx context.Context, dbh *sql.DB, mmc *maxmind.Client, dbmaddrs models.MultiAddressSlice) error {
	txn, err := dbh.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin txn")
	}
	defer db.Rollback(txn)

	for _, dbmaddr := range dbmaddrs {
		log.Debugln("Resolving", dbmaddr.Maddr)
		maddr, err := ma.NewMultiaddr(dbmaddr.Maddr)
		if err != nil {
			log.WithError(err).Warnln("Error parsing multi address")
			continue
		}
		dbmaddr.IsPublic = null.BoolFrom(manet.IsPublicAddr(maddr))
		dbmaddr.IsRelay = null.BoolFrom(isRelayedMaddr(maddr))

		addrInfos, err := mmc.MaddrInfo(ctx, maddr)
		if err != nil {
			log.WithError(err).Warnln("Error deriving address information from maddr ", maddr)
		}

		if len(addrInfos) == 0 {
			dbmaddr.HasManyAddrs = null.BoolFrom(false)
			continue
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
		} else if len(addrInfos) > 1 { // not "else" because the MaddrInfo could have failed and we still want to update the maddr
			dbmaddr.HasManyAddrs = null.BoolFrom(true)
			// Due to dnsaddr protocols each multi address can point to multiple
			// IP addresses each in a different country.
			for addr, addrInfo := range addrInfos {
				// Save the IP address + country information + asn information
				ipaddr := &models.IPAddress{
					Asn:       null.NewInt(int(addrInfo.ASN), addrInfo.ASN != 0),
					IsCloud:   null.Int{},
					Country:   null.NewString(addrInfo.Country, addrInfo.Country != ""),
					Continent: null.NewString(addrInfo.Continent, addrInfo.Continent != ""),
					Address:   addr,
				}
				if err := dbmaddr.AddIPAddresses(ctx, txn, true, ipaddr); err != nil {
					log.WithError(err).WithField("addr", ipaddr.Address).Warnln("Could not insert ip address")
					continue
				}
			}
		}
		if _, err = dbmaddr.Update(ctx, txn, boil.Infer()); err != nil {
			log.WithError(err).WithField("maddr", dbmaddr.Maddr).Warnln("Could not update multi address")
			continue
		}
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
