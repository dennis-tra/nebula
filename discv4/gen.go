package discv4

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/p2p/discover/v4wire"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

// GenRandomPublicKey generates a public key that, when hashed with Keccak256,
// yields a [v4wire.Pubkey] that has a common prefix length of targetCPL.
func GenRandomPublicKey(targetID enode.ID, targetCPL int) (v4wire.Pubkey, error) {
	targetIDStr := fmt.Sprintf("%b", targetID.Bytes()[:4])
	_ = targetIDStr
	targetPrefix := binary.BigEndian.Uint16(targetID[:])

	// For host with ID `L`, an ID `K` belongs to a bucket with ID `B` ONLY IF CommonPrefixLen(L,K) is EXACTLY B.
	// Hence, to achieve a targetPrefix `T`, we must toggle the (T+1)th bit in L & then copy (T+1) bits from L
	// to our randomly generated prefix.
	toggledTargetPrefix := targetPrefix ^ (uint16(0x8000) >> targetCPL)

	randUInt16Bytes := new([2]byte)
	_, err := rand.Read(randUInt16Bytes[:])
	if err != nil {
		return [64]byte{}, fmt.Errorf("read random bytes: %w", err)
	}
	randUint16 := binary.BigEndian.Uint16(randUInt16Bytes[:])

	// generate a mask that starts with targetCPL + 1 ones and the rest zeroes
	mask := (^uint16(0)) << (16 - (targetCPL + 1))

	// toggledTargetPrefix & mask: use the first targetCPL + 1 bits from the toggledTargetPrefix
	// randUint16 & ^mask: use the remaining bits from the random uint16
	// by or'ing them together with | we composed the final prefix
	prefix := (toggledTargetPrefix & mask) | (randUint16 & ^mask)

	// Lookup the preimage in the key prefix map
	key := keyPrefixMap[prefix]

	// generate public key
	out := new([64]byte)
	binary.BigEndian.PutUint32(out[:], key)

	return *out, nil
}
