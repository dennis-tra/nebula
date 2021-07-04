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

func RegisterListenAndServe(host string, port int, cmd string) error {
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "nebula",
	})
	if err != nil {
		return errors.Wrap(err, "new prometheus exporter")
	}

	// Register kademlia metrics
	if err = view.Register(kadmetrics.DefaultViews...); err != nil {
		return errors.Wrap(err, "register kademlia default views")
	}

	// Register nebula views TODO: don't use cmd string parameter here
	if cmd == "crawl" {
		if err = view.Register(DefaultCrawlViews...); err != nil {
			return errors.Wrap(err, "register nebula default crawl views")
		}
	} else if cmd == "monitor" {
		if err = view.Register(DefaultMonitorViews...); err != nil {
			return errors.Wrap(err, "register nebula default monitor views")
		}
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
	ConnectDuration             = stats.Float64("connect_duration", "Duration of connecting to peers", stats.UnitMilliseconds)
	ConnectsCount               = stats.Float64("connects_count", "Number of connection establishment attempts", stats.UnitDimensionless)
	ConnectErrorsCount          = stats.Float64("connect_errors_count", "Number of successful connection establishment errors", stats.UnitDimensionless)
	WorkersWorkingCount         = stats.Float64("workers_working_count", "Number of workers that are currently working", stats.UnitDimensionless)
	FetchNeighborsDuration      = stats.Float64("fetch_neighbors_duration", "Duration of crawling a peer for all neighbours in its buckets", stats.UnitMilliseconds)
	FetchedNeighborsCount       = stats.Float64("fetched_neighbors_count", "Number of neighbors fetched from a peer", stats.UnitDimensionless)
	CrawledPeersCount           = stats.Float64("crawled_peers_count", "Number of distinct peers found for a peer crawl", stats.UnitDimensionless)
	CrawledUpsertDuration       = stats.Float64("crawled_upsert_duration", "Amount of time we need to populate the database with one crawl result", stats.UnitMilliseconds)
	PeerCrawlDuration           = stats.Float64("peer_crawl_duration", "Duration of connecting and querying peers", stats.UnitMilliseconds)
	PeerPingDuration            = stats.Float64("peer_ping_duration", "Duration of pinging peers", stats.UnitMilliseconds)
	PeersToCrawlCount           = stats.Float64("peers_to_crawl_count", "Number of peers in the queue to crawl", stats.UnitDimensionless)
	PeersToPingCount            = stats.Float64("peers_to_ping_count", "Number of peers in the queue to ping", stats.UnitDimensionless)
	PingBlockedDialSuccessCount = stats.Float64("ping_blocked_dial_success_count", "Number of instances where a ping failed but a dial succeeded", stats.UnitDimensionless)
)

// Views
var (
	ConnectDurationView = &view.View{
		Measure:     ConnectDuration,
		Aggregation: defaultMillisecondsDistribution,
	}
	ConnectsCountView = &view.View{
		Measure:     ConnectsCount,
		Aggregation: view.Count(),
	}
	ConnectErrorsCountView = &view.View{
		Measure:     ConnectErrorsCount,
		Aggregation: view.Count(),
	}
	WorkersWorkingCountView = &view.View{
		Measure:     WorkersWorkingCount,
		Aggregation: view.LastValue(),
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
	PeerPingDurationView = &view.View{
		Measure:     PeerPingDuration,
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
	PeersToPingCountView = &view.View{
		Measure:     PeersToPingCount,
		Aggregation: view.LastValue(),
	}
	PingBlockedDialSuccessCountView = &view.View{
		Measure:     PingBlockedDialSuccessCount,
		Aggregation: view.Count(),
	}
)

// DefaultCrawlViews with all views in it.
var DefaultCrawlViews = []*view.View{
	ConnectDurationView,
	ConnectsCountView,
	ConnectErrorsCountView,
	WorkersWorkingCountView,
	FetchNeighborsDurationView,
	FetchedNeighborsCountView,
	CrawledPeersCountView,
	PeerCrawlDurationView,
	PeersToCrawlCountView,
	CrawledUpsertDurationView,
}

// DefaultMonitorViews with all views in it.
var DefaultMonitorViews = []*view.View{
	PeerPingDurationView,
	PeersToPingCountView,
	PingBlockedDialSuccessCountView,
}
