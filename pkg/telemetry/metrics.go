package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	labeledRegistry prometheus.Registerer
	namespace       = "optimum"
)

func SetLabeledRegistry(lr prometheus.Registerer, ns string) {
	labeledRegistry = lr
	namespace = ns
}

// NewCounter creates and registers a new Prometheus counter metric with the given parameters.
func NewCounter(name, subsystem, help string, labels []string) *prometheus.CounterVec {
	return promauto.With(labeledRegistry).NewCounterVec(
		prometheus.CounterOpts{Namespace: namespace, Subsystem: subsystem, Name: name, Help: help},
		labels,
	)
}

// NewGauge creates and registers a new Prometheus gauge metric with the given parameters.
func NewGauge(name, subsystem, help string, labels []string) *prometheus.GaugeVec {
	return promauto.With(labeledRegistry).NewGaugeVec(
		prometheus.GaugeOpts{Namespace: namespace, Subsystem: subsystem, Name: name, Help: help},
		labels,
	)
}

// NewHistogram creates and registers a new Prometheus histogram metric with the given parameters.
func NewHistogram(name, subsystem, help string, labels []string) *prometheus.HistogramVec {
	return promauto.With(labeledRegistry).NewHistogramVec(
		prometheus.HistogramOpts{Namespace: namespace, Subsystem: subsystem, Name: name, Help: help},
		labels,
	)
}

// NewCounterVec creates a CounterVec metrics under the global namespace returns nop if metrics are disabled.
func NewCounterVec(name, subsystem, help string, labels []string) *prometheus.CounterVec {
	return promauto.With(labeledRegistry).NewCounterVec(
		prometheus.CounterOpts{Namespace: namespace, Subsystem: subsystem, Name: name, Help: help},
		labels,
	)
}

// NewGaugeVec creates a Gauge metrics under the global namespace returns nop if metrics are disabled.
func NewGaugeVec(name, subsystem, help string, labels []string) *prometheus.GaugeVec {
	return promauto.With(labeledRegistry).NewGaugeVec(
		prometheus.GaugeOpts{Namespace: namespace, Subsystem: subsystem, Name: name, Help: help},
		labels,
	)
}

// NewSimpleHistogram returns a histogram without labels.
func NewSimpleHistogram(name, subsystem, help string, buckets []float64) prometheus.Histogram {
	return promauto.With(labeledRegistry).NewHistogram(
		prometheus.HistogramOpts{Namespace: namespace, Subsystem: subsystem, Name: name, Help: help, Buckets: buckets},
	)
}
