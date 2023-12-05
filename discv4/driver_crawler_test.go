package discv4

import (
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPeerInfo(t *testing.T) {
	tests := []struct {
		name     string
		enodeURL string
		peerID   string
		maddrs   []string
	}{
		{
			name:     "bootstrap 1",
			enodeURL: "enode://d860a01f9722d78051619d1e2351aba3f43f943f6f00718d1b9baa4101932a1f5011f16bb2b1bb35db20d6fe28fa0bf09636d26a87d31de9ec6203eeedb1f666@18.138.108.67:30303",
			peerID:   "16Uiu2HAm9zKVtAMGmM7kzt9bJRYDXXYPG7RXH672iS53uiRUaQyU",
			maddrs: []string{
				"/ip4/18.138.108.67/udp/30303",
				"/ip4/18.138.108.67/tcp/30303",
			},
		},
		{
			name:     "bootstrap 2",
			enodeURL: "enode://2b252ab6a1d0f971d9722cb839a42cb81db019ba44c08754628ab4a823487071b5695317c8ccd085219c3a03af063495b2f1da8d18218da2d6a82981b45e6ffc@65.108.70.101:30303",
			peerID:   "16Uiu2HAkxL6P5oaKh6cfgkKL6ZL3AE1kGziw51sykS3DLP6YxbP6",
			maddrs: []string{
				"/ip4/65.108.70.101/udp/30303",
				"/ip4/65.108.70.101/tcp/30303",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := enode.ParseV4(tt.enodeURL)
			require.NoError(t, err)

			got, err := NewPeerInfo(node)
			assert.Nil(t, err)
			assert.Equal(t, tt.peerID, got.peerID.String())

			for i, maddr := range got.maddrs {
				assert.Equal(t, tt.maddrs[i], maddr.String())
			}
		})
	}
}
