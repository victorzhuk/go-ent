# GC Tuning

Garbage collector tuning with GOGC and GOMEMLIMIT.

## Quick Reference

| Parameter | Purpose | Default | Containers | Impact |
|-----------|---------|---------|------------|--------|
| `GOGC` | GC trigger percentage | 100 | Same | CPU ↔ Memory trade-off |
| `GOMEMLIMIT` | Soft memory limit | None | 90% limit | Prevents OOM kills |
| `GODEBUG=gctrace=1` | GC trace logs | Off | Same | Debug GC behavior |

```bash
# Set GC target (default 100)
GOGC=50 ./app  # More frequent GC (lower memory, higher CPU)
GOGC=200 ./app # Less frequent GC (higher memory, lower CPU)

# Set memory limit (Go 1.19+)
GOMEMLIMIT=2GiB ./app  # Soft memory limit

# Container deployment (90% of limit)
GOMEMLIMIT=1800MiB GOGC=100 ./app  # For 2GB container

# Disable GC (dangerous, testing only)
GOGC=off ./app  # Memory will grow unbounded

# Enable GC trace
GODEBUG=gctrace=1 ./app
```

## GOMEMLIMIT for Containers

Set to 90% of container memory limit to prevent OOM kills.

```bash
# Kubernetes deployment
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: app
    resources:
      limits:
        memory: "2Gi"
    env:
    - name: GOMEMLIMIT
      value: "1800MiB"  # 90% of 2GB
```

```go
import (
    "fmt"
    "runtime/debug"
)

func init() {
    const containerLimit = 2 * 1024 * 1024 * 1024
    limit := int64(float64(containerLimit) * 0.9)
    debug.SetMemoryLimit(limit)
    fmt.Printf("Memory limit: %d bytes\n", limit)
}
```

**Behavior:**
- Soft limit, not hard cap
- GC targets staying below limit
- Can exceed briefly during spikes
- Reduces but doesn't eliminate OOM risk

## GC Pacer

GOGC controls when GC triggers based on heap growth.

**Formula:** `next_gc = live_heap + (live_heap * GOGC / 100)`

```go
// GOGC=100 (default)
// live_heap=100MB → next_gc=200MB (doubles)

// GOGC=50
// live_heap=100MB → next_gc=150MB (1.5x)

// GOGC=200
// live_heap=100MB → next_gc=300MB (3x)
```

**Trade-offs:**

| GOGC | Memory | CPU | Use Case |
|------|--------|-----|----------|
| 50   | Lower  | Higher | Memory-constrained containers |
| 100  | Medium | Medium | Default, balanced |
| 200  | Higher | Lower | Throughput-critical services |

```go
import "runtime/debug"

// Aggressive GC for memory-constrained env
debug.SetGCPercent(50)

// Relaxed GC for high-throughput
debug.SetGCPercent(200)

// Disable GC (testing only)
debug.SetGCPercent(-1)
```

## Ballast Deprecation

**Old technique (pre-Go 1.19):** Allocate large unused slice to inflate heap baseline.

```go
// DEPRECATED: Don't use this
ballast := make([]byte, 10*1024*1024*1024) // 10GB
runtime.KeepAlive(ballast)
```

**Why it existed:** GOGC triggered based on live heap size. Ballast increased baseline, delaying GC.

**Why deprecated:** GOMEMLIMIT provides better control without memory waste.

**Migration:**
```bash
# Old approach
GOGC=100 ./app  # with ballast in code

# New approach
GOMEMLIMIT=2GiB GOGC=100 ./app  # no ballast needed
```

## Monitoring

### Runtime Metrics

```go
import (
    "fmt"
    "runtime"
    "runtime/metrics"
    "time"
)

func monitorGC() {
    samples := []metrics.Sample{
        {Name: "/gc/heap/goal:bytes"},
        {Name: "/gc/heap/live:bytes"},
        {Name: "/memory/classes/total:bytes"},
        {Name: "/sched/gomaxprocs:threads"},
    }

    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        metrics.Read(samples)
        for _, s := range samples {
            fmt.Printf("%s: %v\n", s.Name, s.Value)
        }

        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        fmt.Printf("Alloc: %d MB, NumGC: %d\n",
            m.Alloc/1024/1024, m.NumGC)
    }
}
```

### GC Trace

```bash
GODEBUG=gctrace=1 ./app
```

**Output:**
```
gc 1 @0.005s 3%: 0.018+0.46+0.003 ms clock, 0.14+0.35/0.86/0.050+0.025 ms cpu, 4->4->2 MB, 5 MB goal, 8 P
```

**Fields:**
- `gc 1`: GC cycle number
- `@0.005s`: Time since program start
- `3%`: Percentage of CPU time in GC
- `4->4->2 MB`: Heap size before, after, live data
- `5 MB goal`: Next GC target
- `8 P`: Number of processors

## Runtime Control

```go
import "runtime/debug"

// Adjust GC target
debug.SetGCPercent(50)  // More aggressive GC

// Force GC
runtime.GC()

// Get GC stats
var stats debug.GCStats
debug.ReadGCStats(&stats)
fmt.Println("Last GC:", stats.LastGC)
```

## Common Mistakes

| Mistake | Impact | Fix |
|---------|--------|-----|
| No GOMEMLIMIT in containers | OOM kills | Set to 90% of limit |
| GOGC=off in production | Unbounded memory | Use GOGC=200 instead |
| Over-tuning without measurement | Wasted effort | Profile first |
| Ignoring GC pause times | Latency spikes | Monitor with gctrace |
| Setting GOMEMLIMIT too high | Still OOM | Leave 10% headroom |
| Using ballast in new code | Memory waste | Use GOMEMLIMIT |

## See Also

- [Memory](./memory.md)
- [Profiling](./profiling.md)
- [Docker](../13-devops/docker.md)
- [Kubernetes](../13-devops/kubernetes.md)
