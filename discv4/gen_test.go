package discv4

import (
	"encoding/hex"
	"math/bits"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestGenRandomPublicKey(t *testing.T) {
	tests := []struct {
		targetID string
	}{
		{targetID: "c845e51a5e470e445ad424f7cb516339237f469ad7b3c903221b5c49ce55863f"},
	}
	for _, tt := range tests {
		t.Run(tt.targetID, func(t *testing.T) {
			nodeID, err := hex.DecodeString(tt.targetID)
			require.NoError(t, err)

			// fmt.Printf("    %08b\n", nodeID[:2])
			for i := 0; i < 16; i++ {
				got, err := GenRandomPublicKey(enode.ID(nodeID), i)
				require.NoError(t, err)

				gotHashed := crypto.Keccak256Hash(got[:])

				// fmt.Printf("[%d] %08b - %08b %08b\n", i, gotHashed[:2], nodeID[0]^gotHashed[0], nodeID[1]^gotHashed[1])

				lz := bits.LeadingZeros8(nodeID[0] ^ gotHashed[0])
				if i > 8 {
					lz += bits.LeadingZeros8(nodeID[1] ^ gotHashed[1])
				}
				assert.Equal(t, i, lz)
			}
		})
	}
}
