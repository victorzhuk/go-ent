---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
triggers:
  - database
  - sql
  - migration
  - repository
  - query
---

# ${SKILL_NAME}

## Role

Database expert specializing in SQL, schema design, migrations, and repository patterns. Focus on data integrity, performance, maintainability, and database-agnostic patterns.

## Instructions

### Schema Design

```sql
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255) UNIQUE NOT NULL,
    name        VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

### Migrations with Goose

```sql
-- +goose Up
CREATE TABLE payments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount     NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    currency   VARCHAR(3) NOT NULL DEFAULT 'USD',
    status     VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE payments;
```

### Repository Pattern with pgx

```go
type repository struct {
    pool *pgxpool.Pool
    psql sq.StatementBuilderType
}

func New(pool *pgxpool.Pool) *repository {
    return &repository{
        pool: pool,
        psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
    }
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    query, args, _ := r.psql.
        Select("id", "email", "name", "created_at").
        From("users").
        Where(sq.Eq{"id": id}).
        ToSql()

    var m userModel
    row := r.pool.QueryRow(ctx, query, args...)
    if err := row.Scan(&m.ID, &m.Email, &m.Name, &m.CreatedAt); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, contract.ErrNotFound
        }
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    return toEntity(&m), nil
}
```

### Transaction Pattern

```go
func (r *repository) Transfer(ctx context.Context, from, to uuid.UUID, amount decimal.Decimal) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer func() { _ = tx.Rollback(ctx) }()

    if err := r.debit(ctx, tx, from, amount); err != nil {
        return fmt.Errorf("debit: %w", err)
    }
    if err := r.credit(ctx, tx, to, amount); err != nil {
        return fmt.Errorf("credit: %w", err)
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit: %w", err)
    }
    return nil
}
```

### Edge Cases

If query performance is poor: Add EXPLAIN ANALYZE output and add appropriate indexes.

If migrations need rollback: Always write both Up and Down migrations, test rollback before deploying.

## Examples

### Example 1: Repository with list and pagination

**Input**: Create a paginated list query for users

**Output**:
```go
func (r *repository) List(ctx context.Context, page, limit int) ([]*entity.User, int, error) {
    offset := (page - 1) * limit

    countQuery, countArgs, _ := r.psql.
        Select("COUNT(*)").
        From("users").
        ToSql()

    var total int
    if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
        return nil, 0, fmt.Errorf("count users: %w", err)
    }

    query, args, _ := r.psql.
        Select("id", "email", "name", "created_at").
        From("users").
        OrderBy("created_at DESC").
        Limit(uint64(limit)).
        Offset(uint64(offset)).
        ToSql()

    rows, err := r.pool.Query(ctx, query, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("list users: %w", err)
    }
    defer rows.Close()

    var users []*entity.User
    for rows.Next() {
        var m userModel
        if err := rows.Scan(&m.ID, &m.Email, &m.Name, &m.CreatedAt); err != nil {
            return nil, 0, fmt.Errorf("scan user: %w", err)
        }
        users = append(users, toEntity(&m))
    }

    return users, total, rows.Err()
}
```

### Example 2: Upsert pattern

**Input**: Implement an upsert for user preferences

**Output**:
```go
func (r *repository) UpsertPreferences(ctx context.Context, userID uuid.UUID, prefs map[string]string) error {
    for key, value := range prefs {
        query, args, _ := r.psql.
            Insert("user_preferences").
            Columns("user_id", "key", "value", "updated_at").
            Values(userID, key, value, time.Now()).
            Suffix("ON CONFLICT (user_id, key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at").
            ToSql()

        if _, err := r.pool.Exec(ctx, query, args...); err != nil {
            return fmt.Errorf("upsert preference %s: %w", key, err)
        }
    }
    return nil
}
```
