# Proposal: Remove WHAT Comments

## Metadata
- **Change ID:** `remove-what-comments`
- **Status:** Proposed (Audited - Scope Revised)
- **Type:** Code Style
- **Priority:** Low
- **Affects Specs:** None (style cleanup)

## Problem

**Updated after audit (2025-01-30):**

The codebase has minimal WHAT comment violations. The original proposal was based on non-existent `internal/execution/` package.

**Rule:**
> Comments explaining WHAT code does = BAD NAMING. Fix the name instead.
> ONLY acceptable (rare): Comments explaining WHY - counterintuitive behavior, legacy requirements, etc.

**Files with violations (post-audit):**
- `internal/genspec/validator.go` - 3 borderline WHAT comments
- `internal/generator/claude.go` - 4 borderline WHAT comments
- `internal/mcp/tools/generate.go` - 3 borderline WHAT comments

**Total:** ~10 borderline WHAT comments to remove

## Examples of Violations

```go
// BAD: WHAT comment (borderline)
// Read file
content, err := os.ReadFile(path)

// BAD: WHAT comment (borderline)
// Build frontmatter
fm := ClaudeFrontmatter{...}

// BAD: WHAT comment (borderline)
// Create template engine
engine := template.NewEngine(pkg.FS)
```

**ACCEPTABLE Comments (WHY or Documentation):**
```go
// Required by legacy API - remove after v2 migration
resp.Header.Set("X-Legacy-Token", token)

// Validate validates the metrics configuration (proper Go doc comment)
func (m *MetricsConfig) Validate() error {
    return nil
}

// Check if registry.db exists before creating BoltDB (WHY comment explains fallback)
registryPath := filepath.Join(input.Path, "openspec", "registry.db")
```

## Proposed Solution

Remove all WHAT comments from the codebase that simply describe what the code does. Keep only:

1. **WHY comments** - Explain counterintuitive behavior, legacy requirements
2. **Package comments** - `// Package execution provides...`
3. **Exported function docs** - Required by golint for public APIs
4. **TODO/FIXME comments** - Action items (if they don't reference internal tickets)

## Scope

**In Scope:**
- `internal/genspec/validator.go` - ~3 borderline WHAT comments
- `internal/generator/claude.go` - ~4 borderline WHAT comments
- `internal/generator/opencode.go` - similar pattern to claude.go
- `internal/mcp/tools/generate.go` - ~3 borderline WHAT comments

**Out of Scope:**
- `internal/execution/` - Package doesn't exist (deleted)
- `internal/config/` - Clean, only proper Go documentation
- `internal/skill/` - Clean, only proper Go documentation
- `internal/spec/` - Clean, only proper Go documentation
- Test files (test step comments are acceptable)
- Generated code
- Third-party code
- Comments that explain WHY (keep these)
- Function documentation comments (legitimate Go docs)

## Impact

- **Breaking Changes:** None (comments don't affect runtime behavior)
- **API Changes:** None
- **Migration Required:** No
- **Testing Required:** No (pure comment removal)

## Risks

- **Very Low Risk:** Only removing comments, no code changes
- **Review Needed:** Human review to ensure we don't accidentally remove WHY comments
- **CI/CD:** No impact, builds continue to work

## Implementation Strategy

1. **Automated removal** where safe (obvious WHAT comments)
2. **Manual review** for borderline cases
3. **Preserve** any WHY comments or genuinely useful explanations

## Success Criteria

- [ ] Borderline WHAT comments removed from genspec/validator.go
- [ ] Borderline WHAT comments removed from generator/claude.go
- [ ] Borderline WHAT comments removed from generator/opencode.go
- [ ] Borderline WHAT comments removed from mcp/tools/generate.go
- [ ] All WHY comments preserved
- [ ] All legitimate function documentation comments preserved
- [ ] Code compiles and tests pass (no accidental code removal)

## Notes

**After Audit (2025-01-30):**
- Original scope was based on non-existent `internal/execution/` package
- Actual scope is ~10-15 borderline comments across 4 files
- Most comments in codebase are legitimate: function docs, WHY comments, or test steps
- This is a very low-priority cleanup task
- Consider archiving this proposal and handling any future violations directly

## Alternatives Considered

1. **Keep comments** - Violates project standards, creates inconsistency
2. **Improve naming instead** - Some cases, but most code is already well-named
3. **Add linter rule** - Good follow-up, but doesn't fix existing violations
