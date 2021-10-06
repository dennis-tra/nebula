package provide

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

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

	exp Experiment
}

// NewScheduler initializes a new libp2p host and scheduler instance.
func NewScheduler(conf *config.Config, dbc *db.Client) (*Scheduler, error) {
	return &Scheduler{
		Service:    service.New("scheduler"),
		dbc:        dbc,
		config:     conf,
		eventsChan: make(chan Event),
		exp: Experiment{
			events:   []Event{},
			involved: map[peer.ID]struct{}{},
		},
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

	// Update experiment data
	s.exp.requesterID = r.h.ID()
	s.exp.providerID = p.h.ID()

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

	ctx, queryEvents := routing.RegisterForQueryEvents(s.Ctx())
	go s.handleQueryEvents(queryEvents)

	// Note start of experiment
	s.exp.startTime = time.Now()

	// Provide the random content from above.
	if err = p.Provide(ctx, content); err != nil {
		return errors.Wrap(err, "provide content")
	}

	// Give some time for the requesters to pick up the provider record.
	// If it times out stop them...
	select {
	case <-time.After(30 * time.Second):
		r.Shutdown()
	case <-r.SigDone():
	}

	var fevents []Event
	for _, event := range s.exp.events {
		if _, isInvolved := s.exp.involved[event.RemoteID()]; isInvolved {
			fevents = append(fevents, event)
		}
	}
	s.exp.events = fevents

	return nil
}

type Experiment struct {
	providerID  peer.ID
	requesterID peer.ID
	startTime   time.Time
	events      []Event
	involved    map[peer.ID]struct{}
}

func (s *Scheduler) handleQueryEvents(queryEvents <-chan *routing.QueryEvent) {
	for event := range queryEvents {
		// no need for locking as only this go routine accesses this map
		s.exp.involved[event.ID] = struct{}{}
	}
}

func (s *Scheduler) readEvents() {
	s.Service.ServiceStarted()
	defer s.Service.ServiceStopped()

	log.Info("Start reading events")
	defer log.Info("Stop reading events")

	for {
		select {
		case event := <-s.eventsChan:
			s.handleEvent(event)
		case <-s.SigShutdown():
			return
		}
	}
}

func (s *Scheduler) handleEvent(event Event) {
	log.Debugf("Read event %T", event)

	switch event.LocalID() {
	case s.exp.providerID:
		s.addEvent(event)
		switch evt := event.(type) {
		case *SendMessageStart:
			if evt.Message.Type == pb.Message_ADD_PROVIDER {
				log.WithField("remoteID", evt.RemoteID().Pretty()[:16]).Infoln("Adding provider record")
			}
		}
	case s.exp.requesterID:
		switch evt := event.(type) {
		case *SendRequestStart:
			if evt.Request.Type == pb.Message_GET_PROVIDERS {
				s.addEvent(event)
			}
		case *SendRequestEnd:
			if evt.Err == nil && evt.Response.Type == pb.Message_GET_PROVIDERS {
				s.addEvent(event)
			}
		}
	default:
		panic("unexpected")
	}
}

func (s *Scheduler) addEvent(event Event) {
	s.exp.events = append(s.exp.events, event)
}
