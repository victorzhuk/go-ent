# Constraints

- Include slog for structured logging with context fields
- Include OpenTelemetry for distributed tracing
- Include Prometheus metrics for monitoring
- Include request/trace ID correlation across services
- Include log levels (Debug, Info, Warn, Error)
- Include sampling for high-volume traces
- Include metric labels for dimensionality
- Exclude logging sensitive data (passwords, tokens, PII)
- Exclude excessive logging in hot paths
- Exclude missing error context in logs
- Exclude hardcoded metric names (use constants)
- Bound to production-ready observability
- Follow structured logging best practices
