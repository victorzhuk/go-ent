## REST Conventions
- Use nouns for resources: `/users`, `/orders`
- Use HTTP verbs: GET (read), POST (create), PUT (replace), PATCH (update), DELETE
- Use plural nouns: `/users` not `/user`
- Nest for relationships: `/users/{id}/orders`
- Use query params for filtering: `/users?status=active&sort=name`

## Status Codes
- 200 OK — successful GET/PUT/PATCH
- 201 Created — successful POST (include Location header)
- 204 No Content — successful DELETE
- 400 Bad Request — validation errors
- 401 Unauthorized — missing/invalid auth
- 403 Forbidden — authenticated but not authorized
- 404 Not Found — resource doesn't exist
- 409 Conflict — duplicate or state conflict
- 422 Unprocessable Entity — semantic validation error
- 429 Too Many Requests — rate limited
- 500 Internal Server Error — unexpected server error

## Error Response Format (RFC 7807)
```json
{
  "type": "https://api.example.com/errors/validation",
  "title": "Validation Error",
  "status": 422,
  "detail": "Email is already in use",
  "instance": "/users",
  "errors": [{"field": "email", "message": "already exists"}]
}
```

## Pagination
```json
// Cursor-based (preferred for large datasets)
GET /users?cursor=eyJpZCI6MTAwfQ&limit=20

{
  "data": [...],
  "pagination": {
    "next_cursor": "eyJpZCI6MTIwfQ",
    "has_more": true
  }
}
```

## Versioning
- URL path: `/api/v1/users` (simple, explicit)
- Header: `Accept: application/vnd.api+json;version=1` (cleaner URLs)
- Never break backward compatibility within a version

## OpenAPI / Swagger
- Write spec first, generate code second
- Document all endpoints, request/response schemas, and error codes
- Use examples for complex objects
- Keep spec in version control alongside code
