# Table-Driven Tests

Table-driven testing is the idiomatic Go approach for testing multiple inputs/outputs. It reduces duplication and makes tests easy to extend.

## Quick Reference

| Pattern                           | Use When                        |
|-----------------------------------|---------------------------------|
| `t.Run(name, func(t *testing.T))` | Subtests with descriptive names |
| `t.Parallel()`                    | Tests can run concurrently      |
| `testify/assert`                  | Readable assertions             |
| `testify/require`                 | Stop test on first failure      |
| `t.Cleanup(func())`               | Resource cleanup                |

## Basic Table-Driven Test

### Simple Example

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a    int
        b    int
        want int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -2, -3, -5},
        {"mixed signs", -2, 3, 1},
        {"zeros", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Add(%d, %d) = %d; want %d",
                    tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

**Key points:**
- Use `struct` for test cases
- `name` field describes the scenario
- `t.Run` creates subtests
- Loop variable shadowing not needed in Go 1.22+

## With testify/assert

### Assert vs Require

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestProcess(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "hello", "HELLO", false},
        {"empty input", "", "", true},
        {"special chars", "hello!", "HELLO!", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Process(tt.input)

            if tt.wantErr {
                require.Error(t, err) // Stop if error expected but not returned
                return
            }

            require.NoError(t, err)     // Stop if unexpected error
            assert.Equal(t, tt.want, got) // Continue even if fails
        })
    }
}
```

**Difference:**
- `assert.*`: Test continues after failure (use for multiple checks)
- `require.*`: Test stops after failure (use for prerequisites)

## Parallel Tests

### Basic Parallel

```go
func TestParallel(t *testing.T) {
    tests := []struct {
        name string
        id   int
    }{
        {"user 1", 1},
        {"user 2", 2},
        {"user 3", 3},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // Run this subtest in parallel

            user := fetchUser(tt.id) // Slow operation
            assert.NotNil(t, user)
        })
    }
}
```

**Key points:**
- Call `t.Parallel()` at start of subtest
- Reduces test execution time
- Safe when tests don't share state

### When Not to Parallelize

```go
func TestDatabase(t *testing.T) {
    // DON'T use t.Parallel() when:
    // - Tests modify shared database
    // - Tests use shared file system
    // - Tests modify global state

    t.Run("create user", func(t *testing.T) {
        // Not parallel - modifies DB
        db.CreateUser(user)
    })

    t.Run("delete user", func(t *testing.T) {
        // Not parallel - depends on previous test
        db.DeleteUser(user.ID)
    })
}
```

## Test Helpers

### Setup and Teardown

```go
func setupTest(t *testing.T) *Database {
    db := NewDatabase()
    if err := db.Connect(); err != nil {
        t.Fatal(err)
    }

    t.Cleanup(func() {
        db.Close()
    })

    return db
}

func TestUserOperations(t *testing.T) {
    db := setupTest(t) // Cleanup runs automatically

    t.Run("create user", func(t *testing.T) {
        err := db.CreateUser(user)
        require.NoError(t, err)
    })
}
```

### Helper Functions

```go
func assertUser(t *testing.T, got, want User) {
    t.Helper() // Marks this function as helper (improves error messages)

    assert.Equal(t, want.ID, got.ID)
    assert.Equal(t, want.Name, got.Name)
    assert.Equal(t, want.Email, got.Email)
}

func TestGetUser(t *testing.T) {
    got := getUser(123)
    want := User{ID: 123, Name: "Alice", Email: "alice@example.com"}

    assertUser(t, got, want) // Error will point to this line, not inside helper
}
```

## Testing Errors

### Error Checking

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   User
        wantErr error
    }{
        {
            name:    "missing name",
            input:   User{Email: "a@b.com"},
            wantErr: ErrNameRequired,
        },
        {
            name:    "invalid email",
            input:   User{Name: "Alice", Email: "invalid"},
            wantErr: ErrInvalidEmail,
        },
        {
            name:  "valid user",
            input: User{Name: "Alice", Email: "a@b.com"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.input.Validate()

            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Custom Error Assertions

```go
func TestCustomError(t *testing.T) {
    err := validateAge(-1)

    var validationErr *ValidationError
    require.ErrorAs(t, err, &validationErr)

    assert.Equal(t, "age", validationErr.Field)
    assert.Contains(t, validationErr.Message, "negative")
}
```

## Testing with Context

### Context-Aware Tests

```go
func TestWithContext(t *testing.T) {
    tests := []struct {
        name    string
        timeout time.Duration
        wantErr bool
    }{
        {"fast operation", 5 * time.Second, false},
        {"timeout", 1 * time.Millisecond, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
            defer cancel()

            err := doWork(ctx)

            if tt.wantErr {
                require.Error(t, err)
                assert.ErrorIs(t, err, context.DeadlineExceeded)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Using t.Context() (Go 1.24+)

```go
func TestWithTContext(t *testing.T) {
    ctx := t.Context() // Cancelled when test completes

    result := doWork(ctx)
    assert.NotNil(t, result)
}
```

## Golden Files

### Testing Against Golden Files

```go
import (
    "os"
    "path/filepath"
    "testing"
)

var update = flag.Bool("update", false, "update golden files")

func TestRender(t *testing.T) {
    tests := []struct {
        name  string
        input Template
    }{
        {"simple", Template{Name: "test"}},
        {"complex", Template{Name: "test", Items: []string{"a", "b"}}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Render(tt.input)

            golden := filepath.Join("testdata", tt.name+".golden")

            if *update {
                os.WriteFile(golden, []byte(got), 0644)
                return
            }

            want, err := os.ReadFile(golden)
            require.NoError(t, err)

            assert.Equal(t, string(want), got)
        })
    }
}

// Run: go test
// Update: go test -update
```

## Table Tests with Functions

### Testing Multiple Operations

```go
func TestCalculator(t *testing.T) {
    tests := []struct {
        name string
        op   func(int, int) int
        a    int
        b    int
        want int
    }{
        {"add", Add, 2, 3, 5},
        {"subtract", Sub, 5, 3, 2},
        {"multiply", Mul, 2, 3, 6},
        {"divide", Div, 6, 3, 2},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.op(tt.a, tt.b)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Advanced Patterns

### Subtests with Setup

```go
func TestUserService(t *testing.T) {
    db := setupTest(t)

    tests := []struct {
        name string
        test func(t *testing.T, db *Database)
    }{
        {
            name: "create user",
            test: func(t *testing.T, db *Database) {
                err := db.CreateUser(user)
                require.NoError(t, err)
            },
        },
        {
            name: "get user",
            test: func(t *testing.T, db *Database) {
                user, err := db.GetUser(123)
                require.NoError(t, err)
                assert.NotNil(t, user)
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.test(t, db)
        })
    }
}
```

### Organizing Large Test Tables

```go
var userValidationTests = []struct {
    name    string
    user    User
    wantErr error
}{
    {"valid user", validUser, nil},
    {"missing name", userWithoutName, ErrNameRequired},
    // ... many more cases
}

var emailValidationTests = []struct {
    name    string
    email   string
    wantErr bool
}{
    {"valid email", "test@example.com", false},
    {"invalid email", "invalid", true},
    // ... many more cases
}

func TestUserValidation(t *testing.T) {
    for _, tt := range userValidationTests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.user.Validate()
            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

## Common Mistakes

| Mistake                            | Fix                                                  |
|------------------------------------|------------------------------------------------------|
| Not using `t.Run`                  | Always use subtests for table tests                  |
| Generic test names                 | Use descriptive names: "empty input" not "test1"     |
| Shadowing loop variable (Go <1.22) | Pass as parameter or use `:=` in loop                |
| `t.Parallel()` with shared state   | Don't parallelize tests that modify shared resources |
| Not calling `t.Helper()`           | Mark helper functions with `t.Helper()`              |
| Using `assert` for prerequisites   | Use `require` to stop test early                     |

## Best Practices

```go
// ✓ Good - descriptive test case names
tests := []struct {
    name string
    // ...
}{
    {"empty input returns error", ...},
    {"valid input succeeds", ...},
    {"nil pointer returns error", ...},
}

// ✗ Bad - generic names
tests := []struct {
    name string
    // ...
}{
    {"test1", ...},
    {"case2", ...},
}

// ✓ Good - clear structure
tests := []struct {
    name    string
    input   int
    want    int
    wantErr bool
}{
    // ...
}

// ✗ Bad - unclear fields
tests := []struct {
    n string
    i int
    w int
    e bool
}{
    // ...
}
```

## See Also

- [Mocking](./mocking.md) - Testing with mocks
- [Integration Tests](./integration.md) - testcontainers patterns
- [Benchmarks](./benchmarks.md) - Performance testing
- [testing package](https://pkg.go.dev/testing)
- [testify](https://github.com/stretchr/testify)
