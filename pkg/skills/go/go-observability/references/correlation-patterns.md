# Correlation Patterns Quick Reference

Extracted from `docs/go/topics/07-observability/correlation.md` (667 lines) → 100 lines.

## Request ID Middleware

```go
import "github.com/google/uuid"

type contextKey string
const requestIDKey = contextKey("request_id")

func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        ctx := context.WithValue(r.Context(), requestIDKey, requestID)
        w.Header().Set("X-Request-ID", requestID)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(requestIDKey).(string); ok {
        return id
    }
    return ""
}
```

## Propagate Across Services

```go
func callService(ctx context.Context, url string) error {
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

    // Propagate request ID
    if requestID := GetRequestID(ctx); requestID != "" {
        req.Header.Set("X-Request-ID", requestID)
    }

    // Propagate trace context (automatic with otelhttp)
    resp, err := http.DefaultClient.Do(req)
    // ...
    return err
}
```
