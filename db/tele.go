package db

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/dennis-tra/nebula-crawler/tele"
)

// pgTelemetry holds the relevant items to manage the engine's pgTelemetry
type pgTelemetry struct {
	tracer               trace.Tracer
	cacheQueriesCount    metric.Int64Counter
	insertVisitHistogram metric.Int64Histogram
}

func newPGTelemetry(tp trace.TracerProvider, mp metric.MeterProvider) (*pgTelemetry, error) {
	meter := mp.Meter(tele.MeterName)

	cacheQueriesCount, err := meter.Int64Counter("cache_queries", metric.WithDescription("Number of queries to the LRU caches"))
	if err != nil {
		return nil, fmt.Errorf("cache_queries counter: %w", err)
	}

	insertVisitHistogram, err := meter.Int64Histogram("insert_visit_timing", metric.WithDescription("Histogram of database query times for visit insertions"), metric.WithUnit("milliseconds"))
	if err != nil {
		return nil, fmt.Errorf("insert_visit_timing histogram: %w", err)
	}

	return &pgTelemetry{
		tracer:               tp.Tracer(tele.TracerName),
		cacheQueriesCount:    cacheQueriesCount,
		insertVisitHistogram: insertVisitHistogram,
	}, nil
}

// chTelemetry holds the relevant items to manage the engine's chTelemetry
type chTelemetry struct {
	tracer                 trace.Tracer
	insertCounter          metric.Int64Counter
	insertLatencyHistogram metric.Int64Histogram
}

func newCHTelemetry(tp trace.TracerProvider, mp metric.MeterProvider) (*chTelemetry, error) {
	meter := mp.Meter(tele.MeterName)

	insertCounter, err := meter.Int64Counter("inserts", metric.WithDescription("Number of written records"), metric.WithUnit("1"))
	if err != nil {
		return nil, fmt.Errorf("inserts counter: %w", err)
	}

	insertHistogram, err := meter.Int64Histogram("insert_latency", metric.WithDescription("Histogram of database query times for insertions"), metric.WithUnit("milliseconds"))
	if err != nil {
		return nil, fmt.Errorf("insert_latency histogram: %w", err)
	}

	return &chTelemetry{
		tracer:                 tp.Tracer(tele.TracerName),
		insertCounter:          insertCounter,
		insertLatencyHistogram: insertHistogram,
	}, nil
}
