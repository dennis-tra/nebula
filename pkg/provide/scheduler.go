package provide

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
)

// The Scheduler handles the scheduling and managing of
//   a) providers - Takes the CID of randomly generated data and tries to insert a provider record in the DHT.
// 					During this process dial-events and more are collected.
//   b) requesters - The requester tries to find the closest peers to the CID of random data and periodically
//                   monitors them for associated provider records.
type Scheduler struct {
	// The configuration of timeouts etc.
	config *config.Config

	// A channel that receives events from the receiver and provider and is read by this scheduler
	eventsChan chan Event

	// Keeps track of the raw events that were dispatched on the eventsChan and also general experiment information.
	measurement Measurement

	// The measurement timestamps that will be used as a prefix for the file names.
	measurements []string
}

// NewScheduler initializes a new libp2p host and scheduler instance.
func NewScheduler(conf *config.Config) (*Scheduler, error) {
	return &Scheduler{
		config:       conf,
		eventsChan:   make(chan Event),
		measurements: []string{},
	}, nil
}

func (s *Scheduler) StartExperiment(ctx context.Context) error {
	r, err := NewRequester(ctx, s.config, s.eventsChan)
	if err != nil {
		return errors.Wrap(err, "new requester")
	}

	// Construct the provider libp2p host
	p, err := NewProvider(ctx, s.config, s.eventsChan)
	if err != nil {
		return errors.Wrap(err, "new provider")
	}

	// Start reading events
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		for {
			select {
			case <-s.eventsChan:
				// discard events
			case <-cancelCtx.Done():
				return
			}
		}
	}()

	// Bootstrap both libp2p hosts by connecting to the canonical bootstrap peers.
	// TODO: use configuration for list of bootstrap peers
	group, errCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return p.Bootstrap(errCtx)
	})
	group.Go(func() error {
		return r.Bootstrap(errCtx)
	})
	if err = group.Wait(); err != nil {
		return errors.Wrap(err, "bootstrap err group")
	}

	// Check if we should wait until the routing table of the provider was refreshed.
	if s.config.RefreshRoutingTable {
		p.RefreshRoutingTable()
	}

	// Stop discarding events
	cancel()

	for i := 0; i < s.config.ProvideRunCount; i++ {
		log.WithField("total", s.config.ProvideRunCount).Infof("Starting measurement run %d", i+1)

		cancelCtx, cancel := context.WithCancel(ctx)
		go s.readEvents(cancelCtx)

		// Initialize active measurement
		s.measurement = Measurement{
			providerID:  p.h.ID(),
			requesterID: r.h.ID(),
			monitored:   []peer.ID{},
			events:      []Event{},
			involved:    sync.Map{},
			refreshRT:   s.config.RefreshRoutingTable,
		}

		// Generate random content that we'll provide in the DHT.
		content, err := NewRandomContent()
		if err != nil {
			cancel()
			return errors.Wrap(err, "new random content")
		}
		log.Infof("Generated random content %s", content.cid.String())
		s.measurement.content = content

		// Start pinging the closest peers to the random content from above for provider records.
		s.measurement.monitored, err = r.MonitorProviders(cancelCtx, content) // mpeers = monitored peers
		if err != nil {
			cancel()
			return errors.Wrap(err, "monitor provider")
		}

		queryCtx, queryEvents := routing.RegisterForQueryEvents(cancelCtx)
		go s.handleQueryEvents(queryEvents)

		// Note start of experiment
		s.measurement.startTime = time.Now()

		// Provide the random content from above.
		if err = p.Provide(queryCtx, content); err != nil {
			cancel()
			return errors.Wrap(err, "provide content")
		}

		// Note end of experiment
		s.measurement.endTime = time.Now()

		// Give some time for the requesters to pick up the provider record.
		// If it times out stop them...
		select {
		case <-time.After(30 * time.Second):
		case <-cancelCtx.Done():
		}

		// Stop reading events
		cancel()

		if err = s.measurement.serialize(s.config.ProvideOutDir); err != nil {
			return errors.Wrap(err, "serialize measurement")
		}

		s.measurements = append(s.measurements, s.measurement.Prefix())
	}

	f, err := os.Create(path.Join(s.config.ProvideOutDir, "measurements.json"))
	if err != nil {
		return errors.Wrap(err, "creating measurements json")
	}
	defer f.Close()

	data, err := json.MarshalIndent(s.measurements, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshal measurements json")
	}

	if _, err = f.Write(data); err != nil {
		return errors.Wrap(err, "writing measurements json")
	}

	return nil
}

// handleQueryEvents is called in the scope of the Provide operation. So every query
// event that we'll receive on the queryEvents channel is _involved_ for the Provide
// process.
func (s *Scheduler) handleQueryEvents(queryEvents <-chan *routing.QueryEvent) {
	for event := range queryEvents {
		// no need for locking as only this go routine accesses this map
		s.measurement.involved.Store(event.ID, struct{}{})
	}
}

// readEvents reads the events channel until it's closed or the scheduler was asked to stop.
func (s *Scheduler) readEvents(ctx context.Context) {
	log.Info("Start reading events")
	defer log.Info("Stop reading events")

	for {
		select {
		case event := <-s.eventsChan:
			s.handleEvent(event)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scheduler) handleEvent(event Event) {
	log.Debugf("Read event %T", event)

	switch event.LocalID() {
	case s.measurement.providerID:
		s.addEvent(event)
		switch evt := event.(type) {
		case *SendMessageStart:
			if evt.Message.Type == pb.Message_ADD_PROVIDER {
				log.WithField("remoteID", utils.FmtPeerID(evt.RemoteID())).Infoln("Adding provider record")
			}
		}
	case s.measurement.requesterID:
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
	s.measurement.events = append(s.measurement.events, event)
}
