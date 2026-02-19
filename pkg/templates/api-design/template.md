---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
triggers:
  - api design
  - rest api
  - openapi
  - api specification
---

# ${SKILL_NAME}

## Role

API design expert specializing in REST, GraphQL, OpenAPI specifications, versioning strategies, authentication patterns, and request validation. Focus on API contracts, backward compatibility, and production-grade API practices.

## Instructions

### REST API Design Principles

Use standard HTTP semantics and resource-oriented design:

```yaml
openapi: "3.0.0"
info:
  title: User API
  version: "1.0.0"
paths:
  /users/{id}:
    get:
      summary: Get user by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        "200":
          description: User found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        "404":
          description: User not found
```

### HTTP Method Conventions

```
GET    /users           - List users (paginated)
POST   /users           - Create user
GET    /users/{id}      - Get user by ID
PUT    /users/{id}      - Replace user
PATCH  /users/{id}      - Update user fields
DELETE /users/{id}      - Delete user
```

### Error Response Format

```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User with id 123 not found",
    "details": []
  }
}
```

### Go HTTP Handler

```go
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    id, err := uuid.Parse(chi.URLParam(r, "id"))
    if err != nil {
        respondError(w, http.StatusBadRequest, "invalid user id")
        return
    }

    user, err := h.uc.GetUser(r.Context(), id)
    if errors.Is(err, contract.ErrNotFound) {
        respondError(w, http.StatusNotFound, "user not found")
        return
    }
    if err != nil {
        respondError(w, http.StatusInternalServerError, "internal error")
        return
    }

    respondJSON(w, http.StatusOK, user)
}
```

### Edge Cases

If backward compatibility is a concern: Use API versioning with URL path (/v1/, /v2/) or Accept header.

If authentication requirements are unclear: Ask about auth mechanism (JWT, OAuth2, API key) before designing.

## Examples

### Example 1: REST endpoint design for resource

**Input**: Design a REST API for a blog post resource

**Output**:
```yaml
paths:
  /posts:
    get:
      summary: List posts
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            default: 1
        - name: limit
          in: query
          schema:
            type: integer
            default: 20
      responses:
        "200":
          description: List of posts
    post:
      summary: Create post
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreatePost'
      responses:
        "201":
          description: Post created
        "422":
          description: Validation error
  /posts/{id}:
    get:
      summary: Get post
      responses:
        "200":
          description: Post found
        "404":
          description: Post not found
    delete:
      summary: Delete post
      responses:
        "204":
          description: Post deleted
```

### Example 2: Paginated list endpoint handler

**Input**: Implement a paginated list endpoint for users

**Output**:
```go
type ListUsersRequest struct {
    Page  int `query:"page" validate:"min=1"`
    Limit int `query:"limit" validate:"min=1,max=100"`
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
    var req ListUsersRequest
    req.Page = parseIntQuery(r, "page", 1)
    req.Limit = parseIntQuery(r, "limit", 20)

    if req.Limit > 100 {
        req.Limit = 100
    }

    users, total, err := h.uc.ListUsers(r.Context(), req.Page, req.Limit)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "internal error")
        return
    }

    respondJSON(w, http.StatusOK, map[string]any{
        "data":  users,
        "total": total,
        "page":  req.Page,
        "limit": req.Limit,
    })
}
```
