# Go Enterprise Development Guide

Comprehensive, reference-style documentation for building production-grade Go applications.

## Navigation

### 01. Fundamentals
Core Go idioms and style conventions.

- [**Idioms**](topics/01-fundamentals/idioms.md) - Effective Go principles and Go proverbs
- [**Style Guide**](topics/01-fundamentals/style-guide.md) - Uber + Google Go Style consensus
- [**Naming Conventions**](topics/01-fundamentals/naming.md) - Variables, types, packages

### 02. Language Features
Modern Go language capabilities (1.22+).

- [**Generics**](topics/02-language/generics.md) - Type constraints, iter.Seq, slices/maps packages
- [**Data Structures**](topics/02-language/data-structures.md) - Slices, maps, Swiss Tables, container/*
- [**Error Handling**](topics/02-language/error-handling.md) - Wrapping, Is/As, custom types, multi-error

### 03. Concurrency
Goroutines, channels, and synchronization patterns.

- [**Goroutines**](topics/03-concurrency/goroutines.md) - Lifecycle management and leak prevention
- [**Channels**](topics/03-concurrency/channels.md) - Channel axioms and patterns
- [**Sync Primitives**](topics/03-concurrency/sync-primitives.md) - Mutex, WaitGroup, Once, Pool
- [**Patterns**](topics/03-concurrency/patterns.md) - Pipeline, fan-in/out, worker pools

### 04. Database
PostgreSQL, Redis, MongoDB integration.

- [**PostgreSQL**](topics/04-database/postgresql.md) - pgx/v5, squirrel, transactions
- [**Redis**](topics/04-database/redis.md) - go-redis/v9, pub/sub, streams
- [**MongoDB**](topics/04-database/mongodb.md) - mongo-driver, BSON, aggregations
- [**Migrations**](topics/04-database/migrations.md) - goose migration patterns

### 05. HTTP & gRPC
Building and consuming APIs.

- [**HTTP Server**](topics/05-http-grpc/http-server.md) - ServeMux 1.22+, middleware, graceful shutdown
- [**HTTP Client**](topics/05-http-grpc/http-client.md) - Transport, connection pooling, timeouts
- [**gRPC**](topics/05-http-grpc/grpc.md) - Interceptors, keepalive, otelgrpc
- [**OpenAPI**](topics/05-http-grpc/openapi.md) - ogen code generation patterns

### 06. Messaging
Message queues and event streaming.

- [**RabbitMQ**](topics/06-messaging/rabbitmq.md) - amqp091-go patterns
- [**Kafka**](topics/06-messaging/kafka.md) - kafka-go producer/consumer patterns
- [**Redis Pub/Sub**](topics/06-messaging/redis-pubsub.md) - Pub/Sub vs Streams comparison
- [**NATS**](topics/06-messaging/nats.md) - NATS and JetStream patterns

### 07. Observability
Logging, tracing, metrics, and correlation.

- [**Logging**](topics/07-observability/logging.md) - slog + zerolog integration
- [**Tracing**](topics/07-observability/tracing.md) - OpenTelemetry setup and propagation
- [**Metrics**](topics/07-observability/metrics.md) - Prometheus client_golang patterns
- [**Correlation**](topics/07-observability/correlation.md) - Request/Correlation ID propagation

### 08. Testing
Unit, integration, fuzz, and benchmark testing.

- [**Table-Driven Tests**](topics/08-testing/table-driven.md) - t.Run, parallel, testify
- [**Mocking**](topics/08-testing/mocking.md) - mockery, manual mocks, when to mock
- [**Integration Tests**](topics/08-testing/integration.md) - testcontainers patterns
- [**Fuzzing**](topics/08-testing/fuzzing.md) - Go 1.18+ fuzzing with examples
- [**Benchmarks**](topics/08-testing/benchmarks.md) - testing.B, benchstat, profiling

### 09. CLI & Configuration
Command-line applications and configuration management.

- [**Cobra**](topics/09-cli-config/cobra.md) - Command structure and subcommands
- [**Configuration**](topics/09-cli-config/configuration.md) - caarlos0/env patterns
- [**Functional Options**](topics/09-cli-config/functional-options.md) - Dave Cheney pattern
- [**Interactive CLI**](topics/09-cli-config/interactive.md) - huh (survey replacement)

### 10. Architecture
Project layout, clean architecture, dependency injection.

- [**Project Layout**](topics/10-architecture/project-layout.md) - Directory structure and packages
- [**Clean Architecture**](topics/10-architecture/clean-architecture.md) - Layers and boundaries
- [**Dependency Injection**](topics/10-architecture/dependency-injection.md) - Manual, Wire, Fx
- [**SOLID Principles**](topics/10-architecture/solid.md) - SOLID in Go context

### 11. Security
Input validation, authentication, rate limiting.

- [**Input Validation**](topics/11-security/input-validation.md) - go-playground/validator
- [**Authentication**](topics/11-security/authentication.md) - JWT, bcrypt patterns
- [**Rate Limiting**](topics/11-security/rate-limiting.md) - x/time/rate patterns
- [**Security Headers**](topics/11-security/security-headers.md) - Middleware, CORS

### 12. Performance
Profiling, memory optimization, GC tuning.

- [**Profiling**](topics/12-performance/profiling.md) - pprof methodology
- [**Memory Optimization**](topics/12-performance/memory.md) - Escape analysis, sync.Pool
- [**GC Tuning**](topics/12-performance/gc-tuning.md) - GOGC, GOMEMLIMIT
- [**Connection Pools**](topics/12-performance/connection-pools.md) - DB, HTTP, gRPC pooling

### 13. DevOps
Containerization, deployment, tooling.

- [**Docker**](topics/13-devops/docker.md) - Multi-stage builds, distroless images
- [**Kubernetes**](topics/13-devops/kubernetes.md) - Probes, automaxprocs
- [**Linting**](topics/13-devops/linting.md) - golangci-lint configuration
- [**CI/CD**](topics/13-devops/ci-cd.md) - GitHub Actions patterns

---

## Quick Start

### New to Go?
Start here in order:
1. [Idioms](topics/01-fundamentals/idioms.md)
2. [Style Guide](topics/01-fundamentals/style-guide.md)
3. [Error Handling](topics/02-language/error-handling.md)
4. [Goroutines](topics/03-concurrency/goroutines.md)
5. [Table-Driven Tests](topics/08-testing/table-driven.md)

### Building a Web API?
1. [HTTP Server](topics/05-http-grpc/http-server.md)
2. [PostgreSQL](topics/04-database/postgresql.md)
3. [Logging](topics/07-observability/logging.md)
4. [Clean Architecture](topics/10-architecture/clean-architecture.md)
5. [Integration Tests](topics/08-testing/integration.md)

### Microservices?
1. [gRPC](topics/05-http-grpc/grpc.md)
2. [Kafka](topics/06-messaging/kafka.md)
3. [Tracing](topics/07-observability/tracing.md)
4. [Project Layout](topics/10-architecture/project-layout.md)
5. [Kubernetes](topics/13-devops/kubernetes.md)

---

## Library Reference

### Core Dependencies

| Category | Library | Version | Guide |
|----------|---------|---------|-------|
| **Database** | pgx | v5 | [PostgreSQL](topics/04-database/postgresql.md) |
| | go-redis | v9 | [Redis](topics/04-database/redis.md) |
| | mongo-driver | v1 | [MongoDB](topics/04-database/mongodb.md) |
| | squirrel | latest | [PostgreSQL](topics/04-database/postgresql.md) |
| | goose | v3 | [Migrations](topics/04-database/migrations.md) |
| **Messaging** | amqp091-go | latest | [RabbitMQ](topics/06-messaging/rabbitmq.md) |
| | kafka-go | latest | [Kafka](topics/06-messaging/kafka.md) |
| | nats.go | latest | [NATS](topics/06-messaging/nats.md) |
| **HTTP/gRPC** | net/http | stdlib | [HTTP Server](topics/05-http-grpc/http-server.md) |
| | grpc-go | latest | [gRPC](topics/05-http-grpc/grpc.md) |
| | ogen | latest | [OpenAPI](topics/05-http-grpc/openapi.md) |
| **Observability** | log/slog | stdlib | [Logging](topics/07-observability/logging.md) |
| | zerolog | latest | [Logging](topics/07-observability/logging.md) |
| | otel | latest | [Tracing](topics/07-observability/tracing.md) |
| | client_golang | latest | [Metrics](topics/07-observability/metrics.md) |
| **Testing** | testify | latest | [Table-Driven](topics/08-testing/table-driven.md) |
| | mockery | latest | [Mocking](topics/08-testing/mocking.md) |
| | testcontainers-go | latest | [Integration](topics/08-testing/integration.md) |
| **CLI** | cobra | latest | [Cobra](topics/09-cli-config/cobra.md) |
| | env | v11 | [Configuration](topics/09-cli-config/configuration.md) |
| | huh | latest | [Interactive](topics/09-cli-config/interactive.md) |
| **Security** | validator | v10 | [Validation](topics/11-security/input-validation.md) |
| | jwt | v5 | [Auth](topics/11-security/authentication.md) |
| | bcrypt | stdlib | [Auth](topics/11-security/authentication.md) |
| **Tools** | golangci-lint | latest | [Linting](topics/13-devops/linting.md) |
| | gofumpt | latest | [Style Guide](topics/01-fundamentals/style-guide.md) |
| | automaxprocs | latest | [Kubernetes](topics/13-devops/kubernetes.md) |

---

## Format

Each guide follows a reference-style format:

```markdown
# Topic Title
Brief description of when/why to use this.

## Quick Reference
| Pattern | Use When |
|---------|----------|

## Code Examples
Practical patterns with explanations.

## Common Mistakes
| Mistake | Fix |
|---------|-----|

## See Also
[Related topics]
```

Optimized for quick lookup, not sequential reading.

---

## Version Notes

- **Go Version**: 1.25+ baseline (examples may use 1.24/1.23 features with version notes)
- **Swiss Tables**: Go 1.24+ (marked where applicable)
- **b.Loop()**: Go 1.24+ (replaces classic b.N loop in benchmarks)
- **Archived Libraries**: survey/v2 (use huh instead)

---

## See Also

- [Go Official Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Proverbs](https://go-proverbs.github.io/)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Google Go Style Guide](https://google.github.io/styleguide/go/)
