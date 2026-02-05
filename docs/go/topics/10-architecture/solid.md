# SOLID Principles

SOLID principles applied to Go: designing maintainable, testable, and extensible systems without over-engineering.

## Quick Reference

| Principle                | Go Pattern                                      |
|--------------------------|-------------------------------------------------|
| Single Responsibility    | One reason to change, cohesive packages         |
| Open/Closed              | Extension via interfaces, strategy pattern      |
| Liskov Substitution      | Interface contracts, behavioral substitutability|
| Interface Segregation    | Small focused interfaces (io.Reader pattern)    |
| Dependency Inversion     | Accept interfaces, inject dependencies          |

## Single Responsibility Principle

A type/function/package should have one reason to change. In Go, this means cohesive structs and focused packages.

### One Reason to Change

```go
// ✓ Good - single responsibility: user business logic
type User struct {
    ID        uuid.UUID
    Email     string
    CreatedAt time.Time
}

func (u *User) Validate() error {
    if u.Email == "" {
        return ErrEmptyEmail
    }
    if !strings.Contains(u.Email, "@") {
        return ErrInvalidEmail
    }
    return nil
}

// ✓ Good - single responsibility: user persistence
type UserRepository struct {
    pool *pgxpool.Pool
}

func (r *UserRepository) Save(ctx context.Context, user *User) error {
    query := `INSERT INTO users (id, email, created_at) VALUES ($1, $2, $3)`
    _, err := r.pool.Exec(ctx, query, user.ID, user.Email, user.CreatedAt)
    return err
}

// ✗ Bad - mixed responsibilities: business logic + persistence + presentation
type User struct {
    ID    uuid.UUID
    Email string
}

func (u *User) Validate() error { /* validation */ }
func (u *User) Save(db *sql.DB) error { /* persistence */ }
func (u *User) ToJSON() string { /* serialization */ }
```

### Struct Design

```go
// ✓ Good - focused struct
type OrderService struct {
    repo OrderRepository
    log  *slog.Logger
}

func (s *OrderService) Create(ctx context.Context, req CreateOrderReq) error {
    order, err := NewOrder(req.UserID, req.Items)
    if err != nil {
        return fmt.Errorf("new order: %w", err)
    }

    if err := s.repo.Save(ctx, order); err != nil {
        return fmt.Errorf("save order: %w", err)
    }

    s.log.Info("order created", "id", order.ID)
    return nil
}

// ✗ Bad - too many responsibilities
type OrderService struct {
    db      *sql.DB
    cache   *redis.Client
    mailer  *smtp.Client
    logger  *slog.Logger
    metrics *prometheus.Registry
}

func (s *OrderService) Create(ctx context.Context, req CreateOrderReq) error {
    // Validates, persists, caches, sends email, logs, records metrics
    // Too many reasons to change!
}
```

### Package Cohesion

```go
// ✓ Good - cohesive package structure
// internal/order/
//   entity.go       - Order domain entity
//   repository.go   - Repository interface
//   service.go      - Business logic
//   postgres/
//     repo.go       - PostgreSQL implementation

// ✗ Bad - utility/helper packages
// internal/utils/
//   helpers.go      - Random utility functions
//   common.go       - Shared code without theme

// ✓ Good - domain-named packages
package order

type Order struct { /* domain entity */ }

package notification

func Send(ctx context.Context, msg Message) error { /* focused behavior */ }
```

## Open/Closed Principle

Open for extension, closed for modification. In Go: extend behavior via interfaces without changing existing code.

### Extension via Interfaces

```go
// ✓ Good - extensible via interface
type Notifier interface {
    Notify(ctx context.Context, msg string) error
}

type NotificationService struct {
    notifiers []Notifier
}

func (s *NotificationService) Send(ctx context.Context, msg string) error {
    for _, n := range s.notifiers {
        if err := n.Notify(ctx, msg); err != nil {
            return err
        }
    }
    return nil
}

// Extensions without modifying NotificationService
type EmailNotifier struct {
    smtp *smtp.Client
}

func (n *EmailNotifier) Notify(ctx context.Context, msg string) error {
    return n.smtp.Send(ctx, msg)
}

type SlackNotifier struct {
    webhook string
}

func (n *SlackNotifier) Notify(ctx context.Context, msg string) error {
    return postToSlack(n.webhook, msg)
}

// ✗ Bad - modification required for new notifiers
type NotificationService struct {
    useEmail bool
    useSlack bool
    useSMS   bool
}

func (s *NotificationService) Send(ctx context.Context, msg string) error {
    // Need to modify this function for each new channel
    if s.useEmail { /* send email */ }
    if s.useSlack { /* send slack */ }
    if s.useSMS   { /* send SMS */ }
}
```

### Strategy Pattern

```go
// ✓ Good - payment strategy
type PaymentProcessor interface {
    Process(ctx context.Context, amount int64) error
}

type PaymentService struct {
    processor PaymentProcessor
}

func (s *PaymentService) Charge(ctx context.Context, amount int64) error {
    return s.processor.Process(ctx, amount)
}

// Strategies
type StripeProcessor struct {
    apiKey string
}

func (p *StripeProcessor) Process(ctx context.Context, amount int64) error {
    // Stripe-specific logic
}

type PayPalProcessor struct {
    clientID string
}

func (p *PayPalProcessor) Process(ctx context.Context, amount int64) error {
    // PayPal-specific logic
}

// Usage - change strategy without modifying PaymentService
service := &PaymentService{
    processor: &StripeProcessor{apiKey: key},
}
```

### Plugin Architecture

```go
// ✓ Good - plugin system
type Plugin interface {
    Name() string
    Execute(ctx context.Context, data []byte) error
}

type PluginRegistry struct {
    plugins map[string]Plugin
}

func (r *PluginRegistry) Register(p Plugin) {
    r.plugins[p.Name()] = p
}

func (r *PluginRegistry) Execute(ctx context.Context, name string, data []byte) error {
    p, ok := r.plugins[name]
    if !ok {
        return ErrPluginNotFound
    }
    return p.Execute(ctx, data)
}

// New plugins extend without modifying registry
type LoggerPlugin struct{}

func (p *LoggerPlugin) Name() string { return "logger" }
func (p *LoggerPlugin) Execute(ctx context.Context, data []byte) error {
    slog.Info("plugin executed", "data", string(data))
    return nil
}
```

## Liskov Substitution Principle

Subtypes must be substitutable for their base types. In Go: implementations must honor interface contracts.

### Interface Contracts

```go
// ✓ Good - all implementations honor contract
type UserRepository interface {
    // Returns user or ErrNotFound
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

// PostgreSQL implementation
type PostgresUserRepo struct {
    pool *pgxpool.Pool
}

func (r *PostgresUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
    var user User
    err := r.pool.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", id).Scan(&user)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrNotFound // Contract honored
    }
    return &user, err
}

// In-memory implementation
type InMemoryUserRepo struct {
    users map[uuid.UUID]*User
}

func (r *InMemoryUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
    user, ok := r.users[id]
    if !ok {
        return nil, ErrNotFound // Contract honored
    }
    return user, nil
}

// ✗ Bad - violates contract (different error semantics)
type BrokenUserRepo struct{}

func (r *BrokenUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
    return nil, errors.New("user does not exist") // NOT ErrNotFound!
}
```

### Error Handling Consistency

```go
// ✓ Good - consistent error semantics
type Storage interface {
    Get(ctx context.Context, key string) ([]byte, error)
}

var ErrKeyNotFound = errors.New("key not found")

// Redis implementation
type RedisStorage struct {
    client *redis.Client
}

func (s *RedisStorage) Get(ctx context.Context, key string) ([]byte, error) {
    val, err := s.client.Get(ctx, key).Bytes()
    if errors.Is(err, redis.Nil) {
        return nil, ErrKeyNotFound
    }
    return val, err
}

// File system implementation
type FileStorage struct {
    basePath string
}

func (s *FileStorage) Get(ctx context.Context, key string) ([]byte, error) {
    data, err := os.ReadFile(filepath.Join(s.basePath, key))
    if errors.Is(err, os.ErrNotExist) {
        return nil, ErrKeyNotFound
    }
    return data, err
}
```

### Behavioral Substitutability

```go
// ✓ Good - all caches behave the same
type Cache interface {
    Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
    Get(ctx context.Context, key string) ([]byte, error)
}

// In-memory cache
type InMemoryCache struct {
    data sync.Map
}

func (c *InMemoryCache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
    c.data.Store(key, val)
    time.AfterFunc(ttl, func() { c.data.Delete(key) })
    return nil
}

func (c *InMemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
    val, ok := c.data.Load(key)
    if !ok {
        return nil, ErrNotFound
    }
    return val.([]byte), nil
}

// Redis cache - same behavior
type RedisCache struct {
    client *redis.Client
}

func (c *RedisCache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
    return c.client.Set(ctx, key, val, ttl).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
    val, err := c.client.Get(ctx, key).Bytes()
    if errors.Is(err, redis.Nil) {
        return nil, ErrNotFound
    }
    return val, err
}
```

## Interface Segregation Principle

Clients should not depend on interfaces they don't use. In Go: prefer many small interfaces over fat ones.

### Small Focused Interfaces

```go
// ✓ Good - segregated interfaces
type Reader interface {
    Read(ctx context.Context, key string) ([]byte, error)
}

type Writer interface {
    Write(ctx context.Context, key string, data []byte) error
}

type Deleter interface {
    Delete(ctx context.Context, key string) error
}

// Compose as needed
type ReadWriter interface {
    Reader
    Writer
}

// Clients depend only on what they need
func ProcessData(ctx context.Context, r Reader, key string) error {
    data, err := r.Read(ctx, key)
    if err != nil {
        return err
    }
    return process(data)
}

// ✗ Bad - fat interface forces unnecessary dependencies
type Storage interface {
    Read(ctx context.Context, key string) ([]byte, error)
    Write(ctx context.Context, key string, data []byte) error
    Delete(ctx context.Context, key string) error
    List(ctx context.Context, prefix string) ([]string, error)
    Stats(ctx context.Context) (Stats, error)
}

// Clients forced to depend on entire interface
func ProcessData(ctx context.Context, s Storage, key string) error {
    // Only needs Read, but depends on all methods
}
```

### io.Reader/Writer Examples

```go
// ✓ Good - stdlib interfaces are small and composable
import "io"

func CompressData(w io.Writer, data []byte) error {
    gw := gzip.NewWriter(w)
    defer gw.Close()
    _, err := gw.Write(data)
    return err
}

func DecompressData(r io.Reader) ([]byte, error) {
    gr, err := gzip.NewReader(r)
    if err != nil {
        return nil, err
    }
    defer gr.Close()
    return io.ReadAll(gr)
}

// Works with files, buffers, network connections, etc.
var buf bytes.Buffer
CompressData(&buf, data)
DecompressData(&buf)
```

### Avoid Fat Interfaces

```go
// ✓ Good - multiple focused interfaces
type UserReader interface {
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
    List(ctx context.Context, limit int) ([]*User, error)
}

type UserWriter interface {
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
}

// Service depends only on what it needs
type UserQueryService struct {
    reader UserReader
}

type UserCommandService struct {
    writer UserWriter
}

// ✗ Bad - monolithic interface
type UserRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
    List(ctx context.Context, limit int) ([]*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id uuid.UUID) error
    Search(ctx context.Context, query string) ([]*User, error)
    Count(ctx context.Context) (int, error)
}
```

## Dependency Inversion Principle

Depend on abstractions, not concretions. In Go: high-level modules define interfaces, low-level modules implement them.

### Depend on Abstractions

```go
// ✓ Good - high-level module defines interface
package order

// Domain defines what it needs
type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
}

type OrderService struct {
    repo OrderRepository // Depends on abstraction
    log  *slog.Logger
}

func NewOrderService(repo OrderRepository, log *slog.Logger) *OrderService {
    return &OrderService{repo: repo, log: log}
}

// ✗ Bad - depends on concrete implementation
package order

import "internal/postgres" // Concrete dependency

type OrderService struct {
    repo *postgres.OrderRepo // Depends on concrete type
}
```

### Constructor Injection

```go
// ✓ Good - dependencies injected
package main

import (
    "internal/order"
    "internal/order/postgres"
)

func main() {
    pool := setupDB()
    logger := slog.Default()

    // Low-level implements high-level interface
    repo := postgres.NewOrderRepo(pool)
    service := order.NewOrderService(repo, logger)

    // Dependencies flow inward
}

// ✗ Bad - service creates its own dependencies
type OrderService struct {
    repo *OrderRepo
}

func NewOrderService(connString string) *OrderService {
    pool := pgxpool.New(context.Background(), connString)
    return &OrderService{
        repo: &OrderRepo{pool: pool}, // Hard-coded dependency
    }
}
```

### Repository Pattern

```go
// ✓ Good - domain defines repository interface
package user

type User struct {
    ID    uuid.UUID
    Email string
}

// Interface defined by consumer
type Repository interface {
    Save(ctx context.Context, user *User) error
    GetByEmail(ctx context.Context, email string) (*User, error)
}

type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

// Implementation in separate package
package userrepo

import "domain/user"

type postgresRepo struct {
    pool *pgxpool.Pool
}

// Implements user.Repository
func (r *postgresRepo) Save(ctx context.Context, u *user.User) error {
    query := `INSERT INTO users (id, email) VALUES ($1, $2)`
    _, err := r.pool.Exec(ctx, query, u.ID, u.Email)
    return err
}

func (r *postgresRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
    var u user.User
    err := r.pool.QueryRow(ctx,
        `SELECT id, email FROM users WHERE email = $1`,
        email,
    ).Scan(&u.ID, &u.Email)
    return &u, err
}

func NewPostgresRepo(pool *pgxpool.Pool) user.Repository {
    return &postgresRepo{pool: pool}
}
```

## Go-Specific Considerations

### Accept Interfaces, Return Structs

```go
// ✓ Good - accept interface, return struct
func ProcessOrder(ctx context.Context, repo OrderRepository) (*OrderResult, error) {
    order, err := repo.GetNext(ctx)
    if err != nil {
        return nil, err
    }
    return &OrderResult{ID: order.ID}, nil
}

// ✗ Bad - returning interface
func ProcessOrder(ctx context.Context, repo OrderRepository) (OrderRepository, error) {
    // Returning interface hides concrete type
}
```

### Interfaces at Consumer

```go
// ✓ Good - interface defined where it's used
package service

type UserRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type UserService struct {
    repo UserRepository
}

// Implementation elsewhere
package postgres

import "service"

type userRepo struct {
    pool *pgxpool.Pool
}

// Implements service.UserRepository
func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*service.User, error) {
    // ...
}

// ✗ Bad - interface in implementation package
package postgres

type UserRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type userRepo struct{}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {}
```

### Embedding for Composition

```go
// ✓ Good - composition via embedding
type LoggingRepository struct {
    repo OrderRepository
    log  *slog.Logger
}

func (r *LoggingRepository) Save(ctx context.Context, order *Order) error {
    r.log.Info("saving order", "id", order.ID)
    return r.repo.Save(ctx, order)
}

// Extend behavior without modification
func NewLoggingRepository(repo OrderRepository, log *slog.Logger) *LoggingRepository {
    return &LoggingRepository{repo: repo, log: log}
}
```

## Common Mistakes

| Mistake                            | Fix                                       |
|------------------------------------|-------------------------------------------|
| Utility packages (`pkg/utils`)     | Use domain-named packages                 |
| Premature abstraction              | Start concrete, extract when needed       |
| God interfaces (10+ methods)       | Split into focused interfaces             |
| Concrete dependencies in domain    | Depend on interfaces, inject at edges     |
| Leaky abstractions                 | Hide implementation details completely    |
| Interface for every struct         | Create interfaces only when needed        |
| Testing via concrete types         | Test against interfaces                   |

## Best Practices

```go
// ✓ Good patterns

// Single responsibility - one struct, one job
type OrderValidator struct{}
type OrderRepository struct{}
type OrderService struct{}

// Open/closed - extend via interface
type Notifier interface { Notify(ctx context.Context, msg string) error }

// Liskov - honor contracts
func (r *PostgresRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
    // Always returns ErrNotFound for missing users
}

// Interface segregation - small interfaces
type Reader interface { Read(ctx context.Context, key string) ([]byte, error) }
type Writer interface { Write(ctx context.Context, key string, data []byte) error }

// Dependency inversion - accept interfaces
func NewService(repo OrderRepository, logger *slog.Logger) *Service

// ✗ Bad patterns

// Mixed responsibilities
type UserService struct {
    db      *sql.DB
    cache   *redis.Client
    mailer  *smtp.Client
}

// Fat interface
type Repository interface {
    Method1()
    Method2()
    // ... 10 more methods
}

// Concrete dependency
type Service struct {
    repo *PostgresRepo // Should be interface
}
```

## See Also

- [Clean Architecture](./clean-architecture.md) - Layered architecture patterns
- [Dependency Injection](./dependency-injection.md) - DI patterns in Go
- [Idioms](../01-fundamentals/idioms.md) - Accept interfaces, return structs
- [Error Handling](../02-language/error-handling.md) - Error contract patterns
