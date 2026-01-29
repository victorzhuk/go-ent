---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
version: "${VERSION}"
author: "${AUTHOR}"
tags: [${TAGS}]
triggers:
  - pattern: "database|sql|migration|query|repository"
    weight: 0.9
  - keywords: ["database", "sql", "migration", "repository", "query", "postgres", "mysql", "postgresql", "sqlite"]
    weight: 0.8
  - filePattern: "*.sql|*migration*.go|*repo*.go"
    weight: 0.7
---

# ${SKILL_NAME}

<role>
Database expert specializing in SQL, schema design, migrations, and repository patterns. 
Focus on data integrity, performance, maintainability, and database-agnostic patterns.
</role>

<instructions>

## Schema Design

Design schemas with normalization and performance in mind:

```sql
-- Users table with proper constraints
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for common query patterns
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at DESC);
```

## Migration Patterns

Use goose for versioned migrations:

```bash
# Create new migration
goose -dir ./migrations create add_users_table sql

# Migration files
migrations/20240101000001_add_users_table.up.sql
migrations/20240101000001_add_users_table.down.sql
```

Example up migration:

```sql
-- 20240101000001_add_users_table.up.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 20240101000001_add_users_table.down.sql
DROP TABLE users;
```

## Repository Pattern

Implement clean repository interfaces:

```go
package user

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
    ID        string
    Email     string
    CreatedAt time.Time
}

type Repository interface {
    Find(ctx context.Context, id string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    Save(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}

type repository struct {
    db *pgxpool.Pool
}

func New(db *pgxpool.Pool) Repository {
    return &repository{db: db}
}

func (r *repository) Find(ctx context.Context, id string) (*User, error) {
    const query = `
        SELECT id, email, created_at
        FROM users
        WHERE id = $1
    `
    
    var u User
    err := r.db.QueryRow(ctx, query, id).Scan(
        &u.ID,
        &u.Email,
        &u.CreatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf("find user %s: %w", id, err)
    }
    return &u, nil
}
```

## Transaction Handling

Use transactions for multi-step operations:

```go
func (r *repository) CreateWithProfile(ctx context.Context, user *User, profile *Profile) error {
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    if err := r.saveUser(ctx, tx, user); err != nil {
        return err
    }
    
    if err := r.saveProfile(ctx, tx, profile); err != nil {
        return err
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }
    return nil
}
```

## Query Building

Use squirrel for dynamic queries:

```go
import "github.com/Masterminds/squirrel"

func (r *repository) FindMany(ctx context.Context, opts FindOptions) ([]*User, error) {
    sq := squirrel.Select("id", "email", "created_at").
        From("users").
        PlaceholderFormat(squirrel.Dollar)

    if opts.Email != "" {
        sq = sq.Where(squirrel.Eq{"email": opts.Email})
    }
    
    if !opts.StartDate.IsZero() {
        sq = sq.Where(squirrel.GteOrEq{"created_at": opts.StartDate})
    }

    sq = sq.Limit(uint64(opts.Limit)).
        Offset(uint64(opts.Offset)).
        OrderBy("created_at DESC")

    query, args, err := sq.ToSql()
    if err != nil {
        return nil, fmt.Errorf("build query: %w", err)
    }

    rows, err := r.db.Query(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("execute query: %w", err)
    }
    defer rows.Close()

    var users []*User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Email, &u.CreatedAt); err != nil {
            return nil, fmt.Errorf("scan row: %w", err)
        }
        users = append(users, &u)
    }
    return users, nil
}
```

## Indexing Strategy

Create indexes based on query patterns:

```sql
-- For equality queries
CREATE INDEX idx_users_email ON users(email);

-- For range queries
CREATE INDEX idx_users_created_at ON users(created_at DESC);

-- Composite indexes for multi-column queries
CREATE INDEX idx_users_status_created ON users(status, created_at DESC);

-- Partial indexes for filtered queries
CREATE INDEX idx_active_users ON users(email) WHERE status = 'active';

-- Unique constraints for data integrity
ALTER TABLE users ADD CONSTRAINT unique_email UNIQUE (email);
```

</instructions>

<constraints>
- Always use parameterized queries to prevent SQL injection
- Use transactions for multi-step operations with proper rollback
- Design schemas with normalization and referential integrity
- Create indexes based on actual query patterns
- Use connection pooling for production
- Handle context cancellation in all database operations
- Implement proper error handling with context wrapping
- Keep repositories focused on data access only
- Use pgx/v5 for PostgreSQL, or appropriate driver for your database
- Follow DDD principles: repository returns domain entities
- Never execute raw SQL with string concatenation
- Use prepared statements or query builders for dynamic queries
</constraints>

<edge_cases>
If database driver is not specified: Ask which database (PostgreSQL, MySQL, SQLite, etc.) and provide appropriate patterns.

If migration tool is not specified: Recommend goose for PostgreSQL/MySQL, or appropriate tool for the database system.

If performance requirements are complex: Suggest specific indexing strategies, query optimization, and potentially delegate to go-perf skill for advanced profiling.

If distributed transactions are needed: Explain the limitations and suggest alternative patterns like eventual consistency or saga pattern.

If ORM is requested: Discuss trade-offs between raw SQL/queries and ORMs, and recommend sticking to repository pattern with query builders for better control.

If schema changes are frequent: Emphasize the importance of versioned migrations and backward-compatible changes.

If database operations are failing: Check connection pooling, query timeouts, and context cancellation handling.
</edge_cases>

<examples>
<example>
<input>Create a user repository with CRUD operations using pgx</input>
<output>
```go
package user

import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
    ID        string
    Email     string
    Name      string
    CreatedAt time.Time
    UpdatedAt *time.Time
}

type Repository interface {
    Find(ctx context.Context, id string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    Save(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}

type repository struct {
    db *pgxpool.Pool
}

func New(db *pgxpool.Pool) Repository {
    return &repository{db: db}
}

func (r *repository) Find(ctx context.Context, id string) (*User, error) {
    const query = `
        SELECT id, email, name, created_at, updated_at
        FROM users
        WHERE id = $1
    `
    
    var u User
    err := r.db.QueryRow(ctx, query, id).Scan(
        &u.ID,
        &u.Email,
        &u.Name,
        &u.CreatedAt,
        &u.UpdatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf("find user %s: %w", id, err)
    }
    return &u, nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
    const query = `
        SELECT id, email, name, created_at, updated_at
        FROM users
        WHERE email = $1
    `
    
    var u User
    err := r.db.QueryRow(ctx, query, email).Scan(
        &u.ID,
        &u.Email,
        &u.Name,
        &u.CreatedAt,
        &u.UpdatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf("find user by email %s: %w", email, err)
    }
    return &u, nil
}

func (r *repository) Save(ctx context.Context, user *User) error {
    const query = `
        INSERT INTO users (id, email, name, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (email) DO UPDATE
        SET name = EXCLUDED.name, updated_at = NOW()
    `
    
    now := time.Now()
    _, err := r.db.Exec(ctx, query,
        user.ID,
        user.Email,
        user.Name,
        user.CreatedAt,
        now,
    )
    if err != nil {
        return fmt.Errorf("save user %s: %w", user.ID, err)
    }
    return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
    const query = `DELETE FROM users WHERE id = $1`
    
    result, err := r.db.Exec(ctx, query, id)
    if err != nil {
        return fmt.Errorf("delete user %s: %w", id, err)
    }
    
    if result.RowsAffected() == 0 {
        return fmt.Errorf("user not found: %s", id)
    }
    return nil
}
```
</output>
</example>

<example>
<input>Create a migration using goose with up and down migrations</input>
<output>
```sql
-- 20240101000001_add_users_table.up.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status_created ON users(status, created_at DESC);

-- 20240101000001_add_users_table.down.sql
DROP TABLE users;

-- 20240101000002_add_orders_table.up.sql
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    total DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT fk_orders_user_id 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status_created ON orders(status, created_at DESC);

-- 20240101000002_add_orders_table.down.sql
DROP TABLE orders;
```
</output>
</example>

<example>
<input>Create a transaction that creates a user and their initial order</input>
<output>
```go
package order

import (
    "context"
    "fmt"
    
    "github.com/jackc/pgx/v5"
    "github.com/google/uuid"
)

type OrderService struct {
    userRepo  user.Repository
    orderRepo order.Repository
    db        *pgxpool.Pool
}

func (s *OrderService) CreateWithUser(ctx context.Context, req CreateRequest) error {
    tx, err := s.db.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)
    
    // Create user
    userID := uuid.New().String()
    newUser := &User{
        ID:    userID,
        Email: req.Email,
        Name:  req.Name,
    }
    
    if err := s.userRepo.SaveTx(ctx, tx, newUser); err != nil {
        return fmt.Errorf("save user: %w", err)
    }
    
    // Create order
    newOrder := &Order{
        ID:     uuid.New().String(),
        UserID: userID,
        Total:  req.Total,
    }
    
    if err := s.orderRepo.SaveTx(ctx, tx, newOrder); err != nil {
        return fmt.Errorf("save order: %w", err)
    }
    
    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }
    
    return nil
}
```
</output>
</example>
</examples>

<output_format>
Provide production-ready database code following established patterns:

1. **Schema**: Well-designed tables with constraints and indexes
2. **Migrations**: Versioned migrations with up/down SQL using goose
3. **Repository**: Clean interface with pgx implementation
4. **Transactions**: Proper transaction handling with rollback on error
5. **Query Building**: Use squirrel for dynamic queries, parameterized queries otherwise
6. **Error Handling**: Wrapped errors with context (operation: %w)
7. **Context**: Propagated through all database operations
8. **Connection Pooling**: Use pgxpool for production

Focus on data integrity, performance, and maintainability.
</output_format>
