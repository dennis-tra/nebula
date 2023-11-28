package discv5

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/bits"

	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rlp"
	beacon "github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
)

type ENREntryEth2 struct {
	beacon.Eth2Data
}

type ENREntryAttnets struct {
	AttnetsNum int
	Attnets    string
}

type ENREntrySyncCommsSubnet struct {
	SyncNets string
}

// ENREntryOpStack from https://github.com/ethereum-optimism/optimism/blob/85d932810bafc9084613b978d42cd770bc044eb4/op-node/p2p/discovery.go#L172
type ENREntryOpStack struct {
	ChainID uint64
	Version uint64
}

var (
	_ enr.Entry = (*ENREntryEth2)(nil)
	_ enr.Entry = (*ENREntryAttnets)(nil)
	_ enr.Entry = (*ENREntrySyncCommsSubnet)(nil)
	_ enr.Entry = (*ENREntryOpStack)(nil)
)

func (e *ENREntryEth2) ENRKey() string            { return "eth2" }
func (e *ENREntryAttnets) ENRKey() string         { return "attnets" }
func (e *ENREntrySyncCommsSubnet) ENRKey() string { return "syncnets" }
func (e *ENREntryOpStack) ENRKey() string         { return "opstack" }

func (e *ENREntryEth2) DecodeRLP(s *rlp.Stream) error {
	b, err := s.Bytes()
	if err != nil {
		return fmt.Errorf("failed to get bytes for attnets ENR entry: %w", err)
	}

	if err := e.Eth2Data.Deserialize(codec.NewDecodingReader(bytes.NewReader(b), uint64(len(b)))); err != nil {
		return fmt.Errorf("deserialize eth2 beacon data ENR entry: %w", err)
	}

	return nil
}

func (e *ENREntryAttnets) DecodeRLP(s *rlp.Stream) error {
	b, err := s.Bytes()
	if err != nil {
		return fmt.Errorf("failed to get bytes for attnets ENR entry: %w", err)
	}

	e.AttnetsNum = bits.OnesCount64(binary.BigEndian.Uint64(b))
	e.Attnets = "0x" + hex.EncodeToString(b)

	return nil
}

func (e *ENREntrySyncCommsSubnet) DecodeRLP(s *rlp.Stream) error {
	b, err := s.Bytes()
	if err != nil {
		return fmt.Errorf("failed to get bytes for syncnets ENR entry: %w", err)
	}

	// check out https://github.com/prysmaticlabs/prysm/blob/203dc5f63b060821c2706f03a17d66b3813c860c/beacon-chain/p2p/subnets.go#L221
	e.SyncNets = "0x" + hex.EncodeToString(b)

	return nil
}

func (e *ENREntryOpStack) DecodeRLP(s *rlp.Stream) error {
	b, err := s.Bytes()
	if err != nil {
		return fmt.Errorf("failed to decode outer ENR entry: %w", err)
	}
	// We don't check the byte length: the below readers are limited, and the ENR itself has size limits.
	// Future "opstack" entries may contain additional data, and will be tagged with a newer Version etc.
	r := bytes.NewReader(b)
	chainID, err := binary.ReadUvarint(r)
	if err != nil {
		return fmt.Errorf("failed to read chain ID var int: %w", err)
	}
	version, err := binary.ReadUvarint(r)
	if err != nil {
		return fmt.Errorf("failed to read Version var int: %w", err)
	}
	e.ChainID = chainID
	e.Version = version
	return nil
}

// "cap"
// "les"
// "opera"
