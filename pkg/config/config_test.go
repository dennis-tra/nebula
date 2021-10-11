package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (Config, func(t *testing.T)) {
	tmpLogLevel := log.GetLevel()
	log.SetLevel(log.ErrorLevel)

	tmpConfigFile := configFile
	configFile = filepath.Join(Prefix, "config_test.json")

	path, err := xdg.ConfigFile(configFile)
	require.NoError(t, err)
	_ = os.Remove(path)

	// Copy config to test with
	config := DefaultConfig

	teardown := func(t *testing.T) {
		_ = os.Remove(path)
		log.SetLevel(tmpLogLevel)
		configFile = tmpConfigFile
	}

	return config, teardown
}

func TestInit_BootstrapPeers(t *testing.T) {
	config, teardown := setup(t)
	defer teardown(t)

	assert.Greater(t, len(config.BootstrapPeers), 0)
}

func TestConfig_BootstrapAddrInfos(t *testing.T) {
	config, teardown := setup(t)
	defer teardown(t)

	addrInfos, err := config.BootstrapAddrInfos()
	require.NoError(t, err)

	assert.Equal(t, len(config.BootstrapPeers), len(addrInfos))
}

func Test_readConfig_noPath_notExists(t *testing.T) {
	_, teardown := setup(t)
	defer teardown(t)

	config, err := read("")
	require.NoError(t, err)
	assert.NotEmpty(t, config.Path)
	assert.False(t, config.Existed)
}

func Test_readConfig_noPath_exists(t *testing.T) {
	config, teardown := setup(t)
	defer teardown(t)

	err := config.Save()
	require.NoError(t, err)

	configLoaded, err := read("")
	require.NoError(t, err)
	assert.True(t, configLoaded.Existed)
}

func TestConfig_Save(t *testing.T) {
	config, teardown := setup(t)
	defer teardown(t)

	require.False(t, config.Existed)
	err := config.Save()
	require.NoError(t, err)
	assert.False(t, config.Existed)
}
