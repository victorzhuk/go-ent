# Redis

Redis integration using go-redis/v9 for caching, pub/sub, and data structures.

## Quick Reference

| Operation | Usage | Notes |
|-----------|-------|-------|
| **Basic KV** | `Set(ctx, key, val, ttl)`, `Get(ctx, key)` | Core operations |
| **Pipeline** | `Pipe()`, `Exec(ctx)` | Batch commands, reduce RTT |
| **Lua Scripts** | `Eval(ctx, script, keys, args)` | Atomic operations |
| **Cluster** | `NewClusterClient(&ClusterOptions{...})` | Horizontal scaling |
| **Sentinel** | `NewFailoverClient(&FailoverOptions{...})` | High availability |
| **Redlock** | `NewRedlock(clients...)` | Distributed locks |
| **TTL** | `Expire(ctx, key, dur)`, `TTL(ctx, key)` | Expiration patterns |

```go
import "github.com/redis/go-redis/v9"

client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

// Basic operations
client.Set(ctx, "key", "value", time.Hour)
val, err := client.Get(ctx, "key").Result()
client.Del(ctx, "key")

// Lists, Sets, Hashes, Sorted Sets all supported
```

## Connection Setup

```go
func NewRedisClient(ctx context.Context) *redis.Client {
    client := redis.NewClient(&redis.Options{
        Addr:         "localhost:6379",
        Password:     "",
        DB:           0,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        PoolSize:     10,
    })

    if err := client.Ping(ctx).Err(); err != nil {
        log.Fatal(err)
    }

    return client
}
```

## Caching Pattern

```go
func GetUser(ctx context.Context, id string) (*User, error) {
    // Try cache first
    cached, err := rdb.Get(ctx, "user:"+id).Result()
    if err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // Cache miss - fetch from DB
    user, err := db.GetUser(ctx, id)
    if err != nil {
        return nil, err
    }

    // Cache for 1 hour
    data, _ := json.Marshal(user)
    rdb.Set(ctx, "user:"+id, data, time.Hour)

    return user, nil
}
```

## Pipeline

Batch multiple commands into a single round-trip to reduce network overhead.

```go
func IncrementCounters(ctx context.Context, ids []string) error {
    pipe := rdb.Pipeline()

    for _, id := range ids {
        pipe.Incr(ctx, "counter:"+id)
        pipe.Expire(ctx, "counter:"+id, 24*time.Hour)
    }

    _, err := pipe.Exec(ctx)
    if err != nil {
        return fmt.Errorf("pipeline exec: %w", err)
    }
    return nil
}

func BulkGetUsers(ctx context.Context, ids []string) (map[string]string, error) {
    pipe := rdb.Pipeline()
    cmds := make([]*redis.StringCmd, len(ids))

    for i, id := range ids {
        cmds[i] = pipe.Get(ctx, "user:"+id)
    }

    if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
        return nil, fmt.Errorf("pipeline exec: %w", err)
    }

    result := make(map[string]string)
    for i, cmd := range cmds {
        val, err := cmd.Result()
        if err == redis.Nil {
            continue
        }
        if err != nil {
            return nil, fmt.Errorf("get result for %s: %w", ids[i], err)
        }
        result[ids[i]] = val
    }
    return result, nil
}
```

## Lua Scripts

Execute atomic operations on Redis server using Lua scripts.

```go
var rateLimitScript = redis.NewScript(`
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local current = redis.call('GET', key)

if current and tonumber(current) >= limit then
    return 0
end

redis.call('INCR', key)
redis.call('EXPIRE', key, window)
return 1
`)

func RateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
    result, err := rateLimitScript.Run(ctx, rdb, []string{key}, limit, int(window.Seconds())).Result()
    if err != nil {
        return false, fmt.Errorf("rate limit script: %w", err)
    }
    return result.(int64) == 1, nil
}

var deductInventoryScript = redis.NewScript(`
local key = KEYS[1]
local amount = tonumber(ARGV[1])
local current = tonumber(redis.call('GET', key) or "0")

if current < amount then
    return -1
end

redis.call('DECRBY', key, amount)
return current - amount
`)

func DeductInventory(ctx context.Context, productID string, qty int) (int64, error) {
    key := "inventory:" + productID
    result, err := deductInventoryScript.Run(ctx, rdb, []string{key}, qty).Result()
    if err != nil {
        return 0, fmt.Errorf("deduct inventory: %w", err)
    }

    remaining := result.(int64)
    if remaining < 0 {
        return 0, fmt.Errorf("insufficient inventory")
    }
    return remaining, nil
}
```

## Cluster and Sentinel

### Redis Cluster (Horizontal Scaling)

```go
func NewClusterClient() *redis.ClusterClient {
    return redis.NewClusterClient(&redis.ClusterOptions{
        Addrs: []string{
            "node1:6379",
            "node2:6379",
            "node3:6379",
        },
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        PoolSize:     10,
    })
}

func ClusterSet(ctx context.Context, key, value string) error {
    // Cluster client handles slot routing automatically
    return cluster.Set(ctx, key, value, time.Hour).Err()
}
```

### Sentinel (High Availability)

```go
func NewSentinelClient() *redis.Client {
    return redis.NewFailoverClient(&redis.FailoverOptions{
        MasterName:    "mymaster",
        SentinelAddrs: []string{"sentinel1:26379", "sentinel2:26379", "sentinel3:26379"},
        DB:            0,
        PoolSize:      10,
    })
}

func handleFailover(ctx context.Context, client *redis.Client) {
    // Client automatically reconnects to new master after failover
    for {
        if err := client.Ping(ctx).Err(); err != nil {
            log.Printf("redis unavailable: %v", err)
            time.Sleep(time.Second)
            continue
        }
        break
    }
}
```

## Redlock

Distributed lock implementation using Redlock algorithm.

```go
import "github.com/go-redsync/redsync/v4"
import "github.com/go-redsync/redsync/v4/redis/goredis/v9"

var rs *redsync.Redsync

func init() {
    pool := goredis.NewPool(rdb)
    rs = redsync.New(pool)
}

func ProcessWithLock(ctx context.Context, resourceID string) error {
    mutex := rs.NewMutex("lock:"+resourceID,
        redsync.WithExpiry(8*time.Second),
        redsync.WithTries(3),
    )

    if err := mutex.LockContext(ctx); err != nil {
        return fmt.Errorf("acquire lock: %w", err)
    }
    defer mutex.UnlockContext(ctx)

    // Critical section
    return processResource(ctx, resourceID)
}

func TryProcessWithLock(ctx context.Context, resourceID string) (bool, error) {
    mutex := rs.NewMutex("lock:"+resourceID, redsync.WithExpiry(5*time.Second))

    if err := mutex.LockContext(ctx); err != nil {
        if err == redsync.ErrTaken {
            return false, nil
        }
        return false, fmt.Errorf("lock error: %w", err)
    }
    defer mutex.UnlockContext(ctx)

    if err := processResource(ctx, resourceID); err != nil {
        return true, err
    }
    return true, nil
}
```

## TTL Patterns

### Automatic Expiration

```go
func CacheWithTTL(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
    serialized, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("marshal: %w", err)
    }
    return rdb.Set(ctx, key, serialized, ttl).Err()
}

func GetWithRefresh(ctx context.Context, key string, ttl time.Duration) (string, error) {
    val, err := rdb.Get(ctx, key).Result()
    if err != nil {
        return "", err
    }

    // Refresh TTL on access
    rdb.Expire(ctx, key, ttl)
    return val, nil
}
```

### TTL Checking

```go
func CheckExpiration(ctx context.Context, key string) (time.Duration, bool, error) {
    ttl, err := rdb.TTL(ctx, key).Result()
    if err != nil {
        return 0, false, fmt.Errorf("get ttl: %w", err)
    }

    if ttl == -2 {
        return 0, false, nil // Key doesn't exist
    }
    if ttl == -1 {
        return 0, true, nil // Key exists but has no expiration
    }
    return ttl, true, nil
}

func SetExpireIfNotSet(ctx context.Context, key string, ttl time.Duration) error {
    existing, err := rdb.TTL(ctx, key).Result()
    if err != nil {
        return fmt.Errorf("get ttl: %w", err)
    }

    // Only set expiration if none exists (-1)
    if existing == -1 {
        return rdb.Expire(ctx, key, ttl).Err()
    }
    return nil
}
```

### Sliding Window Cache

```go
func GetOrFetch(ctx context.Context, key string, ttl time.Duration, fetchFn func(context.Context) (string, error)) (string, error) {
    val, err := rdb.Get(ctx, key).Result()
    if err == nil {
        // Extend TTL on hit (sliding window)
        rdb.Expire(ctx, key, ttl)
        return val, nil
    }

    if err != redis.Nil {
        return "", fmt.Errorf("redis get: %w", err)
    }

    // Cache miss - fetch and cache
    fetched, err := fetchFn(ctx)
    if err != nil {
        return "", err
    }

    rdb.Set(ctx, key, fetched, ttl)
    return fetched, nil
}
```

## Common Mistakes

| Mistake | Why Bad | Fix |
|---------|---------|-----|
| **Multiple sequential calls** | Each call = network RTT | Use `Pipeline()` for bulk operations |
| **Race conditions in counters** | Non-atomic increment checks | Use Lua scripts for check-and-set |
| **Single node in production** | No failover capability | Use Sentinel or Cluster |
| **Redlock with single instance** | Not distributed, defeats purpose | Use 3+ independent instances |
| **Missing TTL on cache keys** | Memory bloat, stale data forever | Always set TTL on cached data |
| **Ignoring `redis.Nil` error** | Cache miss treated as error | Check `err == redis.Nil` explicitly |
| **Hardcoded cluster slots** | Manual slot management breaks | Let client handle slot routing |
| **Long-running Lua scripts** | Blocks Redis single thread | Keep scripts under 5ms, use batching |

## See Also

- [Redis Pub/Sub](../06-messaging/redis-pubsub.md)
- [PostgreSQL](./postgresql.md)
- [Rate Limiting](../11-security/rate-limiting.md)
- [Connection Pools](../12-performance/connection-pools.md)
