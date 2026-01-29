# Security-Focused Code Review

Example showing comprehensive security review for an authentication endpoint.

## Example

<example>
<input>Security-focused code review for authentication endpoint</input>
<output>
## Review Feedback: PR #42 - Add Login Endpoint

### Critical Issues (P0 - Block Merge)

**❌ 1. Timing attack vulnerability in password comparison**

```go
// Before (vulnerable)
if hashedPassword == request.Password {
}
```

**Fix**: Use `crypto/subtle.ConstantTimeCompare`:
```go
if subtle.ConstantTimeCompare([]byte(hashedPassword), []byte(request.Password)) == 1 {
}
```

---

**❌ 2. Rate limiting missing**

```go
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    // ... login logic
}
```

**Fix**: Implement rate limiting:
```go
import "golang.org/x/time/rate"

type Handler struct {
    limiter *rate.Limiter
}

func NewHandler() *Handler {
    return &Handler{
        limiter: rate.NewLimiter(rate.Every(time.Second), 10),
    }
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    if !h.limiter.Allow() {
        http.Error(w, "Too many requests", http.StatusTooManyRequests)
        return
    }
}
```

---

### High Priority Issues (P1 - Request Changes)

**⚠️ 3. Error messages leak information**

```go
if err := db.QueryUser(username); err != nil {
    http.Error(w, "User not found", http.StatusNotFound)
}
```

**Issue**: Reveals whether username exists.

**Fix**: Generic error message:
```go
http.Error(w, "Invalid credentials", http.StatusUnauthorized)
```

---

**⚠️ 4. Missing input validation**

```go
func LoginRequest(r *http.Request) (*LoginRequest, error) {
    var req LoginRequest
    json.NewDecoder(r.Body).Decode(&req)
    return &req
}
```

**Fix**: Add validation:
```go
func (req *LoginRequest) Validate() error {
    if req.Username == "" {
        return errors.New("username required")
    }
    if len(req.Password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    return nil
}
```

---

### Medium Priority Issues (P2 - Suggestions)

**💡 5. Add failed login tracking**

```go
type FailedLoginTracker struct {
    attempts map[string]int
    mutex    sync.Mutex
}

func (t *FailedLoginTracker) RecordFailed(username string) {
    t.mutex.Lock()
    defer t.mutex.Unlock()
    t.attempts[username]++
}

func (t *FailedLoginTracker) IsLocked(username string) bool {
    t.mutex.Lock()
    defer t.mutex.Unlock()
    return t.attempts[username] >= 5
}
```

---

**💡 6. Add security headers**

```go
w.Header().Set("X-Content-Type-Options", "nosniff")
w.Header().Set("X-Frame-Options", "DENY")
w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
```

---

### Summary

**Action**: Request changes (2 P0, 2 P1 issues)

**Blocking**: Timing attack fix, rate limiting, generic error messages, input validation

**Notes**: Security is critical for auth endpoints. P2 items can be follow-up PRs.
</output>
</example>
