package telemetry_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	ioprometheusclient "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

// helpers
func fq(ns, subsystem, name string) string {
	switch {
	case ns != "" && subsystem != "":
		return ns + "_" + subsystem + "_" + name
	case ns != "":
		return ns + "_" + name
	case subsystem != "":
		return subsystem + "_" + name
	default:
		return name
	}
}

func getMF(t *testing.T, reg *prometheus.Registry, name string) *ioprometheusclient.MetricFamily {
	t.Helper()
	mfs, err := reg.Gather()
	require.NoError(t, err)
	for _, mf := range mfs {
		if mf.GetName() == name {
			return mf
		}
	}
	require.Failf(t, "metric family not found", "wanted %q; gathered %d families", name, len(mfs))
	t.FailNow()
	return nil
}

func singleMetric(t *testing.T, mf *ioprometheusclient.MetricFamily) *ioprometheusclient.Metric {
	t.Helper()
	require.Len(t, mf.Metric, 1)

	return mf.Metric[0]
}

func metricLabels(metric *ioprometheusclient.Metric) map[string]string {
	labels := map[string]string{}
	for _, label := range metric.GetLabel() {
		labels[label.GetName()] = label.GetValue()
	}

	return labels
}

func assertSingleMetricFamily(
	t *testing.T,
	reg *prometheus.Registry,
	subsystem, name string,
	metricType ioprometheusclient.MetricType,
	expectedLabels map[string]string,
) *ioprometheusclient.Metric {
	t.Helper()

	mf := getMF(t, reg, fq("constns", subsystem, name))
	require.Equal(t, metricType, mf.GetType())
	metric := singleMetric(t, mf)
	labels := metricLabels(metric)
	for key, value := range expectedLabels {
		require.Equal(t, value, labels[key])
	}

	return metric
}

func TestNewCounter_RegistersAndCollects(t *testing.T) {
	// given
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "customns")

	// when
	cv := telemetry.NewCounterVec("requests_total", "gateway", "number of requests", []string{"code"})
	cv.WithLabelValues("200").Add(2)
	cv.WithLabelValues("500").Inc()

	// then
	mf := getMF(t, reg, fq("customns", "gateway", "requests_total"))
	require.Equal(t, ioprometheusclient.MetricType_COUNTER, mf.GetType())

	var total float64
	seen := map[string]bool{"200": false, "500": false}
	for _, m := range mf.Metric {
		require.Len(t, m.Label, 1)
		require.Equal(t, "code", m.Label[0].GetName())
		seen[m.Label[0].GetValue()] = true
		total += m.GetCounter().GetValue()
	}
	require.True(t, seen["200"] && seen["500"], "expected both 200 and 500 label values")
	require.Equal(t, 3.0, total)
}

func TestNewGauge_RegistersAndCollects(t *testing.T) {
	// given
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "ns2")

	// when
	gv := telemetry.NewGaugeVec("inflight", "gateway", "inflight requests", []string{"route"})
	gv.WithLabelValues("/health").Set(7)
	gv.WithLabelValues("/api").Set(3)

	// then
	mf := getMF(t, reg, fq("ns2", "gateway", "inflight"))
	require.Equal(t, ioprometheusclient.MetricType_GAUGE, mf.GetType())

	got := map[string]float64{}
	for _, m := range mf.Metric {
		require.Len(t, m.Label, 1)
		require.Equal(t, "route", m.Label[0].GetName())
		got[m.Label[0].GetValue()] = m.GetGauge().GetValue()
	}
	require.Equal(t, 7.0, got["/health"])
	require.Equal(t, 3.0, got["/api"])
}

func TestNewCounter_RegistersAndIncrements(t *testing.T) {
	// given
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "ns_cnt")

	// when
	c := telemetry.NewCounter("processed_total", "worker", "number of processed items")
	c.Add(2)
	c.Inc() // now total should be 3

	// then
	mf := getMF(t, reg, fq("ns_cnt", "worker", "processed_total"))
	require.Equal(t, ioprometheusclient.MetricType_COUNTER, mf.GetType())

	require.Len(t, mf.Metric, 1, "expected exactly one metric (no labels)")
	require.NotNil(t, mf.Metric[0].Counter, "expected counter value to be present")
	require.Equal(t, 3.0, mf.Metric[0].GetCounter().GetValue())
}

func TestNewHistogramWithBucketsVec_RegistersAndObserves(t *testing.T) {
	// given
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "ns3")

	// when
	hv := telemetry.NewHistogramWithBuckets(
		"latency_seconds",
		"gateway",
		"request latency",
		[]string{"route"},
		prometheus.ExponentialBuckets(1, 2, 60),
	)
	hv.WithLabelValues("/api").Observe(0.05)
	hv.WithLabelValues("/api").Observe(0.15)
	hv.WithLabelValues("/health").Observe(0.01)

	// then
	mf := getMF(t, reg, fq("ns3", "gateway", "latency_seconds"))
	require.Equal(t, ioprometheusclient.MetricType_HISTOGRAM, mf.GetType())

	counts := map[string]uint64{}
	for _, m := range mf.Metric {
		var route string
		for _, l := range m.Label {
			if l.GetName() == "route" {
				route = l.GetValue()
			}
		}
		counts[route] = m.GetHistogram().GetSampleCount()
	}
	require.EqualValues(t, 2, counts["/api"])
	require.EqualValues(t, 1, counts["/health"])
}

func TestNewHistogramVec_RegistersAndObserves(t *testing.T) {
	// given
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "ns3")

	// when
	hv := telemetry.NewHistogram("latency_seconds", "gateway", "request latency", []string{"route"})
	hv.WithLabelValues("/api").Observe(0.05)
	hv.WithLabelValues("/api").Observe(0.15)
	hv.WithLabelValues("/health").Observe(0.01)

	// then
	mf := getMF(t, reg, fq("ns3", "gateway", "latency_seconds"))
	require.Equal(t, ioprometheusclient.MetricType_HISTOGRAM, mf.GetType())

	counts := map[string]uint64{}
	for _, m := range mf.Metric {
		var route string
		for _, l := range m.Label {
			if l.GetName() == "route" {
				route = l.GetValue()
			}
		}
		counts[route] = m.GetHistogram().GetSampleCount()
	}
	require.EqualValues(t, 2, counts["/api"])
	require.EqualValues(t, 1, counts["/health"])
}

func TestNewSimpleHistogram_BucketsCumulative(t *testing.T) {
	// given
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "ns4")

	// when
	h := telemetry.NewSimpleHistogram("op_seconds", "worker", "op duration", []float64{0.1, 0.5, 1.0})
	h.Observe(0.3)
	h.Observe(0.9)

	// then
	mf := getMF(t, reg, fq("ns4", "worker", "op_seconds"))
	require.Equal(t, ioprometheusclient.MetricType_HISTOGRAM, mf.GetType())
	require.Len(t, mf.Metric, 1)

	buckets := mf.Metric[0].GetHistogram().GetBucket()
	got := map[float64]uint64{}
	for _, b := range buckets {
		got[b.GetUpperBound()] = b.GetCumulativeCount()
	}

	require.EqualValues(t, 0, got[0.1])
	require.EqualValues(t, 1, got[0.5]) // 0.3
	require.EqualValues(t, 2, got[1.0]) // 0.3 + 0.9
}

func TestAliases_NewCounterVec_NewGaugeVec_Work(t *testing.T) {
	// given
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "ns5")

	// when
	cv := telemetry.NewCounterVec("events_total", "core", "events", []string{"k"})
	cv.WithLabelValues("x").Add(5)
	gv := telemetry.NewGaugeVec("depth", "core", "queue depth", []string{"q"})
	gv.WithLabelValues("main").Set(42)

	// then
	mfC := getMF(t, reg, fq("ns5", "core", "events_total"))
	require.Equal(t, ioprometheusclient.MetricType_COUNTER, mfC.GetType())
	var csum float64
	for _, m := range mfC.Metric {
		csum += m.GetCounter().GetValue()
	}
	require.Equal(t, 5.0, csum)

	mfG := getMF(t, reg, fq("ns5", "core", "depth"))
	require.Equal(t, ioprometheusclient.MetricType_GAUGE, mfG.GetType())
	require.Len(t, mfG.Metric, 1)
	require.Equal(t, 42.0, mfG.Metric[0].GetGauge().GetValue())
}

func TestSetLabeledRegistry_OverridesNamespace(t *testing.T) {
	// given
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "overridden")

	// when
	telemetry.NewCounterVec("foo_total", "sub", "help", []string{"l"}).WithLabelValues("v")

	// then
	mf := getMF(t, reg, fq("overridden", "sub", "foo_total"))
	require.Equal(t, ioprometheusclient.MetricType_COUNTER, mf.GetType())
}

func TestNewCounterVec_WithConstLabelsRegistersAndCollects(t *testing.T) {
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "constns")

	cv := telemetry.NewCounterVec(
		"events_total",
		"core",
		"events",
		[]string{"topic"},
		telemetry.WithConstLabels(prometheus.Labels{"node_id": "node-a", "cluster_id": "cluster-a"}),
	)
	cv.WithLabelValues("/demo").Add(2)

	mf := getMF(t, reg, fq("constns", "core", "events_total"))
	require.Equal(t, ioprometheusclient.MetricType_COUNTER, mf.GetType())
	metric := singleMetric(t, mf)
	labels := metricLabels(metric)
	require.Equal(t, "node-a", labels["node_id"])
	require.Equal(t, "cluster-a", labels["cluster_id"])
	require.Equal(t, "/demo", labels["topic"])
	require.Equal(t, 2.0, metric.GetCounter().GetValue())
}

func TestNewCounter_WithConstLabelsRegistersAndCollects(t *testing.T) {
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "constns")

	c := telemetry.NewCounter(
		"processed_total",
		"worker",
		"processed",
		telemetry.WithConstLabels(prometheus.Labels{"node_id": "node-a"}),
	)
	c.Add(3)

	mf := getMF(t, reg, fq("constns", "worker", "processed_total"))
	require.Equal(t, ioprometheusclient.MetricType_COUNTER, mf.GetType())
	metric := singleMetric(t, mf)
	labels := metricLabels(metric)
	require.Equal(t, "node-a", labels["node_id"])
	require.Equal(t, 3.0, metric.GetCounter().GetValue())
}

func TestNewGaugeVec_WithConstLabelsRegistersAndCollects(t *testing.T) {
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "constns")

	gv := telemetry.NewGaugeVec(
		"depth",
		"queue",
		"depth",
		[]string{"topic"},
		telemetry.WithConstLabels(prometheus.Labels{"node_id": "node-a"}),
	)
	gv.WithLabelValues("/demo").Set(7)

	metric := assertSingleMetricFamily(
		t,
		reg,
		"queue",
		"depth",
		ioprometheusclient.MetricType_GAUGE,
		map[string]string{"node_id": "node-a", "topic": "/demo"},
	)
	require.Equal(t, 7.0, metric.GetGauge().GetValue())
}

func TestNewGauge_WithConstLabelsRegistersAndCollects(t *testing.T) {
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "constns")

	g := telemetry.NewGauge(
		"node_up_info",
		"core",
		"node up",
		telemetry.WithConstLabels(prometheus.Labels{"node_id": "node-a", "protocol_version": "v2"}),
	)
	g.Set(1)

	metric := assertSingleMetricFamily(
		t,
		reg,
		"core",
		"node_up_info",
		ioprometheusclient.MetricType_GAUGE,
		map[string]string{"node_id": "node-a", "protocol_version": "v2"},
	)
	require.Equal(t, 1.0, metric.GetGauge().GetValue())
}

func TestNewGauge_WithDifferentConstLabelValuesRegistersMultipleSeries(t *testing.T) {
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "constns")

	telemetry.NewGauge(
		"node_up_info",
		"core",
		"node up",
		telemetry.WithConstLabels(prometheus.Labels{"node_id": "node-a"}),
	).Set(1)
	telemetry.NewGauge(
		"node_up_info",
		"core",
		"node up",
		telemetry.WithConstLabels(prometheus.Labels{"node_id": "node-b"}),
	).Set(1)

	mf := getMF(t, reg, fq("constns", "core", "node_up_info"))
	require.Equal(t, ioprometheusclient.MetricType_GAUGE, mf.GetType())
	require.Len(t, mf.Metric, 2)

	nodes := map[string]bool{}
	for _, metric := range mf.Metric {
		nodes[metricLabels(metric)["node_id"]] = true
		require.Equal(t, 1.0, metric.GetGauge().GetValue())
	}
	require.True(t, nodes["node-a"])
	require.True(t, nodes["node-b"])
}

func TestNewHistogram_WithConstLabelsRegistersAndCollects(t *testing.T) {
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "constns")

	hv := telemetry.NewHistogram(
		"latency_seconds",
		"gateway",
		"latency",
		[]string{"topic"},
		telemetry.WithConstLabels(prometheus.Labels{"node_id": "node-a"}),
	)
	hv.WithLabelValues("/demo").Observe(0.05)

	metric := assertSingleMetricFamily(
		t,
		reg,
		"gateway",
		"latency_seconds",
		ioprometheusclient.MetricType_HISTOGRAM,
		map[string]string{"node_id": "node-a", "topic": "/demo"},
	)
	require.EqualValues(t, 1, metric.GetHistogram().GetSampleCount())
}

func TestNewHistogramWithBuckets_WithConstLabelsRegistersAndCollects(t *testing.T) {
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "constns")

	hv := telemetry.NewHistogramWithBuckets(
		"shards_per_message",
		"gateway",
		"shards per message",
		[]string{"topic"},
		[]float64{1, 2, 4},
		telemetry.WithConstLabels(prometheus.Labels{"node_id": "node-a"}),
	)
	hv.WithLabelValues("/demo").Observe(3)

	metric := assertSingleMetricFamily(
		t,
		reg,
		"gateway",
		"shards_per_message",
		ioprometheusclient.MetricType_HISTOGRAM,
		map[string]string{"node_id": "node-a", "topic": "/demo"},
	)
	require.EqualValues(t, 1, metric.GetHistogram().GetSampleCount())
}
