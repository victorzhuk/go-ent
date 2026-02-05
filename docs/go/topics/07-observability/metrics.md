# Metrics

Production metrics collection and exposition using Prometheus. Prometheus is the de-facto standard for metrics in Go services, with native support for counters, gauges, histograms, and summaries.

**RED Method**: Rate (requests/sec), Errors (error rate), Duration (latency)
**USE Method**: Utilization (% used), Saturation (queue depth), Errors

## Quick Reference

| Concept | Purpose | Example |
|---------|---------|---------|
| `Counter` | Monotonically increasing value | `http_requests_total`, `errors_total` |
| `Gauge` | Value that goes up and down | `goroutines_count`, `memory_bytes` |
| `Histogram` | Distribution of values with buckets | `http_request_duration_seconds` |
| `Summary` | Distribution with calculated quantiles | `rpc_latency_seconds` (rare, use Histogram) |
| `Registry` | Collector registration and exposition | `prometheus.DefaultRegistrar`, custom registry |
| `Labels` | Dimensional data | `{method="GET", status="200"}` |
| `promhttp.Handler()` | HTTP exposition endpoint | `/metrics` endpoint |
| `InstrumentHandler` | HTTP middleware instrumentation | Automatic RED metrics |

**Library**: `github.com/prometheus/client_golang v1.x`

## Metric Types

### Counter

Monotonically increasing counter. Use for events that only go up: requests, errors, bytes sent.

```go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "handler", "status"},
	)

	ordersProcessedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_processed_total",
			Help: "Total number of orders processed",
		},
	)
)

// Usage
httpRequestsTotal.WithLabelValues("GET", "/api/users", "200").Inc()
ordersProcessedTotal.Add(5)
```

**Always suffix with `_total`** for counters.

### Gauge

Value that can go up or down. Use for current state: goroutines, connections, queue size.

```go
var (
	goroutinesCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "goroutines_count",
			Help: "Current number of goroutines",
		},
	)

	queueDepth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_depth",
			Help: "Current queue depth",
		},
		[]string{"queue_name"},
	)

	memoryBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_bytes",
			Help: "Current memory usage in bytes",
		},
	)
)

// Usage
goroutinesCount.Set(float64(runtime.NumGoroutine()))
queueDepth.WithLabelValues("orders").Inc()
queueDepth.WithLabelValues("orders").Dec()
memoryBytes.Set(float64(stats.Alloc))
```

**Suffix with `_bytes`, `_count`, etc.** to indicate unit.

### Histogram

Distribution of values across predefined buckets. Preferred for latency and size distributions.

```go
var (
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
		},
		[]string{"method", "handler"},
	)

	responseSizeBytes = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 2, 10), // 100, 200, 400, 800...
		},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.5, 1, 5},
		},
		[]string{"query_type"},
	)
)

// Usage
timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues("GET", "/api/users"))
defer timer.ObserveDuration()

// Or manual
start := time.Now()
// ... do work ...
httpRequestDuration.WithLabelValues("GET", "/api/users").Observe(time.Since(start).Seconds())
```

**Custom buckets** for specific use cases:
- **Latency**: `[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.5, 1, 5}`
- **Size**: `prometheus.ExponentialBuckets(100, 2, 10)`
- **Default**: `prometheus.DefBuckets` (reasonable for most HTTP)

### Summary

Distribution with client-side calculated quantiles. **Rare**: prefer Histogram (quantiles calculated in Prometheus).

```go
var (
	rpcLatency = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "rpc_latency_seconds",
			Help:       "RPC latency in seconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"service", "method"},
	)
)

// Usage
rpcLatency.WithLabelValues("auth", "Login").Observe(0.042)
```

**Prefer Histogram**: Summaries cannot be aggregated across instances. Use only when you need exact quantiles without server-side calculation.

## Registry

### Default Registry

`promauto` uses `prometheus.DefaultRegisterer` automatically. Simplest approach for most services.

```go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "status"},
	)
)

// Expose metrics
// GET /metrics
// http.Handle("/metrics", promhttp.Handler())
```

### Custom Registry

Isolate metrics, avoid global state, or register collectors conditionally.

```go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry      *prometheus.Registry
	requestsTotal *prometheus.CounterVec
	duration      *prometheus.HistogramVec
}

func New() *Metrics {
	reg := prometheus.NewRegistry()

	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "status"},
	)

	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "handler"},
	)

	reg.MustRegister(requestsTotal, duration)

	return &Metrics{
		registry:      reg,
		requestsTotal: requestsTotal,
		duration:      duration,
	}
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
```

**When to use custom registry**:
- Testing (isolated metrics per test)
- Multi-tenancy (separate metrics per tenant)
- Avoiding global state in libraries

## Labels and Cardinality

Labels add dimensions to metrics. **Cardinality = product of all label values**.

### Best Practices

**Good labels** (bounded cardinality):
```go
// Low cardinality: ~100 unique series
httpRequestsTotal.WithLabelValues(
	"GET",           // method: ~10 values
	"/api/users",    // handler: ~50 values
	"200",           // status: ~20 values
)
```

**Bad labels** (unbounded cardinality):
```go
// HIGH CARDINALITY - DO NOT DO THIS
httpRequestsTotal.WithLabelValues(
	userID,          // unbounded: millions of users
	requestID,       // unbounded: every request unique
	timestamp,       // unbounded: infinite
	clientIP,        // unbounded: millions of IPs
)
```

### Cardinality Calculation

```
cardinality = label1_values × label2_values × label3_values

Example:
methods (10) × handlers (50) × statuses (20) = 10,000 series
```

**Keep cardinality under 10,000 per metric**. High cardinality causes:
- Memory bloat in Prometheus
- Slow queries
- Scrape timeouts

### Label Guidelines

| Do | Don't |
|---|---|
| `{method="GET"}` | `{user_id="12345"}` |
| `{status="200"}` | `{request_id="abc-123"}` |
| `{handler="/api/users"}` | `{timestamp="2024-01-01"}` |
| `{queue="orders"}` | `{client_ip="1.2.3.4"}` |
| `{service="auth"}` | `{full_path="/api/users/12345/orders/67890"}` |

**Drop high-cardinality data or use logs/traces instead.**

## HTTP Instrumentation

### Standard Middleware

Auto-instrumentation for all HTTP handlers.

```go
package server

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.5, 1, 5},
		},
		[]string{"method", "handler", "status"},
	)

	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "handler", "status"},
	)

	httpInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current in-flight HTTP requests",
		},
	)
)

func InstrumentHandler(handlerName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpInFlight.Inc()
		defer httpInFlight.Dec()

		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		status := wrapped.statusCode

		httpDuration.WithLabelValues(r.Method, handlerName, http.StatusText(status)).Observe(duration)
		httpRequestsTotal.WithLabelValues(r.Method, handlerName, http.StatusText(status)).Inc()
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
```

Usage:
```go
mux := http.NewServeMux()
mux.Handle("/metrics", promhttp.Handler())
mux.Handle("/api/users", InstrumentHandler("/api/users", usersHandler))
mux.Handle("/api/orders", InstrumentHandler("/api/orders", ordersHandler))

http.ListenAndServe(":8080", mux)
```

### Using promhttp.InstrumentHandler

Built-in instrumentation from `prometheus/client_golang`.

```go
import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	mux := http.NewServeMux()

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Instrumented handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	instrumentedHandler := promhttp.InstrumentHandlerDuration(
		httpDuration.MustCurryWith(prometheus.Labels{"handler": "/api/users"}),
		promhttp.InstrumentHandlerCounter(
			httpRequestsTotal.MustCurryWith(prometheus.Labels{"handler": "/api/users"}),
			promhttp.InstrumentHandlerInFlight(httpInFlight, handler),
		),
	)

	mux.Handle("/api/users", instrumentedHandler)

	http.ListenAndServe(":8080", mux)
}
```

**Prefer custom middleware** for simpler label management and clearer code.

## Common Patterns

### RED Method Implementation

Rate, Errors, Duration for HTTP services.

```go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Rate
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests (rate)",
		},
		[]string{"method", "handler", "status"},
	)

	// Errors
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total HTTP errors",
		},
		[]string{"method", "handler"},
	)

	// Duration
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.5, 1, 5},
		},
		[]string{"method", "handler"},
	)
)

func RecordRequest(method, handler string, status int, duration float64) {
	requestsTotal.WithLabelValues(method, handler, http.StatusText(status)).Inc()
	requestDuration.WithLabelValues(method, handler).Observe(duration)

	if status >= 500 {
		errorsTotal.WithLabelValues(method, handler).Inc()
	}
}
```

### Queue Metrics

Track queue depth, processing rate, and errors.

```go
var (
	queueDepth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_depth",
			Help: "Current queue depth",
		},
		[]string{"queue_name"},
	)

	queueProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "queue_processed_total",
			Help: "Total messages processed",
		},
		[]string{"queue_name", "status"},
	)

	queueProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "queue_processing_duration_seconds",
			Help:    "Message processing duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"queue_name"},
	)
)

func (q *Queue) Enqueue(msg Message) {
	queueDepth.WithLabelValues(q.name).Inc()
	q.ch <- msg
}

func (q *Queue) processMessage(msg Message) {
	defer queueDepth.WithLabelValues(q.name).Dec()

	start := time.Now()
	err := q.handler(msg)
	queueProcessingDuration.WithLabelValues(q.name).Observe(time.Since(start).Seconds())

	status := "success"
	if err != nil {
		status = "error"
	}
	queueProcessed.WithLabelValues(q.name, status).Inc()
}
```

### Database Connection Pool

Track pool stats from `pgxpool.Stat()`.

```go
var (
	dbConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections",
			Help: "Database connections by state",
		},
		[]string{"state"}, // idle, acquired, constructing
	)

	dbConnAcquireDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "db_conn_acquire_duration_seconds",
			Help:    "Time to acquire connection from pool",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.5, 1},
		},
	)
)

func RecordPoolStats(stat *pgxpool.Stat) {
	dbConnections.WithLabelValues("idle").Set(float64(stat.IdleConns()))
	dbConnections.WithLabelValues("acquired").Set(float64(stat.AcquiredConns()))
	dbConnections.WithLabelValues("constructing").Set(float64(stat.ConstructingConns()))
}

// Call periodically
go func() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		RecordPoolStats(pool.Stat())
	}
}()
```

### Background Job Metrics

Track job execution, duration, and failures.

```go
var (
	jobExecutions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "job_executions_total",
			Help: "Total job executions",
		},
		[]string{"job_name", "status"},
	)

	jobDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "job_duration_seconds",
			Help:    "Job execution duration",
			Buckets: prometheus.ExponentialBuckets(1, 2, 10), // 1s, 2s, 4s, 8s...
		},
		[]string{"job_name"},
	)

	jobLastRun = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "job_last_run_timestamp_seconds",
			Help: "Unix timestamp of last job run",
		},
		[]string{"job_name"},
	)
)

func RunJob(name string, fn func() error) error {
	start := time.Now()
	err := fn()
	jobDuration.WithLabelValues(name).Observe(time.Since(start).Seconds())
	jobLastRun.WithLabelValues(name).Set(float64(time.Now().Unix()))

	status := "success"
	if err != nil {
		status = "failure"
	}
	jobExecutions.WithLabelValues(name, status).Inc()

	return err
}
```

## Naming Conventions

Follow Prometheus naming best practices.

### Metric Names

**Format**: `{namespace}_{subsystem}_{name}_{unit}`

```go
// Good
http_requests_total              // counter suffix
http_request_duration_seconds    // histogram with unit
db_query_duration_seconds
queue_depth                      // gauge, unit implicit
memory_bytes                     // gauge with unit

// Bad
httpRequests                     // camelCase
requests                         // no namespace
http_duration                    // missing unit
http_request_duration_ms         // use seconds, not ms
```

**Rules**:
- `snake_case` always
- Counters end with `_total`
- Units: `_seconds`, `_bytes`, `_ratio` (0-1)
- Time always in seconds, size in bytes
- Boolean state: `_enabled`, `_active`

### Label Names

```go
// Good
{method="GET", handler="/api/users", status="OK"}

// Bad
{Method="GET"}                   // uppercase
{http_method="GET"}              // redundant prefix
{handler="/api/users/12345"}     // high cardinality
```

**Rules**:
- `snake_case` always
- No prefix if metric name provides context
- Bounded values only
- Status as text: `"OK"`, `"Internal Server Error"`

## Common Mistakes

| Mistake | Impact | Fix |
|---------|--------|-----|
| **High cardinality labels** (`user_id`, `request_id`) | Memory bloat, slow queries | Use bounded labels only |
| **Missing `_total` suffix** on counters | Violates Prometheus conventions | Always suffix counters with `_total` |
| **Using Summary instead of Histogram** | Cannot aggregate across instances | Use Histogram, calculate quantiles in Prometheus |
| **Unbounded labels** (`{path=r.URL.Path}`) | Infinite cardinality | Use handler name, not full path |
| **Wrong metric type** (Gauge for request count) | Broken calculations (rate, increase) | Counter for cumulative, Gauge for current state |
| **No buckets customization** | Poor percentile accuracy | Tune buckets for your latency distribution |
| **Time in milliseconds** | Violates conventions | Always use seconds |
| **Including error message in label** | High cardinality | Use error type or code |
| **Not exporting Go runtime metrics** | Missing memory/GC visibility | `import _ "github.com/prometheus/client_golang/prometheus"` |

## See Also

- [Tracing](tracing.md) - Distributed tracing with OpenTelemetry
- [Logging](logging.md) - Structured logging with slog/zerolog
- [Correlation](correlation.md) - Request ID and trace context propagation
- [HTTP Server](../05-http-grpc/http-server.md) - HTTP server patterns and middleware
- [Profiling](../12-performance/profiling.md) - CPU and memory profiling
