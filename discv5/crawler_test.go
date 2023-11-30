package discv5

import (
	"testing"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"

	"github.com/dennis-tra/nebula-crawler/nebtest"
)

func Test_sanitizeAddrs(t *testing.T) {
	tests := []struct {
		name   string
		maddrs []ma.Multiaddr
		want   []ma.Multiaddr
	}{
		{
			name:   "empty",
			maddrs: []ma.Multiaddr{},
			want:   []ma.Multiaddr{},
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
		},
		{
			name: "quic exist",
			maddrs: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/udp/1234/quic"),
			},
			want: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/udp/1234/quic"),
			},
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
		},
		{
			name: "single udp ip4",
			maddrs: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/udp/1234"),
			},
			want: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.4.5.6/tcp/1234"),
			},
		},
		{
			name: "single udp ip6",
			maddrs: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip6/::1/udp/1234"),
			},
			want: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip6/::1/tcp/1234"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, sanitizeAddrs(tt.maddrs), "sanitizeAddrs(%v)", tt.maddrs)
		})
	}
}
