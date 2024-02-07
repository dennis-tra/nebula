package discv5

import (
	"errors"
	"math"
	"testing"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dennis-tra/nebula-crawler/nebtest"
)

func Test_sanitizeAddrs(t *testing.T) {
	tests := []struct {
		name    string
		maddrs  []ma.Multiaddr
		want    []ma.Multiaddr
		wantGen bool
	}{
		{
			name:    "empty",
			maddrs:  []ma.Multiaddr{},
			want:    []ma.Multiaddr{},
			wantGen: false,
		},
		{
			name: "tcp exists",
			maddrs: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/udp/1234"),
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/tcp/1234"),
			},
			want: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/tcp/1234"),
			},
			wantGen: false,
		},
		{
			name: "quic exist",
			maddrs: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/udp/1234/quic"),
			},
			want: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/udp/1234/quic"),
			},
			wantGen: false,
		},
		{
			name: "tcp and quic exist",
			maddrs: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/udp/1234"),
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/tcp/1234"),
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/udp/1234/quic"),
			},
			want: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/tcp/1234"),
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/udp/1234/quic"),
			},
			wantGen: false,
		},
		{
			name: "tcp and quicv1 exist",
			maddrs: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/udp/1234"),
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/tcp/1234"),
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/udp/1234/quic-v1"),
			},
			want: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/tcp/1234"),
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/udp/1234/quic-v1"),
			},
			wantGen: false,
		},
		{
			name: "single udp ip4",
			maddrs: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/udp/1234"),
			},
			want: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/tcp/1234"),
			},
			wantGen: true,
		},
		{
			name: "single udp ip6",
			maddrs: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip6/::1/udp/1234"),
			},
			want: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip6/::1/tcp/1234"),
			},
			wantGen: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addrs, generated := sanitizeAddrs(tt.maddrs)
			assert.Equalf(t, tt.want, addrs, "sanitizeAddrs(%v)", tt.maddrs)
			assert.Equal(t, tt.wantGen, generated)
		})
	}
}

func TestNoSuccessfulRequest(t *testing.T) {
	// fail if err is nil
	require.False(t, noSuccessfulRequest(nil, 0))
	require.False(t, noSuccessfulRequest(nil, 0b11111111))

	err := errors.New("error")

	// list of numbers that are power of two minus one
	// for which noSuccessfulRequest should return true
	// because all bits are set (all failures = no success)
	powerOfTwoMinusOneList := []uint32{
		0b00000000,
		0b00000001,
		0b00000011,
		0b00000111,
		0b00001111,
		0b00011111,
		0b00111111,
		0b01111111,
		0b11111111,
	}

	for i := uint32(0); i < uint32(math.Pow(2, 8)); i++ {
		powerOfTwoMinusOne := false
		for _, v := range powerOfTwoMinusOneList {
			if i == v {
				powerOfTwoMinusOne = true
				break
			}
		}
		// assert that noSuccessfulRequest returns true if and only if
		// all bits are set
		require.Equal(t, powerOfTwoMinusOne, noSuccessfulRequest(err, i))
	}
}
