# Custom Error Types

Example showing domain validation error types with custom methods.

## Example

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
