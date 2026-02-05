# OpenAPI

Code generation with ogen for type-safe OpenAPI clients and servers.

## Quick Reference

| Task | Command/Code |
|------|--------------|
| Install | `go install github.com/ogen-go/ogen/cmd/ogen@latest` |
| Generate | `ogen --target ./internal/api --clean openapi.yaml` |
| Generate with config | `ogen --config ogen.yaml openapi.yaml` |
| Custom package | `ogen --package api --target ./internal/api openapi.yaml` |
| Skip unimplemented | Set `skipUnimplemented: true` in ogen.yaml |
| Generate tests | Set `generateTests: true` in ogen.yaml |
| Add middleware | Implement `ogen.Middleware` interface |
| Map validation error | Check `*ogen.ErrorStatusCode` in error handler |
| Custom error type | Define in `components/responses` with status codes |
| Use generated client | `client, _ := api.NewClient("http://localhost:8080")` |
| Client with retry | Wrap `http.Transport` with custom `RoundTripper` |

## OpenAPI Spec

```yaml
openapi: 3.0.0
info:
  title: User API
  version: 1.0.0
paths:
  /users:
    post:
      operationId: createUser
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      required: [id, name]
      properties:
        id: {type: string}
        name: {type: string}
```

## Ogen Configuration

**ogen.yaml:**

```yaml
package: api
target: ./internal/api
generate:
  - client
  - server
generateTests: true
skipUnimplemented: false
skipTestRegex: "^(List|Get).*"
convenient_errors: true
allow_remote: false
```

**Build integration:**

```go
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --config ogen.yaml openapi.yaml
```

## Implementation

```go
// Generated interface to implement
type Handler interface {
    CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
}

// Your implementation
type userHandler struct {
    svc *user.Service
}

func (h *userHandler) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    user, err := h.svc.Create(ctx, req.Name)
    if err != nil {
        return nil, err
    }
    return &User{ID: user.ID, Name: user.Name}, nil
}
```

## Error Handling

**Define errors in spec:**

```yaml
components:
  responses:
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    ValidationError:
      description: Validation failed
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ValidationError'
  schemas:
    Error:
      type: object
      required: [code, message]
      properties:
        code: {type: string}
        message: {type: string}
    ValidationError:
      type: object
      required: [code, message, fields]
      properties:
        code: {type: string}
        message: {type: string}
        fields:
          type: array
          items:
            type: object
            required: [field, error]
            properties:
              field: {type: string}
              error: {type: string}
```

**Map errors in handler:**

```go
func (h *userHandler) GetUser(ctx context.Context, params GetUserParams) (*User, error) {
    user, err := h.svc.GetByID(ctx, params.ID)
    if err != nil {
        if errors.Is(err, domain.ErrNotFound) {
            return nil, &NotFoundError{Code: "USER_NOT_FOUND", Message: "user not found"}
        }
        return nil, fmt.Errorf("get user: %w", err)
    }
    return toAPIUser(user), nil
}
```

**Custom error handler:**

```go
type errorHandler struct{}

func (e *errorHandler) NewError(ctx context.Context, err error) *ErrorStatusCode {
    var validationErr *ogen.ValidationError
    if errors.As(err, &validationErr) {
        return &ErrorStatusCode{
            StatusCode: 400,
            Response: Error{Code: "VALIDATION_ERROR", Message: validationErr.Error()},
        }
    }
    return &ErrorStatusCode{
        StatusCode: 500,
        Response: Error{Code: "INTERNAL_ERROR", Message: "internal server error"},
    }
}

srv, _ := api.NewServer(handler, api.WithErrorHandler(&errorHandler{}))
```

## Validation

**Validation in spec:**

```yaml
components:
  schemas:
    CreateUserRequest:
      type: object
      required: [email, age]
      properties:
        email:
          type: string
          format: email
          minLength: 5
          maxLength: 100
        age:
          type: integer
          minimum: 18
          maximum: 120
        tags:
          type: array
          minItems: 1
          maxItems: 10
          items: {type: string}
```

**Ogen auto-validates:**

```go
// ogen validates before calling handler:
// - required fields present
// - string length constraints
// - integer min/max
// - array item counts
// - format (email, uuid, date-time)

func (h *userHandler) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // req is already validated by ogen
    // custom business validation only
    if h.isEmailBanned(req.Email) {
        return nil, &ValidationErrorResponse{
            Code: "BANNED_EMAIL",
            Message: "email domain is banned",
        }
    }
    // ...
}
```

**Custom validators:**

```go
type validatingHandler struct {
    Handler
    v *validator.Validate
}

func (h *validatingHandler) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    if err := h.v.Struct(req); err != nil {
        return nil, &ValidationErrorResponse{
            Code: "VALIDATION_ERROR",
            Message: err.Error(),
        }
    }
    return h.Handler.CreateUser(ctx, req)
}
```

## Middleware

**Logging middleware:**

```go
type loggingMiddleware struct {
    log *slog.Logger
}

func (m *loggingMiddleware) Handle(req ogen.Request, next ogen.Next) (ogen.Response, error) {
    start := time.Now()

    resp, err := next(req)

    m.log.Info("api request",
        "method", req.OperationID,
        "duration", time.Since(start),
        "error", err != nil,
    )

    return resp, err
}
```

**Auth middleware:**

```go
type authMiddleware struct {
    tokens *token.Verifier
}

func (m *authMiddleware) Handle(req ogen.Request, next ogen.Next) (ogen.Response, error) {
    auth := req.HTTPRequest.Header.Get("Authorization")
    if auth == "" {
        return nil, &UnauthorizedError{Code: "NO_AUTH", Message: "missing authorization"}
    }

    claims, err := m.tokens.Verify(strings.TrimPrefix(auth, "Bearer "))
    if err != nil {
        return nil, &UnauthorizedError{Code: "INVALID_TOKEN", Message: "invalid token"}
    }

    ctx := context.WithValue(req.Context(), "claims", claims)
    return next(req.WithContext(ctx))
}
```

**Request ID middleware:**

```go
type requestIDMiddleware struct{}

func (m *requestIDMiddleware) Handle(req ogen.Request, next ogen.Next) (ogen.Response, error) {
    rid := req.HTTPRequest.Header.Get("X-Request-ID")
    if rid == "" {
        rid = uuid.NewString()
    }

    ctx := context.WithValue(req.Context(), "request_id", rid)
    resp, err := next(req.WithContext(ctx))

    if resp != nil {
        resp.HTTPResponse.Header.Set("X-Request-ID", rid)
    }

    return resp, err
}
```

**Wire middleware:**

```go
srv, _ := api.NewServer(
    handler,
    api.WithMiddleware(
        &requestIDMiddleware{},
        &loggingMiddleware{log: logger},
        &authMiddleware{tokens: verifier},
    ),
)
```

## Client Usage

**Generated client:**

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/org/project/internal/api"
)

func main() {
    client, err := api.NewClient("http://localhost:8080")
    if err != nil {
        log.Fatal(err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    user, err := client.CreateUser(ctx, &api.CreateUserRequest{
        Email: "user@example.com",
        Age: 25,
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("created user: %s", user.ID)
}
```

**Custom transport with retry:**

```go
type retryTransport struct {
    base    http.RoundTripper
    maxRetries int
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    for i := 0; i < t.maxRetries; i++ {
        resp, err := t.base.RoundTrip(req)
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }
        if err != nil && i == t.maxRetries-1 {
            return nil, err
        }
        time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
    }
    return nil, fmt.Errorf("max retries exceeded")
}

client, _ := api.NewClient(
    "http://localhost:8080",
    api.WithClient(&http.Client{
        Transport: &retryTransport{
            base: http.DefaultTransport,
            maxRetries: 3,
        },
        Timeout: 10 * time.Second,
    }),
)
```

**Custom headers:**

```go
type headerTransport struct {
    base http.RoundTripper
    headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    for k, v := range t.headers {
        req.Header.Set(k, v)
    }
    return t.base.RoundTrip(req)
}

client, _ := api.NewClient(
    "http://localhost:8080",
    api.WithClient(&http.Client{
        Transport: &headerTransport{
            base: http.DefaultTransport,
            headers: map[string]string{
                "X-API-Key": "secret",
                "User-Agent": "my-client/1.0",
            },
        },
    }),
)
```

## Common Mistakes

| Mistake | Problem | Solution |
|---------|---------|----------|
| Missing `operationId` | ogen cannot generate method names | Add unique `operationId` to each endpoint |
| Wrong `$ref` paths | Generation fails with reference errors | Use `#/components/schemas/Name` format |
| Inconsistent error responses | Different error formats per endpoint | Define shared error schemas in components |
| Not versioning API | Breaking changes break clients | Use `/v1/` prefix or header versioning |
| Ignoring validation | Runtime errors for invalid input | Add constraints to schema (required, min, max) |
| Inline schemas | Code duplication, hard to maintain | Extract to `components/schemas` |
| No response descriptions | Generated docs unclear | Add meaningful `description` fields |
| Using `additionalProperties: true` | Type safety lost | Define explicit properties |
| Forgetting auth in spec | No auth validation generated | Add `securitySchemes` and `security` |
| Not testing generated code | Runtime errors in production | Set `generateTests: true`, run tests |

## See Also

- [HTTP Server](./http-server.md)
- [HTTP Client](./http-client.md)
- [gRPC](./grpc.md)
- [Input Validation](../11-security/input-validation.md)
- [Tracing](../07-observability/tracing.md)
- [Authentication](../11-security/authentication.md)
