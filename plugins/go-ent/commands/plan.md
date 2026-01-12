---
description: Create complete OpenSpec change proposal with research and task breakdown
---

# Planning Workflow

Complete planning workflow: clarify → research → design → decompose.

## Input

`$ARGUMENTS`: Feature description or existing change ID

Examples:
- `/ent:plan "Add two-factor authentication with OTP"`
- `/ent:plan add-user-auth` - Continue existing change

## Agent Chain

| Agent               | Purpose                               | Tier     |
|---------------------|---------------------------------------|----------|
| @ent:planner-fast   | Feasibility check, initial triage     | fast     |
| @ent:architect      | System design, architecture decisions | heavy    |
| @ent:planner        | Detailed planning, spec writing       | standard |
| @ent:decomposer     | Task breakdown, dependency analysis   | standard |

**Escalation**: planner-fast → architect → planner → decomposer

---

## Workflow

### Phase 0: Initial Assessment

**Agent**: @ent:planner-fast

**Goal**: Quick feasibility check

**Steps**:
1. Parse feature description
2. Check: Is it clear enough? Are there blockers? What's the complexity?
3. Identify immediate unknowns
4. **Decision**: Proceed or request clarification

**If unclear**: Ask focused questions before continuing

### Phase 1: Clarification

**Goal**: Resolve all unknowns before design

**Checklist**:
- What problem are we solving?
- Who are the users/consumers?
- What are success criteria?
- Performance requirements?
- Security considerations?
- Constraints (time, resources, compatibility)?

**For each unknown**:
1. Identify what's unclear
2. Research existing codebase
3. Ask specific (non-yes/no) questions
4. Document answers in proposal

**Wait for user input** if critical unknowns exist

### Phase 2: Research & Technology Decisions

**Goal**: Evaluate approaches and choose solution

**Process**:
```
For each technology choice:
  Option A:
    + Pros: {advantages}
    - Cons: {limitations}
    ? Unknowns: {what we don't know}
  Option B:
    + Pros: {advantages}
    - Cons: {limitations}

  Recommendation: {choice with rationale}
```

**Research sources**:
- Existing codebase patterns (Serena)
- Official docs (WebFetch)
- Community practices (WebSearch)
- Project conventions (openspec/specs/)

**Output**: `openspec/changes/{id}/research.md`

**Present findings and get approval** before design

### Phase 3: Architecture & Design

**Agent**: @ent:architect

**Create**: `openspec/changes/{id}/proposal.md` and `design.md`

**proposal.md Structure**:
```markdown
## Summary
{What and why}

## Problem
{Current state and pain points}

## Solution
{High-level approach}

## Breaking Changes
- [ ] API changes
- [ ] Database migrations
- [ ] Configuration changes

## Affected Systems
- Component A: {impact}
- Component B: {impact}

## Alternatives Considered
1. Approach A: {why not chosen}
```

**design.md** (when needed):
- Architecture overview
- Data model (entities, relationships, migrations)
- API contracts
- Integration points
- Migration strategy
- Performance considerations
- Security considerations

**Principles**:
- Follow existing patterns unless there's a reason
- Prefer simple over clever
- Design for testability
- Consider failure modes
- Document trade-offs

**Present design and get approval** before task breakdown

### Phase 4: Spec Deltas

**Agent**: @ent:planner

**Create**: `openspec/changes/{id}/specs/` (one capability per directory)

**Spec delta format**:
```markdown
## ADDED Requirements

### REQ-XXX-001: Requirement Title
Description

#### Scenario: Success Case
**WHEN** condition
**THEN** outcome

## MODIFIED Requirements

### REQ-XXX-002: Updated Requirement
~~Old behavior~~
New behavior (reason)

## REMOVED Requirements

### REQ-XXX-003: Deprecated Feature
Explanation
```

**Requirements must have**:
- Clear acceptance criteria
- Concrete scenarios (WHEN/THEN)
- Cross-references to related requirements

**Validate**: `openspec validate {change-id} --strict`

### Phase 5: Task Decomposition

**Agent**: @ent:decomposer

**Create**: `openspec/changes/{id}/tasks.md`

**Task structure**:
```markdown
## 1. Foundation
- [ ] **1.1** Create domain entities
  - Files: internal/domain/user.go
  - Dependencies: none
  - Effort: 2h

## 2. Implementation
- [ ] **2.1** Implement repository
  - Files: internal/repository/user/postgres/
  - Dependencies: 1.1
  - Effort: 4h
  - Parallel with: 2.2
```

**Guidelines**:
- Break into <4h chunks
- Specify exact files
- Mark dependencies explicitly
- Identify parallel work
- Estimate effort
- Group by layer/capability

**Sync to registry**: Tasks auto-sync to `openspec/registry.yaml`

### Phase 6: Validation

**Final checks**:
```
1. openspec validate {change-id} --strict
2. Completeness:
   - [ ] All TBD/TODO resolved
   - [ ] All unknowns researched
   - [ ] All questions answered
   - [ ] All tasks have IDs and files
   - [ ] Dependencies form valid graph (no cycles)
3. Coverage:
   - Each requirement has tests
   - Each spec has implementing tasks
   - Each task traces to requirement
4. Consistency:
   - Proposal matches design
   - Design matches specs
   - Tasks match specs
```

---

## Output Format

```
═══════════════════════════════════════════
PLANNING: {feature description}
═══════════════════════════════════════════

📋 Change: {change-id}
   Type: {feature|enhancement|refactor}
   Breaking: {yes|no}
   Complexity: {low|medium|high}

🔍 Clarification:
   Unknowns resolved: {count}
   Open questions: 0 ✓

🔬 Research:
   Options evaluated: {count}
   Recommendation: {approach}

🏗️ Design:
   Components affected: {count}
   New entities: {count}
   API changes: {count}
   Migrations: {yes|no}

📐 Specification:
   Requirements added: {count}
   Requirements modified: {count}
   Validation: ✅ PASS

🗂️ Task Breakdown:
   Total tasks: {count}
   Parallelizable: {count}
   Critical path: T1→T3→T5
   Estimated effort: {hours}h

<promise>READY FOR EXECUTION</promise>

Next: Use /ent:task to execute
═══════════════════════════════════════════
```

---

## When to Use

**Use `/ent:plan`** for:
- New features requiring design
- Breaking changes
- Architecture changes
- Multi-component changes
- Complex refactoring

**Skip planning** for:
- Simple bug fixes
- Typos, formatting
- Documentation only
- Configuration tweaks

---

## Example Session

```
User: /ent:plan "Add Redis caching for user queries"

@ent:planner-fast: Quick assessment...
  ✅ Feature is clear, medium complexity, proceeding

@ent:planner: Clarifying...
  Q: Which queries should be cached?
  Q: What's the TTL strategy?
  Q: Cache invalidation approach?
[User answers]

@ent:planner: Researching...
  Evaluated: go-redis/redis vs redis/rueidis
  ✓ Recommendation: rueidis (better performance)

@ent:architect: Designing...
  Created proposal.md and design.md
  - Cache layer architecture
  - Repository wrapper pattern
  - Invalidation strategy

@ent:planner: Writing specs...
  Created specs/user-caching/
  6 requirements added, 2 modified
  ✅ Validation PASS

@ent:decomposer: Breaking down...
  12 tasks total (4 parallel)
  ~16 hours estimated
  Critical path: T1→T3→T5→T9

<promise>READY FOR EXECUTION</promise>

Change ID: add-user-caching
Next: /ent:task
```

---

## Integration with Registry

After planning:
1. `openspec validate {change-id} --strict` passes
2. Tasks auto-sync to `openspec/registry.yaml`
3. Use `registry list` to see all tasks
4. Use `registry next` for next unblocked task
5. Execute with `/ent:task` command
