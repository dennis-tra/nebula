package discv5

import (
	"testing"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"

	"github.com/dennis-tra/nebula-crawler/nebtest"
)

func Test_ensureTCPAddr(t *testing.T) {
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
				nebtest.MustMultiaddr(t, "/ip4/123.1.1.1/udp/1234"),
				nebtest.MustMultiaddr(t, "/ip4/123.1.1.1/tcp/1234"),
			},
			want: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.1.1.1/udp/1234"),
				nebtest.MustMultiaddr(t, "/ip4/123.1.1.1/tcp/1234"),
			},
		},
		{
			name: "single udp ip4",
			maddrs: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.1.1.1/udp/1234"),
			},
			want: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip4/123.1.1.1/udp/1234"),
				nebtest.MustMultiaddr(t, "/ip4/123.1.1.1/tcp/1234"),
			},
		},
		{
			name: "single udp ip6",
			maddrs: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip6/::1/udp/1234"),
			},
			want: []ma.Multiaddr{
				nebtest.MustMultiaddr(t, "/ip6/::1/udp/1234"),
				nebtest.MustMultiaddr(t, "/ip6/::1/tcp/1234"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ensureTCPAddr(tt.maddrs), "ensureTCPAddr(%v)", tt.maddrs)
		})
	}
}
