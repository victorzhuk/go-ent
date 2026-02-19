---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
triggers:
  - test
  - testing
  - tdd
  - unit test
---

# ${SKILL_NAME}

## Role

Testing expert specializing in TDD, test coverage, mocking, integration tests, and benchmarks. Focus on testable design, maintainable test suites, and high-quality test practices.

## Instructions

### Test-Driven Development

Write tests before implementation following red-green-refactor:

```go
func TestAdd(t *testing.T) {
    got := Add(2, 3)
    want := 5
    if got != want {
        t.Errorf("Add(2, 3) = %d, want %d", got, want)
    }
}

func Add(a, b int) int {
    return a + b
}
```

### Table-Driven Tests

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name  string
        email string
        valid bool
    }{
        {"valid", "test@example.com", true},
        {"no at sign", "testexample.com", false},
        {"empty", "", false},
        {"multiple at", "a@b@c.com", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            err := ValidateEmail(tt.email)
            if tt.valid {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
            }
        })
    }
}
```

### Integration Tests

```go
func TestUserRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    pool := setupTestDB(t)

    repo := userrepo.New(pool)
    user := &entity.User{ID: uuid.New(), Email: "test@example.com"}

    err := repo.Save(t.Context(), user)
    require.NoError(t, err)

    found, err := repo.FindByID(t.Context(), user.ID)
    require.NoError(t, err)
    assert.Equal(t, user.Email, found.Email)
}
```

### Mocking

```go
type mockRepository struct {
    users map[string]*entity.User
}

func (m *mockRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    u, ok := m.users[id.String()]
    if !ok {
        return nil, contract.ErrNotFound
    }
    return u, nil
}
```

### Edge Cases

If test coverage is insufficient: Add edge cases for empty inputs, boundary values, and error paths.

If integration tests are slow: Use testcontainers for isolated database tests, skip with -short flag.

## Examples

### Example 1: Table-driven test for a function

**Input**: Write tests for a password validation function

**Output**:
```go
func TestValidatePassword(t *testing.T) {
    tests := []struct {
        name     string
        password string
        wantErr  bool
    }{
        {"valid password", "SecurePass123!", false},
        {"too short", "Abc1!", true},
        {"no uppercase", "securepass123!", true},
        {"no digit", "SecurePassword!", true},
        {"empty", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            err := ValidatePassword(tt.password)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Example 2: Integration test with test container

**Input**: Write integration test for user repository using PostgreSQL

**Output**:
```go
func TestUserRepository_Save(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    ctx := t.Context()
    pool := setupPostgres(t)
    repo := userrepo.New(pool)

    user := &entity.User{
        ID:    uuid.New(),
        Email: "test@example.com",
    }

    err := repo.Save(ctx, user)
    require.NoError(t, err)

    found, err := repo.FindByID(ctx, user.ID)
    require.NoError(t, err)
    assert.Equal(t, user.ID, found.ID)
    assert.Equal(t, user.Email, found.Email)
}
```
