# Naming Conventions Quick Reference

Extracted from `docs/go/topics/01-fundamentals/naming.md` (334 lines) → 80 lines of actionable patterns.

## Quick Reference Table

| Element    | Convention             | Example                |
|------------|------------------------|------------------------|
| Exported   | PascalCase             | `User`, `ProcessData`  |
| Unexported | camelCase              | `user`, `processData`  |
| Acronyms   | ALL_CAPS               | `HTTPServer`, `apiURL` |
| Package    | lowercase, single word | `user`, `http`         |
| Receiver   | 1-2 letters            | `u *User`, `s *Server` |

## Common Abbreviations

```go
cfg, ctx, db, err, idx, max, min, msg, num
req, resp, srv, tmp, tx, val, pool, rows
```

## Interfaces

```go
// Good - "-er" suffix for single method
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }

// Good - compose interfaces
type ReadWriter interface { Reader; Writer }

// Good - descriptive for multiple methods
type UserRepository interface {
    Create(User) error
    Get(id string) (*User, error)
}
```

## Functions

```go
// Good - verb phrase
func ProcessData(data []byte) error
func ValidateEmail(email string) bool

// Avoid stuttering with package name
// user.NewUser() → user.New()
// user.GetUser() → user.Get()

// Getters: no "Get" prefix
func (u *User) Name() string { return u.name }

// Setters: "Set" prefix
func (u *User) SetName(name string) { u.name = name }
```

## Packages

```go
// Good - lowercase, single word
package user
package http

// Avoid stuttering
user.Service  // not user.UserService
http.Server   // not http.HTTPServer
```

## Receivers

```go
// Good - short (1-2 letters), consistent
func (u *User) Validate() error
func (s *Server) Start() error

// Bad - verbose or inconsistent
func (user *User) Validate() error  // verbose
func (this *Server) Start() error   // avoid "this"/"self"
```

## Acronyms

```go
// Good - all caps
type HTTPServer struct{}
var apiURL string
var userID int64

// Bad - mixed case
type HttpServer struct{}
var apiUrl string

// Exception - at start of unexported
var httpClient *http.Client
```
