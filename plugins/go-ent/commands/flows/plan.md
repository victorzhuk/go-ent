---
description: Complete planning workflow with research and task breakdown
---

# Flow: Planning

{{include "domains/openspec.md"}}

Complete planning workflow: clarify → research → design → decompose.

## Agent Chain

| Agent           | Phase                          | Tier     |
|-----------------|--------------------------------|----------|
| @ent:planner-fast | Initial assessment            | fast     |
| @ent:architect  | Architecture and design        | heavy    |
| @ent:planner    | Detailed planning              | standard |
| @ent:decomposer | Task breakdown                | standard |

**Escalation**: planner-fast → architect → planner → decomposer

---

## Workflow

### Phase 1: Initial Assessment

**Agent**: @ent:planner-fast

**Goal**: Quick feasibility check

**Steps**:
1. Parse feature description
2. Check: Is it clear enough? Are there blockers? What's the complexity?
3. Identify immediate unknowns
4. **Decision**: Proceed or request clarification

**If unclear**: Ask focused questions before continuing

### Phase 2: Clarification

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
4. Document answers

**Wait for user input** if critical unknowns exist

### Phase 3: Research & Technology Decisions

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
- Existing codebase patterns
- Official docs
- Community practices
- Project conventions

**Present findings and get approval** before design

### Phase 4: Architecture & Design

**Agent**: @ent:architect

**Goal**: Create detailed design documents

**Principles**:
- Follow existing patterns unless there's a reason
- Prefer simple over clever
- Design for testability
- Consider failure modes
- Document trade-offs

**Present design and get approval** before task breakdown

### Phase 5: Specification

**Agent**: @ent:planner

**Goal**: Create detailed requirements

**Requirements must have**:
- Clear acceptance criteria
- Concrete scenarios (WHEN/THEN)
- Cross-references to related requirements

**Validate** specifications

### Phase 6: Task Decomposition

**Agent**: @ent:decomposer

**Goal**: Break down work into executable tasks

**Guidelines**:
- Break into <4h chunks
- Specify exact files
- Mark dependencies explicitly
- Identify parallel work
- Estimate effort
- Group by layer/capability

**Sync** tasks to tracking system

### Phase 7: Validation

**Final checks**:
```
1. Completeness:
   - [ ] All TBD/TODO resolved
   - [ ] All unknowns researched
   - [ ] All questions answered
   - [ ] All tasks have IDs and files
   - [ ] Dependencies form valid graph (no cycles)
2. Coverage:
   - Each requirement has tests
   - Each spec has implementing tasks
   - Each task traces to requirement
3. Consistency:
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

Next: Execute tasks
═══════════════════════════════════════════
```

---

## When to Use

**Use planning workflow** for:
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
