# Connection Pools

Database, HTTP, and gRPC connection pool configuration.

## Quick Reference

| Pool Type | Min Conns | Max Conns | Idle Timeout | Lifetime | Health Check |
|-----------|-----------|-----------|--------------|----------|--------------|
| pgxpool | 2-4 | `NumCPU * 2-4` | 30m | 1h | `pool.Ping(ctx)` |
| HTTP | - | 100 total, 10/host | 90s | - | No built-in |
| gRPC | - | 1-4 conns × streams | - | - | Keepalive 5m |

**Sizing formulas:**
- pgxpool: `MaxConns = min(NumCPU * 4, QPS / 10)`, `MinConns = 2-4`
- gRPC: 1 connection × many streams (multiplexing), scale to 2-4 for high load

**Pool metrics:** Acquire count, wait duration, idle/total conns, timeout errors

## Database Pools

```go
// pgxpool configuration
config.MaxConns = int32(runtime.NumCPU() * 2)
config.MinConns = 2
config.MaxConnLifetime = 1 * time.Hour
config.MaxConnIdleTime = 30 * time.Minute
config.HealthCheckPeriod = 1 * time.Minute
```

## HTTP Client Pool

```go
client := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 10 * time.Second,
}
```

## gRPC Pool

```go
conn, err := grpc.Dial("localhost:50051",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithKeepaliveParams(keepalive.ClientParameters{
        Time:    5 * time.Minute,
        Timeout: 10 * time.Second,
    }),
)
```

## Sizing Formulas

### pgxpool

```go
// Formula: MaxConns = min(NumCPU * multiplier, QPS / avg_query_time_seconds)
// Multiplier: 2-4 depending on workload (2 = CPU-bound, 4 = I/O-bound)
numCPU := runtime.NumCPU()
maxConns := numCPU * 4 // I/O-bound workload

// Alternative: QPS-based sizing
// QPS = 1000, avg query time = 10ms → MaxConns = 1000 / 100 = 10
qps := 1000
avgQueryTime := 10 * time.Millisecond
maxConnsQPS := int32(float64(qps) * avgQueryTime.Seconds())

config.MaxConns = int32(min(maxConns, maxConnsQPS))
config.MinConns = max(2, config.MaxConns/4)
```

### gRPC

```go
// gRPC uses connection multiplexing (many streams per connection)
// Formula: 1-4 connections × unlimited streams
// Scale connections only for very high load (>10k RPS per connection)

// Single connection is usually sufficient
conn, err := grpc.Dial(addr, opts...)

// High load: multiple connections with round-robin
pool := make([]*grpc.ClientConn, 4)
for i := range pool {
    pool[i], err = grpc.Dial(addr, opts...)
}
```

## Health Checks

### pgxpool Health Check

```go
func checkDBHealth(ctx context.Context, pool *pgxpool.Pool) error {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    if err := pool.Ping(ctx); err != nil {
        return fmt.Errorf("ping db: %w", err)
    }
    return nil
}

// Health check callback in config
config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
    var n int
    return conn.QueryRow(ctx, "SELECT 1").Scan(&n)
}
```

### Max Lifetime Rationale

```go
// MaxConnLifetime prevents:
// 1. Stale connections after network issues
// 2. Connection leaks in load balancers
// 3. Memory growth in long-lived connections
// 4. Loss of connection-level stats reset

config.MaxConnLifetime = 1 * time.Hour       // Rotate hourly
config.MaxConnIdleTime = 30 * time.Minute    // Close idle faster
config.HealthCheckPeriod = 1 * time.Minute   // Detect failures early
```

## Pool Metrics

### pgxpool Statistics

```go
func reportPoolStats(pool *pgxpool.Pool) {
    stats := pool.Stat()

    // Connection counts
    slog.Info("pool stats",
        "acquired", stats.AcquiredConns(),      // Currently in use
        "idle", stats.IdleConns(),              // Available
        "total", stats.TotalConns(),            // Acquired + Idle
        "max", stats.MaxConns(),                // Configured max
    )

    // Acquire metrics
    slog.Info("acquire stats",
        "count", stats.AcquireCount(),          // Total acquires
        "duration", stats.AcquireDuration(),    // Total wait time
        "canceled", stats.CanceledAcquireCount(), // Context canceled
    )

    // Pool pressure indicator
    utilizationPct := float64(stats.AcquiredConns()) / float64(stats.MaxConns()) * 100
    if utilizationPct > 80 {
        slog.Warn("pool near capacity", "utilization_pct", utilizationPct)
    }
}
```

### Custom Metrics for Monitoring

```go
// Prometheus metrics
var (
    poolAcquireDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name:    "db_pool_acquire_duration_seconds",
        Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
    })
    poolUtilization = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "db_pool_utilization_ratio",
    })
)

func recordPoolMetrics(pool *pgxpool.Pool) {
    stats := pool.Stat()
    poolUtilization.Set(float64(stats.AcquiredConns()) / float64(stats.MaxConns()))
}
```

## Backpressure

### Pool Exhaustion Behavior

```go
// When pool is exhausted, Acquire() blocks until:
// 1. A connection becomes available (released by another caller)
// 2. Context deadline exceeded
// 3. Context canceled

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

conn, err := pool.Acquire(ctx)
if err != nil {
    // err = context.DeadlineExceeded when pool exhausted for >5s
    return fmt.Errorf("acquire conn: %w", err)
}
defer conn.Release()
```

### Acquire Timeouts

```go
// Service-level timeout wrapper
func withPoolTimeout(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration) (*pgxpool.Conn, error) {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    start := time.Now()
    conn, err := pool.Acquire(ctx)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            slog.Warn("pool acquire timeout", "wait_duration", time.Since(start))
        }
        return nil, err
    }
    return conn, nil
}
```

### Handling Pool Saturation

```go
// Exponential backoff when pool saturated
func acquireWithBackoff(ctx context.Context, pool *pgxpool.Pool) (*pgxpool.Conn, error) {
    backoff := 10 * time.Millisecond
    maxBackoff := 1 * time.Second

    for {
        acquireCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
        conn, err := pool.Acquire(acquireCtx)
        cancel()

        if err == nil {
            return conn, nil
        }

        if !errors.Is(err, context.DeadlineExceeded) {
            return nil, err
        }

        select {
        case <-time.After(backoff):
            backoff = min(backoff*2, maxBackoff)
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
}
```

## Common Mistakes

| Mistake | Impact | Fix |
|---------|--------|-----|
| Pool too small (`MaxConns=1`) | Serialized queries, high latency | Use `NumCPU * 2-4` |
| Infinite acquire timeout | Service hangs forever | Always use `context.WithTimeout` |
| No health checks | Stale connections used | Set `HealthCheckPeriod`, `AfterConnect` |
| Ignoring pool metrics | Can't diagnose bottlenecks | Monitor `AcquireDuration`, utilization |
| Wrong `MaxConnLifetime` | Too short: churn, Too long: leaks | 1h for most workloads |
| No `MinConns` | Cold start latency | Set to 2-4 for warm pool |
| Forgetting `defer conn.Release()` | Permanent connection leak | Always defer immediately |

## See Also

- [PostgreSQL](../04-database/postgresql.md)
- [Redis](../04-database/redis.md)
- [HTTP Client](../05-http-grpc/http-client.md)
- [gRPC](../05-http-grpc/grpc.md)
- [Profiling](profiling.md)
- [Metrics](../07-observability/metrics.md)
