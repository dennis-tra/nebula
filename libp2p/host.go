package libp2p

import (
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
)

// Host wraps a regular libp2p host and exposes additional methods to
// register peers to receive Identify exchange events.
// TODO: I'm not quite happy with all the locking going on here...
type Host struct {
	host.Host

	// the event bus subscription for Identify events
	sub     event.Subscription
	subDone chan struct{}

	// a map of peer registrations ot the respective channel on which the events
	// should be emitted.
	regsMu sync.Mutex
	regs   map[peer.ID]chan event.EvtPeerIdentificationCompleted
}

// WrapHost wraps the given host to allow for registering peers for Identify
// events.
func WrapHost(h host.Host) (*Host, error) {
	sub, err := h.EventBus().Subscribe([]interface{}{new(event.EvtPeerIdentificationCompleted), new(event.EvtPeerIdentificationFailed)})
	if err != nil {
		return nil, fmt.Errorf("establish event subscription: %w", err)
	}

	wrapped := &Host{
		Host:    h,
		sub:     sub,
		subDone: make(chan struct{}),
		regs:    make(map[peer.ID]chan event.EvtPeerIdentificationCompleted),
	}

	go wrapped.consumeEvents()

	return wrapped, nil
}

// RegisterIdentify registers the given peer to receive the corresponding Identify
// events. The returned channel will emit the
// [event.EvtPeerIdentificationCompleted] event or be closed without an event in
// case of an error. The returned channel will emit at most one event.
func (h *Host) RegisterIdentify(pid peer.ID) <-chan event.EvtPeerIdentificationCompleted {
	h.regsMu.Lock()
	defer h.regsMu.Unlock()

	if resultChan, ok := h.regs[pid]; ok {
		return resultChan
	}

	h.regs[pid] = make(chan event.EvtPeerIdentificationCompleted, 1)

	return h.regs[pid]
}

// DeregisterIdentify deregisters the given peer and closes the channel which
// was previously returned in [RegisterIdentify]
func (h *Host) DeregisterIdentify(pid peer.ID) {
	h.regsMu.Lock()
	defer h.regsMu.Unlock()

	if resultChan, ok := h.regs[pid]; ok {
		delete(h.regs, pid)
		close(resultChan)
	}
}

func (h *Host) consumeEvents() {
	defer close(h.subDone)
	for e := range h.sub.Out() {

		h.regsMu.Lock()

		switch evt := e.(type) {
		case event.EvtPeerIdentificationCompleted:
			resultChan, found := h.regs[evt.Peer]
			if !found {
				break
			}

			resultChan <- evt
			delete(h.regs, evt.Peer)
			close(resultChan)
		case event.EvtPeerIdentificationFailed:
			resultChan, found := h.regs[evt.Peer]
			if !found {
				break
			}

			delete(h.regs, evt.Peer)
			close(resultChan)
		default:
			panic(fmt.Sprintf("received unexpected event"))
		}

		h.regsMu.Unlock()
	}
}

func (h *Host) Close() error {
	if err := h.sub.Close(); err != nil {
		log.WithError(err).Warnln("Failed closing event subscription")
	} else {
		select {
		case <-h.subDone:
		case <-time.After(time.Second):
			log.Warnln("Event subscription did not close within 1s")
		}
	}

	return h.Host.Close()
}
