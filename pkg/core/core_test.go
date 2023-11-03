package core

import (
	"context"
	"strconv"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/mock"

	"github.com/dennis-tra/nebula-crawler/pkg/utils"
)

// interface assertions
var (
	_ Handler[*testPeerInfo, DialResult[*testPeerInfo]]  = (*DialHandler[*testPeerInfo])(nil)
	_ Handler[*testPeerInfo, CrawlResult[*testPeerInfo]] = (*CrawlHandler[*testPeerInfo])(nil)
	_ WorkResult[*testPeerInfo]                          = CrawlResult[*testPeerInfo]{}
	_ WorkResult[*testPeerInfo]                          = DialResult[*testPeerInfo]{}
	_ Worker[CrawlResult[*testPeerInfo], WriteResult]    = (*CrawlWriter[*testPeerInfo])(nil)
	_ Worker[DialResult[*testPeerInfo], WriteResult]     = (*DialWriter[*testPeerInfo])(nil)
)

type testPeerInfo struct {
	peerID peer.ID
	addrs  []ma.Multiaddr
}

var _ PeerInfo[*testPeerInfo] = (*testPeerInfo)(nil)

func (p *testPeerInfo) Addrs() []ma.Multiaddr {
	return p.addrs
}

func (p *testPeerInfo) ID() peer.ID {
	return p.peerID
}

func (p *testPeerInfo) Merge(other *testPeerInfo) *testPeerInfo {
	if p.peerID != other.peerID {
		panic("merge peer ID mismatch")
	}

	return &testPeerInfo{
		peerID: p.peerID,
		addrs:  utils.MergeMaddrs(p.addrs, other.addrs),
	}
}

type testDriver struct {
	mock.Mock
}

var _ Driver[*testPeerInfo, CrawlResult[*testPeerInfo]] = (*testDriver)(nil)

func (s *testDriver) NewWorker() (Worker[*testPeerInfo, CrawlResult[*testPeerInfo]], error) {
	args := s.Called()
	return args.Get(0).(Worker[*testPeerInfo, CrawlResult[*testPeerInfo]]), args.Error(1)
}

func (s *testDriver) NewWriter() (Worker[CrawlResult[*testPeerInfo], WriteResult], error) {
	args := s.Called()
	return args.Get(0).(Worker[CrawlResult[*testPeerInfo], WriteResult]), args.Error(1)
}

func (s *testDriver) Tasks() <-chan *testPeerInfo {
	args := s.Called()
	return args.Get(0).(<-chan *testPeerInfo)
}

func (s *testDriver) Close() {
	s.Called()
}

type testHandler struct {
	mock.Mock
}

var _ Handler[*testPeerInfo, CrawlResult[*testPeerInfo]] = (*testHandler)(nil)

func (h *testHandler) HandlePeerResult(r Result[CrawlResult[*testPeerInfo]]) []*testPeerInfo {
	args := h.Called()
	return args.Get(0).([]*testPeerInfo)
}

func (h *testHandler) HandleWriteResult(r Result[WriteResult]) {
	h.Called()
}

type testWorker[IN any, OUT any] struct {
	mock.Mock
}

var _ Worker[string, int] = (*testWorker[string, int])(nil)

func newTestWorker[IN any, OUT any]() *testWorker[IN, OUT] {
	return &testWorker[IN, OUT]{}
}

func newAtoiWorker(t *testing.T) *testWorker[string, int] {
	worker := &testWorker[string, int]{}
	call := worker.On("Work", mock.IsType(context.Background()), mock.IsType(""))
	call.RunFn = func(args mock.Arguments) {
		val, err := strconv.Atoi(args.String(1))
		call.ReturnArguments = mock.Arguments{val, err}
	}
	return worker
}

func newTestCrawler() *testWorker[*testPeerInfo, CrawlResult[*testPeerInfo]] {
	return &testWorker[*testPeerInfo, CrawlResult[*testPeerInfo]]{}
}

func newTestWriter() *testWorker[CrawlResult[*testPeerInfo], WriteResult] {
	return &testWorker[CrawlResult[*testPeerInfo], WriteResult]{}
}

func (w *testWorker[IN, OUT]) Work(ctx context.Context, task IN) (OUT, error) {
	args := w.Called(ctx, task)
	return args.Get(0).(OUT), args.Error(1)
}
