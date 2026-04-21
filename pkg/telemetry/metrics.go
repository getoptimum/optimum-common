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

func SetLabeledRegistry(lr prometheus.Registerer, ns string) {
	labeledRegistry = lr
	namespace = ns
}

// NewHistogram creates and registers a new Prometheus histogram metric with the given parameters.
func NewHistogram(name, subsystem, help string, labels []string, opts ...MetricOption) *prometheus.HistogramVec {
	metricOpts := newMetricOptions(opts...)

	return promauto.With(labeledRegistry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        name,
			Help:        help,
			ConstLabels: metricOpts.constLabels,
		},
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
	metricOpts := newMetricOptions(opts...)

	return promauto.With(labeledRegistry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        name,
			Help:        help,
			ConstLabels: metricOpts.constLabels,
			Buckets:     buckets,
		},
		labels,
	)
}

// NewCounterVec creates a CounterVec metrics under the global namespace returns nop if metrics are disabled.
func NewCounterVec(name, subsystem, help string, labels []string, opts ...MetricOption) *prometheus.CounterVec {
	metricOpts := newMetricOptions(opts...)

	return promauto.With(labeledRegistry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        name,
			Help:        help,
			ConstLabels: metricOpts.constLabels,
		},
		labels,
	)
}

// NewCounter creates a Counter metrics under the global namespace returns nop if metrics are disabled.
func NewCounter(name, subsystem, help string, opts ...MetricOption) prometheus.Counter {
	metricOpts := newMetricOptions(opts...)

	return promauto.With(labeledRegistry).NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        name,
			Help:        help,
			ConstLabels: metricOpts.constLabels,
		},
	)
}

// NewGaugeVec creates a Gauge metrics under the global namespace returns nop if metrics are disabled.
func NewGaugeVec(name, subsystem, help string, labels []string, opts ...MetricOption) *prometheus.GaugeVec {
	metricOpts := newMetricOptions(opts...)

	return promauto.With(labeledRegistry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        name,
			Help:        help,
			ConstLabels: metricOpts.constLabels,
		},
		labels,
	)
}

// NewGauge creates a Gauge metrics under the global namespace returns nop if metrics are disabled.
func NewGauge(name, subsystem, help string, opts ...MetricOption) prometheus.Gauge {
	metricOpts := newMetricOptions(opts...)

	return promauto.With(labeledRegistry).NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        name,
			Help:        help,
			ConstLabels: metricOpts.constLabels,
		},
	)
}

// NewSimpleHistogram returns a histogram without labels.
func NewSimpleHistogram(name, subsystem, help string, buckets []float64, opts ...MetricOption) prometheus.Histogram {
	metricOpts := newMetricOptions(opts...)

	return promauto.With(labeledRegistry).NewHistogram(
		prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        name,
			Help:        help,
			ConstLabels: metricOpts.constLabels,
			Buckets:     buckets,
		},
	)
}
