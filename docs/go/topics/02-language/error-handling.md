# Error Handling

Go's error handling is explicit and idiomatic. Modern error handling uses wrapping, sentinel errors, and the errors package for inspection.

## Quick Reference

| Pattern                          | Use When                                 |
|----------------------------------|------------------------------------------|
| `if err != nil { return err }`   | Propagate error up                       |
| `fmt.Errorf("context: %w", err)` | Add context while preserving error chain |
| `errors.Is(err, target)`         | Check for specific error in chain        |
| `errors.As(err, &target)`        | Extract specific error type from chain   |
| Custom error types               | Need structured error information        |
| Sentinel errors                  | Pre-defined errors (e.g., `io.EOF`)      |

## Basic Error Handling

### Error Propagation

```go
func readFile(path string) ([]byte, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read file %s: %w", path, err)
    }
    return data, nil
}
```

**Key points:**
- Always check `err != nil`
- Use `%w` to wrap errors (preserves error chain)
- Add context (file path, operation) before error
- Format: `"operation details: %w"` (lowercase, no punctuation)

### Happy Path Left

```go
// Good - happy path on left margin
func process(id string) error {
    user, err := db.GetUser(id)
    if err != nil {
        return fmt.Errorf("get user: %w", err)
    }

    if err := user.Validate(); err != nil {
        return fmt.Errorf("validate user: %w", err)
    }

    return user.Save()
}

// Bad - nested indentation
func badProcess(id string) error {
    user, err := db.GetUser(id)
    if err == nil {
        if user.Validate() == nil {
            return user.Save()
        }
    }
    return err
}
```

## Error Wrapping

### Wrap with fmt.Errorf

```go
func (s *Service) CreateOrder(ctx context.Context, order Order) error {
    if err := order.Validate(); err != nil {
        return fmt.Errorf("validate order: %w", err)
    }

    if err := s.repo.Save(ctx, order); err != nil {
        return fmt.Errorf("save order: %w", err)
    }

    return nil
}
```

### Error Chain Example

```go
// Layer 1: Repository
func (r *Repo) GetUser(id string) (*User, error) {
    row := r.db.QueryRow("SELECT * FROM users WHERE id = $1", id)
    var u User
    if err := row.Scan(&u.ID, &u.Name); err != nil {
        return nil, fmt.Errorf("scan user %s: %w", id, err)
    }
    return &u, nil
}

// Layer 2: Service
func (s *Service) ProcessUser(id string) error {
    user, err := s.repo.GetUser(id)
    if err != nil {
        return fmt.Errorf("get user: %w", err)
    }
    return user.Process()
}

// Layer 3: Handler
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
    if err := h.service.ProcessUser(id); err != nil {
        // Error message: "get user: scan user 123: sql: no rows in result set"
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
```

## Error Inspection

### errors.Is (Check Error Type)

```go
import "errors"

var ErrNotFound = errors.New("not found")

func (r *Repo) GetUser(id string) (*User, error) {
    // ...
    if rowCount == 0 {
        return nil, ErrNotFound
    }
    // ...
}

func (s *Service) ProcessUser(id string) error {
    user, err := s.repo.GetUser(id)
    if errors.Is(err, ErrNotFound) {
        return fmt.Errorf("user %s does not exist", id)
    }
    if err != nil {
        return fmt.Errorf("get user: %w", err)
    }
    return user.Process()
}
```

**Key points:**
- `errors.Is` checks entire error chain
- Works even if error was wrapped multiple times
- Use for sentinel errors (predefined package-level errors)

### errors.As (Extract Error Type)

```go
type ValidationError struct {
    Field string
    Msg   string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}

func handleRequest(r Request) error {
    if err := validateRequest(r); err != nil {
        var validationErr *ValidationError
        if errors.As(err, &validationErr) {
            log.Printf("validation failed: field=%s msg=%s",
                validationErr.Field, validationErr.Msg)
            return fmt.Errorf("invalid request: %w", err)
        }
        return err
    }
    return nil
}
```

**Key points:**
- `errors.As` finds first error in chain matching target type
- Target must be pointer to error type
- Useful for extracting structured error information

## Custom Error Types

### Structured Errors

```go
type DBError struct {
    Op    string    // Operation (e.g., "query", "exec")
    Query string    // SQL query
    Err   error     // Underlying error
}

func (e *DBError) Error() string {
    return fmt.Sprintf("db %s: %v", e.Op, e.Err)
}

func (e *DBError) Unwrap() error {
    return e.Err
}

// Usage
func (r *Repo) query(ctx context.Context, sql string) error {
    _, err := r.db.ExecContext(ctx, sql)
    if err != nil {
        return &DBError{
            Op:    "exec",
            Query: sql,
            Err:   err,
        }
    }
    return nil
}

// Inspection
func handleError(err error) {
    var dbErr *DBError
    if errors.As(err, &dbErr) {
        log.Printf("DB error: op=%s query=%s err=%v",
            dbErr.Op, dbErr.Query, dbErr.Err)
    }
}
```

### Error with Stack Trace

```go
import "github.com/pkg/errors"

func readConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, errors.Wrap(err, "read config file")
        // Captures stack trace at wrap point
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, errors.Wrap(err, "unmarshal config")
    }

    return &cfg, nil
}
```

## Sentinel Errors

### Package-Level Errors

```go
package user

import "errors"

var (
    ErrNotFound      = errors.New("user not found")
    ErrAlreadyExists = errors.New("user already exists")
    ErrInvalidEmail  = errors.New("invalid email address")
)

// Usage
func (s *Service) CreateUser(email string) error {
    if !isValidEmail(email) {
        return ErrInvalidEmail
    }

    if s.exists(email) {
        return ErrAlreadyExists
    }

    return s.repo.Save(email)
}

// Consumer checks
if errors.Is(err, user.ErrNotFound) {
    // Handle not found
}
```

**Key points:**
- Define at package level
- Use for well-known error conditions
- Consumers can check with `errors.Is`

## Multi-Error Handling

### Collect Multiple Errors

```go
import "errors"

func validateUser(u User) error {
    var errs []error

    if u.Name == "" {
        errs = append(errs, errors.New("name is required"))
    }

    if u.Email == "" {
        errs = append(errs, errors.New("email is required"))
    }

    if len(u.Password) < 8 {
        errs = append(errs, errors.New("password too short"))
    }

    return errors.Join(errs...) // Returns nil if errs is empty
}

// Inspection (Go 1.20+)
func handleErrors(err error) {
    for _, e := range errors.Unwrap(err) {
        log.Println(e)
    }
}
```

### Using multierror (hashicorp)

```go
import "github.com/hashicorp/go-multierror"

func processItems(items []Item) error {
    var result error

    for _, item := range items {
        if err := item.Process(); err != nil {
            result = multierror.Append(result, err)
        }
    }

    return result
}
```

## Error Handling Patterns

### Defer for Cleanup

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("open file: %w", err)
    }
    defer f.Close() // Cleanup even if error occurs

    data := make([]byte, 1024)
    if _, err := f.Read(data); err != nil {
        return fmt.Errorf("read file: %w", err)
    }

    return nil
}
```

### Named Return for Defer Error Handling

```go
func saveData(path string, data []byte) (err error) {
    f, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("create file: %w", err)
    }

    defer func() {
        if closeErr := f.Close(); closeErr != nil && err == nil {
            err = fmt.Errorf("close file: %w", closeErr)
        }
    }()

    if _, err = f.Write(data); err != nil {
        return fmt.Errorf("write file: %w", err)
    }

    return nil
}
```

### Panic and Recover (Rare)

```go
func safeProcess() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic: %v", r)
        }
    }()

    // Code that might panic
    riskyOperation()

    return nil
}
```

**Use panic/recover only for:**
- Programming errors (nil pointer, index out of bounds)
- Unrecoverable situations
- Library initialization failures

**Never use for:**
- Control flow
- Expected error conditions
- Normal validation

## HTTP Error Mapping

### Domain Errors to HTTP Status

```go
func mapErrorToStatus(err error) int {
    switch {
    case errors.Is(err, ErrNotFound):
        return http.StatusNotFound
    case errors.Is(err, ErrAlreadyExists):
        return http.StatusConflict
    case errors.Is(err, ErrInvalidInput):
        return http.StatusBadRequest
    case errors.Is(err, ErrUnauthorized):
        return http.StatusUnauthorized
    default:
        return http.StatusInternalServerError
    }
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := h.service.Process(r.Context()); err != nil {
        status := mapErrorToStatus(err)
        http.Error(w, err.Error(), status)
        return
    }

    w.WriteHeader(http.StatusOK)
}
```

## Common Mistakes

| Mistake                           | Fix                             |
|-----------------------------------|---------------------------------|
| `fmt.Errorf("error: %v", err)`    | Use `%w` to wrap, not `%v`      |
| `errors.New(fmt.Sprintf(...))`    | Use `fmt.Errorf` directly       |
| Uppercase error messages          | Lowercase, no punctuation       |
| `return nil, err` without context | Add operation context with `%w` |
| Ignoring errors                   | Always check `err != nil`       |
| Using `panic` for validation      | Return error instead            |
| `err.Error() == "..."`            | Use `errors.Is` or `errors.As`  |

## Best Practices

```go
// ✓ Good
return fmt.Errorf("parse config: %w", err)

// ✗ Bad
return fmt.Errorf("Error parsing config file: %v", err)

// ✓ Good - lowercase, no punctuation, context first
return fmt.Errorf("save user %s: %w", id, err)

// ✗ Bad - uppercase, punctuation, error first
return fmt.Errorf("Failed to save user. Error: %v", err)

// ✓ Good - specific error type
var validationErr *ValidationError
if errors.As(err, &validationErr) {
    // Handle validation error
}

// ✗ Bad - string matching
if strings.Contains(err.Error(), "validation") {
    // Fragile, breaks if error message changes
}
```

## See Also

- [Custom Error Types in Go](https://go.dev/blog/error-handling-and-go)
- [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [errors package](https://pkg.go.dev/errors)
- [fmt.Errorf](https://pkg.go.dev/fmt#Errorf)
