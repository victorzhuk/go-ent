---
name: go-db
description: "PostgreSQL, ClickHouse, Redis integration with pgx, squirrel, goose. Auto-activates for: database work, migrations, queries, repositories."
---

# Go Database (2026)

## Stack

- **PostgreSQL 17** - pgx/v5, squirrel
- **ClickHouse 24** - clickhouse-go/v2
- **Redis 7.4** - go-redis/v9
- **Migrations** - goose/v3

## pgx/v5 Connection

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
    
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("ping: %w", err)
    }
    return pool, nil
}
```

## Repository Pattern

```go
// internal/repository/user/pgx/repo.go
type repository struct {
    pool *pgxpool.Pool
    psql sq.StatementBuilderType
}

func New(pool *pgxpool.Pool) contract.UserRepository {
    return &repository{
        pool: pool,
        psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
    }
}

// internal/repository/user/pgx/schema.go
const (
    tableUsers    = "users"
    colID         = "id"
    colEmail      = "email"
    colCreatedAt  = "created_at"
)

// internal/repository/user/pgx/models.go (private)
type userModel struct {
    ID        string    `db:"id"`
    Email     string    `db:"email"`
    CreatedAt time.Time `db:"created_at"`
}

// internal/repository/user/pgx/mappers.go (private)
func toEntity(m *userModel) *entity.User {
    return &entity.User{
        ID:        uuid.MustParse(m.ID),
        Email:     m.Email,
        CreatedAt: m.CreatedAt,
    }
}

func toModel(e *entity.User) *userModel {
    return &userModel{
        ID:        e.ID.String(),
        Email:     e.Email,
        CreatedAt: e.CreatedAt,
    }
}
```

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

func (r *repository) List(ctx context.Context, f Filter) ([]entity.User, error) {
    qb := r.psql.
        Select(colID, colEmail, colCreatedAt).
        From(tableUsers).
        OrderBy(colCreatedAt + " DESC").
        Limit(uint64(f.Limit)).
        Offset(uint64(f.Offset))
    
    if f.Email != "" {
        qb = qb.Where(sq.ILike{colEmail: "%" + f.Email + "%"})
    }
    
    query, args, _ := qb.ToSql()
    rows, err := r.pool.Query(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("query: %w", err)
    }
    defer rows.Close()
    
    var users []entity.User
    for rows.Next() {
        var m userModel
        if err := rows.Scan(&m.ID, &m.Email, &m.CreatedAt); err != nil {
            return nil, fmt.Errorf("scan: %w", err)
        }
        users = append(users, *toEntity(&m))
    }
    return users, nil
}
```

## Transactions

```go
func (r *repository) SaveWithItems(ctx context.Context, order *entity.Order) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin: %w", err)
    }
    defer tx.Rollback(ctx)
    
    // Insert order
    orderQuery, orderArgs, _ := r.psql.
        Insert("orders").
        Columns("id", "user_id", "total").
        Values(order.ID, order.UserID, order.Total).
        ToSql()
    
    if _, err := tx.Exec(ctx, orderQuery, orderArgs...); err != nil {
        return fmt.Errorf("insert order: %w", err)
    }
    
    // Insert items
    for _, item := range order.Items {
        itemQuery, itemArgs, _ := r.psql.
            Insert("order_items").
            Columns("id", "order_id", "product_id", "qty").
            Values(item.ID, order.ID, item.ProductID, item.Qty).
            ToSql()
        
        if _, err := tx.Exec(ctx, itemQuery, itemArgs...); err != nil {
            return fmt.Errorf("insert item: %w", err)
        }
    }
    
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
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

-- +goose Down
DROP TABLE users;
```

## Redis

```go
func NewRedis(cfg *RedisConfig) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:         cfg.Addr,
        Password:     cfg.Password,
        DB:           cfg.DB,
        PoolSize:     cfg.PoolSize,
        MinIdleConns: cfg.MinIdleConns,
    })
}

// Cache pattern
func (r *cachedRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    key := "user:" + id.String()
    
    // Try cache
    data, err := r.redis.Get(ctx, key).Bytes()
    if err == nil {
        var user entity.User
        json.Unmarshal(data, &user)
        return &user, nil
    }
    
    // Fallback to DB
    user, err := r.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    data, _ = json.Marshal(user)
    r.redis.Set(ctx, key, data, 5*time.Minute)
    
    return user, nil
}
```

## ClickHouse

```go
func NewClickHouse(cfg *ClickHouseConfig) (driver.Conn, error) {
    return clickhouse.Open(&clickhouse.Options{
        Addr: []string{cfg.Addr},
        Auth: clickhouse.Auth{
            Database: cfg.Database,
            Username: cfg.Username,
            Password: cfg.Password,
        },
        Settings: clickhouse.Settings{
            "max_execution_time": 60,
        },
        Compression: &clickhouse.Compression{Method: clickhouse.CompressionLZ4},
        DialTimeout: 5 * time.Second,
    })
}
```
