# Code Documentation Index

Go code conventions and patterns for go-ent.

## Overview

This is the index for go-ent coding standards, patterns, and conventions. Each topic has its own focused document.

## Documentation

- **[Error Handling](ERROR_HANDLING.md)** - Error patterns, wrapping, sentinel errors
- **[Architecture](ARCHITECTURE.md)** - Clean Architecture layers, dependency flow
- **[Testing](TESTING.md)** - Table-driven tests, behavior-focused testing
- **[Conventions](CONVENTIONS.md)** - Naming, file organization, code style

## Production Checklist

When writing production code:

- [ ] Request/Correlation ID propagation
- [ ] Health checks: `/healthz`, `/readyz`
- [ ] Metrics: `/metrics` (Prometheus)
- [ ] Structured logging (JSON prod, slog API)
- [ ] Graceful shutdown (30s, fresh context)
- [ ] Timeouts on all external calls
- [ ] Context propagation throughout
- [ ] No panic (recover in handlers)

## Essential Libraries

- **DB:** pgx/v5, squirrel, goose/v3, clickhouse-go/v2
- **MQ:** amqp091-go, kafka-go, redis/v9
- **HTTP:** fasthttp or net/http, ogen (OpenAPI)
- **Config:** env/v11
- **Logging:** log/slog + zerolog
- **Testing:** testify, testcontainers-go
- **Utils:** uuid, decimal, validator/v10
- **Production:** prometheus/client_golang, x/time/rate, x/sync/errgroup
