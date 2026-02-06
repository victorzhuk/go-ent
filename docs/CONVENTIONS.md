# Conventions

Naming, file organization, and code style conventions for go-ent.

## Overview

This document defines naming conventions, file organization, and code style patterns.

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

### Error Handling Patterns

Happy path left, handle errors immediately:

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

// OK for small structs
user := &User{uuid.New(), "John"}

// BAD - unclear
user := &User{
    uuid.New(),
    "John",
}
```

### Zero Comments Explaining WHAT

Comments explaining what code does indicate poor naming. Rename instead:

```go
// BAD
// Check if user is valid
if user.Valid() {

// GOOD
if user.IsValid() {
```

### WHY Comments Only

Use comments only to explain non-obvious reasoning:

```go
// Using buffered channel to prevent blocking on slow consumers
ch := make(chan Event, 100)

// Must close before deferring to prevent resource leak
defer conn.Close()
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

```go
// Package example manages example functionality.
//
// This package provides example operations for go-ent.
package example

import (
    "context"
    "fmt"
)

const (
    MaxRetries = 3
)

var (
    ErrInvalidInput = errors.New("invalid input")
)

type Example struct {
    id string
}

var globalExample *Example

func New() *Example {
    return &Example{}
}

func (e *Example) DoSomething(ctx context.Context) error {
    // implementation
}

func helper() string {
    return "helper"
}
```

## Best Practices

- Default to private, expose only what's necessary
- Prefer domain-named packages over `util` or `helper`
- Use field names in struct literals for clarity
- Group imports with blank lines
- Keep functions focused and small
- Avoid magic numbers (use named constants)
- Prefer stdlib over external dependencies when possible
- Use meaningful variable names, not verbose AI-style names
