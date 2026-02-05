# Security Headers

Security headers and CORS middleware.

## Quick Reference

| Header | Value | Purpose |
|--------|-------|---------|
| `X-Content-Type-Options` | `nosniff` | Prevent MIME type sniffing |
| `X-Frame-Options` | `DENY` / `SAMEORIGIN` | Prevent clickjacking |
| `X-XSS-Protection` | `1; mode=block` | Legacy XSS protection (deprecated) |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains; preload` | Force HTTPS |
| `Content-Security-Policy` | See CSP section | Prevent XSS, data injection |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Control referrer information |
| `Permissions-Policy` | `geolocation=(), microphone=()` | Control browser features |

```go
func securityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        next.ServeHTTP(w, r)
    })
}
```

## CSP Configuration

Content-Security-Policy prevents XSS and data injection attacks.

### CSP Directives

| Directive | Purpose | Example |
|-----------|---------|---------|
| `default-src` | Fallback for other directives | `'self'` |
| `script-src` | JavaScript sources | `'self' 'nonce-{random}'` |
| `style-src` | CSS sources | `'self' 'unsafe-inline'` |
| `img-src` | Image sources | `'self' data: https:` |
| `connect-src` | AJAX, WebSocket | `'self' https://api.example.com` |
| `font-src` | Font sources | `'self' https://fonts.gstatic.com` |
| `frame-ancestors` | Embedding allowed | `'none'` |
| `report-uri` | Violation reporting | `https://example.com/csp-report` |

```go
import (
    "crypto/rand"
    "encoding/base64"
)

func generateNonce() string {
    b := make([]byte, 16)
    rand.Read(b)
    return base64.StdEncoding.EncodeToString(b)
}

type nonceKey struct{}

func cspMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        nonce := generateNonce()
        ctx := context.WithValue(r.Context(), nonceKey{}, nonce)

        csp := fmt.Sprintf(
            "default-src 'self'; script-src 'self' 'nonce-%s'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
            nonce,
        )
        w.Header().Set("Content-Security-Policy", csp)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func getNonce(ctx context.Context) string {
    nonce, _ := ctx.Value(nonceKey{}).(string)
    return nonce
}
```

## Header Explanations

### X-Frame-Options

Prevents clickjacking by controlling iframe embedding.

```go
w.Header().Set("X-Frame-Options", "DENY")        // Never allow framing
w.Header().Set("X-Frame-Options", "SAMEORIGIN")  // Only same origin
```

### X-Content-Type-Options

Prevents MIME type sniffing.

```go
w.Header().Set("X-Content-Type-Options", "nosniff")
```

### Strict-Transport-Security (HSTS)

Forces HTTPS for specified duration.

```go
w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
```

- `max-age`: 2 years (63072000 seconds)
- `includeSubDomains`: Apply to all subdomains
- `preload`: Submit to HSTS preload list

### Referrer-Policy

Controls referrer information sent with requests.

```go
w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
```

## CORS

Basic CORS for public APIs.

```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

## CORS with Credentials

Proper CORS setup for credentialed requests.

```go
func corsWithCredentials(allowedOrigins []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")

            allowed := false
            for _, o := range allowedOrigins {
                if o == origin {
                    allowed = true
                    break
                }
            }

            if allowed {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Access-Control-Allow-Credentials", "true")
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                w.Header().Set("Access-Control-Max-Age", "86400")
            }

            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusNoContent)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### Preflight Handling

```go
func handlePreflight(w http.ResponseWriter, r *http.Request) bool {
    if r.Method != "OPTIONS" {
        return false
    }

    requestMethod := r.Header.Get("Access-Control-Request-Method")
    requestHeaders := r.Header.Get("Access-Control-Request-Headers")

    if requestMethod == "" {
        return false
    }

    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    if requestHeaders != "" {
        w.Header().Set("Access-Control-Allow-Headers", requestHeaders)
    }
    w.Header().Set("Access-Control-Max-Age", "86400")
    w.WriteHeader(http.StatusNoContent)
    return true
}
```

## Testing

### Header Verification

```go
func TestSecurityHeaders(t *testing.T) {
    handler := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req := httptest.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)

    tests := []struct {
        header string
        want   string
    }{
        {"X-Content-Type-Options", "nosniff"},
        {"X-Frame-Options", "DENY"},
        {"Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload"},
        {"Referrer-Policy", "strict-origin-when-cross-origin"},
    }

    for _, tt := range tests {
        t.Run(tt.header, func(t *testing.T) {
            got := w.Header().Get(tt.header)
            if got != tt.want {
                t.Errorf("%s = %q, want %q", tt.header, got, tt.want)
            }
        })
    }
}
```

### CORS Testing

```go
func TestCORS(t *testing.T) {
    allowedOrigins := []string{"https://example.com"}
    handler := corsWithCredentials(allowedOrigins)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    t.Run("allowed origin", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/", nil)
        req.Header.Set("Origin", "https://example.com")
        w := httptest.NewRecorder()
        handler.ServeHTTP(w, req)

        if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
            t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "https://example.com")
        }
        if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
            t.Errorf("Access-Control-Allow-Credentials = %q, want %q", got, "true")
        }
    })

    t.Run("disallowed origin", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/", nil)
        req.Header.Set("Origin", "https://evil.com")
        w := httptest.NewRecorder()
        handler.ServeHTTP(w, req)

        if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
            t.Errorf("Access-Control-Allow-Origin should be empty, got %q", got)
        }
    })

    t.Run("preflight", func(t *testing.T) {
        req := httptest.NewRequest("OPTIONS", "/", nil)
        req.Header.Set("Origin", "https://example.com")
        req.Header.Set("Access-Control-Request-Method", "POST")
        w := httptest.NewRecorder()
        handler.ServeHTTP(w, req)

        if w.Code != http.StatusNoContent {
            t.Errorf("preflight status = %d, want %d", w.Code, http.StatusNoContent)
        }
    })
}
```

## Common Mistakes

| Mistake | Problem | Fix |
|---------|---------|-----|
| Missing HSTS | Traffic vulnerable to downgrade attacks | Add `Strict-Transport-Security` with long `max-age` |
| Weak CSP | `unsafe-inline` everywhere defeats XSS protection | Use nonces for inline scripts, avoid `unsafe-inline` in `script-src` |
| CORS wildcard with credentials | `Access-Control-Allow-Origin: *` with credentials not allowed by spec | Use specific origin validation |
| No `X-Frame-Options` | Vulnerable to clickjacking | Set to `DENY` or `SAMEORIGIN` |
| Wrong header values | Typos in header values silently fail | Test headers in automated tests |
| HSTS without HTTPS | Header ignored on HTTP | Only serve application over HTTPS |
| Overly permissive CORS | Allow all origins for sensitive data | Validate specific allowed origins |
| CSP report-uri without handler | CSP violations not monitored | Implement violation reporting endpoint |

## See Also

- [Rate Limiting](./rate-limiting.md)
- [Authentication](./authentication.md)
- [HTTP Server](../05-http-grpc/http-server.md)
- [Input Validation](./input-validation.md)
