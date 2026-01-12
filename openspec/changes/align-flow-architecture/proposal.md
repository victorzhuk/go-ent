# Align plugins/go-ent with FLOW.md Architecture

## Overview

Major refactor of go-ent plugin structure to implement FLOW.md's dual-executor model: **Claude Code for planning**, **OpenCode workers for execution**.

## Rationale

### Current State (17 commands, 7 agents)
- **Commands**: analyze, apply, archive, clarify, decompose, gen, init, lint, loop, loop-cancel, plan, registry, research, scaffold, status, tdd
- **Agents**: architect, debug, dev, lead, planner, reviewer, tester
- **Pattern**: Single-executor (Claude Code only)
- **Gap**: No `/task`, `/bug` commands; No smoke/heavy escalation; No OpenCode integration

### FLOW.md Vision
```
/plan  (Claude Code)  →  OpenSpec features
/task  (OpenCode)     →  Code, tests
/bug   (OpenCode)     →  Fix, regression test
```

**Planning Agents (Claude Code)**:
- `@planner-smoke` (Haiku) - Quick triage: needs architecture?
- `@architect` (Opus) - Full architecture design
- `@planner` (Sonnet) - Feature specification
- `@decomposer` (Sonnet) - Task breakdown

**Execution Agents (OpenCode)**:
- `@task-smoke` (GLM 4.7) - Task triage
- `@task-heavy` (Kimi K2) - Complex tasks
- `@coder` (GLM 4.7) - Primary coding
- `@reviewer` (GLM 4.7) - Code review
- `@tester` (GLM 4.7) - Test creation
- `@acceptor` (GLM 4.7) - Final acceptance
- `@reproducer` (GLM 4.7) - Bug reproduction
- `@researcher` (GLM 4.7) - Bug research
- `@debugger-smoke` (GLM 4.7) - Initial investigation
- `@debugger-heavy` (Kimi K2) - Complex bugs

### Problem
1. **Commands misaligned**: Have `/plan`, `/apply`, `/status` but missing `/task`, `/bug`
2. **No smoke/heavy pattern**: All agents same model tier, no escalation
3. **No dual-executor**: Everything runs in Claude Code, no OpenCode delegation
4. **Command bloat**: 17 commands vs FLOW's 3 main commands
5. **Agent mismatch**: Current agents don't map to FLOW roles

## Key Components

### 1. Command Consolidation

**Keep** (Core 3):
- `/plan` - Comprehensive planning in Claude Code
- `/task` - NEW: Execute task (OpenCode or Claude Code w/ ACP)
- `/bug` - NEW: Fix bug (OpenCode or Claude Code w/ ACP)

**Keep** (Support):
- `/status` - View progress
- `/apply` - Apply tasks (alias for /task)
- `/registry` - Task management
- `/init` - Project setup

**Archive/Deprecate** (10):
- `/analyze` → part of /plan
- `/clarify` → part of /plan
- `/research` → part of /plan
- `/decompose` → part of /plan
- `/gen` → use /task
- `/scaffold` → use /task
- `/tdd` → use /task with TDD flag
- `/lint` → automatic in /task
- `/loop` → automatic retry in /task
- `/loop-cancel` → /task cancel

### 2. Agent Restructuring

**Planning Agents** (Claude Code):
```
plugins/go-ent/agents/planning/
├── planner-smoke.md    (Haiku)   - Quick assessment
├── architect.md        (Opus)    - Architecture design
├── planner.md          (Sonnet)  - Feature spec
└── decomposer.md       (Sonnet)  - Task breakdown
```

**Execution Agents** (OpenCode):
```
plugins/go-ent/agents/execution/
├── task-smoke.md       (GLM 4.7) - Task triage
├── task-heavy.md       (Kimi K2) - Complex tasks
├── coder.md            (GLM 4.7) - Coding
├── reviewer.md         (GLM 4.7) - Review
├── tester.md           (GLM 4.7) - Testing
├── acceptor.md         (GLM 4.7) - Acceptance
├── reproducer.md       (GLM 4.7) - Bug repro
├── researcher.md       (GLM 4.7) - Bug research
├── debugger-smoke.md   (GLM 4.7) - Debug triage
└── debugger-heavy.md   (Kimi K2) - Complex debug
```

**Deprecated** (moved to execution):
- `dev.md` → `coder.md`
- `debug.md` → `debugger-smoke.md`
- `lead.md` → removed (delegated by claude code directly)

### 3. ACP Integration

**New**: `internal/acp/client.go`
- Connect to OpenCode via Agent Communication Protocol
- Delegate `/task` and `/bug` to external workers
- Support model selection (GLM 4.7 primary, Kimi K2 heavy)
- Fallback to Claude Code execution

### 4. Smoke/Heavy Escalation

**Pattern**:
```go
type EscalationRule struct {
    Triggers []string // ["race condition", "deadlock", "memory leak"]
    MaxRetries int    // 2 attempts before escalate
    ComplexityThreshold float64 // 0.8
}
```

**Implementation**: `internal/agent/escalation.go`

## Architecture

### Current
```
commands/ (17) → agents/ (7) → Claude Code execution
```

### FLOW.md
```
/plan  → planning agents (4) → Claude Code
/task  → execution agents (10) → OpenCode (or Claude Code w/ ACP)
/bug   → execution agents (10) → OpenCode (or Claude Code w/ ACP)
```

## Dependencies

- `add-boltdb-state-system` (state.md for task context)
- ACP protocol implementation (or fallback to Claude Code)

## Success Criteria

- [ ] 3 core commands (/plan, /task, /bug)
- [ ] Planning agents (4) in Claude Code tier
- [ ] Execution agents (10) with GLM/Kimi model specs
- [ ] Smoke/heavy escalation implemented
- [ ] ACP client for OpenCode delegation
- [ ] Fallback to Claude Code when OpenCode unavailable
- [ ] 10 deprecated commands archived
- [ ] Documentation updated (FLOW.md, AGENTS.md)
- [ ] Migration guide for existing workflows

## Migration Strategy

### Phase 1: Add New Commands
- Implement `/task` and `/bug` alongside existing commands
- Users can try new workflow without breaking old one

### Phase 2: Add New Agents
- Create execution agents with model specs
- Restructure planning agents
- Keep old agents working

### Phase 3: Deprecate Old Commands
- Add deprecation warnings to old commands
- Document migration paths
- Provide 1-month transition period

### Phase 4: Remove Deprecated
- Archive old commands to `commands/archive/`
- Remove old agents
- Update all documentation

## Risk Assessment

**High Risk**:
- Breaking existing user workflows
- ACP protocol not ready → Need robust fallback

**Medium Risk**:
- OpenCode availability/reliability
- Model selection logic complexity

**Low Risk**:
- Command consolidation (clear migration path)
- Agent restructuring (internal change)

## Rollback Plan

- Keep old commands in `commands/archive/`
- Feature flag: `FLOW_MODE=enabled|legacy`
- Can revert by restoring archived commands
