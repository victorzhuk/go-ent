# Integration Tests

Integration testing with testcontainers for realistic database/service tests.

## Quick Reference

| Tool           | Use When                        |
|----------------|---------------------------------|
| testcontainers | Real database/service in Docker |
| test fixtures  | Test data setup                 |
| `t.Parallel()` | Independent test isolation      |
| Build tags     | Separate unit/integration tests |

## Basic Testcontainer

```go
import (
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func setupPostgres(t *testing.T) *pgxpool.Pool {
    ctx := context.Background()
   
    req := testcontainers.ContainerRequest{
        Image:        "postgres:16",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_PASSWORD": "password",
            "POSTGRES_DB":       "testdb",
        },
        WaitingFor: wait.ForLog("database system is ready"),
    }
   
    container, err := testcontainers.GenericContainer(ctx,
        testcontainers.GenericContainerRequest{
            ContainerRequest: req,
            Started:          true,
        })
    require.NoError(t, err)
   
    t.Cleanup(func() {
        container.Terminate(ctx)
    })
   
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")
   
    connStr := fmt.Sprintf("postgres://postgres:password@%s:%s/testdb",
        host, port.Port())
   
    pool, err := pgxpool.New(ctx, connStr)
    require.NoError(t, err)
   
    return pool
}
```

## Integration Test

```go
//go:build integration

func TestUserRepository_Create(t *testing.T) {
    pool := setupPostgres(t)
    repo := NewUserRepository(pool)
   
    user := User{ID: "1", Name: "Alice", Email: "alice@example.com"}
   
    err := repo.Create(context.Background(), user)
    require.NoError(t, err)
   
    // Verify
    got, err := repo.GetByID(context.Background(), "1")
    require.NoError(t, err)
    assert.Equal(t, user.Name, got.Name)
}

// Run: go test -tags=integration
```

## Test Fixtures

```go
func seedUsers(t *testing.T, pool *pgxpool.Pool) []User {
    users := []User{
        {ID: "1", Name: "Alice"},
        {ID: "2", Name: "Bob"},
    }
   
    for _, u := range users {
        _, err := pool.Exec(context.Background(),
            `INSERT INTO users (id, name) VALUES ($1, $2)`, u.ID, u.Name)
        require.NoError(t, err)
    }
   
    return users
}
```

## See Also

- [Table-Driven Tests](./table-driven.md)
- [Mocking](./mocking.md)
