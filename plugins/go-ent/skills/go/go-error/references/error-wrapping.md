# Error Wrapping Pattern

Example showing repository layer error wrapping with context.

## Example

<example>
<input>Wrap database error with context at repository layer</input>
<output>
```go
package user

import (
    "database/sql"
    "fmt"
)

type Repository struct {
    db *sql.DB
}

func (r *Repository) GetUser(id string) (*User, error) {
    var u User
    err := r.db.QueryRow(
        "SELECT id, name, email FROM users WHERE id = $1",
        id,
    ).Scan(&u.ID, &u.Name, &u.Email)
    
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    
    return &u, nil
}

func (r *Repository) CreateUser(u User) (*User, error) {
    if u.Email == "" {
        return nil, fmt.Errorf("create user: %w", &InvalidInputError{
            Field:   "email",
            Message: "email is required",
        })
    }
    
    err := r.db.QueryRow(
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        u.Name, u.Email,
    ).Scan(&u.ID)
    
    if err != nil {
        return nil, fmt.Errorf("insert user: %w", err)
    }
    
    return &u, nil
}
```
</output>
</example>
