# Dependency Injection

Production-grade dependency injection patterns in Go: constructor injection (preferred), Wire code generation, and fx runtime DI.

## Quick Reference

| Pattern                    | Use When                                      |
|----------------------------|-----------------------------------------------|
| Constructor Injection      | Default choice, explicit dependencies         |
| `google/wire`              | Many dependencies, compile-time safety        |
| `uber-go/fx`               | Complex lifecycle, hot reload, plugin systems |
| Interface at consumer      | Testing, swappable implementations            |
| Accept interface, return struct | API design, flexibility vs concreteness  |

## Constructor Injection (Preferred)

Manual dependency injection using constructor functions. No magic, explicit dependencies, easy to trace.

### Basic Pattern

```go
package service

import (
    "context"
    "log/slog"
    "github.com/jackc/pgx/v5/pgxpool"
)

// Dependencies are explicit struct fields
type UserService struct {
    pool   *pgxpool.Pool
    logger *slog.Logger
    cache  Cache
}

// Constructor accepts all dependencies
func New(pool *pgxpool.Pool, logger *slog.Logger, cache Cache) *UserService {
    return &UserService{
        pool:   pool,
        logger: logger,
        cache:  cache,
    }
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    // Use dependencies
    s.logger.InfoContext(ctx, "fetching user", "id", id)
    // ...
}
```

### Application Wiring

```go
// cmd/server/main.go
package main

func main() {
    ctx := context.Background()

    // Infrastructure layer
    cfg := config.LoadFromEnv(os.Getenv)
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
    if err != nil {
        logger.Error("failed to connect to database", "error", err)
        os.Exit(1)
    }
    defer pool.Close()

    cache := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
    defer cache.Close()

    // Repository layer
    userRepo := postgres.NewUserRepository(pool)
    orderRepo := postgres.NewOrderRepository(pool)

    // Use Case layer
    userSvc := user.New(userRepo, cache, logger)
    orderSvc := order.New(orderRepo, userRepo, logger)

    // Transport layer
    httpServer := http.NewServer(cfg.HTTPPort, userSvc, orderSvc, logger)

    // Start server
    if err := httpServer.ListenAndServe(); err != nil {
        logger.Error("server failed", "error", err)
        os.Exit(1)
    }
}
```

### Multiple Constructors

```go
// Production constructor
func New(pool *pgxpool.Pool, logger *slog.Logger) *Service {
    return &Service{
        pool:   pool,
        logger: logger,
        timeout: 30 * time.Second,
    }
}

// Test constructor with overrides
func newForTest(pool *pgxpool.Pool) *Service {
    return &Service{
        pool:    pool,
        logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
        timeout: 1 * time.Second,
    }
}
```

### Functional Options Pattern

```go
type Service struct {
    pool    *pgxpool.Pool
    logger  *slog.Logger
    timeout time.Duration
    retries int
}

type Option func(*Service)

func WithTimeout(d time.Duration) Option {
    return func(s *Service) {
        s.timeout = d
    }
}

func WithRetries(n int) Option {
    return func(s *Service) {
        s.retries = n
    }
}

func New(pool *pgxpool.Pool, logger *slog.Logger, opts ...Option) *Service {
    s := &Service{
        pool:    pool,
        logger:  logger,
        timeout: 30 * time.Second, // defaults
        retries: 3,
    }

    for _, opt := range opts {
        opt(s)
    }

    return s
}

// Usage
svc := service.New(
    pool,
    logger,
    service.WithTimeout(10*time.Second),
    service.WithRetries(5),
)
```

## Wire (google/wire)

Code generation for compile-time dependency injection. Generates wiring code, catches errors at build time.

### Installation

```bash
go install github.com/google/wire/cmd/wire@latest
```

### Basic Usage

```go
// internal/di/wire.go
//go:build wireinject
// +build wireinject

package di

import (
    "context"
    "log/slog"
    "os"

    "github.com/google/wire"
    "github.com/jackc/pgx/v5/pgxpool"

    "myapp/internal/config"
    "myapp/internal/repository/postgres"
    "myapp/internal/service/user"
    "myapp/internal/transport/http"
)

// Provider functions
func provideDB(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
    return pgxpool.New(ctx, cfg.DatabaseURL)
}

func provideLogger() *slog.Logger {
    return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

// Wire set - group related providers
var repositorySet = wire.NewSet(
    postgres.NewUserRepository,
    postgres.NewOrderRepository,
)

var serviceSet = wire.NewSet(
    user.New,
    order.New,
)

// Injector function
func InitializeApp(ctx context.Context, cfg *config.Config) (*http.Server, func(), error) {
    wire.Build(
        provideDB,
        provideLogger,
        repositorySet,
        serviceSet,
        http.NewServer,
    )
    return nil, nil, nil // Wire generates this
}
```

Generate code:

```bash
wire gen ./internal/di
```

Generated file (`wire_gen.go`):

```go
// Code generated by Wire. DO NOT EDIT.
//go:generate go run github.com/google/wire/cmd/wire

package di

func InitializeApp(ctx context.Context, cfg *config.Config) (*http.Server, func(), error) {
    pool, err := provideDB(ctx, cfg)
    if err != nil {
        return nil, nil, err
    }

    logger := provideLogger()
    userRepository := postgres.NewUserRepository(pool)
    orderRepository := postgres.NewOrderRepository(pool)
    userService := user.New(userRepository, logger)
    orderService := order.New(orderRepository, userRepository, logger)
    server := http.NewServer(cfg.HTTPPort, userService, orderService, logger)

    return server, func() {
        pool.Close()
    }, nil
}
```

### Wire with Cleanup

```go
// Provider with cleanup
func provideDB(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, func(), error) {
    pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
    if err != nil {
        return nil, nil, err
    }

    cleanup := func() {
        pool.Close()
    }

    return pool, cleanup, nil
}

// Injector
func InitializeApp(ctx context.Context, cfg *config.Config) (*http.Server, func(), error) {
    wire.Build(
        provideDB,
        provideLogger,
        repositorySet,
        serviceSet,
        http.NewServer,
    )
    return nil, nil, nil
}

// Usage in main
func main() {
    cfg := config.LoadFromEnv(os.Getenv)
    server, cleanup, err := di.InitializeApp(context.Background(), cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer cleanup()

    server.Run()
}
```

### Wire Sets and Interfaces

```go
// internal/repository/postgres/providers.go
var Set = wire.NewSet(
    NewUserRepository,
    wire.Bind(new(user.Repository), new(*UserRepository)),
)

// NewUserRepository returns concrete type
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
    return &UserRepository{pool: pool}
}

// Wire binds concrete type to interface
// wire.Bind(new(Interface), new(*ConcreteType))
```

### Wire Build Constraints

```go
// internal/di/wire.go - only for wire generation
//go:build wireinject
// +build wireinject

// internal/di/wire_gen.go - actual implementation
//go:build !wireinject
// +build !wireinject
```

## Fx (uber-go/fx)

Runtime dependency injection with lifecycle management. Best for complex applications requiring hot reload or plugin systems.

### Basic Usage

```go
package main

import (
    "context"
    "go.uber.org/fx"
    "go.uber.org/fx/fxevent"
    "go.uber.org/zap"
)

func main() {
    fx.New(
        fx.Provide(
            provideLogger,
            provideDB,
            postgres.NewUserRepository,
            user.New,
            http.NewServer,
        ),
        fx.Invoke(runServer),
    ).Run()
}

func provideLogger() *zap.Logger {
    logger, _ := zap.NewProduction()
    return logger
}

func provideDB(lc fx.Lifecycle, cfg *config.Config, logger *zap.Logger) *pgxpool.Pool {
    pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
    if err != nil {
        logger.Fatal("failed to connect to database", zap.Error(err))
    }

    // Lifecycle hooks
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            logger.Info("database connected")
            return pool.Ping(ctx)
        },
        OnStop: func(ctx context.Context) error {
            logger.Info("closing database connection")
            pool.Close()
            return nil
        },
    })

    return pool
}

func runServer(lc fx.Lifecycle, server *http.Server, logger *zap.Logger) {
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            go func() {
                logger.Info("starting http server", zap.Int("port", server.Port))
                if err := server.ListenAndServe(); err != nil {
                    logger.Error("server error", zap.Error(err))
                }
            }()
            return nil
        },
        OnStop: func(ctx context.Context) error {
            logger.Info("stopping http server")
            return server.Shutdown(ctx)
        },
    })
}
```

### Fx Modules

```go
// internal/di/modules.go
package di

import (
    "go.uber.org/fx"
    "myapp/internal/repository/postgres"
    "myapp/internal/service/user"
    "myapp/internal/service/order"
)

// Repository module
var RepositoryModule = fx.Module("repository",
    fx.Provide(
        postgres.NewUserRepository,
        postgres.NewOrderRepository,
    ),
)

// Service module
var ServiceModule = fx.Module("service",
    fx.Provide(
        user.New,
        order.New,
    ),
)

// HTTP module
var HTTPModule = fx.Module("http",
    fx.Provide(http.NewServer),
    fx.Invoke(registerRoutes),
)

func registerRoutes(server *http.Server, userSvc *user.Service) {
    server.RegisterRoutes(userSvc)
}
```

Main application:

```go
func main() {
    fx.New(
        fx.Provide(provideConfig, provideLogger, provideDB),
        di.RepositoryModule,
        di.ServiceModule,
        di.HTTPModule,
    ).Run()
}
```

### Fx Value Groups

```go
// Multiple implementations of same interface
type Handler interface {
    Handle(ctx context.Context) error
}

// Provide multiple handlers
func provideUserHandler() Handler {
    return &UserHandler{}
}

func provideOrderHandler() Handler {
    return &OrderHandler{}
}

// Application
fx.New(
    fx.Provide(
        fx.Annotate(provideUserHandler, fx.ResultTags(`group:"handlers"`)),
        fx.Annotate(provideOrderHandler, fx.ResultTags(`group:"handlers"`)),
    ),
    fx.Invoke(func(handlers []Handler) {
        // All handlers injected as slice
        for _, h := range handlers {
            h.Handle(context.Background())
        }
    }),
)
```

### Fx Named Values

```go
fx.New(
    fx.Provide(
        fx.Annotate(
            providePostgresPool,
            fx.ResultTags(`name:"postgres"`),
        ),
        fx.Annotate(
            provideRedisClient,
            fx.ResultTags(`name:"redis"`),
        ),
    ),
    fx.Invoke(func(
        postgres *pgxpool.Pool `name:"postgres"`,
        redis *redis.Client `name:"redis"`,
    ) {
        // Named dependencies
    }),
)
```

### When to Use Fx

**Use fx when:**
- Large application with complex lifecycle (50+ dependencies)
- Hot reload/plugin system required
- Multiple environments with different wiring
- Shared infrastructure across microservices

**Avoid fx when:**
- Simple application (< 20 dependencies) - use constructor injection
- Startup time critical (fx has runtime overhead)
- Team unfamiliar with DI frameworks

## Interface Design

### Accept Interfaces, Return Structs

```go
// Repository interface in domain layer
package user

type Repository interface {
    GetByID(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, user *User) error
}

// Service accepts interface
type Service struct {
    repo Repository
}

// Constructor accepts interface
func New(repo Repository) *Service {
    return &Service{repo: repo}
}

// Implementation returns concrete type
package postgres

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
    return &UserRepository{pool: pool}
}
```

### Consumer-Side Interfaces

Interfaces belong where they're used, not where they're implemented.

```go
// ✓ Good - interface in consumer package
package user

type Repository interface {
    GetByID(ctx context.Context, id string) (*User, error)
}

type Service struct {
    repo Repository
}

// Implementation in separate package
package postgres

import "myapp/internal/domain/user"

type UserRepository struct {
    pool *pgxpool.Pool
}

// Implements user.Repository
func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
    // ...
}

// ✗ Bad - interface with implementation
package postgres

type Repository interface {
    GetByID(ctx context.Context, id string) (*User, error)
}

type UserRepository struct {
    pool *pgxpool.Pool
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
    // ...
}
```

### Avoid God Interfaces

```go
// ✗ Bad - god interface
type Repository interface {
    CreateUser(ctx context.Context, user *User) error
    GetUser(ctx context.Context, id string) (*User, error)
    UpdateUser(ctx context.Context, user *User) error
    DeleteUser(ctx context.Context, id string) error
    CreateOrder(ctx context.Context, order *Order) error
    GetOrder(ctx context.Context, id string) (*Order, error)
    // ... 50 more methods
}

// ✓ Good - focused interfaces
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}

type OrderRepository interface {
    Create(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, id string) (*Order, error)
}
```

### Interface Segregation

```go
// Split large interfaces by use case
type Reader interface {
    GetByID(ctx context.Context, id string) (*User, error)
    List(ctx context.Context, limit int) ([]*User, error)
}

type Writer interface {
    Save(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}

// Compose when both needed
type Repository interface {
    Reader
    Writer
}

// Service only needs Reader
type QueryService struct {
    repo Reader
}

// Service needs both
type CommandService struct {
    repo Repository
}
```

## Testing with DI

### Constructor Injection Testing

```go
// Service
type UserService struct {
    repo UserRepository
}

func New(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}

// Test with mock
func TestUserService_GetUser(t *testing.T) {
    mockRepo := &MockUserRepository{
        GetByIDFunc: func(ctx context.Context, id string) (*User, error) {
            return &User{ID: id, Name: "Alice"}, nil
        },
    }

    svc := New(mockRepo)
    user, err := svc.GetUser(context.Background(), "123")

    require.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
}
```

### Wire Testing

```go
// internal/di/wire_test.go
//go:build wireinject
// +build wireinject

package di

func InitializeTestApp(ctx context.Context) (*user.Service, func(), error) {
    wire.Build(
        provideTestDB,        // In-memory or testcontainer
        provideTestLogger,    // Discard logger
        postgres.NewUserRepository,
        user.New,
    )
    return nil, nil, nil
}

// Test
func TestUserService(t *testing.T) {
    svc, cleanup, err := di.InitializeTestApp(t.Context())
    require.NoError(t, err)
    defer cleanup()

    // Test with real dependencies
    user, err := svc.CreateUser(t.Context(), "Alice", "alice@example.com")
    require.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

### Fx Testing

```go
func TestWithFx(t *testing.T) {
    var svc *user.Service

    app := fx.New(
        fx.Provide(
            provideTestDB,
            provideTestLogger,
            postgres.NewUserRepository,
            user.New,
        ),
        fx.Populate(&svc),
        fx.NopLogger, // Disable fx logs
    )

    require.NoError(t, app.Start(context.Background()))
    defer app.Stop(context.Background())

    // Test with injected service
    user, err := svc.CreateUser(context.Background(), "Alice", "alice@example.com")
    require.NoError(t, err)
}
```

## Common Mistakes

| Mistake                          | Why It's Bad                            | Fix                                    |
|----------------------------------|-----------------------------------------|----------------------------------------|
| Service locator pattern          | Hidden dependencies, hard to test       | Explicit constructor injection         |
| Global state/singletons          | Tight coupling, race conditions         | Pass dependencies explicitly           |
| Circular dependencies            | Indicates poor design                   | Refactor to separate concerns          |
| Interface in implementation pkg  | Wrong ownership, tight coupling         | Move interface to consumer package     |
| Premature abstraction            | Unnecessary complexity                  | Start with concrete types, add interfaces when needed |
| God interfaces                   | Violates ISP, forces unnecessary deps   | Split into focused interfaces          |
| Constructor does work            | Fails silently, hard to test            | Constructors only assign fields, defer work to Init() or Start() |
| Not using `New()` prefix         | Unclear API, breaks convention          | Name constructors `New()` or `New*()` |
| Returning interface from constructor | Couples to interface, limits flexibility | Return concrete type                  |

## Examples

### Simple App (Constructor Injection)

```go
// cmd/server/main.go
func main() {
    if err := run(context.Background(), os.Getenv, os.Stdout, os.Stderr); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func run(ctx context.Context, getenv func(string) string, stdout, stderr io.Writer) error {
    cfg, err := config.LoadFromEnv(getenv)
    if err != nil {
        return err
    }

    logger := slog.New(slog.NewJSONHandler(stdout, nil))

    pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
    if err != nil {
        return fmt.Errorf("connect to database: %w", err)
    }
    defer pool.Close()

    // Wire dependencies
    userRepo := postgres.NewUserRepository(pool)
    userSvc := user.New(userRepo, logger)
    httpServer := http.NewServer(cfg.HTTPPort, userSvc, logger)

    return httpServer.Run(ctx)
}
```

### Medium App (Wire)

```go
// internal/di/wire.go
//go:build wireinject

package di

var infraSet = wire.NewSet(
    provideConfig,
    provideLogger,
    provideDB,
    provideRedis,
)

var repoSet = wire.NewSet(
    postgres.NewUserRepository,
    postgres.NewOrderRepository,
    wire.Bind(new(user.Repository), new(*postgres.UserRepository)),
    wire.Bind(new(order.Repository), new(*postgres.OrderRepository)),
)

var serviceSet = wire.NewSet(
    user.New,
    order.New,
)

func InitializeApp(ctx context.Context) (*http.Server, func(), error) {
    wire.Build(
        infraSet,
        repoSet,
        serviceSet,
        http.NewServer,
    )
    return nil, nil, nil
}
```

### Large App (Fx)

```go
// cmd/server/main.go
func main() {
    fx.New(
        fx.Provide(
            config.Load,
            provideLogger,
            provideDB,
            provideRedis,
            provideKafka,
        ),
        di.RepositoryModule,
        di.ServiceModule,
        di.HTTPModule,
        di.WorkerModule,
        fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
            return &fxevent.ZapLogger{Logger: log}
        }),
    ).Run()
}
```

## See Also

- [Clean Architecture](./clean-architecture.md) - Layer structure and dependency flow
- [SOLID](./solid.md) - Dependency inversion principle
- [Project Layout](./project-layout.md) - Package organization
- [Mocking](../08-testing/mocking.md) - Testing with dependencies
- [Wire Documentation](https://github.com/google/wire)
- [Fx Documentation](https://uber-go.github.io/fx/)
