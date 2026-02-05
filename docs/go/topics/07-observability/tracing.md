# Tracing

Distributed tracing using OpenTelemetry (OTEL), the industry standard for observability in cloud-native applications.

## Quick Reference

| Pattern                              | Use When                          |
|--------------------------------------|-----------------------------------|
| `otel.Tracer("service-name")`        | Create tracer instance            |
| `tracer.Start(ctx, "span-name")`     | Start new span                    |
| `span.End()`                         | End span (always defer)           |
| `span.SetAttributes(attrs...)`       | Add metadata to span              |
| `span.RecordError(err)`              | Record error in span              |
| `trace.SpanFromContext(ctx)`         | Get current span from context     |
| `otelhttp.NewHandler(handler, name)` | Automatic HTTP server tracing     |
| `otelgrpc.UnaryServerInterceptor()`  | Automatic gRPC server tracing     |

## Tracer Setup

### Basic Setup with OTLP Exporter

```go
import (
    "context"
    "fmt"
    "time"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type TracingConfig struct {
    ServiceName    string
    ServiceVersion string
    Environment    string
    OTLPEndpoint   string
    SampleRate     float64
}

func InitTracing(ctx context.Context, cfg TracingConfig) (func(context.Context) error, error) {
    // Create OTLP exporter (supports Jaeger, Tempo, etc.)
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
        otlptracegrpc.WithInsecure(), // Use WithTLSCredentials() in production
    )
    if err != nil {
        return nil, fmt.Errorf("create otlp exporter: %w", err)
    }

    // Create resource with service metadata
    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName(cfg.ServiceName),
            semconv.ServiceVersion(cfg.ServiceVersion),
            semconv.DeploymentEnvironment(cfg.Environment),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("create resource: %w", err)
    }

    // Create trace provider with sampling
    provider := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(res),
        trace.WithSampler(trace.ParentBased(
            trace.TraceIDRatioBased(cfg.SampleRate),
        )),
    )

    // Set global tracer provider
    otel.SetTracerProvider(provider)

    // Return shutdown function
    return provider.Shutdown, nil
}

// Usage in main
func main() {
    ctx := context.Background()

    shutdown, err := InitTracing(ctx, TracingConfig{
        ServiceName:    "user-service",
        ServiceVersion: "1.0.0",
        Environment:    "production",
        OTLPEndpoint:   "localhost:4317",
        SampleRate:     0.1, // 10% sampling
    })
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown(ctx)

    // Application code...
}
```

### Jaeger Exporter (Alternative)

```go
import (
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/trace"
)

func InitJaegerTracing(serviceName string) (func(context.Context) error, error) {
    exporter, err := jaeger.New(
        jaeger.WithCollectorEndpoint(
            jaeger.WithEndpoint("http://localhost:14268/api/traces"),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("create jaeger exporter: %w", err)
    }

    provider := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName(serviceName),
        )),
    )

    otel.SetTracerProvider(provider)
    return provider.Shutdown, nil
}
```

## Creating Spans

### Basic Span Creation

```go
import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
)

func ProcessOrder(ctx context.Context, orderID string) error {
    tracer := otel.Tracer("order-service")

    ctx, span := tracer.Start(ctx, "ProcessOrder")
    defer span.End()

    // Add attributes
    span.SetAttributes(
        attribute.String("order.id", orderID),
        attribute.String("order.status", "processing"),
    )

    // Business logic...
    order, err := getOrder(ctx, orderID)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "failed to get order")
        return err
    }

    span.SetAttributes(attribute.Float64("order.amount", order.Amount))

    return nil
}
```

### Nested Spans

```go
func ProcessOrder(ctx context.Context, orderID string) error {
    tracer := otel.Tracer("order-service")

    ctx, span := tracer.Start(ctx, "ProcessOrder")
    defer span.End()

    // Validate order (child span)
    if err := validateOrder(ctx, orderID); err != nil {
        return err
    }

    // Process payment (child span)
    if err := processPayment(ctx, orderID); err != nil {
        return err
    }

    return nil
}

func validateOrder(ctx context.Context, orderID string) error {
    tracer := otel.Tracer("order-service")

    ctx, span := tracer.Start(ctx, "ValidateOrder")
    defer span.End()

    span.SetAttributes(attribute.String("order.id", orderID))

    // Validation logic...
    return nil
}

func processPayment(ctx context.Context, orderID string) error {
    tracer := otel.Tracer("order-service")

    ctx, span := tracer.Start(ctx, "ProcessPayment")
    defer span.End()

    span.SetAttributes(attribute.String("order.id", orderID))

    // Payment logic...
    return nil
}
```

### Span Options

```go
import (
    "go.opentelemetry.io/otel/trace"
    "go.opentelemetry.io/otel/attribute"
)

func DatabaseQuery(ctx context.Context, query string) error {
    tracer := otel.Tracer("db-client")

    ctx, span := tracer.Start(ctx, "DatabaseQuery",
        trace.WithSpanKind(trace.SpanKindClient),
        trace.WithAttributes(
            attribute.String("db.system", "postgresql"),
            attribute.String("db.statement", query),
        ),
    )
    defer span.End()

    // Execute query...
    return nil
}
```

## Context Propagation

### HTTP Client (Outgoing Requests)

```go
import (
    "net/http"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func MakeHTTPRequest(ctx context.Context, url string) error {
    // Create HTTP client with tracing
    client := http.Client{
        Transport: otelhttp.NewTransport(http.DefaultTransport),
    }

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return err
    }

    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}
```

### HTTP Server (Incoming Requests)

```go
import (
    "net/http"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func SetupHTTPServer() *http.Server {
    mux := http.NewServeMux()

    mux.HandleFunc("/orders", handleOrders)

    // Wrap handler with automatic tracing
    handler := otelhttp.NewHandler(mux, "http-server")

    return &http.Server{
        Addr:    ":8080",
        Handler: handler,
    }
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
    // Context already contains trace info from otelhttp middleware
    ctx := r.Context()

    tracer := otel.Tracer("order-service")
    ctx, span := tracer.Start(ctx, "HandleOrders")
    defer span.End()

    // Process request...
    span.SetAttributes(
        attribute.String("http.method", r.Method),
        attribute.String("http.route", r.URL.Path),
    )
}
```

### gRPC Server

```go
import (
    "google.golang.org/grpc"
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

func SetupGRPCServer() *grpc.Server {
    server := grpc.NewServer(
        grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
        grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
    )

    // Register services...
    return server
}
```

### gRPC Client

```go
import (
    "google.golang.org/grpc"
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

func SetupGRPCClient(target string) (*grpc.ClientConn, error) {
    conn, err := grpc.Dial(target,
        grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
        grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
    )
    if err != nil {
        return nil, err
    }
    return conn, nil
}
```

### Manual Context Propagation

```go
import (
    "go.opentelemetry.io/otel/propagation"
    "net/http"
)

var propagator = propagation.NewCompositeTextMapPropagator(
    propagation.TraceContext{},
    propagation.Baggage{},
)

// Inject trace context into HTTP headers
func InjectTraceContext(ctx context.Context, header http.Header) {
    propagator.Inject(ctx, propagation.HeaderCarrier(header))
}

// Extract trace context from HTTP headers
func ExtractTraceContext(ctx context.Context, header http.Header) context.Context {
    return propagator.Extract(ctx, propagation.HeaderCarrier(header))
}

// Usage
func MakeRequest(ctx context.Context, url string) error {
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

    // Inject trace context
    InjectTraceContext(ctx, req.Header)

    resp, err := http.DefaultClient.Do(req)
    // ...
    return err
}
```

## Span Attributes and Events

### Standard Attributes

```go
import (
    "go.opentelemetry.io/otel/attribute"
    semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func HTTPRequest(ctx context.Context, url string) error {
    tracer := otel.Tracer("http-client")

    ctx, span := tracer.Start(ctx, "HTTPRequest")
    defer span.End()

    // Semantic conventions
    span.SetAttributes(
        semconv.HTTPMethod("GET"),
        semconv.HTTPURL(url),
        semconv.HTTPStatusCode(200),
    )

    // Custom attributes
    span.SetAttributes(
        attribute.String("user.id", "12345"),
        attribute.Int("retry.count", 3),
        attribute.Bool("cache.hit", true),
        attribute.StringSlice("tags", []string{"important", "batch"}),
    )

    return nil
}
```

### Span Events

```go
func ProcessOrder(ctx context.Context, orderID string) error {
    tracer := otel.Tracer("order-service")

    ctx, span := tracer.Start(ctx, "ProcessOrder")
    defer span.End()

    // Add event at specific point in time
    span.AddEvent("order.validation.started")

    if err := validate(orderID); err != nil {
        span.AddEvent("order.validation.failed",
            trace.WithAttributes(
                attribute.String("error", err.Error()),
            ),
        )
        return err
    }

    span.AddEvent("order.validation.completed")

    // Add event with timestamp
    span.AddEvent("payment.processed",
        trace.WithTimestamp(time.Now()),
        trace.WithAttributes(
            attribute.Float64("amount", 99.99),
            attribute.String("currency", "USD"),
        ),
    )

    return nil
}
```

### Baggage (Cross-Service Data)

```go
import (
    "go.opentelemetry.io/otel/baggage"
)

func SetUserContext(ctx context.Context, userID, tenantID string) context.Context {
    member1, _ := baggage.NewMember("user.id", userID)
    member2, _ := baggage.NewMember("tenant.id", tenantID)

    bag, _ := baggage.New(member1, member2)
    return baggage.ContextWithBaggage(ctx, bag)
}

func GetUserFromBaggage(ctx context.Context) string {
    bag := baggage.FromContext(ctx)
    return bag.Member("user.id").Value()
}

// Usage
func HandleRequest(ctx context.Context) {
    ctx = SetUserContext(ctx, "user-123", "tenant-456")

    // Baggage is automatically propagated to downstream services
    callDownstreamService(ctx)
}
```

## Error Handling

### Recording Errors

```go
import (
    "go.opentelemetry.io/otel/codes"
)

func ProcessPayment(ctx context.Context, amount float64) error {
    tracer := otel.Tracer("payment-service")

    ctx, span := tracer.Start(ctx, "ProcessPayment")
    defer span.End()

    if amount <= 0 {
        err := fmt.Errorf("invalid amount: %f", amount)

        // Record error and set status
        span.RecordError(err)
        span.SetStatus(codes.Error, "invalid payment amount")

        return err
    }

    // Process payment...
    if err := chargeCard(ctx, amount); err != nil {
        span.RecordError(err,
            trace.WithAttributes(
                attribute.String("error.type", "payment_failed"),
                attribute.Float64("amount", amount),
            ),
        )
        span.SetStatus(codes.Error, "payment processing failed")
        return err
    }

    span.SetStatus(codes.Ok, "payment successful")
    return nil
}
```

### Error with Stack Trace

```go
import (
    "runtime/debug"
)

func ProcessWithStackTrace(ctx context.Context) error {
    tracer := otel.Tracer("service")

    ctx, span := tracer.Start(ctx, "ProcessWithStackTrace")
    defer span.End()

    err := doSomething()
    if err != nil {
        span.RecordError(err,
            trace.WithAttributes(
                attribute.String("stack", string(debug.Stack())),
            ),
        )
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    return nil
}
```

## Instrumentation

### Database (pgx)

```go
import (
    "github.com/jackc/pgx/v5/pgxpool"
    "go.opentelemetry.io/contrib/instrumentation/github.com/jackc/pgx/v5/otelpgx"
)

func SetupDatabase(ctx context.Context, connString string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(connString)
    if err != nil {
        return nil, err
    }

    // Add OpenTelemetry tracer
    config.ConnConfig.Tracer = otelpgx.NewTracer()

    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, err
    }

    return pool, nil
}
```

### Redis

```go
import (
    "github.com/redis/go-redis/v9"
    "go.opentelemetry.io/contrib/instrumentation/github.com/redis/go-redis/v9/redisotel"
)

func SetupRedis() *redis.Client {
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // Enable tracing
    if err := redisotel.InstrumentTracing(client); err != nil {
        panic(err)
    }

    return client
}
```

### MongoDB

```go
import (
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

func SetupMongo(ctx context.Context) (*mongo.Client, error) {
    opts := options.Client().
        ApplyURI("mongodb://localhost:27017").
        SetMonitor(otelmongo.NewMonitor())

    client, err := mongo.Connect(ctx, opts)
    if err != nil {
        return nil, err
    }

    return client, nil
}
```

### Custom Repository Tracing

```go
type UserRepository struct {
    pool   *pgxpool.Pool
    tracer trace.Tracer
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
    return &UserRepository{
        pool:   pool,
        tracer: otel.Tracer("user-repository"),
    }
}

func (r *UserRepository) GetUser(ctx context.Context, id string) (*User, error) {
    ctx, span := r.tracer.Start(ctx, "GetUser",
        trace.WithSpanKind(trace.SpanKindClient),
        trace.WithAttributes(
            attribute.String("user.id", id),
        ),
    )
    defer span.End()

    query := `SELECT id, name, email FROM users WHERE id = $1`

    var user User
    err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "query failed")
        return nil, err
    }

    span.SetStatus(codes.Ok, "")
    return &user, nil
}
```

## Sampling

### Sampling Strategies

```go
import (
    "go.opentelemetry.io/otel/sdk/trace"
)

// Always sample (development)
func AlwaysSample() trace.Sampler {
    return trace.AlwaysSample()
}

// Never sample
func NeverSample() trace.Sampler {
    return trace.NeverSample()
}

// Sample 10% of traces
func TraceIDRatioSample() trace.Sampler {
    return trace.TraceIDRatioBased(0.1)
}

// Parent-based sampling (respect parent decision)
func ParentBasedSample(rate float64) trace.Sampler {
    return trace.ParentBased(
        trace.TraceIDRatioBased(rate),
        trace.WithRemoteParentSampled(trace.AlwaysSample()),
        trace.WithRemoteParentNotSampled(trace.NeverSample()),
    )
}
```

### Custom Sampler

```go
type CustomSampler struct {
    errorSampler   trace.Sampler
    defaultSampler trace.Sampler
}

func (s *CustomSampler) ShouldSample(params trace.SamplingParameters) trace.SamplingResult {
    // Always sample if error occurred
    for _, attr := range params.Attributes {
        if attr.Key == "error" {
            return s.errorSampler.ShouldSample(params)
        }
    }

    // Use default sampling
    return s.defaultSampler.ShouldSample(params)
}

func (s *CustomSampler) Description() string {
    return "CustomErrorPrioritySampler"
}

// Usage
provider := trace.NewTracerProvider(
    trace.WithSampler(&CustomSampler{
        errorSampler:   trace.AlwaysSample(),
        defaultSampler: trace.TraceIDRatioBased(0.1),
    }),
)
```

### Environment-Based Sampling

```go
func NewSampler(env string) trace.Sampler {
    switch env {
    case "development":
        return trace.AlwaysSample()
    case "staging":
        return trace.TraceIDRatioBased(0.5) // 50%
    case "production":
        return trace.ParentBased(
            trace.TraceIDRatioBased(0.1), // 10%
        )
    default:
        return trace.TraceIDRatioBased(0.01) // 1%
    }
}
```

## Common Mistakes

| Mistake                              | Fix                                                  |
|--------------------------------------|------------------------------------------------------|
| Forgetting `span.End()`              | Always `defer span.End()` immediately after creation |
| Not propagating context              | Pass `ctx` from `tracer.Start()` to child functions  |
| Creating too many spans              | Trace meaningful operations, not every function      |
| Missing error recording              | Always call `span.RecordError(err)` on errors        |
| Not setting span status              | Call `span.SetStatus(codes.Error/Ok, msg)`           |
| Blocking on span export              | Use `trace.WithBatcher()` not `trace.WithSyncer()`   |
| Logging trace ID manually            | Use automatic correlation (see correlation.md)       |
| Creating spans without context       | Use `tracer.Start(ctx, ...)` not background context  |
| Not shutting down provider           | Always defer `provider.Shutdown(ctx)`                |
| Expensive attributes in hot paths    | Use sampling or conditional attributes               |

## Best Practices

```go
// ✓ Good - defer span end
ctx, span := tracer.Start(ctx, "operation")
defer span.End()

// ✗ Bad - manual end (may be skipped on error)
ctx, span := tracer.Start(ctx, "operation")
span.End()

// ✓ Good - propagate context
ctx, span := tracer.Start(ctx, "parent")
defer span.End()
childFunction(ctx) // Pass traced context

// ✗ Bad - break trace chain
ctx, span := tracer.Start(ctx, "parent")
defer span.End()
childFunction(context.Background())

// ✓ Good - record errors
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, "operation failed")
    return err
}

// ✗ Bad - ignore errors in traces
if err != nil {
    return err
}

// ✓ Good - semantic attributes
span.SetAttributes(
    semconv.HTTPMethod("GET"),
    semconv.HTTPStatusCode(200),
)

// ✗ Bad - non-standard attributes
span.SetAttributes(
    attribute.String("method", "GET"),
    attribute.Int("code", 200),
)

// ✓ Good - meaningful span names
tracer.Start(ctx, "ProcessPayment")
tracer.Start(ctx, "ValidateUser")

// ✗ Bad - generic span names
tracer.Start(ctx, "Operation")
tracer.Start(ctx, "DoSomething")
```

## See Also

- [Logging](./logging.md) - Structured logging with trace correlation
- [Metrics](./metrics.md) - Prometheus metrics with trace exemplars
- [Correlation](./correlation.md) - Request ID and trace ID propagation
- [HTTP Server](../05-http-grpc/http-server.md) - HTTP middleware patterns
- [gRPC](../05-http-grpc/grpc.md) - gRPC interceptor patterns
- [PostgreSQL](../04-database/postgresql.md) - Database tracing
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/)
