---
name: go-code
description: "Go 1.25+ implementation patterns, error handling, concurrency. Auto-activates for: writing Go code, implementing features, refactoring, error handling, configuration."
---

# Go Code Patterns (1.25+)

## Versions (Jan 2026)

- **Go 1.25.6** — Current stable
- **Go 1.26** — Coming Feb 2026

## Bootstrap

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
    
    ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
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

## Error Handling

```go
// Wrap with context, lowercase
if err != nil {
    return fmt.Errorf("query user %s: %w", id, err)
}

// Domain errors
var (
    ErrNotFound = errors.New("not found")
    ErrConflict = errors.New("conflict")
)

// Check wrapped
if errors.Is(err, ErrNotFound) { ... }
```

## Concurrency

```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(10)

for _, id := range ids {
    g.Go(func() error {
        return s.process(ctx, id)
    })
}
return g.Wait()
```

## Repository

```go
type repository struct {
    pool *pgxpool.Pool
    psql sq.StatementBuilderType
}

func New(pool *pgxpool.Pool) *repository {
    return &repository{
        pool: pool,
        psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
    }
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    query, args, _ := r.psql.
        Select("id", "email", "created_at").
        From("users").
        Where(sq.Eq{"id": id}).
        ToSql()
    
    var m userModel
    if err := r.pool.QueryRow(ctx, query, args...).Scan(&m.ID, &m.Email, &m.CreatedAt); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, contract.ErrNotFound
        }
        return nil, fmt.Errorf("query: %w", err)
    }
    return toEntity(&m), nil
}
```

## Config (env/v11)

```go
type Config struct {
    App AppConfig `envPrefix:"APP_"`
    DB  DBConfig  `envPrefix:"DB_"`
}

type DBConfig struct {
    DSN         string `env:"DSN,required"`
    MaxConns    int    `env:"MAX_CONNS" envDefault:"25"`
    MaxIdleTime time.Duration `env:"MAX_IDLE_TIME" envDefault:"5m"`
}

func LoadFromEnv(getenv func(string) string) (*Config, error) {
    var cfg Config
    if err := env.ParseWithOptions(&cfg, env.Options{Environment: getenv}); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

## Naming

```go
cfg, repo, srv, pool, ctx, req, resp, err, tx, log
// NOT: applicationConfiguration, userRepositoryInstance
```
