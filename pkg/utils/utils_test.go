package utils

import (
	"sort"
	"testing"

	nt "github.com/dennis-tra/nebula-crawler/pkg/nebtest"

	"github.com/stretchr/testify/require"

	ma "github.com/multiformats/go-multiaddr"
)

func TestMergeMaddrs(t *testing.T) {
	type args struct{}
	maddr1 := nt.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/3000")
	maddr2 := nt.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/3001")
	maddr3 := nt.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/3002")
	tests := []struct {
		name      string
		maddrSet1 []ma.Multiaddr
		maddrSet2 []ma.Multiaddr
		want      []ma.Multiaddr
	}{
		{
			name:      "empty sets",
			maddrSet1: []ma.Multiaddr{},
			maddrSet2: []ma.Multiaddr{},
			want:      []ma.Multiaddr{},
		},
		{
			name:      "only set 1",
			maddrSet1: []ma.Multiaddr{maddr1},
			maddrSet2: []ma.Multiaddr{},
			want:      []ma.Multiaddr{maddr1},
		},
		{
			name:      "only set 2",
			maddrSet1: []ma.Multiaddr{},
			maddrSet2: []ma.Multiaddr{maddr1},
			want:      []ma.Multiaddr{maddr1},
		},
		{
			name:      "single duplicate",
			maddrSet1: []ma.Multiaddr{maddr1},
			maddrSet2: []ma.Multiaddr{maddr1},
			want:      []ma.Multiaddr{maddr1},
		},
		{
			name:      "unique only",
			maddrSet1: []ma.Multiaddr{maddr2},
			maddrSet2: []ma.Multiaddr{maddr1},
			want:      []ma.Multiaddr{maddr1, maddr2},
		},
		{
			name:      "mixed",
			maddrSet1: []ma.Multiaddr{maddr1, maddr2},
			maddrSet2: []ma.Multiaddr{maddr1, maddr3},
			want:      []ma.Multiaddr{maddr1, maddr2, maddr3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := MergeMaddrs(tt.maddrSet1, tt.maddrSet2)

			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].String() < tt.want[j].String()
			})

			sort.Slice(merged, func(i, j int) bool {
				return merged[i].String() < merged[j].String()
			})

			require.Equal(t, tt.want, merged)
		})
	}
}
