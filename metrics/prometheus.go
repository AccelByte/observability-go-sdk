// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registerer = prometheus.DefaultRegisterer
	gatherer   = prometheus.DefaultGatherer
)

// PrometheusProviderOpts represents the Prometheus metrics configuration options.
type PrometheusProviderOpts struct {
	prometheus.Registerer
	prometheus.Gatherer
	DisableGoCollector      bool // default is false = go collector is enabled
	DisableProcessCollector bool // default is false = process collector is enabled
}

// NewPrometheusProvider creates a new Prometheus provider that implements Provider using Prometheus metrics.
func NewPrometheusProvider(opts PrometheusProviderOpts) PrometheusProvider {
	if opts.Registerer != nil {
		registerer = opts.Registerer
	}
	if opts.Gatherer != nil {
		gatherer = opts.Gatherer
	}
	p := PrometheusProvider{
		registerer: registerer,
		gatherer:   gatherer,
	}

	if opts.DisableProcessCollector {
		prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}
	if opts.DisableGoCollector {
		prometheus.Unregister(collectors.NewGoCollector())
	}

	return p
}

// PrometheusProvider represents the implementation for Prometheus provider.
type PrometheusProvider struct {
	registerer prometheus.Registerer
	gatherer   prometheus.Gatherer
}

// NewCounter creates a new Prometheus counter vector metric.
func (p PrometheusProvider) NewCounter(name, help string, labels ...string) CounterVecMetric {
	vec := promauto.With(p.registerer).NewCounterVec(
		prometheus.CounterOpts{
			Name: sanitizeName(name),
			Help: help,
		},
		labels,
	)
	return counterVec{vec}
}

// counterVec represents an internal counter vec type that implements CounterVecMetric
type counterVec struct {
	*prometheus.CounterVec
}

func (c counterVec) With(labels map[string]string) CounterMetric {
	return c.CounterVec.With(labels)
}

// NewGauge creates a new Prometheus gauge vector metric.
func (p PrometheusProvider) NewGauge(name, help string, labels ...string) GaugeVecMetric {
	vec := promauto.With(p.registerer).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: sanitizeName(name),
			Help: help,
		},
		labels,
	)
	return gaugeVec{vec}
}

// gaugeVec represents an internal gauge vec type that implements GaugeVecMetric
type gaugeVec struct {
	*prometheus.GaugeVec
}

func (g gaugeVec) With(labels map[string]string) GaugeMetric {
	return g.GaugeVec.With(labels)
}

// NewHistogram creates a new Prometheus histogram vector metric.
func (p PrometheusProvider) NewHistogram(name, help string, buckets []float64, labels ...string) ObserverVecMetric {
	if len(buckets) <= 0 {
		buckets = prometheus.DefBuckets
	}
	vec := promauto.With(p.registerer).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    sanitizeName(name),
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)
	return histogramVec{vec}
}

// histogramVec represents an internal histogram vec type that implements ObserverVecMetric
type histogramVec struct {
	*prometheus.HistogramVec
}

func (h histogramVec) With(labels map[string]string) ObserverMetric {
	return h.HistogramVec.With(labels)
}

// NewSummary creates a new Prometheus summary vector metric.
func (p PrometheusProvider) NewSummary(name, help string, labels ...string) ObserverVecMetric {
	vec := promauto.With(p.registerer).NewSummaryVec(
		prometheus.SummaryOpts{
			Name: sanitizeName(name),
			Help: help,
		},
		labels,
	)
	return summaryVec{vec}
}

// initBuildInfo initializes one gauge metric with constant 1
func (p PrometheusProvider) initBuildInfo(buildInfo BuildInfo) {
	buildInfoGauge := promauto.With(p.registerer).NewGauge(
		prometheus.GaugeOpts{
			Name: sanitizeName(generateMetricsName(serviceName, "build_info")),
			Help: "A metric with a constant '1' value labeled by version, revision, branch, and goversion from which the service was built",
			ConstLabels: prometheus.Labels{
				"revisionID":         buildInfo.RevisionID,
				"buildDate":          buildInfo.BuildDate,
				"version":            buildInfo.Version,
				"gitHash":            buildInfo.GitHash,
				"roleSeedingVersion": buildInfo.RoleSeedingVersion,
			},
		})
	buildInfoGauge.Set(1)
}

// summaryVec represents an internal summary vec type that implements ObserverVecMetric
type summaryVec struct {
	*prometheus.SummaryVec
}

func (s summaryVec) With(labels map[string]string) ObserverMetric {
	return s.SummaryVec.With(labels)
}

// ServePrometheus exposes Prometheus over HTTP on the given address and metrics endpoint.
// If you plan on exposing the metrics on an already existing HTTP server, use the PrometheusHandler instead.
func ServePrometheus(addr, endpoint string) error {
	mux := http.NewServeMux()
	mux.Handle(endpoint, PrometheusHandler())
	return http.ListenAndServe(addr, mux)
}

// PrometheusHandler creates a new http.Handler that exposes Prometheus metrics over HTTP.
func PrometheusHandler() http.Handler {
	return promhttp.InstrumentMetricHandler(
		registerer,
		promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}),
	)
}

func sanitizeName(name string) string {
	output := []rune(name)

	for i, b := range name {
		if !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || b == ':' || (b >= '0' && b <= '9' && i > 0)) {
			output[i] = '_'
		}
	}

	return string(output)
}
