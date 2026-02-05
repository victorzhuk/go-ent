# Authentication Patterns Quick Reference

Extracted from `docs/go/topics/11-security/authentication.md` (520 lines) → 120 lines of actionable patterns.

## Core Approaches

| Approach         | Use Case                      | Pros                    | Cons                   |
|------------------|-------------------------------|-------------------------|------------------------|
| JWT (Stateless)  | Microservices, SPAs           | Horizontal scaling      | Revocation complexity  |
| Sessions         | Traditional web apps          | Better revocation       | Server-side storage    |
| API Keys         | Service-to-service            | Simple, long-lived      | No user context        |
| OAuth2/OIDC      | Third-party integrations      | Delegated auth          | Complex setup          |

## Quick Reference Table

| Operation            | Code                                                      | Notes                     |
|----------------------|-----------------------------------------------------------|---------------------------|
| Hash password        | `bcrypt.GenerateFromPassword([]byte(pwd), 12)`            | Cost 12-14                |
| Verify password      | `bcrypt.CompareHashAndPassword(hash, pwd)`                | Constant-time             |
| Create JWT           | `jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)` | HS256 single service |
| Parse JWT            | `jwt.ParseWithClaims(token, &Claims{}, keyFunc)`          | Always validate claims    |
| Extract from header  | `strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")` | Handle missing      |
| Set user in context  | `ctx = context.WithValue(ctx, userCtxKey, user)`          | Type-safe key             |
| Get user from context| `ctx.Value(userCtxKey).(*User)`                            | Check nil before cast     |

## Password Hashing (bcrypt)

```go
import "golang.org/x/crypto/bcrypt"

const bcryptCost = 12  // 12-14 for interactive login

func HashPassword(password string) (string, error) {
    if password == "" {
        return "", fmt.Errorf("empty password")
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
    if err != nil {
        return "", fmt.Errorf("hash password: %w", err)
    }
    return string(hash), nil
}

func VerifyPassword(hash, password string) error {
    if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
        if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
            return ErrInvalidCredentials  // Don't expose hash vs password error
        }
        return fmt.Errorf("verify password: %w", err)
    }
    return nil
}
```

## JWT Authentication

```go
import "github.com/golang-jwt/jwt/v5"

type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

func CreateToken(userID, email string, secret []byte) (string, error) {
    claims := Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "my-service",
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secret)
}

func ParseToken(tokenString string, secret []byte) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return secret, nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, fmt.Errorf("invalid token")
}
```

## Middleware Pattern

```go
func AuthMiddleware(secret []byte) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "missing authorization header", http.StatusUnauthorized)
                return
            }

            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            if tokenString == authHeader {
                http.Error(w, "malformed authorization header", http.StatusUnauthorized)
                return
            }

            claims, err := ParseToken(tokenString, secret)
            if err != nil {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), userCtxKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Type-safe context key
type contextKey string
const userCtxKey = contextKey("user")
```
