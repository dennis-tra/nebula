package discv5

import (
	"bytes"

	"github.com/ethereum/go-ethereum/p2p/enr"
	beacon "github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
)

// ENREntryAttnets is the ENR key of the subnet bitfield in the enr.
type ENREntryAttnets []byte

var _ enr.Entry = ENREntryAttnets{}

func (e ENREntryAttnets) ENRKey() string {
	return "attnets"
}

// ENREntryEth2 is the ENR key of the Ethereum consensus object in an enr.
type ENREntryEth2 []byte

var _ enr.Entry = ENREntryEth2{}

func (e ENREntryEth2) ENRKey() string {
	return "eth2"
}

func (e ENREntryEth2) Data() (*beacon.Eth2Data, error) {
	var data beacon.Eth2Data
	if err := data.Deserialize(codec.NewDecodingReader(bytes.NewReader(e), uint64(len(e)))); err != nil {
		return nil, err
	}
	return &data, nil
}

// ENREntrySyncCommsSubnet is the ENR key of the sync committee subnet bitfield in the enr.
type ENREntrySyncCommsSubnet []byte

var _ enr.Entry = ENREntrySyncCommsSubnet{}

func (e ENREntrySyncCommsSubnet) ENRKey() string {
	return "syncnets"
}
