package testutil_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/telemetry/testutil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestMetricReaders(t *testing.T) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{Name: "c_total", Help: "c"})
	counterVec := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cv_total", Help: "cv"}, []string{"l"})
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{Name: "g", Help: "g"})
	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "gv", Help: "gv"}, []string{"l"})
	histogramVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "h", Help: "h"}, []string{"l"})

	counter.Add(2)
	counterVec.WithLabelValues("a").Inc()
	gauge.Set(3)
	gaugeVec.WithLabelValues("a").Set(4)
	histogramVec.WithLabelValues("a").Observe(1)

	require.Equal(t, 2.0, testutil.CounterValue(counter))
	require.Equal(t, 1.0, testutil.CounterVecValue(counterVec, "a"))
	require.Equal(t, 3.0, testutil.GaugeValue(gauge))
	require.Equal(t, 4.0, testutil.GaugeVecValue(gaugeVec, "a"))
	require.Equal(t, uint64(1), testutil.HistogramVecCount(histogramVec, "a"))
}
