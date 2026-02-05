# Authentication

Authentication verifies user identity through various mechanisms: JWT tokens (stateless), sessions (stateful), API keys, and OAuth2/OIDC (delegated). Choose based on your architecture: JWT for microservices/SPAs, sessions for traditional web apps, OAuth2 for third-party integrations.

**Core approaches:**

- **JWT (Stateless)**: Self-contained tokens, no server-side storage, horizontal scaling friendly
- **Sessions (Stateful)**: Server-side storage, better revocation, traditional web apps
- **API Keys**: Simple service-to-service auth, long-lived credentials
- **OAuth2/OIDC**: Delegated authentication, external identity providers

## Quick Reference

| Operation | Code | Notes |
|-----------|------|-------|
| Hash password | `bcrypt.GenerateFromPassword([]byte(pwd), 12)` | Cost 12-14 for interactive login |
| Verify password | `bcrypt.CompareHashAndPassword(hash, pwd)` | Constant-time comparison |
| Create JWT | `jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)` | HS256 for single service |
| Parse JWT | `jwt.ParseWithClaims(token, &Claims{}, keyFunc)` | Always validate claims |
| Extract from header | `strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")` | Handle missing/malformed |
| Set user in context | `ctx = context.WithValue(ctx, userCtxKey, user)` | Type-safe key |
| Get user from context | `ctx.Value(userCtxKey).(*User)` | Check nil before cast |
| Middleware chain | `authMiddleware(corsMiddleware(handler))` | Auth after CORS |
| OIDC validation | `verifier.Verify(ctx, rawIDToken)` | Validates signature + claims |

## Password Hashing

**Use `bcrypt` for password storage** - adaptive cost, built-in salt, industry standard.

```go
import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12 // 12-14 for interactive login, tune based on latency

// HashPassword creates bcrypt hash with recommended cost
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

// VerifyPassword compares password against bcrypt hash
func VerifyPassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials // Don't expose hash vs password error
		}
		return fmt.Errorf("verify password: %w", err)
	}
	return nil
}
```

**Registration flow:**

```go
func (uc *registerUC) Execute(ctx context.Context, req RegisterReq) (*RegisterResp, error) {
	// Validate input (length, complexity)
	if len(req.Password) < 8 {
		return nil, ErrWeakPassword
	}

	// Hash before storage
	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        req.Email,
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}

	if err := uc.repo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}

	return &RegisterResp{UserID: user.ID}, nil
}
```

**Performance consideration:**

```go
// High-volume scenarios: pre-hash with fast algorithm before bcrypt
// NOT recommended for typical auth (adds complexity)
func HashPasswordHighVolume(password string) (string, error) {
	// 1. Fast hash (SHA-256) to handle arbitrarily long passwords
	h := sha256.Sum256([]byte(password))

	// 2. bcrypt the hash (always 32 bytes, consistent cost)
	hash, err := bcrypt.GenerateFromPassword(h[:], 10)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
```

## JWT Authentication

**Stateless tokens** - signed JSON claims, no database lookup per request.

### Creating Tokens

```go
import (
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

func GenerateTokenPair(userID uuid.UUID, email, role string, secret []byte) (*TokenPair, error) {
	now := time.Now()

	// Access token: short-lived (15 minutes)
	accessClaims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "myapp",
			Subject:   userID.String(),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(secret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	// Refresh token: long-lived (7 days), store jti in DB for revocation
	refreshClaims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
		Issuer:    "myapp",
		Subject:   userID.String(),
		ID:        uuid.NewString(), // jti for revocation tracking
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(secret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(15 * time.Minute / time.Second),
	}, nil
}
```

### Validating Tokens

```go
func ValidateAccessToken(tokenString string, secret []byte) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Additional validation
	if claims.Issuer != "myapp" {
		return nil, ErrInvalidIssuer
	}

	return claims, nil
}
```

### RSA Signing (Microservices)

```go
// Load keys once at startup
func loadRSAKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKeyPEM, err := os.ReadFile("private.pem")
	if err != nil {
		return nil, nil, err
	}

	block, _ := pem.Decode(privateKeyPEM)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	publicKeyPEM, err := os.ReadFile("public.pem")
	if err != nil {
		return nil, nil, err
	}

	block, _ = pem.Decode(publicKeyPEM)
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	publicKey := publicKeyInterface.(*rsa.PublicKey)
	return privateKey, publicKey, nil
}

// Sign with RS256
func signTokenRS256(claims *Claims, privateKey *rsa.PrivateKey) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

// Verify with RS256
func verifyTokenRS256(tokenString string, publicKey *rsa.PublicKey) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}
```

## Session Authentication

**Stateful sessions** - server-side storage, better control over revocation.

### Cookie-Based Sessions

```go
import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"
)

type Session struct {
	ID        string
	UserID    uuid.UUID
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Generate cryptographically secure session ID
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Create session and set cookie
func (uc *loginUC) Execute(ctx context.Context, req LoginReq) (*LoginResp, error) {
	user, err := uc.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials // Don't leak user existence
	}

	if err := VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("generate session: %w", err)
	}

	session := &Session{
		ID:        sessionID,
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := uc.sessionRepo.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}

	return &LoginResp{SessionID: sessionID}, nil
}

// Set secure cookie
func setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,  // Prevent XSS
		Secure:   true,  // HTTPS only
		SameSite: http.SameSiteStrictMode, // CSRF protection
	})
}
```

### Session Storage (Redis)

```go
import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
)

type sessionRepo struct {
	rdb *redis.Client
}

func (r *sessionRepo) Save(ctx context.Context, s *Session) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	ttl := time.Until(s.ExpiresAt)
	return r.rdb.Set(ctx, "session:"+s.ID, data, ttl).Err()
}

func (r *sessionRepo) FindByID(ctx context.Context, id string) (*Session, error) {
	data, err := r.rdb.Get(ctx, "session:"+id).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *sessionRepo) Delete(ctx context.Context, id string) error {
	return r.rdb.Del(ctx, "session:"+id).Err()
}
```

## Middleware Pattern

**Extract and validate credentials** - set authenticated user in context.

```go
type ctxKey int

const (
	userCtxKey ctxKey = iota
)

// JWT middleware
func AuthMiddleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				http.Error(w, "invalid authorization format", http.StatusUnauthorized)
				return
			}

			claims, err := ValidateAccessToken(token, secret)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// Set user in context
			ctx := context.WithValue(r.Context(), userCtxKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Session middleware
func SessionMiddleware(repo SessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil {
				http.Error(w, "missing session", http.StatusUnauthorized)
				return
			}

			session, err := repo.FindByID(r.Context(), cookie.Value)
			if err != nil {
				http.Error(w, "invalid session", http.StatusUnauthorized)
				return
			}

			if session.ExpiresAt.Before(time.Now()) {
				http.Error(w, "session expired", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userCtxKey, session.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Extract user from context
func UserFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userCtxKey).(uuid.UUID)
	return userID, ok
}
```

## OAuth2/OIDC

**Delegate authentication** to external providers (Google, GitHub, Auth0).

### OIDC with Google

```go
import (
	"context"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCConfig struct {
	Provider     *oidc.Provider
	OAuth2Config oauth2.Config
	Verifier     *oidc.IDTokenVerifier
}

func NewOIDCConfig(ctx context.Context, issuerURL, clientID, clientSecret, redirectURL string) (*OIDCConfig, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &OIDCConfig{
		Provider:     provider,
		OAuth2Config: oauth2Config,
		Verifier:     verifier,
	}, nil
}

// Initiate login
func (c *OIDCConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateRandomState() // CSRF protection, store in session
	http.Redirect(w, r, c.OAuth2Config.AuthCodeURL(state), http.StatusFound)
}

// Handle callback
func (c *OIDCConfig) HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Verify state parameter (CSRF protection)
	state := r.URL.Query().Get("state")
	if !verifyState(state) {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	// Exchange authorization code for token
	oauth2Token, err := c.OAuth2Config.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Extract ID token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token", http.StatusInternalServerError)
		return
	}

	// Verify ID token signature and claims
	idToken, err := c.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(w, "failed to verify token", http.StatusInternalServerError)
		return
	}

	// Extract claims
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "failed to parse claims", http.StatusInternalServerError)
		return
	}

	// Create or update user in your system
	user, err := upsertUser(ctx, claims.Email, claims.Name)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	// Create session for user
	sessionID := createSession(user.ID)
	setSessionCookie(w, sessionID)
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}
```

## Common Mistakes

| Mistake | Impact | Fix |
|---------|--------|-----|
| Storing passwords in plaintext | Complete credential compromise | Always use bcrypt/argon2 |
| Weak JWT secrets | Token forgery | 32+ random bytes, rotate regularly |
| Not validating JWT claims (exp, iss, aud) | Expired/malicious tokens accepted | Validate all relevant claims |
| Missing HTTPS | Credentials transmitted in clear | Enforce TLS in production |
| Timing attacks in password comparison | User enumeration | Use constant-time comparison (bcrypt does this) |
| Hardcoded secrets in code | Secret leakage in VCS | Use environment variables |
| No rate limiting on login | Brute force attacks | Limit attempts per IP/user |
| Logging sensitive data | Credential exposure | Never log passwords/tokens |
| Using single global context key | Context value collisions | Type-safe unexported keys |
| Not revoking sessions on logout | Session hijacking | Delete session from storage |
| Accepting tokens without "Bearer" prefix | Protocol violations | Validate Authorization header format |
| Using JWT for sessions | Can't revoke until expiry | Use stateful sessions or short JWT TTL |

## See Also

- [Input Validation](input-validation.md) - Request validation, sanitization
- [Rate Limiting](rate-limiting.md) - Brute force protection
- [Security Headers](security-headers.md) - CORS, CSP, HSTS
- [HTTP Server](../05-http-grpc/http-server.md) - Middleware patterns
- [Error Handling](../02-language/error-handling.md) - Secure error responses
