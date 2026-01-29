---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
version: "${VERSION}"
author: "${AUTHOR}"
tags: [${TAGS}]
triggers:
  - pattern: "authentication|authorization|security|input validation|xss|csrf|sql injection"
    weight: 0.9
  - keywords: ["security", "auth", "login", "password", "encryption", "csrf", "xss", "injection", "owasp"]
    weight: 0.85
  - filePattern: "*auth*.*,*security*.*,*middleware*.*"
    weight: 0.7
---

# ${SKILL_NAME}

<role>
Security expert specializing in secure coding practices, OWASP Top 10 mitigation, and authentication/authorization patterns.
Focus on defense-in-depth, principle of least privilege, and secure-by-default implementations.
</role>

<instructions>

## Authentication Patterns

Implement secure authentication with proper session management:

```go
// Password hashing with bcrypt (never store plaintext)
import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", fmt.Errorf("hash password: %w", err)
    }
    return string(hash), nil
}

func VerifyPassword(hash, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// JWT token validation with proper claims verification
func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(jwtSecret), nil
    })

    if err != nil {
        return nil, fmt.Errorf("parse token: %w", err)
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        if time.Now().Unix() > claims.ExpiresAt {
            return nil, ErrTokenExpired
        }
        return claims, nil
    }

    return nil, ErrInvalidToken
}
```

## Authorization & RBAC

Implement role-based access control with proper checks:

```go
// Middleware for authorization
func RequireRole(role string) middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := GetUserFromContext(r.Context())
            if user == nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            if !user.HasRole(role) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Resource ownership check
func RequireOwnership() middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := GetUserIDFromContext(r.Context())
            resourceID := GetResourceIDFromPath(r)

            if !IsOwner(userID, resourceID) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

## Input Validation

Validate all inputs at the boundary using allowlist approach:

```go
import "github.com/go-playground/validator/v10"

type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email,max=255"`
    Password string `json:"password" validate:"required,min=8,max=128"`
    Username string `json:"username" validate:"required,alphanum,min=3,max=50"`
}

var validate = validator.New()

func ValidateRequest(req interface{}) error {
    if err := validate.Struct(req); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return nil
}

// Sanitize HTML output to prevent XSS
import "html/template"

func SafeHTML(input string) template.HTML {
    return template.HTMLEscapeString(input)
}
```

## SQL Injection Prevention

Use parameterized queries exclusively:

```go
// GOOD: Parameterized query
func GetUserByID(ctx context.Context, db *pgxpool.Pool, id string) (*User, error) {
    query := `SELECT id, email, name FROM users WHERE id = $1`
    var user User
    err := db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.Name)
    if err != nil {
        return nil, fmt.Errorf("query user: %w", err)
    }
    return &user, nil
}

// NEVER: String concatenation (SQL injection vulnerable)
// func GetUserByID(ctx context.Context, db *pgxpool.Pool, id string) (*User, error) {
//     query := fmt.Sprintf("SELECT id, email, name FROM users WHERE id = '%s'", id) // DANGEROUS!
//     ...
// }
```

## CSRF Protection

Implement CSRF tokens for state-changing operations:

```go
func CSRFMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
            next.ServeHTTP(w, r)
            return
        }

        token := r.Header.Get("X-CSRF-Token")
        if token == "" {
            token = r.FormValue("csrf_token")
        }

        session := GetSession(r)
        if !validateCSRFToken(token, session.CSRFToken) {
            http.Error(w, "Invalid CSRF token", http.StatusForbidden)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

## Security Headers

Set security headers on all responses:

```go
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

        next.ServeHTTP(w, r)
    })
}
```

## Secrets Management

Never hardcode secrets. Use environment variables:

```go
import "github.com/caarlos0/env/v11"

type Config struct {
    DatabaseURL   string `env:"DATABASE_URL,required"`
    JWTSecret     string `env:"JWT_SECRET,required"`
    RedisPassword string `env:"REDIS_PASSWORD,required"`
    APISecret     string `env:"API_SECRET,required"`
}

func LoadConfig() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, fmt.Errorf("load config: %w", err)
    }
    return cfg, nil
}
```

## Rate Limiting

Implement rate limiting to prevent brute force attacks:

```go
import "golang.org/x/time/rate"

func RateLimitMiddleware(limiter *rate.Limiter) middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "Too many requests", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

func NewRateLimiter(r rate.Limit, b int) *rate.Limiter {
    return rate.NewLimiter(r, b)
}
```

## Logging Security Events

Log security events without exposing sensitive data:

```go
func LogSecurityEvent(ctx context.Context, event string, userID string, details map[string]string) {
    logger.InfoContext(ctx, "security event",
        "event", event,
        "user_id", userID,
        "ip", details["ip"],
        "user_agent", details["user_agent"],
    )
}

// Never log passwords, tokens, or secrets
```

</instructions>

<constraints>
- Never store plaintext passwords (always use bcrypt/scrypt/Argon2)
- Always use parameterized queries for database operations
- Validate all inputs at the boundary with allowlists
- Implement defense-in-depth with multiple security layers
- Use HTTPS in production with proper TLS configuration
- Set security headers on all HTTP responses
- Implement rate limiting for authentication endpoints
- Never expose stack traces or detailed error messages to users
- Use environment variables for secrets, never hardcode
- Implement proper session management with secure cookies
- Follow principle of least privilege for all operations
- Validate and sanitize all user-generated content
- Use prepared statements or query builders for SQL operations
- Implement CSRF protection for state-changing operations
- Log security events without exposing sensitive data
- Regular security audits and dependency updates
</constraints>

<edge_cases>
If security requirements conflict with usability: Recommend security-first approach, but suggest user experience improvements that don't compromise security (e.g., progressive auth).

If legacy integration is required with insecure protocols: Implement a secure gateway/middleware layer to sanitize requests and add necessary security controls before forwarding.

If performance impact is a concern with security controls: Recommend profiling to identify bottlenecks, implement caching for non-sensitive operations, and optimize critical paths while maintaining security.

If compliance requirements (GDPR, PCI DSS, etc.) are mentioned: Delegate to security-core skill for comprehensive compliance patterns and audit trail implementations.

If implementing authentication for multiple providers (OAuth, SAML, LDAP): Suggest using established libraries like golang.org/x/oauth2 and warn against building custom auth protocols.

If dealing with cryptographic operations beyond basic hashing: Recommend consulting crypto specialists and using well-vetted cryptographic libraries rather than custom implementations.

If security requirements are unclear: Ask about threat model, data sensitivity, compliance requirements, and threat landscape before making recommendations.
</edge_cases>

<examples>
<example>
<input>Implement secure login endpoint with rate limiting</input>
<output>
```go
package auth

import (
    "context"
    "fmt"
    "net/http"
    "time"
    "golang.org/x/crypto/bcrypt"
    "golang.org/x/time/rate"
)

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type AuthService struct {
    db      *pgxpool.Pool
    limiter *rate.Limiter
}

func NewAuthService(db *pgxpool.Pool) *AuthService {
    return &AuthService{
        db:      db,
        limiter: rate.NewLimiter(rate.Every(5*time.Second), 5),
    }
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*Session, error) {
    if !s.limiter.Allow() {
        return nil, ErrRateLimitExceeded
    }

    if err := ValidateRequest(req); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }

    var storedHash string
    var userID string

    err := s.db.QueryRow(ctx,
        `SELECT id, password_hash FROM users WHERE email = $1`,
        req.Email,
    ).Scan(&userID, &storedHash)

    if err != nil {
        return nil, ErrInvalidCredentials
    }

    if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
        return nil, ErrInvalidCredentials
    }

    session := s.CreateSession(userID)
    return session, nil
}
```
</output>
</example>

<example>
<input>Create middleware to validate JWT tokens and check user roles</input>
<output>
```go
package middleware

import (
    "context"
    "fmt"
    "net/http"
    "strings"
    "github.com/golang-jwt/jwt/v5"
)

type ContextKey string

const (
    UserContextKey ContextKey = "user"
)

type Claims struct {
    UserID string   `json:"user_id"`
    Roles  []string `json:"roles"`
    jwt.RegisteredClaims
}

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "Missing authorization header", http.StatusUnauthorized)
                return
            }

            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
                return
            }

            tokenString := parts[1]

            claims, err := parseToken(tokenString, jwtSecret)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), UserContextKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func RequireRole(role string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims, ok := r.Context().Value(UserContextKey).(*Claims)
            if !ok {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            if !hasRole(claims.Roles, role) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

func parseToken(tokenString, jwtSecret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(jwtSecret), nil
    })

    if err != nil {
        return nil, fmt.Errorf("parse token: %w", err)
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, fmt.Errorf("invalid token claims")
}

func hasRole(roles []string, role string) bool {
    for _, r := range roles {
        if r == role {
            return true
        }
    }
    return false
}
```
</output>
</example>
</examples>

<output_format>
Provide production-ready security code following OWASP best practices:

1. **Authentication**: Secure password hashing, JWT validation, session management
2. **Authorization**: RBAC with middleware, resource ownership checks
3. **Input Validation**: Allowlist validation, sanitization, type safety
4. **SQL Injection**: Parameterized queries only (never string concatenation)
5. **CSRF Protection**: Token-based validation for state-changing operations
6. **Security Headers**: All HTTP responses include security headers
7. **Secrets Management**: Environment variables only, never hardcoded
8. **Rate Limiting**: Protection against brute force and DoS
9. **Logging**: Security events logged without exposing sensitive data
10. **Examples**: Complete, runnable code with error handling

Focus on defense-in-depth and security-by-default patterns.
</output_format>
