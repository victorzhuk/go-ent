# Archive: Remove WHAT Comments

## Archived Date
2025-01-30

## Archive Reason
**Scope was based on outdated assumptions and non-existent packages.**

### Original Claim
The proposal claimed 180+ WHAT comment violations across:
- `internal/execution/` - Multiple files (engine.go, runner.go, etc.)
- `internal/config/` - config.go, loader.go, etc.
- `internal/mcp/tools/` - 25+ tool implementation files

### Reality After Audit

**Non-Existent Files (from original proposal):**
- ❌ `internal/execution/` - Entire package deleted (14 files listed in original scope)
- ❌ None of the execution package files exist in the codebase

**Clean Packages:**
- ✅ `internal/config/` - Only proper Go documentation comments, zero violations
- ✅ `internal/skill/` - Only proper documentation, zero violations
- ✅ `internal/spec/` - Only proper documentation, zero violations

**Actual Findings:**
- ⚠️ `internal/genspec/validator.go` - 3 borderline WHAT comments
- ⚠️ `internal/generator/claude.go` - 4 borderline WHAT comments
- ⚠️ `internal/mcp/tools/generate.go` - 3 borderline WHAT comments

**Total:** ~10-15 borderline comments (not violations, just optional cleanup)

### Why These Are Borderline

Most "borderline" comments are actually:
1. **Section headers** improving readability of long functions
2. **Test step comments** that clarify test flow
3. **Minimal descriptive comments** that don't clearly violate the rule

Example of borderline case:
```go
// Build frontmatter
fm := ClaudeFrontmatter{...}
```

This could be removed, but the code is clear and the comment doesn't significantly violate the "WHAT comment" rule.

### Audit Details

**Audit Date:** 2025-01-30

**Files Audited:**
- `internal/config/` - 14 .go files (clean)
- `internal/mcp/tools/` - 25 .go files (mostly clean)
- `internal/skill/` - 19 .go files (clean)
- `internal/spec/` - 23 .go files (clean)
- `internal/generator/` - 10 .go files (4 borderline)
- `internal/genspec/` - 6 .go files (3 borderline)

**Total Files Audited:** ~97 files
**Actual Violations:** 0 clear violations
**Borderline Cases:** ~10-15 comments

## Recommendation

**Handle future violations directly as they arise.**

Given the minimal scope (~10-15 borderline comments) and the fact that most are acceptable documentation:

1. ✅ **Archive this proposal** - Original scope was incorrect
2. ✅ **No dedicated work needed** - Comments don't significantly violate project standards
3. ✅ **Handle as normal code review** - If a developer spots a clear WHAT comment during review, remove it directly
4. ✅ **Focus on actual features** - Time is better spent on proposals with real impact

## Alternative Path

If you still want to track this work:

1. Create a new proposal with accurate scope (~10-15 comments)
2. Focus only on clear WHAT comment violations (not borderline cases)
3. Consider if the effort is justified given the minimal impact

## Archive Command

```bash
ent:archive remove-what-comments
```

## Related Files

- Original proposal: `proposal.md`
- Tasks: `tasks.md`
- Audit report: `AUDIT_REPORT.md`
- Summary: `SUMMARY.md`
