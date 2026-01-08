---
name: go-perf
description: "Performance profiling, optimization, benchmarks. Auto-activates for: performance issues, profiling, optimization, memory leaks, benchmarking."
---

# Go Performance (1.25+)

## Profiling

```bash
go test -cpuprofile=cpu.out -bench=. ./...
go tool pprof -http=:8080 cpu.out

go test -memprofile=mem.out -bench=. ./...
go tool pprof -http=:8080 mem.out

# Live
import _ "net/http/pprof"
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Benchmarks
go test -bench=. -benchmem -count=5 ./...
```

## Memory Optimization

```go
// Pre-allocate slices
results := make([]Result, 0, len(items))

// String building
var b strings.Builder
b.Grow(estimatedSize)

// Sync.Pool for hot paths
var bufPool = sync.Pool{
    New: func() any { return new(bytes.Buffer) },
}
```

## Concurrency

```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(runtime.GOMAXPROCS(0))

for _, item := range items {
    g.Go(func() error {
        return process(ctx, item)
    })
}
return g.Wait()
```

## Singleflight (Cache Stampede)

```go
var g singleflight.Group

func getUser(ctx context.Context, id string) (*User, error) {
    v, err, _ := g.Do(id, func() (any, error) {
        return repo.FindByID(ctx, id)
    })
    return v.(*User), err
}
```

## Database

```go
// Connection pool
pool.Config().MaxConns = 25
pool.Config().MinConns = 5

// Batch inserts
batch := &pgx.Batch{}
for _, item := range items {
    batch.Queue("INSERT INTO items (id, name) VALUES ($1, $2)", item.ID, item.Name)
}
br := pool.SendBatch(ctx, batch)
```

## Rate Limiting

```go
limiter := rate.NewLimiter(rate.Limit(1000), 100)
if err := limiter.Wait(ctx); err != nil {
    return err
}
```

## Context7

```
mcp__context7__resolve(library: "pprof")
mcp__context7__resolve(library: "singleflight")
mcp__context7__resolve(library: "rate")
```
