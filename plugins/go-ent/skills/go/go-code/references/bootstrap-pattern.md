# Bootstrap Pattern for Go Applications

## Main Function with Testable run()

```go
func main() {
    if err := run(context.Background(), os.Getenv, os.Stdout, os.Stderr); err != nil {
        slog.Error("fatal", "error", err)
        os.Exit(1)
    }
}

func run(ctx context.Context, getenv func(string) string, stdout, stderr io.Writer) error {
    cfg, err := config.LoadFromEnv(getenv)
    if err != nil {
        return fmt.Errorf("config: %w", err)
    }

    log := slog.New(slog.NewJSONHandler(stdout, nil))
    slog.SetDefault(log)

    app, err := app.New(log, cfg)
    if err != nil {
        return fmt.Errorf("app: %w", err)
    }

    ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
    defer cancel()

    errCh := make(chan error, 1)
    go func() { errCh <- app.Start(ctx) }()

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        log.Info("shutdown signal")
    }

    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()
    return app.Shutdown(shutdownCtx)
}
```

**Why this pattern:**
- **Testable**: Injectable dependencies (getenv, stdout, stderr)
- **Graceful shutdown**: 30s timeout on fresh context
- **No globals**: Except in main()
- **Proper signal handling**: SIGTERM, SIGINT, SIGQUIT
- **Clean error propagation**: All errors bubble up to main
