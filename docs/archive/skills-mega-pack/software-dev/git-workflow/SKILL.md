---
name: git-workflow
description: Git best practices, branching strategies, conventional commits, code review, and CI/CD integration
---

# Git Workflow

## Conventional Commits
```
<type>(<scope>): <description>

feat(auth): add JWT refresh token rotation
fix(api): handle null response from payment gateway
docs(readme): add deployment instructions
refactor(user): extract validation to value object
test(order): add integration tests for checkout flow
chore(deps): upgrade Go to 1.22
perf(query): add composite index for user search
```

## Branching Strategy (Trunk-Based)
- `main` — always deployable, protected
- `feature/<ticket>-<description>` — short-lived feature branches
- `fix/<ticket>-<description>` — bug fixes
- Merge via squash or rebase — keep history clean
- Delete branches after merge
- Use feature flags for long-running features

## Code Review Best Practices
- Keep PRs small (< 400 lines changed)
- Write descriptive PR titles and descriptions
- Include context: what, why, how, testing done
- Review for: correctness, clarity, performance, security
- Approve or request changes — don't leave hanging

## Pre-commit Hooks
- Lint and format on commit
- Run unit tests on push
- Validate commit message format
- Check for secrets/credentials

## CI/CD Integration
- Run full test suite on every PR
- Require passing CI before merge
- Auto-deploy main to staging
- Use release tags for production deployments
