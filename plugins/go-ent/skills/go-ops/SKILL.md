---
name: go-ops
description: "DevOps patterns with Docker, Kubernetes, Helm, CI/CD. Auto-activates for: deployment, containerization, orchestration, CI/CD pipelines, infrastructure."
---

# Go DevOps

## Stack

- Docker 27 / Podman 5
- Kubernetes 1.31
- Helm 3.16
- GitHub Actions / GitLab CI

## Dockerfile

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /server /server
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
```

## Docker Compose

```yaml
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_DSN=postgres://user:pass@db:5432/app?sslmode=disable
    depends_on:
      db:
        condition: service_healthy

  db:
    image: postgres:17-alpine
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user -d app"]
```

## Health Checks

```go
func (h *HealthController) Liveness(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
}

func (h *HealthController) Readiness(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()
    if err := h.db.Ping(ctx); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
}
```

## Metrics

```go
var requestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{Name: "http_requests_total"},
    []string{"method", "path", "status"},
)
mux.Handle("/metrics", promhttp.Handler())
```

## GitHub Actions

```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: go test -race ./...
      - run: golangci-lint run
```

## Makefile

```makefile
build:
	go build -ldflags="-s -w" -o bin/server ./cmd/server

test:
	go test -race -cover ./...

lint:
	golangci-lint run

docker:
	docker build -t app:$(VERSION) .
```
