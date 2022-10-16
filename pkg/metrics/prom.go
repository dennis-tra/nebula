package metrics

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	CrawlLabel = prometheus.Labels{"type": "crawl"}
	DialLabel  = prometheus.Labels{"type": "dial"}
)

var (
	VisitCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "visit_count",
		Namespace: "nebula",
		Help:      "The number of visits",
	}, []string{"type"})
	VisitErrorsCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "visit_error_count",
		Namespace: "nebula",
		Help:      "The number of visits that yielded an error",
	}, []string{"type"})
	FetchedNeighborsCount = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:      "fetched_neighbors_count",
		Namespace: "nebula",
		Help:      "Number of neighbors fetched from a peer",
		Buckets:   []float64{100, 150, 200, 210, 220, 230, 240, 250, 260, 270, 280, 290, 300},
	})
	DistinctVisitedPeersCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name:      "distinct_visited_peers_count",
		Namespace: "nebula",
		Help:      "Number of distinct peers found for a peer crawl",
	})
	VisitQueueLength = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "visit_queue_length",
		Namespace: "nebula",
		Help:      "Number of peers in the queue to visit",
	}, []string{"type"})
	CacheQueriesCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "cache_queries",
		Namespace: "nebula",
		Help:      "The number of cache queries",
	}, []string{"entity", "outcome"})
	InsertVisitHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "insert_visit_query_duration_seconds",
		Help: "Histogram of database query times for visit insertions",
	}, []string{"type", "success"})
)

func ListenAndServe(host string, port int) {
	if err := prometheus.Register(InsertVisitHistogram); err != nil {
		log.WithError(err).Warnln("Error registering histogram")
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	log.WithField("addr", addr).Debugln("Starting prometheus endpoint")
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.WithError(err).Warnln("Error serving prometheus")
	}
}
