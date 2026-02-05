# Redis Pub/Sub

Pub/Sub and Streams with go-redis. Use Streams for durable messaging, Pub/Sub for ephemeral broadcast.

## Quick Reference

| Operation | Pub/Sub | Streams |
|-----------|---------|---------|
| Publish | `Publish(ctx, channel, msg)` | `XAdd(ctx, &XAddArgs{Stream, Values})` |
| Subscribe | `Subscribe(ctx, channels...)` | `XRead/XReadGroup(ctx, &XReadArgs{})` |
| Pattern subscribe | `PSubscribe(ctx, "prefix:*")` | N/A (filter in consumer) |
| Acknowledge | N/A (fire-and-forget) | `XAck(ctx, stream, group, ids...)` |
| Pending messages | N/A | `XPending(ctx, stream, group)` |
| Consumer groups | N/A | `XGroupCreate/XReadGroup` |
| Durability | None (missed if offline) | Persisted on disk |
| Backpressure | Client-side buffering | Block with timeout |

**Library:** `github.com/redis/go-redis/v9`

## Pub/Sub (Fire-and-Forget)

Use for ephemeral broadcasts where message loss is acceptable.

```go
type notifier struct {
    rdb *redis.Client
    log *slog.Logger
}

func (n *notifier) Publish(ctx context.Context, channel, msg string) error {
    if err := n.rdb.Publish(ctx, channel, msg).Err(); err != nil {
        return fmt.Errorf("publish to %s: %w", channel, err)
    }
    return nil
}

func (n *notifier) Subscribe(ctx context.Context, channels ...string) error {
    pubsub := n.rdb.Subscribe(ctx, channels...)
    defer pubsub.Close()

    ch := pubsub.Channel()
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg := <-ch:
            n.log.Info("received", "channel", msg.Channel, "payload", msg.Payload)
        }
    }
}

// Pattern subscription (e.g., "events:*" matches "events:user", "events:order")
func (n *notifier) PSubscribe(ctx context.Context, patterns ...string) error {
    pubsub := n.rdb.PSubscribe(ctx, patterns...)
    defer pubsub.Close()

    ch := pubsub.Channel()
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg := <-ch:
            n.log.Info("received", "pattern", msg.Pattern, "channel", msg.Channel, "payload", msg.Payload)
        }
    }
}
```

## Redis Streams

Use for durable messaging with acknowledgment and consumer groups.

### Basic Streams Operations

```go
type streamProducer struct {
    rdb *redis.Client
}

// Add event to stream with auto-generated ID
func (p *streamProducer) Add(ctx context.Context, stream string, data map[string]interface{}) (string, error) {
    id, err := p.rdb.XAdd(ctx, &redis.XAddArgs{
        Stream: stream,
        MaxLen: 10000, // Prevent unbounded growth
        Approx: true,  // ~10000 entries (faster)
        Values: data,
    }).Result()
    if err != nil {
        return "", fmt.Errorf("xadd to %s: %w", stream, err)
    }
    return id, nil
}

// Read new messages (blocking)
func (p *streamProducer) Read(ctx context.Context, stream string, lastID string) ([]redis.XMessage, error) {
    result, err := p.rdb.XRead(ctx, &redis.XReadArgs{
        Streams: []string{stream, lastID},
        Block:   5 * time.Second, // Block up to 5s
        Count:   100,              // Max messages per read
    }).Result()
    if err == redis.Nil {
        return nil, nil // No new messages
    }
    if err != nil {
        return nil, fmt.Errorf("xread from %s: %w", stream, err)
    }
    if len(result) > 0 {
        return result[0].Messages, nil
    }
    return nil, nil
}
```

## Consumer Groups

Distribute stream processing across multiple consumers with acknowledgment.

```go
type streamConsumer struct {
    rdb   *redis.Client
    log   *slog.Logger
    group string
    name  string
}

// Initialize consumer group (idempotent)
func (c *streamConsumer) CreateGroup(ctx context.Context, stream string) error {
    err := c.rdb.XGroupCreateMkStream(ctx, stream, c.group, "$").Err()
    if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
        return fmt.Errorf("create group %s: %w", c.group, err)
    }
    return nil
}

// Read and process messages as part of consumer group
func (c *streamConsumer) Consume(ctx context.Context, stream string, handler func(map[string]interface{}) error) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        // Read new messages
        result, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
            Group:    c.group,
            Consumer: c.name,
            Streams:  []string{stream, ">"},
            Count:    10,
            Block:    5 * time.Second,
        }).Result()

        if err == redis.Nil {
            continue // No messages
        }
        if err != nil {
            return fmt.Errorf("xreadgroup: %w", err)
        }

        for _, stream := range result {
            for _, msg := range stream.Messages {
                if err := handler(msg.Values); err != nil {
                    c.log.Error("handler failed", "id", msg.ID, "err", err)
                    continue
                }

                // Acknowledge successful processing
                if err := c.rdb.XAck(ctx, stream.Stream, c.group, msg.ID).Err(); err != nil {
                    c.log.Error("ack failed", "id", msg.ID, "err", err)
                }
            }
        }
    }
}

// Claim stale messages (not acked by dead consumers)
func (c *streamConsumer) ClaimPending(ctx context.Context, stream string, minIdle time.Duration) error {
    pending, err := c.rdb.XPendingExt(ctx, &redis.XPendingExtArgs{
        Stream: stream,
        Group:  c.group,
        Start:  "-",
        End:    "+",
        Count:  100,
        Idle:   minIdle,
    }).Result()
    if err != nil {
        return fmt.Errorf("xpending: %w", err)
    }

    if len(pending) == 0 {
        return nil
    }

    ids := make([]string, len(pending))
    for i, p := range pending {
        ids[i] = p.ID
    }

    claimed, err := c.rdb.XClaim(ctx, &redis.XClaimArgs{
        Stream:   stream,
        Group:    c.group,
        Consumer: c.name,
        MinIdle:  minIdle,
        Messages: ids,
    }).Result()
    if err != nil {
        return fmt.Errorf("xclaim: %w", err)
    }

    c.log.Info("claimed stale messages", "count", len(claimed))
    return nil
}
```

## Comparison: Pub/Sub vs Streams

| Feature | Pub/Sub | Streams |
|---------|---------|---------|
| **Durability** | None (lost if no subscriber) | Persisted to disk |
| **Acknowledgment** | None | XACK required |
| **Consumer groups** | No | Yes (load balancing) |
| **Message history** | No | Yes (queryable by ID/time) |
| **Backpressure** | Client buffer overflow | Block with timeout |
| **Patterns** | Yes (`PSUBSCRIBE`) | No (filter in code) |
| **Use case** | Real-time notifications, caching | Task queues, event sourcing |
| **When offline** | Messages lost | Messages waiting |

**Use Pub/Sub when:**
- Real-time broadcasts (notifications, cache invalidation)
- No delivery guarantees needed
- Pattern-based routing required
- Low latency critical (no persistence overhead)

**Use Streams when:**
- Delivery guarantees required
- Consumer groups needed (load balancing)
- Message replay/history required
- Event sourcing or audit log

## Common Mistakes

| Mistake | Impact | Fix |
|---------|--------|-----|
| Pub/Sub for critical messages | Lost if subscriber offline | Use Streams with consumer groups |
| No `MaxLen` on streams | Unbounded memory growth | Set `MaxLen: 10000, Approx: true` |
| Forgetting `XAck` | Messages redelivered forever | Always ack after successful processing |
| No timeout on `XRead` | Goroutine leaks on shutdown | Set `Block: 5*time.Second` |
| Ignoring pending entries | Stale messages accumulate | Periodic `XPending` + `XClaim` cleanup |
| Pattern subscribe without limit | Regex DoS attack | Validate patterns, limit subscriptions |
| No graceful shutdown | In-flight messages lost | Context cancellation + drain loop |

## See Also

- [Redis](../04-database/redis.md) - Connection pooling, configuration
- [RabbitMQ](./rabbitmq.md) - Full-featured message broker
- [Kafka](./kafka.md) - High-throughput event streaming
- [NATS](./nats.md) - Cloud-native messaging
