---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
triggers:
  - security
  - authentication
  - authorization
  - input validation
---

# ${SKILL_NAME}

## Role

Security expert specializing in secure coding practices, OWASP Top 10 mitigation, and authentication/authorization patterns. Focus on defense-in-depth, principle of least privilege, and secure-by-default implementations.

## Instructions

### Authentication Patterns

```go
// JWT validation middleware
func JWTAuth(secret []byte) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if !strings.HasPrefix(authHeader, "Bearer ") {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            token, err := jwt.Parse(strings.TrimPrefix(authHeader, "Bearer "), func(t *jwt.Token) (any, error) {
                if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
                }
                return secret, nil
            })

            if err != nil || !token.Valid {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### Password Hashing

```go
func HashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", fmt.Errorf("hash password: %w", err)
    }
    return string(hash), nil
}

func CheckPassword(password, hash string) error {
    if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
        return fmt.Errorf("invalid password: %w", err)
    }
    return nil
}
```

### Input Validation

```go
var (
    emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    ErrInvalidEmail    = errors.New("invalid email format")
    ErrPasswordTooShort = errors.New("password must be at least 8 characters")
)

func ValidateEmail(email string) error {
    if !emailRegex.MatchString(email) {
        return ErrInvalidEmail
    }
    return nil
}

func ValidatePassword(password string) error {
    if len(password) < 8 {
        return ErrPasswordTooShort
    }
    return nil
}
```

### SQL Injection Prevention

```go
// Always use parameterized queries - NEVER string concatenation
query, args, _ := sq.Select("id", "email").
    From("users").
    Where(sq.Eq{"email": email}).  // Safe: parameterized
    ToSql()

// Never do this:
// query := "SELECT * FROM users WHERE email = '" + email + "'"
```

### Edge Cases

If authentication scheme is unclear: Ask about existing infrastructure (OAuth2 provider, JWT, session-based) before implementing.

If RBAC requirements are complex: Map out all roles and permissions before implementing, consider using Casbin.

## Examples

### Example 1: Rate limiting middleware

**Input**: Add rate limiting to prevent brute force attacks

**Output**:
```go
func RateLimitMiddleware(limit rate.Limit, burst int) func(http.Handler) http.Handler {
    limiter := rate.NewLimiter(limit, burst)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "too many requests", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Per-IP rate limiter
type IPRateLimiter struct {
    limiters sync.Map
    limit    rate.Limit
    burst    int
}

func (rl *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
    v, _ := rl.limiters.LoadOrStore(ip, rate.NewLimiter(rl.limit, rl.burst))
    return v.(*rate.Limiter)
}
```

### Example 2: Secure token generation

**Input**: Generate a secure password reset token

**Output**:
```go
const tokenLength = 32

func GenerateSecureToken() (string, error) {
    bytes := make([]byte, tokenLength)
    if _, err := rand.Read(bytes); err != nil {
        return "", fmt.Errorf("generate token: %w", err)
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *service) CreatePasswordReset(ctx context.Context, email string) error {
    user, err := s.repo.FindByEmail(ctx, email)
    if err != nil {
        return nil // Don't reveal if email exists
    }

    token, err := GenerateSecureToken()
    if err != nil {
        return fmt.Errorf("create reset token: %w", err)
    }

    expiresAt := time.Now().Add(1 * time.Hour)
    if err := s.repo.SaveResetToken(ctx, user.ID, token, expiresAt); err != nil {
        return fmt.Errorf("save reset token: %w", err)
    }

    return s.mailer.SendPasswordReset(ctx, email, token)
}
```
