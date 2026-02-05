# Constraints

- Include proper message acknowledgment after successful processing
- Include consumer group patterns for load distribution
- Include idempotency keys for at-least-once delivery
- Include dead letter queues for failed messages
- Include backoff and retry strategies
- Include graceful shutdown for consumers
- Include message serialization (JSON, Protobuf, Avro)
- Exclude committing messages before processing
- Exclude blocking indefinitely without timeout
- Exclude ignoring consumer lag monitoring
- Exclude missing error handling in consumers
- Bound to reliable event-driven systems
- Follow at-least-once delivery semantics
