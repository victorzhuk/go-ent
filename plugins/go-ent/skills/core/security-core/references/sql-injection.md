# SQL Injection Prevention

Complete example showing how to prevent SQL injection with parameterized queries.

## Example

<example>
<input>Prevent SQL injection with parameterized queries</input>
<output>
```go
package userrepo

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
    pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *repository {
    return &repository{pool: pool}
}

// ❌ VULNERABLE: String concatenation (SQL injection)
func (r *repository) FindByUsernameVulnerable(ctx context.Context, username string) (*User, error) {
    query := fmt.Sprintf("SELECT id, email FROM users WHERE username = '%s'", username)
    // If username = "admin' OR '1'='1", this returns all users!

    var user User
    err := r.pool.QueryRow(ctx, query).Scan(&user.ID, &user.Email)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// ✅ SECURE: Parameterized query
func (r *repository) FindByUsername(ctx context.Context, username string) (*User, error) {
    const query = `SELECT id, username, email FROM users WHERE username = $1`

    var user User
    err := r.pool.QueryRow(ctx, query, username).Scan(
        &user.ID,
        &user.Username,
        &user.Email,
    )

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("query user: %w", err)
    }

    return &user, nil
}

// ✅ SECURE: Batch insert with parameterized query
func (r *repository) CreateBatch(ctx context.Context, users []*User) error {
    const query = `
        INSERT INTO users (id, username, email, created_at)
        VALUES ($1, $2, $3, $4)
    `

    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    for _, user := range users {
        _, err := tx.Exec(ctx, query,
            user.ID,
            user.Username,
            user.Email,
            user.CreatedAt,
        )
        if err != nil {
            return fmt.Errorf("insert user %s: %w", user.ID, err)
        }
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }

    return nil
}

// ✅ SECURE: Dynamic query with allowlist
func (r *repository) FindByField(ctx context.Context, field string, value any) (*User, error) {
    // Allowlist of valid fields to prevent injection
    allowedFields := map[string]bool{
        "username": true,
        "email":    true,
    }

    if !allowedFields[field] {
        return nil, fmt.Errorf("invalid field: %s", field)
    }

    query := fmt.Sprintf("SELECT id, username, email FROM users WHERE %s = $1", field)

    var user User
    err := r.pool.QueryRow(ctx, query, value).Scan(
        &user.ID,
        &user.Username,
        &user.Email,
    )

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("query user: %w", err)
    }

    return &user, nil
}

// ✅ SECURE: Using squirrel for complex queries
func (r *repository) FindWithFilters(ctx context.Context, filters UserFilters) ([]*User, error) {
    query := sq.Select("id", "username", "email").
        From("users").
        PlaceholderFormat(sq.Dollar)

    if filters.Username != "" {
        query = query.Where(sq.Eq{"username": filters.Username})
    }
    if filters.Email != "" {
        query = query.Where(sq.Eq{"email": filters.Email})
    }
    if filters.CreatedAfter != nil {
        query = query.Where(sq.Gt{"created_at": filters.CreatedAfter})
    }

    sql, args, err := query.ToSql()
    if err != nil {
        return nil, fmt.Errorf("build query: %w", err)
    }

    rows, err := r.pool.Query(ctx, sql, args...)
    if err != nil {
        return nil, fmt.Errorf("query users: %w", err)
    }
    defer rows.Close()

    var users []*User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Username, &user.Email); err != nil {
            return nil, fmt.Errorf("scan user: %w", err)
        }
        users = append(users, &user)
    }

    return users, nil
}
```

**Key Security Principles**: Never concatenate user input, use parameterized queries, validate field names with allowlists, use query builders for complex queries.
</output>
</example>
