---
name: code-review
description: Code review best practices: what to look for, how to give feedback, review checklist, and automation
---

# Code Review Excellence

## What to Review
- **Correctness**: Does the code do what it's supposed to?
- **Design**: Is the architecture appropriate? Right abstractions?
- **Readability**: Can someone new understand this easily?
- **Performance**: Any obvious inefficiencies or N+1 queries?
- **Security**: Input validation, auth checks, injection risks?
- **Testing**: Adequate test coverage for the changes?
- **Edge cases**: What happens with empty input, huge input, concurrent access?

## Giving Feedback
- Be specific and actionable: "Extract this to a function because..." not "This is messy"
- Distinguish: blocking (must fix) vs suggestion (nice to have) vs question (learning)
- Prefix with labels: `[nit]`, `[suggestion]`, `[question]`, `[blocking]`
- Explain WHY, not just what: "This avoids N+1 queries" not "Use a JOIN"
- Praise good patterns: positive feedback reinforces good habits
- Keep reviews under 400 lines — request splits for larger PRs

## Review Checklist
- [ ] PR description explains what and why
- [ ] Tests added/updated for changes
- [ ] No hardcoded secrets or credentials
- [ ] Error handling is appropriate
- [ ] Database migrations are backward-compatible
- [ ] API changes are backward-compatible (or versioned)
- [ ] No N+1 queries or obvious performance issues
- [ ] Logging is appropriate (not excessive, not missing)
- [ ] Code follows project conventions

## Automation
- Linters catch style issues — don't review what can be automated
- CI runs tests — focus on logic, not "does it compile"
- CODEOWNERS for routing reviews to domain experts
- PR templates for consistent descriptions
- Automated security scanning (SAST, dependency audit)

## As a Reviewer
- Review within 24 hours
- Ask questions when you don't understand
- Approve when "good enough" — don't block on perfection
- Trust the author's context for domain-specific decisions

## As an Author
- Self-review before requesting review
- Write a clear PR description with context
- Respond to all comments, even if just "done" or "won't fix because..."
- Don't take feedback personally — it's about the code
