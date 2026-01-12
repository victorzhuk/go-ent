---
name: go-db
description: "PostgreSQL, ClickHouse, Redis integration with pgx, squirrel, goose. Auto-activates for: database work, migrations, queries, repositories, caching."
---

# Go Database

## Stack

- **PostgreSQL** — pgx/v5, squirrel
- **ClickHouse** — clickhouse-go/v2
- **Redis** — go-redis/v9
- **Migrations** — goose/v3

## Connection Pool

```go
func NewPool(ctx context.Context, cfg *DBConfig) (*pgxpool.Pool, error) {
    poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
    if err != nil {
        return nil, fmt.Errorf("parse dsn: %w", err)
    }

    poolCfg.MaxConns = int32(cfg.MaxConns)
    poolCfg.MinConns = int32(cfg.MinConns)
    poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
    poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime

    pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
    if err != nil {
        return nil, fmt.Errorf("connect: %w", err)
    }
    return pool, nil
}
```

## Repository Structure

```
internal/repository/user/pgx/
├── repo.go       # Struct + New()
├── models.go     # PRIVATE with DB tags
├── mappers.go    # PRIVATE toEntity/toModel
├── schema.go     # Constants
└── {op}.go       # One file per operation
```

**Key**: Private models with DB tags, public entities without.

## Queries with Squirrel

```go
func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    query, args, _ := r.psql.
        Select(colID, colEmail, colCreatedAt).
        From(tableUsers).
        Where(sq.Eq{colID: id.String()}).
        ToSql()

    var m userModel
    if err := r.pool.QueryRow(ctx, query, args...).Scan(&m.ID, &m.Email, &m.CreatedAt); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, contract.ErrNotFound
        }
        return nil, fmt.Errorf("query: %w", err)
    }
    return toEntity(&m), nil
}
```

**Pattern**: Squirrel for complex queries, raw SQL for simple ones.

## Transactions

```go
func (r *repository) SaveWithItems(ctx context.Context, order *entity.Order) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin: %w", err)
    }
    defer tx.Rollback(ctx)

    // ... inserts ...

    return tx.Commit(ctx)
}
```

## Migrations (goose)

```sql
-- database/migrations/20260102120000_create_users.sql

-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose Down
DROP TABLE users;
```

**Pattern**: Timestamp prefix, Up/Down sections, idempotent when possible.

## Redis Cache

```go
func (r *cachedRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    key := "user:" + id.String()

    data, err := r.redis.Get(ctx, key).Bytes()
    if err == nil {
        var user entity.User
        json.Unmarshal(data, &user)
        return &user, nil
    }

    user, err := r.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    data, _ = json.Marshal(user)
    r.redis.Set(ctx, key, data, 5*time.Minute)
    return user, nil
}
```

**Pattern**: Cache-aside with TTL.

## Context7

```
mcp__context7__resolve(library: "pgx")
mcp__context7__resolve(library: "squirrel")
mcp__context7__resolve(library: "goose")
mcp__context7__resolve(library: "go-redis")
```
