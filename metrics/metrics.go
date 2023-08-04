// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package metrics

import (
	"fmt"
	"runtime/metrics"
)

var (
	defaultProvider        Provider = NewPrometheusProvider(PrometheusProviderOpts{})
	serviceName            string
	namespacePathParameter string
	enableRuntimeMetrics   bool

	httpMetrics ObserverVecMetric
)

// CounterVecMetric represents a vector counter metric containing a variation
// of the same metric under different labels.
type CounterVecMetric interface {
	With(labels map[string]string) CounterMetric
}

// CounterMetric represents a counter metric.
type CounterMetric interface {
	Inc()
	Add(float64)
}

// GaugeVecMetric represents a vector gauge metric containing a variation
// of the same metric under different labels.
type GaugeVecMetric interface {
	With(labels map[string]string) GaugeMetric
}

// GaugeMetric represents a single numerical value that can go up and down.
type GaugeMetric interface {
	Set(float64)
	Inc()
	Dec()
	Add(float64)
	Sub(float64)
	SetToCurrentTime()
}

// ObserverVecMetric represents a vector observer(histogram/summary) metric containing a variation
// of the same metric under different labels.
type ObserverVecMetric interface {
	With(labels map[string]string) ObserverMetric
}

// ObserverMetric represents a Histogram / Summary metric.
type ObserverMetric interface {
	Observe(float64)
}

// Provider represents a metric provider, i.e: Prometheus.
type Provider interface {
	NewCounter(name, help string, labels ...string) CounterVecMetric
	NewGauge(name, help string, labels ...string) GaugeVecMetric
	NewHistogram(name, help string, buckets []float64, labels ...string) ObserverVecMetric
	NewSummary(name, help string, labels ...string) ObserverVecMetric
}

type Opts struct {
	Namespace            string
	EnableRuntimeMetrics bool
	EnableHTTPMetrics    bool

	HTTPMetrics *ObserverVecMetric
}

func Initialize(s string, option *Opts) {
	serviceName = s

	initializeDefaultOption()

	if option != nil {
		overrideDefaultOption(option)
	}

	if enableRuntimeMetrics {
		startRuntimeMetrics()
	}
}

func initializeDefaultOption() {
	namespacePathParameter = defaultNamespacePathParameter

	httpMetrics = HistogramVecWithBuckets(
		generateMetricsName(metricsNameHTTP),
		"HTTP request in histogram",
		[]float64{0.25, 0.5, 1, 1.25, 1.5, 1.75, 2, 2.25, 2.5, 2.75, 3},
		[]string{labelNamespace, labelPath, labelMethod, labelResponseCode},
	)
}

func overrideDefaultOption(option *Opts) {
	if option.Namespace != "" {
		namespacePathParameter = option.Namespace
	}
	if !option.EnableRuntimeMetrics {
		enableRuntimeMetrics = false
	}
	if option.HTTPMetrics != nil {
		httpMetrics = *option.HTTPMetrics
	}
}

func startRuntimeMetrics() {
	runtimeMetricsGaugeMap = make(map[string]GaugeMetric)
	runtimeMetricsHistogram = make(map[string]ObserverMetric)

	descs := metrics.All()
	for _, desc := range descs {
		switch desc.Kind {
		case metrics.KindUint64, metrics.KindFloat64:
			runtimeMetricsGaugeMap[desc.Name] = Gauge(generateMetricsName(desc.Name), desc.Description)
		case metrics.KindFloat64Histogram:
			runtimeMetricsHistogram[desc.Name] = Histogram(generateMetricsName(desc.Name), desc.Description)
		}
	}

	go sendRuntimeMetrics()
}

func generateMetricsName(metricsName string) string {
	return fmt.Sprintf(metricsNameFormat, serviceName, metricsName)
}

// SetProvider allow setting/replacing the default (Prometheus) metrics provider with a new one.
func SetProvider(p Provider) {
	defaultProvider = p
}

// Counter creates a counter metric with default provider.
// Use this function, if the metric does not have any custom dynamic labels,
// which also gives the caller direct access to a CounterMetric.
func Counter(name string, help string) CounterMetric {
	return defaultProvider.NewCounter(name, help).With(map[string]string{})
}

// CounterVec creates a counter vector metric with default provider.
// Use this function instead, if you plan on dynamically adding custom labels
// to the CounterMetric, which involves an extra step of calling
// .With(map[string]string{"label_name": "label_value"}), which then
// gives the caller access to a CounterMetric to work with.
func CounterVec(name string, help string, labels []string) CounterVecMetric {
	return defaultProvider.NewCounter(name, help, labels...)
}

// Gauge creates a gauge metric with default provider.
// Use this function, if the metric does not have any custom dynamic labels,
// which also gives the caller direct access to a GaugeMetric.
func Gauge(name string, help string) GaugeMetric {
	return defaultProvider.NewGauge(name, help).With(map[string]string{})

}

// GaugeVec creates a gauge vector metric with default provider.
// Use this function instead, if you plan on dynamically adding custom labels
// to the GaugeMetric, which involves an extra step of calling
// .With(map[string]string{"label_name": "label_value"}), which then
// gives the caller access to a GaugeMetric to work with.
func GaugeVec(name string, help string, labels []string) GaugeVecMetric {
	return defaultProvider.NewGauge(name, help, labels...)
}

// Histogram creates a histogram metric with default provider.
// Use this function, if the metric does not have any custom dynamic labels,
// which also gives the caller direct access to a ObserverMetric (histogram).
// This will use the default buckets for a histogram:
// []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
func Histogram(name string, help string) ObserverMetric {
	return defaultProvider.NewHistogram(name, help, []float64{}).With(map[string]string{})
}

// HistogramWithBuckets creates a histogram metric with default provider.
// User this function if the metric does not have any custom dynamic labels,
// but you want to specify custom buckets other than the default.
func HistogramWithBuckets(name, help string, buckets []float64) ObserverMetric {
	return defaultProvider.NewHistogram(name, help, buckets, []string{}...).With(map[string]string{})
}

// HistogramVec creates a histogram vector metric with default provider.
// Use this function instead, if you plan on dynamically adding custom labels
// to the ObserverMetric (histogram), which involves an extra step of calling
// .With(map[string]string{"label_name": "label_value"}), which then
// gives the caller access to a ObserverMetric (histogram) to work with.
// This will use the default buckets for a histogram:
// []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
func HistogramVec(name string, help string, labels []string) ObserverVecMetric {
	return defaultProvider.NewHistogram(name, help, []float64{}, labels...)
}

// HistogramVecWithBuckets creates a histogram vector metric with default provider.
// Use this function to create a ObserverMetric (histogram) that with custom labels
// AND you want to specify custom buckets to overwrite the default. Similar to
// HistogramVec, you will need the extra step of calling the object with
// .With(map[string]string{"label_name": "label_value"})
func HistogramVecWithBuckets(name, help string, buckets []float64, labels []string) ObserverVecMetric {
	return defaultProvider.NewHistogram(name, help, buckets, labels...)
}

// Summary creates a summary metric with default provider.
// Use this function, if the metric does not have any custom dynamic labels,
// which also gives the caller direct access to a ObserverMetric (summary).
func Summary(name string, help string) ObserverMetric {
	return defaultProvider.NewSummary(name, help).With(map[string]string{})
}

// SummaryVec creates a summary vector metric with default provider.
// Use this function instead, if you plan on dynamically adding custom labels
// to the ObserverMetric (summary), which involves an extra step of calling
// .With(map[string]string{"label_name": "label_value"}), which then
// gives the caller access to a ObserverMetric (summary) to work with.
func SummaryVec(name string, help string, labels []string) ObserverVecMetric {
	return defaultProvider.NewSummary(name, help, labels...)
}
