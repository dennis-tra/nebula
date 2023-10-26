package libp2p

import (
	"testing"

	"github.com/dennis-tra/nebula-crawler/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dennis-tra/nebula-crawler/pkg/db"
)

func TestStack_BootstrapPeers(t *testing.T) {
	cfg := &CrawlStackConfig{
		BootstrapPeerStrs: config.BootstrapPeersFilecoin,
	}

	s, err := NewCrawlStack(db.InitNoopClient(), nil, cfg)
	require.NoError(t, err)

	addrInfos, err := s.BootstrapPeers()
	require.NoError(t, err)

	assert.Equal(t, len(config.BootstrapPeersFilecoin), len(addrInfos))
}
