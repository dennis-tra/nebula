package nebtest

import (
	"testing"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

// MustMultiaddr returns parses the given multi address string and stops the
// test with an error if that fails.
func MustMultiaddr(t testing.TB, maddrStr string) ma.Multiaddr {
	maddr, err := ma.NewMultiaddr(maddrStr)
	require.NoError(t, err)
	return maddr
}
