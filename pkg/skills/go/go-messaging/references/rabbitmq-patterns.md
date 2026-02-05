# RabbitMQ Patterns Quick Reference

Extracted from `docs/go/topics/06-messaging/rabbitmq.md` (284 lines) → 80 lines.

## Publisher

```go
import "github.com/rabbitmq/amqp091-go"

ch, _ := conn.Channel()
defer ch.Close()

ch.ExchangeDeclare("events", "topic", true, false, false, false, nil)

ch.Publish("events", "order.created", false, false, amqp091.Publishing{
    ContentType: "application/json",
    Body:        []byte(`{"id":"123"}`),
})
```

## Consumer

```go
ch, _ := conn.Channel()
q, _ := ch.QueueDeclare("orders", true, false, false, false, nil)
ch.QueueBind(q.Name, "order.*", "events", false, nil)

msgs, _ := ch.Consume(q.Name, "", false, false, false, false, nil)

for msg := range msgs {
    process(msg.Body)
    msg.Ack(false)
}
```
