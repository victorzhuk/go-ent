# Proposal: Add Domain Types for Multi-Agent System

## Overview

Introduce core domain types that establish the vocabulary for the multi-agent orchestration system: `AgentRole`, `Runtime`, `SpecAction`, `ExecutionStrategy`, `Skill`, and related types.

## Rationale

### Problem Statement

The current codebase lacks domain concepts for:
- **Agent specialization**: No AgentRole enum to distinguish Product, Architect, Senior, Developer, Reviewer, Ops
- **Runtime environments**: No Runtime enum for ClaudeCode, OpenCode, CLI
- **Execution semantics**: No ExecutionStrategy to model single/multi/parallel execution
- **Task classification**: No SpecAction taxonomy to route work appropriately
- **Skill abstraction**: No Skill interface for reusable capabilities

These concepts are mentioned in PRD but not modeled in code.

### Current State

`internal/spec/domain.go` contains:
- `ChangeStatus`, `TaskStatus` (good, keep these)
- `Project`, `ListItem` (spec-related, not agent-related)
- No agent or execution concepts

### Target State

New `internal/domain/` package with:
- `agent.go` - AgentRole, AgentConfig, AgentCapability
- `runtime.go` - Runtime enum, RuntimeCapability
- `action.go` - SpecAction taxonomy
- `execution.go` - ExecutionStrategy, ExecutionContext, ExecutionResult
- `skill.go` - Skill interface, SkillMetadata, SkillContext
- `errors.go` - Domain-specific errors

## Design Decisions

### 1. Separate Domain Package

**Decision**: Create `internal/domain/` instead of extending `internal/spec/domain.go`

**Rationale**:
- Clean separation: spec vs agent concerns
- Avoids circular dependencies: execution → spec
- Follows DDD bounded context pattern
- Easier to test in isolation

### 2. Enum vs String Constants

**Decision**: Use custom string types with constants (e.g., `type AgentRole string`)

**Rationale**:
- Type-safe without reflection
- JSON marshaling works automatically
- Follows existing codebase pattern (ChangeStatus, TaskStatus)
- Easy to extend

### 3. Skill as Interface

**Decision**: Define `Skill` as interface, not struct

**Rationale**:
- Multiple skill implementations (built-in, custom, plugin)
- Testable via mocks
- Plugin system can provide implementations
- Follows dependency inversion principle

## Domain Model

### AgentRole Hierarchy

```
Product      - User needs, requirements, product decisions
Architect    - System design, architecture, technical decisions
Senior       - Complex implementation, debugging, code review
Developer    - Standard implementation, testing
Reviewer     - Code quality, standards enforcement
Ops          - Deployment, monitoring, production issues
```

### SpecAction Taxonomy

```
Discovery:   research, analyze, retrofit
Planning:    proposal, plan, design, split
Execution:   implement, execute, scaffold
Validation:  review, verify, debug, lint
Lifecycle:   approve, archive, status
```

### Execution Strategies

- **Single**: One agent, sequential execution
- **Multi**: Multiple agents in conversation/handoff
- **Parallel**: Independent agents working simultaneously

## Integration Points

### With Existing Code

- `internal/spec/workflow.go` - Add `AgentRole` field to track current agent
- `internal/spec/domain.go` - Import new domain types where needed

### With Future Proposals

- **P2 (config-system)**: Uses `AgentRole`, `Runtime` in configuration
- **P3 (agent-system)**: Uses `AgentRole`, `Skill` for selection
- **P4 (execution-engine)**: Uses `ExecutionStrategy`, `Runtime`

## Breaking Changes

None - this is a pure addition of new types.

## Dependencies

- **Requires**: P0 (restructure-repository) - clean import paths
- **Blocks**: P2 (config-system), P3 (agent-system), P4 (execution-engine)

## Success Criteria

- [ ] All domain types defined with clear doc comments
- [ ] Unit tests for enums (validation, string conversion)
- [ ] No circular dependencies
- [ ] Zero external dependencies (pure domain)
- [ ] Integration with existing `internal/spec/` works
