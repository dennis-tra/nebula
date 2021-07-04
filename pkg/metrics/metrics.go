package metrics

import (
	"fmt"
	"net/http"

	"contrib.go.opencensus.io/exporter/prometheus"
	"contrib.go.opencensus.io/integrations/ocsql"
	kadmetrics "github.com/libp2p/go-libp2p-kad-dht/metrics"
	"github.com/pkg/errors"
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
	pe, err := prometheus.NewExporter(prometheus.Options{Namespace: "nebula"})
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
)

// Distributions
var (
	defaultBytesDistribution        = view.Distribution(1024, 2048, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864, 268435456, 1073741824, 4294967296)
	fibonacciDistribution           = view.Distribution(1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 377)
	defaultMillisecondsDistribution = view.Distribution(0.01, 0.05, 0.1, 0.3, 0.6, 0.8, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000)
)

// Measures
var (
	CrawlConnectDuration      = stats.Float64("crawl_connect_duration", "Duration of connecting to peers during crawl", stats.UnitMilliseconds)
	MonitorConnectDuration    = stats.Float64("monitor_connect_duration", "Duration of connecting to peers during monitoring", stats.UnitMilliseconds)
	CrawlConnectsCount        = stats.Float64("crawl_connects_count", "Number of connection establishment attempts during crawl", stats.UnitDimensionless)
	MonitorConnectsCount      = stats.Float64("monitor_connects_count", "Number of connection establishment attempts during monitoring", stats.UnitDimensionless)
	CrawlConnectErrorsCount   = stats.Float64("crawl_connect_errors_count", "Number of successful connection establishment errors during crawl", stats.UnitDimensionless)
	MonitorConnectErrorsCount = stats.Float64("monitor_connect_errors_count", "Number of successful connection establishment errors during monitoring", stats.UnitDimensionless)
	FetchNeighborsDuration    = stats.Float64("fetch_neighbors_duration", "Duration of crawling a peer for all neighbours in its buckets", stats.UnitMilliseconds)
	FetchedNeighborsCount     = stats.Float64("fetched_neighbors_count", "Number of neighbors fetched from a peer", stats.UnitDimensionless)
	CrawledPeersCount         = stats.Float64("crawled_peers_count", "Number of distinct peers found for a peer crawl", stats.UnitDimensionless)
	CrawledUpsertDuration     = stats.Float64("crawled_upsert_duration", "Amount of time we need to populate the database with one crawl result", stats.UnitMilliseconds)
	PeerCrawlDuration         = stats.Float64("peer_crawl_duration", "Duration of connecting and querying peers", stats.UnitMilliseconds)
	PeersToCrawlCount         = stats.Float64("peers_to_crawl_count", "Number of peers in the queue to crawl", stats.UnitDimensionless)
	PeersToConnectCount       = stats.Float64("peers_to_connect_count", "Number of peers in the queue to connect", stats.UnitDimensionless)
)

// Views
var (
	CrawlConnectDurationView = &view.View{
		Measure:     CrawlConnectDuration,
		Aggregation: defaultMillisecondsDistribution,
	}
	MonitorConnectDurationView = &view.View{
		Measure:     MonitorConnectDuration,
		Aggregation: defaultMillisecondsDistribution,
	}
	CrawlConnectsCountView = &view.View{
		Measure:     CrawlConnectsCount,
		Aggregation: view.Count(),
	}
	MonitorConnectsCountView = &view.View{
		Measure:     MonitorConnectsCount,
		Aggregation: view.Count(),
	}
	CrawlConnectErrorsCountView = &view.View{
		Measure:     CrawlConnectErrorsCount,
		Aggregation: view.Count(),
	}
	MonitorConnectErrorsCountView = &view.View{
		Measure:     MonitorConnectErrorsCount,
		Aggregation: view.Count(),
	}
	FetchNeighborsDurationView = &view.View{
		Measure:     FetchNeighborsDuration,
		TagKeys:     []tag.Key{KeyAgentVersion},
		Aggregation: defaultMillisecondsDistribution,
	}
	FetchedNeighborsCountView = &view.View{
		Measure:     FetchedNeighborsCount,
		TagKeys:     []tag.Key{KeyAgentVersion},
		Aggregation: fibonacciDistribution,
	}
	CrawledPeersCountView = &view.View{
		Measure:     CrawledPeersCount,
		Aggregation: fibonacciDistribution,
	}
	PeerCrawlDurationView = &view.View{
		Measure:     PeerCrawlDuration,
		Aggregation: defaultMillisecondsDistribution,
	}
	CrawledUpsertDurationView = &view.View{
		Measure:     CrawledUpsertDuration,
		Aggregation: ocsql.DefaultMillisecondsDistribution,
	}
	PeersToCrawlCountView = &view.View{
		Measure:     PeersToCrawlCount,
		Aggregation: view.LastValue(),
	}
	PeersToConnectCountView = &view.View{
		Measure:     PeersToConnectCount,
		Aggregation: view.LastValue(),
	}
)

// DefaultCrawlViews with all views in it.
var DefaultCrawlViews = []*view.View{
	CrawlConnectDurationView,
	CrawlConnectsCountView,
	CrawlConnectErrorsCountView,
	FetchNeighborsDurationView,
	FetchedNeighborsCountView,
	CrawledPeersCountView,
	PeerCrawlDurationView,
	PeersToCrawlCountView,
	CrawledUpsertDurationView,
}

// DefaultMonitorViews with all views in it.
var DefaultMonitorViews = []*view.View{
	MonitorConnectDurationView,
	PeersToConnectCountView,
	MonitorConnectsCountView,
	MonitorConnectErrorsCountView,
}
