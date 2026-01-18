# Proposal: Skill Linting Tool

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
- [ ] None - new command

## Alternatives
1. **External linter tool** (chosen): Dedicated command
   - ✅ Focused, composable
2. **Integrate into validate**: Add flags to existing command
   - ❌ Conflates validation and fixing
