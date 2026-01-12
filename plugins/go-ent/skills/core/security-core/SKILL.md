---
name: security-core
description: "Security fundamentals and OWASP principles. Auto-activates for: authentication, authorization, input validation, SQL injection, XSS, CSRF, security headers."
version: 1.0.0
---

# Security Core

## OWASP Top 10 (2021)

### 1. Broken Access Control
- Principle of least privilege
- Validate permissions on every request
- Use RBAC (Role-Based Access Control)

### 2. Cryptographic Failures
- TLS for transit, encryption at rest
- Strong algorithms (AES-256, RSA-2048+)
- bcrypt/argon2 for passwords (never plaintext)

### 3. Injection
- Parameterized queries for SQL
- Escape output for XSS
- Validate and sanitize all inputs
- Use ORM/query builders

### 4. Insecure Design
- Threat modeling in design
- Secure by default
- Defense in depth
- Fail securely

### 5. Security Misconfiguration
- Disable debug in production
- Remove default credentials
- Keep dependencies updated
- Secure headers (CSP, HSTS, X-Frame-Options)

### 6. Vulnerable Components
- Track dependencies
- Automated vulnerability scanning
- Update regularly

### 7. Authentication Failures
- Multi-factor authentication
- Strong password policies
- Rate limiting on auth endpoints
- Secure session management

### 8. Software and Data Integrity
- Verify integrity of updates/packages
- CI/CD pipeline security
- Code signing

### 9. Logging and Monitoring Failures
- Log security events
- Monitor for anomalies
- Don't log sensitive data
- Alert on suspicious activity

### 10. Server-Side Request Forgery (SSRF)
- Validate URLs
- Use allowlists for external requests
- Disable unused URL schemas

## Security Checklist

### Input Validation
- [ ] Validate type, length, format, range
- [ ] Allowlist over blocklist
- [ ] Sanitize before processing
- [ ] Reject unexpected input

### Authentication
- [ ] Strong password requirements
- [ ] Rate limiting (prevent brute force)
- [ ] MFA option available
- [ ] Secure password reset flow
- [ ] Session timeout
- [ ] Logout functionality

### Authorization
- [ ] Check permissions on every request
- [ ] Default deny (fail secure)
- [ ] Principle of least privilege
- [ ] No client-side authorization only

### Data Protection
- [ ] HTTPS everywhere
- [ ] Encrypt sensitive data at rest
- [ ] Secure key management
- [ ] No secrets in code/logs
- [ ] Secure cookies (HttpOnly, Secure, SameSite)

## Common Vulnerabilities

| Vulnerability | Example | Prevention |
|---------------|---------|------------|
| SQL Injection | `SELECT * FROM users WHERE id='` + input | Parameterized queries |
| XSS | `<script>alert(1)</script>` | Escape output, CSP |
| CSRF | Forged form submission | CSRF tokens |
| Path Traversal | `../../etc/passwd` | Validate paths, use allowlist |
| Command Injection | `; rm -rf /` | Avoid shell, validate input |

## Security Headers

```
Content-Security-Policy: default-src 'self'
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Strict-Transport-Security: max-age=31536000
```

## Threat Modeling

**STRIDE Framework**:
- **S**poofing - Fake identity
- **T**ampering - Modify data
- **R**epudiation - Deny actions
- **I**nformation Disclosure - Leak data
- **D**enial of Service - Unavailable
- **E**levation of Privilege - Unauthorized access
