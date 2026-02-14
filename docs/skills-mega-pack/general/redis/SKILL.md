---
name: redis-patterns
description: Redis caching strategies, data structures, pub/sub, Lua scripting, and production best practices
---

# Redis Patterns

## Caching Strategies
```
Cache-Aside (Lazy Loading):
1. Check cache for key
2. If miss: query database
3. Store result in cache with TTL
4. Return result

Write-Through:
1. Write to cache AND database
2. Cache always has latest data
3. Higher write latency, lower read latency
```

## Data Structures
- **String**: Simple key-value, counters, serialized objects
- **Hash**: Object-like structures (user profiles)
- **List**: Queues (LPUSH/RPOP), recent items
- **Set**: Unique collections, intersections, unions
- **Sorted Set**: Leaderboards, priority queues, time-series
- **Stream**: Event streaming, message queues (like Kafka-lite)

## Common Patterns
```redis
-- Rate limiting (sliding window)
MULTI
ZADD ratelimit:{user_id} {now} {request_id}
ZREMRANGEBYSCORE ratelimit:{user_id} 0 {now - window}
ZCARD ratelimit:{user_id}
EXPIRE ratelimit:{user_id} {window}
EXEC

-- Distributed lock
SET lock:{resource} {token} NX EX 30
-- Release (Lua for atomicity)
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
end

-- Session storage
HSET session:{id} user_id 123 role admin expires_at 1700000000
EXPIRE session:{id} 3600
```

## Production Best Practices
- Set `maxmemory` and `maxmemory-policy` (allkeys-lru for cache)
- Use TTL on all cache keys — avoid unbounded growth
- Use pipelining for batch operations
- Monitor with `INFO`, `SLOWLOG`, `MEMORY USAGE`
- Use Redis Cluster for horizontal scaling
- Use separate Redis instances for cache vs persistent data
- Enable AOF persistence for data you can't afford to lose
