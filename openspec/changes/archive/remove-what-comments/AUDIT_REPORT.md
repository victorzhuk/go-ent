# Audit Report: remove-what-comments Proposal

## Date
2025-01-30

## Summary

The original proposal claimed 180+ WHAT comment violations across `internal/execution/`, `internal/config/`, and `internal/mcp/tools/`.

**Key Finding:**
- `internal/execution/` directory **does not exist** (package was deleted)
- Remaining codebase has **very few actual WHAT comment violations**
- Most comments are legitimate: function docs, test steps, or WHY comments

## Files Audited

### Non-Existent (from original proposal)
```
❌ internal/execution/engine.go (deleted)
❌ internal/execution/runner.go (deleted)
❌ internal/execution/opencode.go (deleted)
❌ internal/execution/codemode.go (deleted)
❌ internal/execution/cli.go (deleted)
❌ internal/execution/budget.go (deleted)
❌ internal/execution/sandbox.go (deleted)
❌ internal/execution/parallel.go (deleted)
❌ internal/execution/multi.go (deleted)
❌ internal/execution/single.go (deleted)
❌ internal/execution/strategy.go (deleted)
❌ internal/execution/context.go (deleted)
❌ internal/execution/fallback.go (deleted)
❌ internal/execution/summarization.go (deleted)
```

### Existing Packages Audited

**internal/config/** - 14 .go files
- ✅ Clean - only proper Go documentation comments for exported types/functions
- ✅ No WHAT comment violations found
- Example: `// Validate validates the metrics configuration.` (proper doc comment)

**internal/mcp/tools/** - 25 .go files (excluding test files)
- ✅ Mostly clean
- ⚠️ Some borderline WHAT comments in registry.go (section headers)
- ✅ No clear violations that need removal

**internal/skill/** - 19 .go files
- ✅ Clean - only proper documentation comments
- ✅ No WHAT comment violations found

**internal/spec/** - 23 .go files
- ✅ Clean - mostly proper documentation
- ⚠️ Some borderline WHAT comments in registry_store.go (section headers)

**internal/generator/** - 10 .go files
- ⚠️ Some borderline WHAT comments (section headers in claude.go, opencode.go)
- Example: `// Build frontmatter`, `// Marshal frontmatter`, `// Inline prompts`

**internal/genspec/** - 6 .go files
- ⚠️ Some borderline WHAT comments in validator.go
- Example: `// Read file`, `// Extract frontmatter`, `// Parse frontmatter as map`

## Comment Types Found

### ✅ Acceptable Comments (Keep)

**1. Function Documentation**
```go
// Validate validates the metrics configuration.
func (m *MetricsConfig) Validate() error { ... }
```

**2. WHY Comments (Explain rationale)**
```go
// Check if registry.db exists before creating BoltDB
registryPath := filepath.Join(input.Path, "openspec", "registry.db")
```

**3. Test Step Comments (Clarity)**
```go
// Create validator
v := NewValidator()
// Create temp dir
dir := t.TempDir()
```

**4. Section Headers (Readability)**
```go
// Check aliases first
if alias, ok := r.cfg.Aliases[agentModel]; ok {
    agentModel = alias
}
```

### ⚠️ Borderline WHAT Comments (Optional to Remove)

These are short comments that could be removed but don't significantly violate the rule:

**internal/genspec/validator.go**
- Line 31: `// Read file` → Could be removed (code is clear)
- Line 41: `// Extract frontmatter` → Could be removed
- Line 51: `// Parse frontmatter as map` → Could be removed

**internal/generator/claude.go**
- Line 39: `// Build frontmatter` → Could be removed
- Line 53: `// Marshal frontmatter` → Could be removed
- Line 59: `// Inline prompts` → Could be removed
- Line 62: `// Build final markdown with frontmatter` → Could be removed

**internal/mcp/tools/generate.go**
- Line 90: `// Validate project type` → Could be removed
- Line 106: `// Create template engine` → Could be removed
- Line 109: `// Prepare template variables` → Could be removed

**Total Borderline Cases:** ~20-30 comments across all packages

## Conclusion

**The original proposal scope was wildly overstated.**

**Reality:**
- 14 execution package files don't exist
- Remaining codebase has ~20-30 borderline WHAT comments (optional to remove)
- Most comments are legitimate documentation or WHY comments

**Recommendation:**
Given the small scope (~20-30 comments), it's better to:
1. **Archive this proposal** as outdated/incorrect
2. **Handle any cleanup directly** if/when comments become problematic
3. **Focus on other proposals** with actual scope

**Or:** Update proposal to reflect actual minimal scope if you still want to track it.

## Next Steps (Choose One)

**Option A:** Archive proposal (recommended)
```
ent:archive remove-what-comments
```

**Option B:** Update proposal to minimal scope
- Update `proposal.md` with accurate file list
- Update `tasks.md` with realistic task breakdown
- Focus only on the ~20-30 borderline cases found

**Option C:** Do nothing
- The comments don't significantly violate the rule
- Most are legitimate documentation or section headers
- Time is better spent on actual features
