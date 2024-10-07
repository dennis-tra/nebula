package mainline

import (
	"net"

	v2 "github.com/anacrolix/dht/v2"
	"github.com/anacrolix/dht/v2/krpc"
)

type addr struct {
	krpc.NodeAddr
}

var _ v2.Addr = (*addr)(nil)

func (a addr) Raw() net.Addr {
	return a.NodeAddr.UDP()
}

func (a addr) Port() int {
	return a.NodeAddr.Port
}

func (a addr) IP() net.IP {
	return a.NodeAddr.IP
}

func (a addr) String() string {
	return a.NodeAddr.String()
}

func (a addr) KRPC() krpc.NodeAddr {
	return a.NodeAddr
}
