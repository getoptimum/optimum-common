package telemetry_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	ioprom "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestGetCoordinates_RealService_CachesAndRegistersMetric(t *testing.T) {
	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"country":"United States","countryCode":"US","lat":37.7749,"lon":-122.4194}`))
		require.NoError(t, err)
	}))
	defer svc.Close()
	t.Setenv("OPTIMUM_GEOLOCATION_SERVICE_URL", svc.URL)

	// given: fresh registry/namespace so we can find the metric cleanly
	reg := prometheus.NewRegistry()
	telemetry.SetLabeledRegistry(reg, "testns_real")

	// when: start the loop (it fetches once immediately, then sleeps 10min)
	go telemetry.GetCoordinates()

	// then: cache should update shortly after the first fetch
	require.Eventually(t, func() bool {
		return telemetry.GetCurrentCountryISO() == "US"
	}, 10*time.Second, 100*time.Millisecond)

	iso := telemetry.GetCurrentCountryISO()
	require.Equal(t, "US", iso)
	require.Equal(t, "United States", telemetry.GetCurrentCountry())

	// and the metric should exist with correct labels and value=1
	mf := getMF(t, reg, fq("testns_real", "det_coordinates", "geo_location"))
	require.Equal(t, ioprom.MetricType_GAUGE, mf.GetType())
	require.GreaterOrEqual(t, len(mf.Metric), 1)

	// assert expected labels and value
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

		if got["country_iso"] == iso && latStr == "37.774900" && lonStr == "-122.419400" {
			require.Equal(t, 1.0, m.GetGauge().GetValue())
			found = true
			break
		}
	}

	require.True(t, found, "no metric sample matched expected labels")
}
