# NATS

Lightweight messaging with NATS and JetStream.

## Quick Reference

| Operation | Code |
|-----------|------|
| Connect | `nc, _ := nats.Connect("nats://localhost:4222", nats.Name("service"))` |
| Publish | `nc.Publish("subject", []byte("data"))` |
| Subscribe | `nc.Subscribe("subject", handler)` |
| Queue Group | `nc.QueueSubscribe("subject", "workers", handler)` |
| Request-Reply | `msg, _ := nc.Request("subject", data, time.Second)` |
| JetStream | `js, _ := nc.JetStream()` |
| Stream Create | `js.AddStream(&nats.StreamConfig{Name: "EVENTS", Subjects: []string{"events.>"}}})` |
| Consumer Push | `js.Subscribe("events.created", handler)` |
| Consumer Pull | `js.PullSubscribe("events.created", "durable")` |
| Acknowledge | `msg.Ack()` |
| Drain | `nc.Drain()` |
| Reconnect | `nats.MaxReconnects(10), nats.ReconnectWait(2*time.Second)` |

## Basic Pub/Sub

```go
func NewPublisher(url string) (*nats.Conn, error) {
	nc, err := nats.Connect(url,
		nats.Name("publisher"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}
	return nc, nil
}

func publish(nc *nats.Conn, subject string, data []byte) error {
	if err := nc.Publish(subject, data); err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}
	return nil
}

func subscribe(nc *nats.Conn, subject string, handler nats.MsgHandler) error {
	sub, err := nc.Subscribe(subject, handler)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", subject, err)
	}
	_ = sub
	return nil
}
```

## JetStream

JetStream provides message persistence, exactly-once delivery, and replay capabilities.

### Enable JetStream

```go
func NewJetStream(nc *nats.Conn) (nats.JetStreamContext, error) {
	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		return nil, fmt.Errorf("create jetstream context: %w", err)
	}
	return js, nil
}
```

### Stream Creation

```go
func createStream(js nats.JetStreamContext, name string, subjects []string) error {
	cfg := &nats.StreamConfig{
		Name:     name,
		Subjects: subjects,
		MaxAge:   24 * time.Hour,
		Storage:  nats.FileStorage,
		Replicas: 1,
	}

	_, err := js.AddStream(cfg)
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		return fmt.Errorf("add stream %s: %w", name, err)
	}
	return nil
}
```

### Push Consumer

```go
func subscribePush(js nats.JetStreamContext, subject, durable string) error {
	_, err := js.Subscribe(subject, func(msg *nats.Msg) {
		if err := process(msg.Data); err != nil {
			msg.Nak()
			return
		}
		msg.Ack()
	}, nats.Durable(durable), nats.ManualAck())

	if err != nil {
		return fmt.Errorf("subscribe push %s: %w", subject, err)
	}
	return nil
}
```

### Pull Consumer

```go
func processPull(ctx context.Context, js nats.JetStreamContext, subject, durable string) error {
	sub, err := js.PullSubscribe(subject, durable, nats.ManualAck())
	if err != nil {
		return fmt.Errorf("pull subscribe: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msgs, err := sub.Fetch(10, nats.MaxWait(5*time.Second))
		if err != nil && err != nats.ErrTimeout {
			return fmt.Errorf("fetch messages: %w", err)
		}

		for _, msg := range msgs {
			if err := process(msg.Data); err != nil {
				msg.Nak()
				continue
			}
			msg.Ack()
		}
	}
}
```

### Acknowledgment

```go
msg.Ack()                                // Acknowledge successful processing
msg.Nak()                                // Negative ack, redeliver
msg.NakWithDelay(5 * time.Second)        // Redeliver after delay
msg.Term()                               // Terminate, don't redeliver
msg.InProgress()                         // Extend ack deadline
```

## Request-Reply Pattern

Synchronous request-response over NATS.

```go
func request(nc *nats.Conn, subject string, data []byte) ([]byte, error) {
	msg, err := nc.Request(subject, data, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", subject, err)
	}
	return msg.Data, nil
}

func respondToRequests(nc *nats.Conn, subject string) error {
	_, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		result := process(msg.Data)
		msg.Respond(result)
	})
	if err != nil {
		return fmt.Errorf("subscribe responder: %w", err)
	}
	return nil
}
```

## Queue Groups

Load balance work across multiple subscribers with unique queue names per service.

```go
func subscribeQueue(nc *nats.Conn, subject, queueName string) error {
	_, err := nc.QueueSubscribe(subject, queueName, func(msg *nats.Msg) {
		process(msg.Data)
	})
	if err != nil {
		return fmt.Errorf("queue subscribe: %w", err)
	}
	return nil
}

// Each service instance joins the same queue group
subscribeQueue(nc, "tasks.process", "task-workers")
```

## Reconnection

NATS client automatically reconnects on connection loss.

```go
func connectWithReconnect(url string) (*nats.Conn, error) {
	nc, err := nats.Connect(url,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.ReconnectJitter(500*time.Millisecond, 2*time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			slog.Warn("nats disconnected", "error", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			slog.Info("nats reconnected", "url", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			slog.Error("nats connection closed", "error", nc.LastError())
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	return nc, nil
}
```

## Durable Subscriptions

JetStream durable consumers resume from last acknowledged message.

```go
func subscribeDurable(js nats.JetStreamContext, stream, subject, durable string) error {
	_, err := js.Subscribe(subject, func(msg *nats.Msg) {
		if err := process(msg.Data); err != nil {
			msg.Nak()
			return
		}
		msg.Ack()
	},
		nats.Durable(durable),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.DeliverAll(),
	)

	if err != nil {
		return fmt.Errorf("subscribe durable: %w", err)
	}
	return nil
}
```

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Not handling disconnects | Use reconnect handlers and check `nc.IsConnected()` |
| Blocking in callback | Spawn goroutine or use buffered channel |
| Forgetting to drain | Call `nc.Drain()` before shutdown for graceful close |
| Wrong subject pattern | Use `>` for wildcard (`events.>`), not regex |
| Missing JetStream ack | Always call `msg.Ack()` or `msg.Nak()` with `ManualAck()` |
| Not checking `ErrTimeout` | Pull consumers return `nats.ErrTimeout`, not fatal |
| Synchronous publish in hot path | Use `js.PublishAsync()` for high throughput |

## See Also

- [RabbitMQ](./rabbitmq.md)
- [Kafka](./kafka.md)
- [Redis Pub/Sub](./redis-pubsub.md)
- [Error Handling](../02-language/error-handling.md)
