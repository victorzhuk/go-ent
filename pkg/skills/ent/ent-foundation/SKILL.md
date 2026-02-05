---
name: ent-foundation
description: Core decision framework - judgment, principals, and Go conventions. Preloaded by all agents for consistent decision-making and code style.
triggers:
  - ent-foundation
  - judgment
  - principals
  - conventions
---

## Role

Foundation for thoughtful, principled decision-making and consistent Go code style across all agents.

## Integration with Shared Judgment Framework

This skill complements the judgment framework loaded by all agents via the shared `_foundation.md` prompt. The shared prompt provides the core decision-making framework (The Standard, Principal Hierarchy, Decision Framework, Non-Negotiable Boundaries, When to Ask vs. Decide).

These Go conventions below are the **authoritative reference** for code style, architecture, and implementation patterns. When applying the judgment framework to Go code:

1. **Apply principal hierarchy** from shared prompt to resolve conflicts
2. **Use Go conventions** as the "Best practices" and "Project conventions" principles
3. **Exercise senior developer judgment** within both frameworks
4. **Document decisions** that deviate from conventions
5. **Accept responsibility** for outcomes

See `references/` directory for extended guidance on naming, style, and idioms extracted from comprehensive documentation.

## Go Code Conventions

Go coding standards and Clean Architecture conventions for consistent, maintainable code across the project.

### Naming Conventions

**Variables (short, natural):**
```go
cfg, repo, srv, ctx, req, resp, err, tx, log
pool, client, handler, query, args, rows
```

**Constructors:**
- Public API: `New()`
- Internal only: `new*()`

**Structs:**
- Private by default: `type app struct`
- Public only for domain: `type User struct`

**Receivers:** Short - `s *Service`, `u *User`, `r *Repo`

### Error Handling

**Format (lowercase, wrap with %w):**
```go
// Good
return fmt.Errorf("query user %s: %w", id, err)
return fmt.Errorf("create order: %w", err)

// Bad
return fmt.Errorf("Failed to query user: %w", err)  // uppercase
return err  // no context
```

**Error Types:**
- Domain errors: Custom types (`ErrUserNotFound`)
- Repo errors: Wrap with operation (`query user: %w`)
- UseCase errors: Business context (`create order: %w`)
- Transport: Map to HTTP status codes

### Comments Policy

**ZERO comments explaining WHAT** - rename instead

```go
// ❌ BAD
// Create a new user
user := NewUser(name)

// ✅ GOOD (no comment, clear name)
user := NewUser(name)

// Only WHY comments if non-obvious
// Required by legacy API - remove after v2 migration
resp.Header.Set("X-Legacy-Token", token)
```

### Clean Architecture Layers

**Domain (`internal/domain/`):**
- ZERO external dependencies
- Pure business logic
- NO struct tags
- Public entities/interfaces

**Repository (`internal/repository/`):**
- Structure: `{concept}/{impl}/`
- Files: `repo.go`, `models.go`, `mappers.go`, `schema.go`
- Private models with tags
- Return domain entities

**UseCase (`internal/usecase/`):**
- Private structs/DTOs
- Public interface
- Transaction boundaries

**Transport (`internal/transport/`):**
- Private DTOs with validation tags
- ZERO business logic

### Architecture Rules

1. Domain has ZERO external deps
2. Interfaces at consumer side
3. Dependencies flow inward: Transport → UseCase → Domain ← Repository
4. Accept interfaces, return structs
5. Private by default
6. Context first parameter

### File Organization

```go
import (
    "context"    // stdlib
    "fmt"

    "github.com/org/project/internal/domain"  // internal

    "github.com/google/uuid"  // third-party
)

// Constants/errors
const ConstName = "value"

// Public types first
type PublicType struct{}
func New() *PublicType {}

// Private types
func privateFunc() {}
```

## Examples

### Good Style
```go
type repository struct {
    pool *pgxpool.Pool
    psql sq.SelectBuilder
}

func New(pool *pgxpool.Pool) Repository {
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
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    return toEntity(&m), nil
}
```

### Bad Style (Anti-Patterns)
```go
// ❌ Verbose AI-style naming
applicationConfiguration := config.Load()
userRepositoryInstance := userRepo.New(pool)

// ❌ WHAT comments
// Create a new user
user := NewUser(name)

// ❌ Unwrapped errors
if err != nil {
    return err
}

// ❌ Uppercase error messages
return fmt.Errorf("Failed to query user: %w", err)
```
