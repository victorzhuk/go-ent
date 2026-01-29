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

| Aspect | Current (v3) | New (v4) |
|--------|--------------|----------|
| Frontmatter | 15+ fields | 3 fields (name, description, triggers) |
| Body format | XML-like tags | Markdown sections (## Role, ## Instructions) |
| Loading | 3-tier progressive | Single load |
| Validation | 20+ rules with scoring | 5 essential rules |
| Dependencies | Complex dependency graph | Simple list in frontmatter |
| File patterns | Regex patterns | Simple glob patterns |

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

1. Convert existing skills to v4 format (automated script)
2. Update parser to handle both v3 and v4 during transition
3. Remove v3 support after all skills migrated
4. Update documentation

## Alternatives Considered

1. **Keep current system**: Rejected - too complex for MVP
2. **Use Claude Code native skills**: Rejected - not portable to OpenCode
3. **Use OpenCode native skills**: Rejected - not portable to Claude Code
4. **Minimal Markdown format (chosen)**: Portable, simple, maintainable

## Success Criteria

- [ ] All skills converted to v4 format
- [ ] Parser handles v4 format
- [ ] Validation reduced to 5 essential rules
- [ ] Progressive loading removed
- [ ] Quality scoring removed
- [ ] Skill tools simplified
- [ ] Tests updated and passing
- [ ] Documentation updated

## Effort Estimate

**~16 hours** across 12 tasks
