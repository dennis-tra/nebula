package discv5

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/bits"

	"github.com/ethereum/go-ethereum/core/forkid"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rlp"
	beacon "github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
)

type ENREntryEth struct {
	ForkID forkid.ID
	Tail   []rlp.RawValue `rlp:"tail"`
}

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

type ENREntryLes struct {
	VfxVersion uint
	Rest       []rlp.RawValue `rlp:"tail"`
}

type ENREntrySnap struct {
	Rest []rlp.RawValue `rlp:"tail"`
}

type ENREntryBsc struct {
	Rest []rlp.RawValue `rlp:"tail"`
}

type ENREntryTrust struct {
	Rest []rlp.RawValue `rlp:"tail"`
}

type ENREntryPtStack struct {
	ChainID uint64
	Version uint64
}
type ENREntryOpera struct {
	ForkID forkid.ID
	Rest   []rlp.RawValue `rlp:"tail"`
}

type ENREntryOptimism struct {
	ChainID uint64
	Version uint64
}

// Cap is the structure of a peer capability.
type Cap struct {
	Name    string
	Version uint
}

func (c Cap) String() string {
	return fmt.Sprintf("%s/%d", c.Name, c.Version)
}

type ENREntryCaps []Cap

func (e ENREntryCaps) Len() int      { return len(e) }
func (e ENREntryCaps) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e ENREntryCaps) Less(i, j int) bool {
	return e[i].Name < e[j].Name || (e[i].Name == e[j].Name && e[i].Version < e[j].Version)
}

type ENREntryTestID []byte

var (
	_ enr.Entry = (*ENREntryEth)(nil)
	_ enr.Entry = (*ENREntryEth2)(nil)
	_ enr.Entry = (*ENREntryAttnets)(nil)
	_ enr.Entry = (*ENREntrySyncCommsSubnet)(nil)
	_ enr.Entry = (*ENREntryOpStack)(nil)
	_ enr.Entry = (*ENREntryLes)(nil)
	_ enr.Entry = (*ENREntrySnap)(nil)
	_ enr.Entry = (*ENREntryBsc)(nil)
	_ enr.Entry = (*ENREntryTrust)(nil)
	_ enr.Entry = (*ENREntryPtStack)(nil)
	_ enr.Entry = (*ENREntryOpera)(nil)
	_ enr.Entry = (*ENREntryCaps)(nil)
	_ enr.Entry = (*ENREntryTestID)(nil)
	_ enr.Entry = (*ENREntryOptimism)(nil)
)

func (e *ENREntryEth) ENRKey() string             { return "eth" }
func (e *ENREntryEth2) ENRKey() string            { return "eth2" }
func (e *ENREntryAttnets) ENRKey() string         { return "attnets" }
func (e *ENREntrySyncCommsSubnet) ENRKey() string { return "syncnets" }
func (e *ENREntryOpStack) ENRKey() string         { return "opstack" }
func (e *ENREntryLes) ENRKey() string             { return "les" }
func (e *ENREntrySnap) ENRKey() string            { return "snap" }
func (e *ENREntryBsc) ENRKey() string             { return "bsc" }
func (e *ENREntryTrust) ENRKey() string           { return "trust" }
func (e *ENREntryPtStack) ENRKey() string         { return "ptstack" }
func (e *ENREntryOpera) ENRKey() string           { return "opera" }
func (e *ENREntryOptimism) ENRKey() string        { return "optimism" }
func (e ENREntryCaps) ENRKey() string             { return "cap" }
func (e ENREntryTestID) ENRKey() string           { return "testid" }

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
	} else if len(b) != 8 {
		return fmt.Errorf("attnets ENR entry must be 8 bytes long, got %d", len(b))
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
	var err error
	e.ChainID, e.Version, err = decodeChainIDAndVersion(s)
	return err
}

func (e *ENREntryPtStack) DecodeRLP(s *rlp.Stream) error {
	var err error
	e.ChainID, e.Version, err = decodeChainIDAndVersion(s)
	return err
}

func (e *ENREntryOptimism) DecodeRLP(s *rlp.Stream) error {
	var err error
	e.ChainID, e.Version, err = decodeChainIDAndVersion(s)
	return err
}

func decodeChainIDAndVersion(s *rlp.Stream) (uint64, uint64, error) {
	b, err := s.Bytes()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode outer ENR entry: %w", err)
	}

	r := bytes.NewReader(b)
	chainID, err := binary.ReadUvarint(r)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read chain ID var int: %w", err)
	}

	version, err := binary.ReadUvarint(r)
	if err != nil {
		return chainID, 0, fmt.Errorf("failed to read version var int: %w", err)
	}

	return chainID, version, nil
}
