---
name: go-test
description: 'Testing patterns with testify, testcontainers, table-driven tests. Auto-activates for: writing tests, TDD, coverage, integration tests, mocks.'
version: 2.0.0
author: go-ent
license: MIT
compatibility:
    claude_code: '>=1.0'
    opencode: '>=0.1'
tags:
    - go
    - testing
    - tdd
    - testify
    - testcontainers
quality_score: 91
category: go
triggers:
    keywords:
        - test
        - testing
    file_pattern: '**/*_test.go'
    weight: 0.8
---

## Role

Expert Go testing specialist focused on TDD, test patterns, and comprehensive coverage. Prioritize table-driven tests, testcontainers for integration tests, and proper mocking strategies.

## Instructions
## Commands

```bash
go test ./... -v                    # All tests
go test -race ./...                 # Race detection
go test -run TestXxx -v ./pkg/...   # Specific test
go test -coverprofile=c.out ./...   # Coverage
go test -bench=. -benchmem ./...    # Benchmarks
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
    container, err := postgres.Run(s.ctx, "postgres:alpine",
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

## Constraints

- Include table-driven tests with t.Run() and t.Parallel()
- Include testify/assert and testify/require for assertions
- Include proper test setup/teardown with testcontainers for integration tests
- Include mocks using testify/mock for usecase layer tests
- Include race detection with -race flag for concurrent code
- Include coverage reporting for measuring test coverage
- Exclude testing private implementation details directly
- Exclude global mutable state in tests (use test-scoped state)
- Exclude sleeps or time.Sleep in tests (use timeouts or channels)
- Exclude duplicate test logic across multiple files (extract helpers)
- Bound to layer-specific testing: domain (pure unit), usecase (mock repos), repository (testcontainers), transport (httptest)
- Follow TDD principles when adding new features
- Use descriptive test names that explain what is being tested

## Edge Cases

If test requirements are unclear: Ask about the specific layer being tested, what behavior to verify, and what test type (unit/integration) is appropriate.

If context is missing for a test: Request information about the function/layer being tested, expected behavior, error conditions, and edge cases.

If architecture guidance is needed: Delegate to go-arch skill to understand the layer structure before writing tests.

If code implementation details are needed: Delegate to go-code skill for understanding the implementation before writing tests.

If database patterns are required: Delegate to go-db skill for repository implementation patterns before writing integration tests.

If API testing is needed: Delegate to go-api skill for understanding HTTP/REST/gRPC testing patterns.

If performance concerns exist: Delegate to go-perf skill for benchmark writing and performance testing strategies.

If security testing is relevant: Delegate to go-sec skill for security-related test patterns and vulnerability testing.

## Examples
<example>
<input>Write table-driven tests for entity.NewUser() validation</input>
<output>
```go
func TestNewUser(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        name    string
        wantErr error
    }{
        {
            name:    "valid user",
            email:   "test@example.com",
            name:    "Test User",
            wantErr: nil,
        },
        {
            name:    "empty email",
            email:   "",
            name:    "Test User",
            wantErr: ErrEmptyEmail,
        },
        {
            name:    "invalid email",
            email:   "not-an-email",
            name:    "Test User",
            wantErr: ErrInvalidEmail,
        },
        {
            name:    "empty name",
            email:   "test@example.com",
            name:    "",
            wantErr: ErrEmptyName,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            user, err := entity.NewUser(tt.email, tt.name)

            if tt.wantErr != nil {
                require.Error(t, err)
                assert.ErrorIs(t, err, tt.wantErr)
                assert.Nil(t, user)
                return
            }

            require.NoError(t, err)
            assert.NotNil(t, user)
            assert.NotEqual(t, uuid.Nil, user.ID)
            assert.Equal(t, tt.email, user.Email)
            assert.Equal(t, tt.name, user.Name)
            assert.False(t, user.CreatedAt.IsZero())
        })
    }
}
```

**Pattern**: t.Parallel() for isolation, require.NoError for fatal errors, assert for non-fatal checks.
</output>
</example>

<example>
<input>Write integration test for user repository with testcontainers</input>
<output>
See `references/testcontainers-setup.md` for complete testcontainers setup with PostgreSQL, migrations, and repository tests.
</output>
</example>

<example>
<input>Write usecase test with mocked repository</input>
<output>
See `references/usecase-mocking.md` for complete UseCase testing with gomock, setup functions, and type-safe mocks.
</output>
</example>

## Output Format

Provide comprehensive test coverage with the following structure:

1. **Test Structure**: Table-driven tests with t.Run(), t.Parallel(), descriptive test names
2. **Assertions**: testify/require for fatal checks, testify/assert for non-fatal checks
3. **Mocking**: testify/mock or gomock for dependency mocking, with proper setup/teardown
4. **Integration Tests**: testcontainers for real dependencies (PostgreSQL, Redis, etc.)
5. **Test Organization**: Test suites for related tests, helper functions to reduce duplication
6. **Coverage**: Aim for high coverage (>80%) on business-critical code
7. **Examples**: Complete, runnable test files demonstrating patterns

Focus on testing behavior over implementation details, with clear test names that explain what is being tested.

