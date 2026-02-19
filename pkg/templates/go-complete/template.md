---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
triggers:
  - go code
  - golang
  - go patterns
  - go implementation
---

# ${SKILL_NAME}

## Role

Expert Go developer focused on clean architecture, patterns, and idioms. Prioritize SOLID, DRY, KISS, YAGNI principles with production-grade quality, maintainability, and performance.

## Instructions

### Bootstrap Pattern

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

### Error Handling

```go
if err != nil {
    return fmt.Errorf("query user %s: %w", id, err)
}

var (
    ErrNotFound = errors.New("not found")
    ErrConflict = errors.New("conflict")
)

if errors.Is(err, ErrNotFound) { ... }
```

### Concurrency

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

### Repository Pattern

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

### Configuration

```go
type Config struct {
    App AppConfig `envPrefix:"APP_"`
    DB  DBConfig  `envPrefix:"DB_"`
}

type DBConfig struct {
    DSN         string        `env:"DSN,required"`
    MaxConns    int           `env:"MAX_CONNS" envDefault:"25"`
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

### Edge Cases

If input is unclear or ambiguous: Ask clarifying questions before proceeding with implementation.

If performance concerns arise: Delegate to go-perf skill for profiling and benchmarking guidance.

If architecture questions emerge: Delegate to go-arch skill for system design and layer boundaries.

If database interactions are complex: Delegate to go-db skill for query optimization and transaction patterns.

## Examples

### Example 1: Bootstrap pattern with graceful shutdown

**Input**: Refactor main() to use bootstrap pattern with graceful shutdown

**Output**:
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

### Example 2: Repository with proper error handling

**Input**: Fix error handling in repository - errors not wrapped properly

**Output**:
```go
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
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    return toEntity(&m), nil
}
```
