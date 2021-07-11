package crawl

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func ExampleAgentVersionParsing() {
	fmt.Printf("%#v\n", agentVersionRegex.FindStringSubmatch(`go-ipfs/0.9.0/ce693d`))
	fmt.Printf("%#v\n", agentVersionRegex.SubexpNames())
	fmt.Printf("%#v\n", agentVersionRegex.FindStringSubmatch(`/go-ipfs/0.5.0-dev/ce693d`))
	fmt.Printf("%#v\n", agentVersionRegex.SubexpNames())
	fmt.Printf("%#v\n", agentVersionRegex.FindStringSubmatch(`no-match`))
	fmt.Printf("%#v\n", agentVersionRegex.SubexpNames())
	// Output:
	// []string{"go-ipfs/0.9.0/ce693d", "0.9.0", "", "ce693d"}
	// []string{"", "core", "prerelease", "commit"}
	// []string{"/go-ipfs/0.5.0-dev/ce693d", "0.5.0", "dev", "ce693d"}
	// []string{"", "core", "prerelease", "commit"}
	// []string(nil)
	// []string{"", "core", "prerelease", "commit"}
}

func Test_isNewSession(t *testing.T) {
	tests := []struct {
		name     string
		oldMaddr []string
		newMaddr []string
		want     bool
	}{
		{
			name:     "ip4/tcp: no entry",
			oldMaddr: []string{},
			newMaddr: []string{},
			want:     false,
		},
		{
			name:     "ip4/tcp: same maddrs one entry",
			oldMaddr: []string{"/ip4/1.1.1.1/tcp/1111"},
			newMaddr: []string{"/ip4/1.1.1.1/tcp/1111"},
			want:     false,
		},
		{
			name:     "ip4/tcp: same maddrs + additional entry",
			oldMaddr: []string{"/ip4/1.1.1.1/tcp/1111"},
			newMaddr: []string{"/ip4/1.1.1.1/tcp/1111", "/ip4/1.1.1.1/udp/1112"},
			want:     false,
		},
		{
			name:     "ip4/tcp: same host different port",
			oldMaddr: []string{"/ip4/1.1.1.1/tcp/1111"},
			newMaddr: []string{"/ip4/1.1.1.1/tcp/1112"},
			want:     true,
		},
		{
			name:     "ip4/tcp: same port different host",
			oldMaddr: []string{"/ip4/1.1.1.1/tcp/1111"},
			newMaddr: []string{"/ip4/1.1.1.2/tcp/1111"},
			want:     true,
		},
		{
			name:     "ip4/tcp: same port different host",
			oldMaddr: []string{"/ip4/1.1.1.1/udp/4001/quic", "/ip4/1.1.1.1/tcp/4001"},
			newMaddr: []string{"/ip4/1.1.1.1/udp/4001/quic", "/ip4/1.1.1.1/tcp/4003"},
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldMaddr, err := addrsToMaddrs(tt.oldMaddr)
			require.NoError(t, err)
			newMaddr, err := addrsToMaddrs(tt.newMaddr)
			require.NoError(t, err)

			if got := isNewSession(oldMaddr, newMaddr); got != tt.want {
				t.Errorf("isNewSession() = %v, want %v", got, tt.want)
			}
		})
	}
}
