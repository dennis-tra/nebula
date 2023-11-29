package core

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/dennis-tra/nebula-crawler/tele"
)

// telemetry holds the relevant items to manage the engine's telemetry
type telemetry[I PeerInfo[I], R WorkResult[I]] struct {
	// A tracer to use for tracing - lol, genius comment
	tracer trace.Tracer

	// obsChan is a channel that the engine will consume. The given function
	// will then be executed in the same go-routine as the main run loop.
	// Because open telemetry operates with callbacks for gauges, each
	// gauge-callback will send an anonymous function to this channel that
	// performs the metric observation. This function will then get executed
	// by the engine in the correct go-routine. Sending on the channel is
	// guarded by a context that can be cancelled with the below obsCancel
	// function. If that cancel function is called, the corresponding context
	// will be cancelled, and sending on the obsChan channel won't block. So to
	// stop the telemetry collection, call obsCancel and then close obsChan.
	obsChan   chan func(*Engine[I, R])
	obsCancel context.CancelFunc

	// counts the number of visits we have performed
	visitCount metric.Int64Counter
	taskCount  metric.Int64Counter
	shutdown   chan struct{}
	done       sync.WaitGroup
}

func newTelemetry[I PeerInfo[I], R WorkResult[I]](tp trace.TracerProvider, mp metric.MeterProvider) (*telemetry[I, R], error) {
	meter := mp.Meter(tele.MeterName)

	shutdown := make(chan struct{})

	obsChan := make(chan func(*Engine[I, R]))

	gauges := []struct {
		name        string
		description string
		observeFn   func(o metric.Int64Observer, e *Engine[I, R])
	}{
		{
			name:        "visit_queue_length",
			description: "Number of peers in the queue to visit",
			observeFn: func(o metric.Int64Observer, e *Engine[I, R]) {
				o.Observe(int64(e.peerQueue.Len()))
			},
		},
		{
			name:        "inflight_queue_length",
			description: "Number of inflight crawls",
			observeFn: func(o metric.Int64Observer, e *Engine[I, R]) {
				o.Observe(int64(len(e.inflight)))
			},
		},
		{
			name:        "processed_peers",
			description: "Number of processed peers",
			observeFn: func(o metric.Int64Observer, e *Engine[I, R]) {
				o.Observe(int64(len(e.processed)))
			},
		},
		{
			name:        "write_queue_length",
			description: "Number of processing results ready to be written to the DB",
			observeFn: func(o metric.Int64Observer, e *Engine[I, R]) {
				o.Observe(int64(e.writeQueue.Len()))
			},
		},
		{
			name:        "written_results",
			description: "Number of written peer results",
			observeFn: func(o metric.Int64Observer, e *Engine[I, R]) {
				o.Observe(int64(e.writeCount))
			},
		},
	}

	var done sync.WaitGroup
	for _, gauge := range gauges {
		gauge := gauge
		_, err := meter.Int64ObservableGauge(
			gauge.name,
			metric.WithDescription(gauge.description),
			metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
				done.Add(1)
				defer done.Done()

				select {
				case <-shutdown:
					return nil
				default:
					// pass
				}

				observeFn := func(e *Engine[I, R]) {
					gauge.observeFn(o, e)
				}

				select {
				case <-shutdown:
				case obsChan <- observeFn:
				}

				return nil
			}),
		)
		if err != nil {
			return nil, fmt.Errorf("create %s gauge: %w", gauge.name, err)
		}
	}

	visitCount, err := meter.Int64Counter("visits", metric.WithDescription("Total number visits"))
	if err != nil {
		return nil, fmt.Errorf("visit_count counter: %w", err)
	}

	engineTaskCount, err := meter.Int64Counter("engine_tasks", metric.WithDescription("Number of tasks the engine has processed"))
	if err != nil {
		return nil, fmt.Errorf("engine_tasks counter: %w", err)
	}

	return &telemetry[I, R]{
		tracer:     tp.Tracer(tele.TracerName),
		obsChan:    obsChan,
		shutdown:   shutdown,
		done:       done,
		visitCount: visitCount,
		taskCount:  engineTaskCount,
	}, nil
}

func (t *telemetry[I, R]) Stop() {
	close(t.shutdown)
	t.done.Wait()
	close(t.obsChan)
}
