# GC Tuning Quick Reference

Extracted from `docs/go/topics/12-performance/gc-tuning.md` → 60 lines of actionable patterns.

## Quick Reference Table

| Parameter           | Purpose              | Default | Containers | Impact                  |
|---------------------|----------------------|---------|------------|-------------------------|
| `GOGC`              | GC trigger %         | 100     | Same       | CPU ↔ Memory trade-off  |
| `GOMEMLIMIT`        | Soft memory limit    | None    | 90% limit  | Prevents OOM kills      |
| `GODEBUG=gctrace=1` | GC trace logs        | Off     | Same       | Debug GC behavior       |

## Basic Usage

```bash
# More frequent GC (lower memory, higher CPU)
GOGC=50 ./app

# Less frequent GC (higher memory, lower CPU)
GOGC=200 ./app

# Set memory limit (Go 1.19+)
GOMEMLIMIT=2GiB ./app

# Container deployment (90% of limit)
GOMEMLIMIT=1800MiB GOGC=100 ./app  # For 2GB container

# Disable GC (dangerous, testing only)
GOGC=off ./app

# Enable GC trace
GODEBUG=gctrace=1 ./app
```

## GOMEMLIMIT for Containers

Set to 90% of container memory limit to prevent OOM kills.

```yaml
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
import "runtime/debug"

func init() {
    const containerLimit = 2 * 1024 * 1024 * 1024
    limit := int64(float64(containerLimit) * 0.9)
    debug.SetMemoryLimit(limit)
}

// Behavior:
// - Soft limit, not hard cap
// - GC targets staying below limit
// - Can exceed briefly during spikes
```

## GC Pacer Formula

**Formula:** `next_gc = live_heap + (live_heap * GOGC / 100)`

```go
// GOGC=100 (default): live_heap=100MB → next_gc=200MB (doubles)
// GOGC=50:  live_heap=100MB → next_gc=150MB (50% growth)
// GOGC=200: live_heap=100MB → next_gc=300MB (triples)
```

## Tuning Guidelines

| Scenario               | GOGC | GOMEMLIMIT | Reasoning                    |
|------------------------|------|------------|------------------------------|
| Memory-constrained     | 50   | 90% limit  | Frequent GC, stay under limit|
| CPU-constrained        | 200  | None       | Reduce GC overhead           |
| Balanced (default)     | 100  | 90% limit  | Good starting point          |
| Spiky workload         | 100  | 90% limit  | Limit prevents OOM spikes    |
```
