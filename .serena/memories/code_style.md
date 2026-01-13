# Go Ent - Code Style and Conventions

## Imports

Standard library (sorted) → Third-party → Internal with blank lines. Use `make fmt`.

```go
import (
    "context"
    "fmt"

    "github.com/victorzhuk/go-ent/internal/domain"

    "github.com/google/uuid"
)
```

## Naming

- **Types/Constants**: PascalCase (AgentRole, AgentRoleProduct)
- **Functions**: PascalCase exported, camelCase private
- **Variables**: Short but meaningful (cfg, repo, ctx, srv, pool)
- **Receivers**: Single letters (s *Store, c *AgentConfig)
- **Files**: lowercase_with_underscores.go
- **Constructors**: `New()` public, `new*()` private
- **Structs**: Private by default, public only for domain entities

## Error Handling

- Package-level errors in `errors.go`
- Lowercase messages, no trailing punctuation
- Wrap with context: `fmt.Errorf("context: %w", err)`

```go
// Good
return fmt.Errorf("query user %s: %w", id, err)
return fmt.Errorf("create order: %w", err)

// Bad
return fmt.Errorf("Failed to query user: %w", err)  // uppercase
return err  // no context
```

## Types

- Public types with doc comments, PascalCase fields
- Enums: `type Role string`
- Validation: `Valid() bool`, `Validate() error`

## Organization

```
Package doc → Imports → Constants/Errors → Types → Variables → Public funcs → Private funcs
```

## Testing

- Table-driven tests with `t.Run()` and `t.Parallel()`
- Use `testify/assert`
- Files: `filename_test.go`

```go
tests := []struct {
    name string
    want string
}{
    {"valid input", "expected"},
    {"empty input", ""},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // test
    })
}
```

## Interfaces

- Minimal interfaces at consumer side
- Accept interfaces, return structs

## Context

First parameter, pass through all layers for cancellation/timeouts.

## Linting & Formatting

```bash
make lint  # golangci-lint with errcheck, gosec, revive, staticcheck
make fmt    # goimports for import formatting
```

## Architecture Rules

- Domain: ZERO external deps, NO struct tags
- Repository: private models, mappers, return domain entities
- UseCase: private structs/DTOs, orchestrate domain logic
- Transport: private DTOs with validation tags, request/response mapping only
- App: DI container and lifecycle, boot order: infra → repos → services → usecases

## Running Tests

```bash
# Run specific test function
go test -run TestAgentRole_String ./internal/domain

# Run tests for a package
go test ./internal/...

# Verbose mode
go test -v ./internal/domain

# Run with specific flags
go test -race -run TestAgentConfig_Valid ./internal/domain
```

## Before Coding

- Read existing code patterns in the relevant package
- Follow Clean Architecture principles
- Keep changes minimal and simple
- Test after changes: `make lint` and `make test`
