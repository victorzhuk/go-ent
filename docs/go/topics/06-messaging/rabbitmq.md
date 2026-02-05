# RabbitMQ

Message queue patterns with amqp091-go.

## Quick Reference

| Operation | Code |
|-----------|------|
| Connection | `amqp091.Dial("amqp://user:pass@host:5672/")` |
| Channel | `conn.Channel()` |
| Publish | `ch.Publish(exchange, key, false, false, amqp091.Publishing{...})` |
| Consume | `ch.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, nil)` |
| Declare Exchange | `ch.ExchangeDeclare(name, "topic", true, false, false, false, nil)` |
| Declare Queue | `ch.QueueDeclare(name, true, false, false, false, nil)` |
| Bind Queue | `ch.QueueBind(queue, key, exchange, false, nil)` |
| Set QoS | `ch.Qos(prefetchCount, 0, false)` |
| Manual Ack | `msg.Ack(false)` |
| Nack + Requeue | `msg.Nack(false, true)` |
| DLX Setup | `args["x-dead-letter-exchange"] = dlxName` |

```go
import "github.com/rabbitmq/amqp091-go"

conn, _ := amqp091.Dial("amqp://guest:guest@localhost:5672/")
ch, _ := conn.Channel()

// Declare exchange
ch.ExchangeDeclare("events", "topic", true, false, false, false, nil)

// Publish
ch.Publish("events", "order.created", false, false, amqp091.Publishing{
    ContentType:  "application/json",
    DeliveryMode: amqp091.Persistent,
    Body:         []byte(message),
})

// Consume with manual ack
ch.Qos(10, 0, false)
msgs, _ := ch.Consume("queue_name", "", false, false, false, false, nil)
for msg := range msgs {
    if err := process(msg.Body); err != nil {
        msg.Nack(false, true)
    } else {
        msg.Ack(false)
    }
}
```

## Exchange and Queue Setup

```go
// Declare topic exchange
err := ch.ExchangeDeclare(
    "events",    // name
    "topic",     // type: fanout, topic, direct, headers
    true,        // durable
    false,       // auto-delete
    false,       // internal
    false,       // no-wait
    nil,         // args
)

// Declare durable queue
queue, err := ch.QueueDeclare(
    "orders",    // name
    true,        // durable
    false,       // auto-delete
    false,       // exclusive
    false,       // no-wait
    nil,         // args
)

// Bind queue to exchange with routing key
err = ch.QueueBind(
    queue.Name,           // queue
    "order.*",            // routing key (topic pattern)
    "events",             // exchange
    false,                // no-wait
    nil,                  // args
)
```

## Dead Letter Exchange

```go
func setupQueueWithDLX(ch *amqp091.Channel, queueName, dlxName string) error {
    // Declare DLX
    if err := ch.ExchangeDeclare(dlxName, "fanout", true, false, false, false, nil); err != nil {
        return fmt.Errorf("declare dlx: %w", err)
    }

    // Declare DLQ
    dlq := queueName + ".dlq"
    if _, err := ch.QueueDeclare(dlq, true, false, false, false, nil); err != nil {
        return fmt.Errorf("declare dlq: %w", err)
    }

    // Bind DLQ to DLX
    if err := ch.QueueBind(dlq, "", dlxName, false, nil); err != nil {
        return fmt.Errorf("bind dlq: %w", err)
    }

    // Declare main queue with DLX and message TTL
    args := amqp091.Table{
        "x-dead-letter-exchange": dlxName,
        "x-message-ttl":          300000, // 5 minutes
    }
    _, err := ch.QueueDeclare(queueName, true, false, false, false, args)
    return err
}

// Retry pattern: reject without requeue sends to DLX
func handleWithRetry(msg amqp091.Delivery, maxRetries int) error {
    retries := 0
    if msg.Headers != nil {
        if r, ok := msg.Headers["x-retry-count"].(int32); ok {
            retries = int(r)
        }
    }

    if err := process(msg.Body); err != nil {
        if retries >= maxRetries {
            msg.Reject(false) // Send to DLX
            return fmt.Errorf("max retries exceeded: %w", err)
        }

        // Republish with incremented counter
        msg.Nack(false, false)
        return nil
    }

    return msg.Ack(false)
}
```

## Acknowledgment Modes

```go
// Auto-ack (dangerous in production - lose messages on crash)
msgs, _ := ch.Consume(queue, "", true, false, false, false, nil)

// Manual ack with QoS prefetch
ch.Qos(10, 0, false) // Prefetch 10 messages

msgs, _ := ch.Consume(queue, "", false, false, false, false, nil)
for msg := range msgs {
    if err := process(msg.Body); err != nil {
        // Nack with requeue
        msg.Nack(false, true)
        continue
    }
    // Ack single message
    msg.Ack(false)
}

// Reject single message (no requeue = send to DLX)
msg.Reject(false)

// Nack multiple messages
msg.Nack(true, false) // multiple=true, requeue=false
```

**QoS Prefetch:**

```go
// Limit unacked messages per consumer (fair dispatch)
ch.Qos(
    10,    // prefetch count (0 = unlimited, bad)
    0,     // prefetch size (0 = no limit)
    false, // global (false = per consumer)
)
```

## Reconnection

```go
type Consumer struct {
    url      string
    conn     *amqp091.Connection
    ch       *amqp091.Channel
    closeCh  chan *amqp091.Error
    done     chan struct{}
}

func (c *Consumer) connect() error {
    conn, err := amqp091.Dial(c.url)
    if err != nil {
        return fmt.Errorf("dial: %w", err)
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return fmt.Errorf("channel: %w", err)
    }

    c.conn = conn
    c.ch = ch
    c.closeCh = make(chan *amqp091.Error)
    ch.NotifyClose(c.closeCh)

    // Redeclare topology
    if err := c.setupTopology(); err != nil {
        ch.Close()
        conn.Close()
        return fmt.Errorf("topology: %w", err)
    }

    return nil
}

func (c *Consumer) Run(ctx context.Context) error {
    backoff := time.Second
    maxBackoff := 30 * time.Second

    for {
        if err := c.connect(); err != nil {
            log.Error("connect failed", "error", err, "backoff", backoff)
            select {
            case <-time.After(backoff):
                backoff = min(backoff*2, maxBackoff)
                continue
            case <-ctx.Done():
                return ctx.Err()
            }
        }

        backoff = time.Second
        log.Info("connected to rabbitmq")

        if err := c.consume(ctx); err != nil {
            log.Error("consume failed", "error", err)
        }

        select {
        case <-ctx.Done():
            c.Close()
            return ctx.Err()
        case <-c.closeCh:
            log.Warn("connection closed, reconnecting")
        }
    }
}

func (c *Consumer) setupTopology() error {
    if err := c.ch.ExchangeDeclare("events", "topic", true, false, false, false, nil); err != nil {
        return err
    }
    if _, err := c.ch.QueueDeclare("orders", true, false, false, false, nil); err != nil {
        return err
    }
    return c.ch.QueueBind("orders", "order.*", "events", false, nil)
}
```

## Producer Pattern

```go
type Producer struct {
    conn *amqp091.Connection
    ch   *amqp091.Channel
}

func NewProducer(url string) (*Producer, error) {
    conn, err := amqp091.Dial(url)
    if err != nil {
        return nil, err
    }

    ch, err := conn.Channel()
    if err != nil {
        return nil, err
    }

    return &Producer{conn: conn, ch: ch}, nil
}

func (p *Producer) Publish(exchange, key string, body []byte) error {
    return p.ch.Publish(exchange, key, false, false, amqp091.Publishing{
        ContentType:  "application/json",
        Body:         body,
        DeliveryMode: amqp091.Persistent,
    })
}
```

## Consumer Pattern

```go
func (c *Consumer) Consume(queue string, handler func([]byte) error) error {
    msgs, err := c.ch.Consume(queue, "", false, false, false, false, nil)
    if err != nil {
        return err
    }

    for msg := range msgs {
        if err := handler(msg.Body); err != nil {
            msg.Nack(false, true) // Requeue
        } else {
            msg.Ack(false)
        }
    }

    return nil
}
```

## Common Mistakes

| Mistake | Problem | Solution |
|---------|---------|----------|
| Auto-ack in production | Message loss on crash | Use manual ack with `msg.Ack(false)` |
| Unbounded prefetch | Memory exhaustion | Set `ch.Qos(10, 0, false)` |
| No reconnection logic | Service dies on disconnect | Implement automatic reconnection with backoff |
| Blocking in consumer | Starves other messages | Process async or use worker pool |
| Wrong exchange type | Messages not routed | Use topic for patterns, direct for exact keys |
| Forgetting to redeclare topology | Consumer fails after reconnect | Redeclare exchanges/queues/bindings on reconnect |
| Not using DLX | Poison messages block queue | Setup DLX for failed messages |
| Requeue on every error | Infinite retry loop | Limit retries, use DLX + TTL |

## See Also

- [Kafka](./kafka.md)
- [NATS](./nats.md)
- [Redis Pub/Sub](./redis-pubsub.md)
- [Error Handling](../02-language/error-handling.md)
