---
name: go-api
description: Spec-first API design with Go stdlib net/http, OpenAPI/ogen, gRPC/protobuf, and middleware patterns
---

# Go API Development

## HTTP Server (stdlib 1.22+)
```go
mux := http.NewServeMux()
// Enhanced routing with method + path patterns
mux.HandleFunc("GET /api/v1/users/{id}", getUser)
mux.HandleFunc("POST /api/v1/users", createUser)
mux.HandleFunc("GET /api/v1/users/{id}/orders/{orderID}", getOrder)

// Path value extraction
func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    // ...
}
```

## Spec-First API Design
1. Write OpenAPI 3.1 spec FIRST in `api/openapi.yaml`
2. Generate server stubs with `ogen` or `oapi-codegen`
3. Implement interfaces — never hand-write HTTP routing for API endpoints
4. Keep spec as single source of truth; regenerate on spec changes
5. Validate requests/responses against spec in tests

## Middleware Pattern
```go
func Chain(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
    for i := len(mw) - 1; i >= 0; i-- {
        h = mw[i](h)
    }
    return h
}

// Standard middleware: logging, recovery, CORS, auth, request-id, rate-limit
handler := Chain(mux, Recovery, RequestID, Logger, CORS, RateLimit, Auth)
```

## JSON Handling
- Use `encoding/json` for standard cases; `json/v2` when available
- Define strict request/response structs with json tags
- Use `omitempty` deliberately — empty string vs absent is meaningful
- Validate input with struct tags or explicit validation, not in handlers
- Return consistent error responses: `{"error": {"code": "...", "message": "..."}}`

## gRPC
- Define services in `.proto` files under `api/proto/`
- Generate with `buf` (preferred over raw `protoc`)
- Use interceptors for cross-cutting concerns (auth, logging, tracing)
- Implement health checks with `grpc.health.v1`
- Use `connect-go` for gRPC-compatible HTTP APIs

## API Best Practices
- Version APIs: `/api/v1/...`
- Use proper HTTP status codes (201 Created, 204 No Content, 409 Conflict)
- Paginate list endpoints with cursor-based pagination
- Return RFC 7807 Problem Details for errors
- Implement graceful shutdown with signal handling
- Use request timeouts via `context.WithTimeout`
