---
name: grpc-development
description: gRPC service design with Protocol Buffers, streaming, interceptors, error handling, and service mesh
---

# gRPC Development

## Proto File Design
```protobuf
syntax = "proto3";
package user.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc StreamUpdates(StreamUpdatesRequest) returns (stream UserEvent);
}

message GetUserRequest {
  string id = 1;
}

message User {
  string id = 1;
  string name = 2;
  string email = 3;
  google.protobuf.Timestamp created_at = 4;
}
```

## Streaming Patterns
- **Unary**: Single request → single response (most common)
- **Server streaming**: Single request → stream of responses (feeds, logs)
- **Client streaming**: Stream of requests → single response (uploads, batch)
- **Bidirectional**: Stream both directions (chat, real-time sync)

## Error Handling
- Use standard gRPC status codes (OK, NOT_FOUND, INVALID_ARGUMENT, etc.)
- Include error details with `google.rpc.Status` for rich errors
- Map domain errors to gRPC codes consistently
- Log errors server-side; return safe messages to clients

## Interceptors (Middleware)
- Authentication: validate tokens, extract user context
- Logging: structured request/response logging
- Metrics: request count, latency, error rate
- Tracing: OpenTelemetry span propagation
- Recovery: catch panics, return Internal error

## Best Practices
- Use `buf` for proto linting, breaking change detection, code generation
- Version services: `package user.v1;`
- Use field numbers wisely — never reuse deleted field numbers
- Keep messages backward compatible
- Use well-known types: Timestamp, Duration, Empty, FieldMask
- Implement health checking: `grpc.health.v1.Health`
