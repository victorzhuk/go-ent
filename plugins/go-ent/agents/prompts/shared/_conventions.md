# Code Conventions

Go code style and architecture patterns for the project.

## Naming Conventions

### Variables (short, natural)
```go
cfg, repo, srv, ctx, req, resp, err, tx, log
pool, client, handler, query, args, rows
```

### Constructors
- **Public API**: `New()`
- **Internal only**: `new*()`

### Structs
- Private by default: `type app struct`
- Public only for domain: `type User struct`

### Receivers
```go
s *Service      // service
u *User         // user
r *Repo         // repository
```

## Error Handling

### Format
```go
// Good
return fmt.Errorf("query user %s: %w", id, err)
return fmt.Errorf("create order: %w", err)

// Bad
return fmt.Errorf("Failed to query user: %w", err)  // uppercase
return err                                          // no context
return fmt.Errorf("error: %w", err)                  // useless context
```

### Error Types
- **Domain errors**: Custom types (`ErrUserNotFound`)
- **Repo errors**: Wrap with operation (`query user: %w`)
- **UseCase errors**: Business context (`create order: %w`)
- **Transport**: Map to HTTP status codes

## Comments Policy

### ZERO comments explaining WHAT
```go
// ❌ BAD
// Create a new user
user := NewUser(name)

// Initialize the server
srv := NewServer(cfg)

// Get user by ID from database
func GetUserByID(id string) (*User, error)
```

### Only WHY comments if non-obvious
```go
// ✅ GOOD
// Required by legacy API - remove after v2 migration
resp.Header.Set("X-Legacy-Token", token)

// Counterintuitive: zero means unlimited per vendor docs
if limit == 0 {
    unlimited = true
}
```

## Anti-Patterns

### NO AI-style verbose names
```go
// ❌ BAD
applicationConfiguration := config.Load()
userRepositoryInstance := userRepo.New(pool)
httpServerInstance := NewServer(cfg)
databaseConnectionPool := pgxpool.New(ctx, dsn)

// ✅ GOOD
cfg := config.Load()
repo := userRepo.New(pool)
srv := NewServer(cfg)
pool := pgxpool.New(ctx, dsn)
```

### NO unnecessary verbosity
```go
// ❌ BAD
if err != nil {
    return fmt.Errorf("failed to create user in database: %w", err)
}

// ✅ GOOD
if err != nil {
    return fmt.Errorf("create user: %w", err)
}
```

## Clean Architecture Layers

### Domain (`internal/domain/`)
- **ZERO** external dependencies
- Pure business logic
- NO struct tags
- Public entities/interfaces

```go
type User struct {
    ID        uuid.UUID
    Email     string
    CreatedAt time.Time
}

func NewUser(email string) (*User, error) {
    // validation and entity creation
}
```

### Repository (`internal/repository/`)
- Structure: `{concept}/{impl}/`
- Private models with tags
- Private mappers
- Return domain entities

**Files:**
- `repo.go` - Interface and implementation
- `models.go` - Database models with tags
- `mappers.go` - Entity ↔ Model conversions
- `schema.go` - Table schema definitions

```go
type repository struct {
    pool *pgxpool.Pool
    psql sq.SelectBuilder
}

func NewRepository(pool *pgxpool.Pool) Repository {
    return &repository{pool: pool, psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar)}
}
```

### UseCase (`internal/usecase/`)
- Private structs/DTOs
- Public interface
- Orchestrate domain logic
- Transaction boundaries

```go
type createUserUC struct {
    repo Repository
}

func (uc *createUserUC) Execute(ctx context.Context, req CreateUserReq) (*CreateUserResp, error) {
    // business orchestration
}
```

### Transport (`internal/transport/`)
- Private DTOs with validation tags
- Request/response mapping only
- ZERO business logic

```go
type createUserRequest struct {
    Email string `json:"email" validate:"required,email"`
}
```

## Architecture Rules

1. **Domain has ZERO external deps**
2. **Interfaces at consumer side**
3. **Dependencies flow inward**: Transport → UseCase → Domain ← Repository ← Infrastructure
4. **Accept interfaces, return structs**
5. **Private by default**
6. **Context first parameter**: `func (r *repo) Find(ctx context.Context, id uuid.UUID)`

## File Organization

```go
package pkg

import (
    // stdlib
    "context"
    "fmt"

    // internal
    "github.com/org/project/internal/domain"

    // third-party
    "github.com/google/uuid"
)

const (
    ConstName = "value"
)

var (
    GlobalVar = "value"
)

// Public types first
type PublicType struct{}

func New() *PublicType {}

// Private types
func privateFunc() {}
```

## Testing Patterns

### Table-driven tests
```go
tests := []struct {
    name string
    // fields
}{
    {"valid input", ...},
    {"empty input", ...},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // test
    })
}
```

### Use `testify/assert`
```go
assert.NoError(t, err)
assert.Equal(t, expected, actual)
assert.NotNil(t, result)
```

### Real implementations over mocks
- Use real `net.Listen` instead of mock servers
- Use `testcontainers` for database tests
- Mock only complex external services

## Code Style Checklist

- [ ] Short, natural variable names (`cfg`, not `config`)
- [ ] Errors wrapped with lowercase context
- [ ] No WHAT comments (rename if needed)
- [ ] Domain has zero external deps
- [ ] Interfaces at consumer side
- [ ] Happy path left (check errors first)
- [ ] No magic numbers (use named constants)
- [ ] Context passed through all layers
- [ ] No naked returns
