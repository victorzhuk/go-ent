# Proposal: Add Agent System

## Overview

Implement the core agent orchestration system: role definitions, skill registry, agent selector based on task complexity, and delegation logic from existing `go-ent:lead.md`.

## Rationale

### Problem
No automatic agent selection - users must manually choose which agent/model to use for each task.

### Solution
- **Agent selector**: Analyzes task complexity, selects optimal agent role + model
- **Skill registry**: Loads skills from markdown files, matches to context
- **Delegation matrix**: Implements existing `go-ent:lead.md` decision logic
- **Complexity analyzer**: Classifies tasks as Trivial → Simple → Moderate → Complex → Architectural

## Key Components

1. `internal/agent/selector.go` - Main selection algorithm
2. `internal/agent/complexity.go` - Task complexity analysis
3. `internal/agent/delegate.go` - Delegation decision matrix from go-ent:lead.md
4. `internal/skill/registry.go` - Skill loading and matching

## Dependencies

- Requires: P0 (restructure), P1 (domain-types), P2 (config-system)
- Blocks: P4 (execution-engine)

## Success Criteria

- [ ] Agent selector chooses appropriate role based on complexity
- [ ] Delegation matrix matches existing go-ent:lead.md logic
- [ ] Skill registry loads from markdown files
- [ ] Complexity analyzer classifies tasks accurately
