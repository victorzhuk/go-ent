# Connection Pool Tuning Quick Reference

Extracted from `docs/go/topics/12-performance/connection-pools.md` → 80 lines of actionable patterns.

## Quick Reference Table

| Pool Type | Min Conns | Max Conns           | Idle Timeout | Lifetime | Health Check      |
|-----------|-----------|---------------------|--------------|----------|-------------------|
| pgxpool   | 2-4       | `NumCPU * 2-4`      | 30m          | 1h       | `pool.Ping(ctx)`  |
| HTTP      | -         | 100 total, 10/host  | 90s          | -        | No built-in       |
| gRPC      | -         | 1-4 conns × streams | -            | -        | Keepalive 5m      |

## Sizing Formulas

```go
// pgxpool: MaxConns = min(NumCPU * multiplier, QPS / avg_query_time)
numCPU := runtime.NumCPU()
maxConns := numCPU * 4  // 2 = CPU-bound, 4 = I/O-bound

// QPS-based: QPS=1000, avg=10ms → MaxConns = 1000/100 = 10
qps := 1000
avgQueryTime := 10 * time.Millisecond
maxConnsQPS := int32(float64(qps) * avgQueryTime.Seconds())

config.MaxConns = int32(min(maxConns, maxConnsQPS))
config.MinConns = max(2, config.MaxConns/4)

// gRPC: 1 connection × many streams (multiplexing)
// Scale to 2-4 only for very high load (>10k RPS per conn)
```

## Database Pool (pgxpool)

```go
import "github.com/jackc/pgx/v5/pgxpool"

config, _ := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
config.MaxConns = int32(runtime.NumCPU() * 2)
config.MinConns = 2
config.MaxConnLifetime = 1 * time.Hour
config.MaxConnIdleTime = 30 * time.Minute
config.HealthCheckPeriod = 1 * time.Minute

pool, err := pgxpool.NewWithConfig(ctx, config)
```

## HTTP Client Pool

```go
client := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,  // total across all hosts
        MaxIdleConnsPerHost: 10,   // per host
        IdleConnTimeout:     90 * time.Second,
        MaxConnsPerHost:     100,  // limit concurrent per host
    },
    Timeout: 10 * time.Second,  // total request timeout
}
```

## gRPC Connection Pool

```go
import "google.golang.org/grpc/keepalive"

// Single connection (usually sufficient due to multiplexing)
conn, err := grpc.Dial("localhost:50051",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithKeepaliveParams(keepalive.ClientParameters{
        Time:    5 * time.Minute,   // ping interval
        Timeout: 10 * time.Second,  // ping timeout
    }),
)

// High load: pool of 2-4 connections with round-robin
pool := make([]*grpc.ClientConn, 4)
for i := range pool {
    pool[i], _ = grpc.Dial(addr, opts...)
}
nextConn := atomic.Uint32{}
func getConn() *grpc.ClientConn {
    return pool[nextConn.Add(1) % uint32(len(pool))]
}
```

## Health Checks

```go
// pgxpool
func checkDB(ctx context.Context, pool *pgxpool.Pool) error {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    return pool.Ping(ctx)
}

// HTTP - no built-in, implement custom
func checkHTTP(client *http.Client, url string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := client.Do(req)
    if err != nil { return err }
    resp.Body.Close()
    return nil
}
```
