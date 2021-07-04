package service

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
)

// State represents the lifecycle states of a service.
type State uint8

const (
	// These are the concrete lifecycle manifestations.
	Idle State = iota
	Started
	Stopping
	Stopped
)

// ErrServiceAlreadyStarted is returned if there are multiple calls to ServiceStarted.
// If this happens somethings wrong :/
var ErrServiceAlreadyStarted = errors.New("the service was already started in the past")

// Service represents an entity that runs in a
// separate go routine and where its lifecycle
// needs to be handled externally.
type Service struct {
	// The id of the service for logging purposes
	id string

	// A context that can be used for long running
	// io operations of the service. This context
	// gets cancelled when the service receives a
	// shutdown signal. It's controversial to store
	// a context in a struct field but I believe
	// that it makes sense here.
	ctx    context.Context
	cancel context.CancelFunc

	// The current state of this service.
	lk    sync.RWMutex
	state State

	// Time measurements
	StartTime    time.Time
	ShutdownTime time.Time
	DoneTime     time.Time

	// When a message is sent to this channel it
	// starts to gracefully shut down.
	shutdown chan struct{}

	// When a message is sent to this channel
	// the service has shut down.
	done chan struct{}
}

// New instantiates an initialised Service struct. It
// deliberately does not accept a context as an input
// parameter as I consider long running service life-
// cycle handling with contexts as a bad practice.
// Contexts belong in request/response paths and
// Services should be handled via channels.
func New(id string) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		ctx:      ctx,
		id:       id,
		cancel:   cancel,
		state:    Idle,
		shutdown: make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// ServiceStarted marks this service as started.
func (s *Service) ServiceStarted() {
	log.WithField("serviceId", s.id).Traceln("Service has started")

	s.StartTime = time.Now()

	s.lk.Lock()
	defer s.lk.Unlock()

	if s.state != Idle {
		panic(ErrServiceAlreadyStarted)
	}
	s.state = Started

	go func() {
		select {
		case <-s.shutdown:
		case <-s.done:
		}
		s.cancel()
	}()
}

// SigShutdown exposes the shutdown channel to listen for
// shutdown instructions.
func (s *Service) SigShutdown() chan struct{} {
	return s.shutdown
}

// Identifier returns the service identifier.
func (s *Service) Identifier() string {
	return s.id
}

// SigDone exposes the done channel to listen for
// service termination.
func (s *Service) SigDone() chan struct{} {
	return s.done
}

func (s *Service) IsStopping() bool {
	s.lk.Lock()
	defer s.lk.Unlock()
	return s.state == Stopping
}

func (s *Service) IsStarted() bool {
	s.lk.Lock()
	defer s.lk.Unlock()
	return s.state == Started
}

// ServiceStopped marks this service as stopped and
// ultimately releases an external call to Shutdown.
func (s *Service) ServiceStopped() {
	s.lk.Lock()
	defer s.lk.Unlock()

	if s.state == Idle || s.state == Stopped {
		return
	}
	s.state = Stopped

	close(s.done)
	s.DoneTime = time.Now()
	log.WithField("serviceId", s.id).Traceln("Service has stopped")
}

// ServiceContext returns the context associated with this
// service. This context is passed into requests or similar
// that are initiated from this service. Doing it this way
// we can cancel this contexts when someone shuts down
// the service, which results in all requests being stopped.
func (s *Service) ServiceContext() context.Context {
	return s.ctx
}

// Shutdown instructs the service to gracefully shut down.
// This function blocks until the done channel was closed
// which happens when ServiceStopped is called.
func (s *Service) Shutdown() {
	s.lk.Lock()
	if s.state != Started {
		s.lk.Unlock()
		return
	}

	log.WithField("serviceId", s.id).Traceln("Service shutting down...")
	s.state = Stopping
	s.lk.Unlock()

	close(s.shutdown)
	s.ShutdownTime = time.Now()
	<-s.done
	log.WithField("serviceId", s.id).Traceln("Service was shut down")
}
