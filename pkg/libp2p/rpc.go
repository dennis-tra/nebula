package libp2p

import (
	"context"
	"time"

	kadmetrics "github.com/libp2p/go-libp2p-kad-dht/metrics"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-msgio/protoio"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

// msgSender handles sending wire protocol messages to a given peer
type msgSender struct {
	h         host.Host
	protocols []protocol.ID
	timeout   time.Duration
}

// SendRequest sends a peer a message and waits for its response
func (ms *msgSender) SendRequest(ctx context.Context, p peer.ID, pmes *pb.Message) (*pb.Message, error) {
	ctx, _ = tag.New(ctx, kadmetrics.UpsertMessageType(pmes))
	defer stats.Record(ctx, kadmetrics.SentRequests.M(1))

	start := time.Now()
	tctx, cancel := context.WithTimeout(ctx, ms.timeout)
	defer cancel()
	s, err := ms.h.NewStream(tctx, p, ms.protocols...)
	if err != nil {
		stats.Record(ctx, kadmetrics.SentRequestErrors.M(1))
		return nil, err
	}

	w := protoio.NewDelimitedWriter(s)
	if err = w.WriteMsg(pmes); err != nil {
		stats.Record(ctx, kadmetrics.SentRequestErrors.M(1))
		return nil, err
	}

	r := protoio.NewDelimitedReader(s, network.MessageSizeMax)
	tctx, cancel = context.WithTimeout(ctx, ms.timeout)
	defer cancel()
	defer func() { _ = s.Close() }()

	msg := new(pb.Message)
	if err := ctxReadMsg(tctx, r, msg); err != nil {
		stats.Record(ctx, kadmetrics.SentRequestErrors.M(1))
		_ = s.Reset()
		return nil, err
	}

	stats.Record(ctx,
		kadmetrics.SentBytes.M(int64(pmes.Size())),
		kadmetrics.OutboundRequestLatency.M(float64(time.Since(start))/float64(time.Millisecond)),
	)

	return msg, nil
}

func ctxReadMsg(ctx context.Context, rc protoio.ReadCloser, mes *pb.Message) error {
	errc := make(chan error, 1)
	go func(r protoio.ReadCloser) {
		defer close(errc)
		err := r.ReadMsg(mes)
		errc <- err
	}(rc)

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SendMessage sends a peer a message without waiting on a response
func (ms *msgSender) SendMessage(ctx context.Context, p peer.ID, pmes *pb.Message) error {
	ctx, _ = tag.New(ctx, kadmetrics.UpsertMessageType(pmes))
	defer stats.Record(ctx, kadmetrics.SentMessages.M(1))

	s, err := ms.h.NewStream(ctx, p, ms.protocols...)
	if err != nil {
		stats.Record(ctx, kadmetrics.SentMessageErrors.M(1))
		return err
	}
	defer func() { _ = s.Close() }()

	if err = protoio.NewDelimitedWriter(s).WriteMsg(pmes); err != nil {
		stats.Record(ctx, kadmetrics.SentMessageErrors.M(1))
		return err
	}

	stats.Record(ctx, kadmetrics.SentBytes.M(int64(pmes.Size())))
	return nil
}
