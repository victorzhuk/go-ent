---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
version: "${VERSION}"
author: "${AUTHOR}"
tags: [${TAGS}]
triggers:
  - pattern: "api design|rest api|graphql|openapi|api specification"
    weight: 0.9
  - keywords: ["api", "rest", "graphql", "openapi", "swagger", "api design", "api versioning"]
    weight: 0.8
  - filePattern: "*.yaml|*.yml|*.openapi"
    weight: 0.7
---

# ${SKILL_NAME}

<role>
API design expert specializing in REST, GraphQL, OpenAPI specifications, versioning strategies, authentication patterns, and request validation.
Focus on API contracts, backward compatibility, and production-grade API practices.
</role>

<instructions>

## REST API Design Principles

Use standard HTTP semantics and resource-oriented design:

```yaml
# OpenAPI 3.0 specification
openapi: 3.0.3
info:
  title: User API
  version: 1.0.0
  description: User management API
paths:
  /users:
    get:
      summary: List users
      operationId: listUsers
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
        - name: offset
          in: query
          schema:
            type: integer
            minimum: 0
            default: 0
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/User'
                  meta:
                    $ref: '#/components/schemas/PaginationMeta'
        '400':
          $ref: '#/components/responses/BadRequest'
    post:
      summary: Create user
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
        '400':
          $ref: '#/components/responses/BadRequest'
        '409':
          $ref: '#/components/responses/Conflict'
  /users/{userId}:
    get:
      summary: Get user by ID
      operationId: getUser
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          $ref: '#/components/responses/NotFound'
    put:
      summary: Update user
      operationId: updateUser
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateUserRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '400':
          $ref: '#/components/responses/BadRequest'
        '404':
          $ref: '#/components/responses/NotFound'
    delete:
      summary: Delete user
      operationId: deleteUser
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '204':
          description: No Content
        '404':
          $ref: '#/components/responses/NotFound'
```

**Key principles**:
- Use plural nouns for collections (`/users`, `/posts`)
- Use HTTP verbs correctly (GET for read, POST for create, PUT for full replace, PATCH for partial, DELETE for delete)
- Return proper status codes (200, 201, 204, 400, 404, 409, 500)
- Include pagination for list endpoints
- Use OpenAPI for documentation

## GraphQL Schema Design

Use GraphQL for flexible querying and type safety:

```graphql
# schema.graphql
type Query {
  users(limit: Int = 20, offset: Int = 0): UserConnection!
  user(id: ID!): User
}

type Mutation {
  createUser(input: CreateUserInput!): CreateUserPayload!
  updateUser(id: ID!, input: UpdateUserInput!): UpdateUserPayload!
  deleteUser(id: ID!): DeleteUserPayload!
}

type User {
  id: ID!
  email: String!
  createdAt: DateTime!
  updatedAt: DateTime!
}

type UserConnection {
  edges: [UserEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type UserEdge {
  node: User!
  cursor: String!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}

input CreateUserInput {
  email: String!
}

input UpdateUserInput {
  email: String
}

type CreateUserPayload {
  user: User
  errors: [UserError!]!
}

type UpdateUserPayload {
  user: User
  errors: [UserError!]!
}

type DeleteUserPayload {
  success: Boolean!
  errors: [UserError!]!
}

type UserError {
  field: [String!]!
  message: String!
}

scalar DateTime
```

**Key patterns**:
- Use connection pattern for pagination
- Return payload types with errors for mutations
- Use input types for mutation arguments
- Define custom scalars for complex types

## API Versioning

Version APIs for backward compatibility:

### URL Versioning (Recommended)

```yaml
# /api/v1/users
# /api/v2/users
paths:
  /api/v1/users:
    get:
      summary: List users (v1)
  /api/v2/users:
    get:
      summary: List users (v2)
```

### Header Versioning

```yaml
paths:
  /users:
    get:
      parameters:
        - name: API-Version
          in: header
          description: API version (e.g., v1, v2)
          schema:
            type: string
            default: v1
```

**Guidelines**:
- Use URL versioning for breaking changes
- Maintain old versions for at least 6 months
- Document migration paths between versions
- Use semantic versioning for breaking vs non-breaking changes

## Authentication Patterns

### Bearer Token (JWT)

```yaml
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
security:
  - bearerAuth: []

paths:
  /users:
    get:
      security:
        - bearerAuth: []
```

### API Key

```yaml
components:
  securitySchemes:
    apiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
security:
  - apiKeyAuth: []
```

### OAuth2

```yaml
components:
  securitySchemes:
    oauth2:
      type: oauth2
      flows:
        authorizationCode:
          authorizationUrl: https://example.com/oauth/authorize
          tokenUrl: https://example.com/oauth/token
          scopes:
            read:users: Read user data
            write:users: Create and modify users
security:
  - oauth2:
      - read:users
```

## Request Validation

Validate input using JSON Schema or OpenAPI types:

```yaml
components:
  schemas:
    CreateUserRequest:
      type: object
      required:
        - email
      properties:
        email:
          type: string
          format: email
          maxLength: 255
          description: User email address
        password:
          type: string
          minLength: 8
          maxLength: 128
          pattern: '^(?=.*[A-Z])(?=.*[a-z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$'
          description: Strong password with uppercase, lowercase, number, and special character
        age:
          type: integer
          minimum: 18
          maximum: 120
        preferences:
          type: object
          properties:
            notifications:
              type: boolean
            theme:
              type: string
              enum: [light, dark, auto]
          additionalProperties: false
```

**Validation rules**:
- Use `required` for mandatory fields
- Use `format` for standard types (email, uuid, date-time)
- Use `minLength`/`maxLength` for string length
- Use `minimum`/`maximum` for numeric ranges
- Use `pattern` for complex validation
- Use `enum` for fixed values
- Set `additionalProperties: false` to reject unknown fields

## Error Responses

Use consistent error response format:

```yaml
components:
  responses:
    BadRequest:
      description: Bad request
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    Conflict:
      description: Resource conflict
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
  schemas:
    Error:
      type: object
      required:
        - error
      properties:
        error:
          type: object
          required:
            - code
            - message
          properties:
            code:
              type: string
              example: VALIDATION_ERROR
              description: Machine-readable error code
            message:
              type: string
              example: Email is required
              description: Human-readable error message
            details:
              type: array
              items:
                type: object
                properties:
                  field:
                    type: string
                    example: email
                  message:
                    type: string
                    example: is required
            request_id:
              type: string
              format: uuid
              description: Request ID for tracing
```

## Pagination Patterns

### Offset-based (Simple)

```yaml
parameters:
  - name: limit
    in: query
    schema:
      type: integer
      minimum: 1
      maximum: 100
      default: 20
  - name: offset
    in: query
    schema:
      type: integer
      minimum: 0
      default: 0
```

### Cursor-based (Recommended for large datasets)

```yaml
components:
  schemas:
    CursorPagination:
      type: object
      properties:
        first:
          type: integer
          description: Number of items to return from beginning
        last:
          type: integer
          description: Number of items to return from end
        before:
          type: string
          description: Cursor for items before this point
        after:
          type: string
          description: Cursor for items after this point
```

## Rate Limiting

```yaml
x-ratelimit-limit: '1000'
x-ratelimit-remaining: '999'
x-ratelimit-reset: '1640995200'

responses:
  '429':
    description: Too Many Requests
    headers:
      X-RateLimit-Limit:
        schema:
          type: integer
      X-RateLimit-Remaining:
        schema:
          type: integer
      X-RateLimit-Reset:
        schema:
          type: integer
      Retry-After:
        schema:
          type: integer
```

## CORS Configuration

```yaml
x-cors:
  allowedOrigins:
    - https://example.com
    - https://*.example.com
  allowedMethods:
    - GET
    - POST
    - PUT
    - DELETE
    - OPTIONS
  allowedHeaders:
    - Content-Type
    - Authorization
    - X-API-Key
  exposedHeaders:
    - X-Request-ID
  maxAge: 86400
  credentials: true
```

</instructions>

<constraints>
- Use standard HTTP methods correctly (GET, POST, PUT, PATCH, DELETE)
- Return appropriate HTTP status codes for all responses
- Include OpenAPI specification for REST APIs
- Define GraphQL schemas with proper types and relationships
- Use versioning strategy (URL versioning recommended)
- Implement authentication (JWT, API key, or OAuth2)
- Validate all input requests with clear error messages
- Use consistent error response format
- Include pagination for list endpoints
- Document all endpoints with examples
- Follow REST naming conventions (plural nouns, resource-oriented)
- Exclude deprecated or undocumented endpoints
- Exclude ambiguous or inconsistent error codes
- Expose rate limiting headers for all endpoints
</constraints>

<edge_cases>
If multiple authentication methods are needed: Support multiple auth schemes in OpenAPI spec and document which endpoints accept which methods.

If backward compatibility is required: Use versioning strategy and maintain old endpoints while documenting migration path.

If validation logic is complex: Use JSON Schema with custom patterns or consider external validation service with clear error responses.

If pagination with large datasets: Use cursor-based pagination instead of offset-based to avoid deep pagination issues.

If concurrent modifications occur: Implement optimistic locking with ETag headers or version fields and return 409 Conflict on modification failure.

If API requires real-time updates: Consider WebSocket or Server-Sent Events (SSE) instead of polling for real-time features.

If rate limiting is hit: Return 429 status with Retry-After header and clear error message explaining limit.

If GraphQL query is too complex: Limit query depth, complexity, or field count with specific error responses for violations.

If API versioning strategy is unclear: Use URL versioning (`/api/v1/`) for breaking changes and document all versions with sunset dates.
</edge_cases>

<examples>
<example>
<input>Design a REST API for a blog with posts, comments, and tags</input>
<output>
```yaml
openapi: 3.0.3
info:
  title: Blog API
  version: 1.0.0
  description: Blog content management API
servers:
  - url: https://api.example.com/api/v1
paths:
  /posts:
    get:
      summary: List posts
      operationId: listPosts
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
        - name: offset
          in: query
          schema:
            type: integer
            minimum: 0
            default: 0
        - name: tag
          in: query
          schema:
            type: string
          description: Filter posts by tag
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/Post'
                  meta:
                    $ref: '#/components/schemas/PaginationMeta'
    post:
      summary: Create post
      operationId: createPost
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreatePostRequest'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Post'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
  /posts/{postId}:
    get:
      summary: Get post by ID
      operationId: getPost
      parameters:
        - name: postId
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Post'
        '404':
          $ref: '#/components/responses/NotFound'
    put:
      summary: Update post
      operationId: updatePost
      security:
        - bearerAuth: []
      parameters:
        - name: postId
          in: path
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdatePostRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Post'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '404':
          $ref: '#/components/responses/NotFound'
  /posts/{postId}/comments:
    get:
      summary: List comments for post
      operationId: listPostComments
      parameters:
        - name: postId
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Comment'
    post:
      summary: Create comment on post
      operationId: createComment
      security:
        - bearerAuth: []
      parameters:
        - name: postId
          in: path
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateCommentRequest'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Comment'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
  /tags:
    get:
      summary: List all tags
      operationId: listTags
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Tag'
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
  schemas:
    Post:
      type: object
      properties:
        id:
          type: string
          format: uuid
        title:
          type: string
        content:
          type: string
        author:
          type: object
          properties:
            id:
              type: string
              format: uuid
            name:
              type: string
        tags:
          type: array
          items:
            $ref: '#/components/schemas/Tag'
        createdAt:
          type: string
          format: date-time
        updatedAt:
          type: string
          format: date-time
    CreatePostRequest:
      type: object
      required:
        - title
        - content
      properties:
        title:
          type: string
          minLength: 1
          maxLength: 200
        content:
          type: string
          minLength: 1
        tags:
          type: array
          items:
            type: string
    UpdatePostRequest:
      type: object
      properties:
        title:
          type: string
          minLength: 1
          maxLength: 200
        content:
          type: string
          minLength: 1
        tags:
          type: array
          items:
            type: string
    Comment:
      type: object
      properties:
        id:
          type: string
          format: uuid
        post:
          type: string
          format: uuid
        author:
          type: object
          properties:
            id:
              type: string
              format: uuid
            name:
              type: string
        content:
          type: string
        createdAt:
          type: string
          format: date-time
    CreateCommentRequest:
      type: object
      required:
        - content
      properties:
        content:
          type: string
          minLength: 1
          maxLength: 1000
    Tag:
      type: object
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
        postCount:
          type: integer
    PaginationMeta:
      type: object
      properties:
        total:
          type: integer
        limit:
          type: integer
        offset:
          type: integer
  responses:
    BadRequest:
      description: Bad request
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    Unauthorized:
      description: Unauthorized
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    Error:
      type: object
      required:
        - error
      properties:
        error:
          type: object
          required:
            - code
            - message
          properties:
            code:
              type: string
            message:
              type: string
```
</output>
</example>

<example>
<input>Create a GraphQL schema for an e-commerce API with products, orders, and users</input>
<output>
```graphql
# schema.graphql
type Query {
  products(limit: Int = 20, offset: Int = 0, category: String): ProductConnection!
  product(id: ID!): Product
  orders(userId: ID!, limit: Int = 20, offset: Int = 0): OrderConnection!
  order(id: ID!): Order
  me: User
}

type Mutation {
  createProduct(input: CreateProductInput!): CreateProductPayload!
  updateProduct(id: ID!, input: UpdateProductInput!): UpdateProductPayload!
  deleteProduct(id: ID!): DeleteProductPayload!
  createOrder(input: CreateOrderInput!): CreateOrderPayload!
  updateOrderStatus(id: ID!, status: OrderStatus!): UpdateOrderStatusPayload!
  cancelOrder(id: ID!): CancelOrderPayload!
  register(input: RegisterInput!): RegisterPayload!
  login(input: LoginInput!): LoginPayload!
}

type Product {
  id: ID!
  name: String!
  description: String
  price: Float!
  category: Category!
  stock: Int!
  images: [String!]!
  createdAt: DateTime!
  updatedAt: DateTime!
}

type Category {
  id: ID!
  name: String!
  slug: String!
  productCount: Int!
}

type Order {
  id: ID!
  user: User!
  items: [OrderItem!]!
  status: OrderStatus!
  total: Float!
  shippingAddress: Address!
  createdAt: DateTime!
  updatedAt: DateTime!
}

type OrderItem {
  id: ID!
  product: Product!
  quantity: Int!
  price: Float!
}

type User {
  id: ID!
  email: String!
  name: String!
  addresses: [Address!]!
  orders: [Order!]!
  createdAt: DateTime!
}

type Address {
  id: ID!
  street: String!
  city: String!
  state: String!
  zipCode: String!
  country: String!
}

type ProductConnection {
  edges: [ProductEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type ProductEdge {
  node: Product!
  cursor: String!
}

type OrderConnection {
  edges: [OrderEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type OrderEdge {
  node: Order!
  cursor: String!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}

enum OrderStatus {
  PENDING
  PROCESSING
  SHIPPED
  DELIVERED
  CANCELLED
}

input CreateProductInput {
  name: String!
  description: String
  price: Float!
  categoryId: ID!
  stock: Int!
  images: [String!]!
}

input UpdateProductInput {
  name: String
  description: String
  price: Float
  categoryId: ID
  stock: Int
  images: [String!]
}

input CreateOrderInput {
  items: [OrderItemInput!]!
  shippingAddressId: ID!
}

input OrderItemInput {
  productId: ID!
  quantity: Int!
}

input RegisterInput {
  email: String!
  password: String!
  name: String!
}

input LoginInput {
  email: String!
  password: String!
}

type CreateProductPayload {
  product: Product
  errors: [UserError!]!
}

type UpdateProductPayload {
  product: Product
  errors: [UserError!]!
}

type DeleteProductPayload {
  success: Boolean!
  errors: [UserError!]!
}

type CreateOrderPayload {
  order: Order
  errors: [UserError!]!
}

type UpdateOrderStatusPayload {
  order: Order
  errors: [UserError!]!
}

type CancelOrderPayload {
  success: Boolean!
  errors: [UserError!]!
}

type RegisterPayload {
  user: User
  token: String
  errors: [UserError!]!
}

type LoginPayload {
  user: User
  token: String
  errors: [UserError!]!
}

type UserError {
  field: [String!]!
  message: String!
}

scalar DateTime
```
</output>
</example>
</examples>

<output_format>
Provide production-ready API designs following established patterns:

1. **OpenAPI Spec**: Complete, valid OpenAPI 3.0+ specification
2. **REST Principles**: Resource-oriented design with proper HTTP methods
3. **GraphQL Schema**: Type-safe schema with proper relationships
4. **Versioning**: Clear versioning strategy with migration paths
5. **Authentication**: Documented auth schemes (JWT, API key, OAuth2)
6. **Validation**: Input validation with clear error messages
7. **Error Handling**: Consistent error response format
8. **Pagination**: Proper pagination strategy (cursor-based recommended)
9. **Documentation**: Complete docs with examples and descriptions
10. **Best Practices**: Rate limiting, CORS, security headers

Focus on API contracts, backward compatibility, and developer experience.
</output_format>
