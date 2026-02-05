# Rate Limiting

Rate limiting with golang.org/x/time/rate.

## Quick Reference

| Pattern | Use Case |
|---------|----------|
| `rate.NewLimiter(10, 5)` | 10 req/sec with burst of 5 |
| `limiter.Allow()` | Non-blocking check, returns false if limited |
| `limiter.Reserve()` | Reserve token, check delay |
| `limiter.Wait(ctx)` | Block until token available or context cancelled |
| Redis + Lua | Distributed rate limiting across instances |
| Sliding Window | More accurate than fixed window |
| Per-IP Middleware | Limit by remote address |
| Per-User Middleware | Limit by authenticated user ID |

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Limit(10), 1) // 10 req/sec, burst 1

// In handler
if !limiter.Allow() {
    http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
    return
}
```

## Burst Configuration

Burst allows temporary spikes while maintaining average rate.

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
    // Rate limit would be exceeded even with infinite waiting
    return
}
if r.Delay() > maxWaitTime {
    r.Cancel() // Return token to bucket
    return
}
time.Sleep(r.Delay())

// Wait for token with context timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := limiter.Wait(ctx); err != nil {
    // Context cancelled or deadline exceeded
    return
}
```

## Redis Distributed Limiter

Share rate limit state across multiple service instances using Redis.

```go
import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

// Sliding window rate limiter using Redis sorted sets
type RedisLimiter struct {
    client *redis.Client
    limit  int
    window time.Duration
}

func NewRedisLimiter(client *redis.Client, limit int, window time.Duration) *RedisLimiter {
    return &RedisLimiter{
        client: client,
        limit:  limit,
        window: window,
    }
}

// Lua script ensures atomicity (increment + expiry + count check)
const slidingWindowScript = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

local clearBefore = now - window

redis.call('ZREMRANGEBYSCORE', key, 0, clearBefore)
local count = redis.call('ZCARD', key)

if count < limit then
    redis.call('ZADD', key, now, now)
    redis.call('EXPIRE', key, window)
    return 1
end

return 0
`

func (rl *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {
    now := time.Now().UnixNano()
    windowNanos := rl.window.Nanoseconds()

    result, err := rl.client.Eval(ctx, slidingWindowScript, []string{key},
        now, windowNanos, rl.limit).Int()
    if err != nil {
        return false, fmt.Errorf("eval rate limit script: %w", err)
    }

    return result == 1, nil
}
```

## Sliding Window vs Fixed Window

Sliding window provides more accurate rate limiting.

```go
// Fixed window: can allow 2x limit at window boundary
// Example: 100 req/min limit
// 09:00:59 -> 100 requests OK
// 09:01:00 -> 100 requests OK (200 in 2 seconds!)

// Sliding window: enforces limit over any rolling time window
func (rl *RedisLimiter) SlidingWindow(ctx context.Context, userID string) (bool, error) {
    key := fmt.Sprintf("ratelimit:%s", userID)
    return rl.Allow(ctx, key)
}

// For simple cases without Redis, use token bucket (golang.org/x/time/rate)
// which approximates sliding window behavior with refill rate
```

## Middleware Pattern

Apply rate limiting to HTTP handlers.

```go
// Per-IP rate limiting
type IPRateLimiter struct {
    mu       sync.RWMutex
    limiters map[string]*rate.Limiter
    rate     rate.Limit
    burst    int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
    return &IPRateLimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     r,
        burst:    b,
    }
}

func (ipl *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
    ipl.mu.Lock()
    defer ipl.mu.Unlock()

    limiter, ok := ipl.limiters[ip]
    if !ok {
        limiter = rate.NewLimiter(ipl.rate, ipl.burst)
        ipl.limiters[ip] = limiter
    }
    return limiter
}

func (ipl *IPRateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := getClientIP(r)
        limiter := ipl.getLimiter(ip)

        if !limiter.Allow() {
            w.Header().Set("X-RateLimit-Limit", strconv.Itoa(ipl.burst))
            w.Header().Set("X-RateLimit-Remaining", "0")
            w.Header().Set("Retry-After", "1")
            http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func getClientIP(r *http.Request) string {
    // Check X-Forwarded-For first (behind proxy)
    xff := r.Header.Get("X-Forwarded-For")
    if xff != "" {
        ips := strings.Split(xff, ",")
        return strings.TrimSpace(ips[0])
    }
    // Fallback to RemoteAddr
    ip, _, _ := net.SplitHostPort(r.RemoteAddr)
    return ip
}
```

## Per-User Rate Limiting

```go
type RateLimiter struct {
    mu       sync.RWMutex
    limiters map[string]*rate.Limiter
}

func (rl *RateLimiter) GetLimiter(userID string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    limiter, ok := rl.limiters[userID]
    if !ok {
        limiter = rate.NewLimiter(rate.Limit(10), 1)
        rl.limiters[userID] = limiter
    }
    return limiter
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := getUserID(r)
        limiter := rl.GetLimiter(userID)

        if !limiter.Allow() {
            http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| No burst capacity | Set burst > 1 to allow temporary spikes |
| Single-instance limiter in distributed system | Use Redis or shared state for multi-instance rate limiting |
| Missing 429 headers | Include `X-RateLimit-*` and `Retry-After` headers |
| Wrong rate limit key | Use user ID for authenticated, IP for anonymous |
| No cleanup of old limiters | Implement LRU cache or periodic cleanup to prevent memory leak |

```go
// Bad: No cleanup, memory grows unbounded
type Limiter struct {
    limiters map[string]*rate.Limiter // Never cleaned
}

// Good: LRU cache with max size
import "github.com/hashicorp/golang-lru/v2"

type Limiter struct {
    cache *lru.Cache[string, *rate.Limiter]
}

func NewLimiter(maxEntries int, r rate.Limit, b int) (*Limiter, error) {
    cache, err := lru.NewWithEvict(maxEntries, func(key string, value *rate.Limiter) {
        // Optional: log eviction
    })
    if err != nil {
        return nil, err
    }

    return &Limiter{cache: cache}, nil
}
```

## See Also

- [Authentication](./authentication.md) - User identification for rate limiting
- [Security Headers](./security-headers.md) - Rate limit response headers
- [HTTP Server](./http-server.md) - Middleware patterns
- [Redis](../04-database/redis.md) - Distributed state management
