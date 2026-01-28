# Secure Authentication Implementation

Complete example showing secure authentication with password hashing, rate limiting, and security best practices.

## Example

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
```

**Security Features**: bcrypt hashing, rate limiting, account lockout, constant-time comparison, generic errors, password validation, security logging.
</output>
</example>
