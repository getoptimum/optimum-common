# Package: telemetry

**File:** `geolocator.go`

## Functions

### GetCoordinates

```go
func GetCoordinates()
```

GetCoordinates initializes geolocation metrics and starts a background goroutine
that periodically fetches and updates geolocation data from a remote service.
Fetches occur every 10 minutes. This function should be called once at application startup.

---

### GetCurrentCountry

```go
func GetCurrentCountry() string
```

GetCurrentCountry returns the cached country string

---

### GetCurrentCountryISO

```go
func GetCurrentCountryISO() string
```

GetCurrentCountryISO returns the cached country ISO code
