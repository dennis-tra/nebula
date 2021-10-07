package provide

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	log "github.com/sirupsen/logrus"
)

// Measurement keeps track of the starting conditions and results of an experiment.
type Measurement struct {
	// The libp2p host peer identity of the provider
	providerID peer.ID

	// The libp2p host peer identity of the requester
	requesterID peer.ID

	// The random data that was provided
	content *Content

	// When did the provider start searching for peers for the provider record
	startTime time.Time

	// When did the provider end searching for peers for the provider record
	endTime time.Time

	// All events that occurred during the whole process
	events []Event

	// Keeps track of peers that were involved in the provide process. Since
	// events are dispatched for all dials regardless of whether they were
	// necessary for the provide process or e.g., the routing table refresh.
	// Therefore, the provider registers for query events that happened during the
	// provide process and keep track of all involved peers. This list is
	// ultimately used to remove all events from the events field that targeted
	// peers not relevant for the provide process.
	involved sync.Map

	// monitored represents the list of peers that the requester periodically
	// asked for provider records.
	monitored []peer.ID
}

// filterEvents removes all events that are related to remote peers that were not involved in the Provide process.
func (m *Measurement) filterEvents() {
	// Also mark the monitored peers as involved.
	// This is necessary if the requester found peers
	// that the provider didn't add provider records to.
	for _, mon := range m.monitored {
		m.involved.Store(mon, struct{}{})
	}

	var filtered []Event
	for _, event := range m.events {
		if _, isInvolved := m.involved.Load(event.RemoteID()); isInvolved {
			filtered = append(filtered, event)
		}
	}
	m.events = filtered
}

// checkIntegrity makes sure that for every start event there is a corresponding end event.
func (m *Measurement) checkIntegrity() bool {
	states := map[peer.ID]map[string]int{}
	for _, event := range m.events {

		if _, found := states[event.RemoteID()]; !found {
			states[event.RemoteID()] = map[string]int{}
		}

		if event.IsStart() {
			states[event.RemoteID()][event.Scope()] += 1
		} else {
			states[event.RemoteID()][event.Scope()] -= 1
		}
	}

	for p, scopes := range states {
		for scope, count := range scopes {
			if count != 0 {
				log.Warnln(p, scope)
				return false
			}
		}
	}

	return true
}
