# Constraints

- Include clean architecture with clear layer boundaries
- Include domain-first design with zero external dependencies in domain layer
- Include dependency injection pattern with explicit container
- Include transaction management for write operations
- Include outbox pattern for event-driven systems
- Include graceful shutdown with 30s timeout on fresh context
- Exclude cross-layer dependencies (inward dependency rule only)
- Exclude business logic in transport layer
- Exclude direct infrastructure access from usecases
- Exclude global mutable state or singletons
- Exclude tight coupling between bounded contexts
- Bound to Transport → UseCase → Domain ← Repository flow
- Follow domain-driven design principles for complex domains
- Use interfaces at consumer side, return structs