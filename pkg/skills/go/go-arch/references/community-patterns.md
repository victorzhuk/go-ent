## Clean Architecture Layers
```
internal/
├── domain/       # Entities, value objects, repository interfaces
├── service/      # Business logic, use cases — depends only on domain
├── handler/      # HTTP/gRPC adapters — depends on service
├── repo/         # Database implementations — depends on domain interfaces
└── infra/        # External services, message queues, caches
```
- Dependency rule: inner layers never import outer layers
- Use interfaces at layer boundaries for testability
- Domain layer has zero external dependencies

## Observability (OpenTelemetry)
- Structured logging with `slog` (stdlib): JSON in production, text in dev
- Distributed tracing with OpenTelemetry SDK
- Metrics with Prometheus exposition format
- Health endpoints: `/healthz` (liveness), `/readyz` (readiness)
- Always propagate trace context through `context.Context`

## Graceful Shutdown
```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()

srv := &http.Server{Addr: ":8080", Handler: handler}
go func() { srv.ListenAndServe() }()

<-ctx.Done()
shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
srv.Shutdown(shutCtx)
```

## Configuration
- Use environment variables (12-factor app)
- Parse with struct tags: `env:"DATABASE_URL,required"`
- Validate configuration at startup — fail fast on missing required config
- Never log secrets; redact in config dumps

## Resilience Patterns
- Circuit breaker for external service calls
- Retry with exponential backoff and jitter
- Timeouts on all external calls via context
- Bulkhead: separate goroutine pools for different services
- Rate limiting at API gateway and service level

## Docker Best Practice
```dockerfile
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=build /app /app
ENTRYPOINT ["/app"]
```
