---
name: go-test
description: "Go testing with testify, testcontainers, table-driven tests, synctest. Auto-activates for: writing tests, TDD, coverage improvement."
---

# Go Testing (1.25+)

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

## UseCase with Mocks

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

## Integration (Testcontainers)

```go
type UserRepoSuite struct {
    suite.Suite
    ctx       context.Context
    container *postgres.PostgresContainer
    pool      *pgxpool.Pool
    repo      contract.UserRepository
}

func TestUserRepoSuite(t *testing.T) {
    suite.Run(t, new(UserRepoSuite))
}

func (s *UserRepoSuite) SetupSuite() {
    s.ctx = context.Background()
    container, err := postgres.Run(s.ctx, "postgres:17-alpine",
        postgres.WithDatabase("test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(30*time.Second),
        ),
    )
    s.Require().NoError(err)
    s.container = container

    connStr, _ := container.ConnectionString(s.ctx, "sslmode=disable")
    s.pool, _ = pgxpool.New(s.ctx, connStr)
    s.repo = userRepo.New(s.pool)
}

func (s *UserRepoSuite) TearDownSuite() {
    s.pool.Close()
    s.container.Terminate(s.ctx)
}

func (s *UserRepoSuite) TestSaveAndFind() {
    user, _ := entity.NewUser("test@example.com", "John")
    s.Require().NoError(s.repo.Save(s.ctx, user))
    
    found, err := s.repo.FindByID(s.ctx, user.ID)
    s.Require().NoError(err)
    s.Equal(user.Email, found.Email)
}
```

## HTTP Handler Tests

```go
func TestUserHandler_Create(t *testing.T) {
    tests := []struct {
        name   string
        body   string
        setup  func(*mockUC)
        status int
    }{
        {
            name: "success",
            body: `{"email":"a@b.com","name":"John"}`,
            setup: func(m *mockUC) {
                m.On("Execute", mock.Anything, mock.Anything).
                    Return(&CreateUserResp{ID: "123"}, nil)
            },
            status: http.StatusCreated,
        },
        {
            name:   "invalid json",
            body:   "not json",
            setup:  func(m *mockUC) {},
            status: http.StatusBadRequest,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            uc := new(mockUC)
            tt.setup(uc)
            h := NewUserHandler(uc)

            req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(tt.body))
            rec := httptest.NewRecorder()
            h.Create(rec, req)

            assert.Equal(t, tt.status, rec.Code)
        })
    }
}
```

## Commands

```bash
go test ./... -v                    # All
go test -race ./...                 # Race
go test -coverprofile=c.out ./...   # Coverage
go test -run TestXxx -v ./pkg/...   # Specific
go test -bench=. -benchmem ./...    # Bench
```

## Libraries

- testify v1.9+
- testcontainers-go v0.34+
- go-sqlmock v1.5+
