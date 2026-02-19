---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
triggers:
  - architecture
  - system design
  - clean architecture
  - domain driven design
---

# ${SKILL_NAME}

## Role

Architecture specialist covering layered architecture, clean architecture, DDD, microservices, and design patterns. Prioritize clear boundaries, minimal coupling, and data flow clarity when making architectural decisions.

## Instructions

### Layer Boundaries

Dependencies flow inward only:

```
Transport → UseCase → Domain ← Repository ← Infrastructure
```

```go
// Domain layer: pure, no external deps
package domain

type User struct {
    ID    uuid.UUID
    Email string
}

type UserRepository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*User, error)
    Save(ctx context.Context, user *User) error
}
```

```go
// UseCase layer: orchestrates domain + repos
package userusecase

type UseCase struct {
    repo domain.UserRepository
}

func (uc *UseCase) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    return uc.repo.FindByID(ctx, id)
}
```

### DDD Bounded Contexts

```
/internal
  /user           # User bounded context
    /domain       # Entities, value objects, domain interfaces
    /usecase      # Application services
    /repo         # Repository implementations
    /transport    # HTTP/gRPC handlers
```

### Architectural Decision Records

```markdown
# ADR-001: Use PostgreSQL for primary storage

## Status: Accepted

## Context
Need reliable persistence with ACID guarantees and complex queries.

## Decision
Use PostgreSQL with pgx/v5 driver and squirrel query builder.

## Consequences
+ Strong consistency, rich query capabilities
- Requires PostgreSQL infrastructure
```

### Edge Cases

If bounded context boundaries are unclear: Map domain events and aggregates first, then define context boundaries.

If layering violations are found: Introduce interfaces at the boundary to invert dependencies.

If microservices vs monolith decision is needed: Start with modular monolith, extract when deployment or scaling demands it.

## Examples

### Example 1: Clean architecture layer setup

**Input**: Set up clean architecture layers for a payment service

**Output**:
```
/internal/payment
  /domain
    entity.go        # Payment, Transaction structs
    repository.go    # PaymentRepository interface
    service.go       # Domain services
  /usecase
    create.go        # CreatePayment use case
    refund.go        # RefundPayment use case
  /repo
    postgres.go      # PostgreSQL implementation
    models.go        # DB models
    mappers.go       # entity <-> model conversion
  /transport
    http.go          # HTTP handlers
    middleware.go    # Auth, logging middleware
```

### Example 2: Dependency injection setup

**Input**: Set up dependency injection for the application

**Output**:
```go
func New(cfg *config.Config, pool *pgxpool.Pool, log *slog.Logger) (*App, error) {
    // Infrastructure
    userRepo := userrepo.New(pool)
    paymentRepo := paymentrepo.New(pool)

    // Use cases
    userUC := userusecase.New(userRepo, log)
    paymentUC := paymentusecase.New(paymentRepo, userRepo, log)

    // Transport
    router := chi.NewRouter()
    userHandler := userhttp.NewHandler(userUC)
    paymentHandler := paymenthttp.NewHandler(paymentUC)

    userHandler.Register(router)
    paymentHandler.Register(router)

    return &App{router: router, cfg: cfg}, nil
}
```
