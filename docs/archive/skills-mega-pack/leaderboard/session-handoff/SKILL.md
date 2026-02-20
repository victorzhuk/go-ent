---
name: session-handoff
description: Session handoff patterns for AI agents: context preservation, task continuity, and progress tracking
---

# Session Handoff

## Purpose
When switching between AI agent sessions (new conversation, different model, different tool), preserve context so work continues seamlessly.

## Handoff Document Template
```markdown
# Session Handoff — [Project/Task Name]

## Current State
- Branch: `feature/xyz`
- Last commit: `abc123 — "feat: add user validation"`
- Working directory: `/path/to/project`

## What Was Done
1. Implemented user validation in `internal/handler/user.go`
2. Added test cases in `internal/handler/user_test.go`
3. Fixed N+1 query in `internal/repo/user.go`

## What's Next
1. [ ] Add integration tests for validation edge cases
2. [ ] Update API documentation
3. [ ] Run full test suite and fix any failures

## Key Decisions Made
- Using custom validator instead of library because [reason]
- Error messages follow RFC 7807 format

## Open Issues
- Rate limiter middleware needs Redis configuration
- Migration 003 needs review before merging

## Files Changed
- `internal/handler/user.go` — added validation
- `internal/handler/user_test.go` — new test file
- `internal/repo/user.go` — query optimization

## Context the Next Agent Needs
- Using Go 1.22+ stdlib routing (no framework)
- Tests use explicit comparisons, no assertion libraries
- Follow existing patterns in `internal/handler/order.go` for reference
```

## Best Practices
- Create handoff documents at natural breakpoints
- Include enough context for a new agent to continue without re-reading everything
- List specific files changed and their purpose
- Document decisions and their rationale
- Include "what's next" with clear, actionable items
- Reference existing patterns for consistency
