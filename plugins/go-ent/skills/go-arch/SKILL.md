---
name: go-arch
description: "Clean Architecture, DDD, microservices patterns for Go. Auto-activates for: architecture decisions, system design, layer organization, dependency injection, bounded contexts."
---

# Go Architecture

## Layer Structure

```
internal/
├── domain/           # ZERO external deps, NO tags
│   ├── entity/
│   ├── contract/     # Interfaces (repos, services)
│   └── event/
├── usecase/          # Business orchestration
├── repository/       # Data access
│   └── {store}/pgx/
├── transport/        # HTTP/gRPC handlers
│   └── http/
└── app/              # Bootstrap, DI
    ├── app.go
    ├── di.go
    └── uc.go
```

## Dependency Rule

```
Transport → UseCase → Domain ← Repository ← Infrastructure
```

## DI Container

```go
type container struct {
    infra *infraDeps
    repos *repoDeps
    ucs   *usecaseDeps
}

func newContainer(cfg *config.Config, log *slog.Logger) (*container, error) {
    c := &container{}
    if err := c.buildInfra(cfg); err != nil {
        return nil, fmt.Errorf("infra: %w", err)
    }
    c.buildRepos()
    c.buildUseCases(log)
    return c, nil
}
```

## Transaction Manager

```go
type TxManager interface {
    WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *txManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
    tx, err := m.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin: %w", err)
    }
    
    if err := fn(injectTx(ctx, tx)); err != nil {
        tx.Rollback(ctx)
        return err
    }
    return tx.Commit(ctx)
}
```

## Outbox Pattern

```go
func (uc *createOrderUC) Execute(ctx context.Context, req CreateOrderReq) error {
    return uc.tx.WithTx(ctx, func(ctx context.Context) error {
        if err := uc.orderRepo.Save(ctx, order); err != nil {
            return fmt.Errorf("save order: %w", err)
        }
        return uc.outbox.Save(ctx, &Outbox{
            Topic:   "orders.created",
            Payload: mustMarshal(OrderCreated{ID: order.ID}),
        })
    })
}
```

## Decision Matrix

| Scenario       | Pattern              |
|----------------|----------------------|
| Simple CRUD    | Clean Architecture   |
| Complex domain | DDD bounded contexts |
| Cross-service  | Event-driven, outbox |
| High load      | CQRS, read replicas  |

## Graceful Shutdown

```go
func (a *app) Shutdown(ctx context.Context) error {
    a.httpSrv.Shutdown(ctx)
    a.workers.Stop()
    a.container.close()
    return nil
}
```
