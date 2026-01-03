---
name: go-api
description: "Spec-first API design with OpenAPI/ogen and gRPC/protobuf. Auto-activates for: API design, OpenAPI specs, code generation, protobuf."
---

# Go API Design (2026) - Spec-First

## Approach

**Spec-First Always:**
1. Write OpenAPI 3.1 spec → Generate with ogen → Wrap server
2. Write Proto spec → Generate with buf/protoc → Implement service

## Stack Decision

| Scenario | Transport | Generator |
|----------|-----------|-----------|
| Public API, HTTP/2 | `net/http` | ogen |
| High-load (100k+ RPS) | `fasthttp` | ogen |
| Microservices internal | gRPC | buf + protoc |
| Mixed REST + gRPC | grpc-gateway | buf |

## Project Structure

```
project/
├── api/
│   ├── openapi/
│   │   └── v1/
│   │       ├── openapi.yaml      # Main spec
│   │       ├── paths/            # Path definitions
│   │       └── components/       # Schemas, responses
│   └── proto/
│       └── v1/
│           ├── user.proto
│           └── order.proto
├── gen/                          # Generated (gitignore internals)
│   ├── api/v1/                   # ogen output
│   └── proto/v1/                 # buf output
├── internal/
│   ├── transport/
│   │   ├── http/
│   │   │   ├── server.go         # Wraps ogen server
│   │   │   └── handler.go        # Implements ogen Handler
│   │   └── grpc/
│   │       ├── server.go
│   │       └── handler.go        # Implements proto services
│   └── ...
└── Makefile
```

---

## OpenAPI + ogen

### 1. Write Spec

```yaml
# api/openapi/v1/openapi.yaml
openapi: "3.1.0"
info:
  title: User Service API
  version: "1.0.0"
servers:
  - url: /api/v1

paths:
  /users:
    get:
      operationId: listUsers
      parameters:
        - $ref: "#/components/parameters/Limit"
        - $ref: "#/components/parameters/Offset"
      responses:
        "200":
          description: Users list
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/UsersResponse"
    post:
      operationId: createUser
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/CreateUserRequest"
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
        "422":
          $ref: "#/components/responses/ValidationError"

  /users/{id}:
    get:
      operationId: getUser
      parameters:
        - $ref: "#/components/parameters/UserID"
      responses:
        "200":
          description: User found
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
        "404":
          $ref: "#/components/responses/NotFound"

components:
  schemas:
    User:
      type: object
      required: [id, email, name, createdAt]
      properties:
        id:
          type: string
          format: uuid
        email:
          type: string
          format: email
        name:
          type: string
        createdAt:
          type: string
          format: date-time

    CreateUserRequest:
      type: object
      required: [email, name]
      properties:
        email:
          type: string
          format: email
          maxLength: 255
        name:
          type: string
          minLength: 2
          maxLength: 100

    UsersResponse:
      type: object
      required: [users, total]
      properties:
        users:
          type: array
          items:
            $ref: "#/components/schemas/User"
        total:
          type: integer

    Error:
      type: object
      required: [code, message]
      properties:
        code:
          type: string
        message:
          type: string

  parameters:
    UserID:
      name: id
      in: path
      required: true
      schema:
        type: string
        format: uuid
    Limit:
      name: limit
      in: query
      schema:
        type: integer
        default: 20
        maximum: 100
    Offset:
      name: offset
      in: query
      schema:
        type: integer
        default: 0

  responses:
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"
    ValidationError:
      description: Validation error
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"
```

### 2. Generate with ogen

```makefile
# Makefile
OGEN_VERSION := v1.8.1

.PHONY: gen-api
gen-api:
	go run github.com/ogen-go/ogen/cmd/ogen@$(OGEN_VERSION) \
		--target gen/api/v1 \
		--package apiv1 \
		--clean \
		api/openapi/v1/openapi.yaml
```

### 3. Implement Handler

```go
// internal/transport/http/handler.go
package http

import (
    "context"
    
    apiv1 "github.com/org/app/gen/api/v1"
    "github.com/org/app/internal/usecase"
)

type Handler struct {
    createUserUC usecase.CreateUserUC
    getUserUC    usecase.GetUserUC
    listUsersUC  usecase.ListUsersUC
    log          *slog.Logger
}

func NewHandler(
    createUserUC usecase.CreateUserUC,
    getUserUC usecase.GetUserUC,
    listUsersUC usecase.ListUsersUC,
    log *slog.Logger,
) *Handler {
    return &Handler{
        createUserUC: createUserUC,
        getUserUC:    getUserUC,
        listUsersUC:  listUsersUC,
        log:          log,
    }
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
    
    return &apiv1.User{
        ID:        apiv1.NewOptUUID(resp.ID),
        Email:     resp.Email,
        Name:      resp.Name,
        CreatedAt: resp.CreatedAt,
    }, nil
}

func (h *Handler) GetUser(ctx context.Context, params apiv1.GetUserParams) (apiv1.GetUserRes, error) {
    resp, err := h.getUserUC.Execute(ctx, params.ID)
    if err != nil {
        return h.mapError(err), nil
    }
    
    return &apiv1.User{
        ID:        apiv1.NewOptUUID(resp.ID),
        Email:     resp.Email,
        Name:      resp.Name,
        CreatedAt: resp.CreatedAt,
    }, nil
}

func (h *Handler) ListUsers(ctx context.Context, params apiv1.ListUsersParams) (apiv1.ListUsersRes, error) {
    resp, err := h.listUsersUC.Execute(ctx, usecase.ListUsersReq{
        Limit:  params.Limit.Or(20),
        Offset: params.Offset.Or(0),
    })
    if err != nil {
        return h.mapError(err), nil
    }
    
    users := make([]apiv1.User, len(resp.Users))
    for i, u := range resp.Users {
        users[i] = apiv1.User{
            ID:        apiv1.NewOptUUID(u.ID),
            Email:     u.Email,
            Name:      u.Name,
            CreatedAt: u.CreatedAt,
        }
    }
    
    return &apiv1.UsersResponse{Users: users, Total: resp.Total}, nil
}

func (h *Handler) mapError(err error) apiv1.ErrorStatusCode {
    switch {
    case errors.Is(err, contract.ErrNotFound):
        return &apiv1.ErrorStatusCode{
            StatusCode: 404,
            Response:   apiv1.Error{Code: "not_found", Message: "resource not found"},
        }
    case errors.Is(err, contract.ErrConflict):
        return &apiv1.ErrorStatusCode{
            StatusCode: 409,
            Response:   apiv1.Error{Code: "conflict", Message: "already exists"},
        }
    default:
        h.log.Error("internal error", "error", err)
        return &apiv1.ErrorStatusCode{
            StatusCode: 500,
            Response:   apiv1.Error{Code: "internal_error", Message: "something went wrong"},
        }
    }
}
```

### 4a. Wrap with net/http (Default)

```go
// internal/transport/http/server.go
package http

import (
    "net/http"
    
    apiv1 "github.com/org/app/gen/api/v1"
)

type Server struct {
    srv *http.Server
    log *slog.Logger
}

func NewServer(cfg *Config, handler *Handler, log *slog.Logger) (*Server, error) {
    ogenSrv, err := apiv1.NewServer(handler,
        apiv1.WithPathPrefix("/api/v1"),
    )
    if err != nil {
        return nil, fmt.Errorf("create ogen server: %w", err)
    }
    
    mux := http.NewServeMux()
    mux.HandleFunc("GET /healthz", healthLiveness)
    mux.HandleFunc("GET /readyz", healthReadiness)
    mux.Handle("GET /metrics", promhttp.Handler())
    mux.Handle("/api/v1/", withMiddleware(ogenSrv, log))
    
    return &Server{
        srv: &http.Server{
            Addr:              cfg.Addr,
            Handler:           mux,
            ReadTimeout:       cfg.ReadTimeout,
            ReadHeaderTimeout: cfg.ReadHeaderTimeout,
            WriteTimeout:      cfg.WriteTimeout,
            IdleTimeout:       cfg.IdleTimeout,
        },
        log: log,
    }, nil
}

func withMiddleware(h http.Handler, log *slog.Logger) http.Handler {
    h = recoveryMiddleware(h, log)
    h = loggingMiddleware(h, log)
    h = requestIDMiddleware(h)
    return h
}

func (s *Server) Start() error {
    s.log.Info("http server starting", "addr", s.srv.Addr)
    if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
        return fmt.Errorf("listen: %w", err)
    }
    return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
    return s.srv.Shutdown(ctx)
}
```

### 4b. Wrap with fasthttp (High-Load)

```go
// internal/transport/fasthttp/server.go
package fasthttp

import (
    "github.com/valyala/fasthttp"
    "github.com/valyala/fasthttp/fasthttpadaptor"
    
    apiv1 "github.com/org/app/gen/api/v1"
)

type Server struct {
    srv  *fasthttp.Server
    addr string
    log  *slog.Logger
}

func NewServer(cfg *Config, handler *Handler, log *slog.Logger) (*Server, error) {
    ogenSrv, err := apiv1.NewServer(handler, apiv1.WithPathPrefix("/api/v1"))
    if err != nil {
        return nil, fmt.Errorf("create ogen server: %w", err)
    }
    
    // Adapt ogen (net/http) to fasthttp
    ogenHandler := fasthttpadaptor.NewFastHTTPHandler(ogenSrv)
    
    requestHandler := func(ctx *fasthttp.RequestCtx) {
        path := string(ctx.Path())
        
        switch path {
        case "/healthz":
            ctx.SetStatusCode(fasthttp.StatusOK)
            ctx.SetBodyString(`{"status":"ok"}`)
        case "/readyz":
            ctx.SetStatusCode(fasthttp.StatusOK)
            ctx.SetBodyString(`{"status":"ready"}`)
        default:
            ogenHandler(ctx)
        }
    }
    
    return &Server{
        srv: &fasthttp.Server{
            Handler:            requestHandler,
            ReadTimeout:        cfg.ReadTimeout,
            WriteTimeout:       cfg.WriteTimeout,
            MaxRequestBodySize: cfg.MaxRequestBodySize,
        },
        addr: cfg.Addr,
        log:  log,
    }, nil
}

func (s *Server) Start() error {
    s.log.Info("fasthttp server starting", "addr", s.addr)
    return s.srv.ListenAndServe(s.addr)
}

func (s *Server) Shutdown() error {
    return s.srv.Shutdown()
}
```

---

## gRPC + Protobuf

### 1. Write Proto Spec

```protobuf
// api/proto/v1/user.proto
syntax = "proto3";

package user.v1;

option go_package = "github.com/org/app/gen/proto/v1;userv1";

import "google/protobuf/timestamp.proto";
import "buf/validate/validate.proto";

service UserService {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc GetUser(GetUserRequest) returns (User);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}

message CreateUserRequest {
  string email = 1 [(buf.validate.field).string.email = true];
  string name = 2 [(buf.validate.field).string = {min_len: 2, max_len: 100}];
}

message CreateUserResponse {
  string id = 1;
}

message GetUserRequest {
  string id = 1 [(buf.validate.field).string.uuid = true];
}

message User {
  string id = 1;
  string email = 2;
  string name = 3;
  google.protobuf.Timestamp created_at = 4;
}

message ListUsersRequest {
  int32 limit = 1;
  int32 offset = 2;
}

message ListUsersResponse {
  repeated User users = 1;
  int32 total = 2;
}
```

### 2. Generate with buf

```yaml
# buf.gen.yaml
version: v2
plugins:
  - remote: buf.build/protocolbuffers/go
    out: gen/proto
    opt: paths=source_relative
  - remote: buf.build/grpc/go
    out: gen/proto
    opt: paths=source_relative
```

```makefile
.PHONY: gen-proto
gen-proto:
	buf generate
```

### 3. Implement gRPC Server

```go
// internal/transport/grpc/handler.go
package grpc

import (
    "context"
    
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"
    
    userv1 "github.com/org/app/gen/proto/v1"
)

type UserHandler struct {
    userv1.UnimplementedUserServiceServer
    createUC usecase.CreateUserUC
    getUserC usecase.GetUserUC
    listUC   usecase.ListUsersUC
    log      *slog.Logger
}

func (h *UserHandler) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
    resp, err := h.createUC.Execute(ctx, usecase.CreateUserReq{
        Email: req.Email,
        Name:  req.Name,
    })
    if err != nil {
        return nil, h.mapError(err)
    }
    return &userv1.CreateUserResponse{Id: resp.ID.String()}, nil
}

func (h *UserHandler) mapError(err error) error {
    switch {
    case errors.Is(err, contract.ErrNotFound):
        return status.Error(codes.NotFound, "not found")
    case errors.Is(err, contract.ErrConflict):
        return status.Error(codes.AlreadyExists, "already exists")
    default:
        h.log.Error("internal error", "error", err)
        return status.Error(codes.Internal, "internal error")
    }
}
```

---

## Makefile

```makefile
OGEN_VERSION := v1.8.1

.PHONY: gen
gen: gen-api gen-proto

.PHONY: gen-api
gen-api:
	go run github.com/ogen-go/ogen/cmd/ogen@$(OGEN_VERSION) \
		--target gen/api/v1 \
		--package apiv1 \
		--clean \
		api/openapi/v1/openapi.yaml

.PHONY: gen-proto
gen-proto:
	buf generate

.PHONY: lint-api
lint-api:
	npx @redocly/cli lint api/openapi/v1/openapi.yaml

.PHONY: lint-proto
lint-proto:
	buf lint
```
