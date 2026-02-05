# Profiling

Performance profiling using pprof.

## Quick Reference

| Profile Type | Endpoint | Use Case |
|--------------|----------|----------|
| CPU | `/debug/pprof/profile?seconds=30` | CPU-bound operations, hot functions |
| Heap | `/debug/pprof/heap` | Memory usage, allocation patterns |
| Allocs | `/debug/pprof/allocs` | Allocation count regardless of GC |
| Goroutine | `/debug/pprof/goroutine` | Goroutine leaks, concurrent behavior |
| Block | `/debug/pprof/block` | Lock contention, channel blocking |
| Mutex | `/debug/pprof/mutex` | Mutex contention patterns |
| Trace | `/debug/pprof/trace?seconds=5` | Execution trace for detailed analysis |

```go
import _ "net/http/pprof"

// In main:
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Access profiles at:
// http://localhost:6060/debug/pprof/
```

## CPU Profiling

```bash
# Collect CPU profile (30 seconds)
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze
go tool pprof cpu.prof

# Commands in pprof:
# top10    - show top 10 functions
# list funcName - show source code
# web      - open in browser (needs graphviz)
```

## Memory Profiling

```bash
# Heap profile
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Allocations
curl http://localhost:6060/debug/pprof/allocs > allocs.prof
go tool pprof allocs.prof
```

## Benchmark Profiling

```go
func BenchmarkProcess(b *testing.B) {
    for i := 0; i < b.N; i++ {
        process(data)
    }
}

// Run with profiling:
// go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof
// go tool pprof cpu.prof
```

## Goroutine Profile

Identify goroutine leaks and analyze concurrent behavior.

```bash
# Collect goroutine profile
curl http://localhost:6060/debug/pprof/goroutine > goroutine.prof

# Analyze with pprof
go tool pprof goroutine.prof
```

```go
// Production: track goroutine count
func (s *server) healthz(w http.ResponseWriter, r *http.Request) {
    count := runtime.NumGoroutine()
    if count > 10000 {
        log.Warn("goroutine count high", "count", count)
    }
    json.NewEncoder(w).Encode(map[string]int{"goroutines": count})
}
```

Common leak patterns:

```go
// Bad: goroutine leak (no context cancellation)
go func() {
    ticker := time.NewTicker(time.Second)
    for range ticker.C {
        doWork()
    }
}()

// Good: proper cleanup
go func() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            doWork()
        }
    }
}()
```

## Block Profile

Detect contention on mutexes and channels.

```go
import "runtime"

func init() {
    // Enable block profiling (production: use low rate)
    runtime.SetBlockProfileRate(1000000) // 1ms
}
```

```bash
# Collect block profile
curl http://localhost:6060/debug/pprof/block > block.prof

# Analyze
go tool pprof block.prof
```

Analyze contention:

```go
// Production: expose mutex profile separately
import "runtime"

func init() {
    runtime.SetMutexProfileFraction(100) // sample 1% of events
}

// Collect mutex profile
// curl http://localhost:6060/debug/pprof/mutex > mutex.prof
```

## Trace Tool

Detailed execution trace for GC pauses, scheduling, and network blocking.

```bash
# Collect trace (5 seconds, generates large files)
curl http://localhost:6060/debug/pprof/trace?seconds=5 > trace.out

# Analyze with trace tool
go tool trace trace.out
```

Programmatic tracing:

```go
import (
    "os"
    "runtime/trace"
)

func processWithTrace(ctx context.Context) error {
    f, err := os.Create("trace.out")
    if err != nil {
        return err
    }
    defer f.Close()

    if err := trace.Start(f); err != nil {
        return err
    }
    defer trace.Stop()

    // Critical section to analyze
    return doComplexWork(ctx)
}
```

Trace is ideal for:
- GC pause analysis
- Scheduler behavior (goroutine scheduling)
- Network and syscall blocking
- Goroutine coordination patterns

## Continuous Profiling

Production profiling with always-on collection.

```go
import (
    "log"
    "net/http"
    _ "net/http/pprof"
    "runtime"
)

func setupProfiling(addr string) {
    runtime.SetMutexProfileFraction(100)
    runtime.SetBlockProfileRate(1000000)

    mux := http.NewServeMux()
    mux.HandleFunc("/debug/pprof/", pprof.Index)
    mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
    mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
    mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
    mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

    srv := &http.Server{
        Addr:    addr,
        Handler: mux,
    }

    go func() {
        if err := srv.ListenAndServe(); err != nil {
            log.Printf("pprof server: %v", err)
        }
    }()
}
```

Pyroscope integration (continuous profiling):

```go
import (
    "github.com/grafana/pyroscope-go"
)

func initPyroscope(appName, serverURL string) {
    pyroscope.Start(pyroscope.Config{
        ApplicationName: appName,
        ServerAddress:   serverURL,
        ProfileTypes: []pyroscope.ProfileType{
            pyroscope.ProfileCPU,
            pyroscope.ProfileAllocObjects,
            pyroscope.ProfileAllocSpace,
            pyroscope.ProfileInuseObjects,
            pyroscope.ProfileInuseSpace,
        },
    })
}
```

## Common Patterns

```go
// Find allocation hotspots
go tool pprof -alloc_space heap.prof

// Find memory usage
go tool pprof -inuse_space heap.prof

// Compare profiles (before/after optimization)
go tool pprof -base=old.prof new.prof

// Interactive analysis
go tool pprof -http=:8080 cpu.prof
```

## Common Mistakes

| Mistake | Impact | Solution |
|---------|--------|----------|
| Profiling in debug mode | `-N -l` disables optimizations | Use release build or `-gcflags=-l` |
| Short collection duration | Insufficient samples, noise | CPU: 30s+, Trace: 5-10s max |
| No baseline comparison | Can't measure improvement | Always profile before optimization |
| Profiling wrong environment | Dev machine != prod load | Profile staging/prod or realistic load |
| Ignoring allocations | Memory churn impacts GC | Profile both CPU and allocs |
| Profiling with race detector | 10x+ slowdown, skewed results | Profile without `-race` |
| Not enabling block/mutex | Hidden contention issues | Enable in production (low rate) |

## See Also

- [Memory](./memory.md) - Memory optimization techniques
- [GC Tuning](./gc-tuning.md) - Garbage collector optimization
- [Benchmarks](../08-testing/benchmarks.md) - Writing benchmarks
- [Tracing](../07-observability/tracing.md) - Distributed tracing
