# Simplify Skill Format

## Summary

Reduce the skill system from a complex XML-like tagged format to a minimal Markdown-based structure that aligns with native Claude Code and OpenCode conventions. Remove progressive loading, complex validation, and over-engineered features that add maintenance burden without MVP value.

## Problem

The current skill system has grown overly complex:

1. **XML-like tags in Markdown**: Skills use `<triggers>`, `<role>`, `<instructions>`, `<constraints>` tags embedded in Markdown
2. **Progressive loading**: Three-tier loading system (metadata → core → extended) adds complexity for marginal token savings
3. **Complex validation**: 20+ validation rules with quality scoring, overlap detection, fixer suggestions
4. **Dual format maintenance**: Separate parsing logic for different skill formats
5. **Heavy metadata**: Skills carry 15+ fields in frontmatter, many unused

This complexity conflicts with the vision of a "slim and clean" orchestration tool that leverages native agent capabilities.

## Solution

### New Skill Format (v4 Minimal)

**File**: `skills/{category}/{name}/SKILL.md`

```markdown
---
name: go-error
description: Error handling patterns for Go
triggers:
  - error handling
  - wrap errors
  - fmt.Errorf
---

## Role

Expert Go error handling engineer.

## Instructions

### Error Wrapping
Always wrap errors with context using %w.

### Custom Errors
Define package-level errors for domain concepts.

## Examples

```go
// Good
return fmt.Errorf("query user %s: %w", id, err)

// Bad
return fmt.Errorf("Failed to query user: %w", err)
```
```

### Key Changes

| Aspect | Current (v3) | New (v4) | Rationale |
|--------|--------------|----------|-----------|
| Frontmatter | 15+ fields | 3 fields (name, description, triggers) | Only essentials |
| Triggers | Complex object with weight/file_pattern | Flat array | Runtime matching, simpler |
| Category | Explicit field | Inferred from path | `skills/{category}/` |
| Body sections | 6 required (Role, Instructions, Constraints, Edge Cases, Examples, Output Format) | 3 required (Role, Instructions, Examples) | YAGNI - move heavy sections to references/ |
| Constraints/Edge Cases | Required sections | Optional → references/ or merged | Keep core skill lean |
| Output Format | Required section | Merged into Instructions | Often redundant |
| Loading | 3-tier progressive | Single load | Simpler code |
| Validation | 20+ rules with scoring | 5 essential rules | Less maintenance |
| Quality Score | Computed and stored | Removed | Not needed at runtime |
| References/ | Convention only | Validated structure | Consistency |
| XML Support | Yes (v2 compatibility) | No | Markdown only |

### Design Decisions (from Exploration)

1. **Constraints/Edge Cases/Output Format → References/**
   - Research shows optimal prompts have 3-4 sections (Identity, Instructions, Examples, Context)
   - Move sections with >5 items to `references/` directory
   - Keeps core SKILL.md focused and readable

2. **Triggers: Flat Array (No Weight)**
   - Weight computed at runtime from match ratio
   - File pattern inferred from category
   - Simpler mental model for skill authors

3. **Category: Path-Based Inference**
   - `skills/go/*` → category: go
   - `skills/core/*` → category: core
   - `skills/ent/*` → category: ent
   - Explicit contract via path structure

4. **References/ Validation**
   - Max depth: 1 subdirectory
   - No frontmatter in reference files
   - Max file size: 50KB
   - Ensures consistency without being overly prescriptive

5. **No Backward Compatibility**
   - Clean break from v3
   - One-time migration via conversion script
   - Single code path (no version detection)

## Affected Systems

- `internal/skill/parser.go` - Simplify parsing logic
- `internal/skill/validator.go` - Reduce validation rules
- `internal/skill/registry.go` - Remove progressive loading
- `internal/skill/scorer.go` - Remove quality scoring
- `internal/skill/overlap.go` - Remove overlap detection
- `internal/skill/fixer.go` - Remove auto-fixing
- `internal/mcp/tools/skill_*.go` - Simplify skill tools
- `internal/cli/.claude/skills/` - Update all skill files
- `internal/cli/.opencode/skills/` - Update all skill files

## Breaking Changes

- [x] Skill format changes (v3 → v4)
- [x] Remove progressive loading API
- [x] Remove quality scoring from registry
- [x] Simplify MCP skill tools

## Migration Path

1. **Create migration tool** (Task 9)
   - Build `ent skill migrate v3-to-v4` command
   - Support --dry-run, --backup, --all flags
   - Convert triggers, sections, frontmatter

2. **Convert all skills** (Task 10)
   - Run migration on all ~30 skills in pkg/skills/
   - Create backups before conversion
   - Manual review for skills with complex references/

3. **Update parser** (Tasks 1-3)
   - Add v4 parsing support
   - Remove progressive loading
   - Infer category from path

4. **Simplify validation** (Task 4)
   - Reduce to 5 essential rules
   - Add optional reference validation
   - Remove v3-specific rules

5. **Remove legacy systems** (Tasks 5-7)
   - Delete quality scoring
   - Delete overlap detection
   - Delete auto-fixing

6. **Update documentation**
   - New SKILL-AUTHORING.md for v4
   - Migration guide
   - Update CLI reference

## Reference Documents

- [v4 Specification](/home/zhuk/.local/share/opencode/memories/skill-format-v4-spec.md) - Complete format specification
- [Migration Script Design](/home/zhuk/.local/share/opencode/memories/skill-migration-script-design.md) - Detailed migration tool design

## Alternatives Considered

1. **Keep current system**: Rejected - too complex for MVP
2. **Use Claude Code native skills**: Rejected - not portable to OpenCode
3. **Use OpenCode native skills**: Rejected - not portable to Claude Code
4. **Minimal Markdown format (chosen)**: Portable, simple, maintainable

## Success Criteria

### Implementation
- [x] Parser handles v4 format (flat triggers, 3 sections)
- [x] Category inference from path works correctly
- [x] Validation reduced to 5 essential rules + 2 optional
- [x] References/ validation implemented
- [x] Progressive loading removed
- [x] Quality scoring removed
- [x] Overlap detection removed
- [x] Auto-fixing removed
- [x] Registry simplified (no load levels)
- [x] MCP skill tools updated

### Migration
- [x] Migration tool created (`cmd/skill-convert/main.go`)
- [x] All 27 skills converted to v4
- [x] Backups created for all converted skills
- [x] References/ directories created where needed (18 skills)
- [x] No v3 frontmatter fields remain

### Quality
- [x] All tests updated and passing
- [x] `go build ./...` succeeds
- [x] Skills validate without errors

### Documentation
- [x] v4 specification created (memory: skill-format-v4-spec)
- [x] Migration script design documented (memory: skill-migration-script-design)

## Effort Estimate

**~16 hours** across 12 tasks
