---
name: ent:task-flow
description: Execute tasks with TDD and validation
---

# Flow: Task Execution

{{include "context/openspec.md"}}

Execute tasks from tracking system with test-driven development and validation.

## Delegation Strategy

| Phase | Approach | If Subagent: Model + Type |
|-------|----------|---------------------------|
| Assessment | Inline (quick triage) | haiku, Explore |
| Deep analysis | Inline or subagent | opus, general-purpose |
| Implementation | Inline (go-code skill) | sonnet, general-purpose |
| Review | Subagent (read-only) | opus, Explore |
| Testing | Inline (go-test skill) | -- |
| Acceptance | Inline or subagent | opus, Explore |

---

## Workflow

### Phase 1: Assessment

**Approach**: Inline (quick triage)

**Goal**: Quick task assessment and routing

**Steps**:
1. Load task from tracking system
2. Assess complexity (see escalation triggers below)
3. Load change context (proposal, design, requirements)
4. **Decision**: Proceed to implementation or escalate for deeper analysis

**Escalation triggers** (needs deeper analysis):
- Task requires algorithm design
- Security-critical implementation
- Multiple integration points (>2)
- Unclear requirements after initial analysis
- Previous attempt failed

### Phase 2: Deep Analysis (Conditional)

**Approach**: Inline or subagent for complex analysis

For complex tasks requiring deep reasoning:
```
Task(model: "opus", subagent_type: "general-purpose",
  prompt: "Analyze task: {description}. Clarify ambiguous requirements, design implementation approach, identify risks and edge cases. Provide clear spec for implementation.")
```

**Steps**:
1. Analyze ambiguous requirements
2. Design implementation approach
3. Identify risks and edge cases
4. Document clarified requirements

### Phase 3: Implementation

**Approach**: Inline (skills auto-activate)

**Goal**: Implement task with TDD

**For test tasks (TDD cycle)**:
1. Write failing tests first (RED)
2. Implement minimal solution (GREEN)
3. Refactor and clean up

**For implementation tasks**:
1. Use code navigation tools for context
2. Follow project conventions
3. Implement requirements
4. Write tests alongside code
5. Run build and test validation

### Phase 4: Review (Conditional)

**Approach**: Subagent (read-only)

**Condition**: When changes are non-trivial or touch critical paths

```
Task(model: "opus", subagent_type: "Explore",
  prompt: "Review changes in {files}. Check for bugs, security issues, adherence to Clean Architecture, proper error handling. Only report issues with confidence >= 80%.")
```

**Outcome**:
- **APPROVED** -> Continue to testing
- **CHANGES_REQUESTED** -> Address feedback, then re-review

### Phase 5: Testing

**Approach**: Inline (go-test skill auto-activates)

**Goal**: Validate implementation with tests

**Steps**:
1. Run test suite with race detector
2. Verify coverage >= 80% for new code
3. Check edge cases covered
4. **Decision**: PASS or FAIL

**If FAIL**: Fix the issue and re-test

### Phase 6: Acceptance

**Approach**: Inline or subagent

For formal acceptance validation:
```
Task(model: "opus", subagent_type: "Explore",
  prompt: "Validate implementation of {task-id} against spec scenarios. Check all acceptance criteria, test coverage, and non-regression.")
```

**Steps**:
1. Load spec scenarios
2. Verify all acceptance criteria met
3. Check non-regression
4. Sign off or reject

**Outcome**:
- **ACCEPTED** -> Mark task complete
- **NEEDS_WORK** -> Return to implementation
- **REJECTED** -> Escalate to planning

### Phase 7: Complete

Update tracking system:
- Mark task as completed
- Add completion notes
- Update progress

---

## Output Format

```
TASK: {task-id}

Task: {description}
   Change: {change-id}
   Priority: {priority}
   Dependencies: {count} (all complete)

Implementation:
   Files modified: {count}
   Lines added: +{num}
   Lines removed: -{num}

Testing:
   Tests written: {count}
   Coverage: {percent}%
   Race detector: PASS

Validation:
   Build: PASS
   Tests: PASS ({passed}/{total})
   Lint: PASS

COMPLETE

Progress: {percent}% ({completed}/{total} tasks)
Next: {next-task-id} (priority: {level})
```
