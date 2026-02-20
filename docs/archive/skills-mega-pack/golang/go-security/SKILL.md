---
name: go-security
description: Security patterns for Go: input validation, authentication, authorization, secrets management, OWASP mitigations
---

# Go Security Patterns

## Input Validation
- Validate ALL input at the boundary (handlers) before passing to business logic
- Use allowlists, not denylists
- Validate types, ranges, lengths, and formats
- Sanitize output for the target context (HTML, SQL, shell)
- Never trust user input — including headers, query params, path params

## Authentication
- Use bcrypt (`golang.org/x/crypto/bcrypt`) for password hashing — cost >= 12
- Implement JWT with `golang-jwt/jwt/v5` — always validate `alg`, `exp`, `iss`
- Use short-lived access tokens (15min) with longer-lived refresh tokens
- Store refresh tokens server-side; revoke on logout
- Implement rate limiting on auth endpoints

## Authorization
- Check permissions at the service layer, not just the handler
- Use RBAC or ABAC patterns with explicit permission checks
- Never rely on client-side authorization checks alone
- Log authorization failures for security monitoring

## Secrets Management
- Never hardcode secrets in source code
- Use environment variables or secret managers (Vault, AWS Secrets Manager)
- Rotate secrets regularly; support zero-downtime rotation
- Never log secrets — use redaction in structured logging
- Use `crypto/rand` for generating tokens, never `math/rand`

## Common Vulnerabilities
- SQL Injection: Always use parameterized queries
- XSS: Sanitize output with `html/template` (auto-escapes)
- CSRF: Use SameSite cookies + CSRF tokens for session-based auth
- Path Traversal: Use `filepath.Clean` and validate paths
- SSRF: Validate and allowlist URLs for server-side requests
- Rate Limiting: Implement at middleware level for all endpoints

## Cryptography
- Use `crypto/aes` with GCM mode for symmetric encryption
- Use `crypto/ecdsa` or `crypto/ed25519` for signing
- Use `crypto/tls` with TLS 1.3 minimum
- Never implement custom cryptographic algorithms
