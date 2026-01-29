# Squirrel Queries with Joins

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
