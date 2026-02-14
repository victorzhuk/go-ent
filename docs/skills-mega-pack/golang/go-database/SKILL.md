---
name: go-database
description: PostgreSQL with pgx, query building with squirrel, migrations with goose, Redis caching, and ClickHouse analytics
---

# Go Database Patterns

## PostgreSQL with pgx
```go
// Connection pool setup
config, _ := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
config.MaxConns = 25
config.MinConns = 5
config.MaxConnLifetime = 30 * time.Minute
config.MaxConnIdleTime = 5 * time.Minute
pool, _ := pgxpool.NewWithConfig(ctx, config)
```

## Repository Pattern
```go
type UserRepo struct { pool *pgxpool.Pool }

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
    var u User
    err := r.pool.QueryRow(ctx,
        `SELECT id, name, email, created_at FROM users WHERE id = $1`, id,
    ).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrNotFound
    }
    return &u, err
}
```

## Query Building with squirrel
```go
query := sq.Select("id", "name", "email").
    From("users").
    Where(sq.Eq{"status": "active"}).
    OrderBy("created_at DESC").
    Limit(20)
sql, args, _ := query.PlaceholderFormat(sq.Dollar).ToSql()
```

## Migrations with goose
- Store migrations in `migrations/` directory
- Use SQL migrations (not Go) for portability
- Name: `YYYYMMDDHHMMSS_description.sql`
- Always include both `-- +goose Up` and `-- +goose Down`
- Run migrations on startup or as a separate step — never auto-migrate in production

## Transactions
```go
tx, err := pool.Begin(ctx)
if err != nil { return err }
defer tx.Rollback(ctx) // safe: no-op after commit

// ... execute statements against tx ...

return tx.Commit(ctx)
```

## Redis Caching
- Use `go-redis/redis/v9` with context support
- Cache-aside pattern: check cache → miss → query DB → set cache
- Use consistent TTLs with jitter to prevent thundering herd
- Serialize to JSON or msgpack; prefer msgpack for performance
- Use pipelines for batch operations

## Best Practices
- Always use parameterized queries — never string concatenation
- Use connection pooling with proper limits
- Set statement timeouts at the database level
- Index columns used in WHERE, JOIN, ORDER BY
- Use EXPLAIN ANALYZE to verify query plans
- Prefer batch operations (pgx.CopyFrom) for bulk inserts
