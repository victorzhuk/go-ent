---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
triggers:
  - go code
  - golang
  - go implementation
---

# ${SKILL_NAME}

## Role

Expert Go developer focused on clean code, idiomatic patterns, and best practices. Prioritize readability, maintainability, and simplicity in all implementations.

## Instructions

### Code Structure

Follow standard Go project layout:

```go
package example

import (
    "context"
    "fmt"
)

// Public types
type Example struct {
    Name string
}

// Public function
func New(name string) *Example {
    return &Example{Name: name}
}
```

### Error Handling

```go
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

var ErrNotFound = errors.New("not found")
```

### Context Propagation

```go
func Process(ctx context.Context, data string) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    return nil
}
```

### Testing

```go
tests := []struct {
    name  string
    input string
    want  string
}{
    {"valid input", "test", "test"},
    {"empty input", "", ""},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        got := Process(tt.input)
        assert.Equal(t, tt.want, got)
    })
}
```

### Edge Cases

If input is unclear or ambiguous: Ask clarifying questions before proceeding.

If performance concerns arise: Suggest simple optimizations first, delegate to go-perf for advanced profiling.

If testing requirements are extensive: Delegate to go-test for comprehensive testing strategies.

## Examples

### Example 1: Validate email format

**Input**: Create a simple function to validate email format

**Output**:
```go
package email

import (
    "errors"
    "regexp"
)

var (
    ErrInvalidFormat = errors.New("invalid email format")
)

var emailRegex = regexp.MustCompile(`+"`"+`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`+"`"+`)

func Validate(email string) error {
    if !emailRegex.MatchString(email) {
        return fmt.Errorf("%w: %s", ErrInvalidFormat, email)
    }
    return nil
}
```

### Example 2: Simple repository with CRUD operations

**Input**: Create a simple repository with basic CRUD operations

**Output**:
```go
package userrepo

import (
    "context"
    "fmt"
)

type User struct {
    ID   string
    Name string
}

type repository struct {
    store map[string]*User
}

func New() *repository {
    return &repository{store: make(map[string]*User)}
}

func (r *repository) Find(ctx context.Context, id string) (*User, error) {
    user, ok := r.store[id]
    if !ok {
        return nil, fmt.Errorf("user not found: %s", id)
    }
    return user, nil
}

func (r *repository) Save(ctx context.Context, user *User) error {
    r.store[user.ID] = user
    return nil
}
```
