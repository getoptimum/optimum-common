package endpoints

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	BootstrapBaseURL    = "https://bootstrap.getoptimum.io"
	DevBootstrapBaseURL = "https://dev-bootstrap.getoptimum.io"
	CloudflareTraceURL  = "https://cloudflare.com/cdn-cgi/trace"
	CheckIPURL          = "https://checkip.amazonaws.com/"
)

func BootstrapBaseURLOrDefault(baseURL string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return BootstrapBaseURL
	}
	return baseURL
}

func BootstrapConfigURL(baseURL, chainID, clusterID, serviceVersion, gatewayID string) string {
	configURL := fmt.Sprintf("%s/api/v1/%s/%s/config", BootstrapBaseURLOrDefault(baseURL), chainID, clusterID)
	query := url.Values{}
	if serviceVersion != "" {
		query.Set("service_version", serviceVersion)
	}
	if gatewayID != "" {
		query.Set("gateway_id", gatewayID)
	}
	if len(query) == 0 {
		return configURL
	}
	return configURL + "?" + query.Encode()
}

func BootstrapGeolocationURL(baseURL string) string {
	return BootstrapBaseURLOrDefault(baseURL) + "/api/v1/ip/location"
}

func BootstrapTraceURL(baseURL string) string {
	return BootstrapBaseURLOrDefault(baseURL) + "/cdn-cgi/trace"
}
