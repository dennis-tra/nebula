package main

import (
	"time"

	"github.com/friendsofgo/errors"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/maxmind"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
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
		&cli.BoolFlag{
			Name:        "unresolved",
			Usage:       "Whether to only resolve the yet unresolved multi addresses",
			EnvVars:     []string{"NEBULA_RESOLVE_UNRESOLVED"},
			DefaultText: "false",
			Value:       false,
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
	dbc, err := db.InitClient(conf)
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

	offset := 0
	limit := c.Int("batch-size")

	// Track the beginning of the resolution and put it to every association
	// row in the multi_addresses_x_ip_addresses table. This allows tracking
	// shifts in ip addresses behind e.g., dnsaddr multi addresses.
	resolutionTime := time.Now()

	// Start the main loop
	for {
		log.Infoln("Fetching multi addresses...")

		var err error
		var dbmaddrs models.MultiAddressSlice
		if c.Bool("unresolved") {
			dbmaddrs, err = dbc.FetchUnresolvedMultiAddresses(c.Context, offset, limit)
		} else {
			dbmaddrs, err = dbc.FetchMultiAddresses(c.Context, offset, limit)
		}
		if err != nil {
			return errors.Wrap(err, "fetching multi addresses")
		}
		offset += limit
		log.Infof("Fetched %d multi addresses", len(dbmaddrs))

		// No new multi addresses to resolve
		if len(dbmaddrs) == 0 {
			break
		}

		// Save the resolved IP addresses + their countries in a transaction
		txn, err := dbh.BeginTx(c.Context, nil)
		if err != nil {
			return errors.Wrap(err, "begin txn")
		}

		for _, dbmaddr := range dbmaddrs {
			log.Infoln("Resolving", dbmaddr.Maddr)
			maddr, err := ma.NewMultiaddr(dbmaddr.Maddr)
			if err != nil {
				log.WithError(err).Warnln("Error parsing multi address")
				continue
			}

			countries, err := mmc.MaddrCountry(c.Context, maddr)
			if err != nil {
				log.WithError(err).Warnln("Error deriving countries from maddr ", maddr)
				continue
			}

			// Due to dnsaddr protocols each multi address can point to multiple
			// IP addresses each in a different country.
			for addr, country := range countries {

				// Save the IP address + country information
				ipaddr := &models.IPAddress{
					Address: addr,
					Country: country,
				}
				if err := ipaddr.Upsert(c.Context, dbh, true,
					[]string{models.IPAddressColumns.Address, models.IPAddressColumns.Country},
					boil.Whitelist(models.IPAddressColumns.UpdatedAt), // need to update at least one field to retrieve the ID
					boil.Infer(),
				); err != nil {
					log.WithError(err).Warnln("Could not upsert ip address %s", ipaddr)
					continue
				}

				// Save the association of this IP address to the source multi address
				association := &models.MultiAddressesXIPAddress{
					MultiAddressID: dbmaddr.ID,
					IPAddressID:    ipaddr.ID,
					ResolvedAt:     resolutionTime,
				}
				if err := association.Upsert(c.Context, dbh, false, []string{
					models.MultiAddressesXIPAddressColumns.MultiAddressID,
					models.MultiAddressesXIPAddressColumns.ResolvedAt,
					models.MultiAddressesXIPAddressColumns.IPAddressID,
				}, boil.None(), boil.Infer()); err != nil {
					log.WithError(err).Warnln("Could not add ip address %s to db multi address", ipaddr)
					continue
				}
			}
		}

		if err = txn.Commit(); err != nil {
			if err2 := txn.Rollback(); err2 != nil {
				log.WithError(err).Warnln("Could not roll back txn")
			}
			return errors.Wrap(err, "committing txn")
		}
	}

	log.Infof("Done")
	return nil
}
