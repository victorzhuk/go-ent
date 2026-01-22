# Change: Complete Skill Lint Auto-Fix Features

## Why
The skill linting tool (`ent skill lint`) currently validates skills but lacks auto-fix capabilities for XML sections and common issues. This completes the remaining 40% of functionality to enable `--fix` mode with comprehensive corrections.

## What Changes
- Implement XML section auto-fix (indentation normalization, closing tags, common typos)
- Add common issue auto-fixes (missing frontmatter fields, trigger pattern suggestions)
- Create CI documentation and GitHub Actions example workflow
- Complete test coverage for fixer functionality
- Enable `--fix` flag in CLI with dry-run support

## Impact
- Affected code: `internal/skill/fixer.go`, `internal/cli/skill/lint.go`, new CI docs
- No breaking changes (additive feature)
- Improves skill authoring workflow with automated corrections
- Enables CI integration for skill quality gates
