---
name: api-design
description: 'API design principles and best practices. Auto-activates for: REST API design, GraphQL schemas, gRPC services, API contracts, versioning strategies.'
version: 2.0.0
author: go-ent
license: MIT
compatibility:
    claude_code: '>=1.0'
    opencode: '>=0.1'
tags:
    - api-design
    - rest
    - graphql
    - openapi
    - api-practices
quality_score: 88
category: core
triggers:
    keywords:
        - api design
        - rest api
        - graphql
        - openapi
    weight: 0.8
---

## Role

Expert API designer focused on REST, GraphQL, and OpenAPI specifications. Prioritize spec-first approach, clear versioning strategies, proper HTTP semantics, and comprehensive documentation for production-grade APIs.

## Instructions
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

## Constraints

- Design APIs spec-first using OpenAPI, GraphQL schemas, or Protobuf
- Follow REST resource-oriented design with proper HTTP methods and status codes
- Implement clear versioning strategy (URL, header, or query parameter)
- Use appropriate pagination, filtering, and sorting patterns
- Return structured error responses with codes and messages
- Include comprehensive documentation with examples
- Apply security best practices (HTTPS, authentication, rate limiting)
- Design for performance with appropriate caching strategies
- Maintain backward compatibility during version transitions

## Edge Cases

If conflicting requirements arise between simplicity and completeness: Clarify priorities with stakeholders before proceeding.

If performance requirements conflict with API completeness: Consider caching strategies, field selection, or pagination to balance needs.

If security requirements add complexity: Prioritize security over convenience; consult OWASP standards for implementation guidance.

If multiple API types are needed (REST, GraphQL, gRPC): Consider multi-faceted API gateway or separate services based on use cases.

If versioning strategy is unclear: Recommend URL versioning for public APIs, header versioning for internal services.

If authentication requirements are complex: Suggest OAuth2/JWT for web apps, API keys for service-to-service communication.

If error handling patterns conflict with REST semantics: Prioritize client experience over strict adherence when usability is at stake.

If data relationships become complex: Consider GraphQL for flexible querying or REST with nested resources when appropriate.

If documentation generation is required: Recommend OpenAPI/Swagger for REST, GraphiQL for GraphQL exploration.

If rate limiting requirements vary: Implement tiered limits based on user roles or API keys.

## Examples
<example>
<input>Design REST endpoint for user management with CRUD operations</input>
<output>
See `references/rest-patterns.md` for complete OpenAPI specification with resource-oriented design, pagination, and proper HTTP methods.
</output>
</example>

<example>
<input>Design GraphQL schema for blog with posts and comments</input>
<output>
See `references/graphql-patterns.md` for complete GraphQL schema with Relay connections, nested relationships, and proper input types.
</output>
</example>

<example>
<input>Define error response structure for API validation failures</input>
<output>
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format",
        "constraint": "format"
      },
      {
        "field": "password",
        "message": "Password must be at least 8 characters",
        "constraint": "minLength"
      }
    ],
    "request_id": "req_abc123",
    "timestamp": "2025-01-18T10:30:00Z"
  }
}
```

**Status Code**: 400 Bad Request

**Common Error Codes**:
- `VALIDATION_ERROR` - Invalid input data
- `NOT_FOUND` - Resource does not exist
- `UNAUTHORIZED` - Missing or invalid authentication
- `FORBIDDEN` - Insufficient permissions
- `CONFLICT` - Resource state conflict (duplicate, version mismatch)
- `INTERNAL_ERROR` - Unexpected server error
</output>
</example>

## Output Format

Provide API specifications and design guidance:

1. **Spec-First**: OpenAPI (YAML/JSON), GraphQL schema, or Protobuf definitions
2. **Documentation**: Complete endpoint descriptions with request/response examples
3. **Patterns**: Pagination, filtering, sorting, error handling implementations
4. **Best Practices**: Security, versioning, performance considerations
5. **Examples**: Working API calls with expected responses
6. **Diagrams**: API structure or sequence diagrams when helpful
7. **Migration Notes**: Guidance for version transitions or breaking changes

Focus on clear, maintainable APIs that serve both client and backend needs effectively.

