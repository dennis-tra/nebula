package core

type DialHandlerConfig struct{}

type DialHandler[I PeerInfo[I]] struct {
	cfg *DialHandlerConfig
}

func NewDialHandler[I PeerInfo[I]](cfg *DialHandlerConfig) *DialHandler[I] {
	return &DialHandler[I]{
		cfg: cfg,
	}
}

func (h *DialHandler[I]) HandlePeerResult(result Result[DialResult[I]]) []I {
	return nil
}

func (h *DialHandler[I]) HandleWriteResult(result Result[WriteResult]) {}
