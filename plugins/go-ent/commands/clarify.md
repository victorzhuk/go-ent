---
description: Ask focused questions to clarify underspecified requirements
argument-hint: <change-id>
---

# Requirement Clarification

Generate structured clarifying questions for underspecified or ambiguous requirements.

## Input

Change ID: $ARGUMENTS (from `openspec list`)

## Path Resolution

Change directory: `openspec/changes/$ARGUMENTS/`

For the steps below, `$CHANGE_ROOT` refers to `openspec/changes/$ARGUMENTS/`.

## Scope

Analyze across:
- `$CHANGE_ROOT/proposal.md` - Problem statement and changes
- `$CHANGE_ROOT/design.md` - Technical decisions
- `$CHANGE_ROOT/specs/` - Spec deltas with requirements and scenarios

## Steps

1. Validate change exists: `openspec show $ARGUMENTS`
2. Resolve change directory path (see Path Resolution above)
3. Read all change artifacts
3. Identify ambiguities across 9 categories:
   - Functional Scope - What exactly should/shouldn't be built?
   - Domain & Data Model - Entity definitions, relationships
   - Interaction & UX Flow - User journeys, edge cases
   - Non-Functional Attributes - Performance, security, reliability
   - Integration & Dependencies - External systems, APIs
   - Edge Cases - Error scenarios, boundary conditions
   - Constraints & Tradeoffs - Technical/business limitations
   - Terminology - Unclear terms, assumptions
   - Completion Signals - When is it done? Success criteria?
4. Prioritize by implementation impact (High/Medium/Low)
5. Select top 5 questions maximum
6. Format as structured questions with context

## Output Format

```markdown
# Clarification Needed: <change-id>

Found <N> areas requiring clarification:

## Question 1 (High Impact)
**Source**: proposal.md, line 12
**Context**: "Users should be able to configure settings"
**Question**: What specific settings should be configurable?
- Option A: ...
- Option B: ...
- Something else?

## Question 2 (Medium Impact)
**Source**: specs/auth/spec.md
**Context**: No scenario for rate limiting
**Question**: Should OTP attempts be rate-limited? If so, what limits?

---
Reply with answers to refine the proposal.
```

## Guardrails

- Maximum 5 questions per invocation
- Each question should be actionable (not philosophical)
- Include enough context for informed decision
- Prioritize questions that block implementation
- Don't ask about details that can be inferred from patterns

## No Files Created

Output is direct to conversation for immediate user feedback.
