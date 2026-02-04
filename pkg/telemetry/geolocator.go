package telemetry

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// defaultGeolocationServiceURL is the default URL for geolocation service
	defaultGeolocationServiceURL = "https://bootstrap.getoptimum.io/api/v1/ip/location"
)

type geoLocationResp struct {
	Country    string  `json:"country"`
	CountryISO string  `json:"countryCode"`
	Latitude   float64 `json:"lat"`
	Longitude  float64 `json:"lon"`
}

var (
	geoLocation        *prometheus.GaugeVec
	currentGeoLocation atomic.Pointer[geoLocationResp]
	initOnce           sync.Once
)

// GetCoordinates initializes geolocation metrics and starts a background goroutine
// that periodically fetches and updates geolocation data from a remote service.
// Fetches occur every 10 minutes. This function should be called once at application startup.
func GetCoordinates() {
	initOnce.Do(func() {
		geoLocation = NewGaugeVec(
			"geo_location",
			"det_coordinates",
			"Current geolocation coordinates with country information",
			[]string{"country_iso", "latitude", "longitude"},
		)
	})
	for {
		fetchCoordinatesWithURL(defaultGeolocationServiceURL)
		time.Sleep(10 * time.Minute) // Fetch every 10 minutes
	}
}

func fetchCoordinatesWithURL(serviceURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, code, err := utils.GetCurl[geoLocationResp](ctx, serviceURL, nil)
	if code != http.StatusOK || err != nil {
		return
	}

	if res.Latitude == 0 && res.Longitude == 0 {
		return
	}
	if res.CountryISO == "" {
		res.CountryISO = "unknown"
	}
	if res.Country == "" {
		res.Country = "unknown"
	}

	// Store the geolocation data
	currentGeoLocation.Store(res)

	lat := strconv.FormatFloat(res.Latitude, 'f', 6, 64)
	lon := strconv.FormatFloat(res.Longitude, 'f', 6, 64)
	geoLocation.WithLabelValues(res.CountryISO, lat, lon).Set(1)
}

// GetCurrentCountry returns the cached country string
func GetCurrentCountry() string {
	if cGl := currentGeoLocation.Load(); cGl != nil {
		return cGl.Country
	}
	return "unknown"
}

// GetCurrentCountryISO returns the cached country ISO code
func GetCurrentCountryISO() string {
	if cGl := currentGeoLocation.Load(); cGl != nil {
		return cGl.CountryISO
	}
	return "unknown"
}
