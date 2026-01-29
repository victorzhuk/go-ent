---
name: ent-handoffs
description: "Agent delegation patterns and irreversible action checkpoints. Preloaded by coordinator agents."
version: "1.0.0"
author: "go-ent"
disable-model-invocation: true
user-invocable: false
---

## Role

Guidance for when to delegate between agents, escalation patterns, and safety checkpoints before irreversible operations.

## Irreversible Action Checkpoints

Before performing any irreversible operation, pause and verify:

- File deletions
- Destructive git operations (`git push --force`, `git reset --hard`)
- Database schema changes
- Dependency version upgrades
- Breaking API changes
- Production configuration changes

**Checkpoint Process:**
1. Confirm requirement - Is this truly necessary?
2. Verify target - Is the file/branch/environment correct?
3. Check alternatives - Can we rename, revert, or deprecate instead?
4. Plan backup - What's the rollback strategy?
5. Apply judgment - Would a senior dev do this without asking?

**If any uncertainty exists: ASK BEFORE PROCEEDING**

## Handoff Agents

### @ent/coder
**Use when**: Implementation, writing new features, following Clean Architecture
**From**: architect, planner, debugger

### @ent/tester
**Use when**: Writing tests, TDD cycles, improving coverage
**From**: coder, debugger

### @ent/reviewer
**Use when**: Code review, quality checks, before merging
**From**: coder, debugger, tester
**Escalation**: → reviewer-heavy for complex architectural review

### @ent/debugger
**Use when**: Multi-file bugs, integration issues, test failures
**From**: coder, tester
**Escalation**: → debugger-heavy for concurrency, performance, multi-service failures

### @ent/architect
**Use when**: System design, architecture decisions, technology selection
**From**: Initial planning phase

### @ent/planner
**Use when**: Task breakdown, implementation plans, effort estimation
**From**: architect

## Handoff vs. Escalation

**Handoff**: Transferring to agent with different specialization (clear deliverable, known scope)
**Escalation**: Asking for help, approval, or additional expertise (uncertainty, safety concerns)

**Decision Flow:**
```
Uncertain? → Apply principals → Apply judgment → Still uncertain? → ASK → Need guidance? → ESCALATE
```

## Common Flows

### Feature Implementation
```
architect → planner → coder → tester → reviewer
```

### Bug Fix
```
debugger-fast OR debugger → tester → reviewer
```

### Architecture Change
```
architect → reviewer-heavy → planner → coder
```

## Handoff Requirements

When handing off, include:
- Context on what was done
- Relevant files/paths
- Expected outcomes
- Constraints/considerations

See full handoff guide in `plugins/go-ent/agents/prompts/shared/_handoffs.md`
