---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
triggers:
  - code review
  - debug
  - security
  - architecture
---

# ${SKILL_NAME}

## Role

Core development specialist covering security, code review, debugging, and architecture patterns. Prioritize best practices, maintainability, and evidence-based problem solving across all development concerns.

## Instructions

### Code Review Checklist

```markdown
- [ ] Error handling: errors wrapped with context, not swallowed
- [ ] Context propagation: ctx first parameter, passed through all layers
- [ ] No magic numbers: named constants used
- [ ] No global state: dependencies injected
- [ ] Tests present: table-driven with parallel execution
- [ ] No over-engineering: no premature abstractions
```

### Debugging Approach

```go
// Add structured logging at decision points
log.Info("processing request",
    "user_id", userID,
    "action", "payment",
    "amount", amount,
)

// Use context for tracing
ctx = context.WithValue(ctx, correlationIDKey, correlationID)

// Check for common issues
if err != nil {
    log.Error("operation failed",
        "error", err,
        "component", "payment",
        "user_id", userID,
    )
    return fmt.Errorf("process payment: %w", err)
}
```

### Security Patterns

```go
// Validate all inputs
func ValidateUserInput(input string) error {
    if len(input) > 255 {
        return fmt.Errorf("input too long: %d chars (max 255)", len(input))
    }
    if !allowedPattern.MatchString(input) {
        return fmt.Errorf("invalid input format")
    }
    return nil
}

// Use parameterized queries (never string concatenation)
query, args, _ := sq.Select("id", "email").
    From("users").
    Where(sq.Eq{"id": id}).
    ToSql()
```

### Edge Cases

If the bug is intermittent: Add structured logging with correlation IDs to trace through distributed calls.

If security vulnerability is found: Assess severity, patch immediately, review for similar patterns in codebase.

If architecture decision is needed: Document as ADR with context, decision, and consequences.

## Examples

### Example 1: Code review feedback for error handling

**Input**: Review this function's error handling

**Output**:
```go
// Before: errors not wrapped with context
func (s *service) CreateUser(ctx context.Context, email string) error {
    _, err := s.repo.FindByEmail(ctx, email)
    if err == nil {
        return errors.New("user exists")
    }
    err = s.repo.Save(ctx, &User{Email: email})
    if err != nil {
        return err
    }
    return nil
}

// After: proper error wrapping and domain errors
func (s *service) CreateUser(ctx context.Context, email string) error {
    existing, err := s.repo.FindByEmail(ctx, email)
    if err != nil && !errors.Is(err, ErrNotFound) {
        return fmt.Errorf("check existing user: %w", err)
    }
    if existing != nil {
        return fmt.Errorf("%w: %s", ErrConflict, email)
    }
    if err := s.repo.Save(ctx, &User{Email: email}); err != nil {
        return fmt.Errorf("save user: %w", err)
    }
    return nil
}
```

### Example 2: Debug a race condition

**Input**: Debug a race condition in concurrent map access

**Output**:
```go
// Problem: concurrent map read/write
type Cache struct {
    data map[string]string // race condition!
}

// Fix: use sync.RWMutex
type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    v, ok := c.data[key]
    return v, ok
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value
}

// Or use sync.Map for high-contention scenarios
var cache sync.Map
cache.Store(key, value)
v, ok := cache.Load(key)
```
