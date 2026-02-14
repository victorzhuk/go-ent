---
name: writing-plans
description: Writing implementation plans, technical proposals, RFCs, and Architecture Decision Records
---

# Writing Technical Plans

## Implementation Plan Template
```markdown
# [Feature Name] Implementation Plan

## Context
What is the current state? Why does this need to change?

## Goals
- Primary goal
- Secondary goals
- Non-goals (explicitly state what's out of scope)

## Design
### Architecture
High-level approach and component diagram

### Data Model
Schema changes, new tables, migrations

### API Changes
New/modified endpoints, request/response schemas

### Dependencies
External services, libraries, infrastructure

## Implementation Phases
### Phase 1: Foundation (est. X days)
- [ ] Task 1.1
- [ ] Task 1.2

### Phase 2: Core Logic (est. X days)
- [ ] Task 2.1
- [ ] Task 2.2

### Phase 3: Testing & Polish (est. X days)
- [ ] Task 3.1
- [ ] Task 3.2

## Risks & Mitigations
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| ...  | High   | Medium     | ...        |

## Testing Strategy
Unit, integration, E2E test plans

## Rollout Plan
Feature flags, staged rollout, monitoring

## Open Questions
- Question 1?
- Question 2?
```

## Architecture Decision Record (ADR)
```markdown
# ADR-001: Use PostgreSQL as Primary Database

## Status: Accepted

## Context
We need a database for our new microservice...

## Decision
We will use PostgreSQL 16 with pgx driver...

## Consequences
- Good: ACID compliance, JSON support, mature ecosystem
- Bad: Operational complexity vs managed NoSQL
- Neutral: Team needs PostgreSQL expertise
```

## Best Practices
- Write plans BEFORE coding — thinking time is cheaper than coding time
- Include non-goals to prevent scope creep
- Break work into independently testable/deployable phases
- Identify risks early and plan mitigations
- Get feedback on the plan before starting implementation
- Update the plan as you learn during implementation
