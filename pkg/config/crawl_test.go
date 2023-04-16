package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestInit_BootstrapPeers(t *testing.T) {
	config := Crawl{
		Root:           &Root{Protocols: cli.NewStringSlice()},
		BootstrapPeers: cli.NewStringSlice(),
		Network:        string(NetworkIPFS),
	}
	err := config.ConfigureNetwork()
	require.NoError(t, err)

	assert.Greater(t, len(config.BootstrapPeers.Value()), 0)
}

func TestConfig_BootstrapAddrInfos(t *testing.T) {
	config := Crawl{
		Root:           &Root{Protocols: cli.NewStringSlice()},
		BootstrapPeers: cli.NewStringSlice(),
		Network:        string(NetworkIPFS),
	}

	addrInfos, err := config.BootstrapAddrInfos()
	require.NoError(t, err)

	assert.Equal(t, len(config.BootstrapPeers.Value()), len(addrInfos))
}
