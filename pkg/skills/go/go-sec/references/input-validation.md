# Input Validation Quick Reference

Extracted from `docs/go/topics/11-security/input-validation.md` (487 lines) → 120 lines of actionable patterns.

## Quick Reference Table

| Category          | Pattern                              | Use Case                       |
|-------------------|--------------------------------------|--------------------------------|
| Struct Tags       | `validate:"required,email,min=5"`    | Declarative validation         |
| Custom Validators | `v.RegisterValidation("custom", fn)` | Business logic validation      |
| Cross-Field       | `validate:"eqfield=Password"`        | Validate field against another |
| HTTP Binding      | `c.ShouldBindJSON(&req)`             | Validate HTTP request bodies   |
| XSS Prevention    | `html.EscapeString()`                | Escape HTML output             |
| SQL Injection     | Parameterized queries                | Never concatenate SQL          |
| Path Traversal    | `filepath.Clean()`, check `"../"`    | Sanitize file paths            |

## Validator Setup

```go
import "github.com/go-playground/validator/v10"

var validate *validator.Validate

func init() {
    validate = validator.New(validator.WithRequiredStructEnabled())
}

func Validate(s interface{}) error {
    if err := validate.Struct(s); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return nil
}
```

## Common Struct Tags

```go
type CreateUserReq struct {
    Email     string `json:"email" validate:"required,email,max=255"`
    Username  string `json:"username" validate:"required,min=3,max=50,alphanum"`
    Password  string `json:"password" validate:"required,min=8,max=128"`
    Age       int    `json:"age" validate:"required,min=18,max=120"`
    Website   string `json:"website" validate:"omitempty,url"`
    Role      string `json:"role" validate:"required,oneof=user admin moderator"`
    AcceptTOS bool   `json:"accept_tos" validate:"required,eq=true"`
}

type CreateOrderReq struct {
    ProductID string  `json:"product_id" validate:"required,uuid4"`
    Quantity  int     `json:"quantity" validate:"required,min=1,max=1000"`
    Price     float64 `json:"price" validate:"required,gt=0"`
    Currency  string  `json:"currency" validate:"required,iso4217"`
}
```

## Custom Validators

```go
// Register custom validator
validate.RegisterValidation("phone_us", func(fl validator.FieldLevel) bool {
    phone := fl.Field().String()
    re := regexp.MustCompile(`^\+1[0-9]{10}$`)
    return re.MatchString(phone)
})

// Use in struct
type User struct {
    Phone string `validate:"required,phone_us"`
}
```

## XSS Prevention

```go
import "html"

// Escape user input before rendering
func SafeHTML(input string) string {
    return html.EscapeString(input)
}

// In template
// {{.UserInput | html}}  // Go templates auto-escape by default
```

## SQL Injection Prevention

```go
import "github.com/Masterminds/squirrel"

// Good - parameterized query
query, args, _ := sq.
    Select("id", "email").
    From("users").
    Where(sq.Eq{"email": userInput}).  // Safe - parameterized
    ToSql()

row := db.QueryRow(ctx, query, args...)

// Bad - string concatenation (DO NOT DO THIS)
query := "SELECT * FROM users WHERE email = '" + userInput + "'"
```

## Path Traversal Prevention

```go
import (
    "path/filepath"
    "strings"
)

func SafeFilePath(userPath, baseDir string) (string, error) {
    // Clean and resolve
    clean := filepath.Clean(userPath)

    // Check for traversal attempts
    if strings.Contains(clean, "..") {
        return "", fmt.Errorf("invalid path: traversal detected")
    }

    // Build absolute path
    abs := filepath.Join(baseDir, clean)

    // Ensure result is within base directory
    if !strings.HasPrefix(abs, baseDir) {
        return "", fmt.Errorf("invalid path: outside base directory")
    }

    return abs, nil
}
```

## User-Friendly Error Messages

```go
import "github.com/go-playground/validator/v10"

func FormatValidationErrors(err error) map[string]string {
    errors := make(map[string]string)
    for _, e := range err.(validator.ValidationErrors) {
        field := e.Field()
        switch e.Tag() {
        case "required":
            errors[field] = field + " is required"
        case "email":
            errors[field] = field + " must be a valid email"
        case "min":
            errors[field] = field + " must be at least " + e.Param() + " characters"
        default:
            errors[field] = field + " is invalid"
        }
    }
    return errors
}
```
