---
name: go-error
description: "Implement Go error handling patterns (wrapping, custom errors, error types). Use for error design."
version: "2.0.0"
author: "go-ent"
tags: ["go", "error", "error-handling"]
triggers:
  - keywords: ["error handling", "error wrapping", "custom error", "error type", "sentinel error", "error chain"]
    file_patterns: ["errors.go", "**/errors/*.go", "**/*error*.go"]
    weight: 0.8
---

# Go Error Handling

<role>
Expert Go error handling engineer specializing in error design patterns, wrapping strategies, custom error types, and production-grade error management. Focus on clear error contexts, proper error chains, and idiomatic Go error patterns.
</role>

<instructions>

## Error Handling Stack

- **Error Wrapping** — fmt.Errorf with %w for wrapping
- **Custom Errors** — Error types with methods (Error(), Is(), Unwrap())
- **Sentinel Errors** — errors.New() for comparison
- **Error Chains** — errors.Is() and errors.As() for inspection
- **Domain Errors** — Package-level error types for business logic

## Error Wrapping Pattern

```go
package user

import (
    "fmt"
)

func (r *Repository) GetUser(id string) (*User, error) {
    row := r.db.QueryRow("SELECT id, name FROM users WHERE id = $1", id)
    
    var u User
    if err := row.Scan(&u.ID, &u.Name); err != nil {
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    
    return &u, nil
}
```

**Key points**:
- Wrap errors with context using %w
- Add operation context at each layer
- Don't wrap sentinel errors from your own package
- Preserve error types for Is() and As()

## Custom Error Types

```go
package user

import "fmt"

type InvalidInputError struct {
    Field   string
    Message string
}

func (e *InvalidInputError) Error() string {
    return fmt.Sprintf("invalid input: field '%s' %s", e.Field, e.Message)
}

type UserNotFoundError struct {
    ID string
}

func (e *UserNotFoundError) Error() string {
    return fmt.Sprintf("user not found: %s", e.ID)
}

func (e *UserNotFoundError) Is(target error) bool {
    _, ok := target.(*UserNotFoundError)
    return ok
}
```

**Usage**:
```go
if err != nil {
    if _, ok := err.(*UserNotFoundError); ok {
        return nil, nil
    }
    return err
}
```

## Sentinel Errors

```go
package user

import "errors"

var (
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidInput = errors.New("invalid input")
    ErrUserExists   = errors.New("user already exists")
)

func (r *Repository) GetUser(id string) (*User, error) {
    var u User
    err := r.db.QueryRow("SELECT id, name FROM users WHERE id = $1", id).Scan(&u.ID, &u.Name)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("query user: %w", err)
    }
    return &u, nil
}
```

**Comparison**:
```go
if errors.Is(err, ErrUserNotFound) {
    // Handle not found
}
```

## Error Chains

```go
func (s *Service) CreateUser(req CreateUserRequest) (*User, error) {
    if err := s.validate(req); err != nil {
        return nil, fmt.Errorf("validate request: %w", err)
    }
    
    u, err := s.repo.Create(req)
    if err != nil {
        return nil, fmt.Errorf("create user in repository: %w", err)
    }
    
    if err := s.notify(u); err != nil {
        return nil, fmt.Errorf("notify user: %w", err)
    }
    
    return u, nil
}
```

**Unwrapping**:
```go
err := s.CreateUser(req)

if errors.Is(err, user.ErrUserExists) {
    // Check specific sentinel
}

var notFound *user.UserNotFoundError
if errors.As(err, &notFound) {
    // Extract custom error details
    fmt.Printf("User not found: %s\n", notFound.ID)
}

fmt.Printf("Full chain: %v\n", err)
```

## Domain-Specific Errors

```go
package payment

import "fmt"

type PaymentError struct {
    Code    string
    Message string
    Cause   error
}

func (e *PaymentError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("payment error [%s]: %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("payment error [%s]: %s", e.Code, e.Message)
}

func (e *PaymentError) Unwrap() error {
    return e.Cause
}

func (e *PaymentError) Is(target error) bool {
    t, ok := target.(*PaymentError)
    return ok && e.Code == t.Code
}

var (
    ErrInsufficientFunds = &PaymentError{Code: "INSUFFICIENT_FUNDS", Message: "insufficient funds"}
    ErrPaymentFailed     = &PaymentError{Code: "PAYMENT_FAILED", Message: "payment failed"}
    ErrInvalidCard       = &PaymentError{Code: "INVALID_CARD", Message: "invalid card"}
)
```

## Error Handling Best Practices

1. **Wrap errors with context** — Add operation context at each layer
2. **Use %w for wrapping** — Preserves error type for Is() and As()
3. **Sentinel errors** — For predictable errors (not found, invalid input)
4. **Custom errors** — For domain-specific logic with additional data
5. **Don't check errors twice** — Handle error where it occurs or return it
6. **Handle errors immediately** — Don't defer error handling
7. **Log at appropriate layer** — Don't log and wrap (double logging)
8. **Provide useful context** — Include relevant values in error messages

## Error Documentation

```go
package user

import "errors"

var (
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidEmail = errors.New("invalid email format")
    ErrUserExists   = errors.New("user already exists")
)

type InvalidInputError struct {
    Field   string
    Message string
}

func (e *InvalidInputError) Error() string {
    return fmt.Sprintf("invalid input: %s: %s", e.Field, e.Message)
}

func (e *InvalidInputError) Is(target error) bool {
    _, ok := target.(*InvalidInputError)
    return ok
}
```

## Error Testing

```go
func TestGetUser_NotFound(t *testing.T) {
    repo := NewTestRepository()
    _, err := repo.GetUser("nonexistent")
    
    if !errors.Is(err, ErrUserNotFound) {
        t.Errorf("expected ErrUserNotFound, got %v", err)
    }
}

func TestCreateUser_InvalidInput(t *testing.T) {
    repo := NewTestRepository()
    _, err := repo.CreateUser(User{Email: "invalid"})
    
    var invalid *InvalidInputError
    if !errors.As(err, &invalid) {
        t.Fatalf("expected InvalidInputError, got %v", err)
    }
    
    if invalid.Field != "email" {
        t.Errorf("expected field 'email', got %s", invalid.Field)
    }
}
```

</instructions>

<constraints>
- Include error wrapping with fmt.Errorf and %w verb
- Include custom error types with Error() and optional Is()/Unwrap() methods
- Include sentinel errors using errors.New()
- Include error chain inspection with errors.Is() and errors.As()
- Include domain-specific error patterns
- Include error testing patterns
- Exclude wrapping errors without adding context
- Exclude double logging (log and wrap)
- Exclude checking errors twice (handle once)
- Exclude using error messages for control flow (use error types)
- Exclude returning error without context in outer layers
- Exclude creating custom errors when sentinel errors suffice
</constraints>

<edge_cases>
If error occurs at domain boundary: Wrap with context describing the operation that failed.

If error is already wrapped: Use %w to preserve the error chain for Is()/As() inspection.

If error should be compared: Use sentinel errors or implement Is() method for custom types.

If error contains sensitive data: Redact or omit from error message, log separately with debug level.

If multiple errors occur: Collect and return aggregate error or return first with context about failures.

If error should be ignored: Check with errors.Is() or errors.As() and handle specifically, don't use string matching.

If error needs additional data: Use custom error type with fields, don't encode in error message.

If error occurs in goroutine: Return via channel or sync.ErrGroup with context.

If error handling requires retry: Implement retry logic with backoff, don't rely on error message parsing.
</edge_cases>

<examples>

<example>
<input>Wrap database error with context at repository layer</input>
<output>
```go
package user

import (
    "database/sql"
    "fmt"
)

type Repository struct {
    db *sql.DB
}

func (r *Repository) GetUser(id string) (*User, error) {
    var u User
    err := r.db.QueryRow(
        "SELECT id, name, email FROM users WHERE id = $1",
        id,
    ).Scan(&u.ID, &u.Name, &u.Email)
    
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    
    return &u, nil
}

func (r *Repository) CreateUser(u User) (*User, error) {
    if u.Email == "" {
        return nil, fmt.Errorf("create user: %w", &InvalidInputError{
            Field:   "email",
            Message: "email is required",
        })
    }
    
    err := r.db.QueryRow(
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        u.Name, u.Email,
    ).Scan(&u.ID)
    
    if err != nil {
        return nil, fmt.Errorf("insert user: %w", err)
    }
    
    return &u, nil
}
```
</output>
</example>

<example>
<input>Implement custom error types for domain validation</input>
<output>
```go
package user

import "fmt"

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed: %s: %s", e.Field, e.Message)
}

type UserNotFoundError struct {
    ID string
}

func (e *UserNotFoundError) Error() string {
    return fmt.Sprintf("user not found: %s", e.ID)
}

func (e *UserNotFoundError) Is(target error) bool {
    _, ok := target.(*UserNotFoundError)
    return ok
}

type UserAlreadyExistsError struct {
    Email string
}

func (e *UserAlreadyExistsError) Error() string {
    return fmt.Sprintf("user already exists with email: %s", e.Email)
}

func (e *UserAlreadyExistsError) Is(target error) bool {
    _, ok := target.(*UserAlreadyExistsError)
    return ok
}

func ValidateUser(u User) error {
    if u.Name == "" {
        return &ValidationError{Field: "name", Message: "name is required"}
    }
    if u.Email == "" {
        return &ValidationError{Field: "email", Message: "email is required"}
    }
    return nil
}
```
</output>
</example>

<example>
<input>Define sentinel errors and handle with errors.Is()</input>
<output>
```go
package user

import "errors"

var (
    ErrUserNotFound     = errors.New("user not found")
    ErrInvalidInput     = errors.New("invalid input")
    ErrUserExists       = errors.New("user already exists")
    ErrUnauthorized     = errors.New("unauthorized access")
    ErrInternalError    = errors.New("internal server error")
)

type Service struct {
    repo *Repository
}

func (s *Service) GetUser(id string) (*User, error) {
    u, err := s.repo.GetUser(id)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            return nil, fmt.Errorf("get user: %w", ErrUserNotFound)
        }
        return nil, fmt.Errorf("get user: %w", err)
    }
    return u, nil
}

func (s *Service) DeleteUser(id string) error {
    if err := s.repo.Delete(id); err != nil {
        if errors.Is(err, ErrUserNotFound) {
            return ErrUserNotFound
        }
        return fmt.Errorf("delete user: %w", err)
    }
    return nil
}

func (s *Service) HandleGetUser(id string) error {
    u, err := s.GetUser(id)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            return fmt.Errorf("user %s not found", id)
        }
        return err
    }
    
    fmt.Printf("Found user: %s\n", u.Name)
    return nil
}
```
</output>
</example>

<example>
<input>Inspect error chains with errors.As() to extract error details</input>
<output>
```go
package user

import (
    "errors"
    "fmt"
)

func HandleError(err error) {
    if err == nil {
        return
    }
    
    switch {
    case errors.Is(err, ErrUserNotFound):
        fmt.Println("User not found")
        
    case errors.Is(err, ErrInvalidInput):
        var ve *ValidationError
        if errors.As(err, &ve) {
            fmt.Printf("Validation error: %s - %s\n", ve.Field, ve.Message)
        }
        
    case errors.Is(err, ErrUserExists):
        var uae *UserAlreadyExistsError
        if errors.As(err, &uae) {
            fmt.Printf("User exists: %s\n", uae.Email)
        }
        
    default:
        var paymentErr *payment.PaymentError
        if errors.As(err, &paymentErr) {
            fmt.Printf("Payment error [%s]: %s\n", paymentErr.Code, paymentErr.Message)
            if paymentErr.Cause != nil {
                fmt.Printf("  Caused by: %v\n", paymentErr.Cause)
            }
            return
        }
        
        fmt.Printf("Unexpected error: %v\n", err)
    }
}

func ExampleUsage() {
    err := CreateUser(User{Email: "existing@example.com"})
    HandleError(err)
    
    err = GetUser("nonexistent")
    HandleError(err)
    
    err = CreateUser(User{Name: "John", Email: ""})
    HandleError(err)
}
```
</output>
</example>

<example>
<input>Build error chain across application layers with proper context</input>
<output>
```go
package user

import (
    "fmt"
    "net/http"
)

type Service struct {
    repo *Repository
}

func (s *Service) CreateUser(req CreateUserRequest) (*User, error) {
    if err := s.validate(req); err != nil {
        return nil, fmt.Errorf("validate request: %w", err)
    }
    
    u, err := s.repo.Create(req.User)
    if err != nil {
        if errors.Is(err, ErrUserExists) {
            return nil, fmt.Errorf("create user: %w", ErrUserExists)
        }
        return nil, fmt.Errorf("create user in repository: %w", err)
    }
    
    return u, nil
}

func (s *Service) validate(req CreateUserRequest) error {
    if req.Name == "" {
        return &ValidationError{Field: "name", Message: "name is required"}
    }
    if req.Email == "" {
        return &ValidationError{Field: "email", Message: "email is required"}
    }
    return nil
}

type Handler struct {
    service *Service
}

func (h *Handler) HandleCreateUser(w http.ResponseWriter, r *http.Request) error {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return fmt.Errorf("decode request: %w", err)
    }
    
    u, err := h.service.CreateUser(req)
    if err != nil {
        var ve *ValidationError
        if errors.As(err, &ve) {
            return fmt.Errorf("invalid request: %w", ve)
        }
        if errors.Is(err, ErrUserExists) {
            return fmt.Errorf("user already exists: %w", err)
        }
        return fmt.Errorf("create user: %w", err)
    }
    
    return json.NewEncoder(w).Encode(u)
}

func main() {
    handler := &Handler{service: &Service{repo: &Repository{}}}
    
    err := handler.HandleCreateUser(w, r)
    if err != nil {
        var ve *ValidationError
        if errors.As(err, &ve) {
            http.Error(w, ve.Error(), http.StatusBadRequest)
            return
        }
        if errors.Is(err, ErrUserExists) {
            http.Error(w, "user already exists", http.StatusConflict)
            return
        }
        http.Error(w, "internal server error", http.StatusInternalServerError)
        log.Printf("create user failed: %v", err)
    }
}
```
</output>
</example>

</examples>

<output_format>
Provide error handling guidance with the following structure:

1. **Error Patterns**: Wrapping, custom errors, sentinel errors, domain errors
2. **Error Wrapping**: Use fmt.Errorf with %w, add operation context
3. **Custom Types**: Implement Error() with optional Is()/Unwrap()
4. **Sentinel Errors**: Use errors.New() for comparison with errors.Is()
5. **Error Chains**: Inspect with errors.Is() and errors.As()
6. **Error Context**: Add relevant operation context at each layer
7. **Error Testing**: Test with errors.Is() and errors.As()
8. **Examples**: Complete, runnable code showing error patterns

Focus on production-ready error handling with clear context and proper error chains.
</output_format>
