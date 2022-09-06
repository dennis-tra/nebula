package crawl

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
)

func TestNewCrawler_correctInit(t *testing.T) {
	conf := &config.Config{}
	crawler, err := NewCrawler(nil, conf)
	require.NoError(t, err)

	assert.NotNil(t, crawler.pm)
	assert.Equal(t, conf, crawler.config)
	assert.Equal(t, "crawler-01", crawler.id)

	// increments crawler service number
	crawler, err = NewCrawler(nil, conf)
	require.NoError(t, err)
	assert.Equal(t, "crawler-02", crawler.id)
}

func TestCrawler_StartCrawling_stopsOnShutdown(t *testing.T) {
	crawler, err := NewCrawler(nil, &config.Config{})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go crawler.StartCrawling(ctx, queue.NewFIFO(), queue.NewFIFO())

	time.Sleep(time.Millisecond * 100)

	cancel()
	<-crawler.done
}

func TestCrawler_handleCrawlJob_unlinked(t *testing.T) {
	utils.IDLength = 4

	ctx := context.Background()
	net := mocknet.New()

	h, err := net.GenPeer()
	require.NoError(t, err)

	remote, err := net.GenPeer()
	require.NoError(t, err)

	crawler, err := NewCrawler(h, &config.Config{})
	require.NoError(t, err)

	pi := peer.AddrInfo{
		ID:    remote.ID(),
		Addrs: remote.Addrs(),
	}

	result := crawler.handleCrawlJob(ctx, pi)
	assert.Error(t, result.ConnectError)
	assert.NotZero(t, result.CrawlStartTime)
	assert.NotZero(t, result.CrawlEndTime)
	assert.NotZero(t, result.ConnectStartTime)
	assert.NotZero(t, result.ConnectEndTime)
	assert.Zero(t, result.Agent)
	assert.Nil(t, result.RoutingTable.Neighbors)
	assert.Equal(t, pi, result.Peer)
}
