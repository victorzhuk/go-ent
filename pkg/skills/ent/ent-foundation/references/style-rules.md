# Style Rules Quick Reference

Extracted from `docs/go/topics/01-fundamentals/style-guide.md` (453 lines) → 100 lines of actionable patterns.

## Quick Reference Table

| Rule           | Pattern                                |
|----------------|----------------------------------------|
| Naming         | camelCase unexported, PascalCase exported |
| Line length    | 80-100 chars (soft limit)              |
| Imports        | Grouped: stdlib, external, internal    |
| Error messages | lowercase, no punctuation              |
| Pointers       | Large structs, mutability              |

## File Organization Pattern

```go
package user

import (
    "context"        // stdlib
    "fmt"

    "github.com/google/uuid"  // external

    "example.com/project/internal/db"  // internal
)

const defaultTimeout = 30 * time.Second  // constants

var ErrNotFound = errors.New("user not found")  // package vars

type User struct { ID string }  // types

func New() *User { return &User{} }  // constructors

func (u *User) Validate() error { return nil }  // public methods

func (u *User) sanitize() {}  // private methods
```

## Error Handling

```go
// Good - lowercase, context first, wrap with %w
return fmt.Errorf("parse config: %w", err)
return fmt.Errorf("save user: %w", err)
return errors.New("name is required")

// Bad - uppercase, %v instead of %w, no context
return fmt.Errorf("Failed: %v", err)
return err  // loses context
```

## Receivers

```go
// Pointer receiver: method modifies, large struct, consistency
func (u *User) SetName(name string) { u.Name = name }

// Value receiver: small struct, no modification
func (c Color) String() string { return fmt.Sprintf("#%06x", int(c)) }

// Consistency: if one method needs pointer, use pointer for all
```

## Reduce Nesting (Happy Path Left)

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
        }
    }
    return err
}
```

## Initialize Slices

```go
// Good - make with capacity when size known
users := make([]User, 0, 100)

// Good - nil slice for unknown size
var users []User

// Bad - empty slice literal for unknown size
users := []User{}  // unnecessary allocation
```

## Map Initialization

```go
// Good - make for known size
m := make(map[string]int, 100)

// Good - literal for static data
m := map[string]int{"one": 1, "two": 2}

// Bad - no size hint for large maps
m := make(map[string]int)  // will resize
```

## Common Mistakes

| Mistake           | Fix                              |
|-------------------|----------------------------------|
| `apiUrl`          | `apiURL` (acronym all caps)      |
| `user.UserService`| `user.Service` (avoid stutter)   |
| `this *User`      | `u *User` (short receiver)       |
| `util.Helper`     | `email.Validator` (domain name)  |
| `return err`      | `return fmt.Errorf("op: %w", err)` |
