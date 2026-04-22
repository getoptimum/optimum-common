package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	labeledRegistry prometheus.Registerer
	namespace       = "optimum"
)

type metricOptions struct {
	constLabels prometheus.Labels
}

// MetricOption customizes a metric created by the telemetry helpers.
type MetricOption func(*metricOptions)

// WithConstLabels attaches fixed labels to every series emitted by the metric.
func WithConstLabels(labels prometheus.Labels) MetricOption {
	return func(opts *metricOptions) {
		opts.constLabels = labels
	}
}

func newMetricOptions(opts ...MetricOption) metricOptions {
	var out metricOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&out)
		}
	}

	return out
}

func baseOpts(name, subsystem, help string, opts ...MetricOption) prometheus.Opts {
	metricOpts := newMetricOptions(opts...)

	return prometheus.Opts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        name,
		Help:        help,
		ConstLabels: metricOpts.constLabels,
	}
}

func counterOpts(name, subsystem, help string, opts ...MetricOption) prometheus.CounterOpts {
	return prometheus.CounterOpts(baseOpts(name, subsystem, help, opts...))
}

func gaugeOpts(name, subsystem, help string, opts ...MetricOption) prometheus.GaugeOpts {
	return prometheus.GaugeOpts(baseOpts(name, subsystem, help, opts...))
}

func histogramOpts(
	name, subsystem, help string,
	buckets []float64,
	opts ...MetricOption,
) prometheus.HistogramOpts {
	base := baseOpts(name, subsystem, help, opts...)

	return prometheus.HistogramOpts{
		Namespace:   base.Namespace,
		Subsystem:   base.Subsystem,
		Name:        base.Name,
		Help:        base.Help,
		ConstLabels: base.ConstLabels,
		Buckets:     buckets,
	}
}

func SetLabeledRegistry(lr prometheus.Registerer, ns string) {
	labeledRegistry = lr
	namespace = ns
}

// NewHistogram creates and registers a new Prometheus histogram metric with the given parameters.
func NewHistogram(name, subsystem, help string, labels []string, opts ...MetricOption) *prometheus.HistogramVec {
	return promauto.With(labeledRegistry).NewHistogramVec(
		histogramOpts(name, subsystem, help, nil, opts...),
		labels,
	)
}

// NewHistogramWithBuckets creates a Histogram metrics with custom buckets.
func NewHistogramWithBuckets(
	name, subsystem, help string,
	labels []string,
	buckets []float64,
	opts ...MetricOption,
) *prometheus.HistogramVec {
	return promauto.With(labeledRegistry).NewHistogramVec(
		histogramOpts(name, subsystem, help, buckets, opts...),
		labels,
	)
}

// NewCounterVec creates a CounterVec metrics under the global namespace returns nop if metrics are disabled.
func NewCounterVec(name, subsystem, help string, labels []string, opts ...MetricOption) *prometheus.CounterVec {
	return promauto.With(labeledRegistry).NewCounterVec(
		counterOpts(name, subsystem, help, opts...),
		labels,
	)
}

// NewCounter creates a Counter metrics under the global namespace returns nop if metrics are disabled.
func NewCounter(name, subsystem, help string, opts ...MetricOption) prometheus.Counter {
	return promauto.With(labeledRegistry).NewCounter(
		counterOpts(name, subsystem, help, opts...),
	)
}

// NewGaugeVec creates a Gauge metrics under the global namespace returns nop if metrics are disabled.
func NewGaugeVec(name, subsystem, help string, labels []string, opts ...MetricOption) *prometheus.GaugeVec {
	return promauto.With(labeledRegistry).NewGaugeVec(
		gaugeOpts(name, subsystem, help, opts...),
		labels,
	)
}

// NewGauge creates a Gauge metrics under the global namespace returns nop if metrics are disabled.
func NewGauge(name, subsystem, help string, opts ...MetricOption) prometheus.Gauge {
	return promauto.With(labeledRegistry).NewGauge(
		gaugeOpts(name, subsystem, help, opts...),
	)
}

// NewSimpleHistogram returns a histogram without labels.
func NewSimpleHistogram(name, subsystem, help string, buckets []float64, opts ...MetricOption) prometheus.Histogram {
	return promauto.With(labeledRegistry).NewHistogram(
		histogramOpts(name, subsystem, help, buckets, opts...),
	)
}
