# Rate Limiting Quick Reference

Extracted from `docs/go/topics/11-security/rate-limiting.md` (354 lines) → 80 lines of actionable patterns.

## Quick Reference Table

| Pattern                   | Use Case                              |
|---------------------------|---------------------------------------|
| `rate.NewLimiter(10, 5)`  | 10 req/sec with burst of 5            |
| `limiter.Allow()`         | Non-blocking check, returns false if limited |
| `limiter.Reserve()`       | Reserve token, check delay            |
| `limiter.Wait(ctx)`       | Block until token available           |
| Redis + Lua               | Distributed rate limiting             |
| Sliding Window            | More accurate than fixed window       |
| Per-IP Middleware         | Limit by remote address               |
| Per-User Middleware       | Limit by authenticated user ID        |

## Basic Usage

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Limit(10), 1)  // 10 req/sec, burst 1

// In handler
if !limiter.Allow() {
    http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
    return
}
```

## Burst Configuration

```go
// Allow bursts of up to 20 requests, then enforce 10 req/sec average
limiter := rate.NewLimiter(rate.Limit(10), 20)

// Check if allowed (non-blocking)
if limiter.Allow() {
    // Process request
}

// Reserve a token (check delay before consuming)
r := limiter.Reserve()
if !r.OK() {
    return  // Rate limit exceeded
}
if r.Delay() > maxWaitTime {
    r.Cancel()  // Return token to bucket
    return
}
time.Sleep(r.Delay())

// Wait for token with context timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := limiter.Wait(ctx); err != nil {
    return  // Context cancelled or deadline exceeded
}
```

## Per-IP Middleware

```go
import (
    "net/http"
    "sync"
    "golang.org/x/time/rate"
)

type IPRateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    r        rate.Limit
    b        int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
    return &IPRateLimiter{
        limiters: make(map[string]*rate.Limiter),
        r:        r,
        b:        b,
    }
}

func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
    i.mu.Lock()
    defer i.mu.Unlock()

    limiter, exists := i.limiters[ip]
    if !exists {
        limiter = rate.NewLimiter(i.r, i.b)
        i.limiters[ip] = limiter
    }
    return limiter
}

func (i *IPRateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := r.RemoteAddr
        limiter := i.getLimiter(ip)

        if !limiter.Allow() {
            http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

## Redis Distributed Limiter

```go
import "github.com/redis/go-redis/v9"

type RedisLimiter struct {
    client *redis.Client
    limit  int
    window time.Duration
}

func (r *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {
    now := time.Now().UnixNano()
    windowStart := now - r.window.Nanoseconds()

    pipe := r.client.Pipeline()

    // Remove old entries
    pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprint(windowStart))

    // Count current entries
    pipe.ZCard(ctx, key)

    // Add new entry
    pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

    // Set expiration
    pipe.Expire(ctx, key, r.window)

    cmds, err := pipe.Exec(ctx)
    if err != nil {
        return false, err
    }

    count := cmds[1].(*redis.IntCmd).Val()
    return count < int64(r.limit), nil
}
```
