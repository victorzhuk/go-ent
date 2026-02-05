# Kafka

Event streaming with kafka-go.

## Quick Reference

| Feature | Config/Method | Notes |
|---------|---------------|-------|
| **Producer** | `kafka.NewWriter(cfg)` | Thread-safe, reusable |
| Acks | `RequiredAcks: -1/1/0` | all/leader/none |
| Compression | `CompressionCodec: snappy.NewCompressionCodec()` | snappy/gzip/lz4/zstd |
| Batching | `BatchSize: 100, BatchTimeout: 10ms` | Throughput vs latency |
| Idempotence | `Idempotent: true` | Exactly-once semantics |
| Write | `WriteMessages(ctx, msgs...)` | Sync with retries |
| **Consumer** | `kafka.NewReader(cfg)` | NOT thread-safe |
| Group ID | `GroupID: "service-v1"` | Consumer group membership |
| Partition | `Partition: 0` | Direct partition (no GroupID) |
| Offset | `StartOffset: kafka.FirstOffset` | Start position |
| Auto-commit | `CommitInterval: time.Second` | Auto vs manual |
| Read | `ReadMessage(ctx)` | Blocking, auto-commits |
| Fetch | `FetchMessage(ctx)` | Manual commit required |
| Commit | `CommitMessages(ctx, msg)` | Manual offset commit |
| **Offset Mgmt** | `SetOffset(offset)` | Seek to position |
| At-least-once | Use `FetchMessage` + `CommitMessages` | Default pattern |
| At-most-once | Auto-commit + process after read | Data loss risk |
| **Error Handling** | Check `err` on all ops | Transient vs permanent |

## Producer Configuration

```go
import (
    "github.com/segmentio/kafka-go"
    "github.com/segmentio/kafka-go/compress/snappy"
)

w := &kafka.Writer{
    Addr:     kafka.TCP("broker1:9092", "broker2:9092"),
    Topic:    "events",

    // Acks: -1 = all replicas, 1 = leader only, 0 = no ack
    RequiredAcks: kafka.RequireAll,

    // Compression reduces network and disk usage
    Compression: snappy.NewCompressionCodec(),

    // Batching: buffer messages for throughput
    BatchSize:    100,              // max messages per batch
    BatchBytes:   1048576,          // max bytes per batch (1MB)
    BatchTimeout: 10 * time.Millisecond,

    // Idempotence: exactly-once producer semantics
    Idempotent: true,

    // Retries for transient failures
    MaxAttempts: 3,

    // Balancer: how to distribute messages across partitions
    Balancer: &kafka.Hash{},  // or kafka.LeastBytes, kafka.RoundRobin
}
defer w.Close()

err := w.WriteMessages(ctx,
    kafka.Message{
        Key:   []byte("user-123"),
        Value: []byte(`{"event":"created"}`),
        Headers: []kafka.Header{
            {Key: "trace-id", Value: []byte("abc123")},
        },
    },
)
```

**Acks Tradeoff:**
- `RequireAll (-1)`: Strongest durability, highest latency
- `RequireOne (1)`: Balanced (most common)
- `RequireNone (0)`: Fastest, data loss on leader failure

## Consumer Groups

```go
r := kafka.NewReader(kafka.ReaderConfig{
    Brokers: []string{"broker1:9092", "broker2:9092"},
    Topic:   "events",

    // GroupID enables consumer group semantics
    GroupID: "service-v1",

    // Partition assignment strategy
    GroupBalancers: []kafka.GroupBalancer{
        kafka.RangeBalancer{},  // default
        kafka.RoundRobinBalancer{},
    },

    // Start offset for new consumer group
    StartOffset: kafka.LastOffset,  // or kafka.FirstOffset

    // Commit interval (auto-commit with ReadMessage)
    CommitInterval: time.Second,

    // Session timeout: how long broker waits for heartbeat
    SessionTimeout: 10 * time.Second,

    // Heartbeat interval
    HeartbeatInterval: 3 * time.Second,

    // Partition watchdog (rebalance if consumer stuck)
    PartitionWatchInterval: 5 * time.Second,

    // Max wait time for fetch
    MaxWait: 500 * time.Millisecond,

    // Min/max bytes per fetch
    MinBytes: 10e3,  // 10KB
    MaxBytes: 10e6,  // 10MB
})
defer r.Close()

// Rebalancing: occurs when consumers join/leave
// OnPartitionsAssigned/Revoked hooks available via Stats()
```

**Consumer Group Rules:**
- Each partition consumed by exactly one consumer in group
- Consumers > partitions = idle consumers
- Partitions > consumers = multiple partitions per consumer
- Rebalance triggers: consumer join/leave, partition change

## Offset Management

```go
// AUTO-COMMIT: ReadMessage commits after successful read
msg, err := r.ReadMessage(ctx)
if err != nil {
    return fmt.Errorf("read message: %w", err)
}
// offset committed automatically

// MANUAL COMMIT: FetchMessage requires explicit commit
msg, err := r.FetchMessage(ctx)
if err != nil {
    return fmt.Errorf("fetch message: %w", err)
}

if err := processMessage(msg); err != nil {
    // handle error, don't commit
    return fmt.Errorf("process: %w", err)
}

// commit only after successful processing
if err := r.CommitMessages(ctx, msg); err != nil {
    return fmt.Errorf("commit: %w", err)
}

// SEEKING: manually set offset position
err = r.SetOffset(kafka.FirstOffset)  // reprocess from start
err = r.SetOffsetAt(ctx, time.Now().Add(-24*time.Hour))  // rewind 24h
```

**Delivery Guarantees:**

| Pattern | Implementation | Guarantee | Risk |
|---------|----------------|-----------|------|
| At-most-once | Auto-commit before process | Message loss | Failure after read |
| At-least-once | Manual commit after process | Duplicate processing | Failure before commit |
| Exactly-once | Idempotent + transactional | No duplicates/loss | Requires coordination |

## Batching

```go
// PRODUCE BATCH: WriteMessages is already batched
messages := make([]kafka.Message, 0, 100)
for i := 0; i < 100; i++ {
    messages = append(messages, kafka.Message{
        Value: []byte(fmt.Sprintf("msg-%d", i)),
    })
}
err := w.WriteMessages(ctx, messages...)  // single network call

// CONSUME BATCH: use FetchMessage in loop
const batchSize = 100
batch := make([]kafka.Message, 0, batchSize)

for len(batch) < batchSize {
    fetchCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
    msg, err := r.FetchMessage(fetchCtx)
    cancel()

    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            break  // partial batch ok
        }
        return fmt.Errorf("fetch: %w", err)
    }
    batch = append(batch, msg)
}

// process batch
if err := processBatch(batch); err != nil {
    return fmt.Errorf("process batch: %w", err)
}

// commit last offset (commits all preceding)
if len(batch) > 0 {
    if err := r.CommitMessages(ctx, batch[len(batch)-1]); err != nil {
        return fmt.Errorf("commit: %w", err)
    }
}
```

**Batching Tuning:**
- Producer: increase `BatchSize`/`BatchBytes` for throughput
- Consumer: increase `MaxBytes`/`MaxWait` for larger fetches
- Trade latency for throughput in high-volume scenarios

## Error Handling

```go
// PRODUCER ERRORS: transient vs permanent
err := w.WriteMessages(ctx, msg)
if err != nil {
    var kafkaErr kafka.Error
    if errors.As(err, &kafkaErr) {
        if kafkaErr.Temporary() {
            // retry after backoff
            return fmt.Errorf("transient kafka error: %w", err)
        }
        // permanent: log and skip or DLQ
        logger.Error("permanent kafka error", "err", err)
        return nil
    }
    return fmt.Errorf("write message: %w", err)
}

// CONSUMER ERRORS: poison pill handling
for {
    msg, err := r.FetchMessage(ctx)
    if err != nil {
        if errors.Is(err, context.Canceled) {
            return nil
        }
        logger.Error("fetch failed", "err", err)
        time.Sleep(time.Second)
        continue
    }

    if err := processMessage(msg); err != nil {
        logger.Error("process failed",
            "topic", msg.Topic,
            "partition", msg.Partition,
            "offset", msg.Offset,
            "err", err,
        )

        // send to DLQ after N retries
        if shouldDLQ(msg) {
            dlqProducer.WriteMessages(ctx, kafka.Message{
                Value: msg.Value,
                Headers: append(msg.Headers,
                    kafka.Header{Key: "error", Value: []byte(err.Error())},
                    kafka.Header{Key: "original-topic", Value: []byte(msg.Topic)},
                ),
            })
        }

        // commit to move forward (prevent infinite reprocessing)
        r.CommitMessages(ctx, msg)
        continue
    }

    r.CommitMessages(ctx, msg)
}
```

## Common Mistakes

| Mistake | Problem | Fix |
|---------|---------|-----|
| Auto-commit without processing | Message loss on failure | Use `FetchMessage` + manual `CommitMessages` |
| `RequiredAcks: 0` in prod | Data loss on leader failure | Use `RequireOne` or `RequireAll` |
| Not handling rebalance | Duplicate processing | Accept rebalance semantics, ensure idempotent handlers |
| Blocking in consumer loop | Partition starvation, rebalance | Process async or batch with timeout |
| Unbounded batches | Memory exhaustion | Cap batch size and timeout |
| Single consumer for multi-partition | Underutilized parallelism | Scale consumers = partitions |
| Shared `Reader` across goroutines | Race conditions, corruption | One reader per goroutine or lock |
| No compression | High network/disk usage | Enable snappy or zstd |
| Ignoring producer errors | Silent message loss | Check `WriteMessages` error, implement DLQ |
| Committing before processing | At-most-once (data loss) | Process then commit for at-least-once |

## See Also

- [RabbitMQ](./rabbitmq.md)
- [NATS](./nats.md)
- [Redis Pub/Sub](./redis-pubsub.md)
- [Error Handling](../02-language/error-handling.md)
