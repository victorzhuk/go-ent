# Naming Conventions

Go naming conventions for variables, types, packages, and receivers.

## Quick Reference

| Element    | Convention             | Example                |
|------------|------------------------|------------------------|
| Exported   | PascalCase             | `User`, `ProcessData`  |
| Unexported | camelCase              | `user`, `processData`  |
| Acronyms   | ALL_CAPS               | `HTTPServer`, `apiURL` |
| Package    | lowercase, single word | `user`, `http`         |
| Receiver   | 1-2 letters            | `u *User`, `s *Server` |

## Variables

### Short Names for Short Scopes

```go
// Good - short scope
for i := 0; i < 10; i++ {
    // ...
}

// Good - common abbreviations
var (
    cfg Config
    db  *Database
    ctx context.Context
    req *http.Request
    err error
)

// Good - longer names for longer scopes
func ProcessUserData(userDataConfiguration UserDataConfig) {
    // userDataConfiguration is clear in this larger scope
}
```

### Common Abbreviations

```go
cfg   // config
ctx   // context
db    // database
err   // error
idx   // index
max   // maximum
min   // minimum
msg   // message
num   // number
req   // request
resp  // response
srv   // server
tmp   // temporary
tx    // transaction
val   // value
```

## Types

### Struct Names

```go
// Good - noun, describes what it is
type User struct{}
type Server struct{}
type Request struct{}

// Bad - verb or unclear
type ProcessUser struct{}
type DoServer struct{}
```

### Interface Names

```go
// Good - "-er" suffix for single method
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}

// Good - compose interfaces
type ReadWriter interface {
    Reader
    Writer
}

// Good - descriptive for multiple methods
type UserRepository interface {
    Create(User) error
    Get(id string) (*User, error)
    Update(User) error
}
```

## Functions and Methods

### Function Names

```go
// Good - verb phrase
func ProcessData(data []byte) error
func ValidateEmail(email string) bool
func GetUserByID(id string) (*User, error)

// Don't stutter with package name
// user.NewUser() → user.New()
// user.GetUser() → user.Get()
```

### Getters and Setters

```go
type User struct {
    name string
    age  int
}

// Good - no "Get" prefix
func (u *User) Name() string {
    return u.name
}

// Good - "Set" prefix for setters
func (u *User) SetName(name string) {
    u.name = name
}

// Bad - "Get" prefix
func (u *User) GetName() string {
    return u.name
}
```

## Packages

### Package Names

```go
// Good - lowercase, single word
package user
package http
package time

// Good - short, descriptive
package email
package auth
package cache

// Bad - underscores or mixed case
package user_service
package httpServer

// Bad - generic names
package util
package common
package helper
```

### Avoid Stuttering

```go
// Bad
user.UserService // stutters with package name
http.HTTPServer

// Good
user.Service
http.Server
```

## Constants and Variables

### Constants

```go
// Good - descriptive, grouped
const (
    MaxRetries      = 3
    DefaultTimeout  = 30 * time.Second
    StatusPending   = "pending"
    StatusCompleted = "completed"
)

// Good - typed constants
const (
    StatusActive Status = "active"
    StatusInactive Status = "inactive"
)
```

### Package-Level Variables

```go
// Good - descriptive error variables
var (
    ErrNotFound      = errors.New("not found")
    ErrAlreadyExists = errors.New("already exists")
    ErrInvalidInput  = errors.New("invalid input")
)

// Good - exported for configuration
var (
    DefaultPort    = 8080
    DefaultTimeout = 30 * time.Second
)
```

## Receivers

### Receiver Names

```go
// Good - short, consistent
func (u *User) Validate() error
func (s *Server) Start() error
func (c *Client) Connect() error

// Bad - "this" or "self"
func (this *User) Validate() error
func (self *Server) Start() error

// Bad - full word
func (user *User) Validate() error
```

### Consistency

```go
// Good - same receiver name throughout type
type User struct {
    name string
}

func (u *User) Name() string {
    return u.name
}

func (u *User) SetName(name string) {
    u.name = name
}

// Bad - inconsistent receiver names
func (u *User) Name() string {
    return u.name
}

func (user *User) SetName(name string) {
    user.name = name
}
```

## Acronyms

### Acronym Casing

```go
// Good - acronyms are all caps
type HTTPServer struct{}
type URLPath struct{}
var apiURL string
var userID int64

// Bad - mixed case acronyms
type HttpServer struct{}
type UrlPath struct{}
var apiUrl string
var userId int64

// Exception - at start of unexported
var httpClient *http.Client
var urlParser URLParser
```

## Test Names

### Test Functions

```go
// Good - descriptive test names
func TestUserValidation(t *testing.T)
func TestCreateUser_ValidInput(t *testing.T)
func TestProcessPayment_InsufficientFunds(t *testing.T)

// Good - table-driven test names
tests := []struct {
    name string
    // ...
}{
    {"valid email returns no error", ...},
    {"empty email returns error", ...},
    {"malformed email returns error", ...},
}
```

## File Names

### Source Files

```go
// Good - lowercase, underscores for multi-word
user.go
http_server.go
user_repository.go

// Good - _test.go for tests
user_test.go
http_server_test.go
```

## Common Mistakes

| Mistake               | Fix                               |
|-----------------------|-----------------------------------|
| `getUserByID`         | `GetUserByID` (acronym all caps)  |
| `user.UserService`    | `user.Service` (avoid stuttering) |
| `this *User`          | `u *User` (short receiver)        |
| `util.Helper`         | `email.Validator` (specific name) |
| `package userService` | `package user` (single word)      |

## See Also

- [Idioms](./idioms.md) - Go idioms
- [Style Guide](./style-guide.md) - Uber + Google consensus
- [Effective Go - Names](https://go.dev/doc/effective_go#names)
