# Documentation Index

Complete documentation for the go-ent project.

## Quick Links

- **[Go Enterprise Development Guide](go/INDEX.md)** - Comprehensive Go reference guide
- **[Development Guide](DEVELOPMENT.md)** - Project development workflow
- **[Architecture](ARCHITECTURE.md)** - System architecture and design

## Go Development

### [Go Enterprise Development Guide](go/INDEX.md)

Comprehensive, reference-style documentation for building production-grade Go applications.

**Covers:**
- Fundamentals (idioms, style, naming)
- Language features (generics, error handling, data structures)
- Concurrency patterns (goroutines, channels, sync primitives)
- Database integration (PostgreSQL, Redis, MongoDB)
- HTTP & gRPC services
- Messaging (RabbitMQ, Kafka, NATS)
- Observability (logging, tracing, metrics)
- Testing (table-driven, mocking, integration, fuzzing)
- CLI applications (Cobra, configuration)
- Architecture patterns (clean architecture, DI, SOLID)
- Security (validation, authentication, rate limiting)
- Performance (profiling, memory, GC tuning)
- DevOps (Docker, Kubernetes, CI/CD)

**Format:** Quick reference tables → Code examples → Common mistakes → Best practices

### Legacy Guide

- **[go/ent_dev_guide.md](go/ent_dev_guide.md)** - Original development guide (archived for reference)

## Project Documentation

### Development

- **DEVELOPMENT.md** - Development setup, workflows, and contributing guidelines
- **tools/** - Tool-specific documentation and configuration

### Architecture

- **ARCHITECTURE.md** - System design, patterns, and architectural decisions
- **research/** - Design research and prototypes
- **archive/** - Archived documentation and deprecated patterns

## Contributing

When adding new documentation:

1. **Go topics**: Add to appropriate category in `go/topics/`
2. **Project docs**: Create in `docs/` root or appropriate subdirectory
3. **Tools**: Document in `tools/` directory
4. **Update this index** when adding major documentation sections

## See Also

- [Project README](../README.md) - Project overview and quick start
- [CLAUDE.md](../CLAUDE.md) - Project-specific instructions for AI assistants
