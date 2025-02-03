package tele

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/prometheus-community/ecs_exporter/ecscollector"
	"github.com/prometheus-community/ecs_exporter/ecsmetadata"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	promexp "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmeter "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.19.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/atomic"
)

const (
	MeterName  = "github.com/dennis-tra/nebula-crawler"
	TracerName = "github.com/dennis-tra/nebula-crawler"
)

// HealthStatus is a global variable that indicates the health of the current
// process. This could either be "hey, I've started crawling and everything
// works fine" or "hey, I'm monitoring the network and all good".
var HealthStatus = atomic.NewBool(false)

// NewMeterProvider initializes a new opentelemetry meter provider that exports
// metrics using prometheus. To serve the prometheus endpoint call
// [ListenAndServe] down below.
func NewMeterProvider() (metric.MeterProvider, error) {
	// initialize AWS Elastic Container Service collector and register it with
	// the default prometheus registry. If we are not running in an ECS
	// environment, don't do anything.
	client, err := ecsmetadata.NewClientFromEnvironment()
	if err == nil {
		log.Infoln("Registering ECS collector")
		collector := ecscollector.NewCollector(client, slog.Default())
		if err := prometheus.DefaultRegisterer.Register(collector); err != nil {
			return nil, fmt.Errorf("register collector: %w", err)
		}
	}

	exporter, err := promexp.New(promexp.WithNamespace("nebula"))
	if err != nil {
		return nil, fmt.Errorf("new prometheus exporter: %w", err)
	}

	return sdkmeter.NewMeterProvider(sdkmeter.WithReader(exporter)), nil
}

// NewTracerProvider initializes a new tracer provider to send traces to the
// given host and port combination. If any of the two values is the zero value
// this function will return a no-op tracer provider which effectively disables
// tracing.
func NewTracerProvider(ctx context.Context, host string, port int) (trace.TracerProvider, error) {
	if host == "" || port == 0 {
		return noop.NewTracerProvider(), nil
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%d", host, port)),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("new otel trace exporter: %w", err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("nebula"),
		)),
	), nil
}

// ListenAndServe starts an HTTP server and exposes prometheus and pprof
// metrics. It also exposes a health endpoint that can be probed with
// `nebula health`.
func ListenAndServe(host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)
	log.WithField("addr", addr).Debugln("Starting telemetry endpoint")

	// profile 1% of contention events
	runtime.SetMutexProfileFraction(1)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(rw http.ResponseWriter, req *http.Request) {
		log.Debugln("Responding to health check")
		if HealthStatus.Load() {
			rw.WriteHeader(http.StatusOK)
		} else {
			rw.WriteHeader(http.StatusServiceUnavailable)
		}
	})

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.WithError(err).Warnln("Error serving prometheus")
	}
}
