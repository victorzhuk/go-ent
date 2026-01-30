# Constraints

- Include multi-stage Docker builds for smaller images
- Include distroless or minimal base images for production
- Include non-root user in containers
- Include health checks in containers and applications
- Include resource limits and requests in Kubernetes
- Include security scanning in CI/CD pipelines (Trivy, Snyk)
- Include structured logging with JSON format in production
- Include observability (metrics, traces, logs) for production
- Include graceful shutdown handling (30s timeout)
- Include proper secret management (never in code)
- Exclude root user in production containers
- Exclude latest tags for base images (use specific versions)
- Exclude sensitive data in logs (passwords, tokens, PII)
- Exclude hardcoded credentials or API keys
- Bound to cloud-native patterns and best practices
- Follow 12-factor app methodology