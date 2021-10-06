package provide

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/transport"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	"github.com/libp2p/go-tcp-transport"
	websocket "github.com/libp2p/go-ws-transport"
	ma "github.com/multiformats/go-multiaddr"
)

// TCPTransport is a thin wrapper around the actual *tcp.TcpTransport implementation.
// It intercepts calls to Dial to track when which peer is dialed.
type TCPTransport struct {
	local peer.ID
	ec    chan<- Event
	trpt  *tcp.TcpTransport
}

func NewTCPTransport(local peer.ID, ec chan<- Event) func(*tptu.Upgrader) *TCPTransport {
	return func(u *tptu.Upgrader) *TCPTransport {
		return &TCPTransport{
			local: local,
			ec:    ec,
			trpt:  tcp.NewTCPTransport(u),
		}
	}
}

func (t *TCPTransport) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (transport.CapableConn, error) {
	t.ec <- &DialStart{
		BaseEvent: NewBaseEvent(t.local, p),
		Transport: "tcp",
		Maddr:     raddr,
	}
	dial, err := t.trpt.Dial(ctx, raddr, p)
	t.ec <- &DialEnd{
		BaseEvent: NewBaseEvent(t.local, p),
		Transport: "tcp",
		Maddr:     raddr,
		Err:       err,
	}
	return dial, err
}

func (t *TCPTransport) CanDial(addr ma.Multiaddr) bool {
	return t.trpt.CanDial(addr)
}

func (t *TCPTransport) Listen(laddr ma.Multiaddr) (transport.Listener, error) {
	return t.trpt.Listen(laddr)
}

func (t *TCPTransport) Protocols() []int {
	return t.trpt.Protocols()
}

func (t *TCPTransport) Proxy() bool {
	return t.trpt.Proxy()
}

// WSTransport is a thin wrapper around the actual *websocket.WebsocketTransport
// implementation. It intercepts calls to Dial to track when which peer is dialed.
type WSTransport struct {
	local peer.ID
	ec    chan<- Event
	trpt  *websocket.WebsocketTransport
}

func NewWSTransport(local peer.ID, ec chan<- Event) func(u *tptu.Upgrader) *WSTransport {
	return func(u *tptu.Upgrader) *WSTransport {
		return &WSTransport{
			local: local,
			ec:    ec,
			trpt:  websocket.New(u),
		}
	}
}

func (ws *WSTransport) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (transport.CapableConn, error) {
	ws.ec <- &DialStart{
		BaseEvent: NewBaseEvent(ws.local, p),
		Transport: "ws",
		Maddr:     raddr,
	}
	dial, err := ws.trpt.Dial(ctx, raddr, p)
	ws.ec <- &DialEnd{
		BaseEvent: NewBaseEvent(ws.local, p),
		Transport: "ws",
		Maddr:     raddr,
		Err:       err,
	}
	return dial, err
}

func (ws *WSTransport) CanDial(addr ma.Multiaddr) bool {
	return ws.trpt.CanDial(addr)
}

func (ws *WSTransport) Listen(laddr ma.Multiaddr) (transport.Listener, error) {
	return ws.trpt.Listen(laddr)
}

func (ws *WSTransport) Protocols() []int {
	return ws.trpt.Protocols()
}

func (ws *WSTransport) Proxy() bool {
	return ws.trpt.Proxy()
}
