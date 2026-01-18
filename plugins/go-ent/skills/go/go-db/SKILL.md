---
name: go-db
description: "PostgreSQL, ClickHouse, Redis integration with pgx, squirrel, goose. Auto-activates for: database work, migrations, queries, repositories, caching."
version: "2.0.0"
author: "go-ent"
tags: ["go", "database", "pgx", "squirrel", "postgres"]
---

# Go Database

<role>
Expert Go database engineer specializing in PostgreSQL, ClickHouse, and Redis integration. Focus on repository pattern, query optimization, migrations, and connection pooling.
</role>

<instructions>

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

</instructions>

<constraints>
- Include repository pattern with private models and public entities
- Include squirrel for complex queries (joins, dynamic WHERE, pagination)
- Include proper connection pooling with pgxpool
- Include error mapping (pgx.ErrNoRows → domain errors)
- Include transaction support for multi-operation writes
- Include migration management with goose
- Include caching strategy with Redis (cache-aside pattern)
- Include connection lifecycle management (begin, commit, rollback)
- Exclude raw SQL in application code (use squirrel or prepared statements)
- Exclude N+1 query problems (use joins or batch queries)
- Exclude unparameterized queries (use prepared statements to prevent injection)
- Exclude database-specific types leaking into domain layer
- Exclude running migrations in production without proper testing
- Bound to repository layer with entity types from domain
- Follow SQL best practices (indexes, constraints, proper data types)
- Use context for all database operations with proper timeouts
</constraints>

<edge_cases>
If query complexity is high (multiple joins, CTEs needed): Suggest creating a database view or materialized view instead of complex queries in code.

If performance concerns exist: Delegate to go-perf skill for query optimization, indexing strategies, and performance profiling.

If schema changes are required: Emphasize using migrations with goose and ensuring backward compatibility.

If caching strategy is needed: Suggest cache-aside pattern with appropriate TTL and cache invalidation.

If database technology choice is unclear: Recommend PostgreSQL for transactional workloads, ClickHouse for analytics, Redis for caching.

If transaction boundaries are complex: Suggest using transaction manager pattern with context injection.

If connection pool tuning is needed: Suggest adjusting MaxConns/MinConns based on application load patterns.

If code implementation patterns are required: Delegate to go-code skill for Go-specific database interaction patterns.

If architecture decisions are needed: Delegate to go-arch skill for repository pattern integration with clean architecture.
</edge_cases>

<examples>
<example>
<input>Implement repository with squirrel for complex query with joins and filters</input>
<output>
```go
package userrepo

import (
    "context"
    "fmt"

    "github.com/Masterminds/squirrel"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/google/uuid"
)

type repository struct {
    pool *pgxpool.Pool
    psql squirrel.StatementBuilderType
}

func New(pool *pgxpool.Pool) *repository {
    return &repository{
        pool: pool,
        psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
    }
}

type listFilter struct {
    email    *string
    status   *string
    limit    int
    offset   int
}

func (r *repository) List(ctx context.Context, filter listFilter) ([]*entity.User, error) {
    query := r.psql.
        Select(
            "u.id", "u.email", "u.name", "u.status",
            "p.id as profile_id", "p.bio",
        ).
        From("users u").
        Join("profiles p ON p.user_id = u.id").
        Where(sq.Eq{"u.status": "active"}).
        Limit(uint64(filter.limit)).
        Offset(uint64(filter.offset))

    if filter.email != nil {
        query = query.Where(sq.ILike{"u.email": *filter.email + "%"})
    }

    if filter.status != nil {
        query = query.Where(sq.Eq{"u.status": *filter.status})
    }

    sql, args, _ := query.ToSql()

    rows, err := r.pool.Query(ctx, sql, args...)
    if err != nil {
        return nil, fmt.Errorf("query: %w", err)
    }
    defer rows.Close()

    var users []*entity.User
    for rows.Next() {
        var m userModel
        if err := rows.Scan(
            &m.ID, &m.Email, &m.Name, &m.Status,
            &m.ProfileID, &m.Bio,
        ); err != nil {
            return nil, fmt.Errorf("scan: %w", err)
        }
        users = append(users, toEntity(&m))
    }

    return users, nil
}
```

**Benefits**: Dynamic WHERE clauses, type-safe query building, joins across tables.
</output>
</example>

<example>
<input>Implement cached repository with cache-aside pattern</input>
<output>
```go
package userrepo

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type cachedRepository struct {
    repo  *repository
    redis *redis.Client
    ttl   time.Duration
}

func NewCached(repo *repository, redis *redis.Client, ttl time.Duration) *cachedRepository {
    return &cachedRepository{
        repo:  repo,
        redis: redis,
        ttl:   ttl,
    }
}

func (r *cachedRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    key := fmt.Sprintf("user:%s", id)

    // Try cache first
    data, err := r.redis.Get(ctx, key).Bytes()
    if err == nil {
        var user entity.User
        if err := json.Unmarshal(data, &user); err == nil {
            return &user, nil
        }
    }

    // Cache miss - get from DB
    user, err := r.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("db query: %w", err)
    }

    // Populate cache
    data, _ = json.Marshal(user)
    _ = r.redis.Set(ctx, key, data, r.ttl).Err()

    return user, nil
}

func (r *cachedRepository) Save(ctx context.Context, user *entity.User) error {
    // Save to DB first
    if err := r.repo.Save(ctx, user); err != nil {
        return fmt.Errorf("db save: %w", err)
    }

    // Invalidate cache
    key := fmt.Sprintf("user:%s", user.ID)
    _ = r.redis.Del(ctx, key).Err()

    return nil
}
```

**Pattern**: Cache-aside ensures cache is only populated on demand, invalidation happens on writes.
</output>
</example>

<example>
<input>Write goose migration with proper rollback strategy</input>
<output>
```sql
-- +goose Up
BEGIN;

CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    status user_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_created_at ON users(created_at DESC);

CREATE TRIGGER update_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;

-- +goose Down
BEGIN;

DROP TRIGGER IF EXISTS update_updated_at ON users;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_status;

COMMIT;
```

**Best practices**:
- Wrap in transactions
- Drop in reverse order of creation
- Use IF EXISTS for safe rollback
- Include indexes and triggers
- Add updated_at trigger for audit trail
</output>
</example>
</examples>

<output_format>
Provide database implementation guidance with the following structure:

1. **Repository Pattern**: Private models with DB tags, public entities, clear separation
2. **Query Building**: Squirrel for complex queries (joins, dynamic conditions, pagination)
3. **Connection Management**: Proper pool configuration with pgxpool, connection lifecycle
4. **Transactions**: Begin, commit, rollback with proper error handling
5. **Migrations**: Goose with Up/Down sections, idempotent, timestamp prefix
6. **Caching**: Redis integration with cache-aside pattern, TTL, invalidation
7. **Examples**: Complete, runnable repository implementations
8. **Error Handling**: Map database errors (ErrNoRows, constraints) to domain errors

Focus on production-ready database patterns that balance performance with maintainability.
</output_format>
