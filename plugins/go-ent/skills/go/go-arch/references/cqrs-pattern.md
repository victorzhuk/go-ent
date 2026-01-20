# CQRS Pattern Reference

Complete implementation of CQRS pattern with read/write separation.

## Command Side (Write)

```go
type CreateOrderCmd struct {
    UserID  string
    Items   []OrderItem
    Address string
}

type OrderCommandHandler struct {
    eventStore contract.EventStore
}

func (h *OrderCommandHandler) Handle(ctx context.Context, cmd CreateOrderCmd) error {
    events := []domain.Event{
        domain.OrderCreated{
            ID:      uuid.New(),
            UserID:  cmd.UserID,
            Items:   cmd.Items,
            Address: cmd.Address,
            Time:    time.Now(),
        },
    }
    
    return h.eventStore.Save(ctx, events[0].AggregateID(), events)
}
```

## Query Side (Read)

```go
type OrderQueryModel struct {
    ID        string
    UserID    string
    Status    string
    Total     decimal.Decimal
    CreatedAt time.Time
    UpdatedAt time.Time
}

type OrderQueryHandler struct {
    readDB *pgxpool.Pool
}

func (h *OrderQueryHandler) GetOrdersByUser(ctx context.Context, userID string) ([]OrderQueryModel, error) {
    const q = `SELECT id, user_id, status, total, created_at, updated_at
               FROM order_read_model WHERE user_id = $1 ORDER BY created_at DESC`
    
    rows, err := h.readDB.Query(ctx, q, userID)
    if err != nil {
        return nil, fmt.Errorf("query orders: %w", err)
    }
    defer rows.Close()
    
    var orders []OrderQueryModel
    for rows.Next() {
        var o OrderQueryModel
        if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.Total, &o.CreatedAt, &o.UpdatedAt); err != nil {
            return nil, fmt.Errorf("scan order: %w", err)
        }
        orders = append(orders, o)
    }
    
    return orders, nil
}
```

## Projection (Event → Read Model)

```go
type OrderProjector struct {
    readDB  *pgxpool.Pool
    events  contract.EventConsumer
}

func (p *OrderProjector) Start(ctx context.Context) error {
    return p.events.Subscribe(ctx, "order.*", func(e domain.Event) error {
        switch ev := e.(type) {
        case domain.OrderCreated:
            return p.onOrderCreated(ctx, ev)
        case domain.OrderPaid:
            return p.onOrderPaid(ctx, ev)
        }
        return nil
    })
}

func (p *OrderProjector) onOrderCreated(ctx context.Context, e domain.OrderCreated) error {
    const q = `INSERT INTO order_read_model (id, user_id, status, total, created_at, updated_at)
               VALUES ($1, $2, 'created', $3, $4, $5)`
    
    total := calculateTotal(e.Items)
    _, err := p.readDB.Exec(ctx, q, e.ID, e.UserID, total, e.Time, e.Time)
    return fmt.Errorf("insert projection: %w", err)
}

func (p *OrderProjector) onOrderPaid(ctx context.Context, e domain.OrderPaid) error {
    const q = `UPDATE order_read_model SET status = 'paid', updated_at = $2 WHERE id = $1`
    
    _, err := p.readDB.Exec(ctx, q, e.ID, e.Time)
    return fmt.Errorf("update projection: %w", err)
}
```

## Benefits

- Write side optimized for consistency (event store)
- Read side optimized for queries (denormalized tables)
- Independent scaling (can add more read replicas)
- Event sourcing provides full audit trail
- Eventual consistency handles high write loads
