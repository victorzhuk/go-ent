## When to Use Queues
- Decouple producers from consumers
- Handle traffic spikes (buffering)
- Async processing (email, notifications, reports)
- Event-driven architecture between services
- Reliable delivery when consumers are temporarily down

## Patterns
- **Pub/Sub**: One message to many consumers (events, notifications)
- **Work Queue**: One message to one consumer (task distribution)
- **Request-Reply**: Async RPC with correlation IDs
- **Dead Letter Queue**: Failed messages for investigation
- **Saga/Choreography**: Distributed transactions via events

## Kafka
- Topics partitioned for parallelism
- Consumer groups for load balancing
- Retain messages by time/size (log-based)
- Use Avro/Protobuf with Schema Registry
- Exactly-once with idempotent producers + transactional consumers
- Monitor: consumer lag, partition assignment, throughput

## RabbitMQ
- Exchanges route to queues (direct, topic, fanout, headers)
- Acknowledgments for reliable delivery
- Prefetch count for consumer flow control
- Use durable queues and persistent messages for reliability
- Dead letter exchanges for failed messages

## Best Practices
- Design messages to be idempotent — consumers may process duplicates
- Include correlation IDs for tracing
- Schema evolution: add optional fields, never remove or rename
- Monitor queue depth, consumer lag, processing time
- Set message TTL to prevent unbounded growth
- Use backoff and retry with exponential delays
- Log message processing failures with full context
