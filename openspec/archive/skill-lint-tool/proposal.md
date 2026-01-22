# Proposal: Skill Linting Tool

## Status: complete

## Summary
Add `skill lint` command with auto-fix capabilities for common issues and CI/CD integration.

## Problem
- Manual validation is slow
- No auto-fix for common issues
- No CI integration

## Solution
```bash
skill lint [--fix] [--json] [path]
```

Features:
- Format validation with detailed errors
- Auto-fix: format frontmatter, fix common issues
- JSON output for CI pipelines
- Exit codes for success/failure

## Breaking Changes
- [x] None - new command

## Implementation Notes

### Task 4.1: GitHub Actions Example ✅

**Created:**
1. `.github/workflows/skill-lint.yml` - GitHub Actions workflow with:
   - Validation-only job (runs on push/PR)
   - Auto-fix job (manual trigger)
   - JSON output and artifact uploads
   - Exit code handling for CI failures

2. `docs/SKILL_LINT_CI.md` - Comprehensive documentation covering:
   - When to use validation-only vs auto-fix
   - Exit codes and their meanings
   - Workflow customization guide
   - Troubleshooting common issues
   - Best practices

**Key Features:**
- Automatic validation on PRs and pushes
- Manual auto-fix workflow for maintenance
- Structured JSON output for debugging
- Artifacts retained for 30 days
- PR comments with lint results (auto-fix job)

## Alternatives
1. **External linter tool** (chosen): Dedicated command
   - ✅ Focused, composable
2. **Integrate into validate**: Add flags to existing command
   - ❌ Conflates validation and fixing
