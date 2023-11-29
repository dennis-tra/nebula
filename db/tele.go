package db

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/dennis-tra/nebula-crawler/tele"
)

// telemetry holds the relevant items to manage the engine's telemetry
type telemetry struct {
	tracer               trace.Tracer
	cacheQueriesCount    metric.Int64Counter
	insertVisitHistogram metric.Int64Histogram
}

func newTelemetry(tp trace.TracerProvider, mp metric.MeterProvider) (*telemetry, error) {
	meter := mp.Meter(tele.MeterName)

	cacheQueriesCount, err := meter.Int64Counter("cache_queries", metric.WithDescription("Number of queries to the LRU caches"))
	if err != nil {
		return nil, fmt.Errorf("cache_queries counter: %w", err)
	}

	insertVisitHistogram, err := meter.Int64Histogram("insert_visit_timing", metric.WithDescription("Histogram of database query times for visit insertions"), metric.WithUnit("milliseconds"))
	if err != nil {
		return nil, fmt.Errorf("cache_queries counter: %w", err)
	}

	return &telemetry{
		tracer:               tp.Tracer(tele.TracerName),
		cacheQueriesCount:    cacheQueriesCount,
		insertVisitHistogram: insertVisitHistogram,
	}, nil
}
