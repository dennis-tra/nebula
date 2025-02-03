package db

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

type NoopClient struct{}

func NewNoopClient() *NoopClient {
	return &NoopClient{}
}

func (n *NoopClient) InitCrawl(ctx context.Context, version string) error {
	return nil
}

func (n *NoopClient) SealCrawl(ctx context.Context, args *SealCrawlArgs) error {
	return nil
}

func (n *NoopClient) QueryBootstrapPeers(ctx context.Context, limit int) ([]peer.AddrInfo, error) {
	return []peer.AddrInfo{}, nil
}

func (n *NoopClient) InsertVisit(ctx context.Context, args *VisitArgs) error {
	return nil
}

func (n *NoopClient) InsertCrawlProperties(ctx context.Context, properties map[string]map[string]int) error {
	return nil
}

func (n *NoopClient) InsertNeighbors(ctx context.Context, peerID peer.ID, neighbors []peer.ID, errorBits uint16) error {
	return nil
}

func (n *NoopClient) SelectPeersToProbe(ctx context.Context) ([]peer.AddrInfo, error) {
	return []peer.AddrInfo{}, nil
}

func (n *NoopClient) Close() error {
	return nil
}
