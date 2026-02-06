# Error Handling

Error handling conventions and patterns for go-ent.

## Overview

This document defines error handling patterns used throughout the go-ent codebase.

## Package-Level Errors

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

## Error Wrapping

Always wrap errors with context using `fmt.Errorf` and the `%w` verb:

```go
// Good - provides context
return fmt.Errorf("query user %s: %w", id, err)
return fmt.Errorf("parse task %s: %w", taskID, err)

// Bad - no context
return fmt.Errorf("failed: %w", err)
return err
```

## Error Message Format

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

## Error Type Checking

Use `errors.Is()` for sentinel errors and `errors.As()` for custom error types:

```go
// Sentinel error checking
if errors.Is(err, ErrNotFound) {
    // Handle not found
}

// Custom error type checking
var notFoundErr *NotFoundError
if errors.As(err, &notFoundErr) {
    log.Printf("resource %s not found", notFoundErr.ID)
}
```

## Error Placement

- Repository: Wrap database errors with query context
- UseCase: Wrap errors with business context
- Domain: Define domain-specific error types
- Transport: Map errors to HTTP status codes

## Examples

### Repository Layer

```go
func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
    query := "SELECT id, name FROM users WHERE id = $1"
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

### UseCase Layer

```go
func (uc *createUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
    user, err := domain.NewUser(req.Name)
    if err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }

    if err := uc.repo.Save(ctx, user); err != nil {
        return nil, fmt.Errorf("save user: %w", err)
    }

    return &CreateUserResponse{ID: user.ID}, nil
}
```

### Domain Layer

```go
package user

var (
    ErrEmptyName = errors.New("empty name")
    ErrInvalidEmail = errors.New("invalid email")
)

func NewUser(name, email string) (*User, error) {
    if name == "" {
        return nil, ErrEmptyName
    }
    if !isValidEmail(email) {
        return nil, ErrInvalidEmail
    }
    return &User{
        ID:    generateID(),
        Name:  name,
        Email: email,
    }, nil
}
```
