---
description: Comprehensive planning workflow with research, design, and task decomposition
argument-hint: <feature-description-or-change-id>
allowed-tools: Read, Bash, Edit, mcp__plugin_serena_serena, mcp__go_ent__workflow_start, mcp__go_ent__workflow_status
---

# Planning Workflow

Complete planning workflow integrating research, clarification, task decomposition, and quality validation with explicit wait points for user approval.

## Input

$ARGUMENTS can be:
- **New feature**: Description of what to build
- **Existing change**: Change ID from `openspec list`

## Path Resolution

Change directory: `openspec/changes/$ARGUMENTS/`

For the steps below, `$CHANGE_ROOT` refers to `openspec/changes/$ARGUMENTS/`.

## Workflow Management

**IMPORTANT**: Use workflow state tracking for guided execution:

1. **Start workflow**: `workflow_start` with change_id and phase="discovery"
2. **At each wait point**: Save state and wait for user approval
3. **User approves**: Workflow continues to next phase

## Three-Phase Process with Wait Points

### Phase 0: Clarification & Research

**Goal**: Resolve all unknowns before proceeding

1. Start workflow tracking
2. If new feature:
   - Run `/openspec:proposal` with description
   - Get change ID for subsequent steps
3. Run `/go-ent:clarify <change-id>`
4. **WAIT POINT 1**: Present findings and clarifying questions
   - Use workflow state to mark: `wait_point="user-clarification"`
   - Stop execution, present questions to user
   - User provides answers
   - Resume after approval
5. Update proposal/design based on answers
6. Run `/go-ent:research <change-id>`
7. Conduct research for all identified unknowns
8. **WAIT POINT 2**: Present research findings
   - Use workflow state to mark: `wait_point="research-review"`
   - Present: technology choices, alternatives, recommendations
   - User approves approach
9. **VERIFY**: No "NEEDS CLARIFICATION" or "TBD" items remain
10. If blockers exist, escalate to user for decisions

### Phase 1: Design & Contracts

**Goal**: Create complete specifications with scenarios

1. Update workflow phase to "design"
2. Review/update `$CHANGE_ROOT/proposal.md`:
   - Clear "why" and "what changes"
   - Mark breaking changes
   - Identify affected systems
3. Create/update `$CHANGE_ROOT/design.md` (when needed):
   - Architectural decisions
   - Data model changes
   - Integration points
   - Migration strategy
4. Draft spec deltas in `$CHANGE_ROOT/specs/`:
   - One capability per directory
   - Use `## ADDED|MODIFIED|REMOVED Requirements`
   - Each requirement has `#### Scenario:` with WHEN/THEN
   - Cross-reference related capabilities
5. Validate: `openspec validate <change-id> --strict`
6. Fix all validation errors before proceeding
7. **WAIT POINT 3**: Present design for approval
   - Use workflow state to mark: `wait_point="design-approval"`
   - Present: architecture decisions, data models, API contracts
   - User reviews and approves design
   - Resume after approval

### Phase 2: Task Generation & Validation

**Goal**: Create implementable, validated task plan

1. Update workflow phase to "planning"
2. Run `/go-ent:decompose <change-id>`
   - Generate task IDs and dependency graph
   - Mark parallelizable tasks
   - Associate file paths
3. Run `/go-ent:analyze <change-id>`
   - Verify requirement coverage
   - Check consistency
   - Validate task graph
4. Fix any issues found
5. Final validation: `openspec validate <change-id> --strict`
7. **WAIT POINT 4 (FINAL)**: Present complete plan for approval
   - Use workflow state to mark: `wait_point="plan-approval"`
   - Present: full task breakdown, priorities, dependencies
   - Estimated effort per task
   - User approves plan
   - Mark workflow as completed after approval

## Success Criteria

Plan is ready for `/openspec:apply` when:

- [ ] All clarifying questions answered
- [ ] Research complete (no TBD items)
- [ ] Design decisions documented
- [ ] Spec deltas validated
- [ ] Tasks have IDs, dependencies, file paths
- [ ] Consistency analysis passes
- [ ] `openspec validate --strict` passes
- [ ] User has approved the plan

## Example Workflow

```bash
# Start with new feature
/go-ent:plan "Add two-factor authentication with OTP"

# Phase 0
→ Creates proposal, gets change-id: add-2fa
→ Runs /go-ent:clarify add-2fa
→ Asks 5 questions
← User answers questions
→ Runs /go-ent:research add-2fa
→ Evaluates OTP libraries, SMS providers
→ Updates research.md

# Phase 1
→ Updates proposal.md and design.md
→ Creates spec deltas with scenarios
→ Validates with openspec

# Phase 2
→ Runs /go-ent:decompose add-2fa
→ Generates tasks with IDs T001-T015
→ Runs /go-ent:analyze add-2fa
→ Checks coverage: 100%
→ Final validation passes

# Ready for implementation
/openspec:apply add-2fa
```

## When to Use

**Use `/go-ent:plan`** for:
- New features (not bug fixes)
- Breaking changes
- Architecture changes
- Performance optimizations changing behavior
- Security pattern updates

**Use simpler workflow** for:
- Bug fixes (just `/openspec:proposal`)
- Typos, formatting
- Non-breaking dependency updates
- Configuration changes

## Guardrails

- Don't skip user approval between phases
- Don't proceed with TBD items in research
- Don't implement until validation passes
- Don't skip clarification for complex changes

## Output

Fully validated change proposal ready for implementation with:
- `$CHANGE_ROOT/proposal.md`, `$CHANGE_ROOT/design.md`
- `$CHANGE_ROOT/research.md` with decisions
- `$CHANGE_ROOT/tasks.md` with task IDs and dependencies
- Spec deltas in `$CHANGE_ROOT/specs/` validated by OpenSpec

Output directory: `openspec/changes/$ARGUMENTS/`

## Next Step

After approval: `/openspec:apply <change-id>`
