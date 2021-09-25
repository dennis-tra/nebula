package maxmind

import (
	"context"
	"fmt"
	"testing"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_AddrCountry(t *testing.T) {
	client, err := NewClient()
	require.NoError(t, err)

	tests := []struct {
		addr    string
		want    string
		wantErr bool
	}{
		{addr: "invalid", want: "", wantErr: true},
		{addr: "159.69.43.228", want: "DE", wantErr: false},
		{addr: "100.0.0.2", want: "US", wantErr: false},
		{addr: "111.250.198.94", want: "TW", wantErr: false},
		{addr: "130.188.225.47", want: "FI", wantErr: false},
		{addr: "46.17.96.99", want: "NL", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s | iso: %s | err: %v", tt.addr, tt.want, tt.wantErr), func(t *testing.T) {
			got, err := client.AddrCountry(tt.addr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClient_MaddrCountry(t *testing.T) {
	client, err := NewClient()
	require.NoError(t, err)

	tests := []struct {
		addr    string
		want    string
		wantErr bool
	}{
		{addr: "/ip4/46.17.96.99/tcp/6666/p2p/Qme8g49gm3q4Acp7xWBKg3nAa9fxZ1YmyDJdyGgoG6LsXh/p2p-circuit", want: "NL", wantErr: false},
		{addr: "/p2p-circuit/p2p/QmPG5bax9kfpQUVDrzfahmh44Ab6egDeZ2QDWeTY279HLJ", want: "", wantErr: true},
		{addr: "/dnsaddr/bootstrap.libp2p.io", want: "", wantErr: true},
		//{addr: "/dns4/k8s-dev-ipfsp2pt-c0b76d02d7-969229bd37f82282.elb.ca-central-1.amazonaws.com/tcp/4001", want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s | iso: %s | err: %v", tt.addr, tt.want, tt.wantErr), func(t *testing.T) {
			maddr, err := ma.NewMultiaddr(tt.addr)
			require.NoError(t, err)

			got, err := client.MaddrCountry(context.Background(), maddr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
