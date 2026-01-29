# Input Validation and XSS Prevention

Complete example showing input validation and XSS prevention with security headers.

## Example

<example>
<input>Implement input validation and XSS prevention</input>
<output>
```go
package handlers

import (
    "encoding/json"
    "html"
    "net/http"
    "regexp"
    "strings"

    "github.com/go-playground/validator/v10"
)

// Validator struct
type Validator struct {
    validate *validator.Validate
}

func NewValidator() *Validator {
    v := validator.New()
    v.RegisterValidation("username", validateUsername)
    v.RegisterValidation("safehtml", validateSafeHTML)
    return &Validator{validate: v}
}

// Custom validation for username
func validateUsername(fl validator.FieldLevel) bool {
    username := fl.Field().String()
    // Only allow alphanumeric, dash, underscore
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{3,20}$`, username)
    return matched
}

// Custom validation for safe HTML (no tags)
func validateSafeHTML(fl validator.FieldLevel) bool {
    input := fl.Field().String()
    // Check for HTML tags
    return !strings.ContainsAny(input, "<>")
}

// Request DTO with validation tags
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,username"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,max=128"`
    Bio      string `json:"bio" validate:"max=500,safehtml"`
}

// Response DTO with escaped output
type UserResponse struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Bio      string `json:"bio"`
}

type Handler struct {
    validator *Validator
}

func NewHandler() *Handler {
    return &Handler{validator: NewValidator()}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // 1. Content-Type validation
    ct := r.Header.Get("Content-Type")
    if ct != "application/json" {
        http.Error(w, "Invalid content type", http.StatusUnsupportedMediaType)
        return
    }

    // 2. Decode request
    var req CreateUserRequest
    decoder := json.NewDecoder(r)
    decoder.DisallowUnknownFields() // Prevent mass assignment

    if err := decoder.Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // 3. Validate input
    if err := h.validator.validate.Struct(&req); err != nil {
        var validationErrors []string
        for _, err := range err.(validator.ValidationErrors) {
            validationErrors = append(validationErrors, fmt.Sprintf(
                "%s failed validation: %s",
                err.Field(),
                err.Tag(),
            ))
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]any{
            "error": "Validation failed",
            "details": validationErrors,
        })
        return
    }

    // 4. Process (business logic)
    user, err := h.userService.Create(r.Context(), &req)
    if err != nil {
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }

    // 5. Prepare response with escaped output
    resp := UserResponse{
        ID:       user.ID,
        Username: html.EscapeString(user.Username),  // Escape HTML
        Email:    html.EscapeString(user.Email),     // Escape HTML
        Bio:      html.EscapeString(user.Bio),       // Escape HTML
    }

    // 6. Set security headers
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Content-Type-Options", "nosniff")
    w.Header().Set("X-Frame-Options", "DENY")

    json.NewEncoder(w).Encode(resp)
}

// Security headers middleware
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Security-Policy",
            "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
            "style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self';")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Content-Security-Policy", "frame-ancestors 'none'")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        next.ServeHTTP(w, r)
    })
}

// URL sanitization
func SanitizeURL(input string) string {
    input = strings.ToLower(input)
    if strings.Contains(input, "javascript:") || strings.Contains(input, "data:") || strings.Contains(input, "vbscript:") {
        return ""
    }
    return html.EscapeString(input)
}
```

**XSS Prevention Checklist**: Input validation with allowlists, output encoding, CSP headers, disable unknown fields, security headers middleware, file upload validation, URL sanitization.
</output>
</example>
