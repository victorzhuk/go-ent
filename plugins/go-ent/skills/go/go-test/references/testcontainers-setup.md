# Testcontainers Setup for Integration Tests

## Repository Integration Test Suite

```go
type UserRepoSuite struct {
    suite.Suite
    ctx       context.Context
    container *testcontainers.PostgreSQLContainer
    pool      *pgxpool.Pool
    repo      *userRepo.Repository
}

func (s *UserRepoSuite) SetupSuite() {
    s.ctx = context.Background()

    container, err := postgres.Run(s.ctx, "postgres:alpine",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("testuser"),
        postgres.WithPassword("testpass"),
    )
    s.Require().NoError(err)
    s.container = container

    connStr, err := container.ConnectionString(s.ctx, "sslmode=disable")
    s.Require().NoError(err)

    pool, err := pgxpool.New(s.ctx, connStr)
    s.Require().NoError(err)
    s.pool = pool

    // Run migrations
    _, err = s.pool.Exec(s.ctx, `
        CREATE TABLE users (
            id UUID PRIMARY KEY,
            email VARCHAR(255) NOT NULL,
            name VARCHAR(255) NOT NULL,
            created_at TIMESTAMPTZ NOT NULL
        )
    `)
    s.Require().NoError(err)

    s.repo = userRepo.New(s.pool)
}

func (s *UserRepoSuite) TearDownSuite() {
    if s.pool != nil {
        s.pool.Close()
    }
    if s.container != nil {
        _ = testcontainers.TerminateContainer(s.ctx, s.container)
    }
}

func (s *UserRepoSuite) TestSave() {
    user := entity.User{
        ID:        uuid.Must(uuid.NewV7()),
        Email:     "test@example.com",
        Name:      "Test User",
        CreatedAt: time.Now(),
    }

    err := s.repo.Save(s.ctx, &user)
    s.NoError(err)

    // Verify
    found, err := s.repo.FindByID(s.ctx, user.ID)
    s.NoError(err)
    s.Equal(user.ID, found.ID)
    s.Equal(user.Email, found.Email)
    s.Equal(user.Name, found.Name)
}

func TestUserRepoSuite(t *testing.T) {
    suite.Run(t, new(UserRepoSuite))
}
```

**Pattern**: SetupSuite for container, TearDownSuite for cleanup, testify/suite for test organization.
