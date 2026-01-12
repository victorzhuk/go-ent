---
name: api-design
description: "API design principles and best practices. Auto-activates for: REST API design, GraphQL schemas, gRPC services, API contracts, versioning strategies."
version: 1.0.0
---

# API Design

## REST Principles

### Resource-Oriented
- Nouns not verbs: `/users`, `/orders`
- HTTP methods = CRUD: GET, POST, PUT/PATCH, DELETE
- Collections vs instances: `/users` vs `/users/{id}`

### Status Codes

| Code | Meaning | Use |
|------|---------|-----|
| 200 | OK | GET/PUT/PATCH success |
| 201 | Created | POST success |
| 204 | No Content | DELETE success |
| 400 | Bad Request | Invalid input |
| 401 | Unauthorized | Missing/invalid auth |
| 403 | Forbidden | No permission |
| 404 | Not Found | Resource missing |
| 409 | Conflict | Concurrency/uniqueness |
| 500 | Internal Error | Server failure |

### Versioning

| Strategy | Format | Pros/Cons |
|----------|--------|-----------|
| URL | `/v1/users` | Clear, simple (breaks caching) |
| Header | `Accept: v=1` | Clean URLs (less obvious) |
| Query | `?version=1` | Flexible (not RESTful) |

## Design Patterns

### Pagination
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 150
  },
  "links": {
    "next": "/users?page=2"
  }
}
```

### Filtering
```
GET /users?status=active&role=admin&created_after=2025-01-01
```

### Sorting
```
GET /users?sort=created_at:desc,name:asc
```

## Error Responses

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input",
    "details": [
      {"field": "email", "message": "Invalid format"}
    ]
  }
}
```

## API Comparison

| Type | Best For | Trade-offs |
|------|----------|------------|
| REST | CRUD, public APIs | Over-fetching/under-fetching |
| GraphQL | Complex data, mobile apps | Query complexity, caching |
| gRPC | Microservices, performance | Binary format, less tooling |

## Security

- HTTPS only
- Rate limiting
- Authentication (OAuth2, JWT)
- Input validation
- CORS policies
- No sensitive data in URLs

## Documentation

- OpenAPI/Swagger for REST
- GraphQL introspection
- Protobuf definitions for gRPC
- Include examples
- Document error responses
