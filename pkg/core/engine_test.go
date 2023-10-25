package core

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/pkg/db"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestEngineConfig_Validate(t *testing.T) {
	t.Run("default config valid", func(t *testing.T) {
		cfg := DefaultEngineConfig()
		assert.NoError(t, cfg.Validate())
	})

	t.Run("zero or negative crawlers", func(t *testing.T) {
		cfg := DefaultEngineConfig()
		cfg.CrawlerCount = 0
		assert.Error(t, cfg.Validate())
		cfg.CrawlerCount = -1
		assert.Error(t, cfg.Validate())
		cfg.CrawlerCount = 1
		assert.NoError(t, cfg.Validate())
	})

	t.Run("zero or negative writers", func(t *testing.T) {
		cfg := DefaultEngineConfig()
		cfg.WriterCount = 0
		assert.Error(t, cfg.Validate())
		cfg.WriterCount = -1
		assert.Error(t, cfg.Validate())
		cfg.WriterCount = 1
		assert.NoError(t, cfg.Validate())
	})

	t.Run("limit can be anything", func(t *testing.T) {
		cfg := DefaultEngineConfig()
		cfg.Limit = 0
		assert.NoError(t, cfg.Validate())
		cfg.Limit = -100
		assert.NoError(t, cfg.Validate())
		cfg.Limit = 100
		assert.NoError(t, cfg.Validate())
	})
}

func TestNewEngine(t *testing.T) {
	stack := &testStack{}
	stack.On("NewCrawler").Return(newTestCrawler(), nil)
	stack.On("NewWriter").Return(newTestWriter(), nil)

	t.Run("nil config", func(t *testing.T) {
		eng, err := NewEngine[*testPeerInfo](stack, nil)
		assert.NotNil(t, eng)
		assert.NoError(t, err)
	})

	t.Run("invalid config", func(t *testing.T) {
		cfg := DefaultEngineConfig()
		cfg.CrawlerCount = 0
		assert.Error(t, cfg.Validate())

		eng, err := NewEngine[*testPeerInfo](stack, cfg)
		assert.Nil(t, eng)
		assert.Error(t, err)
	})

	t.Run("valid config", func(t *testing.T) {
		cfg := DefaultEngineConfig()
		eng, err := NewEngine[*testPeerInfo](stack, cfg)
		assert.NotNil(t, eng)
		assert.NoError(t, err)

		assert.Len(t, eng.crawlerPool.workers, cfg.CrawlerCount)
		assert.Len(t, eng.writerPool.workers, cfg.WriterCount)
		assert.NotNil(t, eng.crawlQueue)
		assert.NotNil(t, eng.writeQueue)
		assert.NotNil(t, eng.runData)
		assert.NotNil(t, eng.runData.PeerMappings)
		assert.NotNil(t, eng.runData.RoutingTables)
		assert.NotNil(t, eng.runData.ConnErrs)
		assert.NotNil(t, eng.inflight)
		assert.NotNil(t, eng.crawled)
	})
}

func TestNewEngine_Run(t *testing.T) {
	logrus.SetLevel(logrus.PanicLevel)

	t.Run("no bootstrap", func(t *testing.T) {
		stack := &testStack{}
		stack.On("NewCrawler").Return(newTestCrawler(), nil)
		stack.On("NewWriter").Return(newTestWriter(), nil)
		stack.On("OnClose").Times(1)
		stack.On("BootstrapPeers").Return([]*testPeerInfo{}, nil)

		cfg := DefaultEngineConfig()
		eng, err := NewEngine[*testPeerInfo](stack, cfg)
		require.NoError(t, err)

		data, err := eng.Run(context.Background())
		require.NoError(t, err)

		assert.Equal(t, 0, data.QueuedPeers)
		assert.Equal(t, 0, data.CrawledPeers)
		assert.Len(t, data.RoutingTables, 0)
		assert.Len(t, data.PeerMappings, 0)
		assert.Len(t, data.ConnErrs, 0)
	})

	t.Run("single peer", func(t *testing.T) {
		ctx := context.Background()

		testMaddr, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/3000")
		require.NoError(t, err)

		testPeer := &testPeerInfo{
			peerID: peer.ID("test-peer"),
			addrs:  []ma.Multiaddr{testMaddr},
		}
		cr := CrawlResult[*testPeerInfo]{
			CrawlerID: "1",
			Info:      testPeer,
		}

		crawler := newTestCrawler()
		crawler.On("Work", mock.IsType(ctx), mock.IsType(testPeer)).Return(cr, nil)

		writer := newTestWriter()
		writer.On("Work", mock.IsType(ctx), mock.IsType(cr)).Return(WriteResult{
			InsertVisitResult: &db.InsertVisitResult{},
		}, nil)

		stack := &testStack{}
		stack.On("NewCrawler").Return(crawler, nil)
		stack.On("NewWriter").Return(writer, nil)
		stack.On("OnPeerCrawled", cr, nil).Times(1)
		stack.On("OnClose").Times(1)
		stack.On("BootstrapPeers").Return([]*testPeerInfo{testPeer}, nil)

		cfg := DefaultEngineConfig()
		eng, err := NewEngine[*testPeerInfo](stack, cfg)
		require.NoError(t, err)

		data, err := eng.Run(ctx)
		require.NoError(t, err)

		assert.Equal(t, 0, data.QueuedPeers)
		assert.Equal(t, 1, data.CrawledPeers)
		assert.Len(t, data.RoutingTables, 0)
		assert.Len(t, data.PeerMappings, 0)
		assert.Len(t, data.ConnErrs, 0)
	})
}
