package metrics

import (
	"fmt"
	"net/http"

	"contrib.go.opencensus.io/exporter/prometheus"
	kadmetrics "github.com/libp2p/go-libp2p-kad-dht/metrics"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

func RegisterListenAndServe(host string, port int) error {
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "nebula",
	})
	if err != nil {
		return errors.Wrap(err, "new prometheus exporter")
	}

	// Register the views
	if err = view.Register(kadmetrics.DefaultViews...); err != nil {
		return errors.Wrap(err, "register kademlia default views")
	}

	// Register the views
	if err = view.Register(DefaultViews...); err != nil {
		return errors.Wrap(err, "register nebula default views")
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", pe)
		if err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), mux); err != nil {
			log.Fatalf("Failed to run Prometheus /metrics endpoint: %v", err)
		}
	}()
	return nil
}

//func RegisterDB(db *db.Client) error {
//	return db.Use(gormprom.New(gormprom.Config{
//		DBName:          "nebula_db", // use `DBName` as metrics label
//		RefreshInterval: 15,          // Refresh metrics interval (default 15 seconds)
//		StartServer:     true,        // start http server to expose metrics
//		HTTPServerPort:  6667,        // configure http server port, default port 8080 (if you have configured multiple instances, only the first `HTTPServerPort` will be used to start server)
//		//MetricsCollector: []prometheus.MetricsCollector{
//		//	&prometheus.DBStats{
//		//		VariableNames: []string{"Threads_running"},
//		//	},
//		//}, // user defined metrics
//	}))
//}

// Keys
var (
	KeyAgentVersion, _ = tag.NewKey("agent_version")
)

// UpsertAgentVersion .
func UpsertAgentVersion(av string) tag.Mutator {
	return tag.Upsert(KeyAgentVersion, av)
}

// Distributions
var (
	defaultBytesDistribution        = view.Distribution(1024, 2048, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864, 268435456, 1073741824, 4294967296)
	fibonacciDistribution           = view.Distribution(1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 377)
	defaultMillisecondsDistribution = view.Distribution(0.01, 0.05, 0.1, 0.3, 0.6, 0.8, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000)
)

// Measures
var (
	ConnectDuration        = stats.Float64("connect_duration", "Duration of connecting to peers", stats.UnitMilliseconds)
	ConnectsCount          = stats.Float64("connects_count", "Number of connection establishment attempts", stats.UnitDimensionless)
	ConnectErrors          = stats.Float64("connect_errors_count", "Number of successful connection establishment errors", stats.UnitDimensionless)
	WorkersWorkingCount    = stats.Float64("workers_working_count", "Number of workers that are currently working", stats.UnitDimensionless)
	FetchNeighborsDuration = stats.Float64("fetch_neighbors_duration", "Duration of crawling a peer for all neighbours in its buckets", stats.UnitMilliseconds)
	FetchedNeighborsCount  = stats.Float64("fetched_neighbors_count", "Number of neighbors fetched from a peer", stats.UnitDimensionless)
	CrawledPeersCount      = stats.Float64("crawled_peers_count", "Number of distinct peers found for a peer crawl", stats.UnitDimensionless)
	PeerCrawlDuration      = stats.Float64("peer_crawl_duration", "Duration of connecting and querying peers", stats.UnitMilliseconds)
	PeersToCrawlCount      = stats.Float64("peers_to_crawl_count", "Number of peers in the queue to crawl", stats.UnitDimensionless)
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
	ConnectErrorsView = &view.View{
		Measure:     ConnectErrors,
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
		TagKeys:     []tag.Key{KeyAgentVersion},
		Aggregation: fibonacciDistribution,
	}
	PeerCrawlDurationView = &view.View{
		Measure:     PeerCrawlDuration,
		TagKeys:     []tag.Key{KeyAgentVersion},
		Aggregation: defaultMillisecondsDistribution,
	}
	PeersToCrawlCountView = &view.View{
		Measure:     PeersToCrawlCount,
		Aggregation: view.LastValue(),
	}
)

// DefaultViews with all views in it.
var DefaultViews = []*view.View{
	ConnectDurationView,
	ConnectsCountView,
	ConnectErrorsView,
	WorkersWorkingCountView,
	FetchNeighborsDurationView,
	FetchedNeighborsCountView,
	CrawledPeersCountView,
	PeerCrawlDurationView,
	PeersToCrawlCountView,
}
