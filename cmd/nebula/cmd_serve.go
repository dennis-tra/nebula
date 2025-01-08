package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/dennis-tra/nebula-crawler/config"
	v1 "github.com/dennis-tra/nebula-crawler/proto/nebula/v1"
	"github.com/dennis-tra/nebula-crawler/proto/nebula/v1/nebulav1connect"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/dennis-tra/nebula-crawler/db"
)

var serveConfig = &config.Serve{
	Root: rootConfig,
	Host: "localhost",
	Port: 8080,
}

// ServeCommand .
var ServeCommand = &cli.Command{
	Name:   "serve",
	Usage:  "Serves data from a Nebula database",
	Action: ServeAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "host",
			Usage:       "Let the server listen on the specified host",
			EnvVars:     []string{"NEBULA_SERVE_HOST"},
			Value:       serveConfig.Host,
			Destination: &serveConfig.Host,
		},
		&cli.IntFlag{
			Name:        "port",
			Usage:       "Let the server listen on the specified port",
			EnvVars:     []string{"NEBULA_SERVE_PORT"},
			Value:       serveConfig.Port,
			Destination: &serveConfig.Port,
		},
	},
}

// ServeAction is the function that is called when running `nebula resolve`.
func ServeAction(c *cli.Context) error {
	log.Infoln("Start serving Nebula data...")
	defer log.Infoln("Stopped serving Nebula data.")

	ctx := c.Context

	// initialize a new database client based on the given configuration.
	// Options are Postgres, JSON, and noop (dry-run).
	dbc, err := db.NewServerClient(ctx, rootConfig.Database)
	if err != nil {
		return fmt.Errorf("new database client: %w", err)
	}
	defer func() {
		if err := dbc.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) && !strings.Contains(err.Error(), "use of closed network connection") {
			log.WithError(err).Warnln("Failed closing database handle")
		}
	}()

	mux := http.NewServeMux()
	path, handler := nebulav1connect.NewNebulaServiceHandler(&nebulaServiceServer{
		dbc: dbc,
	})
	mux.Handle(path, handler)

	address := fmt.Sprintf("%s:%d", serveConfig.Host, serveConfig.Port)

	s := http.Server{
		Addr:    address,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		log.WithField("addr", address).Infoln("Start listening...")
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).WithField("addr", address).Error("Failed to serve gRPC server")
		}
	}()

	select {
	case <-done:
	case <-c.Context.Done():
	}

	shutdownTimeout := 30 * time.Second
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.WithField("timeout", shutdownTimeout).WithField("addr", address).Infoln("Shutting down...")
	if err := s.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("Failed to shutdown gRPC server")
	}

	return nil
}

// petStoreServiceServer implements the PetStoreService API.
type nebulaServiceServer struct {
	dbc db.ServerClient
}

var _ nebulav1connect.NebulaServiceHandler = (*nebulaServiceServer)(nil)

func (n *nebulaServiceServer) GetPeer(ctx context.Context, c *connect.Request[v1.GetPeerRequest]) (*connect.Response[v1.GetPeerResponse], error) {
	log.WithField("multihash", c.Msg.MultiHash).Info("GetPeer")

	dbPeer, dbProtocols, err := n.dbc.GetPeer(ctx, c.Msg.MultiHash)
	if err != nil {
		return nil, err
	}

	v1Maddrs := make([]*v1.MultiAddress, 0, len(dbPeer.R.MultiAddresses))
	for _, dbMaddr := range dbPeer.R.MultiAddresses {
		var asn *int32
		if !dbMaddr.Asn.IsZero() {
			val := int32(dbMaddr.Asn.Int)
			asn = &val
		}

		var isCloud *int32
		if !dbMaddr.IsCloud.IsZero() {
			val := int32(dbMaddr.IsCloud.Int)
			asn = &val
		}

		var country *string
		if !dbMaddr.Country.IsZero() {
			country = &dbMaddr.Country.String
		}

		var continent *string
		if !dbMaddr.Continent.IsZero() {
			continent = &dbMaddr.Country.String
		}

		var ip *string
		if !dbMaddr.Addr.IsZero() {
			ip = &dbMaddr.Addr.String
		}

		v1Maddrs = append(v1Maddrs, &v1.MultiAddress{
			MultiAddress: dbMaddr.Maddr,
			Asn:          asn,
			IsCloud:      isCloud,
			Country:      country,
			Continent:    continent,
			Ip:           ip,
		})
	}

	protocols := make([]string, 0, len(dbProtocols))
	for _, dbProtocol := range dbProtocols {
		protocols = append(protocols, dbProtocol.Protocol)
	}

	var av *string
	if dbPeer.R.AgentVersion != nil {
		av = &dbPeer.R.AgentVersion.AgentVersion
	}

	resp := &connect.Response[v1.GetPeerResponse]{
		Msg: &v1.GetPeerResponse{
			MultiHash:      dbPeer.MultiHash,
			AgentVersion:   av,
			MultiAddresses: v1Maddrs,
			Protocols:      protocols,
		},
	}

	return resp, nil
}
