package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/dennis-tra/nebula-crawler/config"
	v1 "github.com/dennis-tra/nebula-crawler/proto/nebula/v1"
	"github.com/dennis-tra/nebula-crawler/proto/nebula/v1/nebulav1connect"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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
	mux := http.NewServeMux()
	path, handler := nebulav1connect.NewNebulaServiceHandler(&nebulaServiceServer{})
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
type nebulaServiceServer struct{}

var _ nebulav1connect.NebulaServiceHandler = (*nebulaServiceServer)(nil)

func (n nebulaServiceServer) GetPeer(ctx context.Context, c *connect.Request[v1.GetPeerRequest]) (*connect.Response[v1.GetPeerResponse], error) {
	log.WithField("multihash", c.Msg.MultiHash).Info("GetPeer")
	resp := &connect.Response[v1.GetPeerResponse]{
		Msg: &v1.GetPeerResponse{
			MultiHash: c.Msg.MultiHash,
		},
	}

	return resp, nil
}
