package utils

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
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

// MergeMaddrs takes two slices of multi addresses and merges them into a single
// one.
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

func ToEnode(peerID peer.ID, maddrs []ma.Multiaddr) (*enode.Node, error) {
	// init ENR and fill with secp256k1 public key from peerID
	var r enr.Record
	pubKey, err := peerID.ExtractPublicKey()
	if err != nil {
		return nil, fmt.Errorf("extract public key: %w", err)
	}

	raw, err := pubKey.Raw()
	if err != nil {
		return nil, fmt.Errorf("extract raw bytes from public key: %w", err)
	}

	x, y := secp256k1.DecompressPubkey(raw)
	r.Set(enode.Secp256k1(ecdsa.PublicKey{Curve: secp256k1.S256(), X: x, Y: y}))

	// Parse multi addresses from database
	addrInfo := peer.AddrInfo{ID: peerID}
	for _, maddr := range maddrs {
		ip4 := false
		if comp, err := maddr.ValueForProtocol(ma.P_IP4); err == nil {
			ip4 = true
			if ip := net.ParseIP(comp); ip != nil {
				r.Set(enr.IPv4(ip))
			}
		} else if comp, err := maddr.ValueForProtocol(ma.P_IP6); err == nil {
			if ip := net.ParseIP(comp); ip != nil {
				r.Set(enr.IPv6(ip))
			}
		}

		if comp, err := maddr.ValueForProtocol(ma.P_UDP); err == nil {
			if udp, err := strconv.Atoi(comp); err == nil && udp <= math.MaxUint16 {
				if ip4 {
					r.Set(enr.UDP(uint16(udp)))
				} else {
					r.Set(enr.UDP6(uint16(udp)))
				}
			}
		} else if comp, err := maddr.ValueForProtocol(ma.P_TCP); err == nil {
			if tcp, err := strconv.Atoi(comp); err == nil && tcp <= math.MaxUint16 {
				if ip4 {
					r.Set(enr.TCP(uint16(tcp)))
				} else {
					r.Set(enr.TCP6(uint16(tcp)))
				}
			}
		}

		addrInfo.Addrs = append(addrInfo.Addrs, maddr)
	}

	// use custom identity scheme to not check the signature.
	return enode.New(nebulaIdentityScheme{}, &r)
}

// nebulaIdentityScheme is an always-valid ID scheme. When a new [enode.Node] is
// constructed, the Verify method won't check the signature, and we just assume
// the record is valid. However, the NodeAddr method returns the correct node
// identifier.
type nebulaIdentityScheme struct{}

// Verify doesn't check the signature or anything. It assumes all records to be
// valid.
func (nebulaIdentityScheme) Verify(r *enr.Record, sig []byte) error {
	return nil
}

// NodeAddr returns the node's ID. The logic is copied from the [enode.V4ID]
// implementation.
func (nebulaIdentityScheme) NodeAddr(r *enr.Record) []byte {
	var pubkey enode.Secp256k1
	err := r.Load(&pubkey)
	if err != nil {
		return nil
	}
	buf := make([]byte, 64)
	math.ReadBits(pubkey.X, buf[:32])
	math.ReadBits(pubkey.Y, buf[32:])
	return crypto.Keccak256(buf)
}
