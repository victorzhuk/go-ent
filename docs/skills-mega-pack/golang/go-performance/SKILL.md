---
name: go-performance
description: Go performance profiling, optimization, benchmarking, memory management, and production tuning
---

# Go Performance

## Profiling
```bash
# CPU profile
go test -cpuprofile cpu.prof -bench .
go tool pprof cpu.prof

# Memory profile
go test -memprofile mem.prof -bench .
go tool pprof -alloc_space mem.prof

# Trace
go test -trace trace.out -bench .
go tool trace trace.out

# Live profiling (add to server)
import _ "net/http/pprof"
go func() { http.ListenAndServe(":6060", nil) }()
```

## Benchmarking
- Always use `b.ResetTimer()` after expensive setup
- Use `b.ReportAllocs()` to track allocations
- Run with `-count=5` and compare with `benchstat`
- Use `-benchmem` flag for memory statistics
- Benchmark realistic workloads, not micro-operations

## Memory Optimization
- Pre-allocate slices: `make([]T, 0, expectedCap)`
- Use `sync.Pool` for frequently allocated temporary objects
- Avoid `interface{}` in hot paths — causes allocations
- Use `strings.Builder` for string concatenation
- Reuse buffers with `bytes.Buffer` pool
- Watch for hidden allocations: closures capturing variables, `append` growth

## Concurrency Optimization
- Profile goroutine count — more isn't always better
- Use `GOMAXPROCS` tuning for containerized environments
- Prefer `sync.RWMutex` when reads vastly outnumber writes
- Use atomic operations (`sync/atomic`) for simple counters
- Batch work to reduce goroutine overhead

## Compiler Hints
- Use `//go:inline` sparingly for critical hot-path functions
- `//go:nosplit` and `//go:noescape` for advanced optimization
- Enable PGO (Profile-Guided Optimization) for production builds:
  `go build -pgo=default.pgo`
- Use build tags for architecture-specific optimizations

## Production Tuning
- Set `GOMEMLIMIT` for containerized deployments
- Monitor GC pause times with `runtime/metrics`
- Use `GOGC` tuning (default 100) — increase for throughput, decrease for latency
- Enable HTTP/2 for multiplexed connections
- Use connection pooling for database and HTTP clients
