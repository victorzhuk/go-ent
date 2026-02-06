# Architecture

Clean Architecture patterns for go-ent.

## Overview

This document defines the architectural layers and their boundaries in go-ent.

## Layer Structure

```
Transport → UseCase → Domain ← Repository ← Infrastructure
```

Dependencies flow inward only. Each layer has specific responsibilities.

## Domain Layer

**Purpose:** Core business logic and types with zero external dependencies.

**Rules:**
- No external dependencies (stdlib only)
- Pure business logic
- No struct tags (no JSON, YAML, etc.)
- Defines interfaces for capabilities, not implementation

**Location:** `internal/domain/`

**Example:**
```go
package domain

type User struct {
    ID   string
    Name string
}

func NewUser(name string) (*User, error) {
    if name == "" {
        return nil, ErrEmptyName
    }
    return &User{
        ID:   generateID(),
        Name: name,
    }, nil
}
```

## Repository Layer

**Purpose:** Data access and storage operations.

**Rules:**
- Accept interfaces from domain/usecase
- Return structs
- Wrap errors with query context
- Interface defined at consumer side

**Location:** `internal/{package}/repository/`

**Example:**
```go
package repository

type UserRepository interface {
    FindByID(ctx context.Context, id string) (*domain.User, error)
    Save(ctx context.Context, user *domain.User) error
}

type userRepository struct {
    pool *pgxpool.Pool
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
    query := "SELECT id, name FROM users WHERE id = $1"
    var u userModel
    if err := r.pool.QueryRow(ctx, query, id).Scan(&u.ID, &u.Name); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, fmt.Errorf("user %s: %w", id, ErrNotFound)
        }
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    return toDomain(&u), nil
}
```

## UseCase Layer

**Purpose:** Business logic and orchestration.

**Rules:**
- Accept repository interfaces
- Return domain structs or response DTOs
- Handle transactions
- Coordinate multiple repositories

**Location:** `internal/{package}/usecase/`

**Example:**
```go
package usecase

type CreateUserUseCase interface {
    Execute(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error)
}

type createUserUseCase struct {
    repo UserRepository
    log  *slog.Logger
}

func (uc *createUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
    user, err := domain.NewUser(req.Name)
    if err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }

    if err := uc.repo.Save(ctx, user); err != nil {
        return nil, fmt.Errorf("save user: %w", err)
    }

    return &CreateUserResponse{ID: user.ID}, nil
}
```

## Transport Layer

**Purpose:** External interfaces (HTTP, gRPC, CLI, MCP).

**Rules:**
- Accept usecase interfaces
- Private DTOs with validation tags
- Map to/from domain types
- Handle HTTP status codes
- Zero business logic

**Location:** `cmd/`, `internal/mcp/`, `internal/cli/`

**Example:**
```go
package transport

type CreateUserRequest struct {
    Name string `json:"name" validate:"required,min=1"`
}

type UserHandler struct {
    uc CreateUserUseCase
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    resp, err := h.uc.Execute(r.Context(), req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(resp)
}
```

## Interface Definition Rule

**Rule:** Define interfaces at the consumer side, not the implementation side.

```go
// GOOD - Interface defined where it's consumed
package usecase

type Store interface {  // usecase defines the contract
    Save(ctx context.Context, task *domain.Task) error
    FindByID(ctx context.Context, id string) (*domain.Task, error)
}

type TaskUseCase struct {
    store Store  // accepts interface
}

// Implementation defines it differently
package repository

type boltStore struct {  // private struct
    db *bolt.DB
}

func (s *boltStore) Save(ctx context.Context, task *domain.Task) error { /* ... */ }
func (s *boltStore) FindByID(ctx context.Context, id string) (*domain.Task, error) { /* ... */ }
```

## Dependency Injection Pattern

Use constructor injection with struct composition:

```go
// Application wiring
func NewApplication(cfg *Config, log *slog.Logger) (*Application, error) {
    db, err := bolt.Open(cfg.DBPath, cfg.DBMode, nil)
    if err != nil {
        return nil, fmt.Errorf("open db: %w", err)
    }

    repo := repository.NewBoltStore(db)
    uc := usecase.NewTaskUseCase(repo, log)
    handler := transport.NewTaskHandler(uc)

    return &Application{handler: handler}, nil
}
```

## Layer Boundaries

### Domain
- **Can use:** stdlib only
- **Cannot use:** external packages, struct tags, persistence code

### Repository
- **Can use:** stdlib, external DB drivers, domain types
- **Cannot use:** HTTP handlers, business logic

### UseCase
- **Can use:** stdlib, domain types, repository interfaces
- **Cannot use:** HTTP handlers, direct DB access, transport types

### Transport
- **Can use:** stdlib, usecase interfaces, external HTTP libs
- **Cannot use:** direct repository access, business logic, domain mutations
