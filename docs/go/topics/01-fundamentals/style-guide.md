# Style Guide

Consensus patterns from Uber Go Style Guide and Google Go Style Guide.

## Quick Reference

| Rule           | Pattern                                               |
|----------------|-------------------------------------------------------|
| Naming         | `camelCase` for unexported, `PascalCase` for exported |
| Line length    | 80-100 characters (soft limit)                        |
| Imports        | Grouped: stdlib, external, internal                   |
| Error messages | lowercase, no punctuation                             |
| Pointers       | Use for large structs, mutability                     |

## Naming Conventions

### Variables

```go
// Good - concise, clear
var cfg Config
var db *Database
var mu sync.Mutex
var wg sync.WaitGroup

// Bad - verbose
var applicationConfiguration Config
var databaseConnection *Database
```

### Functions

```go
// Good - verb phrase
func ProcessUser(user User) error
func ValidateEmail(email string) bool
func GetUserByID(id string) (*User, error)

// Bad - noun or unclear
func User(user User) error
func Email(email string) bool
```

### Constants

```go
// Good - use const blocks
const (
    StatusPending  = "pending"
    StatusApproved = "approved"
    StatusRejected = "rejected"
)

// Good - typed constants
const (
    MaxRetries   = 3
    Timeout      = 30 * time.Second
    DefaultPort  = 8080
)
```

### Acronyms

```go
// Good - acronyms are all caps
type HTTPServer struct{}
func ServeHTTP()
var apiURL string
var userID int

// Bad - mixed case
type HttpServer struct{}
func ServeHttp()
var apiUrl string
var userId int
```

## Code Organization

### File Organization

```go
// File: user.go

package user

import (
    // Standard library
    "context"
    "fmt"
    "time"

    // External dependencies
    "github.com/google/uuid"

    // Internal packages
    "example.com/project/internal/db"
)

// Constants and variables
const defaultTimeout = 30 * time.Second

var ErrNotFound = errors.New("user not found")

// Types
type User struct {
    ID    string
    Name  string
    Email string
}

// Constructors
func New(name, email string) *User {
    return &User{
        ID:    uuid.New().String(),
        Name:  name,
        Email: email,
    }
}

// Public methods
func (u *User) Validate() error {
    if u.Name == "" {
        return errors.New("name required")
    }
    return nil
}

// Private methods
func (u *User) sanitize() {
    u.Name = strings.TrimSpace(u.Name)
}

// Public functions
func ProcessUsers(users []User) error {
    // ...
}

// Private functions
func processInternal(u User) error {
    // ...
}
```

### Package Organization

```go
// Bad - generic names
common/
utils/
helpers/

// Good - domain names
email/
user/
payment/
```

## Error Handling

### Error Messages

```go
// Good - lowercase, no punctuation, context first
return fmt.Errorf("parse config: %w", err)
return fmt.Errorf("connect to database: %w", err)
return errors.New("name is required")

// Bad - uppercase, punctuation, unclear
return fmt.Errorf("Failed to parse config: %v", err)
return fmt.Errorf("Error: %v", err)
return errors.New("Name is required.")
```

### Error Wrapping

```go
// Good - wrap with context
if err != nil {
    return fmt.Errorf("save user: %w", err)
}

// Bad - lose context
if err != nil {
    return err
}

// Bad - use %v (not %w)
if err != nil {
    return fmt.Errorf("error: %v", err)
}
```

## Receivers

### Receiver Names

```go
// Good - short (1-2 letters)
func (u *User) Validate() error
func (s *Server) Start() error
func (c *Client) Connect() error

// Bad - verbose
func (user *User) Validate() error
func (this *Server) Start() error
func (self *Client) Connect() error
```

### Pointer vs Value Receivers

```go
// Use pointer receiver when:
// 1. Method modifies receiver
func (u *User) SetName(name string) {
    u.Name = name
}

// 2. Receiver is large struct
type LargeStruct struct {
    data [1000]int
}

func (l *LargeStruct) Process() {
    // Use pointer to avoid copying
}

// 3. Consistency (if one method needs pointer, use pointer for all)
func (u *User) Validate() error  { return nil }
func (u *User) Save() error      { return nil } // Pointer for consistency

// Use value receiver when:
// 1. Method doesn't modify receiver
// 2. Receiver is small struct or primitive
func (c Color) String() string {
    return fmt.Sprintf("#%06x", int(c))
}
```

## Function Parameters

### Parameter Ordering

```go
// Good - context first, options last
func ProcessUser(ctx context.Context, userID string, opts ...Option) error

// Bad - context not first
func ProcessUser(userID string, ctx context.Context) error
```

### Named Result Parameters

```go
// Good - named returns for documentation
func divide(a, b int) (result int, err error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// Avoid naked returns (return without values)
func divide(a, b int) (result int, err error) {
    if b == 0 {
        err = errors.New("division by zero")
        return // Bad - naked return
    }
    result = a / b
    return // Bad - naked return
}
```

## Formatting

### Line Length

```go
// Good - break long lines
user, err := service.GetUserByEmailAndDomain(
    ctx,
    email,
    domain,
)

// Good - break long conditions
if user.IsActive &&
    user.EmailVerified &&
    user.HasPermission("admin") {
    // ...
}
```

### Grouping

```go
// Good - group related declarations
const (
    StatusActive   = "active"
    StatusInactive = "inactive"
)

var (
    ErrNotFound = errors.New("not found")
    ErrInvalid  = errors.New("invalid")
)

// Bad - ungrouped
const StatusActive = "active"
const StatusInactive = "inactive"
var ErrNotFound = errors.New("not found")
var ErrInvalid = errors.New("invalid")
```

## Comments

### Package Comments

```go
// Package user provides user management functionality.
//
// This package handles user creation, validation, and persistence.
package user
```

### Function Comments

```go
// GetUser retrieves a user by ID.
// Returns ErrNotFound if user doesn't exist.
func GetUser(ctx context.Context, id string) (*User, error) {
    // ...
}

// Don't repeat function signature in comment
// Bad:
// GetUser gets a user by ID and returns a User pointer and error
```

### Exported Names

```go
// Good - comment exported names
// User represents a system user.
type User struct {
    ID   string
    Name string
}

// ErrNotFound is returned when a user is not found.
var ErrNotFound = errors.New("user not found")
```

## Common Patterns

### Reduce Nesting

```go
// Good
func process() error {
    if err := step1(); err != nil {
        return err
    }

    if err := step2(); err != nil {
        return err
    }

    return step3()
}

// Bad - deeply nested
func process() error {
    if err := step1(); err == nil {
        if err := step2(); err == nil {
            return step3()
        } else {
            return err
        }
    } else {
        return err
    }
}
```

### Initialize Slices

```go
// Good - make with capacity when size known
users := make([]User, 0, 100)

// Good - nil slice for unknown size
var users []User

// Bad - empty slice literal for unknown size
users := []User{} // Allocates unnecessarily
```

### Map Initialization

```go
// Good - make for known size
m := make(map[string]int, 100)

// Good - literal for static data
m := map[string]int{
    "one": 1,
    "two": 2,
}

// Bad - allocate without size hint
m := make(map[string]int)
// ... add 1000 items (multiple resizes)
```

## Tools

### Required Tools

```bash
# Format code
gofmt -w .
# Or stricter formatting
gofumpt -w .

# Lint
golangci-lint run

# Imports
goimports -w .
```

### golangci-lint Configuration

```yaml
# .golangci.yml
linters:
  enable:
    - errcheck
    - gofmt
    - goimports
    - govet
    - ineffassign
    - staticcheck
    - unused
```

## See Also

- [Idioms](./idioms.md) - Go idioms and proverbs
- [Naming](./naming.md) - Naming conventions
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Google Go Style Guide](https://google.github.io/styleguide/go/)
