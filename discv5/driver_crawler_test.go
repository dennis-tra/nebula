package discv5

import (
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPeerInfo(t *testing.T) {
	type args struct {
		node *enode.Node
	}
	tests := []struct {
		name   string
		enr    string
		peerID string
		maddrs []string
	}{
		{
			name:   "bootstrap 1",
			enr:    "enr:-Ku4QEWzdnVtXc2Q0ZVigfCGggOVB2Vc1ZCPEc6j21NIFLODSJbvNaef1g4PxhPwl_3kax86YPheFUSLXPRs98vvYsoBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpC1MD8qAAAAAP__________gmlkgnY0gmlwhDZBrP2Jc2VjcDI1NmsxoQM6jr8Rb1ktLEsVcKAPa08wCsKUmvoQ8khiOl_SLozf9IN1ZHCCIyg",
			peerID: "16Uiu2HAmGbaF2hpC4iumV7eWiK51nWmsZf7mK2R9nCM6rrvRQCwu",
			maddrs: []string{
				"/ip4/54.65.172.253/udp/9000",
			},
		},
		{
			name:   "bootstrap 2",
			enr:    "enr:-LK4QKWrXTpV9T78hNG6s8AM6IO4XH9kFT91uZtFg1GcsJ6dKovDOr1jtAAFPnS2lvNltkOGA9k29BUN7lFh_sjuc9QBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpC1MD8qAAAAAP__________gmlkgnY0gmlwhANAdd-Jc2VjcDI1NmsxoQLQa6ai7y9PMN5hpLe5HmiJSlYzMuzP7ZhwRiwHvqNXdoN0Y3CCI4yDdWRwgiOM",
			peerID: "16Uiu2HAm9TFzGRAgJKVyfZ54nDbCYNwEAFnpz6ybjM7KaVPLJZpy",
			maddrs: []string{
				"/ip4/3.64.117.223/udp/9100",
				"/ip4/3.64.117.223/tcp/9100",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := enode.Parse(enode.ValidSchemes, tt.enr)
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
