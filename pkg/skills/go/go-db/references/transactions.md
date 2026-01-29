# Database Transactions with Error Handling

<example>
<input>Implement transaction with multiple operations and rollback on error</input>
<output>
```go
package orderrepo

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/google/uuid"
)

func (r *repository) SaveWithItems(ctx context.Context, order *entity.Order, items []*entity.OrderItem) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    orderID := uuid.New()

    _, err = tx.Exec(ctx, `
        INSERT INTO orders (id, user_id, status, total, created_at)
        VALUES ($1, $2, $3, $4, NOW())
    `, orderID, order.UserID, order.Status, order.Total)
    if err != nil {
        return fmt.Errorf("insert order: %w", err)
    }

    for _, item := range items {
        _, err := tx.Exec(ctx, `
            INSERT INTO order_items (id, order_id, product_id, quantity, price)
            VALUES ($1, $2, $3, $4, $5)
        `, uuid.New(), orderID, item.ProductID, item.Quantity, item.Price)
        if err != nil {
            return fmt.Errorf("insert item %s: %w", item.ProductID, err)
        }
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }

    order.ID = orderID
    return nil
}

func (r *repository) UpdateStatus(ctx context.Context, orderID uuid.UUID, status string) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    result, err := tx.Exec(ctx, `
        UPDATE orders
        SET status = $1, updated_at = NOW()
        WHERE id = $2 AND status != 'cancelled'
    `, status, orderID)
    if err != nil {
        return fmt.Errorf("update order: %w", err)
    }

    if result.RowsAffected() == 0 {
        return ErrOrderNotFound
    }

    if status == "cancelled" {
        _, err := tx.Exec(ctx, `
            UPDATE products p
            SET stock = stock + oi.quantity
            FROM order_items oi
            WHERE oi.order_id = $1 AND oi.product_id = p.id
        `, orderID)
        if err != nil {
            return fmt.Errorf("restore stock: %w", err)
        }
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }

    return nil
}
```

**Pattern**: Defer rollback ensures cleanup on error, explicit commit when successful.
</output>
</example>
