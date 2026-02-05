# Logging

Production logging using `log/slog` (Go 1.21+) with zerolog as the performance-optimized handler.

## Quick Reference

| Pattern                          | Use When                 |
|----------------------------------|--------------------------|
| `slog.Info("msg", "key", value)` | Structured logging       |
| `slog.With("key", value)`        | Add context to logger    |
| `slog.SetDefault(logger)`        | Set global logger        |
| `slog.NewJSONHandler(w, opts)`   | JSON output (production) |
| `zeroslog.NewHandler(w)`         | High-performance JSON    |

## Basic Setup

### Development (Text Output)

```go
import (
    "log/slog"
    "os"
)

func setupLogger() {
    handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
        AddSource: true, // Add source file:line
    })

    logger := slog.New(handler)
    slog.SetDefault(logger)
}

func main() {
    setupLogger()

    slog.Info("server starting", "port", 8080)
    slog.Debug("debug message", "data", map[string]string{"key": "value"})
}
```

### Production (JSON with zerolog)

```go
import (
    "log/slog"
    "os"

    "github.com/rs/zerolog"
    zeroslog "github.com/samber/slog-zerolog/v2"
)

func setupProductionLogger() *slog.Logger {
    zerologLogger := zerolog.New(os.Stdout).
        Level(zerolog.InfoLevel).
        With().
        Timestamp().
        Logger()

    handler := zeroslog.Option{
        Level:  slog.LevelInfo,
        Logger: &zerologLogger,
    }.NewZerologHandler()

    logger := slog.New(handler)
    slog.SetDefault(logger)

    return logger
}
```

## Structured Logging

### Basic Logging

```go
// Info level
slog.Info("user created",
    "user_id", user.ID,
    "email", user.Email,
)

// Error level with error
if err != nil {
    slog.Error("failed to create user",
        "error", err,
        "email", user.Email,
    )
}

// Debug level
slog.Debug("processing request",
    "method", r.Method,
    "path", r.URL.Path,
)

// Warn level
slog.Warn("rate limit approaching",
    "current", currentRate,
    "limit", maxRate,
)
```

### Log Levels

```go
const (
    LevelDebug = slog.LevelDebug // -4
    LevelInfo  = slog.LevelInfo  // 0
    LevelWarn  = slog.LevelWarn  // 4
    LevelError = slog.LevelError // 8
)

// Custom levels
const (
    LevelTrace = slog.Level(-8)
    LevelFatal = slog.Level(12)
)

// Set minimum level
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo, // Only Info and above
})
```

## Context Loggers

### Logger with Context

```go
func (s *Service) ProcessUser(ctx context.Context, userID string) error {
    // Create logger with context
    logger := slog.With(
        "user_id", userID,
        "service", "user_service",
    )

    logger.Info("starting user processing")

    user, err := s.repo.GetUser(ctx, userID)
    if err != nil {
        logger.Error("failed to get user", "error", err)
        return err
    }

    logger.Info("user processing complete",
        "name", user.Name,
        "email", user.Email,
    )

    return nil
}
```

### Logger from Context

```go
type contextKey string

const loggerKey contextKey = "logger"

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
    return context.WithValue(ctx, loggerKey, logger)
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
    if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
        return logger
    }
    return slog.Default()
}

// Usage
func handler(w http.ResponseWriter, r *http.Request) {
    logger := LoggerFromContext(r.Context())
    logger.Info("handling request", "path", r.URL.Path)
}
```

## HTTP Middleware

### Request Logging

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Wrap response writer to capture status
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

        // Add request ID to logger
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        logger := slog.With(
            "request_id", requestID,
            "method", r.Method,
            "path", r.URL.Path,
        )

        // Add logger to context
        ctx := ContextWithLogger(r.Context(), logger)

        logger.Info("request started")

        next.ServeHTTP(wrapped, r.WithContext(ctx))

        duration := time.Since(start)
        logger.Info("request completed",
            "status", wrapped.statusCode,
            "duration_ms", duration.Milliseconds(),
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
    w.statusCode = code
    w.ResponseWriter.WriteHeader(code)
}
```

## Grouping Attributes

### LogValue for Complex Types

```go
type User struct {
    ID    string
    Name  string
    Email string
}

func (u User) LogValue() slog.Value {
    return slog.GroupValue(
        slog.String("id", u.ID),
        slog.String("name", u.Name),
        slog.String("email", u.Email),
    )
}

// Usage
slog.Info("user created", "user", user)
// Output: {"msg":"user created","user":{"id":"123","name":"Alice","email":"alice@example.com"}}
```

### Manual Grouping

```go
slog.Info("request details",
    slog.Group("request",
        slog.String("method", r.Method),
        slog.String("path", r.URL.Path),
        slog.Int("content_length", int(r.ContentLength)),
    ),
    slog.Group("user",
        slog.String("id", userID),
        slog.String("ip", r.RemoteAddr),
    ),
)
```

## Performance Optimization

### Conditional Logging

```go
// Bad - always evaluates expensive operation
slog.Debug("expensive", "data", expensiveOperation())

// Good - only evaluates if debug enabled
if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
    slog.Debug("expensive", "data", expensiveOperation())
}

// Good - use LogAttrs for zero allocation
slog.LogAttrs(context.Background(), slog.LevelDebug, "message",
    slog.String("key", value),
)
```

### Zerolog for High Performance

```go
import "github.com/rs/zerolog/log"

// Zerolog native (if slog API not required)
log.Info().
    Str("user_id", userID).
    Int("count", count).
    Msg("processing users")

// With structured context
logger := log.With().
    Str("service", "api").
    Str("version", "1.0.0").
    Logger()

logger.Info().Msg("service started")
```

## Error Logging

### Logging Errors

```go
func (s *Service) ProcessOrder(ctx context.Context, orderID string) error {
    logger := LoggerFromContext(ctx)

    order, err := s.repo.GetOrder(ctx, orderID)
    if err != nil {
        logger.Error("failed to get order",
            "error", err,
            "order_id", orderID,
        )
        return fmt.Errorf("get order: %w", err)
    }

    if err := order.Validate(); err != nil {
        logger.Warn("invalid order",
            "error", err,
            "order_id", orderID,
        )
        return fmt.Errorf("validate order: %w", err)
    }

    logger.Info("order processed successfully",
        "order_id", orderID,
        "amount", order.Amount,
    )

    return nil
}
```

### Stack Traces (with pkg/errors)

```go
import (
    "github.com/pkg/errors"
)

func processFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        err = errors.Wrap(err, "read file")
        slog.Error("file processing failed",
            "error", err,
            "stack", fmt.Sprintf("%+v", err), // Stack trace
        )
        return err
    }
    return nil
}
```

## Environment-Based Configuration

### Development vs Production

```go
func NewLogger(env string) *slog.Logger {
    var handler slog.Handler

    switch env {
    case "development":
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level:     slog.LevelDebug,
            AddSource: true,
        })

    case "production":
        zerologLogger := zerolog.New(os.Stdout).
            Level(zerolog.InfoLevel).
            With().
            Timestamp().
            Caller().
            Logger()

        handler = zeroslog.Option{
            Level:  slog.LevelInfo,
            Logger: &zerologLogger,
        }.NewZerologHandler()

    default:
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelInfo,
        })
    }

    return slog.New(handler)
}
```

## Sampling (Reduce Log Volume)

### Sample High-Frequency Logs

```go
import "github.com/rs/zerolog"

func setupSampledLogger() {
    sampled := zerolog.New(os.Stdout).
        Sample(&zerolog.LevelSampler{
            TraceSampler: &zerolog.BurstSampler{
                Burst:  5,
                Period: 1 * time.Second,
            },
            DebugSampler: &zerolog.BurstSampler{
                Burst:  10,
                Period: 1 * time.Second,
            },
        })

    // Only allows 5 trace logs per second (burst)
    sampled.Trace().Msg("high frequency event")
}
```

## Testing with Logs

### Capture Logs in Tests

```go
import (
    "bytes"
    "testing"
)

func TestService(t *testing.T) {
    var buf bytes.Buffer
    handler := slog.NewJSONHandler(&buf, nil)
    logger := slog.New(handler)

    svc := NewService(logger)
    svc.Process()

    // Assert log output
    logs := buf.String()
    assert.Contains(t, logs, "processing started")
}
```

### Discard Logs in Tests

```go
import "io"

func TestQuiet(t *testing.T) {
    handler := slog.NewJSONHandler(io.Discard, nil)
    logger := slog.New(handler)

    svc := NewService(logger)
    svc.Process() // No log output
}
```

## Common Patterns

### Panic Recovery with Logging

```go
func recoverMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                logger := LoggerFromContext(r.Context())
                logger.Error("panic recovered",
                    "error", err,
                    "stack", string(debug.Stack()),
                )
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### Correlation ID

```go
func correlationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        correlationID := r.Header.Get("X-Correlation-ID")
        if correlationID == "" {
            correlationID = uuid.New().String()
        }

        logger := slog.With("correlation_id", correlationID)
        ctx := ContextWithLogger(r.Context(), logger)

        w.Header().Set("X-Correlation-ID", correlationID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Common Mistakes

| Mistake                          | Fix                                                             |
|----------------------------------|-----------------------------------------------------------------|
| String formatting in log args    | Use structured logging: `"user_id", id` not `"user_id: %s", id` |
| Logging sensitive data           | Sanitize PII, passwords, tokens                                 |
| Not setting global logger        | Call `slog.SetDefault(logger)`                                  |
| Expensive operations in log args | Check level before expensive calls                              |
| Missing context                  | Use `slog.With()` to add context                                |
| Inconsistent key names           | Standardize: `user_id` not `userId`, `userID`                   |

## Best Practices

```go
// ✓ Good - structured with consistent keys
slog.Info("user created",
    "user_id", user.ID,
    "email", user.Email,
)

// ✗ Bad - unstructured message
slog.Info(fmt.Sprintf("User %s created with email %s", user.ID, user.Email))

// ✓ Good - appropriate level
slog.Error("database connection failed", "error", err)

// ✗ Bad - wrong level
slog.Info("database connection failed", "error", err)

// ✓ Good - sanitize sensitive data
slog.Info("user logged in",
    "user_id", user.ID,
    "email_domain", emailDomain(user.Email),
)

// ✗ Bad - logging passwords
slog.Info("login attempt",
    "password", password, // Never log passwords!
)
```

## See Also

- [Tracing](./tracing.md) - Distributed tracing with OpenTelemetry
- [Metrics](./metrics.md) - Prometheus metrics
- [Correlation](./correlation.md) - Request ID propagation
- [slog package](https://pkg.go.dev/log/slog)
- [zerolog](https://github.com/rs/zerolog)
