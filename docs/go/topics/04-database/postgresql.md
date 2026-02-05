# PostgreSQL

Production-grade PostgreSQL integration using pgx/v5 (preferred) and squirrel for query building.

## Quick Reference

| Pattern                        | Use When                      |
|--------------------------------|-------------------------------|
| `pgx.Connect(ctx, connString)` | Single connection             |
| `pgxpool.New(ctx, connString)` | Connection pool (recommended) |
| `squirrel.StatementBuilder`    | Dynamic query building        |
| `tx.Exec(ctx, ...)`            | Transaction execution         |
| `rows.Scan(&vars...)`          | Query results                 |

## Connection Setup

### Connection Pool (Recommended)

```go
import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
    URL          string
    MaxConns     int32
    MinConns     int32
    MaxConnLife  time.Duration
    MaxConnIdle  time.Duration
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(cfg.URL)
    if err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }

    // Connection pool settings
    config.MaxConns = cfg.MaxConns           // Default: 4
    config.MinConns = cfg.MinConns           // Default: 0
    config.MaxConnLifetime = cfg.MaxConnLife // Default: 1h
    config.MaxConnIdleTime = cfg.MaxConnIdle // Default: 30m

    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, fmt.Errorf("create pool: %w", err)
    }

    // Verify connection
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("ping: %w", err)
    }

    return pool, nil
}
```

### Connection String

```go
const (
    // Standard format
    connString = "postgres://user:pass@localhost:5432/dbname?sslmode=disable"

    // With connection pool params
    connStringWithPool = "postgres://user:pass@localhost:5432/dbname?" +
        "pool_max_conns=10&" +
        "pool_min_conns=2&" +
        "pool_max_conn_lifetime=1h"
)
```

## Basic Queries

### Query Single Row

```go
type User struct {
    ID    int64
    Name  string
    Email string
}

func (r *Repository) GetUser(ctx context.Context, id int64) (*User, error) {
    query := `SELECT id, name, email FROM users WHERE id = $1`

    var user User
    err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("query user: %w", err)
    }

    return &user, nil
}
```

### Query Multiple Rows

```go
func (r *Repository) ListUsers(ctx context.Context, limit int) ([]User, error) {
    query := `SELECT id, name, email FROM users LIMIT $1`

    rows, err := r.pool.Query(ctx, query, limit)
    if err != nil {
        return nil, fmt.Errorf("query users: %w", err)
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
            return nil, fmt.Errorf("scan user: %w", err)
        }
        users = append(users, user)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("rows error: %w", err)
    }

    return users, nil
}
```

### Using pgx.CollectRows (pgx v5)

```go
import "github.com/jackc/pgx/v5"

func (r *Repository) ListUsers(ctx context.Context, limit int) ([]User, error) {
    query := `SELECT id, name, email FROM users LIMIT $1`

    rows, err := r.pool.Query(ctx, query, limit)
    if err != nil {
        return nil, fmt.Errorf("query: %w", err)
    }
    defer rows.Close()

    users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
    if err != nil {
        return nil, fmt.Errorf("collect rows: %w", err)
    }

    return users, nil
}
```

## Exec (Insert, Update, Delete)

### Insert

```go
func (r *Repository) CreateUser(ctx context.Context, user User) (int64, error) {
    query := `
        INSERT INTO users (name, email, created_at)
        VALUES ($1, $2, $3)
        RETURNING id
    `

    var id int64
    err := r.pool.QueryRow(ctx, query, user.Name, user.Email, time.Now()).Scan(&id)
    if err != nil {
        return 0, fmt.Errorf("insert user: %w", err)
    }

    return id, nil
}
```

### Update

```go
func (r *Repository) UpdateUser(ctx context.Context, id int64, name string) error {
    query := `UPDATE users SET name = $1, updated_at = $2 WHERE id = $3`

    tag, err := r.pool.Exec(ctx, query, name, time.Now(), id)
    if err != nil {
        return fmt.Errorf("update user: %w", err)
    }

    if tag.RowsAffected() == 0 {
        return ErrNotFound
    }

    return nil
}
```

### Delete

```go
func (r *Repository) DeleteUser(ctx context.Context, id int64) error {
    query := `DELETE FROM users WHERE id = $1`

    tag, err := r.pool.Exec(ctx, query, id)
    if err != nil {
        return fmt.Errorf("delete user: %w", err)
    }

    if tag.RowsAffected() == 0 {
        return ErrNotFound
    }

    return nil
}
```

## Transactions

### Basic Transaction

```go
func (r *Repository) CreateUserWithProfile(ctx context.Context, user User, profile Profile) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback(ctx) // Safe to call even after commit

    // Insert user
    var userID int64
    err = tx.QueryRow(ctx,
        `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`,
        user.Name, user.Email,
    ).Scan(&userID)
    if err != nil {
        return fmt.Errorf("insert user: %w", err)
    }

    // Insert profile
    _, err = tx.Exec(ctx,
        `INSERT INTO profiles (user_id, bio) VALUES ($1, $2)`,
        userID, profile.Bio,
    )
    if err != nil {
        return fmt.Errorf("insert profile: %w", err)
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit: %w", err)
    }

    return nil
}
```

### Transaction Helper

```go
func (r *Repository) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin: %w", err)
    }
    defer tx.Rollback(ctx)

    if err := fn(tx); err != nil {
        return err
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit: %w", err)
    }

    return nil
}

// Usage
err := repo.WithTx(ctx, func(tx pgx.Tx) error {
    if err := createUser(tx, user); err != nil {
        return err
    }
    return createProfile(tx, profile)
})
```

## Query Builder with Squirrel

### Setup

```go
import (
    sq "github.com/Masterminds/squirrel"
    "github.com/jackc/pgx/v5"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
```

### Dynamic Filters

```go
func (r *Repository) SearchUsers(ctx context.Context, filters UserFilters) ([]User, error) {
    query := psql.Select("id", "name", "email").From("users")

    if filters.Name != "" {
        query = query.Where(sq.Like{"name": "%" + filters.Name + "%"})
    }

    if filters.Email != "" {
        query = query.Where(sq.Eq{"email": filters.Email})
    }

    if filters.CreatedAfter != nil {
        query = query.Where(sq.GtOrEq{"created_at": filters.CreatedAfter})
    }

    if filters.Limit > 0 {
        query = query.Limit(uint64(filters.Limit))
    }

    sql, args, err := query.ToSql()
    if err != nil {
        return nil, fmt.Errorf("build query: %w", err)
    }

    rows, err := r.pool.Query(ctx, sql, args...)
    if err != nil {
        return nil, fmt.Errorf("query: %w", err)
    }
    defer rows.Close()

    return pgx.CollectRows(rows, pgx.RowToStructByName[User])
}
```

### Insert with Squirrel

```go
func (r *Repository) BulkInsert(ctx context.Context, users []User) error {
    query := psql.Insert("users").Columns("name", "email", "created_at")

    for _, user := range users {
        query = query.Values(user.Name, user.Email, time.Now())
    }

    sql, args, err := query.ToSql()
    if err != nil {
        return fmt.Errorf("build query: %w", err)
    }

    _, err = r.pool.Exec(ctx, sql, args...)
    if err != nil {
        return fmt.Errorf("bulk insert: %w", err)
    }

    return nil
}
```

### Update with Squirrel

```go
func (r *Repository) UpdateUserFields(ctx context.Context, id int64, updates map[string]interface{}) error {
    updates["updated_at"] = time.Now()

    query := psql.Update("users").
        SetMap(updates).
        Where(sq.Eq{"id": id})

    sql, args, err := query.ToSql()
    if err != nil {
        return fmt.Errorf("build query: %w", err)
    }

    tag, err := r.pool.Exec(ctx, sql, args...)
    if err != nil {
        return fmt.Errorf("update: %w", err)
    }

    if tag.RowsAffected() == 0 {
        return ErrNotFound
    }

    return nil
}
```

## Advanced Patterns

### JSONB Handling

```go
type UserMeta struct {
    Preferences map[string]interface{} `json:"preferences"`
    Tags        []string               `json:"tags"`
}

func (r *Repository) SaveUserMeta(ctx context.Context, userID int64, meta UserMeta) error {
    query := `
        UPDATE users
        SET metadata = $1
        WHERE id = $2
    `

    _, err := r.pool.Exec(ctx, query, meta, userID)
    return err
}

func (r *Repository) GetUserMeta(ctx context.Context, userID int64) (*UserMeta, error) {
    query := `SELECT metadata FROM users WHERE id = $1`

    var meta UserMeta
    err := r.pool.QueryRow(ctx, query, userID).Scan(&meta)
    if err != nil {
        return nil, fmt.Errorf("query meta: %w", err)
    }

    return &meta, nil
}
```

### Array Handling

```go
func (r *Repository) GetUsersByIDs(ctx context.Context, ids []int64) ([]User, error) {
    query := `SELECT id, name, email FROM users WHERE id = ANY($1)`

    rows, err := r.pool.Query(ctx, query, ids)
    if err != nil {
        return nil, fmt.Errorf("query: %w", err)
    }
    defer rows.Close()

    return pgx.CollectRows(rows, pgx.RowToStructByName[User])
}
```

### Prepared Statements (rarely needed)

```go
// pgx automatically prepares frequently used queries

func (r *Repository) setupPreparedStatements(ctx context.Context) error {
    _, err := r.pool.Prepare(ctx, "get_user",
        `SELECT id, name, email FROM users WHERE id = $1`)
    return err
}

func (r *Repository) GetUserPrepared(ctx context.Context, id int64) (*User, error) {
    var user User
    err := r.pool.QueryRow(ctx, "get_user", id).Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```

## Connection Pool Tuning

### Pool Configuration

```go
config, _ := pgxpool.ParseConfig(connString)

// Maximum connections
config.MaxConns = int32(runtime.NumCPU() * 2) // Rule of thumb

// Minimum connections (keep-alive)
config.MinConns = 2

// Connection lifetime (prevent long-lived connections)
config.MaxConnLifetime = 1 * time.Hour

// Idle connection timeout
config.MaxConnIdleTime = 30 * time.Minute

// Health check period
config.HealthCheckPeriod = 1 * time.Minute
```

### Monitoring Pool Stats

```go
func (r *Repository) PoolStats() *pgxpool.Stat {
    return r.pool.Stat()
}

// Metrics
stats := repo.PoolStats()
fmt.Printf("Acquired connections: %d\n", stats.AcquiredConns())
fmt.Printf("Idle connections: %d\n", stats.IdleConns())
fmt.Printf("Total connections: %d\n", stats.TotalConns())
fmt.Printf("Max connections: %d\n", stats.MaxConns())
```

## Error Handling

### Common Errors

```go
import (
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) CreateUser(ctx context.Context, user User) error {
    query := `INSERT INTO users (email) VALUES ($1)`

    _, err := r.pool.Exec(ctx, query, user.Email)
    if err != nil {
        // No rows found
        if errors.Is(err, pgx.ErrNoRows) {
            return ErrNotFound
        }

        // Constraint violation
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) {
            // Unique violation
            if pgErr.Code == "23505" {
                return ErrAlreadyExists
            }
            // Foreign key violation
            if pgErr.Code == "23503" {
                return ErrForeignKeyViolation
            }
        }

        return fmt.Errorf("insert user: %w", err)
    }

    return nil
}
```

### PostgreSQL Error Codes

| Code  | Meaning               |
|-------|-----------------------|
| 23505 | Unique violation      |
| 23503 | Foreign key violation |
| 23502 | Not null violation    |
| 23514 | Check violation       |
| 40001 | Serialization failure |

## Common Mistakes

| Mistake                           | Fix                                   |
|-----------------------------------|---------------------------------------|
| Not closing rows                  | Always `defer rows.Close()`           |
| Ignoring `rows.Err()`             | Check after loop completes            |
| Single connection instead of pool | Use `pgxpool` for production          |
| Not using placeholders            | Use `$1, $2` not string concatenation |
| Forgetting `defer tx.Rollback()`  | Always defer rollback                 |
| Not checking `RowsAffected()`     | Verify update/delete affected rows    |

## Best Practices

```go
// ✓ Good - use pool
pool, err := pgxpool.New(ctx, connString)

// ✗ Bad - single connection
conn, err := pgx.Connect(ctx, connString)

// ✓ Good - parameterized query
query := `SELECT * FROM users WHERE id = $1`
pool.Query(ctx, query, userID)

// ✗ Bad - SQL injection risk
query := fmt.Sprintf("SELECT * FROM users WHERE id = %d", userID)

// ✓ Good - check rows affected
tag, _ := pool.Exec(ctx, query, id)
if tag.RowsAffected() == 0 {
    return ErrNotFound
}

// ✗ Bad - assume success
pool.Exec(ctx, query, id)
```

## See Also

- [Migrations](./migrations.md) - Schema migration patterns
- [Redis](./redis.md) - Caching patterns
- [MongoDB](./mongodb.md) - NoSQL alternative
- [pgx documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [squirrel documentation](https://pkg.go.dev/github.com/Masterminds/squirrel)
