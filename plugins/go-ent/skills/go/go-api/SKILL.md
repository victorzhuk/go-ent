---
name: go-api
description: "Spec-first API design with OpenAPI/ogen and gRPC/protobuf. Auto-activates for: API design, OpenAPI specs, code generation, protobuf, REST endpoints, gRPC services."
---

# Go API — Spec-First

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
