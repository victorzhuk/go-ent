---
name: security-best-practices
description: Application security, OWASP Top 10, authentication, authorization, input validation, and secrets management
---

# Security Best Practices

## OWASP Top 10 Mitigations
1. **Broken Access Control**: Check permissions server-side on every request
2. **Cryptographic Failures**: Use TLS 1.3, hash passwords with bcrypt/argon2, encrypt at rest
3. **Injection**: Parameterized queries, input validation, output encoding
4. **Insecure Design**: Threat modeling, security requirements, abuse cases
5. **Security Misconfiguration**: Harden defaults, remove unused features, automate config
6. **Vulnerable Components**: Dependency scanning, automated updates, SBOM
7. **Auth Failures**: MFA, rate limiting, session management, secure password policies
8. **Data Integrity Failures**: Verify software updates, use signed artifacts
9. **Logging Failures**: Log security events, monitor for anomalies, protect logs
10. **SSRF**: Validate URLs, allowlist destinations, block internal network access

## Input Validation
- Validate on the server side — client validation is UX, not security
- Use allowlists over denylists
- Validate type, length, range, format
- Reject invalid input with clear error messages (no internal details)
- Sanitize output for the target context (HTML, SQL, JSON, shell)

## Authentication
- Use bcrypt (cost >= 12) or argon2id for password hashing
- Implement rate limiting on login endpoints
- Use short-lived JWTs (15 min) + refresh tokens
- Implement account lockout after failed attempts
- Use secure session configuration (HttpOnly, Secure, SameSite)

## Secrets Management
- Never commit secrets to version control
- Use environment variables or secret managers
- Rotate secrets regularly
- Use different secrets per environment
- Audit secret access

## Headers
```
Content-Security-Policy: default-src 'self'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-XSS-Protection: 0 (use CSP instead)
```
