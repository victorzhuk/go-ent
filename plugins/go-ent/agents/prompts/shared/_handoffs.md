# Agent Handoffs

When and how to hand off between agents for specialized tasks.

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
