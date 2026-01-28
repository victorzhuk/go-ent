# Rollback Patterns for Database Migrations

## Data Migration with Rollback Strategy

```sql
-- +goose Up
UPDATE users
SET status = 'inactive'
WHERE created_at < NOW() - INTERVAL '1 year';

-- +goose Down
UPDATE users
SET status = 'active'
WHERE created_at < NOW() - INTERVAL '1 year';
```

**Pattern**: Make data migrations idempotent and reversible when possible.

## Non-Rollbackable Operations

```sql
-- +goose Up
DROP TABLE IF EXISTS old_users CASCADE;

-- +goose Down
-- Non-rollbackable - recreate empty table structure
CREATE TABLE IF NOT EXISTS old_users (
    id UUID PRIMARY KEY,
    email VARCHAR(255),
    created_at TIMESTAMPTZ
);
```

**Pattern**: Use CASCADE for drops, acknowledge non-rollbackable operations in comments. Create empty structure in Down for schema consistency.

## Complex Go Migration with Rollback

```go
func upMigrateUserStatus(tx *sql.Tx) error {
    rows, err := tx.Query(`
        SELECT id, created_at
        FROM users
        WHERE status IS NULL
    `)
    if err != nil {
        return fmt.Errorf("query users: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var id string
        var createdAt time.Time
        if err := rows.Scan(&id, &createdAt); err != nil {
            return fmt.Errorf("scan row: %w", err)
        }

        status := "active"
        if createdAt.Before(time.Now().AddDate(-1, 0, 0)) {
            status = "inactive"
        }

        if _, err := tx.Exec(
            "UPDATE users SET status = $1 WHERE id = $2",
            status, id,
        ); err != nil {
            return fmt.Errorf("update user: %w", err)
        }
    }

    return rows.Err()
}

func downMigrateUserStatus(tx *sql.Tx) error {
    _, err := tx.Exec("UPDATE users SET status = NULL WHERE status IN ('active', 'inactive')")
    if err != nil {
        return fmt.Errorf("revert migration: %w", err)
    }
    return nil
}
```

**Pattern**: Store original state or use NULL to indicate "pre-migration" state for reversible Go migrations.
