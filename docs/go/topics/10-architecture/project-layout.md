# Project Layout

Standard Go project structure for enterprise applications.

## Quick Reference

| Aspect | Recommendation | When to Use |
|--------|----------------|-------------|
| **Flat vs Layered** | Flat for <5 packages, layered for complex apps | Flat: libraries, tools. Layered: services, APIs |
| **internal/** | Use for private code (enforced by compiler) | Always for application code, never for libraries |
| **cmd/** | One subdirectory per binary | Multiple executables (server, worker, CLI) |
| **pkg/** | Public libraries only | Reusable code for external projects |
| **Test files** | Alongside source (`*_test.go`) | All unit tests |
| **Integration tests** | Separate package or `test/` directory | Tests requiring external dependencies |
| **Vendor** | Prefer modules, vendor for reproducibility | Air-gapped builds, Docker layer caching |

## Standard Layout

```
myproject/
├── cmd/                    # Main applications
│   ├── server/
│   │   └── main.go
│   └── worker/
│       └── main.go
├── internal/               # Private application code
│   ├── domain/             # Business entities
│   ├── usecase/            # Business logic
│   ├── repository/         # Data access
│   └── transport/          # HTTP/gRPC handlers
├── pkg/                    # Public libraries
├── api/                    # API definitions (OpenAPI, protobuf)
├── configs/                # Configuration files
├── scripts/                # Build, install scripts
├── deployments/            # Docker, k8s manifests
├── test/                   # Integration tests, test data
├── go.mod
└── go.sum
```

## Flat vs Layered

### Flat Structure

**Use when:**
- Project has <5 packages
- Single responsibility (library, tool)
- No layered architecture needed

```
mylib/
├── client.go           # Public API
├── client_test.go
├── config.go
├── internal.go         # Private helpers
└── go.mod
```

### Layered Structure

**Use when:**
- Application has multiple layers (transport, usecase, domain)
- Team size >3 developers
- Bounded contexts or microservices

```
myservice/
├── cmd/server/main.go
├── internal/
│   ├── domain/         # Entities, business rules
│   ├── usecase/        # Business logic
│   ├── repository/     # Data access
│   └── transport/      # HTTP/gRPC
└── go.mod
```

## Internal Structure

### By Layer (Recommended)

```
internal/
├── domain/
│   └── user/
│       ├── user.go         # Entity
│       ├── repository.go   # Interface
│       └── errors.go
├── usecase/
│   └── user/
│       ├── create.go       # CreateUserUseCase
│       └── get.go          # GetUserUseCase
├── repository/
│   └── postgres/
│       ├── user.go         # Repository implementation
│       ├── models.go       # DB models
│       └── mappers.go      # DB ↔ Domain
└── transport/
    ├── http/
    │   ├── user.go         # User handlers
    │   └── dto.go          # Request/response DTOs
    └── grpc/
        └── user.go         # gRPC service
```

### By Feature (Alternative)

```
internal/
├── user/
│   ├── domain.go           # Entity
│   ├── usecase.go          # Business logic
│   ├── repository.go       # Data access
│   └── handler.go          # HTTP handlers
└── order/
    ├── domain.go
    ├── usecase.go
    ├── repository.go
    └── handler.go
```

**Use feature-based when:**
- Microservices with clear bounded contexts
- Features rarely share code
- Small teams (<5 people)

## internal/ Package

The `internal/` directory is enforced by the Go compiler. Code in `internal/` can only be imported by code in the same module or subdirectories of the parent of `internal/`.

**Rules:**
- Always use for application code
- Never use for libraries (use private unexported symbols instead)
- Organize by layer or feature, not by type

```go
// ✅ Allowed
import "github.com/myorg/myproject/internal/domain"

// ❌ Blocked by compiler
import "github.com/otherorg/myproject/internal/domain"
```

### Common Patterns

**API/Domain/Infra:**
```
internal/
├── api/            # Transport layer (HTTP, gRPC)
├── domain/         # Business entities
└── infra/          # Infrastructure (DB, cache, external APIs)
```

**Clean Architecture:**
```
internal/
├── domain/         # Entities + interfaces
├── usecase/        # Business logic
├── repository/     # Data access implementations
└── transport/      # HTTP/gRPC handlers
```

## Multi-Binary Projects

### Structure

```
cmd/
├── server/         # HTTP API server
│   └── main.go
├── worker/         # Background job worker
│   └── main.go
├── migrate/        # Database migrations
│   └── main.go
└── cli/            # Command-line tool
    └── main.go

internal/           # Shared code
├── app/            # Application bootstrap
├── domain/
└── repository/
```

### Main Functions

**cmd/server/main.go:**
```go
package main

import (
    "context"
    "os"

    "github.com/myorg/project/internal/app"
)

func main() {
    ctx := context.Background()
    if err := app.RunServer(ctx, os.Getenv, os.Stdout, os.Stderr); err != nil {
        os.Exit(1)
    }
}
```

**cmd/worker/main.go:**
```go
package main

import (
    "context"
    "os"

    "github.com/myorg/project/internal/app"
)

func main() {
    ctx := context.Background()
    if err := app.RunWorker(ctx, os.Getenv, os.Stdout, os.Stderr); err != nil {
        os.Exit(1)
    }
}
```

## Test Organization

### Unit Tests

Place test files alongside source code:

```
internal/domain/user/
├── user.go
├── user_test.go        # Unit tests
├── repository.go
└── repository_test.go
```

### Integration Tests

**Option 1: Separate package**
```
internal/repository/postgres/
├── user.go
├── user_test.go            # Unit tests
└── user_integration_test.go # Same package, build tag
```

```go
//go:build integration

package postgres_test

import "testing"
```

**Option 2: Dedicated directory**
```
test/
├── integration/
│   ├── user_test.go
│   └── order_test.go
└── testdata/
    └── fixtures.json
```

### Test Data

```
internal/domain/user/
├── user.go
├── user_test.go
└── testdata/
    ├── valid_user.json
    └── invalid_email.json
```

**Access in tests:**
```go
data, _ := os.ReadFile("testdata/valid_user.json")
```

## Vendor vs Modules

### Prefer Modules (Default)

```bash
go mod download      # Download dependencies
go mod verify        # Verify checksums
go mod tidy          # Clean up
```

### When to Vendor

**Use `go mod vendor` when:**
- Air-gapped builds (no internet access)
- Docker layer caching optimization
- Corporate proxy issues
- Reproducible builds (CI/CD)

```bash
go mod vendor        # Create vendor/ directory
go build -mod=vendor # Build using vendored deps
```

**Gitignore:**
```gitignore
vendor/   # Only commit if required
```

## Common Mistakes

| Mistake | Why It's Bad | Fix |
|---------|--------------|-----|
| Deep nesting (`internal/app/service/user/domain/entity/`) | Hard to navigate, cognitive overhead | Max 3-4 levels: `internal/domain/user/` |
| Wrong `internal/` usage in libraries | Prevents users from importing your code | Use unexported symbols instead |
| Mixing layers (`domain/` imports `transport/`) | Breaks dependency rule | Reverse: transport → usecase → domain |
| Circular dependencies (`user` ↔ `order`) | Compilation errors, tight coupling | Extract shared code to separate package |
| Monolithic `cmd/main.go` (1000+ lines) | Hard to test, violates SRP | Extract to `internal/app/` with `run()` function |
| Generic packages (`utils/`, `helpers/`) | Poor cohesion, dumping ground | Use domain-specific names (`user/validation`) |
| No `pkg/` vs `internal/` distinction | Unclear API boundaries | Public libs in `pkg/`, private in `internal/` |

## Package Naming

```go
// ✅ Good - specific domain names
internal/user/
internal/order/
internal/payment/

// ❌ Bad - generic names
internal/models/
internal/utils/
internal/handlers/
```

## See Also

- [Clean Architecture](./clean-architecture.md)
- [Dependency Injection](./dependency-injection.md)
- [SOLID Principles](./solid.md)
- [Integration Testing](../08-testing/integration.md)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
