# gRPC Service with Protobuf

## Protobuf Service Definition

```protobuf
syntax = "proto3";

package user.v1;

option go_package = "myapp/gen/proto/user/v1;userv1";

service UserService {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}

message CreateUserRequest {
  string email = 1;
  string name = 2;
}

message CreateUserResponse {
  string id = 1;
  string email = 2;
  string name = 3;
}

message GetUserRequest {
  string id = 1;
}

message GetUserResponse {
  string id = 1;
  string email = 2;
  string name = 3;
}

message ListUsersRequest {
  int32 page = 1;
  int32 per_page = 2;
}

message ListUsersResponse {
  repeated User users = 1;
  int32 total = 2;
}

message User {
  string id = 1;
  string email = 2;
  string name = 3;
}
```

## gRPC Handler Implementation

```go
package grpchandler

import (
    "context"
    "errors"

    userv1 "myapp/gen/proto/user/v1"
    "myapp/internal/contract"
    "myapp/internal/usecase"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type UserHandler struct {
    userv1.UnimplementedUserServiceServer
    createUC usecase.CreateUserUC
    getUC    usecase.GetUserUC
}

func NewUserHandler(createUC usecase.CreateUserUC, getUC usecase.GetUserUC) *UserHandler {
    return &UserHandler{
        createUC: createUC,
        getUC:    getUC,
    }
}

func (h *UserHandler) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
    resp, err := h.createUC.Execute(ctx, usecase.CreateUserReq{
        Email: req.Email,
        Name:  req.Name,
    })
    if err != nil {
        return nil, h.mapError(err)
    }

    return &userv1.CreateUserResponse{
        Id:    resp.ID.String(),
        Email: req.Email,
        Name:  req.Name,
    }, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
    id, err := uuid.Parse(req.Id)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
    }

    resp, err := h.getUC.Execute(ctx, usecase.GetUserReq{ID: id})
    if err != nil {
        return nil, h.mapError(err)
    }

    return &userv1.GetUserResponse{
        Id:    resp.ID.String(),
        Email: resp.Email,
        Name:  resp.Name,
    }, nil
}

// mapError maps domain errors to gRPC status codes
func (h *UserHandler) mapError(err error) error {
    switch {
    case errors.Is(err, contract.ErrNotFound):
        return status.Error(codes.NotFound, "user not found")
    case errors.Is(err, contract.ErrConflict):
        return status.Error(codes.AlreadyExists, "user already exists")
    case errors.Is(err, contract.ErrValidation):
        return status.Error(codes.InvalidArgument, "invalid input")
    default:
        return status.Error(codes.Internal, "internal error")
    }
}
```

**Pattern**: Map domain errors to gRPC status codes, use UnimplementedServer for forward compatibility, validate input at handler boundary.
