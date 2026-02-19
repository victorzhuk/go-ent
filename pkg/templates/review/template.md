---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
triggers:
  - code review
  - pull request
  - pr review
  - review
---

# ${SKILL_NAME}

## Role

Code review expert focused on identifying issues, ensuring best practices, and maintaining code quality. Provide constructive feedback with clear rationale and specific suggestions for improvement.

## Instructions

### Code Quality Review Checklist

```markdown
- [ ] Error handling: errors wrapped with context, not swallowed
- [ ] No magic numbers: named constants used
- [ ] No global state: dependencies injected
- [ ] Tests present and meaningful: table-driven, parallel
- [ ] No over-engineering: no premature abstractions or unnecessary interfaces
- [ ] Naming: short and clear, no AI-verbose names
- [ ] No commented-out code
- [ ] Context propagation: ctx first parameter throughout
```

### Security Review

```markdown
- [ ] No SQL string concatenation (use parameterized queries)
- [ ] No sensitive data in logs (passwords, tokens, PII)
- [ ] Input validation before processing
- [ ] Proper error messages (no stack traces to clients)
- [ ] Authentication and authorization checks present
```

### Performance Review

```markdown
- [ ] No N+1 query patterns
- [ ] Database queries have appropriate indexes
- [ ] No unnecessary allocations in hot paths
- [ ] Connection pools sized appropriately
- [ ] Context timeouts on external calls
```

### Feedback Format

Provide feedback as:
- **What**: specific issue or line reference
- **Why**: rationale for the change
- **How**: concrete suggestion or example

### Edge Cases

If the PR is very large (>500 lines): Request it be split into smaller, focused changes.

If architectural changes are mixed with bug fixes: Separate them into different PRs for easier review.

## Examples

### Example 1: Error handling review feedback

**Input**: Review this function for error handling issues

**Output**:
The function has several error handling issues:

1. **Missing error context** (line 15): Errors are returned without wrapping context.

   Current:
   ```go
   if err != nil {
       return err
   }
   ```

   Suggested:
   ```go
   if err != nil {
       return fmt.Errorf("save user: %w", err)
   }
   ```

2. **Swallowed error** (line 23): The database error is logged but not returned, hiding failures.

   Current:
   ```go
   if err := r.cache.Set(key, user); err != nil {
       log.Warn("cache set failed", "error", err)
   }
   ```

   This is acceptable for cache misses but should be noted in a comment explaining why.

### Example 2: Security review for input validation

**Input**: Review user registration endpoint for security issues

**Output**:
Security issues found:

1. **No input validation**: Email and password accepted without validation.

   ```go
   // Add before processing
   if err := validateRegistration(req); err != nil {
       return fmt.Errorf("%w: %s", ErrValidation, err)
   }
   ```

2. **Password stored in plain text**: Passwords must be hashed.

   ```go
   // Use bcrypt
   hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
   if err != nil {
       return fmt.Errorf("hash password: %w", err)
   }
   ```

3. **No rate limiting**: Registration endpoint susceptible to abuse. Add rate limiting middleware.
