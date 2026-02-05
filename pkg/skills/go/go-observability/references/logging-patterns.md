# Logging Patterns Quick Reference

Extracted from `docs/go/topics/07-observability/logging.md` (668 lines) → 120 lines.

## slog Basics

```go
import "log/slog"

// Setup
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

// Logging
slog.Info("user created", "user_id", id, "email", email)
slog.Error("failed to save", "error", err, "user_id", id)
```

## Context Logging

```go
func (s *service) Handle(ctx context.Context, req Request) error {
    logger := slog.With("request_id", getRequestID(ctx), "user_id", req.UserID)
    logger.Info("handling request")

    if err := s.process(ctx, req); err != nil {
        logger.Error("processing failed", "error", err)
        return err
    }

    logger.Info("request handled successfully")
    return nil
}
```

## Log Levels

```go
slog.Debug("detailed debug info")    // Development only
slog.Info("normal operation")         // Production events
slog.Warn("warning condition")        // Potential issues
slog.Error("error occurred")          // Errors requiring attention
```
