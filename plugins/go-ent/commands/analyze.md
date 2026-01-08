---
description: Cross-document consistency validation (read-only)
argument-hint: <change-id>
---

# Consistency Analysis

Validate consistency across proposal, design, tasks, and spec deltas.

## Input

Change ID: $ARGUMENTS (from `openspec list`)

## Path Resolution

Change directory: `openspec/changes/$ARGUMENTS/`

For the steps below, `$CHANGE_ROOT` refers to `openspec/changes/$ARGUMENTS/`.

## Read-Only Operation

This command only reads and reports. No files are modified.

## Steps

1. Validate change exists: `openspec show $ARGUMENTS`
2. Resolve change directory path (see Path Resolution above)
3. Read all change artifacts:
   - `$CHANGE_ROOT/proposal.md`
   - `$CHANGE_ROOT/design.md` (if exists)
   - `$CHANGE_ROOT/tasks.md`
   - `$CHANGE_ROOT/specs/` - Spec deltas
3. Perform consistency checks:
   - **Coverage**: All requirements have corresponding tasks
   - **References**: Task story references point to valid requirements
   - **Files**: File paths in tasks exist or are clearly new
   - **Alignment**: Design decisions match spec scenarios
   - **Uniqueness**: No orphaned or duplicate tasks
   - **Dependencies**: Task graph forms valid DAG (no cycles)
4. Calculate coverage metrics
5. Identify issues with severity (High/Medium/Low)
6. Generate recommendations

## Checks Performed

### 1. Requirement Coverage
- Every spec requirement should have implementation tasks
- Every task should reference at least one requirement

### 2. Reference Validation
- Task story references match actual requirement headings
- File paths are realistic (existing or clearly new files)

### 3. Consistency
- Design decisions align with spec scenarios
- Proposal "what changes" matches spec deltas
- Task breakdown matches proposal scope

### 4. Task Graph Validation
- No circular dependencies (T001 → T002 → T001)
- Dependencies reference valid task IDs
- Parallelization markers [P] used correctly

## Output Format

```markdown
# Consistency Analysis: <change-id>

## Summary
- Requirements: 5 | Tasks: 12 | Coverage: 100%
- Issues Found: 2

## Coverage Matrix
| Requirement | Tasks | Status |
|-------------|-------|--------|
| Two-Factor Auth | T001-T004 | ✓ OK |
| OTP Email | T005-T007 | ✓ OK |
| Recovery Codes | - | ✗ MISSING |

## Issues

### Issue 1: Missing Task Coverage (High)
**Requirement**: specs/auth/spec.md#recovery-codes
**Problem**: No implementation tasks defined
**Suggestion**: Add tasks for backup code generation and storage

### Issue 2: Invalid Reference (Medium)
**Task**: T003
**Problem**: References `specs/auth/spec.md#sms-delivery` but requirement not found
**Suggestion**: Check requirement name or add missing spec

## Recommendations
1. Add tasks for "Recovery Codes" requirement
2. Fix reference in T003 to match actual requirement
3. Consider adding integration test tasks
```

## Validation

Run `openspec validate $ARGUMENTS --strict` after fixing issues.

## No Files Modified

All output is direct to conversation for review.
