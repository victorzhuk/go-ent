# Repository Pattern for Go

## Structure

```
internal/repository/user/pgx/
├── repo.go       # Struct + New()
├── models.go     # PRIVATE with DB tags
├── mappers.go    # PRIVATE toEntity/toModel
└── create.go     # Operations (one file each)
```

## Repository Implementation

```go
package userrepo

import (
    "context"
    "fmt"

    "github.com/Masterminds/squirrel"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
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

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    query, args, _ := r.psql.
        Select("id", "email", "created_at").
        From("users").
        Where(squirrel.Eq{"id": id}).
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

**Key points:**
- Use squirrel for complex queries
- Map `pgx.ErrNoRows` to domain error
- Wrap errors with operation context
- Private models, public entities
