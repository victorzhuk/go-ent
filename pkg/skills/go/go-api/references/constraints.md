# Constraints

- Include spec-first design (OpenAPI for REST, protobuf for gRPC)
- Include code generation with ogen/protoc
- Include transport layer with zero business logic
- Include proper error mapping (domain errors → HTTP status codes / gRPC codes)
- Include request/response DTOs with validation
- Include context propagation throughout handlers
- Include proper logging and metrics in handlers
- Exclude business logic in transport layer (delegate to usecases)
- Exclude direct database access from handlers
- Exclude unvalidated input processing
- Exclude manual JSON marshaling/unmarshaling (use generated types)
- Exclude breaking changes to public API without versioning
- Bound to transport layer only, call usecases for business logic
- Follow REST conventions for HTTP (status codes, resource naming)
- Follow gRPC best practices (streaming, deadlines, metadata)