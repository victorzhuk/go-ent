---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
triggers:
  - debug
  - troubleshoot
  - diagnose
  - fix bug
---

# ${SKILL_NAME}

## Role

Debugging expert specializing in systematic troubleshooting, effective logging, distributed tracing, and debugging strategies. Focus on root cause analysis, diagnostic techniques, and efficient problem-solving approaches.

## Instructions

### Systematic Debugging Methodology

Follow a structured approach:

1. Reproduce the issue consistently
2. Isolate the failing component
3. Form hypothesis about root cause
4. Test hypothesis with minimal change
5. Fix and verify with tests

### Structured Logging

```go
log.Error("payment failed",
    "error", err,
    "user_id", userID,
    "amount", amount,
    "currency", currency,
    "correlation_id", correlationID,
)
```

### Race Condition Detection

```bash
go test -race ./...
go run -race cmd/server/main.go
```

```go
// Protect shared state
type SafeCache struct {
    mu    sync.RWMutex
    store map[string]any
}

func (c *SafeCache) Get(key string) (any, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    v, ok := c.store[key]
    return v, ok
}
```

### Performance Profiling

```go
import _ "net/http/pprof"

// In main.go
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

```bash
go tool pprof http://localhost:6060/debug/pprof/heap
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### Goroutine Leak Detection

```go
func TestNoGoroutineLeaks(t *testing.T) {
    defer goleak.VerifyNone(t)
    // test code that spawns goroutines
}
```

### Edge Cases

If issue only appears in production: Add detailed structured logging with correlation IDs, use sampling for high-volume events.

If performance regression is found: Profile with pprof before and after the change, establish baselines.

If goroutine leak is suspected: Use goleak in tests, check for missing context cancellation propagation.

## Examples

### Example 1: Debug memory leak with pprof

**Input**: The service memory keeps growing, how do I find the leak?

**Output**:
```go
// Add pprof endpoint
import _ "net/http/pprof"

// Start debug server
go func() {
    if err := http.ListenAndServe("localhost:6060", nil); err != nil {
        log.Error("pprof server failed", "error", err)
    }
}()
```

```bash
# Take baseline heap profile
curl -s http://localhost:6060/debug/pprof/heap > heap1.out

# Wait and take second profile
sleep 60
curl -s http://localhost:6060/debug/pprof/heap > heap2.out

# Compare profiles
go tool pprof -base heap1.out heap2.out

# In pprof shell
(pprof) top10
(pprof) list YourFunction
```

### Example 2: Add correlation ID tracing

**Input**: Add request tracing to debug intermittent failures

**Output**:
```go
type contextKey string

const correlationIDKey contextKey = "correlation_id"

func WithCorrelationID(ctx context.Context) context.Context {
    return context.WithValue(ctx, correlationIDKey, uuid.New().String())
}

func CorrelationID(ctx context.Context) string {
    id, _ := ctx.Value(correlationIDKey).(string)
    return id
}

// Middleware
func CorrelationIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := WithCorrelationID(r.Context())
        w.Header().Set("X-Correlation-ID", CorrelationID(ctx))
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```
