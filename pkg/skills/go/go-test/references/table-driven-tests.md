# Table-Driven Tests Quick Reference

Extracted from `docs/go/topics/08-testing/table-driven.md` (436 lines) → 100 lines of actionable patterns.

## Quick Reference Table

| Pattern                           | Use When                        |
|-----------------------------------|---------------------------------|
| `t.Run(name, func(t *testing.T))` | Subtests with descriptive names |
| `t.Parallel()`                    | Tests can run concurrently      |
| `testify/assert`                  | Readable assertions             |
| `testify/require`                 | Stop test on first failure      |
| `t.Cleanup(func())`               | Resource cleanup                |

## Basic Table-Driven Test

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
                t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

## With testify/assert

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
                require.Error(t, err)  // Stop if error expected but not returned
                return
            }

            require.NoError(t, err)       // Stop if unexpected error
            assert.Equal(t, tt.want, got) // Continue even if fails
        })
    }
}

// assert.*: Test continues after failure (use for multiple checks)
// require.*: Test stops after failure (use for prerequisites)
```

## Parallel Tests

```go
func TestParallel(t *testing.T) {
    tests := []struct {
        name string
        input int
        want int
    }{
        {"case 1", 1, 2},
        {"case 2", 2, 4},
        {"case 3", 3, 6},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()  // Run subtests concurrently

            got := Double(tt.input)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Context-Based Tests

```go
func TestWithContext(t *testing.T) {
    tests := []struct {
        name    string
        timeout time.Duration
        wantErr bool
    }{
        {"normal operation", 5 * time.Second, false},
        {"timeout", 1 * time.Millisecond, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := t.Context()  // Automatically cancelled when test ends
            ctx, cancel := context.WithTimeout(ctx, tt.timeout)
            defer cancel()

            err := SlowOperation(ctx)
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```
