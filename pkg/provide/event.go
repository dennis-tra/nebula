package provide

import (
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	ma "github.com/multiformats/go-multiaddr"
)

// The Event interface must be implemented by all ec
// that need to be tracked.
type Event interface {
	LocalID() peer.ID
	RemoteID() peer.ID
	TimeStamp() time.Time
	Scope() string
	IsStart() bool
	Error() error
}

// BaseEvent captures the minimum number of fields and event
// needs to have. This is intended to be embedded into other
// structs.
type BaseEvent struct {
	Local  peer.ID
	Remote peer.ID
	Time   time.Time
}

func NewBaseEvent(local peer.ID, remote peer.ID) BaseEvent {
	return BaseEvent{
		Local:  local,
		Remote: remote,
		Time:   time.Now(),
	}
}

func (e *BaseEvent) LocalID() peer.ID {
	return e.Local
}

func (e *BaseEvent) RemoteID() peer.ID {
	return e.Remote
}

func (e *BaseEvent) TimeStamp() time.Time {
	return e.Time
}

func (e *BaseEvent) Error() error {
	return nil
}

// The DialStart event is dispatch when the TCP or Websocket
// trpt modules start dialing a certain peer under the
// given multi address.
type DialStart struct {
	BaseEvent
	Transport string
	Maddr     ma.Multiaddr
}

func (d *DialStart) Scope() string {
	return "dial"
}

func (d *DialStart) IsStart() bool {
	return true
}

// The DialEnd event is dispatch when the TCP or Websocket
// trpt modules have finished dialing a certain peer under
// the given multi address.
type DialEnd struct {
	BaseEvent
	Transport string
	Maddr     ma.Multiaddr
	Err       error
}

func (d *DialEnd) Scope() string {
	return "dial"
}

func (d *DialEnd) IsStart() bool {
	return false
}

func (d *DialEnd) Error() error {
	return d.Err
}

type SendRequestStart struct {
	BaseEvent
	Request      *pb.Message
	AgentVersion string
	Protocols    []string
}

func NewSendRequestStart(h host.Host, local peer.ID, remote peer.ID, pmes *pb.Message) *SendRequestStart {
	event := &SendRequestStart{
		BaseEvent: NewBaseEvent(local, remote),
		Request:   pmes,
		Protocols: []string{},
	}

	// Extract agent
	if agent, err := h.Peerstore().Get(remote, "AgentVersion"); err == nil {
		event.AgentVersion = agent.(string)
	}

	// Extract protocols
	if protocols, err := h.Peerstore().GetProtocols(remote); err == nil {
		event.Protocols = protocols
	}

	return event
}

func (s *SendRequestStart) Scope() string {
	return "send_request"
}

func (s *SendRequestStart) IsStart() bool {
	return true
}

type SendRequestEnd struct {
	BaseEvent
	Response     *pb.Message
	AgentVersion string
	Protocols    []string
	Err          error
}

func NewSendRequestEnd(h host.Host, local peer.ID, remote peer.ID) *SendRequestEnd {
	event := &SendRequestEnd{
		BaseEvent: NewBaseEvent(local, remote),
		Protocols: []string{},
	}

	// Extract agent
	if agent, err := h.Peerstore().Get(remote, "AgentVersion"); err == nil {
		event.AgentVersion = agent.(string)
	}

	// Extract protocols
	if protocols, err := h.Peerstore().GetProtocols(remote); err == nil {
		event.Protocols = protocols
	}

	return event
}

func (s *SendRequestEnd) Scope() string {
	return "send_request"
}

func (s *SendRequestEnd) IsStart() bool {
	return false
}

func (e *SendRequestEnd) Error() error {
	return e.Err
}

type SendMessageStart struct {
	BaseEvent
	Message      *pb.Message
	AgentVersion string
	Protocols    []string
}

func NewSendMessageStart(h host.Host, local peer.ID, remote peer.ID, pmes *pb.Message) *SendMessageStart {
	event := &SendMessageStart{
		BaseEvent: NewBaseEvent(local, remote),
		Message:   pmes,
		Protocols: []string{},
	}

	// Extract agent
	if agent, err := h.Peerstore().Get(remote, "AgentVersion"); err == nil {
		event.AgentVersion = agent.(string)
	}

	// Extract protocols
	if protocols, err := h.Peerstore().GetProtocols(remote); err == nil {
		event.Protocols = protocols
	}

	return event
}

func (s *SendMessageStart) Scope() string {
	return "send_message"
}

func (s *SendMessageStart) IsStart() bool {
	return true
}

type SendMessageEnd struct {
	BaseEvent
	Message      *pb.Message
	AgentVersion string
	Protocols    []string
	Err          error
}

func NewSendMessageEnd(h host.Host, local peer.ID, remote peer.ID, pmes *pb.Message) *SendMessageEnd {
	event := &SendMessageEnd{
		BaseEvent: NewBaseEvent(local, remote),
		Protocols: []string{},
	}

	// Extract agent
	if agent, err := h.Peerstore().Get(remote, "AgentVersion"); err == nil {
		event.AgentVersion = agent.(string)
	}

	// Extract protocols
	if protocols, err := h.Peerstore().GetProtocols(remote); err == nil {
		event.Protocols = protocols
	}

	return event
}

func (s *SendMessageEnd) Scope() string {
	return "send_message"
}

func (s *SendMessageEnd) IsStart() bool {
	return false
}

func (e *SendMessageEnd) Error() error {
	return e.Err
}
