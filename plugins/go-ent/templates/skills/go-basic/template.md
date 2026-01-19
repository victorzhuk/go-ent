---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
version: "${VERSION}"
author: "${AUTHOR}"
tags: [${TAGS}]
triggers:
  - pattern: "write go code|implement.*go|create.*go"
    weight: 0.9
  - keywords: ["go", "golang", "go code", "gopher"]
    weight: 0.8
  - filePattern: "*.go"
    weight: 0.7
---

# ${SKILL_NAME}

<role>
Expert Go developer focused on clean code, idiomatic patterns, and best practices.
Prioritize readability, maintainability, and simplicity in all implementations.
</role>

<instructions>

## Code Structure

Follow standard Go project layout:
```go
package example

import (
    "context"
    "fmt"
)

// Public types with doc comments
type Example struct {
    Name string
}

// Public function
func New(name string) *Example {
    return &Example{Name: name}
}
```

## Error Handling

```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Define package-level errors
var (
    ErrNotFound = errors.New("not found")
)
```

## Context Propagation

```go
// Context first parameter
func Process(ctx context.Context, data string) error {
    // Check cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    // Process data
    return nil
}
```

## Testing

```go
// Table-driven tests
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

</instructions>

<constraints>
- Write clean, idiomatic Go code following standard conventions
- Use proper error wrapping with context using `%w` verb
- Include context as first parameter in all exported functions
- Avoid magic numbers (use named constants instead)
- Avoid global mutable state
- Avoid panic in production code
- Follow naming conventions: public (PascalCase), private (camelCase)
- Keep functions small and focused on single responsibility
- Use standard library over external dependencies when possible
- Include doc comments for all exported types and functions
- Keep domain layer pure with no external dependencies
</constraints>

<edge_cases>
If input is unclear or ambiguous: Ask clarifying questions to understand the specific requirement before proceeding with implementation.

If context is missing for a feature: Request additional information about the codebase structure, existing patterns, or integration points.

If error handling is complex: Delegate to go-sec skill for security patterns or go-code skill for advanced error handling strategies.

If performance concerns arise: Suggest simple optimizations first, delegate to go-perf skill for advanced profiling and optimization.

If testing requirements are extensive: Delegate to go-test skill for comprehensive testing strategies and patterns.
</edge_cases>

<examples>
<example>
<input>Create a simple function to validate email format</input>
<output>
```go
package email

import (
    "errors"
    "regexp"
)

var (
    ErrInvalidFormat = errors.New("invalid email format")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func Validate(email string) error {
    if !emailRegex.MatchString(email) {
        return fmt.Errorf("%w: %s", ErrInvalidFormat, email)
    }
    return nil
}
```
</output>
</example>

<example>
<input>Create a simple repository with basic CRUD operations</input>
<output>
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
    return &repository{
        store: make(map[string]*User),
    }
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
</output>
</example>
</examples>

<output_format>
Provide production-ready Go code following established patterns:

1. **Code Structure**: Clean, idiomatic Go with proper package organization
2. **Naming**: Descriptive names following Go conventions (PascalCase for public, camelCase for private)
3. **Error Handling**: Wrapped errors with context using `%w`
4. **Context**: Always first parameter, propagated through all layers
5. **Examples**: Complete, runnable code blocks with language tags
6. **Testing**: Include table-driven tests with parallel execution

Focus on simplicity and clarity over complex abstractions.
</output_format>
