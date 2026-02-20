---
name: clean-architecture
description: Clean Architecture, DDD, SOLID principles, dependency injection, and layered design patterns
---

# Clean Architecture & Design Patterns

## Clean Architecture Layers
```
┌─────────────────────────────────────────┐
│  Frameworks & Drivers (outermost)       │  HTTP, DB, UI, external services
├─────────────────────────────────────────┤
│  Interface Adapters                      │  Controllers, gateways, presenters
├─────────────────────────────────────────┤
│  Application Business Rules             │  Use cases, application services
├─────────────────────────────────────────┤
│  Enterprise Business Rules (innermost)  │  Entities, value objects, domain events
└─────────────────────────────────────────┘
```

## Dependency Rule
- Dependencies point INWARD only
- Inner layers define interfaces; outer layers implement them
- Domain layer has ZERO framework dependencies
- Use dependency injection to wire implementations

## SOLID Principles
- **S**ingle Responsibility: One reason to change per class/module
- **O**pen/Closed: Open for extension, closed for modification
- **L**iskov Substitution: Subtypes must be substitutable for base types
- **I**nterface Segregation: Many specific interfaces > one general interface
- **D**ependency Inversion: Depend on abstractions, not concretions

## Domain-Driven Design
- **Entities**: Objects with identity that persist over time
- **Value Objects**: Immutable objects defined by their attributes
- **Aggregates**: Cluster of entities with a root entity as entry point
- **Repositories**: Abstract persistence — interface in domain, impl in infrastructure
- **Domain Events**: Decouple bounded contexts
- **Bounded Contexts**: Each context has its own ubiquitous language

## Patterns
- **Repository Pattern**: Abstract data access behind interfaces
- **Unit of Work**: Group operations in a transaction
- **CQRS**: Separate read and write models for complex domains
- **Event Sourcing**: Store events, not state (when audit trail is critical)
- **Specification Pattern**: Encapsulate business rules as composable objects

## Anti-Patterns to Avoid
- Anemic domain model (logic in services, dumb entities)
- God classes / God modules
- Leaky abstractions (domain depending on infrastructure types)
- Over-engineering (not every app needs full DDD)
- Premature abstraction (extract interfaces when you have 2+ implementations)
