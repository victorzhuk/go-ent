---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
version: "${VERSION}"
author: "${AUTHOR}"
tags: [${TAGS}]
triggers:
  - pattern: "go code|golang implementation|write go|go patterns"
    weight: 0.9
  - keywords: ["go", "golang", "go-code", "go patterns"]
    weight: 0.8
  - filePattern: "*.go"
    weight: 0.7
---

# ${SKILL_NAME}

<role>
Expert Go developer focused on clean architecture, patterns, and idioms. 
Prioritize SOLID, DRY, KISS, YAGNI principles with production-grade quality, maintainability, and performance.
</role>

<instructions>

## Bootstrap Pattern

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

**Why this pattern**:
- Testable (injectable dependencies)
- Graceful shutdown (30s timeout)
- No globals except in `main()`
- Proper signal handling

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

**Rules**:
- Always wrap with context
- Lowercase, no trailing punctuation
- Use `%w` for wrapping
- Domain errors as package-level vars

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

**Patterns**:
- `errgroup` for parallel work with error handling
- `sync.WaitGroup` for fire-and-forget
- Channels for communication
- `context` for cancellation
- `sync.Once` for lazy initialization

## Repository Pattern

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

**Key points**:
- Use squirrel for complex queries
- Map `pgx.ErrNoRows` to domain error
- Wrap errors with operation context
- Private models, public entities

## Configuration

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

**Pattern**: Injectable `getenv` for testing

## Naming Conventions

```go
// Good: Short, clear
cfg, repo, srv, pool, ctx, req, resp, err, tx, log

// Bad: Verbose
applicationConfiguration, userRepositoryInstance
```

**Guidelines**:
- Short variable names in small scopes
- Descriptive names for public APIs
- Receivers: single letter or short abbreviation (`s`, `r`, `uc`)
- Avoid single letters except: `i` (index), `t` (test), `w` (writer), `r` (reader)

## Common Patterns

### Functional Options
```go
type Option func(*Server)

func WithPort(port int) Option {
    return func(s *Server) { s.port = port }
}

func New(opts ...Option) *Server {
    s := &Server{port: 8080}
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

### Table-Driven Tests
```go
tests := []struct {
    name string
    input string
    want  int
}{
    {"valid", "123", 123},
    {"empty", "", 0},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        got := parse(tt.input)
        assert.Equal(t, tt.want, got)
    })
}
```

### Context Propagation
```go
// Always first parameter
func Process(ctx context.Context, data Data) error {
    // Check cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    // Pass to downstream calls
    return s.repo.Save(ctx, data)
}
```

</instructions>

<constraints>
- Include clean, idiomatic Go code following standard conventions
- Include proper error wrapping with context using `%w` verb
- Include context propagation as first parameter throughout layers
- Include domain entities with zero external dependencies
- Include dependency injection pattern (accept interfaces, return structs)
- Exclude magic numbers (use named constants instead)
- Exclude global mutable state (pass dependencies explicitly)
- Exclude panic in production code (use error handling instead)
- Exclude over-engineering and premature abstractions (YAGNI)
- Exclude AI-style verbose naming and unnecessary comments
- Bound to clean layered architecture: Transport → UseCase → Domain ← Repository
- Follow DI pattern with explicit dependency graphs
- Keep domain layer pure with no external dependencies
</constraints>

<edge_cases>
If input is unclear or ambiguous: Ask clarifying questions to understand the specific requirement before proceeding with implementation.

If context is missing for a feature: Request additional information about architecture decisions, existing patterns, or integration points.

If performance concerns arise: Delegate to go-perf skill for profiling, optimization strategies, and benchmarking guidance.

If architecture questions emerge: Delegate to go-arch skill for system design, layer boundaries, and structural decisions.

If testing requirements are needed: Delegate to go-test skill for test coverage, table-driven tests, and mocking strategies.

If security considerations are relevant: Delegate to go-sec skill for authentication, authorization, and input validation patterns.

If database interactions are complex: Delegate to go-db skill for query optimization, transaction patterns, and connection pooling.

If API design is involved: Delegate to go-api skill for OpenAPI specs, gRPC services, and REST endpoint patterns.

</edge_cases>

<examples>
<example>
<input>Refactor main() to use bootstrap pattern with graceful shutdown</input>
<output>
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
</output>
</example>

<example>
<input>Fix error handling in this function - it's not wrapping errors properly</input>
<output>
```go
// Before
func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    var m userModel
    if err := r.pool.QueryRow(ctx, query, args...).Scan(&m.ID, &m.Email, &m.CreatedAt); err != nil {
        return nil, err
    }
    return toEntity(&m), nil
}

// After
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
</output>
</example>

<example>
<input>Implement repository pattern with proper error handling and domain mapping</input>
<output>
```go
package userrepo

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
)

type repository struct {
    db *sqlx.DB
}

func New(db *sqlx.DB) *repository {
    return &repository{db: db}
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
    const query = `SELECT id, email, created_at FROM users WHERE id = $1`

    var m userModel
    if err := r.db.GetContext(ctx, &m, query, id.String()); err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    return toEntity(&m), nil
}

func (r *repository) Save(ctx context.Context, user *User) error {
    const query = `INSERT INTO users (id, email, created_at) VALUES ($1, $2, $3)`

    m := toModel(user)
    _, err := r.db.ExecContext(ctx, query, m.ID, m.Email, m.CreatedAt)
    if err != nil {
        return fmt.Errorf("save user %s: %w", user.ID, err)
    }
    return nil
}
```
</output>
</example>

<example>
<input>Create a use case layer with proper orchestration and transaction handling</input>
<output>
```go
package userusecase

import (
    "context"
    "fmt"
)

type UseCase struct {
    repo   Repository
    logger *slog.Logger
}

func New(repo Repository, logger *slog.Logger) *UseCase {
    return &UseCase{
        repo:   repo,
        logger: logger,
    }
}

func (uc *UseCase) CreateUser(ctx context.Context, email string) (*User, error) {
    if email == "" {
        return nil, fmt.Errorf("email required")
    }

    existing, err := uc.repo.FindByEmail(ctx, email)
    if err == nil && existing != nil {
        return nil, fmt.Errorf("user already exists: %s", email)
    }

    user := &User{
        ID:    uuid.New(),
        Email: email,
    }

    if err := uc.repo.Save(ctx, user); err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }

    uc.logger.Info("user created", "id", user.ID, "email", user.Email)
    return user, nil
}
```
</output>
</example>
</examples>

<output_format>
Provide production-ready Go code following established patterns:

1. **Code Structure**: Clean, idiomatic Go with proper package organization
2. **Naming**: Short, natural variable names (cfg, repo, ctx, req, resp)
3. **Error Handling**: Wrapped errors with lowercase context using `%w`
4. **Context**: Always first parameter, propagated through all layers
5. **Interfaces**: Minimal interfaces at consumer side, return structs
6. **Examples**: Complete, runnable code blocks with language tags
7. **Explanations**: Clear, concise justifications for pattern choices

Focus on practical implementation with minimal abstractions unless complexity demands it.
</output_format>
