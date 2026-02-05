# Tracing Patterns Quick Reference

Extracted from `docs/go/topics/07-observability/tracing.md` (644 lines) → 150 lines.

## OpenTelemetry Setup

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("my-service")

func processOrder(ctx context.Context, order Order) error {
    ctx, span := tracer.Start(ctx, "processOrder")
    defer span.End()

    span.SetAttributes(
        attribute.String("order.id", order.ID),
        attribute.Int("order.items", len(order.Items)),
    )

    if err := saveOrder(ctx, order); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    return nil
}
```

## Trace Propagation

```go
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

// HTTP client
client := &http.Client{
    Transport: otelhttp.NewTransport(http.DefaultTransport),
}

// HTTP server
handler := otelhttp.NewHandler(myHandler, "my-service")
http.ListenAndServe(":8080", handler)
```
