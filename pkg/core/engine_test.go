package core

import (
	"context"
	"testing"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/nebtest"
	"github.com/libp2p/go-libp2p/core/peer"
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
		defer goleak.VerifyNone(t, goleakIgnore...)

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
		defer goleak.VerifyNone(t, goleakIgnore...)

		ctx, cancel := context.WithCancel(context.Background()) // cancelCtx to satisfy mock.IsType below
		defer cancel()

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

func TestNewEngine_Run_parking_peers(t *testing.T) {
	defer goleak.VerifyNone(t, goleakIgnore...)
	logrus.SetLevel(logrus.PanicLevel)

	ctx, cancel := context.WithCancel(context.Background()) // cancelCtx to satisfy mock.IsType below
	defer cancel()

	bootstrapPeer := &testPeerInfo{
		peerID: peer.ID("bootstrap"),
		addrs:  []ma.Multiaddr{nebtest.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/3000")},
	}
	targetPeerWithoutAddrs := &testPeerInfo{
		peerID: peer.ID("target"),
		addrs:  []ma.Multiaddr{},
	}
	intermediatePeer := &testPeerInfo{
		peerID: peer.ID("intermediate"),
		addrs:  []ma.Multiaddr{nebtest.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/3002")},
	}
	targetPeerWithAddrs := &testPeerInfo{
		peerID: peer.ID("target"),
		addrs:  []ma.Multiaddr{nebtest.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/3001")},
	}

	tasksChan := make(chan *testPeerInfo, 1)
	tasksChan <- bootstrapPeer
	close(tasksChan)

	bootstrapPeerCrawlResult := CrawlResult[*testPeerInfo]{
		CrawlerID: "1",
		Info:      bootstrapPeer,
		RoutingTable: &RoutingTable[*testPeerInfo]{
			PeerID: bootstrapPeer.peerID,
			Neighbors: []*testPeerInfo{
				targetPeerWithoutAddrs,
				intermediatePeer,
			},
		},
	}

	intermediatePeerCrawlResult := CrawlResult[*testPeerInfo]{
		CrawlerID: "1",
		Info:      intermediatePeer,
		RoutingTable: &RoutingTable[*testPeerInfo]{
			PeerID: intermediatePeer.peerID,
			Neighbors: []*testPeerInfo{
				targetPeerWithAddrs,
				bootstrapPeer,
			},
		},
	}

	targetPeerCrawlResult := CrawlResult[*testPeerInfo]{
		CrawlerID: "1",
		Info:      targetPeerWithAddrs,
		RoutingTable: &RoutingTable[*testPeerInfo]{
			PeerID: targetPeerWithAddrs.peerID,
			Neighbors: []*testPeerInfo{
				intermediatePeer,
				bootstrapPeer,
			},
		},
	}

	crawler := newTestCrawler()
	writerCfg := &CrawlWriterConfig{
		AddrTrackType: config.AddrTypeAny,
	}
	writer := NewCrawlWriter[*testPeerInfo]("1", db.InitNoopClient(), 1, writerCfg)

	crawler.On("Work", mock.IsType(ctx), bootstrapPeer).
		Return(bootstrapPeerCrawlResult, nil).Times(1)
	crawler.On("Work", mock.IsType(ctx), intermediatePeer).
		Return(intermediatePeerCrawlResult, nil).Times(1)
	crawler.On("Work", mock.IsType(ctx), targetPeerWithAddrs).
		Return(targetPeerCrawlResult, nil).Times(1)

	driver := &testDriver{}
	driver.On("NewWorker").Return(crawler, nil)
	driver.On("NewWriter").Return(writer, nil)
	driver.On("Close").Times(1)
	driver.On("Tasks").Return((<-chan *testPeerInfo)(tasksChan))

	handler := NewCrawlHandler[*testPeerInfo](&CrawlHandlerConfig{})

	cfg := DefaultEngineConfig()
	cfg.WorkerCount = 1
	eng, err := NewEngine[*testPeerInfo, CrawlResult[*testPeerInfo]](driver, handler, cfg)
	require.NoError(t, err)

	queuedPeers, err := eng.Run(ctx)
	require.NoError(t, err)

	require.Len(t, queuedPeers, 0)
}
