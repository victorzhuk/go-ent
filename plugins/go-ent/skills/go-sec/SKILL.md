---
name: go-sec
description: "Security patterns, OWASP, authentication, authorization. Auto-activates for: security concerns, authentication, authorization, input validation, secrets."
---

# Go Security

## Tools

```bash
gosec ./...           # Static analysis
govulncheck ./...     # Dependency vulnerabilities  
staticcheck ./...     # Additional checks
```

## Injection Prevention

```go
// ❌ SQL Injection
query := "SELECT * FROM users WHERE id = " + userID

// ✅ Parameterized
sq.Select("*").From("users").Where(sq.Eq{"id": userID})
```

## Authentication

```go
// Password hashing
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
err := bcrypt.CompareHashAndPassword(hash, []byte(password))

// Constant-time comparison
if subtle.ConstantTimeCompare([]byte(a), []byte(b)) != 1 {
    return ErrInvalidToken
}
```

## Sensitive Data

```go
type User struct {
    Email    string `json:"email"`
    Password string `json:"-"` // never serialize
}
```

## JWT

```go
token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, ErrInvalidAlgorithm
    }
    return secretKey, nil
})
```

## Security Headers

```go
func SecureHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000")
        next.ServeHTTP(w, r)
    })
}
```

## Rate Limiting

```go
limiter := rate.NewLimiter(rate.Limit(100), 10)
if !limiter.Allow() {
    http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
    return
}
```

## Audit Checklist

- [ ] Parameterized queries
- [ ] Passwords hashed (bcrypt/argon2)
- [ ] JWT algorithm verified
- [ ] Input validated
- [ ] Secrets from environment
- [ ] Rate limiting on auth
- [ ] Security headers set
- [ ] govulncheck clean
- [ ] No sensitive data in logs
