---
name: go-perf
description: "Performance profiling, optimization, caching, concurrency tuning. Auto-activates for: performance issues, profiling, optimization, benchmarks."
---

# Go Performance (1.25+)

## Profiling

```bash
# CPU profile
go test -cpuprofile=cpu.out -bench=. ./...
go tool pprof -http=:8080 cpu.out

# Memory profile
go test -memprofile=mem.out -bench=. ./...
go tool pprof -http=:8080 mem.out

# Live profiling
import _ "net/http/pprof"
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
go tool pprof http://localhost:6060/debug/pprof/heap

# Trace
go test -trace=trace.out ./...
go tool trace trace.out

# Benchmarks
go test -bench=. -benchmem -count=5 ./...
```

## Memory Optimization

```go
// Pre-allocate slices
func process(items []Item) []Result {
    results := make([]Result, 0, len(items)) // pre-alloc
    for _, item := range items {
        results = append(results, convert(item))
    }
    return results
}

// String building
var b strings.Builder
b.Grow(estimatedSize)
for _, v := range items {
    b.WriteString(v.Name)
    b.WriteByte(',')
}
s := b.String()

// Sync.Pool for hot paths
var bufPool = sync.Pool{
    New: func() any { return new(bytes.Buffer) },
}

func process() {
    buf := bufPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufPool.Put(buf)
    }()
    // use buf
}
```

## Concurrency Patterns

```go
// Worker pool with errgroup
func processItems(ctx context.Context, items []Item) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(runtime.GOMAXPROCS(0)) // CPU-bound
    
    for _, item := range items {
        g.Go(func() error {
            return process(ctx, item)
        })
    }
    return g.Wait()
}

// Semaphore for I/O-bound
sem := make(chan struct{}, 100)
for _, item := range items {
    sem <- struct{}{}
    go func() {
        defer func() { <-sem }()
        process(item)
    }()
}

// Rate limiter
limiter := rate.NewLimiter(rate.Limit(1000), 100) // 1000/s, burst 100
if err := limiter.Wait(ctx); err != nil {
    return err
}
```

## Caching

```go
// In-memory TTL cache
import "github.com/jellydator/ttlcache/v3"

cache := ttlcache.New[string, *User](
    ttlcache.WithTTL[string, *User](5 * time.Minute),
    ttlcache.WithCapacity[string, *User](10000),
)
go cache.Start() // cleanup goroutine

// Get or load
item := cache.Get(key)
if item == nil {
    user, _ := repo.FindByID(ctx, id)
    cache.Set(key, user, ttlcache.DefaultTTL)
    return user
}
return item.Value()

// Singleflight for cache stampede
import "golang.org/x/sync/singleflight"

var g singleflight.Group

func getUser(ctx context.Context, id string) (*User, error) {
    v, err, _ := g.Do(id, func() (any, error) {
        return repo.FindByID(ctx, id)
    })
    if err != nil {
        return nil, err
    }
    return v.(*User), nil
}
```

## Database Optimization

```go
// Connection pool tuning
pool.Config().MaxConns = 25
pool.Config().MinConns = 5
pool.Config().MaxConnLifetime = 1 * time.Hour
pool.Config().MaxConnIdleTime = 30 * time.Minute

// Batch inserts
batch := &pgx.Batch{}
for _, item := range items {
    batch.Queue("INSERT INTO items (id, name) VALUES ($1, $2)", item.ID, item.Name)
}
br := pool.SendBatch(ctx, batch)
defer br.Close()

// COPY for bulk
rows := [][]any{}
for _, item := range items {
    rows = append(rows, []any{item.ID, item.Name})
}
_, err := pool.CopyFrom(ctx, pgx.Identifier{"items"}, []string{"id", "name"}, pgx.CopyFromRows(rows))
```

## JSON Optimization

```go
// sonic for high-performance JSON (AMD64/ARM64)
import "github.com/bytedance/sonic"

data, _ := sonic.Marshal(obj)
sonic.Unmarshal(data, &obj)

// Pre-compiled encoder
var userEncoder = sonic.ConfigDefault.NewEncoder(os.Stdout)
userEncoder.Encode(user)
```

## Benchmark Template

```go
func BenchmarkCreateUser(b *testing.B) {
    repo := setupRepo(b)
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        user, _ := entity.NewUser(fmt.Sprintf("user%d@test.com", i), "Test")
        repo.Save(context.Background(), user)
    }
}

func BenchmarkCreateUserParallel(b *testing.B) {
    repo := setupRepo(b)
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            user, _ := entity.NewUser(fmt.Sprintf("user%d@test.com", i), "Test")
            repo.Save(context.Background(), user)
            i++
        }
    })
}
```

## Go 1.25+ Improvements

- Swiss maps (faster map operations)
- Improved GC latency
- Better inlining decisions
- FIPS 140-3 crypto without overhead
