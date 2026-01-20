# CRUD Operations with Error Handling

<example>
<input>Implement CRUD operations with proper error handling</input>
<output>
```go
package userrepo

import (
    "context"
    "errors"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/google/uuid"
)

var (
    ErrUserNotFound      = errors.New("user not found")
    ErrEmailDuplicate    = errors.New("email already exists")
    ErrConnectionTimeout = errors.New("connection timeout")
)

func (r *repository) Create(ctx context.Context, user *entity.User) error {
    _, err := r.pool.Exec(ctx, `
        INSERT INTO users (id, email, name, status)
        VALUES ($1, $2, $3, $4)
    `, user.ID, user.Email, user.Name, user.Status)
    
    if err != nil {
        if isUniqueViolation(err) {
            return fmt.Errorf("%w: %s", ErrEmailDuplicate, user.Email)
        }
        if ctx.Err() == context.DeadlineExceeded {
            return fmt.Errorf("%w: %w", ErrConnectionTimeout, err)
        }
        return fmt.Errorf("insert user: %w", err)
    }
    return nil
}

func (r *repository) Update(ctx context.Context, id uuid.UUID, user *entity.User) error {
    result, err := r.pool.Exec(ctx, `
        UPDATE users
        SET email = $1, name = $2, status = $3, updated_at = NOW()
        WHERE id = $4
    `, user.Email, user.Name, user.Status, id)
    
    if err != nil {
        if isUniqueViolation(err) {
            return fmt.Errorf("%w: %s", ErrEmailDuplicate, user.Email)
        }
        return fmt.Errorf("update user: %w", err)
    }
    
    if result.RowsAffected() == 0 {
        return fmt.Errorf("%w: %s", ErrUserNotFound, id)
    }
    
    return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
    result, err := r.pool.Exec(ctx, `
        DELETE FROM users
        WHERE id = $1
    `, id)
    
    if err != nil {
        return fmt.Errorf("delete user: %w", err)
    }
    
    if result.RowsAffected() == 0 {
        return fmt.Errorf("%w: %s", ErrUserNotFound, id)
    }
    
    return nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
    var m userModel
    err := r.pool.QueryRow(ctx, `
        SELECT id, email, name, status, created_at
        FROM users
        WHERE email = $1
    `, email).Scan(&m.ID, &m.Email, &m.Name, &m.Status, &m.CreatedAt)
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, fmt.Errorf("%w: %s", ErrUserNotFound, email)
        }
        return nil, fmt.Errorf("query by email: %w", err)
    }
    
    return toEntity(&m), nil
}

func isUniqueViolation(err error) bool {
    var pgErr *pgx.PgError
    return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
```

**Error patterns**: Unique constraint detection, context timeout handling, not found mapping.
</output>
</example>
