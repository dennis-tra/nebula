package metrics

import (
	"fmt"
	"net/http"

	ocprom "contrib.go.opencensus.io/exporter/prometheus"
	"contrib.go.opencensus.io/integrations/ocsql"
	kadmetrics "github.com/libp2p/go-libp2p-kad-dht/metrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

func RegisterCrawlMetrics() error {
	if err := view.Register(kadmetrics.DefaultViews...); err != nil {
		return errors.Wrap(err, "register kademlia default views")
	}
	if err := view.Register(DefaultCrawlViews...); err != nil {
		return errors.Wrap(err, "register nebula default crawl views")
	}
	return nil
}

func RegisterMonitorMetrics() error {
	if err := view.Register(kadmetrics.DefaultViews...); err != nil {
		return errors.Wrap(err, "register kademlia default views")
	}

	if err := view.Register(DefaultMonitorViews...); err != nil {
		return errors.Wrap(err, "register nebula default monitor views")
	}
	return nil
}

func ListenAndServe(host string, port int) error {
	// Register default Go and process metrics
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// Initialize new exporter instance
	pe, err := ocprom.NewExporter(ocprom.Options{
		Namespace: "nebula",
		Registry:  registry,
	})
	if err != nil {
		return errors.Wrap(err, "new prometheus exporter")
	}

	// Enable ocsql metrics with OpenCensus
	ocsql.RegisterAllViews()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", pe)
		if err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), mux); err != nil {
			log.Fatalf("Failed to run Prometheus /metrics endpoint: %v", err)
		}
	}()
	return nil
}

// Keys
var (
	KeyAgentVersion, _ = tag.NewKey("agent_version")
	KeyError, _        = tag.NewKey("error")
)

// Distributions
var (
	neighborsDistribution = view.Distribution(100, 150, 200, 210, 220, 230, 240, 250, 260, 270, 280, 290, 300)
)

// Measures
var (
	CrawlConnectsCount         = stats.Float64("crawl_connects_count", "Number of connection establishment attempts during crawl", stats.UnitDimensionless)
	MonitorDialCount           = stats.Float64("monitor_dials_count", "Number of dial attempts during monitoring", stats.UnitDimensionless)
	CrawlConnectErrorsCount    = stats.Float64("crawl_connect_errors_count", "Number of successful connection establishment errors during crawl", stats.UnitDimensionless)
	MonitorDialErrorsCount     = stats.Float64("monitor_dial_errors_count", "Number of successful dial errors during monitoring", stats.UnitDimensionless)
	FetchedNeighborsCount      = stats.Float64("fetched_neighbors_count", "Number of neighbors fetched from a peer", stats.UnitDimensionless)
	CrawledPeersCount          = stats.Float64("crawled_peers_count", "Number of distinct peers found for a peer crawl", stats.UnitDimensionless)
	CrawledUpsertDuration      = stats.Float64("crawled_upsert_duration", "Amount of time we need to populate the database with one crawl result", stats.UnitMilliseconds)
	PeersToCrawlCount          = stats.Float64("peers_to_crawl_count", "Number of peers in the queue to crawl", stats.UnitDimensionless)
	PeersToDialCount           = stats.Float64("peers_to_dial_count", "Number of peers in the queue to dial", stats.UnitDimensionless)
	PeersToDialErrorsCount     = stats.Float64("peers_to_dial_errors_count", "Number of errors when dialing peers", stats.UnitDimensionless)
	AgentVersionCacheHitCount  = stats.Float64("agent_version_cache_hit_count", "Number of agent version cache hits", stats.UnitDimensionless)
	AgentVersionCacheMissCount = stats.Float64("agent_version_cache_miss_count", "Number of agent version cache misses", stats.UnitDimensionless)
	ProtocolCacheHitCount      = stats.Float64("protocol_cache_hit_count", "Number of protocol cache hits", stats.UnitDimensionless)
	ProtocolCacheMissCount     = stats.Float64("protocol_cache_miss_count", "Number of protocol cache misses", stats.UnitDimensionless)
)

// Views
var (
	CrawlConnectsCountView = &view.View{
		Measure:     CrawlConnectsCount,
		Aggregation: view.Count(),
	}
	MonitorDialsCountView = &view.View{
		Measure:     MonitorDialCount,
		Aggregation: view.Count(),
	}
	CrawlConnectErrorsCountView = &view.View{
		Measure:     CrawlConnectErrorsCount,
		Aggregation: view.Count(),
	}
	MonitorDialErrorsCountView = &view.View{
		Measure:     MonitorDialErrorsCount,
		Aggregation: view.Count(),
	}
	FetchedNeighborsCountView = &view.View{
		Measure:     FetchedNeighborsCount,
		TagKeys:     []tag.Key{KeyAgentVersion},
		Aggregation: neighborsDistribution,
	}
	CrawledPeersCountView = &view.View{
		Measure:     CrawledPeersCount,
		Aggregation: view.Count(),
	}
	CrawledUpsertDurationView = &view.View{
		Measure:     CrawledUpsertDuration,
		Aggregation: ocsql.DefaultMillisecondsDistribution,
	}
	PeersToCrawlCountView = &view.View{
		Measure:     PeersToCrawlCount,
		Aggregation: view.LastValue(),
	}
	PeersToDialCountView = &view.View{
		Measure:     PeersToDialCount,
		Aggregation: view.LastValue(),
	}
	PeersToDialErrorsCountView = &view.View{
		Measure:     PeersToDialErrorsCount,
		TagKeys:     []tag.Key{KeyError},
		Aggregation: view.Count(),
	}
	AgentVersionCacheHitCountView = &view.View{
		Measure:     AgentVersionCacheHitCount,
		Aggregation: view.Count(),
	}
	AgentVersionCacheMissCountView = &view.View{
		Measure:     AgentVersionCacheMissCount,
		Aggregation: view.Count(),
	}
	ProtocolCacheHitCountView = &view.View{
		Measure:     ProtocolCacheHitCount,
		Aggregation: view.Count(),
	}
	ProtocolCacheMissCountView = &view.View{
		Measure:     ProtocolCacheMissCount,
		Aggregation: view.Count(),
	}
)

// DefaultCrawlViews with all views in it.
var DefaultCrawlViews = []*view.View{
	CrawlConnectsCountView,
	CrawlConnectErrorsCountView,
	FetchedNeighborsCountView,
	CrawledPeersCountView,
	PeersToCrawlCountView,
	CrawledUpsertDurationView,
}

// DefaultMonitorViews with all views in it.
var DefaultMonitorViews = []*view.View{
	PeersToDialCountView,
	MonitorDialsCountView,
	MonitorDialErrorsCountView,
	PeersToDialErrorsCountView,
	AgentVersionCacheHitCountView,
	AgentVersionCacheMissCountView,
	ProtocolCacheHitCountView,
	ProtocolCacheMissCountView,
}
