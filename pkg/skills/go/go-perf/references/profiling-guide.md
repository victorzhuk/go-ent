# Profiling Guide Quick Reference

Extracted from `docs/go/topics/12-performance/profiling.md` → 100 lines of actionable patterns.

## Profile Types Quick Reference

| Profile    | Endpoint                              | Use Case                          |
|------------|---------------------------------------|-----------------------------------|
| CPU        | `/debug/pprof/profile?seconds=30`     | CPU-bound operations, hot funcs   |
| Heap       | `/debug/pprof/heap`                   | Memory usage, allocation patterns |
| Allocs     | `/debug/pprof/allocs`                 | Allocation count (ignores GC)     |
| Goroutine  | `/debug/pprof/goroutine`              | Goroutine leaks, concurrency      |
| Block      | `/debug/pprof/block`                  | Lock contention, channel blocking |
| Mutex      | `/debug/pprof/mutex`                  | Mutex contention patterns         |
| Trace      | `/debug/pprof/trace?seconds=5`        | Detailed execution analysis       |

## Setup pprof

```go
import _ "net/http/pprof"

// In main:
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Access: http://localhost:6060/debug/pprof/
```

## CPU Profiling

```bash
# Collect 30-second profile
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze
go tool pprof cpu.prof

# pprof commands:
# top10           - show top 10 functions
# list funcName   - show source code
# web             - open graph in browser (needs graphviz)
# png             - save as image
```

## Memory Profiling

```bash
# Heap profile
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Allocations (total count)
curl http://localhost:6060/debug/pprof/allocs > allocs.prof
go tool pprof allocs.prof

# In pprof:
# -alloc_space   - total allocation size
# -inuse_space   - currently in-use memory
# -alloc_objects - total allocation count
# -inuse_objects - currently in-use objects
```

## Benchmark Profiling

```go
func BenchmarkProcess(b *testing.B) {
    data := generateTestData()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        process(data)
    }
}

// Run with profiling:
// go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof
// go tool pprof cpu.prof
```

## Goroutine Leak Detection

```bash
# Collect goroutine profile
curl http://localhost:6060/debug/pprof/goroutine > goroutine.prof
go tool pprof goroutine.prof
```

```go
// Track goroutine count in health check
func (s *server) healthz(w http.ResponseWriter, r *http.Request) {
    count := runtime.NumGoroutine()
    if count > 10000 {
        log.Warn("high goroutine count", "count", count)
    }
    json.NewEncoder(w).Encode(map[string]int{"goroutines": count})
}

// Common leak: missing context cancellation
// Bad
go func() {
    ticker := time.NewTicker(time.Second)
    for range ticker.C { doWork() }
}()

// Good - with cancellation
go func() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done(): return
        case <-ticker.C: doWork()
        }
    }
}()
```

## Block Profile (Lock Contention)

```go
// Enable block profiling
runtime.SetBlockProfileRate(1)

// Collect
curl http://localhost:6060/debug/pprof/block > block.prof
go tool pprof block.prof
```

## Mutex Profile

```go
// Enable mutex profiling
runtime.SetMutexProfileFraction(1)

// Collect
curl http://localhost:6060/debug/pprof/mutex > mutex.prof
go tool pprof mutex.prof
```

## Execution Trace

```bash
# Collect 5-second trace
curl http://localhost:6060/debug/pprof/trace?seconds=5 > trace.out

# View in browser
go tool trace trace.out
```
