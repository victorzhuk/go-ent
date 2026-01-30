# Constraints

- Include input validation at all boundaries (API, CLI, file input)
- Include output encoding for HTML, JSON, SQL to prevent injection
- Include parameterized queries for all database operations
- Include proper password hashing (bcrypt/argon2, never plaintext)
- Include secret management via environment variables or secret stores
- Include constant-time comparisons for security-sensitive data
- Include security headers in all HTTP responses
- Include rate limiting on authentication endpoints
- Exclude hard-coded secrets, API keys, or credentials in code
- Exclude dynamic SQL construction (use parameterized queries)
- Exclude error messages that leak sensitive information
- Exclude insecure default configurations
- Exclude deprecated or weak cryptographic algorithms
- Bound to OWASP Top 10 security best practices
- Follow principle of least privilege for all operations