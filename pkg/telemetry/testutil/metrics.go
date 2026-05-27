package testutil

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func CounterValue(c prometheus.Counter) float64 {
	var m dto.Metric
	_ = c.(prometheus.Metric).Write(&m)
	return m.GetCounter().GetValue()
}

func CounterVecValue(c *prometheus.CounterVec, labels ...string) float64 {
	var m dto.Metric
	_ = c.WithLabelValues(labels...).(prometheus.Metric).Write(&m)
	return m.GetCounter().GetValue()
}

func GaugeValue(g prometheus.Gauge) float64 {
	var m dto.Metric
	_ = g.(prometheus.Metric).Write(&m)
	return m.GetGauge().GetValue()
}

func GaugeVecValue(g *prometheus.GaugeVec, labels ...string) float64 {
	var m dto.Metric
	_ = g.WithLabelValues(labels...).(prometheus.Metric).Write(&m)
	return m.GetGauge().GetValue()
}

func HistogramVecCount(h *prometheus.HistogramVec, labels ...string) uint64 {
	var m dto.Metric
	_ = h.WithLabelValues(labels...).(prometheus.Metric).Write(&m)
	return m.GetHistogram().GetSampleCount()
}
