# Tasks: Complete Skill Lint Auto-Fix Features

## 1. Implement XML section auto-fix
- [x] 1.1 Add `FixXMLSections()` to `internal/skill/fixer.go`
- [x] 1.2 Normalize indentation within XML sections (2-space standard)
- [x] 1.3 Add missing closing tags (detect unclosed `<example>`, `<instructions>`, etc.)
- [x] 1.4 Fix common tag typos (`<instruction>` → `<instructions>`, etc.)
- [x] 1.5 Preserve content while fixing structure

## 2. Implement common issue fixes
- [x] 2.1 Add `FixCommonIssues()` method to fixer
- [x] 2.2 Auto-add missing required frontmatter fields (name, description)
- [x] 2.3 Suggest trigger patterns based on file location and name
- [x] 2.4 Fix YAML frontmatter syntax errors (quotes, indentation)

## 3. Add CI documentation
- [x] 3.1 Create `.github/workflows/skill-lint.yml` example workflow
- [x] 3.2 Document CI integration in `docs/SKILL_LINT_CI.md`
- [x] 3.3 Add usage examples for `--fix` and `--dry-run` flags
- [x] 3.4 Document exit codes and CI failure modes

## 4. Test auto-fix functionality
- [x] 4.1 Add tests for `FixXMLSections()` in `internal/skill/fixer_test.go`
- [x] 4.2 Add tests for `FixCommonIssues()`
- [x] 4.3 Run auto-fix on all skills in `plugins/go-ent/skills/` (validation test)
- [x] 4.4 Add integration test for `--fix` CLI flag

## 5. Complete CLI integration
- [x] 5.1 Wire up `--fix` flag to call fixer methods
- [x] 5.2 Add `--dry-run` flag (show what would be fixed without writing)
- [x] 5.3 Add color-coded diff output for fixes
- [x] 5.4 Document flags in `ent skill lint --help`
