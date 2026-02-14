---
name: docker-devops
description: Docker, Kubernetes, CI/CD pipelines, infrastructure as code, and deployment best practices
---

# Docker & DevOps

## Dockerfile Best Practices
```dockerfile
# Multi-stage build
FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile

FROM node:22-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN pnpm build

FROM node:22-alpine AS runtime
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=build /app/dist ./dist
COPY --from=build /app/node_modules ./node_modules
USER app
EXPOSE 3000
CMD ["node", "dist/main.js"]
```

## Docker Compose (Development)
```yaml
services:
  app:
    build: .
    ports: ["3000:3000"]
    depends_on:
      postgres: { condition: service_healthy }
    environment:
      DATABASE_URL: postgres://user:pass@postgres:5432/app
  postgres:
    image: postgres:16-alpine
    healthcheck:
      test: pg_isready -U user
      interval: 5s
    volumes: [pgdata:/var/lib/postgresql/data]
volumes:
  pgdata:
```

## Kubernetes Essentials
- Use Deployments for stateless apps
- Use StatefulSets for databases (prefer managed databases)
- Set resource requests AND limits
- Use liveness and readiness probes
- Use ConfigMaps for config, Secrets for credentials
- Use HPA (Horizontal Pod Autoscaler) for auto-scaling
- Use NetworkPolicies for pod-to-pod security

## CI/CD Pipeline
```yaml
# GitHub Actions example
on:
  push: { branches: [main] }
  pull_request: { branches: [main] }

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: make test
      - run: make lint
  deploy:
    needs: test
    if: github.ref == 'refs/heads/main'
    steps:
      - run: make deploy
```

## Security
- Never store secrets in images or Dockerfiles
- Use non-root users in containers
- Scan images for vulnerabilities (Trivy, Snyk)
- Use read-only file systems where possible
- Pin base image versions (not `latest`)
