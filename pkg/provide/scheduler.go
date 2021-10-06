package provide

import (
	"golang.org/x/sync/errgroup"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
)

// The Scheduler handles the scheduling and managing of
//   a) providers - TODO
//   b) requesters - TODO
type Scheduler struct {
	// Service represents an entity that runs in a separate go routine and where its lifecycle
	// needs to be handled externally. This is true for this scheduler, so we're embedding it here.
	*service.Service

	// The database client
	dbc *db.Client

	// The configuration of timeouts etc.
	config *config.Config

	//
	eventsChan chan Event
}

// NewScheduler initializes a new libp2p host and scheduler instance.
func NewScheduler(conf *config.Config, dbc *db.Client) (*Scheduler, error) {
	return &Scheduler{
		Service:    service.New("scheduler"),
		dbc:        dbc,
		config:     conf,
		eventsChan: make(chan Event),
	}, nil
}

func (s *Scheduler) StartExperiment() error {
	// Generate random content that we'll provide in the DHT.
	content, err := NewRandomContent()
	if err != nil {
		return errors.Wrap(err, "new random content")
	}
	log.Infof("Generated random content %s", content.cid.String())

	// Start reading events
	go s.readEvents()

	// Construct the requester (must come before new provider due to monkey patching)
	r, err := NewRequester(s.Ctx(), s.config, s.eventsChan)
	if err != nil {
		return errors.Wrap(err, "new requester")
	}

	// Construct the provider libp2p host
	p, err := NewProvider(s.Ctx(), s.config, s.eventsChan)
	if err != nil {
		return errors.Wrap(err, "new provider")
	}

	// Bootstrap both libp2p hosts by connecting to the canonical bootstrap peers.
	// TODO: use configuration for list of bootstrap peers
	group, ctx := errgroup.WithContext(s.Ctx())
	group.Go(func() error {
		return p.Bootstrap(ctx)
	})
	group.Go(func() error {
		return r.Bootstrap(ctx)
	})
	if err = group.Wait(); err != nil {
		return errors.Wrap(err, "bootstrap err group")
	}

	// Check if we should wait until the routing table of the provider was refreshed.
	if s.config.RefreshRoutingTable {
		p.RefreshRoutingTable()
	}

	// Start pinging the closest peers to the random content from above for provider records.
	if err = r.MonitorProviders(content); err != nil {
		return errors.Wrap(err, "monitor provider")
	}

	// Provide the random content from above.
	if err = p.Provide(ctx, content); err != nil {
		return errors.Wrap(err, "provide")
	}

	r.Shutdown()

	return nil
}

func (s *Scheduler) readEvents() {
	s.Service.ServiceStarted()
	defer s.Service.ServiceStopped()

	log.Info("Start reading events")
	defer log.Info("Stop reading events")

	for {
		var event Event
		select {
		case event = <-s.eventsChan:
		case <-s.SigShutdown():
			return
		}
		log.Debugf("Read event %T", event)
	}
}
