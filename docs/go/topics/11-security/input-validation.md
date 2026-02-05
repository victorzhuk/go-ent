# Input Validation

Input validation is the first line of defense in application security. It ensures data integrity, prevents injection attacks, and protects against malformed inputs that could cause unexpected behavior or security vulnerabilities.

**Validation vs Sanitization:**
- **Validation**: Check if input meets expected format/constraints (reject if invalid)
- **Sanitization**: Transform input to make it safe (escape, encode, or strip dangerous content)

Both are necessary for defense-in-depth. Always validate first, then sanitize where needed.

---

## Quick Reference

| Category | Pattern | Use Case |
|----------|---------|----------|
| **Struct Tags** | `validate:"required,email,min=5"` | Declarative validation rules |
| **Custom Validators** | `v.RegisterValidation("custom", fn)` | Business logic validation |
| **Cross-Field** | `validate:"eqfield=Password"` | Validate field against another |
| **HTTP Binding** | `c.ShouldBindJSON(&req)` + `validate.Struct(req)` | Validate HTTP request bodies |
| **Query Params** | `validate:"required,uuid4"` | Validate URL parameters |
| **Error Handling** | Type assert `validator.ValidationErrors` | User-friendly error messages |
| **XSS Prevention** | `html.EscapeString()`, `template.HTMLEscapeString()` | Escape HTML output |
| **SQL Injection** | Parameterized queries (pgx, squirrel) | Never concatenate SQL strings |
| **Path Traversal** | `filepath.Clean()`, check `strings.Contains("../")` | Sanitize file paths |

---

## Validator Library

Use `github.com/go-playground/validator/v10` for comprehensive validation.

### Setup

```go
package validation

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

// Validate validates a struct using validator tags
func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}
```

### Struct Tags

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
	ShipTo    Address `json:"ship_to" validate:"required"`
}

type Address struct {
	Street  string `json:"street" validate:"required,min=5,max=200"`
	City    string `json:"city" validate:"required,min=2,max=100"`
	ZipCode string `json:"zip_code" validate:"required,postcode_iso3166_alpha2=US"`
	Country string `json:"country" validate:"required,iso3166_1_alpha2"`
}
```

### Built-in Validators

```go
// String validators
`validate:"required"`         // Non-zero value required
`validate:"omitempty"`         // Skip validation if empty
`validate:"min=3,max=50"`      // Length constraints
`validate:"len=10"`            // Exact length
`validate:"email"`             // Valid email format
`validate:"url"`               // Valid URL format
`validate:"uri"`               // Valid URI format
`validate:"uuid4"`             // Valid UUID v4
`validate:"alphanum"`          // Alphanumeric only
`validate:"alpha"`             // Letters only
`validate:"numeric"`           // Numbers only
`validate:"hexadecimal"`       // Hex string
`validate:"base64"`            // Base64 encoded
`validate:"oneof=red green blue"` // Must be one of values

// Number validators
`validate:"eq=5"`              // Equal to value
`validate:"ne=0"`              // Not equal to value
`validate:"gt=0"`              // Greater than
`validate:"gte=1"`             // Greater than or equal
`validate:"lt=100"`            // Less than
`validate:"lte=99"`            // Less than or equal

// Date/Time validators
`validate:"datetime=2006-01-02"` // Valid date format

// Network validators
`validate:"ip"`                // Valid IP address
`validate:"ipv4"`              // Valid IPv4
`validate:"ipv6"`              // Valid IPv6
`validate:"cidr"`              // Valid CIDR notation
`validate:"mac"`               // Valid MAC address

// File validators
`validate:"file"`              // File exists
`validate:"dir"`               // Directory exists

// Format validators
`validate:"json"`              // Valid JSON
`validate:"jwt"`               // Valid JWT token
`validate:"iso3166_1_alpha2"`  // Country code (US, CA)
`validate:"iso4217"`           // Currency code (USD, EUR)
```

---

## Custom Validators

Register custom validation functions for business logic.

### Simple Custom Validator

```go
package validation

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())

	// Register custom validators
	validate.RegisterValidation("strong_password", validateStrongPassword)
	validate.RegisterValidation("username", validateUsername)
	validate.RegisterValidation("no_profanity", validateNoProfanity)
}

// validateStrongPassword ensures password has uppercase, lowercase, number, special char
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// validateUsername ensures username doesn't contain reserved words
func validateUsername(fl validator.FieldLevel) bool {
	username := strings.ToLower(fl.Field().String())

	reserved := []string{"admin", "root", "system", "moderator", "support"}
	for _, r := range reserved {
		if username == r {
			return false
		}
	}

	return true
}

var profanityList = []string{"badword1", "badword2"} // Load from config

func validateNoProfanity(fl validator.FieldLevel) bool {
	content := strings.ToLower(fl.Field().String())

	for _, word := range profanityList {
		if strings.Contains(content, word) {
			return false
		}
	}

	return true
}
```

### Cross-Field Validation

```go
type ResetPasswordReq struct {
	Password        string `json:"password" validate:"required,min=8,strong_password"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

type DateRangeReq struct {
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
}

// Custom cross-field validator
func init() {
	validate.RegisterStructValidation(validatePriceRange, PriceRangeReq{})
}

type PriceRangeReq struct {
	MinPrice float64 `json:"min_price" validate:"required,gte=0"`
	MaxPrice float64 `json:"max_price" validate:"required,gte=0"`
}

func validatePriceRange(sl validator.StructLevel) {
	req := sl.Current().Interface().(PriceRangeReq)

	if req.MaxPrice <= req.MinPrice {
		sl.ReportError(req.MaxPrice, "max_price", "MaxPrice", "gtfield", "MinPrice")
	}
}
```

### Conditional Validation

```go
type PaymentReq struct {
	Method      string `json:"method" validate:"required,oneof=card paypal crypto"`
	CardNumber  string `json:"card_number" validate:"required_if=Method card,credit_card"`
	PayPalEmail string `json:"paypal_email" validate:"required_if=Method paypal,email"`
	WalletAddr  string `json:"wallet_addr" validate:"required_if=Method crypto,hexadecimal"`
}

type ShippingReq struct {
	Type          string `json:"type" validate:"required,oneof=standard express"`
	ExpressDate   string `json:"express_date" validate:"required_if=Type express,datetime=2006-01-02"`
	TrackingOptIn bool   `json:"tracking_opt_in"`
	TrackingEmail string `json:"tracking_email" validate:"required_if=TrackingOptIn true,email"`
}
```

---

## HTTP Request Validation

Validate all HTTP inputs: JSON bodies, query parameters, path parameters, and headers.

### JSON Body Validation

```go
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type createUserReq struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Password string `json:"password" validate:"required,min=8,strong_password"`
}

func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserReq

	// Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON format")
		return
	}

	// Validate
	if err := h.validate.Struct(req); err != nil {
		respondValidationErrors(w, err)
		return
	}

	// Process valid request
	// ...
}
```

### Query Parameter Validation

```go
type listUsersQuery struct {
	Page     int    `validate:"min=1"`
	PageSize int    `validate:"min=1,max=100"`
	SortBy   string `validate:"omitempty,oneof=name email created_at"`
	Order    string `validate:"omitempty,oneof=asc desc"`
	Role     string `validate:"omitempty,oneof=user admin moderator"`
}

func (h *handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	query := listUsersQuery{
		Page:     parseIntDefault(q.Get("page"), 1),
		PageSize: parseIntDefault(q.Get("page_size"), 20),
		SortBy:   q.Get("sort_by"),
		Order:    parseDefault(q.Get("order"), "asc"),
		Role:     q.Get("role"),
	}

	if err := h.validate.Struct(query); err != nil {
		respondValidationErrors(w, err)
		return
	}

	// Process valid query
	// ...
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func parseDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
```

### Path Parameter Validation

```go
type uuidParam struct {
	ID string `validate:"required,uuid4"`
}

func (h *handler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract path param (example with chi router)
	idStr := chi.URLParam(r, "id")

	param := uuidParam{ID: idStr}
	if err := h.validate.Struct(param); err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID format")
		return
	}

	id, _ := uuid.Parse(idStr) // Safe after validation

	// Process valid ID
	// ...
}
```

### Header Validation

```go
type authHeaders struct {
	Authorization string `validate:"required,startswith=Bearer "`
	RequestID     string `validate:"required,uuid4"`
	ContentType   string `validate:"required,eq=application/json"`
}

func (h *handler) validateHeaders(r *http.Request) error {
	headers := authHeaders{
		Authorization: r.Header.Get("Authorization"),
		RequestID:     r.Header.Get("X-Request-ID"),
		ContentType:   r.Header.Get("Content-Type"),
	}

	return h.validate.Struct(headers)
}
```

---

## Sanitization

Validation rejects invalid input. Sanitization cleans potentially dangerous content.

### XSS Prevention

```go
package sanitize

import (
	"html"
	"html/template"
	"regexp"
	"strings"
)

// SanitizeHTML removes dangerous HTML tags and attributes
func SanitizeHTML(input string) string {
	// Strip script tags
	re := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	input = re.ReplaceAllString(input, "")

	// Strip event handlers
	re = regexp.MustCompile(`(?i)\s*on\w+\s*=\s*["'][^"']*["']`)
	input = re.ReplaceAllString(input, "")

	// Escape remaining HTML
	return html.EscapeString(input)
}

// SanitizeForTemplate prepares user input for HTML templates
func SanitizeForTemplate(input string) template.HTML {
	return template.HTML(template.HTMLEscapeString(input))
}

// StripTags removes all HTML tags
func StripTags(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(input, "")
}
```

### SQL Injection Prevention

**ALWAYS use parameterized queries. NEVER concatenate user input into SQL strings.**

```go
// ✅ SAFE: Parameterized query (pgx)
func (r *repo) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, created_at FROM users WHERE email = $1`

	var u User
	err := r.pool.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.CreatedAt)
	return &u, err
}

// ✅ SAFE: Query builder (squirrel)
func (r *repo) FindByRole(ctx context.Context, role string) ([]*User, error) {
	query, args, _ := r.psql.
		Select("id", "email", "role").
		From("users").
		Where(sq.Eq{"role": role}).
		ToSql()

	rows, err := r.pool.Query(ctx, query, args...)
	// ...
}

// ❌ UNSAFE: String concatenation (NEVER DO THIS)
func (r *repo) FindByEmailUNSAFE(ctx context.Context, email string) (*User, error) {
	query := "SELECT * FROM users WHERE email = '" + email + "'" // SQL injection!
	// ...
}
```

### Path Traversal Prevention

```go
package sanitize

import (
	"errors"
	"path/filepath"
	"strings"
)

var ErrPathTraversal = errors.New("path traversal detected")

// SanitizeFilePath prevents directory traversal attacks
func SanitizeFilePath(userPath, baseDir string) (string, error) {
	// Clean path (remove .., redundant slashes)
	cleanPath := filepath.Clean(userPath)

	// Check for traversal attempts
	if strings.Contains(cleanPath, "..") {
		return "", ErrPathTraversal
	}

	// Join with base directory
	fullPath := filepath.Join(baseDir, cleanPath)

	// Ensure result is still under base directory
	if !strings.HasPrefix(fullPath, baseDir) {
		return "", ErrPathTraversal
	}

	return fullPath, nil
}

// Example usage
func (h *handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")

	safePath, err := sanitize.SanitizeFilePath(filename, "/var/uploads")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid file path")
		return
	}

	http.ServeFile(w, r, safePath)
}
```

---

## Error Messages

Provide user-friendly error messages without leaking sensitive information.

### User-Friendly Errors

```go
package handler

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func respondValidationErrors(w http.ResponseWriter, err error) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		respondError(w, http.StatusBadRequest, "validation failed")
		return
	}

	fieldErrors := make(map[string]string)
	for _, fe := range ve {
		fieldErrors[fe.Field()] = friendlyErrorMessage(fe)
	}

	respondJSON(w, http.StatusBadRequest, map[string]any{
		"error":  "validation failed",
		"fields": fieldErrors,
	})
}

func friendlyErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Must be at least " + fe.Param() + " characters"
	case "max":
		return "Must be at most " + fe.Param() + " characters"
	case "uuid4":
		return "Must be a valid UUID"
	case "oneof":
		return "Must be one of: " + fe.Param()
	case "strong_password":
		return "Password must contain uppercase, lowercase, number, and special character"
	case "eqfield":
		return "Must match " + fe.Param()
	case "gtfield":
		return "Must be greater than " + fe.Param()
	default:
		return "Invalid value"
	}
}
```

### Example Response

```json
{
  "error": "validation failed",
  "fields": {
    "email": "Must be a valid email address",
    "password": "Password must contain uppercase, lowercase, number, and special character",
    "age": "Must be at least 18"
  }
}
```

### Avoid Information Leakage

```go
// ❌ BAD: Leaks database schema
"error": "column 'user_secret_key' does not exist"

// ✅ GOOD: Generic error
"error": "unable to process request"

// ❌ BAD: Reveals validation logic
"error": "password must not contain user's email (john@example.com)"

// ✅ GOOD: Generic constraint
"error": "password does not meet security requirements"

// ❌ BAD: Confirms user existence
"error": "user john@example.com already exists"

// ✅ GOOD: Ambiguous
"error": "unable to create account with provided email"
```

---

## Common Mistakes

| Mistake | Problem | Solution |
|---------|---------|----------|
| **Trusting client input** | Client-side validation can be bypassed | Always validate on server |
| **Insufficient validation** | Validating format but not business rules | Add custom validators for domain logic |
| **Verbose error messages** | Leaking internal details (schema, stack traces) | Return generic errors, log details |
| **Client-side validation only** | JavaScript can be disabled/manipulated | Duplicate all validation server-side |
| **Missing sanitization** | XSS, HTML injection in user-generated content | Sanitize before storage AND display |
| **Wrong regex** | `^admin$` vs `admin` (latter matches "administrator") | Use anchors `^...$` for exact match |
| **No length limits** | DoS via large payloads | Set `max` validators on all strings |
| **Validating after use** | Data used before validation completes | Validate FIRST, then process |
| **Inconsistent validation** | Different rules in different handlers | Centralize validation logic |
| **Ignoring empty values** | `omitempty` allows malicious empty strings | Use `required` when field is mandatory |

---

## See Also

- [Authentication](authentication.md) - Secure authentication patterns
- [HTTP Server](../05-http-grpc/http-server.md) - HTTP handler patterns
- [Error Handling](../02-language/error-handling.md) - Error patterns and sentinel errors
- [PostgreSQL](../04-database/postgresql.md) - Parameterized query examples
- [Security Headers](security-headers.md) - HTTP security headers for defense-in-depth
