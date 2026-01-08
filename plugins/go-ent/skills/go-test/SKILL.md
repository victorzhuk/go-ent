---
name: go-test
description: "Testing patterns with testify, testcontainers, table-driven tests. Auto-activates for: writing tests, TDD, coverage, integration tests, mocks."
---

# Go Testing (1.25+)

## Commands

```bash
go test ./... -v                    # All tests
go test -race ./...                 # Race detection
go test -run TestXxx -v ./pkg/...   # Specific test
go test -coverprofile=c.out ./...   # Coverage
go test -bench=. -benchmem ./...    # Benchmarks
```

## Go 1.25+ Testing Features

```go
// testing/synctest (experimental since 1.24)
// GOEXPERIMENT=synctest
import "testing/synctest"

func TestConcurrent(t *testing.T) {
    synctest.Run(func() {
        // Fake clock, deterministic goroutines
        go worker()
        synctest.Wait() // wait for all to block
    })
}
```

## Table-Driven Tests

```go
func TestNewUser(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid", "test@example.com", false},
        {"empty", "", true},
        {"invalid", "not-email", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            user, err := entity.NewUser(tt.email, "Test")
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.email, user.Email)
        })
    }
}
```

## Mocks

```go
type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Save(ctx context.Context, u *entity.User) error {
    return m.Called(ctx, u).Error(0)
}

func TestCreateUser(t *testing.T) {
    repo := new(mockUserRepo)
    repo.On("Save", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
    
    uc := NewCreateUserUC(repo, slog.New(slog.DiscardHandler))
    _, err := uc.Execute(context.Background(), CreateUserReq{Email: "a@b.com"})
    
    require.NoError(t, err)
    repo.AssertExpectations(t)
}
```

## Testcontainers

```go
func (s *UserRepoSuite) SetupSuite() {
    s.ctx = context.Background()
    container, err := postgres.Run(s.ctx, "postgres:17-alpine",
        postgres.WithDatabase("test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    s.Require().NoError(err)
    s.container = container

    connStr, _ := container.ConnectionString(s.ctx, "sslmode=disable")
    s.pool, _ = pgxpool.New(s.ctx, connStr)
    s.repo = userRepo.New(s.pool)
}
```

## Testing by Layer

| Layer      | Test Type   | Tools              |
|------------|-------------|--------------------|
| Domain     | Pure unit   | testify            |
| UseCase    | Mock repos  | testify/mock       |
| Repository | Integration | testcontainers     |
| Transport  | HTTP test   | httptest + mock UC |

## Context7

```
mcp__context7__resolve(library: "testify")
mcp__context7__resolve(library: "testcontainers-go")
```
