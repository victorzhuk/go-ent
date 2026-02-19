## Language Fundamentals
- Use Go 1.22+ features: range-over-func, enhanced routing, loop variable semantics
- Prefer `errors.New` and `fmt.Errorf` with `%w` for wrapping; never use bare `string` errors
- Always handle errors explicitly — never use `_` for error returns unless in tests with justification
- Use `context.Context` as the first parameter for any function that does I/O or may be cancelled
- Prefer value receivers for small structs; pointer receivers for large structs or when mutation is needed

## Project Structure
```
project/
├── cmd/            # Entry points (main packages)
│   └── server/
├── internal/       # Private packages (enforced by Go)
│   ├── domain/     # Business entities and interfaces
│   ├── service/    # Business logic (use cases)
│   ├── handler/    # HTTP/gRPC handlers
│   ├── repo/       # Data access implementations
│   └── config/     # Configuration loading
├── pkg/            # Public reusable packages (use sparingly)
├── api/            # OpenAPI specs, proto files
├── migrations/     # Database migrations
└── go.mod
```

## Error Handling Patterns
```go
// Sentinel errors for expected conditions
var ErrNotFound = errors.New("not found")

// Custom error types for rich context
type ValidationError struct {
    Field   string
    Message string
}
func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s: %s", e.Field, e.Message)
}

// Wrapping for call chain context
if err := repo.Get(ctx, id); err != nil {
    return fmt.Errorf("service.GetUser(%s): %w", id, err)
}

// Checking errors
if errors.Is(err, ErrNotFound) { ... }
var ve *ValidationError
if errors.As(err, &ve) { ... }
```

## Concurrency
- Use `errgroup.Group` for parallel work with error propagation
- Always pass `context.Context` to goroutines for cancellation
- Prefer channels for communication; mutexes for shared state protection
- Use `sync.Once` for lazy initialization, `sync.Pool` for high-alloc paths
- Never start goroutines in init(); prefer explicit lifecycle management

## Naming Conventions
- Packages: short, lowercase, singular (`user`, not `users` or `userService`)
- Interfaces: describe behavior (`Reader`, `Storer`), not implementation
- Unexported by default; only export what's part of the public API
- Avoid stuttering: `user.User` is fine, `user.UserService` is not — use `user.Service`

## Testing
- Table-driven tests with descriptive subtest names
- Use explicit `if got != want` comparisons — avoid assertion libraries
- Test files: `*_test.go` in the same package for white-box, `*_test` package for black-box
- Use `testdata/` directory for fixtures
- Use `t.Helper()` in test helper functions
- Use `t.Parallel()` for independent tests
- Integration tests behind build tags: `//go:build integration`

## Performance
- Profile before optimizing: `pprof`, `trace`, benchmarks
- Pre-allocate slices/maps when size is known: `make([]T, 0, n)`
- Avoid unnecessary allocations in hot paths; use `sync.Pool`
- Use `strings.Builder` for string concatenation
- Benchmark with `testing.B` and compare with `benchstat`

## Dependencies
- Minimize external dependencies — prefer stdlib
- Use `go mod tidy` to clean dependencies
- Pin versions in `go.mod`; review transitive deps
