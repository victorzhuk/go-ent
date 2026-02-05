# Request Correlation

Request correlation enables tracking a single request across multiple services, logs, traces, and metrics. A correlation ID (also called request ID or trace ID) is a unique identifier that flows through the entire request lifecycle, making debugging distributed systems practical.

## Quick Reference

| Operation | Code |
|-----------|------|
| Context with correlation ID | `ctx = correlation.WithID(ctx, id)` |
| Extract correlation ID | `id := correlation.FromContext(ctx)` |
| slog with correlation | `slog.InfoContext(ctx, "msg", "request_id", id)` |
| HTTP header injection | `r.Header.Set("X-Request-ID", id)` |
| HTTP header extraction | `id := r.Header.Get("X-Request-ID")` |
| Span with correlation | `span.SetAttributes(attribute.String("request.id", id))` |

---

## Request ID Generation

### UUID v7

UUID v7 provides time-ordered unique identifiers with millisecond precision. Preferred for most use cases.

```go
package correlation

import (
	"github.com/google/uuid"
)

// GenerateID creates a new UUID v7 correlation ID.
// UUID v7 is time-ordered, making logs naturally sortable.
func GenerateID() string {
	return uuid.Must(uuid.NewV7()).String()
}
```

### ULID

ULID (Universally Unique Lexicographically Sortable Identifier) offers better sortability than UUID v7.

```go
import (
	"github.com/oklog/ulid/v2"
	"math/rand"
	"time"
)

var entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)

func GenerateULID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}
```

### Custom Format

For specific requirements (e.g., including service prefix, timestamp encoding).

```go
import (
	"fmt"
	"time"
)

// GenerateCustomID creates format: service-timestamp-random
// Example: api-20260205143022-a1b2c3d4
func GenerateCustomID(serviceName string) string {
	timestamp := time.Now().Format("20060102150405")
	random := uuid.Must(uuid.NewV7()).String()[:8]
	return fmt.Sprintf("%s-%s-%s", serviceName, timestamp, random)
}
```

---

## Context Propagation

### Typed Context Keys

Always use unexported struct types for context keys to avoid collisions.

```go
package correlation

import (
	"context"
)

// contextKey is an unexported type for correlation context keys.
// This prevents collisions with other packages using context.
type contextKey struct{}

var (
	requestIDKey     = contextKey{}
	correlationIDKey = contextKey{}
)

// WithRequestID adds request ID to context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID extracts request ID from context.
// Returns empty string if not present.
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// WithCorrelationID adds correlation ID to context.
// Used when request spans multiple user-initiated operations.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey, id)
}

// CorrelationID extracts correlation ID from context.
func CorrelationID(ctx context.Context) string {
	id, _ := ctx.Value(correlationIDKey).(string)
	return id
}
```

### Custom Context Package

Production pattern for correlation context management.

```go
package correlation

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type Context struct {
	RequestID     string
	CorrelationID string
	UserID        string
	SessionID     string
}

type ctxKey struct{}

var key = ctxKey{}

// WithContext adds correlation context to ctx.
func WithContext(ctx context.Context, c *Context) context.Context {
	return context.WithValue(ctx, key, c)
}

// FromContext extracts correlation context.
// Returns zero-value Context if not present.
func FromContext(ctx context.Context) *Context {
	c, ok := ctx.Value(key).(*Context)
	if !ok {
		return &Context{}
	}
	return c
}

// New creates context with generated request ID.
func New(ctx context.Context) context.Context {
	return WithContext(ctx, &Context{
		RequestID: uuid.Must(uuid.NewV7()).String(),
	})
}

// LogAttrs returns slog attributes for correlation context.
func (c *Context) LogAttrs() []slog.Attr {
	attrs := make([]slog.Attr, 0, 4)
	if c.RequestID != "" {
		attrs = append(attrs, slog.String("request_id", c.RequestID))
	}
	if c.CorrelationID != "" {
		attrs = append(attrs, slog.String("correlation_id", c.CorrelationID))
	}
	if c.UserID != "" {
		attrs = append(attrs, slog.String("user_id", c.UserID))
	}
	if c.SessionID != "" {
		attrs = append(attrs, slog.String("session_id", c.SessionID))
	}
	return attrs
}
```

---

## HTTP Integration

### Middleware

Extract or generate correlation IDs from HTTP headers.

```go
package middleware

import (
	"net/http"

	"github.com/yourorg/project/internal/correlation"
)

const (
	HeaderRequestID     = "X-Request-ID"
	HeaderCorrelationID = "X-Correlation-ID"
)

// Correlation adds correlation IDs to request context and response headers.
func Correlation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract or generate request ID
		requestID := r.Header.Get(HeaderRequestID)
		if requestID == "" {
			requestID = correlation.GenerateID()
		}

		// Extract correlation ID (optional)
		correlationID := r.Header.Get(HeaderCorrelationID)

		// Add to context
		ctx := correlation.WithContext(r.Context(), &correlation.Context{
			RequestID:     requestID,
			CorrelationID: correlationID,
		})

		// Echo back in response headers
		w.Header().Set(HeaderRequestID, requestID)
		if correlationID != "" {
			w.Header().Set(HeaderCorrelationID, correlationID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

### HTTP Client

Inject correlation IDs into outgoing requests.

```go
package httpclient

import (
	"context"
	"net/http"

	"github.com/yourorg/project/internal/correlation"
)

// Transport wraps http.RoundTripper to inject correlation headers.
type Transport struct {
	Base http.RoundTripper
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	c := correlation.FromContext(req.Context())

	if c.RequestID != "" {
		req.Header.Set("X-Request-ID", c.RequestID)
	}
	if c.CorrelationID != "" {
		req.Header.Set("X-Correlation-ID", c.CorrelationID)
	}

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

// NewClient creates HTTP client with correlation injection.
func NewClient() *http.Client {
	return &http.Client{
		Transport: &Transport{},
	}
}
```

---

## Structured Logging

### slog with Correlation

Always log correlation IDs for request traceability.

```go
package handler

import (
	"log/slog"
	"net/http"

	"github.com/yourorg/project/internal/correlation"
)

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := correlation.FromContext(r.Context())

	// Log with correlation attributes
	slog.InfoContext(r.Context(), "handling request",
		slog.String("request_id", c.RequestID),
		slog.String("correlation_id", c.CorrelationID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)

	// Or use helper
	slog.InfoContext(r.Context(), "handling request",
		slog.Group("request", c.LogAttrs()...),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
}
```

### Custom Handler

Automatically inject correlation IDs into all logs.

```go
package logging

import (
	"context"
	"log/slog"

	"github.com/yourorg/project/internal/correlation"
)

// Handler wraps slog.Handler to inject correlation IDs.
type Handler struct {
	base slog.Handler
}

func NewHandler(base slog.Handler) *Handler {
	return &Handler{base: base}
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	c := correlation.FromContext(ctx)

	if c.RequestID != "" {
		r.AddAttrs(slog.String("request_id", c.RequestID))
	}
	if c.CorrelationID != "" {
		r.AddAttrs(slog.String("correlation_id", c.CorrelationID))
	}
	if c.UserID != "" {
		r.AddAttrs(slog.String("user_id", c.UserID))
	}

	return h.base.Handle(ctx, r)
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{base: h.base.WithAttrs(attrs)}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{base: h.base.WithGroup(name)}
}
```

---

## Trace Correlation

### OpenTelemetry Integration

Link correlation IDs with trace and span IDs.

```go
package tracing

import (
	"context"

	"github.com/yourorg/project/internal/correlation"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan creates span with correlation attributes.
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	tracer := otel.Tracer("app")
	ctx, span := tracer.Start(ctx, name)

	c := correlation.FromContext(ctx)
	if c.RequestID != "" {
		span.SetAttributes(attribute.String("request.id", c.RequestID))
	}
	if c.CorrelationID != "" {
		span.SetAttributes(attribute.String("correlation.id", c.CorrelationID))
	}

	return ctx, span
}

// ExtractTraceID gets OpenTelemetry trace ID from context.
// Useful for logging trace_id alongside request_id.
func ExtractTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}
	return span.SpanContext().TraceID().String()
}
```

### Unified Correlation

Use trace ID as correlation ID when using OpenTelemetry.

```go
package correlation

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

// NewWithTrace creates correlation context using trace ID if available.
func NewWithTrace(ctx context.Context) context.Context {
	var requestID, correlationID string

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		// Use trace ID as correlation ID
		correlationID = span.SpanContext().TraceID().String()
	}

	requestID = uuid.Must(uuid.NewV7()).String()

	return WithContext(ctx, &Context{
		RequestID:     requestID,
		CorrelationID: correlationID,
	})
}
```

---

## gRPC Integration

### Server Interceptor

Extract correlation IDs from gRPC metadata.

```go
package grpcmw

import (
	"context"

	"github.com/yourorg/project/internal/correlation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryCorrelation extracts correlation IDs from gRPC metadata.
func UnaryCorrelation() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		requestID := getFirst(md, "x-request-id")
		if requestID == "" {
			requestID = correlation.GenerateID()
		}

		correlationID := getFirst(md, "x-correlation-id")

		ctx = correlation.WithContext(ctx, &correlation.Context{
			RequestID:     requestID,
			CorrelationID: correlationID,
		})

		// Echo back in response metadata
		grpc.SetHeader(ctx, metadata.Pairs(
			"x-request-id", requestID,
		))

		return handler(ctx, req)
	}
}

func getFirst(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
```

### Client Interceptor

Inject correlation IDs into outgoing gRPC calls.

```go
// UnaryClientCorrelation injects correlation IDs into gRPC metadata.
func UnaryClientCorrelation() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		c := correlation.FromContext(ctx)

		md := metadata.New(nil)
		if c.RequestID != "" {
			md.Set("x-request-id", c.RequestID)
		}
		if c.CorrelationID != "" {
			md.Set("x-correlation-id", c.CorrelationID)
		}

		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
```

---

## Common Mistakes

| Mistake | Why It's Bad | Fix |
|---------|-------------|-----|
| String keys in context | Collision with other packages | Use unexported struct type: `type ctxKey struct{}` |
| Missing header forwarding | Breaks correlation chain | Always inject IDs into outgoing requests |
| Inconsistent field names | Hard to search logs | Standardize: `request_id`, `correlation_id` |
| Not logging on errors | Impossible to trace failures | Always log with correlation IDs on error paths |
| Generating ID per span | Multiple IDs for same request | Generate once, propagate everywhere |
| No correlation in async tasks | Background work not traceable | Copy correlation context before spawning goroutine |
| Using trace ID as request ID | Conflates concepts | Use separate IDs: request (per-hop), correlation (end-to-end) |

---

## See Also

- [Tracing](./tracing.md) - Distributed tracing with OpenTelemetry
- [Metrics](./metrics.md) - Prometheus metrics with correlation labels
- [Logging](./logging.md) - Structured logging with slog
- [HTTP Server](../05-http-grpc/http-server.md) - HTTP middleware patterns
- [gRPC](../05-http-grpc/grpc.md) - gRPC interceptors and metadata
- [Error Handling](../02-language/error-handling.md) - Error context propagation
