# Metrics Patterns Quick Reference

Extracted from `docs/go/topics/07-observability/metrics.md` (667 lines) → 120 lines.

## Prometheus Metrics

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    requestCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
)

func init() {
    prometheus.MustRegister(requestCounter, requestDuration)
}
```

## Recording Metrics

```go
func (s *server) handleRequest(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

    // Handle request...

    status := 200
    requestCounter.WithLabelValues(r.Method, r.URL.Path, fmt.Sprint(status)).Inc()
    requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
}
```
