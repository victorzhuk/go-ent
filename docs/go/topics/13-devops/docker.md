# Docker for Go Applications

Containerizing Go applications with multi-stage builds for minimal, secure, production-ready images.

## Quick Reference

| Pattern | Use Case | Notes |
|---------|----------|-------|
| **Multi-stage build** | Separate build and runtime | Builder stage + minimal runtime |
| **scratch base** | Static binaries | Zero dependencies, smallest size |
| **distroless base** | CGO or timezone/CA certs | ~2MB, includes libc and CA certs |
| **COPY order** | go.mod/go.sum first | Maximize layer caching |
| **.dockerignore** | Exclude unnecessary files | Speeds up builds, reduces context |
| **HEALTHCHECK** | Container health monitoring | Kubernetes liveness/readiness |
| **Non-root user** | Security best practice | Run as nobody or custom user |
| **CGO_ENABLED=0** | Static linking | Required for scratch base |
| **Build flags** | Strip debug info | `-ldflags="-s -w"` reduces size |

## Multi-Stage Build

**Pattern:** Build in full Go environment, run in minimal container.

```dockerfile
# syntax=docker/dockerfile:1

# Builder stage
FROM golang:1.23-alpine AS builder

# Install build dependencies (if needed for CGO)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy dependency files first (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always)" \
    -o /app \
    ./cmd/api

# Runtime stage
FROM scratch

# Copy CA certs and timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary
COPY --from=builder /app /app

# Non-root user (nobody is UID 65534)
USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/app"]
```

**Why multi-stage:**
- Builder stage: ~1GB with full toolchain
- Runtime stage: ~10MB with only binary
- Final image contains zero build artifacts

## Base Image Selection

### scratch (Static Binaries)

**When:** Pure Go code, no CGO, no external dependencies.

```dockerfile
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app /app

USER 65534:65534
ENTRYPOINT ["/app"]
```

**Pros:**
- Smallest possible size (~5-15MB)
- Minimal attack surface
- No shell, no package manager

**Cons:**
- No debugging tools (no shell)
- No timezone data (copy from builder)
- No CA certificates (copy from builder)
- CGO requires libc

### distroless (CGO or Standard Features)

**When:** Need timezone data, CA certs, or CGO.

```dockerfile
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app /app

EXPOSE 8080
ENTRYPOINT ["/app"]
```

**Variants:**
- `static-debian12`: Static binaries, no libc (~2MB base)
- `base-debian12`: Includes glibc for CGO (~20MB base)
- `cc-debian12`: Includes gcc runtime libraries

**Pros:**
- Includes CA certs and timezone data
- Non-root user by default
- Still very small
- Supports CGO (base variant)

**Cons:**
- No shell (use `:debug` tag for debugging)
- Slightly larger than scratch

### alpine (Development/Debugging)

**When:** Need shell access, debugging tools, or quick iteration.

```dockerfile
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app /app

RUN adduser -D -u 1000 appuser
USER appuser

ENTRYPOINT ["/app"]
```

**Pros:**
- Shell and standard tools
- Small size (~5MB base + tools)
- Easy debugging

**Cons:**
- Uses musl libc (compatibility issues)
- Larger attack surface
- Not recommended for production

## Optimization

### .dockerignore

**Pattern:** Exclude unnecessary files from build context.

```dockerignore
# Version control
.git
.gitignore

# Build artifacts
bin/
dist/
*.exe
*.test
*.prof

# IDE files
.vscode/
.idea/
*.swp
*.swo

# Documentation
*.md
docs/

# CI/CD
.github/
.gitlab-ci.yml

# Local development
.env
.env.local
docker-compose.yml
Makefile

# Test files (if not needed in container)
*_test.go
testdata/

# Temporary files
tmp/
*.tmp
*.log
```

**Impact:**
- Faster builds (smaller context)
- No accidentally copied secrets
- Cleaner layer caching

### Layer Caching

**Pattern:** Order COPY commands by change frequency.

```dockerfile
# 1. Copy dependency files (changes rarely)
COPY go.mod go.sum ./
RUN go mod download

# 2. Copy source code (changes frequently)
COPY . .

# 3. Build (only rebuilds if source changed)
RUN go build -o /app ./cmd/api
```

**Impact:**
- Dependency layer cached unless go.mod/go.sum change
- Source changes don't invalidate dependency cache
- Faster iteration during development

### Build Flags

**Pattern:** Strip debug info and optimize for size.

```dockerfile
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /app \
    ./cmd/api
```

**Flags:**
- `CGO_ENABLED=0`: Static linking, no C dependencies
- `-ldflags="-s -w"`: Strip symbol table and debug info (~30% size reduction)
- `-trimpath`: Remove file system paths from binary (reproducible builds)
- `-ldflags="-X main.version=..."`: Inject version at build time

**Size comparison:**
- Default build: ~20MB
- With `-s -w`: ~14MB
- With UPX compression: ~5MB (not recommended, slower startup)

## Security

### Non-Root User

**Pattern:** Run as unprivileged user.

```dockerfile
# Option 1: Use nobody (UID 65534)
FROM scratch
USER 65534:65534

# Option 2: Use distroless nonroot (UID 65532)
FROM gcr.io/distroless/static-debian12:nonroot
# Already runs as nonroot

# Option 3: Create user in alpine
FROM alpine:3.19
RUN adduser -D -u 1000 appuser
USER appuser
```

**Why:**
- Limits damage if container is compromised
- Prevents writing to filesystem
- Required by many Kubernetes security policies

### Read-Only Filesystem

**Pattern:** Mount root filesystem as read-only.

```dockerfile
# In Dockerfile
FROM scratch
USER 65534:65534
ENTRYPOINT ["/app"]
```

```yaml
# In docker-compose.yml or Kubernetes
services:
  api:
    read_only: true
    tmpfs:
      - /tmp:size=10M,mode=1777
```

**Why:**
- Prevents malicious writes
- Forces explicit tmpfs for temporary files
- Catches bugs that assume writable filesystem

### No Secrets in Layers

**Never:**
```dockerfile
# BAD: Secret in layer
COPY .env .
RUN curl -H "Authorization: Bearer ${TOKEN}" ...
```

**Instead:**
```dockerfile
# GOOD: Secrets at runtime
ENV CONFIG_PATH=/secrets/config.json
```

```bash
# Pass at runtime
docker run --env-file .env myapp
docker run -v /secrets:/secrets:ro myapp
```

**Or use BuildKit secrets:**
```dockerfile
# syntax=docker/dockerfile:1
RUN --mount=type=secret,id=token \
    curl -H "Authorization: Bearer $(cat /run/secrets/token)" ...
```

```bash
docker build --secret id=token,src=.token .
```

### Security Scanning

**Pattern:** Scan images for vulnerabilities.

```dockerfile
# Multi-stage with trivy scan
FROM golang:1.23-alpine AS builder
# ... build steps ...

FROM scratch AS runtime
COPY --from=builder /app /app
USER 65534:65534
ENTRYPOINT ["/app"]

# Scan stage
FROM runtime AS scan
COPY --from=aquasec/trivy:latest /usr/local/bin/trivy /trivy
RUN /trivy filesystem --exit-code 1 --no-progress /
```

```bash
# Local scanning
docker build --target runtime -t myapp:latest .
trivy image myapp:latest

# CI/CD scanning
trivy image --severity HIGH,CRITICAL myapp:latest
```

## Healthcheck

### Dockerfile HEALTHCHECK

**Pattern:** Define health check in Dockerfile.

```dockerfile
FROM scratch

COPY --from=builder /app /app

USER 65534:65534

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app", "healthcheck"]

ENTRYPOINT ["/app"]
```

**Application code:**
```go
// main.go
func main() {
    if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
        if err := healthcheck(); err != nil {
            fmt.Fprintf(os.Stderr, "healthcheck failed: %v\n", err)
            os.Exit(1)
        }
        os.Exit(0)
    }

    // Normal startup
    run()
}

func healthcheck() error {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/healthz", nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unhealthy: status %d", resp.StatusCode)
    }
    return nil
}
```

### HTTP Healthcheck Endpoint

**Pattern:** Dedicated /healthz endpoint.

```go
func (a *app) healthHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
    defer cancel()

    // Check database
    if err := a.db.PingContext(ctx); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "unhealthy",
            "error":  "database unreachable",
        })
        return
    }

    // Check dependencies
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
    })
}
```

**Separate readiness:**
```go
// /healthz: Liveness (is process alive?)
func liveness(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    io.WriteString(w, "ok")
}

// /readyz: Readiness (can accept traffic?)
func readiness(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := db.Ping(); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
    }
}
```

## Docker Compose

### Local Development

**Pattern:** Multi-service development environment.

```yaml
version: '3.8'

services:
  api:
    build:
      context: .
      target: builder
    volumes:
      - .:/app
      - go-modules:/go/pkg/mod
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://user:pass@db:5432/mydb?sslmode=disable
      - REDIS_URL=redis://redis:6379/0
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started
    command: go run ./cmd/api

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
      POSTGRES_DB: mydb
    ports:
      - "5432:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data

volumes:
  go-modules:
  postgres-data:
  redis-data:
```

### Production Compose

**Pattern:** Production-ready with health checks and security.

```yaml
version: '3.8'

services:
  api:
    image: myapp:latest
    read_only: true
    tmpfs:
      - /tmp:size=10M
    user: "65534:65534"
    cap_drop:
      - ALL
    security_opt:
      - no-new-privileges:true
    environment:
      - DATABASE_URL=${DATABASE_URL}
    depends_on:
      db:
        condition: service_healthy
    healthcheck:
      test: ["/app", "healthcheck"]
      interval: 30s
      timeout: 3s
      retries: 3
    restart: unless-stopped
```

## Common Mistakes

| Mistake | Problem | Fix |
|---------|---------|-----|
| **Running as root** | Security risk, pod security violations | Use `USER 65534` or distroless:nonroot |
| **COPY . before deps** | Cache invalidation on every code change | `COPY go.mod go.sum` → `go mod download` → `COPY .` |
| **Large images** | Slow pulls, large attack surface | Multi-stage build with scratch/distroless |
| **CGO with scratch** | Binary fails with missing libc | Use distroless/base or `CGO_ENABLED=0` |
| **Hardcoded versions** | Stale dependencies, security issues | Pin base image tags: `golang:1.23-alpine` |
| **No .dockerignore** | Slow builds, accidentally copied secrets | Create .dockerignore with .git, .env, etc. |
| **Secrets in ENV** | Exposed in `docker inspect` | Use secrets management or mounted files |
| **No healthcheck** | Failed containers keep receiving traffic | Add HEALTHCHECK or /healthz endpoint |
| **Missing CA certs** | HTTPS calls fail in scratch | Copy from builder: `/etc/ssl/certs/ca-certificates.crt` |
| **Writable filesystem** | Security risk, allows persistence | Use read_only: true with tmpfs for /tmp |

## See Also

- **[Kubernetes](kubernetes.md)** - Container orchestration and deployments
- **[CI/CD](ci-cd.md)** - Automated builds and image publishing
- **[Linting](linting.md)** - Static analysis and Dockerfile linting (hadolint)
- **[HTTP Server](../05-http-grpc/http-server.md)** - Health check endpoint implementation
- **[Configuration](../09-cli-config/configuration.md)** - Environment variables and secrets management
