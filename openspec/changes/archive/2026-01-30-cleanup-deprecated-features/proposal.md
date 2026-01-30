# Cleanup Deprecated Features

## Summary

Remove unused code, dead packages, and structural issues that accumulated during development. Clean up the codebase to align with the simplified architecture vision based on comprehensive investigation findings.

## Investigation Findings

### Three-Directory Pattern Analysis

The codebase uses a clear three-directory pattern that should be preserved:

- **`pkg/`** - Source code and canonical definitions
- **`.claude/`** - Generated files by Claude Code plugin
- **`.opencode/`** - Generated files by OpenCode

**Decision**: DO NOT touch `.claude/` or `.opencode/` directories. These are generated artifacts that should remain in git.

### Structural Issues Discovered

1. **Template terminology ambiguity**: "template" refers to three different concepts:
   - Skill templates for creating new skills
   - Go templates (text/template)
   - Template engine code

2. **Skill location duplication**: Skills exist in four locations:
   - `pkg/skills/` - Source code for skill system
   - `pkg/skills/templates/` - Skill templates
   - `pkg/skills/plugins/` - Unused subtree (dead code)
   - `internal/skill/` - Internal skill logic

3. **Schema location confusion**: Schemas exist in two locations:
   - `schemas/` - YAML schemas (canonical)
   - `pkg/schemas/` - Go schema definitions

4. **Marketplace package**: `internal/marketplace/` is confirmed dead code:
   - Zero production usage (no imports, no tests reference it)
   - No main.go entry point
   - Not part of any active workflow

## Problem

The codebase has accumulated technical debt in four main categories:

### 1. Test Artifacts Committed to Git

Five coverage files exist in the root directory:
- `coverage.out`
- `reg_coverage.out`
- `coverage_skill.out`
- `coverage_template.out`
- `combined_coverage.out`

These were committed before `.gitignore` was properly configured and should be removed.

### 2. Dead Code: Marketplace Package

`internal/marketplace/` is entirely dead code:
- No production usage (confirmed via code search)
- No active tests reference it
- No main.go entry point
- Not integrated into CLI or MCP server

### 3. MCP Tool Test Artifacts

`internal/mcp/tools/` contains test artifacts:
- `exports/` directory (temporary export files)
- Multiple `.json`, `.csv`, `.prom` files
- `testdata/fallback/` (unused test data)

### 4. Duplicate Structures

- `pkg/skills/plugins/` subtree (unused)
- Schema duplication between `schemas/` and `pkg/schemas/`

**Impact**:
- Slower builds (compiling unused code)
- Confusion for new contributors (what is marketplace?)
- Maintenance overhead (dead code still appears in search)
- Larger repository size

## Solution

### Phase 1: Remove Coverage Files from Git

```bash
# Remove committed coverage files
git rm coverage.out
git rm reg_coverage.out
git rm coverage_skill.out
git rm coverage_template.out
git rm combined_coverage.out

# Already in .gitignore for future:
# *.out
# coverage*.out
```

### Phase 2: Remove Marketplace Package (Dead Code)

```bash
# Remove entire marketplace package
git rm -r internal/marketplace/
```

**Rationale**: Confirmed zero production usage via comprehensive code search.

### Phase 3: Clean MCP Tool Test Artifacts

```bash
# Remove exports directory
git rm -r internal/mcp/tools/exports/

# Remove JSON/CSV/PROM test artifacts
git rm internal/mcp/tools/*.json
git rm internal/mcp/tools/*.csv
git rm internal/mcp/tools/*.prom

# Remove unused testdata/fallback
git rm -r internal/mcp/tools/testdata/fallback/
```

**Note**: Keep `internal/mcp/tools/testdata/` directory structure - some files are actively used by tests.

### Phase 4: Remove Unused Subtrees

```bash
# Remove unused plugins subtree
git rm -r pkg/skills/plugins/
```

### Phase 5: Schema Consolidation (Documentation Only)

Current state:
- `schemas/` - YAML schemas (canonical location)
- `pkg/schemas/` - Go schema definitions (used by code)

**Action**: Add documentation clarifying that `schemas/` contains canonical YAML definitions. No code changes needed - pkg/schemas/ is actively used.

### Phase 6: Document Structure

Create `docs/ARCHITECTURE.md` or update existing documentation to explain:
- Three-directory pattern (pkg/ vs .claude/ vs .opencode/)
- Schema location rationale
- Template terminology disambiguation
- Skill system organization

## Affected Systems

- Root directory - Remove coverage files
- `internal/marketplace/` - Remove entire package (dead code)
- `internal/mcp/tools/exports/` - Remove
- `internal/mcp/tools/testdata/fallback/` - Remove
- `pkg/skills/plugins/` - Remove entire subtree
- Documentation - Update with structure explanation

**NOT affected**:
- `openspec/changes/archive/` - Stays in git (historical record)
- `.claude/` - Generated files stay (DO NOT TOUCH)
- `.opencode/` - Generated files stay (DO NOT TOUCH)
- `internal/template/testdata/` - Actively used by tests
- `pkg/schemas/` - Used by code
- `schemas/` - Canonical YAML schemas

## Structural Issues (To Document)

### Template Terminology

"Template" means three different things in this codebase:

1. **Skill Templates**: Pre-defined skill templates for creating new skills
   - Location: `pkg/skills/templates/`
   - Purpose: Provide starting points for skill creation
   - Usage: CLI command `ent skill new --template <name>`

2. **Go Templates**: Standard library text/template
   - Usage: Template engine in skill generation
   - Files: `.tmpl` files throughout codebase

3. **Template Engine**: Code that processes Go templates
   - Location: `internal/template/`
   - Purpose: Parse and execute templates

**Recommendation**: Document this distinction clearly.

### Skill Organization

Skills exist in four locations with distinct purposes:

- `pkg/skills/` - Source code for skill system (core logic)
- `pkg/skills/templates/` - Skill templates (user-facing)
- `internal/skill/` - Internal skill processing logic
- `pkg/skills/plugins/` - **UNUSED** (removed in this change)

### Schema Locations

Two schema locations with different purposes:

- `schemas/` - YAML schema definitions (canonical, human-readable)
- `pkg/schemas/` - Go schema parsing code (used by application)

**Recommendation**: Document that `schemas/` is the canonical source for schema definitions.

## Breaking Changes

- [x] Remove marketplace package (no production usage)
- [x] Remove unused plugins subtree
- [x] Remove test artifacts from git
- [x] Remove coverage files from git

**No user-facing breaking changes**: All removed code was unused in production.

## Migration Path

1. Create backup branch: `git checkout -b backup-cleanup`
2. Phase 1: Remove coverage files from git
3. Phase 2: Remove marketplace package
4. Phase 3: Clean MCP tool test artifacts
5. Phase 4: Remove unused plugins subtree
6. Phase 5: Update documentation (schema explanation)
6. Phase 6: Create/update architecture documentation
7. Verify: `make build && make test`
8. Commit with detailed message

## Alternatives Considered

1. **Move archive to .gitignore**: Rejected - User wants archive in git as historical record
2. **Consolidate schemas immediately**: Rejected - No code benefit, pkg/schemas/ is actively used
3. **Keep everything**: Rejected - Dead code adds confusion and build overhead
4. **Phase-by-phase cleanup (chosen)**: Safer, easier to revert if needed

## Success Criteria

- [ ] Coverage files removed from git
- [ ] Marketplace package removed
- [ ] MCP tool test artifacts removed
- [ ] Unused plugins subtree removed
- [ ] Build passes: `make build`
- [ ] Tests pass: `make test`
- [ ] Documentation updated with structure explanation
- [ ] Clean `git status` after changes
- [ ] No regressions in functionality

## Effort Estimate

**~4 hours** across 6 phases

**Breakdown**:
- Phase 1 (Coverage files): 15 minutes
- Phase 2 (Marketplace): 30 minutes (verify no dependencies)
- Phase 3 (MCP artifacts): 30 minutes
- Phase 4 (Plugins subtree): 15 minutes
- Phase 5 (Schema docs): 30 minutes
- Phase 6 (Architecture docs): 1 hour
- Testing & verification: 1 hour

## Related Changes

None - this is a standalone cleanup task.

## Notes

- Archive directory stays in git per user decision (historical record)
- `.claude/` and `.opencode/` directories are NOT touched (generated files)
- Focus is on removing clearly dead code and test artifacts only
- Schema consolidation deferred to future if needed (no urgent benefit)
