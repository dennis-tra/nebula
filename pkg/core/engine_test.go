package core

import (
	"context"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestEngineConfig_Validate(t *testing.T) {
	t.Run("default config valid", func(t *testing.T) {
		cfg := DefaultEngineConfig()
		assert.NoError(t, cfg.Validate())
	})

	t.Run("zero or negative crawlers", func(t *testing.T) {
		cfg := DefaultEngineConfig()
		cfg.WorkerCount = 0
		assert.Error(t, cfg.Validate())
		cfg.WorkerCount = -1
		assert.Error(t, cfg.Validate())
		cfg.WorkerCount = 1
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
	driver := &testDriver{}
	driver.On("NewWorker").Return(newTestCrawler(), nil)
	driver.On("NewWriter").Return(newTestWriter(), nil)
	driver.On("Tasks").Return(make(<-chan *testPeerInfo))

	handlerCfg := &CrawlHandlerConfig{
		TrackNeighbors: true,
	}

	t.Run("nil config", func(t *testing.T) {
		handler := NewCrawlHandler[*testPeerInfo](handlerCfg)
		eng, err := NewEngine[*testPeerInfo, CrawlResult[*testPeerInfo]](driver, handler, nil)
		assert.NotNil(t, eng)
		assert.NoError(t, err)
	})

	t.Run("invalid config", func(t *testing.T) {
		cfg := DefaultEngineConfig()
		cfg.WorkerCount = 0
		assert.Error(t, cfg.Validate())

		handler := NewCrawlHandler[*testPeerInfo](handlerCfg)
		eng, err := NewEngine[*testPeerInfo, CrawlResult[*testPeerInfo]](driver, handler, cfg)
		assert.Nil(t, eng)
		assert.Error(t, err)
	})

	t.Run("valid config", func(t *testing.T) {
		cfg := DefaultEngineConfig()
		handler := NewCrawlHandler[*testPeerInfo](handlerCfg)
		eng, err := NewEngine[*testPeerInfo, CrawlResult[*testPeerInfo]](driver, handler, cfg)
		assert.NotNil(t, eng)
		assert.NoError(t, err)

		assert.Len(t, eng.workerPool.workers, cfg.WorkerCount)
		assert.Len(t, eng.writerPool.workers, cfg.WriterCount)
		assert.NotNil(t, eng.peerQueue)
		assert.NotNil(t, eng.writeQueue)
		assert.NotNil(t, eng.handler)
		assert.NotNil(t, handler.PeerMappings)
		assert.NotNil(t, handler.RoutingTables)
		assert.NotNil(t, handler.CrawlErrs)
		assert.NotNil(t, eng.inflight)
		assert.NotNil(t, eng.processed)
	})
}

func TestNewEngine_Run(t *testing.T) {
	logrus.SetLevel(logrus.PanicLevel)

	t.Run("no bootstrap", func(t *testing.T) {
		tasksChan := make(chan *testPeerInfo)
		driver := &testDriver{}
		driver.On("NewWorker").Return(newTestCrawler(), nil)
		driver.On("NewWriter").Return(newTestWriter(), nil)
		driver.On("Close").Times(1)
		driver.On("Tasks").Return((<-chan *testPeerInfo)(tasksChan))

		close(tasksChan)

		handlerCfg := &CrawlHandlerConfig{
			TrackNeighbors: true,
		}
		handler := NewCrawlHandler[*testPeerInfo](handlerCfg)
		cfg := DefaultEngineConfig()
		eng, err := NewEngine[*testPeerInfo, CrawlResult[*testPeerInfo]](driver, handler, cfg)
		require.NoError(t, err)

		queuedPeers, err := eng.Run(context.Background())
		require.NoError(t, err)

		assert.Equal(t, 0, len(queuedPeers))
		assert.Equal(t, 0, handler.CrawledPeers)
		assert.Len(t, handler.RoutingTables, 0)
		assert.Len(t, handler.PeerMappings, 0)
		assert.Len(t, handler.CrawlErrs, 0)
	})

	t.Run("single peer", func(t *testing.T) {
		ctx := context.Background()

		tasksChan := make(chan *testPeerInfo, 1)

		testMaddr, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/3000")
		require.NoError(t, err)

		testPeer := &testPeerInfo{
			peerID: peer.ID("test-peer"),
			addrs:  []ma.Multiaddr{testMaddr},
		}

		tasksChan <- testPeer

		cr := CrawlResult[*testPeerInfo]{
			CrawlerID:    "1",
			Info:         testPeer,
			RoutingTable: &RoutingTable[*testPeerInfo]{},
		}

		crawler := newTestCrawler()
		crawler.On("Work", mock.IsType(ctx), mock.IsType(testPeer)).Return(cr, nil)

		writer := newTestWriter()
		writer.On("Work", mock.IsType(ctx), mock.IsType(cr)).Return(WriteResult{
			InsertVisitResult: &db.InsertVisitResult{},
		}, nil)

		driver := &testDriver{}
		driver.On("NewWorker").Return(crawler, nil)
		driver.On("NewWriter").Return(writer, nil)
		driver.On("Close").Times(1)
		driver.On("Tasks").Return((<-chan *testPeerInfo)(tasksChan))

		close(tasksChan)

		handlerCfg := &CrawlHandlerConfig{
			TrackNeighbors: false,
		}
		handler := NewCrawlHandler[*testPeerInfo](handlerCfg)

		cfg := DefaultEngineConfig()
		eng, err := NewEngine[*testPeerInfo, CrawlResult[*testPeerInfo]](driver, handler, cfg)
		require.NoError(t, err)

		queuedPeers, err := eng.Run(ctx)
		require.NoError(t, err)

		assert.Equal(t, 0, len(queuedPeers))
		assert.Equal(t, 1, handler.CrawledPeers)
		assert.Len(t, handler.RoutingTables, 0)
		assert.Len(t, handler.PeerMappings, 0)
		assert.Len(t, handler.CrawlErrs, 0)
	})
}
