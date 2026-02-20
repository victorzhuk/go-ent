# Conventions

Naming, file organization, code style, error handling, and testing conventions for go-ent.

## Naming Conventions

### Variables

Short but meaningful names following common patterns:

```go
ctx   // context
cfg   // configuration
req   // request
resp  // response
err   // error
tx    // transaction
repo  // repository
srv   // service
log   // logger
```

### Constructors

- `New()` - Public constructor
- `new*()` - Internal constructor

```go
func NewRepository(pool *pgxpool.Pool) Repository {
    return &repository{pool: pool}
}

func newPool(cfg Config) (*pgxpool.Pool, error) {
    // internal implementation
}
```

### Structs

- Private by default
- Public only for domain entities

```go
type repository struct {  // private
    pool *pgxpool.Pool
}

type User struct {  // public domain entity
    ID   uuid.UUID
    Name string
}
```

### Receivers

Single-letter receivers based on type:

```go
s *service
r *repository
u *User
```

### File Names

Lowercase with underscores:

```
user_repository.go
user_service.go
errors.go
models.go
```

## Code Style

### Imports

Group with blank lines: stdlib → third-party → internal:

```go
import (
    "context"
    "fmt"
    "os"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"

    "github.com/victorzhuk/go-ent/internal/domain"
)
```

Run `make fmt` to format imports automatically.

### Happy Path Left

Handle errors immediately, keep happy path unindented:

```go
// GOOD
item, ok := cache[key]
if !ok {
    return ErrNotFound
}
return item

// BAD
if ok := cache[key]; ok {
    return item
}
return ErrNotFound
```

### Struct Initialization

Use field names for clarity:

```go
// GOOD
user := &User{
    ID:   uuid.New(),
    Name: "John",
}
```

### Comments

Only explain WHY, never WHAT. If you're writing a comment explaining what code does, rename instead:

```go
// BAD — rename the function instead
// Check if user is valid
if user.Valid() {

// GOOD name
if user.IsValid() {

// GOOD comment — explains non-obvious reasoning
// Using buffered channel to prevent blocking on slow consumers
ch := make(chan Event, 100)
```

## File Organization

### Package Order

1. Package documentation
2. Imports
3. Constants
4. Errors
5. Types
6. Variables
7. Public functions
8. Private functions

## Error Handling

### Package-Level Errors

Define sentinel errors at the package level in an `errors.go` file:

```go
// spec/parser/errors.go
package parser

import "errors"

var (
    ErrInvalidFormat = errors.New("invalid format")
    ErrDuplicateID   = errors.New("duplicate task id")
    ErrEmptyContent  = errors.New("empty content")
)
```

### Error Wrapping

Always wrap errors with context using `fmt.Errorf` and the `%w` verb:

```go
// Good — provides context
return fmt.Errorf("query user %s: %w", id, err)
return fmt.Errorf("parse task %s: %w", taskID, err)

// Bad — no context
return fmt.Errorf("failed: %w", err)
return err
```

### Error Message Format

- Lowercase messages
- No trailing punctuation
- Include operation context and identifiers
- Use `%w` for wrapping to enable `errors.Is()` checking

```go
// Good
fmt.Errorf("create order: %w", err)
fmt.Errorf("save skill %s: %w", skillID, err)

// Bad
fmt.Errorf("Failed to create order: %w", err)
fmt.Errorf("Error saving skill: %w", err)
```

### Error Type Checking

Use `errors.Is()` for sentinel errors and `errors.As()` for custom error types:

```go
if errors.Is(err, ErrNotFound) {
    // Handle not found
}

var notFoundErr *NotFoundError
if errors.As(err, &notFoundErr) {
    log.Printf("resource %s not found", notFoundErr.ID)
}
```

### Error Placement by Layer

- Repository: Wrap database errors with query context
- UseCase: Wrap errors with business context
- Domain: Define domain-specific error types
- Transport: Map errors to HTTP status codes

### Repository Example

```go
func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
    var u userModel
    if err := r.pool.QueryRow(ctx, query, id).Scan(&u.ID, &u.Name); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, fmt.Errorf("user %s: %w", id, ErrNotFound)
        }
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    return toDomain(&u), nil
}
```

## Testing

### Table-Driven Tests

Use table-driven tests with `t.Run()` and `t.Parallel()`:

```go
func TestParseTask(t *testing.T) {
    tests := []struct {
        name    string
        content []byte
        want    *domain.Task
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid task",
            content: []byte("# Task\nMetadata: value"),
            want:    &domain.Task{ID: "task-id"},
        },
        {
            name:    "empty content",
            content: []byte(""),
            wantErr: true,
            errMsg:  "empty content",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            got, err := NewTaskParser().Parse(tt.content)

            if tt.wantErr {
                require.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.want.ID, got.ID)
        })
    }
}
```

### Test Organization

- Unit tests: `*_test.go` in same package
- Integration tests: `*_integration_test.go`
- Use `testify/assert` and `testify/require`
- Use `t.Parallel()` for concurrent tests

### Behavior-Focused Testing

Test observable behavior through interfaces, not internal implementation details.

### Temporary Directories

Use `t.TempDir()` for file-based tests:

```go
func TestLoad(t *testing.T) {
    tmpDir := t.TempDir()
    skillPath := filepath.Join(tmpDir, "skill1", "SKILL.md")
    require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0o750))
    require.NoError(t, os.WriteFile(skillPath, []byte(skillContent), 0o600))

    r := NewRegistry()
    err := r.Load(tmpDir)
    require.NoError(t, err)
}
```

### Best Practices

- Use `t.Parallel()` for independent test cases
- Use `require` for setup/failure conditions (fail fast)
- Use `assert` for assertions (continue on failure)
- Use `t.TempDir()` instead of manual temp directory creation
- Keep test cases focused and independent
- Use descriptive test case names

## Production Checklist

When writing production code:

- [ ] Request/Correlation ID propagation
- [ ] Health checks: `/healthz`, `/readyz`
- [ ] Metrics: `/metrics` (Prometheus)
- [ ] Structured logging (JSON prod, slog API)
- [ ] Graceful shutdown (30s, fresh context)
- [ ] Timeouts on all external calls
- [ ] Context propagation throughout
- [ ] No panic (recover in handlers)

## Essential Libraries

- **DB:** pgx/v5, squirrel, goose/v3, clickhouse-go/v2
- **MQ:** amqp091-go, kafka-go, redis/v9
- **HTTP:** fasthttp or net/http, ogen (OpenAPI)
- **Config:** env/v11
- **Logging:** log/slog + zerolog
- **Testing:** testify, testcontainers-go
- **Utils:** uuid, decimal, validator/v10
- **Production:** prometheus/client_golang, x/time/rate, x/sync/errgroup
