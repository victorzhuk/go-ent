## Service Design
- One service per bounded context (DDD)
- Services own their data — no shared databases
- Communicate via APIs (sync) or events (async)
- Each service is independently deployable
- Size: team can understand and maintain it (2-pizza team rule)

## Communication Patterns
- **Synchronous**: REST/gRPC for queries and commands needing immediate response
- **Asynchronous**: Message queues for events and commands that can be deferred
- **API Gateway**: Single entry point for clients; routes to services
- **BFF (Backend for Frontend)**: Gateway per client type (web, mobile)

## Resilience
- **Circuit Breaker**: Open circuit after N failures, half-open to test recovery
- **Retry with Backoff**: Exponential backoff with jitter
- **Timeout**: Always set timeouts on external calls
- **Bulkhead**: Isolate failures — separate thread/connection pools per dependency
- **Fallback**: Degrade gracefully with cached/default responses

## Data Management
- **Database per Service**: Each service owns its schema
- **Saga Pattern**: Distributed transactions via compensating actions
  - Choreography: services react to events
  - Orchestration: central coordinator manages flow
- **CQRS**: Separate read and write models
- **Event Sourcing**: Store events, derive state

## Observability
- Distributed tracing across services (OpenTelemetry)
- Centralized logging with correlation IDs
- Service mesh metrics (Istio, Linkerd)
- Health checks and readiness probes
- Dashboards per service and system-wide

## Anti-Patterns
- Distributed monolith: services tightly coupled
- Shared database: defeats the purpose of microservices
- Too many services: operational overhead exceeds benefits
- Synchronous chains: A calls B calls C calls D = fragile
- No observability: flying blind in production
