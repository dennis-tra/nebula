package core

type DialHandlerConfig struct{}

type DialHandler[I PeerInfo] struct {
	cfg *DialHandlerConfig
}

var _ Handler[PeerInfo, DialResult[PeerInfo]] = (*DialHandler[PeerInfo])(nil)

func NewDialHandler[I PeerInfo](cfg *DialHandlerConfig) *DialHandler[I] {
	return &DialHandler[I]{
		cfg: cfg,
	}
}

func (h *DialHandler[I]) HandleWorkResult(result Result[DialResult[I]]) []I {
	dr := result.Value
	_ = dr
	return nil
}

func (h *DialHandler[I]) HandleWriteResult(result Result[WriteResult]) {
	wr := result.Value
	_ = wr
}
