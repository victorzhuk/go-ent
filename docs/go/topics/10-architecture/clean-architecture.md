# Clean Architecture

Clean Architecture in Go using layers, dependency inversion, and bounded contexts.

## Quick Reference

| Layer      | Responsibility           | Dependencies        |
|------------|--------------------------|---------------------|
| Domain     | Business logic, entities | None                |
| Use Case   | Application logic        | Domain only         |
| Repository | Data access              | Domain (interfaces) |
| Transport  | HTTP/gRPC handlers       | Use Cases           |

## Layer Structure

```
internal/
├── domain/          # Entities, business logic (no dependencies)
│   ├── user/
│   │   ├── user.go
│   │   └── repository.go  # Interface
│   └── order/
├── usecase/         # Application logic
│   └── user/
│       └── service.go
├── repository/      # Data access implementations
│   └── postgres/
│       └── user.go  # Implements domain/user/repository.go
└── transport/       # HTTP/gRPC handlers
    └── http/
        └── user_handler.go
```

## Domain Layer

```go
// internal/domain/user/user.go
package user

type User struct {
    ID    string
    Name  string
    Email string
}

func (u User) Validate() error {
    if u.Name == "" {
        return ErrNameRequired
    }
    return nil
}

// Repository interface (defined in domain)
type Repository interface {
    Create(ctx context.Context, user User) error
    GetByID(ctx context.Context, id string) (*User, error)
}
```

## Use Case Layer

```go
// internal/usecase/user/service.go
package user

import "internal/domain/user"

type Service struct {
    repo user.Repository  // Depends on interface
}

func New(repo user.Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, name, email string) error {
    u := user.User{
        ID:    generateID(),
        Name:  name,
        Email: email,
    }

    if err := u.Validate(); err != nil {
        return err
    }

    return s.repo.Create(ctx, u)
}
```

## Repository Layer

```go
// internal/repository/postgres/user.go
package postgres

import (
    "internal/domain/user"
    "github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
    pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
    return &UserRepository{pool: pool}
}

// Implements user.Repository
func (r *UserRepository) Create(ctx context.Context, u user.User) error {
    query := `INSERT INTO users (id, name, email) VALUES ($1, $2, $3)`
    _, err := r.pool.Exec(ctx, query, u.ID, u.Name, u.Email)
    return err
}
```

## Transport Layer

```go
// internal/transport/http/user_handler.go
package http

type UserHandler struct {
    userSvc *user.Service
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    err := h.userSvc.CreateUser(r.Context(), req.Name, req.Email)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}
```

## Dependency Injection

```go
// cmd/server/main.go
func main() {
    // Infrastructure
    pool := setupDatabase()
   
    // Repositories
    userRepo := postgres.NewUserRepository(pool)
   
    // Use Cases
    userSvc := user.New(userRepo)
   
    // Transport
    handler := http.NewUserHandler(userSvc)
   
    // Server
    mux := http.NewServeMux()
    mux.HandleFunc("POST /users", handler.CreateUser)
   
    http.ListenAndServe(":8080", mux)
}
```

## See Also

- [Project Layout](./project-layout.md)
- [Dependency Injection](./dependency-injection.md)
- [SOLID](./solid.md)
