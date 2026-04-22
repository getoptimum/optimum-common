package endpoints_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/internal/endpoints"
	"github.com/stretchr/testify/require"
)

func TestBootstrapBaseURLOrDefault(t *testing.T) {
	require.Equal(t, endpoints.BootstrapBaseURL, endpoints.BootstrapBaseURLOrDefault(""))
	require.Equal(t, endpoints.BootstrapBaseURL, endpoints.BootstrapBaseURLOrDefault("  "))
	require.Equal(t, "https://dev-bootstrap.getoptimum.io", endpoints.BootstrapBaseURLOrDefault("https://dev-bootstrap.getoptimum.io/"))
}

func TestBootstrapURLs(t *testing.T) {
	require.Equal(
		t,
		"https://bootstrap.getoptimum.io/api/v1/hoodi/optimum/config?service_version=gateway-v0.0.1-rc11",
		endpoints.BootstrapConfigURL(endpoints.BootstrapBaseURL, "hoodi", "optimum", "gateway-v0.0.1-rc11"),
	)
	require.Equal(
		t,
		"https://bootstrap.getoptimum.io/api/v1/ip/location",
		endpoints.BootstrapGeolocationURL(endpoints.BootstrapBaseURL),
	)
	require.Equal(
		t,
		"https://bootstrap.getoptimum.io/cdn-cgi/trace",
		endpoints.BootstrapTraceURL(endpoints.BootstrapBaseURL),
	)
}
