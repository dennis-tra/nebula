package utils

import (
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

// MaddrsToAddrs maps a slice of multi addresses to their string representation.
func MaddrsToAddrs(maddrs []ma.Multiaddr) []string {
	addrs := make([]string, len(maddrs))
	for i, maddr := range maddrs {
		addrs[i] = maddr.String()
	}
	return addrs
}

// AddrsToMaddrs maps a slice of addresses to their multiaddress representation.
func AddrsToMaddrs(addrs []string) ([]ma.Multiaddr, error) {
	maddrs := make([]ma.Multiaddr, len(addrs))
	for i, addr := range addrs {
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		maddrs[i] = maddr
	}

	return maddrs, nil
}

// AddrInfoFilterPrivateMaddrs strips private multiaddrs from the given peer address information.
func AddrInfoFilterPrivateMaddrs(pi peer.AddrInfo) peer.AddrInfo {
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

// AddrInfoFilterPublicMaddrs strips public multiaddrs from the given peer address information.
func AddrInfoFilterPublicMaddrs(pi peer.AddrInfo) peer.AddrInfo {
	filtered := peer.AddrInfo{
		ID:    pi.ID,
		Addrs: []ma.Multiaddr{},
	}

	// Just keep public multi addresses
	for _, maddr := range pi.Addrs {
		if manet.IsPublicAddr(maddr) {
			continue
		}
		filtered.Addrs = append(filtered.Addrs, maddr)
	}

	return filtered
}

// FilterPrivateMaddrs strips private multiaddrs from the given peer address information.
func FilterPrivateMaddrs(maddrs []ma.Multiaddr) []ma.Multiaddr {
	var filtered []ma.Multiaddr
	for _, maddr := range maddrs {
		if manet.IsPrivateAddr(maddr) {
			continue
		}
		filtered = append(filtered, maddr)
	}
	return filtered
}

// FilterPublicMaddrs strips public multiaddrs from the given peer address information.
func FilterPublicMaddrs(maddrs []ma.Multiaddr) []ma.Multiaddr {
	var filtered []ma.Multiaddr
	for _, maddr := range maddrs {
		if manet.IsPublicAddr(maddr) {
			continue
		}
		filtered = append(filtered, maddr)
	}
	return filtered
}

// MergeMaddrs strips private multiaddrs from the given peer address information.
func MergeMaddrs(maddrSet1 []ma.Multiaddr, maddrSet2 []ma.Multiaddr) []ma.Multiaddr {
	maddrSetOut := make(map[string]ma.Multiaddr, len(maddrSet1))
	for _, maddr := range maddrSet1 {
		if _, found := maddrSetOut[string(maddr.Bytes())]; found {
			continue
		}
		maddrSetOut[string(maddr.Bytes())] = maddr
	}

	for _, maddr := range maddrSet2 {
		if _, found := maddrSetOut[string(maddr.Bytes())]; found {
			continue
		}
		maddrSetOut[string(maddr.Bytes())] = maddr
	}

	maddrsOut := make([]ma.Multiaddr, 0, len(maddrSetOut))
	for _, maddr := range maddrSetOut {
		maddrsOut = append(maddrsOut, maddr)
	}

	return maddrsOut
}

// IsResourceLimitExceeded returns true if the given error represents an error
// related to a limit of the local resource manager.
func IsResourceLimitExceeded(err error) bool {
	return err != nil && strings.HasSuffix(err.Error(), network.ErrResourceLimitExceeded.Error())
}
