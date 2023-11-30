package discv5

import (
	"testing"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"

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
