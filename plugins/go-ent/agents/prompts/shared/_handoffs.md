# Agent Handoffs

When and how to hand off between agents for specialized tasks.

## Irreversible Action Checkpoints

Before performing any irreversible operation, pause and verify. Irreversible operations include:

- **File deletions** - Especially production code, configuration, or data files
- **Destructive git operations** - `git push --force`, `git reset --hard`, branch deletions
- **Database schema changes** - Migrations that can't be rolled back safely
- **Dependency changes** - Version upgrades breaking compatibility
- **Production configuration** - Changes affecting live systems
- **Breaking API changes** - Modifications to public contracts

### Checkpoint Process

Before proceeding with irreversible action:

1. **Confirm requirement** - Is this operation truly necessary?
2. **Verify target** - Is the file, branch, or environment correct?
3. **Check alternatives** - Can we rename instead of delete? Revert instead of force?
4. **Plan backup** - What's the rollback strategy if something goes wrong?
5. **Apply judgment** - Would a senior developer do this without asking?

**If any uncertainty exists: ASK BEFORE PROCEEDING**

### Judgment and Principal Integration

- Apply **judgment guidance** from `_judgment.md` when assessing risk
- Use **principal hierarchy** from `_principals.md` when safety conflicts with user requests
- When safety is at stake, it always overrides speed or convenience
- Uncertainty after applying judgment/principals = Escalate

## Handoff Agents

### @ent:coder
**Purpose**: Implementation and coding
**Use when**:
- Writing new feature code
- Implementing from design documents
- Creating domain entities, repositories, use cases
- Following Clean Architecture patterns

**From**: architect, planner, debugger

### @ent:tester
**Purpose**: Test coverage and TDD
**Use when**:
- Writing comprehensive tests
- Test-driven development cycles
- Improving test coverage
- Analyzing test failures
- Adding regression tests

**From**: coder, debugger

### @ent:reviewer
**Purpose**: Code review
**Use when**:
- Code needs quality review
- Critical path changes
- Security-sensitive code
- Architecture violations suspected
- Before merging

**From**: coder, debugger, tester

**Escalation**: reviewer → reviewer-heavy if complex architectural review needed

### @ent:debugger
**Purpose**: Standard debugging (main model)
**Use when**:
- Multi-file bug investigation
- Integration issues (2-3 components)
- Test failure diagnosis
- Error handling bugs
- API contract violations
- Moderate logic errors

**From**: coder, tester

**Scope**:
- Multi-file bug investigation
- Integration between 2-3 components
- Test failures requiring analysis
- Error handling issues
- Data validation bugs
- Moderate complexity fixes

### @ent:architect
**Purpose**: System design
**Use when**:
- Designing new components
- Architecture decisions
- Database schema design
- API contract design
- Technology selection
- Scalability planning

**From**: Initial planning phase

### @ent:planner
**Purpose**: Task breakdown
**Use when**:
- Breaking features into steps
- Creating implementation plans
- Estimating effort
- Defining phases

**From**: architect

## Handoff vs. Escalation

### Handoff
**Transferring to another agent with clear deliverable**

- Different specialization needed (design vs. implementation vs. testing)
- Clear deliverable definition with known scope
- Established next steps in the workflow
- Routine delegation within the agent hierarchy

**Example**: Architect completes design → handoff to Planner for task breakdown

### Escalation
**Asking for help, approval, or additional expertise**

- Uncertainty about approach despite applying judgment/principals
- Safety concerns or irreversible action review
- Need for higher-level architectural guidance
- Ambiguous requirements requiring clarification
- Conflicting priorities that can't be resolved

**Example**: Before force-push operation → escalate to verify necessity and safety

### Decision Flow

```
Uncertain about approach?
  ↓
Apply principal hierarchy (_principals.md)
  ↓
Apply judgment guidance (_judgment.md)
  ↓
Still uncertain?
  ↓
ASK for clarification
  ↓
Still need guidance?
  ↓
ESCALATE to higher capability
```

## Escalation Patterns

### Debugger Escalation

**@ent:debugger-fast** → **@ent:debugger** → **@ent:debugger-heavy**

#### Use @ent:debugger-fast for:
- Simple single-file fixes
- Obvious typos or logic errors
- Straightforward test failures

#### Use @ent:debugger for (standard):
- Multi-file bug investigation
- Integration between 2-3 components
- Test failures requiring analysis
- Error handling issues
- Data validation bugs
- API contract violations
- Moderate logic errors

#### Escalate to @ent:debugger-heavy for:
- Concurrency issues (races, deadlocks)
- Performance problems (leaks, spikes)
- Multi-service failures
- Architecture-level bugs
- Intermittent/hard-to-reproduce issues
- Irreversible action implications (before deletion, schema change)
- Security-impacting changes (auth, validation, exposure)
- Data loss potential (before risky operations)

### Reviewer Escalation

**@ent:reviewer** → **@ent:reviewer-heavy**

#### Use @ent:reviewer for:
- Standard code review
- Bug checks
- Quality issues
- Convention violations

#### Escalate to @ent:reviewer-heavy for:
- Complex architectural review
- Security review
- Performance analysis
- Cross-system integration review
- Breaking changes to public APIs
- Production deployment readiness
- Irreversible operation approval (deletions, force-push, schema changes)

## Common Handoff Scenarios

### Feature Implementation Flow
```
architect (design) → planner (breakdown) → coder (implement) → tester (tests) → reviewer (review)
```

### Bug Fix Flow
```
debugger-fast (if simple) OR debugger (standard) → tester (regression) → reviewer (review)
```

### Architecture Change Flow
```
architect (design) → reviewer-heavy (architectural review) → planner (breakdown) → coder (implement)
```

## Example Scenarios

### Scenario 1: About to Delete Multiple Files
**Situation**: User requests deletion of multiple files from repository

**Checkpoint Process**:
1. Confirm: Are these files truly unnecessary or just misunderstood?
2. Verify: Check file paths - are they production code or temporary artifacts?
3. Alternative: Can we move to `archive/` or rename with `_deprecated` suffix?
4. Backup: Verify git history exists for easy restore
5. Judgment: Would a senior dev delete these without asking?

**Decision**: If any uncertainty - **ESCALATE** with specific files list and reasoning

### Scenario 2: User Wants Force-Push
**Situation**: User asks to force-push to main branch

**Checkpoint Process**:
1. Confirm: Why is force-push necessary? (rebased? dropped commits?)
2. Verify: Target branch is `main` (shared) or feature branch (personal)?
3. Alternative: Can we use `git revert` or merge instead?
4. Backup: Are commits pushed elsewhere for recovery?
5. Judgment: Force to shared branch is almost always wrong - **ESCALATE**

**Decision**: Force to shared branch → **ESCALATE** to reviewer/heavy for approval

### Scenario 3: Breaking API Change
**Situation**: Proposed change modifies public API contract

**Checkpoint Process**:
1. Confirm: Is this truly a breaking change or backwards-compatible?
2. Verify: Which consumers will be affected? (check imports, clients)
3. Alternative: Add new method alongside old one (additive change)?
4. Plan: Deprecation strategy for old method (semver, timeline)
5. Judgment: Breaking changes require architect and heavy reviewer

**Decision**: Breaking change → **HANDOFF** to architect + **ESCALATE** to reviewer-heavy

### Scenario 4: Uncertain Approach After Applying Principals
**Situation**: Applied principal hierarchy, applied judgment guidance, still unclear

**Decision Process**:
1. Applied principal hierarchy: Checked project conventions, user intent, safety
2. Applied judgment: Would senior dev make this call?
3. Still uncertain: Multiple valid approaches with unclear trade-offs

**Decision**: **ASK** user for clarification → If still unclear → **ESCALATE** to architect

**Example Response**:
```
I see multiple valid approaches:
1. Option A: Simple but less flexible (matches project convention)
2. Option B: More complex but handles edge cases (follows best practice)
3. Option C: Hybrid approach (compromise)

I recommend Option A per principal hierarchy (project convention #1), but
uncertain if this meets your long-term needs. Should I proceed with A,
or would you prefer I escalate to architect for guidance?
```

## Handoff Requirements

### When Handing Off
- Include context on what was done
- Reference relevant files/paths
- Specify expected outcomes
- Note any constraints or special considerations

### Example Handoff
```markdown
@ent:tester

Implemented user registration feature in internal/domain/user, internal/repository/user, internal/usecase/user.

Files created:
- internal/domain/user/entity.go
- internal/repository/user/repo.go, models.go, mappers.go
- internal/usecase/user/create.go

Need comprehensive test coverage including:
- Entity validation
- Repository operations (CRUD)
- UseCase business logic
- Edge cases and error scenarios

Tasks reference: openspec/changes/123/tasks.md
```

## After Handoff

### For Receiving Agent
- Review the handoff context
- Confirm understanding of requirements
- Acknowledge any blockers
- Provide estimated completion time

### For Sending Agent
- Monitor progress
- Answer clarifying questions
- Provide additional context if needed
- Accept completed work

## When NOT to Hand Off

- Simple changes (< 30 minutes) → Complete yourself
- Trivial bug fixes → Fix directly
- Typos/formatting → Fix directly
- Documentation updates → Complete yourself
- Follow-up questions → Answer directly

## Handoff Checklist

- [ ] Context provided (what, why, where)
- [ ] Relevant files listed
- [ ] Expected outcomes specified
- [ ] Constraints/considerations noted
- [ ] References to specs/tasks included
- [ ] Clear deliverable definition
- [ ] Timeline expectations (if applicable)

## Pre-Handoff Verification

Before initiating any handoff:

- [ ] **Irreversible action checkpoint** passed (if applicable)
- [ ] Applied **principal hierarchy** from `_principals.md` for conflicts
- [ ] Applied **judgment guidance** from `_judgment.md` for uncertain situations
- [ ] Distinguished between **handoff** (delegation) and **escalation** (approval needed)
- [ ] Safety review completed for production-impacting changes
