# Package: telemetry

**File:** `metrics.go`

## Functions

### NewCounter

```go
func NewCounter(name , subsystem , help string) prometheus.Counter
```

NewCounter creates a Counter metrics under the global namespace returns nop if metrics are disabled.

---

### NewCounterVec

```go
func NewCounterVec(name , subsystem , help string, labels []string) *prometheus.CounterVec
```

NewCounterVec creates a CounterVec metrics under the global namespace returns nop if metrics are disabled.

---

### NewGaugeVec

```go
func NewGaugeVec(name , subsystem , help string, labels []string) *prometheus.GaugeVec
```

NewGaugeVec creates a Gauge metrics under the global namespace returns nop if metrics are disabled.

---

### NewHistogram

```go
func NewHistogram(name , subsystem , help string, labels []string) *prometheus.HistogramVec
```

NewHistogram creates and registers a new Prometheus histogram metric with the given parameters.

---

### NewHistogramWithBuckets

```go
func NewHistogramWithBuckets(name , subsystem , help string, labels []string, buckets []float64) *prometheus.HistogramVec
```

NewHistogramWithBuckets creates a Histogram metrics with custom buckets.

---

### NewSimpleHistogram

```go
func NewSimpleHistogram(name , subsystem , help string, buckets []float64) prometheus.Histogram
```

NewSimpleHistogram returns a histogram without labels.

---

### SetLabeledRegistry

```go
func SetLabeledRegistry(lr prometheus.Registerer, ns string)
```
