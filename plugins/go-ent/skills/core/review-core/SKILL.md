---
name: review-core
description: "Code review best practices and checklist. Auto-activates for: pull request reviews, code quality checks, style violations, best practices enforcement."
version: 1.0.0
---

# Code Review Core

## Review Mindset

- Be kind and constructive
- Explain the "why", not just "what's wrong"
- Suggest alternatives
- Focus on code, not author
- Recognize good work

## Essential Checklist

### Functionality
- Does what PR claims
- Edge cases handled
- Errors handled properly
- No obvious bugs

### Design
- Follows project architecture
- Appropriate abstraction
- DRY and SOLID principles
- Not over-engineered

### Readability
- Self-documenting (clear names)
- Comments for "why", not "what"
- Consistent style
- Functions small and focused
- No magic numbers

### Tests
- Tests present and passing
- Edge cases covered
- Deterministic (no flaky tests)
- Fast execution

### Performance
- No obvious bottlenecks
- Efficient algorithms
- Proper resource management
- Caching where appropriate

### Security
- Input validation
- No SQL injection, XSS
- Secrets not in code
- Auth/authz correct

## Review Priorities

| Priority | Focus | Action |
|----------|-------|--------|
| P0 | Bugs, security issues | Block merge |
| P1 | Design flaws, maintainability | Request changes |
| P2 | Style, minor improvements | Suggest, don't block |
| P3 | Nitpicks, preferences | Comment, approve anyway |

## Common Issues

| Issue | Detection | Fix |
|-------|-----------|-----|
| N+1 queries | Multiple DB calls in loop | Use batch loading |
| Race conditions | Shared state without sync | Add mutex/channels |
| Memory leaks | Goroutines never stop | Add context cancellation |
| Error swallowing | Empty catch blocks | Log or propagate |
| Tight coupling | Hard dependencies | Use interfaces |

## Giving Feedback

**Good feedback**:
```
This could lead to a race condition when multiple requests
modify the cache concurrently. Consider using sync.RWMutex
to protect the map access.
```

**Bad feedback**:
```
This is wrong. Use mutex.
```

## Reviewing Changes

1. Read PR description thoroughly
2. Understand the problem being solved
3. Review design/approach first
4. Then dive into implementation
5. Test locally if unclear
6. Check tests before approval
7. Approve or request changes with clear reasoning
