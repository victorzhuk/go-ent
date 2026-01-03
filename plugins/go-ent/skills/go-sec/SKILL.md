---
name: go-sec
description: "Security audit, OWASP, auth, crypto. Auto-activates for: security review, authentication, authorization, vulnerability fixes."
---

# Go Security (2026)

## Tools

```bash
gosec ./...           # Static analysis
govulncheck ./...     # Dependency vulnerabilities  
staticcheck ./...     # Additional checks
```

## OWASP Top 10

### 1. Injection

```go
// ❌ SQL Injection
query := "SELECT * FROM users WHERE id = " + userID

// ✅ Parameterized (squirrel)
sq.Select("*").From("users").Where(sq.Eq{"id": userID})

// ❌ Command injection
exec.Command("sh", "-c", "ls " + userInput)

// ✅ No shell
exec.Command("ls", userInput)
```

### 2. Authentication

```go
// Password hashing
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
err := bcrypt.CompareHashAndPassword(hash, []byte(password))

// Constant-time comparison
if subtle.ConstantTimeCompare([]byte(a), []byte(b)) != 1 {
    return ErrInvalidToken
}
```

### 3. Sensitive Data

```go
// ❌ Logging secrets
log.Info("user login", "password", req.Password)

// ✅ Redact
type User struct {
    Email    string `json:"email"`
    Password string `json:"-"` // never serialize
}
```

### 4. JWT Security

```go
token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
    // Verify algorithm
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, ErrInvalidAlgorithm
    }
    return secretKey, nil
})

// Validate claims
if !claims.VerifyExpiresAt(time.Now(), true) {
    return ErrTokenExpired
}
```

### 5. Access Control

```go
func (h *Handler) GetOrder(ctx context.Context, orderID uuid.UUID) (*Order, error) {
    userID := auth.UserIDFromContext(ctx)
    order, err := h.repo.FindByID(ctx, orderID)
    if err != nil {
        return nil, err
    }
    // Verify ownership
    if order.UserID != userID {
        return nil, ErrForbidden
    }
    return order, nil
}
```

### 6. Security Headers

```go
func SecureHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000")
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        next.ServeHTTP(w, r)
    })
}
```

### 7. Input Validation

```go
type CreateUserReq struct {
    Email string `validate:"required,email,max=255"`
    Name  string `validate:"required,min=2,max=100,alphanumunicode"`
}

validate := validator.New()
if err := validate.Struct(req); err != nil {
    return ErrValidation
}
```

### 8. Rate Limiting

```go
limiter := rate.NewLimiter(rate.Limit(100), 10) // 100/s, burst 10

func RateLimit(limiter *rate.Limiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

## Secrets Management

```go
// ✅ Environment
apiKey := os.Getenv("API_KEY")

// ❌ Never hardcode
const apiKey = "sk-123..."
```

## FIPS 140-3 (Go 1.24+)

```bash
GOFIPS140=v1.0.0 go build ./...
GODEBUG=fips140=on ./app
```

## Audit Checklist

- [ ] Parameterized queries (no string concat)
- [ ] Passwords hashed (bcrypt/argon2)
- [ ] JWT algorithm verified
- [ ] Input validated
- [ ] Secrets from environment
- [ ] Rate limiting on auth endpoints
- [ ] Security headers set
- [ ] TLS 1.3 minimum
- [ ] govulncheck clean
- [ ] No sensitive data in logs
