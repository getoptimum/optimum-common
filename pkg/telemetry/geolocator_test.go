package telemetry_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	io_prom "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

// --- tests ---
func Test_GetCoordinates_RealService_CachesAndRegistersMetric(t *testing.T) {
	// given: fresh registry/namespace so we can find the metric cleanly
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "testns_real")

	// when: start the loop (it fetches once immediately, then sleeps 10min)
	go telemetry.GetCoordinates()

	// then: cache should become non-"unknown" shortly if the service is reachable
	require.Eventually(t, func() bool {
		return telemetry.GetCurrentCountryISO() != "unknown"
	}, 10*time.Second, 100*time.Millisecond)

	iso := telemetry.GetCurrentCountryISO()
	require.NotEmpty(t, iso)
	require.NotEqual(t, "unknown", iso)
	require.NotEmpty(t, telemetry.GetCurrentCountry())

	// and the metric should exist with correct labels and value=1
	mf := getMF(t, reg, fq("testns_real", "det_coordinates", "geo_location"))
	require.Equal(t, io_prom.MetricType_GAUGE, mf.GetType())
	require.GreaterOrEqual(t, len(mf.Metric), 1)

	// we don’t know exact coords ahead of time; assert label presence & consistency
	found := false
	for _, m := range mf.Metric {
		lbls := m.GetLabel()
		require.Len(t, lbls, 3)

		// build map name->value for easier asserts
		got := map[string]string{
			lbls[0].GetName(): lbls[0].GetValue(),
			lbls[1].GetName(): lbls[1].GetValue(),
			lbls[2].GetName(): lbls[2].GetValue(),
		}

		// required label keys exist
		_, hasISO := got["country_iso"]
		latStr, hasLat := got["latitude"]
		lonStr, hasLon := got["longitude"]
		require.True(t, hasISO && hasLat && hasLon, "missing expected labels")

		// iso matches cached iso; coords are numeric strings (not necessarily stable)
		if got["country_iso"] == iso {
			// numeric check
			if _, err := strconv.ParseFloat(latStr, 64); err == nil {
				if _, err := strconv.ParseFloat(lonStr, 64); err == nil {
					require.Equal(t, 1.0, m.GetGauge().GetValue())
					found = true
					break
				}
			}
		}
	}
	require.True(t, found, "no metric sample matched cached ISO/valid coords")
}
