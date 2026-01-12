---
name: arch-core
description: "Architecture patterns and system design principles. Auto-activates for: architecture decisions, design patterns, system boundaries, component interaction, architectural trade-offs."
version: 1.0.0
---

# Architecture Core

## Core Principles

### Separation of Concerns
- Single, well-defined responsibility per component
- Clear boundaries between layers/modules
- Minimal coupling, high cohesion

### Dependency Management
- Depend on abstractions, not implementations
- Inversion of Control for dependencies
- Interface-based design at boundaries

## Common Patterns

| Pattern | When to Use | Trade-offs |
|---------|-------------|------------|
| Layered | Clear separation (UI, business, data) | Can become rigid |
| Clean Architecture | Framework/DB independence | More boilerplate |
| Hexagonal | Business logic isolation | Complex for simple apps |
| CQRS | Different read/write needs | Increased complexity |
| Event-Driven | Async, loosely coupled systems | Harder to debug |

## Architectural Decisions

**ADR Template**:
```markdown
# ADR-001: Title

## Context
Problem and constraints

## Decision
Chosen approach with rationale

## Consequences
Positive and negative outcomes

## Alternatives
Other options and why rejected
```

## System Trade-offs

| Approach | Pros | Cons |
|----------|------|------|
| Monolith | Simple, low latency, easy to develop | Scaling limits, deployment coupling |
| Microservices | Independent scaling/deployment | Complexity, distributed challenges |
| Serverless | Auto-scaling, pay-per-use | Cold starts, vendor lock-in |

## Anti-Patterns

- **Big Ball of Mud** - No clear structure
- **God Object** - One class does everything
- **Tight Coupling** - Changes ripple everywhere
- **Premature Optimization** - Complexity without proof of need
- **Golden Hammer** - One pattern for all problems

## Design Checklist

- [ ] Clear component boundaries
- [ ] Dependencies point inward (Clean Architecture)
- [ ] Interfaces at boundaries
- [ ] Testable design
- [ ] Scalability considered
- [ ] Security by design
- [ ] Fail-safe defaults
