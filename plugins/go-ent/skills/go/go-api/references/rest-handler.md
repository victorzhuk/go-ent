# REST API Handler with OpenAPI

## Complete REST Handler Implementation

```go
package handler

import (
    "context"
    "errors"
    "log/slog"

    apiv1 "myapp/gen/api/v1"
    "myapp/internal/contract"
    "myapp/internal/usecase"
)

type Handler struct {
    createUserUC usecase.CreateUserUC
    getUserUC    usecase.GetUserUC
    log          *slog.Logger
}

var _ apiv1.Handler = (*Handler)(nil)

func New(createUC usecase.CreateUserUC, getUC usecase.GetUserUC, log *slog.Logger) *Handler {
    return &Handler{
        createUserUC: createUC,
        getUserUC:    getUC,
        log:          log,
    }
}

// CreateUser implements POST /users
func (h *Handler) CreateUser(ctx context.Context, req *apiv1.CreateUserRequest) (apiv1.CreateUserRes, error) {
    resp, err := h.createUserUC.Execute(ctx, usecase.CreateUserReq{
        Email: req.Email,
        Name:  req.Name,
    })
    if err != nil {
        return h.mapError(err), nil
    }

    return &apiv1.User{
        ID:    apiv1.NewOptUUID(resp.ID),
        Email: req.Email,
        Name:  req.Name,
    }, nil
}

// GetUser implements GET /users/{id}
func (h *Handler) GetUser(ctx context.Context, params apiv1.GetUserParams) (apiv1.GetUserRes, error) {
    resp, err := h.getUserUC.Execute(ctx, usecase.GetUserReq{
        ID: params.ID,
    })
    if err != nil {
        return h.mapError(err), nil
    }

    return &apiv1.User{
        ID:    apiv1.NewOptUUID(resp.ID),
        Email: resp.Email,
        Name:  resp.Name,
    }, nil
}

// mapError maps domain errors to HTTP status codes
func (h *Handler) mapError(err error) apiv1.ErrorStatusCode {
    switch {
    case errors.Is(err, contract.ErrNotFound):
        return &apiv1.ErrorStatusCode{
            StatusCode: 404,
            Response: apiv1.Error{
                Code:    "not_found",
                Message: "Resource not found",
            },
        }
    case errors.Is(err, contract.ErrConflict):
        return &apiv1.ErrorStatusCode{
            StatusCode: 409,
            Response: apiv1.Error{
                Code:    "conflict",
                Message: "Resource already exists",
            },
        }
    case errors.Is(err, contract.ErrValidation):
        return &apiv1.ErrorStatusCode{
            StatusCode: 400,
            Response: apiv1.Error{
                Code:    "validation_error",
                Message: "Invalid input",
            },
        }
    default:
        h.log.ErrorContext(ctx, "internal error", "error", err)
        return &apiv1.ErrorStatusCode{
            StatusCode: 500,
            Response: apiv1.Error{
                Code:    "internal_error",
                Message: "Internal server error",
            },
        }
    }
}
```

**Pattern**: Map domain errors to HTTP status codes at transport layer, zero business logic in handlers, structured error responses.
