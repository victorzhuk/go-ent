---
description: Execute tasks with TDD and validation
---

# Flow: Task Execution

{{include "domains/openspec.md"}}

Execute tasks from tracking system with test-driven development and validation.

## Agent Chain

| Agent            | Phase                              | Tier     |
|------------------|------------------------------------|----------|
| @ent:task-fast   | Quick assessment, routing          | fast     |
| @ent:task-heavy  | Complex task analysis              | heavy    |
| @ent:coder       | Code implementation                | standard |
| @ent:reviewer    | Code review (optional)             | standard |
| @ent:tester      | Test creation and validation       | fast     |
| @ent:acceptor    | Acceptance validation              | fast     |

**Escalation**: task-fast → (task-heavy if complex) → coder → reviewer → tester → acceptor

---

## Workflow

### Phase 1: Assessment

**Agent**: @ent:task-fast

**Goal**: Quick task assessment and routing

**Steps**:
1. Load task from tracking system
2. Assess complexity (see escalation triggers)
3. Load change context (proposal, design, requirements)
4. **Decision**: Proceed to @ent:coder or escalate to @ent:task-heavy

**Escalation to @ent:task-heavy if**:
- Task requires algorithm design
- Security-critical implementation
- Multiple integration points (>2)
- Unclear requirements after initial analysis
- Complexity score > 0.8
- Previous attempt failed

### Phase 2: Deep Analysis (Conditional)

**Agent**: @ent:task-heavy

**Goal**: Clarify complex requirements, design approach

**Steps**:
1. Analyze ambiguous requirements
2. Design implementation approach
3. Identify risks and edge cases
4. Document clarified requirements
5. Hand off to @ent:coder with clear spec

### Phase 3: Implementation

**Agent**: @ent:coder

**Goal**: Implement task with TDD

**For test tasks (TDD cycle)**:
1. Write failing tests first (RED)
2. Implement minimal solution (GREEN)
3. Refactor and clean up
4. Run validation

**For implementation tasks**:
1. Use code navigation tools for context
2. Follow project conventions
3. Implement requirements
4. Write tests alongside code
5. Run build and test validation

### Phase 4: Review (Conditional)

**Agent**: @ent:reviewer

**Condition**: When changes are non-trivial or touch critical paths

**Goal**: Code quality review

**Outcome**:
- **APPROVED** → Continue to @ent:tester
- **CHANGES_REQUESTED** → Return to @ent:coder with specific fixes

### Phase 5: Testing

**Agent**: @ent:tester

**Goal**: Validate implementation with tests

**Steps**:
1. Run test suite with race detector
2. Verify coverage >= 80% for new code
3. Check edge cases covered
4. **Decision**: PASS or FAIL

**If FAIL**: Return to @ent:coder with failure details

### Phase 6: Acceptance

**Agent**: @ent:acceptor

**Goal**: Final validation against spec

**Steps**:
1. Load spec scenarios
2. Verify all acceptance criteria met
3. Check non-regression
4. Sign off or reject

**Outcome**:
- **ACCEPTED** → Mark task complete
- **NEEDS_WORK** → Return to @ent:coder
- **REJECTED** → Escalate to architect

### Phase 7: Complete

Update tracking system:
- Mark task as completed
- Add completion notes
- Update progress

---

## Output Format

```
═══════════════════════════════════════════
TASK: {task-id}
═══════════════════════════════════════════

📋 Task: {description}
   Change: {change-id}
   Priority: {priority}
   Dependencies: {count} (all complete)

🔨 Implementation:
   Files modified: {count}
   Lines added: +{num}
   Lines removed: -{num}

🧪 Testing:
   Tests written: {count}
   Coverage: {percent}%
   Race detector: PASS

✅ Validation:
   Build: PASS
   Tests: PASS ({passed}/{total})
   Lint: PASS

<promise>COMPLETE</promise>

Progress: {percent}% ({completed}/{total} tasks)
Next: {next-task-id} (priority: {level})
═══════════════════════════════════════════
```
