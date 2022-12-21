package utils

import (
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

// IDLength is here as a variable so that it can be decreased for tests with mocknet where IDs are way shorter.
// The call to FmtPeerID would panic if this value stayed at 16.
var IDLength = 16

func FmtPeerID(id peer.ID) string {
	return id.String()[:IDLength]
}

// MaddrsToAddrs maps a slice of multi addresses to their string representation.
func MaddrsToAddrs(maddrs []ma.Multiaddr) []string {
	addrs := make([]string, len(maddrs))
	for i, maddr := range maddrs {
		addrs[i] = maddr.String()
	}
	return addrs
}

// FilterPrivateMaddrs strips private multiaddrs from the given peer address information.
func FilterPrivateMaddrs(pi peer.AddrInfo) peer.AddrInfo {
	filtered := peer.AddrInfo{
		ID:    pi.ID,
		Addrs: []ma.Multiaddr{},
	}

	// Just keep public multi addresses
	for _, maddr := range pi.Addrs {
		if manet.IsPrivateAddr(maddr) {
			continue
		}
		filtered.Addrs = append(filtered.Addrs, maddr)
	}

	return filtered
}

func IsResourceLimitExceeded(err error) bool {
	return err != nil && strings.HasSuffix(err.Error(), network.ErrResourceLimitExceeded.Error())
}
