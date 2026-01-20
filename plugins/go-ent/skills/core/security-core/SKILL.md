---
name: security-core
description: "Security fundamentals and OWASP principles. Auto-activates for: authentication, authorization, input validation, SQL injection, XSS, CSRF, security headers."
version: "2.0.0"
author: "go-ent"
tags: ["security", "owasp", "auth", "authorization", "input-validation"]
---

# Security Core

<role>
Security specialist focused on OWASP principles, authentication patterns, and input validation. Prioritize defense in depth, least privilege, and secure-by-default approaches.
</role>

<instructions>

## OWASP Top 10 (2021)

1. **Access Control**: Least privilege, RBAC, validate permissions
2. **Crypto**: TLS, encryption at rest, strong algorithms, bcrypt/argon2
3. **Injection**: Parameterized queries, escape output, validate inputs
4. **Design**: Threat modeling, secure by default, defense in depth
5. **Misconfiguration**: Disable debug, remove defaults, secure headers
6. **Vulnerable Components**: Track dependencies, scan, update
7. **Auth Failures**: MFA, strong passwords, rate limiting, secure sessions
8. **Integrity**: Verify updates, CI/CD security, code signing
9. **Logging**: Log events, monitor, don't log sensitive data
10. **SSRF**: Validate URLs, allowlists, disable unused schemas

## Security Checklist

**Input Validation**: Type, length, format, range; allowlist over blocklist; sanitize; reject unexpected

**Authentication**: Strong passwords, rate limiting, MFA, secure reset, session timeout, logout

**Authorization**: Check permissions, default deny, least privilege, no client-side auth

**Data Protection**: HTTPS, encrypt at rest, secure key management, no secrets in code/logs, secure cookies

## Common Vulnerabilities

| Vulnerability | Prevention |
|---------------|------------|
| SQL Injection | Parameterized queries |
| XSS | Escape output, CSP |
| CSRF | CSRF tokens |
| Path Traversal | Validate paths, allowlist |
| Command Injection | Avoid shell, validate input |

## Security Headers

```
Content-Security-Policy: default-src 'self'
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Strict-Transport-Security: max-age=31536000
```

## Threat Modeling

**STRIDE**: Spoofing, Tampering, Repudiation, Information Disclosure, Denial of Service, Elevation of Privilege

</instructions>

<constraints>
- Apply defense in depth across all layers
- Implement least privilege principle by default
- Validate all input at application boundaries
- Use parameterized queries to prevent injection attacks
- Encrypt sensitive data in transit and at rest
- Implement proper authentication and authorization
- Never store secrets in code or configuration files
- Use strong, up-to-date cryptographic algorithms
- Log security events without exposing sensitive data
- Regularly update dependencies and scan for vulnerabilities

</constraints>

<edge_cases>
If authentication requirements are complex: Simplify by using proven frameworks (OAuth2, OpenID Connect) rather than custom implementations.

If performance conflicts with security: Prioritize security. Consider caching, connection pooling, or hardware acceleration to address performance.

If using new or unfamiliar technology: Request security review before production deployment, especially for cryptographic operations.

If legacy code has security issues: Apply defense in depth with WAF, input validation, and monitoring while planning remediation.

If API requires public access without authentication: Implement rate limiting, IP allowlists, and request signing for critical operations.

If third-party integration requires credentials: Store securely with key management service, rotate regularly, and monitor usage.

If sensitive data must be logged for debugging: Redact or tokenize sensitive fields, restrict log access, and implement log retention policies.

If compliance requirements exist (PCI-DSS, HIPAA, GDPR): Map security controls to compliance requirements and document evidence of implementation.

If dealing with highly sensitive data: Implement additional controls (MFA, hardware security modules, air-gapped systems).

If security adds significant UX friction: Balance security with usability through risk-based authentication and渐进式实施.
</edge_cases>

<examples>
<example>
<input>Implement secure authentication with password hashing</input>
<output>
```go
package auth

import (
    "context"
    "crypto/subtle"
    "errors"
    "fmt"
    "time"

    "golang.org/x/crypto/bcrypt"
    "golang.org/x/time/rate"
)

var (
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrUserNotFound      = errors.New("user not found")
    ErrAccountLocked     = errors.New("account locked")
    ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

type PasswordHasher interface {
    Hash(password string) (string, error)
    Compare(hashedPassword, password string) (bool, error)
}

type BCryptHasher struct {
    cost int
}

func NewBCryptHasher(cost int) *BCryptHasher {
    return &BCryptHasher{cost: cost}
}

func (h *BCryptHasher) Hash(password string) (string, error) {
    if len(password) < 8 {
        return "", fmt.Errorf("password too short (min 8 characters)")
    }
    if len(password) > 128 {
        return "", fmt.Errorf("password too long (max 128 characters)")
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
    if err != nil {
        return "", fmt.Errorf("hash password: %w", err)
    }
    return string(hash), nil
}

func (h *BCryptHasher) Compare(hashedPassword, password string) (bool, error) {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    if err == bcrypt.ErrMismatchedHashAndPassword {
        return false, nil
    }
    if err != nil {
        return false, fmt.Errorf("compare password: %w", err)
    }
    return true, nil
}

type RateLimiter struct {
    limiter *rate.Limiter
}

func NewRateLimiter(rps int) *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(rate.Every(time.Second/time.Duration(rps)), rps),
    }
}

func (l *RateLimiter) Allow() bool {
    return l.limiter.Allow()
}

type AuthenticationService struct {
    repo         UserRepository
    hasher      PasswordHasher
    rateLimiter *RateLimiter
    failedAttempts *FailedLoginTracker
}

func NewAuthenticationService(
    repo UserRepository,
    hasher PasswordHasher,
    rateLimiter *RateLimiter,
) *AuthenticationService {
    return &AuthenticationService{
        repo:         repo,
        hasher:      hasher,
        rateLimiter: rateLimiter,
        failedAttempts: NewFailedLoginTracker(),
    }
}

func (s *AuthenticationService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    // Rate limiting - prevent brute force
    if !s.rateLimiter.Allow() {
        return nil, ErrRateLimitExceeded
    }

    // Check if account is locked
    if s.failedAttempts.IsLocked(req.Username) {
        return nil, ErrAccountLocked
    }

    // Fetch user (don't reveal whether user exists)
    user, err := s.repo.FindByUsername(ctx, req.Username)
    if err != nil {
        // Log failure for security monitoring
        s.failedAttempts.RecordFailed(req.Username)
        return nil, ErrInvalidCredentials
    }

    // Compare passwords using constant-time comparison
    // to prevent timing attacks
    hashedPassword, err := s.repo.GetPasswordHash(ctx, user.ID)
    if err != nil {
        s.failedAttempts.RecordFailed(req.Username)
        return nil, ErrInvalidCredentials
    }

    match, err := s.hasher.Compare(hashedPassword, req.Password)
    if err != nil {
        s.failedAttempts.RecordFailed(req.Username)
        return nil, ErrInvalidCredentials
    }

    if !match {
        s.failedAttempts.RecordFailed(req.Username)
        return nil, ErrInvalidCredentials
    }

    // Clear failed attempts on successful login
    s.failedAttempts.Clear(req.Username)

    // Generate JWT token
    token, err := s.generateToken(user)
    if err != nil {
        return nil, fmt.Errorf("generate token: %w", err)
    }

    // Log successful login for audit
    logSecurityEvent("login_success", map[string]any{
        "user_id": user.ID,
        "username": user.Username,
    })

    return &LoginResponse{
        Token: token,
        User:  user,
    }, nil
}

func (s *AuthenticationService) Register(ctx context.Context, req *RegisterRequest) error {
    // Validate password strength
    if err := validatePassword(req.Password); err != nil {
        return fmt.Errorf("invalid password: %w", err)
    }

    // Hash password
    hashedPassword, err := s.hasher.Hash(req.Password)
    if err != nil {
        return fmt.Errorf("hash password: %w", err)
    }

    // Create user
    user := &User{
        Username: req.Username,
        Email:    req.Email,
        Password: hashedPassword,
    }

    if err := s.repo.Create(ctx, user); err != nil {
        return fmt.Errorf("create user: %w", err)
    }

    logSecurityEvent("user_registered", map[string]any{
        "user_id": user.ID,
        "username": user.Username,
    })

    return nil
}

func validatePassword(password string) error {
    if len(password) < 8 {
        return fmt.Errorf("must be at least 8 characters")
    }
    if len(password) > 128 {
        return fmt.Errorf("must be at most 128 characters")
    }
    hasUpper, hasLower, hasDigit, hasSpecial := false, false, false, false
    for _, c := range password {
        switch {
        case c >= 'A' && c <= 'Z':
            hasUpper = true
        case c >= 'a' && c <= 'z':
            hasLower = true
        case c >= '0' && c <= '9':
            hasDigit = true
        case c >= ' ' && c <= '~':
            hasSpecial = true
        }
    }
    if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
        return fmt.Errorf("must contain uppercase, lowercase, digit, and special character")
    }
    return nil
}

**Security Features**:**Security Features**: bcrypt hashing, rate limiting, account lockout, constant-time comparison, generic errors, password validation, security logging.</example>

<example>
<input>Prevent SQL injection with parameterized queries</input>
<output>
```go
package userrepo

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
    pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *repository {
    return &repository{pool: pool}
}

// ❌ VULNERABLE: String concatenation (SQL injection)
func (r *repository) FindByUsernameVulnerable(ctx context.Context, username string) (*User, error) {
    query := fmt.Sprintf("SELECT id, email FROM users WHERE username = '%s'", username)
    // If username = "admin' OR '1'='1", this returns all users!

    var user User
    err := r.pool.QueryRow(ctx, query).Scan(&user.ID, &user.Email)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// ✅ SECURE: Parameterized query
func (r *repository) FindByUsername(ctx context.Context, username string) (*User, error) {
    const query = `SELECT id, username, email FROM users WHERE username = $1`

    var user User
    err := r.pool.QueryRow(ctx, query, username).Scan(
        &user.ID,
        &user.Username,
        &user.Email,
    )

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("query user: %w", err)
    }

    return &user, nil
}

// ✅ SECURE: Batch insert with parameterized query
func (r *repository) CreateBatch(ctx context.Context, users []*User) error {
    const query = `
        INSERT INTO users (id, username, email, created_at)
        VALUES ($1, $2, $3, $4)
    `

    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    for _, user := range users {
        _, err := tx.Exec(ctx, query,
            user.ID,
            user.Username,
            user.Email,
            user.CreatedAt,
        )
        if err != nil {
            return fmt.Errorf("insert user %s: %w", user.ID, err)
        }
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }

    return nil
}

// ✅ SECURE: Dynamic query with allowlist
func (r *repository) FindByField(ctx context.Context, field string, value any) (*User, error) {
    // Allowlist of valid fields to prevent injection
    allowedFields := map[string]bool{
        "username": true,
        "email":    true,
    }

    if !allowedFields[field] {
        return nil, fmt.Errorf("invalid field: %s", field)
    }

    query := fmt.Sprintf("SELECT id, username, email FROM users WHERE %s = $1", field)

    var user User
    err := r.pool.QueryRow(ctx, query, value).Scan(
        &user.ID,
        &user.Username,
        &user.Email,
    )

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("query user: %w", err)
    }

    return &user, nil
}

// ✅ SECURE: Using squirrel for complex queries
func (r *repository) FindWithFilters(ctx context.Context, filters UserFilters) ([]*User, error) {
    query := sq.Select("id", "username", "email").
        From("users").
        PlaceholderFormat(sq.Dollar)

    if filters.Username != "" {
        query = query.Where(sq.Eq{"username": filters.Username})
    }
    if filters.Email != "" {
        query = query.Where(sq.Eq{"email": filters.Email})
    }
    if filters.CreatedAfter != nil {
        query = query.Where(sq.Gt{"created_at": filters.CreatedAfter})
    }

    sql, args, err := query.ToSql()
    if err != nil {
        return nil, fmt.Errorf("build query: %w", err)
    }

    rows, err := r.pool.Query(ctx, sql, args...)
    if err != nil {
        return nil, fmt.Errorf("query users: %w", err)
    }
    defer rows.Close()

    var users []*User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Username, &user.Email); err != nil {
            return nil, fmt.Errorf("scan user: %w", err)
        }
        users = append(users, &user)
    }

    return users, nil
}

**Key Security Principles**:**Key Security Principles**: Never concatenate user input, use parameterized queries, validate field names with allowlists, use query builders for complex queries.</example>

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

**XSS Prevention Checklist**:**XSS Prevention Checklist**: Input validation with allowlists, output encoding, CSP headers, disable unknown fields, security headers middleware, file upload validation, URL sanitization.</example>
</examples>

<triggers>
  keywords:
    - "security"
    - "authentication"
    - "authorization"
  weight: 0.8
</triggers>

<output_format>
Provide security guidance and implementations:

1. **Vulnerability Prevention**: Code examples showing secure patterns
2. **OWASP Compliance**: Mapping to OWASP Top 10 controls
3. **Input Validation**: Comprehensive validation for all input vectors
4. **Authentication/Authorization**: Secure auth implementations
5. **Defense in Depth**: Multiple layers of security controls
6. **Monitoring**: Logging, alerting, and detection recommendations
7. **Remediation Steps**: Clear fixes for identified vulnerabilities

Focus on practical, implementable security controls that align with industry best practices and standards.
</output_format>
