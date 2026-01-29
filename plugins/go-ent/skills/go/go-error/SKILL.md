---
name: go-error
description: 'Implement Go error handling patterns (wrapping, custom errors, error types). Auto-activates for: error handling, error wrapping, custom errors, sentinel errors, error chains.'
version: 2.0.0
author: go-ent
license: MIT
compatibility:
    claude_code: '>=1.0'
    opencode: '>=0.1'
tags:
    - go
    - error
    - error-handling
quality_score: 86
category: go
triggers:
    keywords:
        - error handling
        - error wrapping
        - custom error
        - error type
        - sentinel error
        - error chain
    file_pattern: errors.go|**/errors/*.go|**/*error*.go
    weight: 0.8
---

## Role

Expert Go error handling engineer specializing in error design patterns, wrapping strategies, custom error types, and production-grade error management. Focus on clear error contexts, proper error chains, and idiomatic Go error patterns.

## Instructions
## Error Handling Stack

- **Error Wrapping** — fmt.Errorf with %w for wrapping
- **Custom Errors** — Error types with methods (Error(), Is(), Unwrap())
- **Sentinel Errors** — errors.New() for comparison
- **Error Chains** — errors.Is() and errors.As() for inspection
- **Domain Errors** — Package-level error types for business logic

## Error Wrapping Pattern

```go
package user

import (
    "fmt"
)

func (r *Repository) GetUser(id string) (*User, error) {
    row := r.db.QueryRow("SELECT id, name FROM users WHERE id = $1", id)
    
    var u User
    if err := row.Scan(&u.ID, &u.Name); err != nil {
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    
    return &u, nil
}
```

**Key points**:
- Wrap errors with context using %w
- Add operation context at each layer
- Don't wrap sentinel errors from your own package
- Preserve error types for Is() and As()

## Custom Error Types

```go
package user

import "fmt"

type InvalidInputError struct {
    Field   string
    Message string
}

func (e *InvalidInputError) Error() string {
    return fmt.Sprintf("invalid input: field '%s' %s", e.Field, e.Message)
}

type UserNotFoundError struct {
    ID string
}

func (e *UserNotFoundError) Error() string {
    return fmt.Sprintf("user not found: %s", e.ID)
}

func (e *UserNotFoundError) Is(target error) bool {
    _, ok := target.(*UserNotFoundError)
    return ok
}
```

**Usage**:
```go
if err != nil {
    if _, ok := err.(*UserNotFoundError); ok {
        return nil, nil
    }
    return err
}
```

## Sentinel Errors

```go
package user

import "errors"

var (
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidInput = errors.New("invalid input")
    ErrUserExists   = errors.New("user already exists")
)

func (r *Repository) GetUser(id string) (*User, error) {
    var u User
    err := r.db.QueryRow("SELECT id, name FROM users WHERE id = $1", id).Scan(&u.ID, &u.Name)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("query user: %w", err)
    }
    return &u, nil
}
```

**Comparison**:
```go
if errors.Is(err, ErrUserNotFound) {
    // Handle not found
}
```

## Error Chains

```go
func (s *Service) CreateUser(req CreateUserRequest) (*User, error) {
    if err := s.validate(req); err != nil {
        return nil, fmt.Errorf("validate request: %w", err)
    }
    
    u, err := s.repo.Create(req)
    if err != nil {
        return nil, fmt.Errorf("create user in repository: %w", err)
    }
    
    if err := s.notify(u); err != nil {
        return nil, fmt.Errorf("notify user: %w", err)
    }
    
    return u, nil
}
```

**Unwrapping**:
```go
err := s.CreateUser(req)

if errors.Is(err, user.ErrUserExists) {
    // Check specific sentinel
}

var notFound *user.UserNotFoundError
if errors.As(err, &notFound) {
    // Extract custom error details
    fmt.Printf("User not found: %s\n", notFound.ID)
}

fmt.Printf("Full chain: %v\n", err)
```

## Domain-Specific Errors

```go
package payment

import "fmt"

type PaymentError struct {
    Code    string
    Message string
    Cause   error
}

func (e *PaymentError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("payment error [%s]: %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("payment error [%s]: %s", e.Code, e.Message)
}

func (e *PaymentError) Unwrap() error {
    return e.Cause
}

func (e *PaymentError) Is(target error) bool {
    t, ok := target.(*PaymentError)
    return ok && e.Code == t.Code
}

var (
    ErrInsufficientFunds = &PaymentError{Code: "INSUFFICIENT_FUNDS", Message: "insufficient funds"}
    ErrPaymentFailed     = &PaymentError{Code: "PAYMENT_FAILED", Message: "payment failed"}
    ErrInvalidCard       = &PaymentError{Code: "INVALID_CARD", Message: "invalid card"}
)
```

## Error Handling Best Practices

1. **Wrap errors with context** — Add operation context at each layer
2. **Use %w for wrapping** — Preserves error type for Is() and As()
3. **Sentinel errors** — For predictable errors (not found, invalid input)
4. **Custom errors** — For domain-specific logic with additional data
5. **Don't check errors twice** — Handle error where it occurs or return it
6. **Handle errors immediately** — Don't defer error handling
7. **Log at appropriate layer** — Don't log and wrap (double logging)
8. **Provide useful context** — Include relevant values in error messages

## Error Documentation

```go
package user

import "errors"

var (
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidEmail = errors.New("invalid email format")
    ErrUserExists   = errors.New("user already exists")
)

type InvalidInputError struct {
    Field   string
    Message string
}

func (e *InvalidInputError) Error() string {
    return fmt.Sprintf("invalid input: %s: %s", e.Field, e.Message)
}

func (e *InvalidInputError) Is(target error) bool {
    _, ok := target.(*InvalidInputError)
    return ok
}
```

## Error Testing

```go
func TestGetUser_NotFound(t *testing.T) {
    repo := NewTestRepository()
    _, err := repo.GetUser("nonexistent")
    
    if !errors.Is(err, ErrUserNotFound) {
        t.Errorf("expected ErrUserNotFound, got %v", err)
    }
}

func TestCreateUser_InvalidInput(t *testing.T) {
    repo := NewTestRepository()
    _, err := repo.CreateUser(User{Email: "invalid"})
    
    var invalid *InvalidInputError
    if !errors.As(err, &invalid) {
        t.Fatalf("expected InvalidInputError, got %v", err)
    }
    
    if invalid.Field != "email" {
        t.Errorf("expected field 'email', got %s", invalid.Field)
    }
}
```

## Constraints

- Include error wrapping with fmt.Errorf and %w verb
- Include custom error types with Error() and optional Is()/Unwrap() methods
- Include sentinel errors using errors.New()
- Include error chain inspection with errors.Is() and errors.As()
- Include domain-specific error patterns
- Include error testing patterns
- Exclude wrapping errors without adding context
- Exclude double logging (log and wrap)
- Exclude checking errors twice (handle once)
- Exclude using error messages for control flow (use error types)
- Exclude returning error without context in outer layers
- Exclude creating custom errors when sentinel errors suffice

## Edge Cases

If error occurs at domain boundary: Wrap with context describing the operation that failed.

If error is already wrapped: Use %w to preserve the error chain for Is()/As() inspection.

If error should be compared: Use sentinel errors or implement Is() method for custom types.

If error contains sensitive data: Redact or omit from error message, log separately with debug level.

If multiple errors occur: Collect and return aggregate error or return first with context about failures.

If error should be ignored: Check with errors.Is() or errors.As() and handle specifically, don't use string matching.

If error needs additional data: Use custom error type with fields, don't encode in error message.

If error occurs in goroutine: Return via channel or sync.ErrGroup with context.

If error handling requires retry: Implement retry logic with backoff, don't rely on error message parsing.

## Examples
<example>
<input>Implement error wrapping across layers</input>
<output>
```go
// Domain layer - custom error types
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Repository layer - wrap infrastructure errors
func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
    var user User
    err := r.pool.QueryRow(ctx, "SELECT * FROM users WHERE email = $1", email).Scan(&user.ID, &user.Email)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, contract.ErrNotFound
        }
        return nil, fmt.Errorf("query user by email: %w", err)
    }
    return &user, nil
}

// UseCase layer - business context
func (uc *CreateUserUC) Execute(ctx context.Context, req CreateUserReq) (*CreateUserResp, error) {
    if req.Email == "" {
        return nil, &ValidationError{Field: "email", Message: "required"}
    }
    
    user, err := uc.userRepo.FindByEmail(ctx, req.Email)
    if err != nil && !errors.Is(err, contract.ErrNotFound) {
        return nil, fmt.Errorf("check existing user: %w", err)
    }
    if user != nil {
        return nil, contract.ErrConflict
    }
    
    // Create user...
    if err := uc.userRepo.Save(ctx, newUser); err != nil {
        return nil, fmt.Errorf("save user: %w", err)
    }
    
    return &CreateUserResp{ID: newUser.ID}, nil
}
```

**Pattern**: Domain errors at top, wrap with context at each layer, map infrastructure errors to domain errors, use errors.Is() for comparison.
</output>
</example>

For additional error handling examples, see:
- `references/error-wrapping.md` - Repository layer error wrapping
- `references/custom-errors.md` - Domain validation error types
- `references/sentinel-errors.md` - Sentinel errors with errors.Is()
- `references/error-chains.md` - Error chains with errors.As()
- `references/layered-errors.md` - Cross-layer error propagation

## Output Format

Provide error handling guidance with the following structure:

1. **Error Patterns**: Wrapping, custom errors, sentinel errors, domain errors
2. **Error Wrapping**: Use fmt.Errorf with %w, add operation context
3. **Custom Types**: Implement Error() with optional Is()/Unwrap()
4. **Sentinel Errors**: Use errors.New() for comparison with errors.Is()
5. **Error Chains**: Inspect with errors.Is() and errors.As()
6. **Error Context**: Add relevant operation context at each layer
7. **Error Testing**: Test with errors.Is() and errors.As()
8. **Examples**: Complete, runnable code showing error patterns

Focus on production-ready error handling with clear context and proper error chains.

