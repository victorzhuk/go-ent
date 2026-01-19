---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
version: "${VERSION}"
author: "${AUTHOR}"
tags: [${TAGS}]
triggers:
  - pattern: "architecture|design|system design|component|layer|boundary"
    weight: 0.8
  - keywords: ["architecture", "system design", "microservices", "layered architecture", "clean architecture", "ddd", "domain driven design", "component", "boundary", "data flow"]
    weight: 0.7
  - pattern: "separation|concern|coupling|cohesion"
    weight: 0.7
  - pattern: "bounded context|domain|ubiquitous language"
    weight: 0.8
  - pattern: "adr|architectural decision|pattern"
    weight: 0.7
---

# ${SKILL_NAME}

<role>
Architecture specialist covering layered architecture, clean architecture, DDD, microservices, and design patterns. Prioritize clear boundaries, minimal coupling, and data flow clarity when making architectural decisions.
Apply systematic approaches to component design, dependency management, and architectural evolution.
</role>

<instructions>

## Architecture Fundamentals

### Core Principles

**Separation of Concerns**:
- Each component has a single, well-defined responsibility
- Clear boundaries between layers and modules
- Business logic isolated from infrastructure concerns
- UI separated from business rules

**Dependency Management**:
- Dependencies point inward (Clean Architecture)
- Depend on abstractions, not implementations
- Inversion of Control at boundaries
- Interface-based design for decoupling

**Cohesion and Coupling**:
- High cohesion within components (related code together)
- Low coupling between components (minimal dependencies)
- Stable dependencies principle (depend on more stable things)
- Acyclic dependencies (no circular references)

### Data Flow

**Request Flow (Layered)**:
```
HTTP Request → Transport → UseCase → Domain ← Repository ← Infrastructure
               (DTO)       (DTO)    (Entity)  (Entity)   (Model)
```

**Key Points**:
- Transport: HTTP/gRPC handlers, request validation, response formatting
- UseCase: Orchestration, transaction boundaries, business workflows
- Domain: Pure business logic, entities, value objects, domain services
- Repository: Data access abstraction, queries, persistence
- Infrastructure: External systems, databases, message queues

**Response Flow**:
- Domain entities mapped to transport DTOs
- UseCase returns domain objects, not infrastructure types
- Transport handles serialization and HTTP concerns
- No cross-layer type leakage

## Layered Architecture

### Standard Layers

**Transport Layer** (`/internal/transport/`):
- HTTP/gRPC handlers and middleware
- Request/response DTOs
- Input validation
- Serialization/deserialization
- HTTP concerns (status codes, headers, CORS)

**UseCase Layer** (`/internal/usecase/`):
- Application workflows and orchestration
- Transaction boundaries
- Business use cases
- Domain entity coordination
- Error handling at application level

**Domain Layer** (`/internal/domain/`):
- Business entities and value objects
- Domain services (pure business logic)
- Business rules and invariants
- Repository interfaces
- NO external dependencies

**Repository Layer** (`/internal/repository/`):
- Data access implementation
- ORM/database operations
- Query building and execution
- Persistence concerns
- Mappers between domain and infrastructure types

**Infrastructure Layer** (`/internal/infrastructure/` or `/pkg/`):
- External service clients
- Database drivers
- Message queue clients
- Third-party integrations
- Technical utilities

### Dependency Rules

```
Transport → UseCase → Domain ← Repository ← Infrastructure
          ↓           ↑         ↓
         (DTO)      (Entity)  (Interface)
```

**Allowed Dependencies**:
- Transport can use UseCase interfaces
- UseCase can use Domain and Repository interfaces
- Domain has NO dependencies (pure)
- Repository implements Domain interfaces, uses Infrastructure
- Infrastructure is independent

**Forbidden**:
- NO Domain → Transport dependencies
- NO UseCase → Infrastructure direct access (go through Repository)
- NO Domain → Infrastructure (Repository pattern enforces this)
- NO Repository → UseCase (dependency inversion)

### Layer Organization

**Package Structure**:
```
internal/
├── transport/
│   ├── http/
│   │   ├── handlers.go      # HTTP handlers
│   │   ├── dto.go           # Request/response DTOs
│   │   └── middleware.go   # HTTP middleware
│   └── grpc/
│       └── handlers.go
├── usecase/
│   ├── order.go            # UseCase interface + implementation
│   ├── user.go
│   └── dto.go              # Internal UseCase DTOs
├── domain/
│   ├── order.go            # Entities, value objects
│   ├── user.go
│   └── repository.go       # Repository interfaces
├── repository/
│   ├── order/
│   │   ├── repo.go         # Repository implementation
│   │   ├── models.go       # DB models
│   │   └── mappers.go      # Domain ↔ DB mapping
│   └── user/
│       └── ...
└── infrastructure/
    ├── db/                 # Database setup
    ├── mq/                 # Message queue
    └── cache/              # Cache implementation
```

## Clean Architecture

### Concentric Circles

**Outer Circles (Mechanisms)**:
- Frameworks & Drivers (HTTP, DB, UI)
- Interface Adapters (Controllers, Presenters, Gateways)
- Use Cases (Application Business Rules)

**Inner Circles (Entities)**:
- Enterprise Business Rules (core domain)

### Key Principles

**Rule of Dependencies**:
- Inner circles know nothing about outer circles
- Data and code flow inward only
- Outer circles implement interfaces defined by inner circles
- No coupling to frameworks in inner layers

**Benefits**:
- Testable (inner layers have no framework dependencies)
- Framework-agnostic business logic
- Easy to swap implementations (DB, HTTP framework)
- Independent of UI, database, frameworks, agencies

### Implementation Pattern

**Domain Entities** (innermost):
```go
// internal/domain/order/entity.go
package order

type Order struct {
    ID        string
    Items     []Item
    Status    Status
    CreatedAt time.Time
}

func (o *Order) AddItem(item Item) error {
    if o.Status != StatusDraft {
        return ErrCannotModify
    }
    o.Items = append(o.Items, item)
    return nil
}
```

**UseCases** (application layer):
```go
// internal/usecase/order/usecase.go
package order

type UseCase struct {
    repo domain.OrderRepository
}

func New(repo domain.OrderRepository) *UseCase {
    return &UseCase{repo: repo}
}

func (uc *UseCase) CreateOrder(ctx context.Context, req CreateRequest) (*domain.Order, error) {
    order := domain.NewOrder()
    for _, item := range req.Items {
        if err := order.AddItem(item); err != nil {
            return nil, fmt.Errorf("add item: %w", err)
        }
    }
    if err := uc.repo.Save(ctx, order); err != nil {
        return nil, fmt.Errorf("save order: %w", err)
    }
    return order, nil
}
```

**Repository Interface** (domain layer):
```go
// internal/domain/order/repository.go
package order

type Repository interface {
    Save(ctx context.Context, order *Order) error
    FindByID(ctx context.Context, id string) (*Order, error)
}
```

**Repository Implementation** (outer circle):
```go
// internal/repository/postgres/order.go
package postgres

type OrderRepository struct {
    db *pgxpool.Pool
}

func (r *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
    // Postgres-specific implementation
}
```

## Domain-Driven Design (DDD)

### Strategic Patterns

**Bounded Contexts**:
- Explicit boundaries around domain models
- Each context has its own ubiquitous language
- Different contexts may model same concept differently
- Integration through explicit interfaces/APIs

**Context Mapping**:
- **Customer/Supplier**: Upstream context provides contract, downstream conforms
- **Conformist**: Downstream adopts upstream's model
- **Anti-Corruption Layer**: Translate between incompatible models
- **Shared Kernel**: Common domain model used by multiple contexts
- **Open Host Service**: Provides generic API via published protocol

**Identifying Bounded Contexts**:
- Natural linguistic boundaries (different business terms)
- Organizational boundaries (different teams/departments)
- Data ownership (who modifies what data)
- Rate of change (fast-changing vs stable domains)
- Scaling needs (different performance requirements)

### Tactical Patterns

**Entities**:
- Identity that persists over time
- Lifecycle matters (creation, modification, deletion)
- Mutable state with invariants
- Example: Order, User, Account

```go
type Order struct {
    id        OrderID
    items     []OrderItem
    status    OrderStatus
    createdAt time.Time
}

func (o *Order) Confirm() error {
    if len(o.items) == 0 {
        return ErrEmptyOrder
    }
    o.status = StatusConfirmed
    return nil
}
```

**Value Objects**:
- No identity, defined by attributes
- Immutable
- Replace when changed
- Example: Money, Address, Email

```go
type Money struct {
    amount   decimal.Decimal
    currency string
}

func NewMoney(amount decimal.Decimal, currency string) (Money, error) {
    if amount.IsNegative() {
        return Money{}, ErrInvalidAmount
    }
    return Money{amount: amount, currency: currency}, nil
}

func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, ErrCurrencyMismatch
    }
    return Money{
        amount:   m.amount.Add(other.amount),
        currency: m.currency,
    }, nil
}
```

**Aggregates**:
- Cluster of entities and value objects
- Treated as single unit for changes
- One root entity (aggregate root)
- External access only through root
- Enforce consistency boundaries

```go
type Order struct {
    id        OrderID        // Aggregate root
    items     []OrderItem    // Part of aggregate
    status    OrderStatus    // Part of aggregate
}

func (o *Order) AddItem(item OrderItem) error {
    // Enforce invariants
    if o.status != StatusDraft {
        return ErrCannotModify
    }
    if len(o.items) >= MaxItems {
        return ErrTooManyItems
    }
    o.items = append(o.items, item)
    return nil
}

// External code only operates on Order, never OrderItem directly
```

**Domain Events**:
- Something that happened in the domain
- Published by aggregates
- Side-effect free (pure events)
- Used for integration and eventual consistency

```go
type OrderConfirmed struct {
    OrderID    string
    OccurredAt time.Time
}

type Order struct {
    events []domain.Event
}

func (o *Order) Confirm() error {
    // Business logic
    o.status = StatusConfirmed
    
    // Publish event
    o.events = append(o.events, OrderConfirmed{
        OrderID:    o.id.String(),
        OccurredAt: time.Now(),
    })
    
    return nil
}

func (o *Order) Events() []domain.Event {
    events := o.events
    o.events = nil
    return events
}
```

**Repositories**:
- Collection-like interface for aggregates
- Only accessed via aggregate root
- Encapsulate persistence details
- Domain defines interface, infrastructure implements

```go
// Domain layer (interface)
type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    FindByID(ctx context.Context, id OrderID) (*Order, error)
    FindByCustomer(ctx context.Context, customerID CustomerID) ([]*Order, error)
}

// Infrastructure layer (implementation)
type PostgresOrderRepository struct {
    db *pgxpool.Pool
}

func (r *PostgresOrderRepository) Save(ctx context.Context, order *Order) error {
    // Persist aggregate
}
```

## Microservices Architecture

### Service Boundaries

**Criteria for Service Boundaries**:
- **Domain cohesion**: Related business capabilities
- **Data ownership**: One service owns specific data
- **Team autonomy**: Independent development and deployment
- **Failure isolation**: One service failure doesn't cascade
- **Scaling needs**: Different performance requirements
- **Technology diversity**: Different tech stacks allowed

**Anti-patterns**:
- Database-driven boundaries (split by tables)
- Microservices for micro-services (too granular)
- Distributed monolith (tightly coupled services)
- Shared database (breaks ownership)

### Communication Patterns

**Synchronous (Request/Response)**:
- REST or gRPC
- HTTP/HTTPS for REST
- Protobuf for gRPC
- Use for: immediate response needed, strong consistency

**Asynchronous (Event-Driven)**:
- Message queues (RabbitMQ, Kafka)
- Event bus
- Use for: eventual consistency, decoupling, high throughput

**Pattern Selection**:
| Scenario | Synchronous | Asynchronous |
|----------|-------------|--------------|
| User needs response | ✅ | ❌ |
| High throughput | ❌ | ✅ |
| Strong consistency | ✅ | ❌ |
| Loose coupling | ❌ | ✅ |
| Simple debugging | ✅ | ❌ |

### Data Consistency

**Eventual Consistency**:
- Each service owns its data
- No distributed transactions (2PC)
- Events propagate changes
- Time-window for consistency

**Saga Pattern**:
- Break transaction into local transactions
- Each step publishes event
- Compensating actions on failure
- Orchestrated or choreographed

**Example: Order Saga**:
```
1. Order Service: Create Order (PENDING)
2. → Publish OrderCreated
3. Payment Service: Process Payment
   → Success: Publish PaymentCompleted
   → Failure: Publish PaymentFailed
4. Inventory Service: Reserve Items
   → Success: Publish InventoryReserved
   → Failure: Publish InventoryFailed
5. Order Service: Update Status
   → All success: CONFIRMED
   → Any failure: CANCELLED (with compensation)
```

## Design Patterns

### Creational Patterns

**Factory Method**:
- Use when: object creation needs flexibility
- Example: Creating different handler types

```go
type HandlerFactory interface {
    CreateHandler(config Config) (Handler, error)
}

type HTTPHandlerFactory struct{}
func (f *HTTPHandlerFactory) CreateHandler(c Config) (Handler, error) {
    return &HTTPHandler{config: c}, nil
}
```

**Builder**:
- Use when: complex object construction
- Example: Building query objects

```go
type QueryBuilder struct {
    query string
    args  []interface{}
}

func (b *QueryBuilder) Select(columns string) *QueryBuilder {
    b.query = fmt.Sprintf("SELECT %s", columns)
    return b
}

func (b *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
    b.query += fmt.Sprintf(" WHERE %s", condition)
    b.args = append(b.args, args...)
    return b
}

func (b *QueryBuilder) Build() (string, []interface{}) {
    return b.query, b.args
}
```

### Structural Patterns

**Adapter**:
- Use when: interface incompatibility
- Example: Integrating external API

```go
type ExternalPaymentClient struct{}

func (c *ExternalPaymentClient) ProcessPayment(amount int) error {
    // External API call
}

type PaymentAdapter struct {
    client *ExternalPaymentClient
}

func (a *PaymentAdapter) Pay(ctx context.Context, amount decimal.Decimal) error {
    // Convert domain type to external API format
    cents := amount.Mul(decimal.NewFromInt(100)).IntPart()
    return a.client.ProcessPayment(int(cents))
}
```

**Decorator**:
- Use when: add behavior without modifying structure
- Example: Adding logging, caching, metrics

```go
type CacheDecorator struct {
    repo   Repository
    cache  Cache
    ttl    time.Duration
}

func (d *CacheDecorator) FindByID(ctx context.Context, id string) (*Entity, error) {
    if cached, ok := d.cache.Get(id); ok {
        return cached, nil
    }
    
    entity, err := d.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    d.cache.Set(id, entity, d.ttl)
    return entity, nil
}
```

### Behavioral Patterns

**Strategy**:
- Use when: interchangeable algorithms
- Example: Different payment methods

```go
type PaymentStrategy interface {
    Process(ctx context.Context, amount decimal.Decimal) error
}

type CreditCardStrategy struct{}
func (s *CreditCardStrategy) Process(ctx context.Context, amount decimal.Decimal) error {
    // Credit card processing
}

type PayPalStrategy struct{}
func (s *PayPalStrategy) Process(ctx context.Context, amount decimal.Decimal) error {
    // PayPal processing
}

type PaymentProcessor struct {
    strategy PaymentStrategy
}

func (p *PaymentProcessor) SetStrategy(s PaymentStrategy) {
    p.strategy = s
}

func (p *PaymentProcessor) ProcessPayment(ctx context.Context, amount decimal.Decimal) error {
    return p.strategy.Process(ctx, amount)
}
```

**Observer**:
- Use when: one-to-many dependency
- Example: Event publishing

```go
type EventPublisher interface {
    Publish(event Event)
}

type EventBus struct {
    subscribers []EventPublisher
}

func (b *EventBus) Subscribe(publisher EventPublisher) {
    b.subscribers = append(b.subscribers, publisher)
}

func (b *EventBus) Publish(event Event) {
    for _, sub := range b.subscribers {
        sub.Publish(event)
    }
}
```

## Architectural Decision Records (ADRs)

### ADR Template

```markdown
# ADR-001: Title

## Status
Accepted | Superseded by ADR-XXX | Deprecated

## Context
What is the issue that we're seeing that is motivating this decision or change?

## Decision
What is the change that we're proposing and/or doing?

## Consequences
What becomes easier or more difficult to do because of this change?

## Alternatives
What other approaches did we consider, and why didn't we choose them?
```

### Example ADR

```markdown
# ADR-001: Use PostgreSQL as Primary Database

## Status
Accepted

## Context
- We need a relational database for transactional data
- Team has PostgreSQL experience
- Requires strong consistency for orders and payments

## Decision
Use PostgreSQL as the primary database with:
- pgx driver for Go
- Connection pooling with pgxpool
- Migrations via goose
- Read replicas for queries

## Consequences
**Positive**:
- Strong ACID guarantees
- Team familiarity
- Rich ecosystem and tooling
- JSON support for flexible fields

**Negative**:
- Scaling writes requires sharding
- Manual indexing required
- Schema migrations needed

## Alternatives Considered
- **MySQL**: Similar, but team prefers PostgreSQL features
- **MongoDB**: Lacks strong consistency guarantees needed
- **CockroachDB**: Less mature, team unfamiliar
```

## Component Design Checklist

- [ ] Clear responsibility (single concern)
- [ ] Stable interface (changes don't break callers)
- [ ] Minimal coupling (few dependencies)
- [ ] High cohesion (related code together)
- [ ] Testable (can unit test in isolation)
- [ ] Documented (clear purpose and usage)
- [ ] Error handling defined (errors propagated or handled)
- [ ] Logging appropriate (not too much, not too little)
- [ ] Performance considered (not premature, but aware)
- [ ] Security reviewed (input validation, authorization)

</instructions>

<constraints>
- Maintain strict layer boundaries (Transport → UseCase → Domain ← Repository)
- Domain layer must have ZERO external dependencies
- Use interfaces at boundaries for decoupling
- Apply dependency inversion: depend on abstractions
- Define bounded contexts around business domains
- Keep aggregates small and focused
- Use value objects for immutability
- Enforce invariants within aggregates
- Prefer composition over inheritance
- Document architectural decisions with ADRs
- Apply YAGNI: build only what's needed now
- Design for testability from the start
- Avoid distributed monolith anti-pattern
- Each microservice should own its data
- Use eventual consistency for cross-service operations
</constraints>

<edge_cases>
If domain is ambiguous or complex: Apply event storming to identify bounded contexts, aggregates, and domain events before designing structure.

If performance requirements conflict with clean architecture: Profile first, optimize hot paths, consider read models or caching in infrastructure layer.

If team lacks DDD experience: Start with simpler layered architecture, introduce DDD patterns incrementally as understanding grows.

If service boundaries are unclear: Use Conway's Law (organization mirrors communication structure) as a guide, align services with team boundaries.

If distributed transactions are needed: Reconsider service boundaries; prefer larger services with shared database over distributed monolith with complex transactions.

If legacy database must be used: Create anti-corruption layer to translate legacy schema to domain model, isolate legacy concerns.

If microservices seem overkill: Start with modular monolith, extract services when clear boundaries emerge and benefits outweigh complexity.

If tight coupling detected: Introduce interfaces at boundaries, refactor to reduce dependencies, consider splitting bounded contexts.

If data consistency becomes problematic: Identify where eventual consistency is acceptable, use sagas for cross-service operations, consider reducing service boundaries.

If performance profiling reveals bottlenecks: Optimize queries first, add caching in infrastructure layer, consider read models for complex queries, avoid premature optimization.
</edge_cases>

<examples>
<example>
<input>Design order processing system architecture</input>
<output>
## Architecture Design: Order Processing System

### Bounded Contexts

**1. Order Context**
- Responsibilities: Order lifecycle, order items, order status
- Domain: Order, OrderItem, OrderStatus (Draft, Confirmed, Cancelled, Completed)
- Aggregates: Order (root)
- Events: OrderCreated, OrderConfirmed, OrderCancelled

**2. Inventory Context**
- Responsibilities: Stock management, product catalog, reservations
- Domain: Product, Stock, Reservation
- Aggregates: Product (root)
- Events: StockReserved, StockReleased

**3. Payment Context**
- Responsibilities: Payment processing, refunds, transactions
- Domain: Payment, Transaction, Refund
- Aggregates: Payment (root)
- Events: PaymentCompleted, PaymentFailed

### Service Boundaries

**Reasoning**:
- Orders are core business, independent scaling needs
- Inventory managed by warehouse team, different performance profile
- Payments require PCI compliance, separate security domain
- Each context has clear data ownership

### Communication Pattern

**Synchronous (User-facing)**:
```
Client → Order Service (REST API)
```

**Asynchronous (Internal)**:
```
Order Service → Message Queue
   → Payment Service (PaymentCompleted)
   → Inventory Service (StockReserved)
   → Order Service (Confirm)
```

### Layered Architecture (Order Service)

```
internal/
├── transport/
│   └── http/
│       ├── handlers.go       # POST /orders, GET /orders/:id
│       ├── dto.go            # CreateOrderRequest, OrderResponse
│       └── middleware.go     # Auth, logging, recovery
├── usecase/
│   ├── order.go              # CreateOrder, ConfirmOrder, CancelOrder
│   └── dto.go                # Internal use case DTOs
├── domain/
│   ├── order.go              # Order entity, business rules
│   ├── events.go             # OrderCreated, OrderConfirmed
│   └── repository.go         # OrderRepository interface
├── repository/
│   └── postgres/
│       ├── repo.go           # Implements OrderRepository
│       ├── models.go         # DB schema (orders, order_items)
│       └── mappers.go        # Domain ↔ DB mapping
└── infrastructure/
    ├── db/                   # Postgres connection pool
    ├── mq/                   # RabbitMQ publisher
    └── cache/                # Redis for read caching
```

### Domain Model

**Order Aggregate**:
```go
// internal/domain/order/order.go
package order

type OrderID string
type OrderStatus string

const (
    StatusDraft     OrderStatus = "draft"
    StatusConfirmed OrderStatus = "confirmed"
    StatusCancelled OrderStatus = "cancelled"
)

type Order struct {
    id        OrderID
    customer  CustomerID
    items     []OrderItem
    status    OrderStatus
    createdAt time.Time
    events    []Event
}

type OrderItem struct {
    productID ProductID
    quantity  int
    price     Money
}

func NewOrder(customerID CustomerID) *Order {
    return &Order{
        id:        generateOrderID(),
        customer:  customerID,
        status:    StatusDraft,
        createdAt: time.Now(),
    }
}

func (o *Order) AddItem(productID ProductID, quantity int, price Money) error {
    if o.status != StatusDraft {
        return ErrCannotModifyConfirmedOrder
    }
    if quantity <= 0 {
        return ErrInvalidQuantity
    }
    o.items = append(o.items, OrderItem{
        productID: productID,
        quantity:  quantity,
        price:     price,
    })
    return nil
}

func (o *Order) Confirm() error {
    if len(o.items) == 0 {
        return ErrEmptyOrder
    }
    o.status = StatusConfirmed
    o.events = append(o.events, OrderConfirmed{
        OrderID:    o.id,
        CustomerID: o.customer,
        Total:      o.Total(),
    })
    return nil
}

func (o *Order) Total() Money {
    var total Money
    for _, item := range o.items {
        total = total.Add(item.price.Mul(decimal.NewFromInt(int64(item.quantity))))
    }
    return total
}

func (o *Order) Events() []Event {
    events := o.events
    o.events = nil
    return events
}
```

### Use Case Layer

**Create Order Workflow**:
```go
// internal/usecase/order/usecase.go
package order

type UseCase struct {
    repo    domain.OrderRepository
    publisher EventPublisher
}

func New(repo domain.OrderRepository, publisher EventPublisher) *UseCase {
    return &UseCase{repo: repo, publisher: publisher}
}

type CreateOrderRequest struct {
    CustomerID string
    Items      []CreateItemRequest
}

type CreateItemRequest struct {
    ProductID string
    Quantity  int
    Price     decimal.Decimal
}

func (uc *UseCase) CreateOrder(ctx context.Context, req CreateOrderRequest) (*domain.Order, error) {
    // Create order
    order := domain.NewOrder(domain.CustomerID(req.CustomerID))
    
    // Add items
    for _, item := range req.Items {
        price, err := domain.NewMoney(item.Price, "USD")
        if err != nil {
            return nil, fmt.Errorf("invalid price: %w", err)
        }
        if err := order.AddItem(domain.ProductID(item.ProductID), item.Quantity, price); err != nil {
            return nil, fmt.Errorf("add item %s: %w", item.ProductID, err)
        }
    }
    
    // Persist
    if err := uc.repo.Save(ctx, order); err != nil {
        return nil, fmt.Errorf("save order: %w", err)
    }
    
    // Publish domain events
    for _, event := range order.Events() {
        if err := uc.publisher.Publish(ctx, event); err != nil {
            log.Printf("publish event: %v", err)
        }
    }
    
    return order, nil
}

func (uc *UseCase) ConfirmOrder(ctx context.Context, orderID string) error {
    order, err := uc.repo.FindByID(ctx, domain.OrderID(orderID))
    if err != nil {
        return fmt.Errorf("find order: %w", err)
    }
    
    if err := order.Confirm(); err != nil {
        return fmt.Errorf("confirm order: %w", err)
    }
    
    if err := uc.repo.Save(ctx, order); err != nil {
        return fmt.Errorf("save order: %w", err)
    }
    
    // Publish events (OrderConfirmed triggers payment and inventory)
    for _, event := range order.Events() {
        if err := uc.publisher.Publish(ctx, event); err != nil {
            log.Printf("publish event: %v", err)
        }
    }
    
    return nil
}
```

### Repository Pattern

**Domain Interface**:
```go
// internal/domain/order/repository.go
package order

type Repository interface {
    Save(ctx context.Context, order *Order) error
    FindByID(ctx context.Context, id OrderID) (*Order, error)
    FindByCustomer(ctx context.Context, customerID CustomerID) ([]*Order, error)
}
```

**Postgres Implementation**:
```go
// internal/repository/postgres/order.go
package postgres

type OrderRepository struct {
    db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
    return &OrderRepository{db: db}
}

func (r *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback(ctx)
    
    // Save order
    const orderQuery = `
        INSERT INTO orders (id, customer_id, status, created_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO UPDATE SET status = $3
    `
    if _, err := tx.Exec(ctx, orderQuery, order.ID(), order.Customer(), order.Status(), order.CreatedAt()); err != nil {
        return fmt.Errorf("insert order: %w", err)
    }
    
    // Save items
    const itemQuery = `
        INSERT INTO order_items (order_id, product_id, quantity, price)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (order_id, product_id) DO UPDATE SET quantity = $3, price = $4
    `
    for _, item := range order.Items() {
        if _, err := tx.Exec(ctx, itemQuery, order.ID(), item.ProductID(), item.Quantity(), item.Price()); err != nil {
            return fmt.Errorf("insert item: %w", err)
        }
    }
    
    return tx.Commit(ctx)
}

func (r *OrderRepository) FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error) {
    const query = `
        SELECT id, customer_id, status, created_at 
        FROM orders WHERE id = $1
    `
    
    row := r.db.QueryRow(ctx, query, id)
    
    var dto OrderDTO
    if err := row.Scan(&dto.ID, &dto.CustomerID, &dto.Status, &dto.CreatedAt); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, domain.ErrOrderNotFound
        }
        return nil, fmt.Errorf("scan order: %w", err)
    }
    
    return MapToDomain(dto)
}
```

### Transport Layer

**HTTP Handler**:
```go
// internal/transport/http/order_handler.go
package http

type OrderHandler struct {
    usecase *order.UseCase
}

type CreateOrderRequest struct {
    CustomerID string `json:"customer_id" validate:"required"`
    Items      []Item `json:"items" validate:"required,dive"`
}

type Item struct {
    ProductID string          `json:"product_id" validate:"required"`
    Quantity  int             `json:"quantity" validate:"required,min=1"`
    Price     decimal.Decimal `json:"price" validate:"required,gt=0"`
}

type OrderResponse struct {
    ID        string    `json:"id"`
    Customer  string    `json:"customer"`
    Status    string    `json:"status"`
    Total     string    `json:"total"`
    CreatedAt time.Time `json:"created_at"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    var req CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }
    
    if err := validator.New().Struct(&req); err != nil {
        http.Error(w, fmt.Sprintf("validation: %v", err), http.StatusBadRequest)
        return
    }
    
    // Convert to use case DTO
    items := make([]order.CreateItemRequest, len(req.Items))
    for i, item := range req.Items {
        items[i] = order.CreateItemRequest{
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
            Price:     item.Price,
        }
    }
    
    ucReq := order.CreateOrderRequest{
        CustomerID: req.CustomerID,
        Items:      items,
    }
    
    // Execute use case
    domainOrder, err := h.usecase.CreateOrder(r.Context(), ucReq)
    if err != nil {
        log.Printf("create order: %v", err)
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    
    // Map to response
    resp := OrderResponse{
        ID:        domainOrder.ID().String(),
        Customer:  domainOrder.Customer().String(),
        Status:    string(domainOrder.Status()),
        Total:     domainOrder.Total().String(),
        CreatedAt: domainOrder.CreatedAt(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(resp)
}
```

### Data Flow Example

**Request: POST /orders**
```
1. HTTP Request → Handler
   Body: {customer_id: "123", items: [...]}

2. Handler validates request
   DTO: CreateOrderRequest

3. Handler maps to UseCase DTO
   DTO: order.CreateOrderRequest

4. UseCase creates domain.Order
   Entity: Order (Draft status)

5. UseCase calls repository.Save()
   Interface: domain.OrderRepository

6. Repository persists to PostgreSQL
   Tables: orders, order_items

7. UseCase publishes OrderCreated event
   Message Queue: RabbitMQ

8. Payment Service receives event
   Processes payment asynchronously

9. Inventory Service receives event
   Reserves stock asynchronously

10. Handler returns HTTP 201
    Response: {id: "456", status: "draft", total: "100.00"}
```

### Key Architectural Decisions

**ADR-001: Separate Order, Inventory, Payment Contexts**
- Context: Different business domains, different teams, different scaling needs
- Decision: Three separate bounded contexts with event-driven communication
- Consequences: Looser coupling, independent deployment, eventual consistency

**ADR-002: Use PostgreSQL per Service**
- Context: Strong consistency required for orders, team experience
- Decision: PostgreSQL with pgx driver, no shared database
- Consequences: Strong ACID guarantees, team familiarity, manual scaling

**ADR-003: Event-Driven Inter-Service Communication**
- Context: Services need to communicate but synchronous coupling is problematic
- Decision: Use message queue (RabbitMQ) for async communication
- Consequences: Loose coupling, eventual consistency, harder debugging

### Monitoring and Observability

**Metrics**:
- Order creation rate (orders/sec)
- Order confirmation latency (p50, p95, p99)
- Error rate by status code
- Message queue lag

**Tracing**:
- Distributed tracing for cross-service requests
- Trace ID propagation through all layers

**Logging**:
- Structured logs with correlation IDs
- Log key events: OrderCreated, OrderConfirmed, PaymentFailed
</output>
</example>

<example>
<input>Refactor monolith to microservices</input>
<output>
## Microservices Migration Strategy

### Current State (Monolith)

**Architecture**:
```
internal/
├── transport/http/
│   ├── order_handler.go
│   ├── payment_handler.go
│   ├── inventory_handler.go
│   └── user_handler.go
├── usecase/
│   ├── order.go
│   ├── payment.go
│   ├── inventory.go
│   └── user.go
├── domain/
│   ├── order/
│   ├── payment/
│   ├── inventory/
│   └── user/
└── repository/
    └── postgres/ (shared database)
```

**Issues**:
- Single database (coupling between services)
- Single deployment (all or nothing)
- Shared domain (mixed concerns)
- Harder to scale individual components
- Team conflicts on shared code

### Migration Strategy: Strangler Fig Pattern

**Approach**:
1. Identify bounded contexts
2. Create new services alongside monolith
3. Route traffic gradually
4. Decommission monolith incrementally

### Step 1: Identify Bounded Contexts

**Analysis**:
- **Order Context**: Order lifecycle, status transitions
- **Payment Context**: Payment processing, transactions
- **Inventory Context**: Stock management, reservations
- **User Context**: Authentication, profile management

**Dependencies**:
```
Order → Payment (synchronous)
Order → Inventory (synchronous)
Payment → Order (callbacks)
Inventory → Order (callbacks)
```

### Step 2: Create Order Service

**New Repository**:
```
services/order-service/
├── internal/
│   ├── transport/http/
│   ├── usecase/
│   ├── domain/
│   └── repository/postgres/
├── api/
│   └── openapi.yaml
├── Dockerfile
└── go.mod
```

**API Definition**:
```yaml
# api/openapi.yaml
openapi: 3.0.0
info:
  title: Order Service
  version: 1.0.0
paths:
  /orders:
    post:
      summary: Create order
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrderRequest'
      responses:
        '201':
          description: Order created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OrderResponse'
  /orders/{id}:
    get:
      summary: Get order
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Order details
components:
  schemas:
    CreateOrderRequest:
      type: object
      required: [customer_id, items]
      properties:
        customer_id:
          type: string
        items:
          type: array
          items:
            $ref: '#/components/schemas/OrderItem'
    OrderItem:
      type: object
      required: [product_id, quantity]
      properties:
        product_id:
          type: string
        quantity:
          type: integer
```

**Database Migration**:
```sql
-- New order-service database
CREATE DATABASE order_service;

-- Copy orders table from monolith
CREATE TABLE orders (
    id VARCHAR(36) PRIMARY KEY,
    customer_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) NOT NULL,
    total DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE order_items (
    order_id VARCHAR(36) NOT NULL REFERENCES orders(id),
    product_id VARCHAR(36) NOT NULL,
    quantity INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    PRIMARY KEY (order_id, product_id)
);

-- Migrate existing data
INSERT INTO order_service.public.orders
SELECT * FROM monolith.public.orders;

INSERT INTO order_service.public.order_items
SELECT * FROM monolith.public.order_items;
```

### Step 3: Implement Anti-Corruption Layer

**Problem**: Monolith uses shared database, direct SQL queries

**Solution**: Introduce ACL in monolith to communicate with new services

```go
// Monolith ACL: Order Service Client
type OrderServiceClient struct {
    httpClient *http.Client
    baseURL    string
}

func NewOrderServiceClient(baseURL string) *OrderServiceClient {
    return &OrderServiceClient{
        httpClient: &http.Client{Timeout: 30 * time.Second},
        baseURL:    baseURL,
    }
}

func (c *OrderServiceClient) CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error) {
    body, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("marshal: %w", err)
    }
    
    httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/orders", bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("do request: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }
    
    var order Order
    if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
        return nil, fmt.Errorf("decode: %w", err)
    }
    
    return &order, nil
}
```

### Step 4: Gradual Traffic Routing

**API Gateway Pattern**:
```go
// API Gateway routes requests
type APIGateway struct {
    orderService      *OrderServiceClient
    monolithHandler   *MonolithHandler
    featureFlagClient FeatureFlagClient
}

func (g *APIGateway) CreateOrder(w http.ResponseWriter, r *http.Request) {
    // Feature flag: 10% traffic to new service
    if g.featureFlagClient.IsEnabled("order-service-migration", 10) {
        g.orderService.CreateOrder(w, r)
        return
    }
    
    // Fallback to monolith
    g.monolithHandler.CreateOrder(w, r)
}
```

**Monitoring**:
- Compare error rates between monolith and service
- Compare latency (p50, p95)
- Monitor feature flag metrics
- Gradually increase traffic to service

### Step 5: Extract Payment Service

**Bounded Context**: Payment processing, PCI compliance

**Architecture**:
```
services/payment-service/
├── internal/
│   ├── domain/
│   │   ├── payment.go          # Payment entity
│   │   └── events.go           # PaymentCompleted, PaymentFailed
│   ├── usecase/
│   │   └── payment.go          # ProcessPayment, RefundPayment
│   ├── repository/
│   │   └── postgres/
│   └── transport/
│       └── grpc/               # gRPC for low-latency
├── api/
│   └── payment.proto
└── Dockerfile
```

**Communication Pattern**:
- **Order Service → Payment Service**: gRPC (synchronous, low latency)
- **Payment Service → Order Service**: Events (async, status updates)

```protobuf
// api/payment.proto
syntax = "proto3";

package payment;

service PaymentService {
  rpc ProcessPayment(ProcessPaymentRequest) returns (ProcessPaymentResponse);
  rpc GetPayment(GetPaymentRequest) returns (GetPaymentResponse);
  rpc RefundPayment(RefundPaymentRequest) returns (RefundPaymentResponse);
}

message ProcessPaymentRequest {
  string order_id = 1;
  string customer_id = 2;
  int64 amount_cents = 3;
  string currency = 4;
  PaymentMethod payment_method = 5;
}

message ProcessPaymentResponse {
  string payment_id = 1;
  PaymentStatus status = 2;
}
```

### Step 6: Event-Driven Communication

**Problem**: Services need to communicate without tight coupling

**Solution**: Message broker for asynchronous events

```go
// Order Service: Publish events
type EventPublisher struct {
    publisher amqp Publisher
}

func (p *EventPublisher) PublishOrderCreated(ctx context.Context, order *Order) error {
    event := OrderCreatedEvent{
        OrderID:    order.ID,
        CustomerID: order.CustomerID,
        Total:      order.Total,
        Items:      order.Items,
    }
    
    body, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("marshal: %w", err)
    }
    
    return p.publisher.Publish(ctx, "orders.created", body)
}
```

**Payment Service: Consume events**:
```go
type EventConsumer struct {
    processor *PaymentProcessor
    consumer  amqp.Consumer
}

func (c *EventConsumer) Start(ctx context.Context) error {
    messages, err := c.consumer.Consume(ctx, "orders.created")
    if err != nil {
        return fmt.Errorf("consume: %w", err)
    }
    
    for msg := range messages {
        var event OrderCreatedEvent
        if err := json.Unmarshal(msg.Body, &event); err != nil {
            log.Printf("unmarshal: %v", err)
            msg.Ack(false)
            continue
        }
        
        // Process payment
        payment, err := c.processor.ProcessPayment(ctx, event)
        if err != nil {
            log.Printf("process payment: %v", err)
            msg.Nack(false, true) // Requeue
            continue
        }
        
        // Publish payment completed
        c.publisher.PublishPaymentCompleted(ctx, payment)
        
        msg.Ack(false)
    }
    
    return nil
}
```

### Step 7: Decommission Monolith

**Phased Decommission**:
1. **Phase 1** (Month 1): Extract Order Service (10% traffic)
2. **Phase 2** (Month 2): Full traffic to Order Service
3. **Phase 3** (Month 2-3): Extract Payment Service
4. **Phase 4** (Month 3-4): Extract Inventory Service
5. **Phase 5** (Month 4-5): Extract User Service
6. **Phase 6** (Month 5-6): Decommission monolith

**Validation Before Decommission**:
- [ ] All traffic routed to new services
- [ ] Error rates acceptable (target: < 0.1%)
- [ ] Latency acceptable (target: < 200ms p95)
- [ ] Data migration complete
- [ ] Monitoring and alerting configured
- [ ] Runbooks documented
- [ ] Team trained on new architecture

### Challenges and Solutions

**Challenge 1: Distributed Transactions**

**Problem**: Order needs to be confirmed only if payment and inventory succeed

**Solution**: Saga Pattern
```go
// Order Service: Saga orchestrator
type OrderSaga struct {
    orderRepo    OrderRepository
    paymentSvc   PaymentServiceClient
    inventorySvc InventoryServiceClient
}

func (s *OrderSaga) CreateOrder(ctx context.Context, req CreateOrderRequest) error {
    // Step 1: Create order (PENDING)
    order, err := s.orderRepo.Create(ctx, req)
    if err != nil {
        return fmt.Errorf("create order: %w", err)
    }
    
    // Step 2: Process payment
    payment, err := s.paymentSvc.ProcessPayment(ctx, order)
    if err != nil {
        // Compensate: Cancel order
        s.orderRepo.Cancel(ctx, order.ID)
        return fmt.Errorf("process payment: %w", err)
    }
    
    // Step 3: Reserve inventory
    if err := s.inventorySvc.ReserveStock(ctx, order); err != nil {
        // Compensate: Refund payment, Cancel order
        s.paymentSvc.RefundPayment(ctx, payment.ID)
        s.orderRepo.Cancel(ctx, order.ID)
        return fmt.Errorf("reserve inventory: %w", err)
    }
    
    // Step 4: Confirm order
    return s.orderRepo.Confirm(ctx, order.ID)
}
```

**Challenge 2: Data Consistency**

**Problem**: Eventually consistent system

**Solution**: 
- Embrace eventual consistency
- Use compensating transactions
- Document SLAs for consistency
- Implement read models for queries

**Challenge 3: Debugging**

**Problem**: Distributed system, harder to trace issues

**Solution**:
- Distributed tracing (OpenTelemetry)
- Correlation IDs across services
- Centralized logging
- Structured logs

```go
// Middleware: Propagate correlation ID
func CorrelationIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        correlationID := r.Header.Get("X-Correlation-ID")
        if correlationID == "" {
            correlationID = uuid.New().String()
        }
        
        ctx := context.WithValue(r.Context(), "correlation_id", correlationID)
        
        w.Header().Set("X-Correlation-ID", correlationID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Use correlation ID in logs
log.WithField("correlation_id", ctx.Value("correlation_id")).Info("Processing order")
```

### Migration Checklist

**Before Starting**:
- [ ] Identify bounded contexts
- [ ] Define service contracts (APIs)
- [ ] Set up monitoring and logging
- [ ] Prepare infrastructure (CI/CD, Kubernetes)
- [ ] Train team on microservices patterns

**During Migration**:
- [ ] Use feature flags for gradual rollout
- [ ] Monitor metrics closely (error rate, latency)
- [ ] Implement circuit breakers for fault tolerance
- [ ] Test rollback procedures
- [ ] Document migration steps

**After Migration**:
- [ ] Decommission monolith
- [ ] Update documentation
- [ ] Archive old code
- [ ] Conduct retrospective
- [ ] Plan next iteration

### Key Takeaways

1. **Start Small**: Extract one bounded context at a time
2. **Use Strangler Fig**: New services alongside monolith, gradual migration
3. **API-First**: Define contracts before implementation
4. **Observability**: Monitor everything from day one
5. **Embrace Async**: Use events for loose coupling
6. **Fail Gracefully**: Circuit breakers, retries, fallbacks
7. **Test Rollback**: Always have a way back
</output>
</example>
</examples>

<output_format>
Provide comprehensive architecture guidance covering:

1. **Architecture Fundamentals**: Separation of concerns, dependency management, data flow
2. **Layered Architecture**: Transport, UseCase, Domain, Repository, Infrastructure layers
3. **Clean Architecture**: Concentric circles, dependency rules, implementation patterns
4. **DDD**: Bounded contexts, aggregates, value objects, domain events
5. **Microservices**: Service boundaries, communication patterns, data consistency
6. **Design Patterns**: Creational, structural, behavioral patterns with examples
7. **ADRs**: Documenting architectural decisions with templates

Combine patterns with clear examples, component designs, and migration strategies.
</output_format>
