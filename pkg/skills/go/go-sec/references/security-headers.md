# Security Headers Quick Reference

Extracted from `docs/go/topics/11-security/security-headers.md` (464 lines) → 80 lines of actionable patterns.

## Essential Security Headers

| Header                     | Value                                 | Purpose                          |
|----------------------------|---------------------------------------|----------------------------------|
| X-Content-Type-Options     | `nosniff`                             | Prevent MIME type sniffing       |
| X-Frame-Options            | `DENY` or `SAMEORIGIN`                | Prevent clickjacking             |
| X-XSS-Protection           | `1; mode=block`                       | Enable XSS filter (legacy)       |
| Strict-Transport-Security  | `max-age=31536000; includeSubDomains` | Force HTTPS                      |
| Content-Security-Policy    | `default-src 'self'`                  | Prevent XSS, data injection      |
| Referrer-Policy            | `no-referrer` or `strict-origin`      | Control referrer information     |
| Permissions-Policy         | `geolocation=(), microphone=()`       | Disable unused browser features  |

## Middleware Pattern

```go
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Prevent MIME sniffing
        w.Header().Set("X-Content-Type-Options", "nosniff")

        // Prevent clickjacking
        w.Header().Set("X-Frame-Options", "DENY")

        // XSS protection (legacy but harmless)
        w.Header().Set("X-XSS-Protection", "1; mode=block")

        // Force HTTPS (only if serving over HTTPS)
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

        // CSP: only allow resources from same origin
        w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")

        // Referrer policy
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

        // Permissions policy
        w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

        next.ServeHTTP(w, r)
    })
}
```

## Content Security Policy (CSP)

```go
// Strict CSP (most secure)
csp := "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'"

// CSP with CDN
csp := "default-src 'self'; script-src 'self' https://cdn.example.com; style-src 'self' https://cdn.example.com 'unsafe-inline'"

// CSP for SPA with API
csp := "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self' https://api.example.com"

w.Header().Set("Content-Security-Policy", csp)
```

## CORS Configuration

```go
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")

            // Check if origin is allowed
            allowed := false
            for _, o := range allowedOrigins {
                if o == origin || o == "*" {
                    allowed = true
                    break
                }
            }

            if allowed {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                w.Header().Set("Access-Control-Max-Age", "3600")
            }

            // Handle preflight
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusNoContent)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```
