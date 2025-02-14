package core

import (
	"context"
)

type DialHandlerConfig struct{}

type DialHandler[I PeerInfo[I]] struct {
	cfg *DialHandlerConfig
}

func NewDialHandler[I PeerInfo[I]](cfg *DialHandlerConfig) *DialHandler[I] {
	return &DialHandler[I]{
		cfg: cfg,
	}
}

func (h *DialHandler[I]) HandlePeerResult(ctx context.Context, result Result[DialResult[I]]) []I {
	return nil
}

func (h *DialHandler[I]) HandleWriteResult(ctx context.Context, result Result[WriteResult]) {}

func (h *DialHandler[I]) Summary(state *EngineState) *Summary {
	return &Summary{
		PeersRemaining: state.PeersQueued,
	}
}
