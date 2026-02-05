# Redis Pub/Sub Quick Reference

Extracted from `docs/go/topics/06-messaging/redis-pubsub.md` (283 lines) → 70 lines.

## Basic Pub/Sub

```go
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

// Subscribe
pubsub := rdb.Subscribe(ctx, "events")
defer pubsub.Close()

ch := pubsub.Channel()
for msg := range ch {
    process(msg.Payload)
}

// Publish
rdb.Publish(ctx, "events", "message")
```

## Limitations

- No persistence (messages lost if no subscribers)
- No acknowledgment (fire-and-forget)
- No consumer groups
- Use NATS or Kafka for reliable messaging
