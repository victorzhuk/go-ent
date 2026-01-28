---
name: go-api
description: "Spec-first API design with OpenAPI/ogen and gRPC/protobuf. Auto-activates for: API design, OpenAPI specs, code generation, protobuf, REST endpoints, gRPC services."
version: "2.0.0"
author: "go-ent"
license: "MIT"
compatibility:
  claude_code: ">=1.0"
  opencode: ">=0.1"
tags: ["go", "api", "http", "openapi", "ogen"]
quality_score: 90
category: "go"
---

<triggers>
keywords:
  - "go api"
  - "rest"
  - "grpc"
  - "ogen"
  - "protobuf"
file_pattern: "**/api/*.go"
weight: 0.8
</triggers>

# Go API — Spec-First

<role>
Expert Go API designer specializing in REST and gRPC services. Focus on spec-first development, code generation, proper error handling, and transport layer separation.
</role>

<instructions>

## Approach

1. **REST**: Write OpenAPI spec → Generate with ogen → Wrap server
2. **gRPC**: Write Proto spec → Generate with protoc → Implement service

## Stack Decision

| Scenario | Transport | Generator |
|----------|-----------|-----------|
| Public API | `net/http` | ogen |
| High-load (100k+ RPS) | `fasthttp` | ogen |
| Microservices | gRPC | protoc |

## Project Structure

```
api/
├── openapi/v1/
│   └── openapi.yaml
└── proto/v1/
    └── user.proto
gen/
├── api/v1/       # ogen output
└── proto/v1/     # protoc output
```

## Generate Commands

```makefile
gen-api:
	go run github.com/ogen-go/ogen/cmd/ogen@latest \
		--target gen/api/v1 --package apiv1 --clean \
		api/openapi/v1/openapi.yaml

gen-proto:
	protoc -I api/proto/v1 \
		--go_out=gen/proto/v1 \
		api/proto/v1/user.proto
```

## REST Handler

```go
type Handler struct {
    createUserUC usecase.CreateUserUC
    log          *slog.Logger
}

var _ apiv1.Handler = (*Handler)(nil)

func (h *Handler) CreateUser(ctx context.Context, req *apiv1.CreateUserRequest) (apiv1.CreateUserRes, error) {
    resp, err := h.createUserUC.Execute(ctx, usecase.CreateUserReq{
        Email: req.Email,
        Name:  req.Name,
    })
    if err != nil {
        return h.mapError(err), nil
    }
    return &apiv1.User{ID: apiv1.NewOptUUID(resp.ID)}, nil
}

func (h *Handler) mapError(err error) apiv1.ErrorStatusCode {
    switch {
    case errors.Is(err, contract.ErrNotFound):
        return &apiv1.ErrorStatusCode{StatusCode: 404, Response: apiv1.Error{Code: "not_found"}}
    case errors.Is(err, contract.ErrConflict):
        return &apiv1.ErrorStatusCode{StatusCode: 409, Response: apiv1.Error{Code: "conflict"}}
    default:
        h.log.Error("internal error", "error", err)
        return &apiv1.ErrorStatusCode{StatusCode: 500, Response: apiv1.Error{Code: "internal_error"}}
    }
}
```

**Pattern**: Map domain errors to HTTP status codes at transport layer.

## gRPC Handler

```go
type UserHandler struct {
    userv1.UnimplementedUserServiceServer
    createUC usecase.CreateUserUC
}

func (h *UserHandler) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
    resp, err := h.createUC.Execute(ctx, usecase.CreateUserReq{Email: req.Email})
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    return &userv1.CreateUserResponse{Id: resp.ID.String()}, nil
}
```

## Context7

```
mcp__context7__resolve(library: "ogen")
mcp__context7__resolve(library: "protoc")
mcp__context7__resolve(library: "grpc-go")
```

</instructions>

<constraints>
- Include spec-first design (OpenAPI for REST, protobuf for gRPC)
- Include code generation with ogen/protoc
- Include transport layer with zero business logic
- Include proper error mapping (domain errors → HTTP status codes / gRPC codes)
- Include request/response DTOs with validation
- Include context propagation throughout handlers
- Include proper logging and metrics in handlers
- Exclude business logic in transport layer (delegate to usecases)
- Exclude direct database access from handlers
- Exclude unvalidated input processing
- Exclude manual JSON marshaling/unmarshaling (use generated types)
- Exclude breaking changes to public API without versioning
- Bound to transport layer only, call usecases for business logic
- Follow REST conventions for HTTP (status codes, resource naming)
- Follow gRPC best practices (streaming, deadlines, metadata)
</constraints>

<edge_cases>
If API requirements are unclear: Ask about transport type (REST/gRPC), target consumers, performance requirements, and versioning strategy.

If spec is ambiguous or incomplete: Request clarification on endpoints, request/response structures, error handling, and authentication requirements.

If performance concerns exist: Delegate to go-perf skill for optimization strategies, caching, and high-load patterns.

If code implementation details are needed: Delegate to go-code skill for Go-specific handler implementation patterns.

If architecture guidance is needed: Delegate to go-arch skill for transport layer integration with clean architecture.

If database integration is required: Delegate to go-db skill for repository patterns behind the API layer.

If authentication/authorization is needed: Delegate to go-sec skill for security patterns and middleware.

If validation requirements are complex: Suggest using validation middleware or domain-level validation.

If versioning strategy is needed: Recommend URL versioning (/v1/, /v2/) or header-based versioning.
</edge_cases>

<examples>
<example>
<input>Implement REST handler with validation</input>
<output>
```go
type Handler struct {
    createUserUC contract.CreateUserUC
    validator    *validator.Validate
    logger       *slog.Logger
}

type createUserReq struct {
    Email string `json:"email" validate:"required,email"`
    Name  string `json:"name" validate:"required,min=2,max=100"`
}

type createUserResp struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    var req createUserReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    
    if err := h.validator.Struct(req); err != nil {
        respondError(w, http.StatusBadRequest, "validation failed")
        return
    }
    
    resp, err := h.createUserUC.Execute(ctx, usecase.CreateUserReq{
        Email: req.Email,
        Name:  req.Name,
    })
    if err != nil {
        if errors.Is(err, contract.ErrConflict) {
            respondError(w, http.StatusConflict, "user already exists")
            return
        }
        h.logger.ErrorContext(ctx, "create user failed", "error", err)
        respondError(w, http.StatusInternalServerError, "internal error")
        return
    }
    
    respondJSON(w, http.StatusCreated, createUserResp{
        ID:        resp.ID.String(),
        Email:     req.Email,
        CreatedAt: time.Now(),
    })
}
```

**Pattern**: Private DTOs with validation tags, map domain errors to HTTP status codes, structured logging, never leak internal errors.
</output>
</example>

For additional API implementation examples, see:
- `references/rest-handler.md` - REST API handler with OpenAPI
- `references/grpc-handler.md` - gRPC service with protobuf
- `references/middleware.md` - Logging and metrics middleware
</examples>

<output_format>
Provide API design and implementation guidance with the following structure:

1. **Spec-First Approach**: OpenAPI for REST, protobuf for gRPC, code generation
2. **Transport Layer**: Zero business logic, error mapping, request/response handling
3. **Handler Implementation**: Clean delegation to usecases, context propagation
4. **Error Handling**: Domain errors mapped to HTTP status codes or gRPC status
5. **Middleware**: Logging, metrics, request ID, authentication, validation
6. **Code Generation**: ogen/protoc commands, generated type usage
7. **Examples**: Complete OpenAPI specs, protobuf definitions, handler implementations
8. **Best Practices**: REST conventions, gRPC patterns, versioning strategy

Focus on production-ready API patterns that balance usability, performance, and maintainability.
</output_format>
