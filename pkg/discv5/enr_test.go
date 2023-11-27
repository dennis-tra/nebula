package discv5

import (
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestENREntryAttnets_DecodeRLP(t *testing.T) {
	enrStr := "enr:-Ma4QGagfQuxa-p1IOq9Id3b8C2wdftJEx5HPesyqFx__qtvE-0dsq6pdudI5qYbANJ1yilNJkW3AjRSSVApeEGBrYSCCRiHYXR0bmV0c4gAADwAAAAAAIRldGgykCGm-DYDAABk__________-CaWSCdjSCaXCEUkKtHIRxdWljgiMpiXNlY3AyNTZrMaED0fxBax-7g0smHY-UismExjzR3jjRCexY_5ZeuVcxm6yIc3luY25ldHMAg3RjcIIjKIN1ZHCCIyg"
	node, err := enode.Parse(enode.V4ID{}, enrStr)
	require.NoError(t, err)

	var attnetsEntry ENREntryAttnets
	err = node.Load(&attnetsEntry)
	require.NoError(t, err)

	assert.EqualValues(t, 4, attnetsEntry.AttnetsNum)
	assert.Equal(t, "0x00003c0000000000", attnetsEntry.Attnets)

	var eth2Entry ENREntryEth2
	err = node.Load(&eth2Entry)
	require.NoError(t, err)

	assert.Equal(t, "0x21a6f836", eth2Entry.ForkDigest.String())

	var syncnetsEntry ENREntrySyncCommsSubnet
	err = node.Load(&syncnetsEntry)
	require.NoError(t, err)

	assert.Equal(t, "0x00", syncnetsEntry.SyncNets)
}
