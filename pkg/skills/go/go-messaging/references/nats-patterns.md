# NATS Patterns Quick Reference

Extracted from `docs/go/topics/06-messaging/nats.md` (282 lines) → 70 lines.

## Pub/Sub

```go
import "github.com/nats-io/nats.go"

nc, _ := nats.Connect(nats.DefaultURL)
defer nc.Close()

// Publish
nc.Publish("events", []byte("message"))

// Subscribe
nc.Subscribe("events", func(msg *nats.Msg) {
    process(msg.Data)
})
```

## JetStream (Persistence)

```go
js, _ := nc.JetStream()

// Create stream
js.AddStream(&nats.StreamConfig{
    Name:     "EVENTS",
    Subjects: []string{"events.*"},
})

// Publish
js.Publish("events.order", []byte("data"))

// Consume
js.Subscribe("events.*", func(msg *nats.Msg) {
    process(msg.Data)
    msg.Ack()
})
```
