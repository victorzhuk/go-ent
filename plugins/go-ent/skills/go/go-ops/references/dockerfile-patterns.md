# Docker Patterns for Go Services

## Production-Ready Dockerfile

```dockerfile
# Build stage
FROM golang:1.25.5-trixie AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with optimizations
ARG VERSION=dev
ARG BUILD_TIME
ARG COMMIT_HASH

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.CommitHash=${COMMIT_HASH}" \
    -o /server ./cmd/server

# Production stage - minimal image
FROM gcr.io/distroless/static-debian13:nonroot

# Copy binary from builder
COPY --from=builder /server /server

# Security: non-root user
USER nonroot:nonroot

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/server", "healthcheck"] || exit 1

# Expose port
EXPOSE 8080

# Run binary
ENTRYPOINT ["/server"]
```

## Makefile Targets

```makefile
VERSION ?= $(shell git describe --tags --always)
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_HASH ?= $(shell git rev-parse --short HEAD)

docker-build:
\tdocker build \
\t\t--build-arg VERSION=$(VERSION) \
\t\t--build-arg BUILD_TIME=$(BUILD_TIME) \
\t\t--build-arg COMMIT_HASH=$(COMMIT_HASH) \
\t\t-t app:$(VERSION) \
\t\t-t app:latest \
\t\t.
```

**Pattern**: Multi-stage build with distroless base, non-root user, health checks, build args for versioning.
