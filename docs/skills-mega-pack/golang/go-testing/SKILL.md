---
name: go-testing
description: Table-driven tests, testcontainers, benchmarks, fuzzing, and TDD workflow patterns for Go
---

# Go Testing Patterns

## Table-Driven Tests
```go
func TestParseAmount(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    int64
        wantErr bool
    }{
        {name: "valid cents",     input: "12.34",  want: 1234, wantErr: false},
        {name: "whole dollars",   input: "100",    want: 10000, wantErr: false},
        {name: "negative",        input: "-5.00",  want: -500, wantErr: false},
        {name: "invalid",         input: "abc",    want: 0,    wantErr: true},
        {name: "empty",           input: "",       want: 0,    wantErr: true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            got, err := ParseAmount(tt.input)
            if (err != nil) != tt.wantErr {
                t.Fatalf("ParseAmount(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("ParseAmount(%q) = %d, want %d", tt.input, got, tt.want)
            }
        })
    }
}
```

## Assertions — Use Explicit Comparisons
- Do NOT use testify or assertion libraries
- Use `if got != want { t.Errorf(...) }` for simple values
- Use `cmp.Diff` from `github.com/google/go-cmp` for struct comparison
- Use `t.Fatal` for setup failures, `t.Error` for assertion failures

## Integration Tests with testcontainers
```go
//go:build integration

func TestUserRepo_Integration(t *testing.T) {
    ctx := context.Background()
    pg, _ := postgres.Run(ctx, "postgres:16-alpine",
        postgres.WithDatabase("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready").WithStartupTimeout(30*time.Second),
        ),
    )
    t.Cleanup(func() { pg.Terminate(ctx) })
    // run migrations, create repo, test...
}
```

## TDD Workflow
1. Write a failing test that describes desired behavior
2. Write minimal code to make the test pass
3. Refactor while keeping tests green
4. Repeat — red, green, refactor

## Benchmarks
```go
func BenchmarkProcess(b *testing.B) {
    data := setupBenchData()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        Process(data)
    }
}
// Run: go test -bench=. -benchmem -count=5 | benchstat
```

## Fuzzing
```go
func FuzzParseAmount(f *testing.F) {
    f.Add("12.34")
    f.Add("0")
    f.Add("-99.99")
    f.Fuzz(func(t *testing.T, input string) {
        _, _ = ParseAmount(input) // should not panic
    })
}
```

## Test Helpers
```go
func newTestServer(t *testing.T) *httptest.Server {
    t.Helper()
    // setup ...
    t.Cleanup(func() { server.Close() })
    return server
}
```
