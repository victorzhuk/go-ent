# Kafka Patterns Quick Reference

Extracted from `docs/go/topics/06-messaging/kafka.md` (293 lines) → 80 lines.

## Producer

```go
import "github.com/segmentio/kafka-go"

writer := &kafka.Writer{
    Addr:     kafka.TCP("localhost:9092"),
    Topic:    "events",
    Balancer: &kafka.LeastBytes{},
}
defer writer.Close()

err := writer.WriteMessages(ctx, kafka.Message{
    Key:   []byte("key"),
    Value: []byte("value"),
})
```

## Consumer

```go
reader := kafka.NewReader(kafka.ReaderConfig{
    Brokers: []string{"localhost:9092"},
    Topic:   "events",
    GroupID: "my-group",
})
defer reader.Close()

for {
    msg, err := reader.FetchMessage(ctx)
    if err != nil { break }

    process(msg)
    reader.CommitMessages(ctx, msg)
}
```
