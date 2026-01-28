---
name: go-db
description: "PostgreSQL, ClickHouse, Redis integration with pgx, squirrel, goose. Auto-activates for: database work, migrations, queries, repositories, caching."
version: "2.0.0"
author: "go-ent"
license: "MIT"
compatibility:
  claude_code: ">=1.0"
  opencode: ">=0.1"
tags: ["go", "database", "pgx", "squirrel", "postgres"]
quality_score: 89
category: "go"
depends_on: [go-code]
---

<triggers>
keywords:
  - "database"
  - "sql"
  - "postgres"
  - "repository"
  - "migration"
file_pattern: "**/*_repo.go"
weight: 0.8
</triggers>

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
<input>Implement user repository with CRUD operations</input>
<output>
```go
type Repository struct {
    pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
    return &Repository{pool: pool}
}

// Save inserts or updates user
func (r *Repository) Save(ctx context.Context, user *entity.User) error {
    query := `
        INSERT INTO users (id, email, name, created_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO UPDATE
        SET email = EXCLUDED.email, name = EXCLUDED.name
    `
    
    _, err := r.pool.Exec(ctx, query, user.ID, user.Email, user.Name, user.CreatedAt)
    if err != nil {
        return fmt.Errorf("save user: %w", err)
    }
    
    return nil
}

// FindByID retrieves user by ID
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    query := `SELECT id, email, name, created_at FROM users WHERE id = $1`
    
    var user entity.User
    err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, contract.ErrNotFound
        }
        return nil, fmt.Errorf("find user: %w", err)
    }
    
    return &user, nil
}
```

**Pattern**: Parameterized queries ($1, $2), error wrapping with context, domain error mapping (ErrNoRows → ErrNotFound).
</output>
</example>

For additional database implementation examples, see:
- `references/squirrel-queries.md` - Complex queries with joins
- `references/caching-patterns.md` - Redis cache-aside pattern
- `references/goose-migrations.md` - Migration with rollback strategy
- `references/transactions.md` - Transaction handling
- `references/crud-operations.md` - Complete CRUD patterns

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
