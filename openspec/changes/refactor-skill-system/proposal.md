---
name: refactor-skill-system
description: Refactor the go-ent skill system from simple markdown format to a structured XML-tagged format following research-backed patterns from `docs/research/SKILL.md`. This change introduces validation, quality scoring, and migration tooling to ensure skills follow best practices for Claude prompt engineering.
status: completed
---

# Refactor Skill System to XML-Tagged Format

## Summary

Refactor the go-ent skill system from simple markdown format to a structured XML-tagged format following research-backed patterns from `docs/research/SKILL.md`. This change introduces validation, quality scoring, and migration tooling to ensure skills follow best practices for Claude prompt engineering.

**Approach**: Template-first - create exemplary go-code skill, validate, then migrate 13 remaining skills.

## Completion Status

**Status**: ✅ Complete (100%)

All phases completed:
- ✅ Phase 1: Infrastructure (domain types, validator, quality scorer)
- ✅ Phase 2: MCP Tooling (skill_validate, skill_quality)
- ✅ Phase 3: Template Skill (go-code v2 format)
- ✅ Phase 4: Migration (all 14 skills migrated to v2)
- ✅ Phase 5: Tooling (skill-sync command, Makefile targets)
- ✅ Phase 6: Documentation (DEVELOPMENT.md, SKILL-AUTHORING.md)

**Tasks**: 31/31 completed - see `tasks.md`

## Problem

### Current State

The skill system uses simple markdown with minimal structure:

```yaml
---
name: go-code
description: "Description. Auto-activates for: triggers."
---
# Skill Title
- Plain markdown content
- Code blocks
```

### Pain Points

1. **No structure validation** - Skills can have any content structure, leading to inconsistent quality
2. **Missing research patterns** - `docs/research/SKILL.md` (254 lines) documents XML tags, roles, examples, edge cases - but current skills use none of these
3. **No quality measurement** - Cannot assess skill effectiveness or track improvements
4. **Inconsistent format** - Each skill author interprets structure differently
5. **Limited metadata** - Only `name` and `description` tracked; no versioning, tags, or authorship
6. **No migration path** - Adding new fields or structure requires manual updates across 14 skills

### Gap Analysis

| Research Recommends | Current Skills Have |
|---------------------|---------------------|
| `<role>` tag with expert persona | None |
| `<instructions>` with clear steps | Section headers |
| `<constraints>` for boundaries | None |
| `<edge_cases>` for handling | None |
| `<examples>` with input/output | Code snippets only |
| `<output_format>` specification | None |
| Version tracking | None |
| Quality scoring | None |

## Solution

### High-Level Approach

**Phase 1: Infrastructure (10h)**
- Extend domain types and parser for v2 frontmatter fields
- Create validator with 9 validation rules
- Implement quality scorer (0-100 scale)
- Extend registry with validation methods

**Phase 2: MCP Tooling (3.5h)**
- `skill_validate` tool - validate single or all skills
- `skill_quality` tool - quality reports with threshold filtering
- Register tools in MCP server

**Phase 3: Template Skill (3.5h)**
- Refactor go-code to v2 format with all XML sections
- Validate template achieves quality score >= 90
- Use as reference for remaining skills

**Phase 4: Migration (13h)**
- Tier 1: go-arch, go-test, go-db, go-api (4h)
- Tier 2: go-sec, go-perf, go-review, go-ops (4h)
- Tier 3: api-design, arch-core, debug-core, review-core, security-core (5h)

**Phase 5: Tooling (1.5h)**
- Create skill-sync command
- Add Makefile targets

**Phase 6: Documentation (3h)**
- Update DEVELOPMENT.md
- Create SKILL-AUTHORING.md guide

### Target Format (v2)

```xml
---
name: skill-name
description: What this skill does and when to use it
version: 2.0.0
author: go-ent
tags: [tag1, tag2]
---

# Skill Name

<role>
Expert persona with domain expertise and behavioral guidelines
</role>

<instructions>
Clear, specific task instructions
Output format requirements
</instructions>

<constraints>
- What to include
- What to exclude
- Boundaries and limitations
</constraints>

<edge_cases>
If input is unclear: [handling]
If information is missing: [handling]
If request is out of scope: [handling]
</edge_cases>

<examples>
<example>
<input>Representative input</input>
<output>Expected output format</output>
</example>
<example>
<input>Edge case input</input>
<output>Edge case handling</output>
</example>
</examples>

<output_format>
Exact specification of expected output structure
</output_format>
```

### Backward Compatibility

- Parser detects format version by checking for `<role>` or `<instructions>` XML tags
- Registry handles both v1 (current) and v2 (new) skills during migration
- Old-style "Auto-activates for:" trigger extraction continues working
- No breaking changes - gradual migration with validation gates

## Breaking Changes

- [ ] API changes - None (additive only)
- [ ] Database migrations - None
- [ ] Configuration changes - None (backward compatible)

## Affected Systems

### Skill Parser (`internal/skill/parser.go`)
- **Impact**: Extended to parse new frontmatter fields (version, author, tags, allowed-tools)
- **Change**: Additive - supports both v1 and v2 formats

### Skill Registry (`internal/skill/registry.go`)
- **Impact**: New validation and quality reporting methods
- **Change**: Additive - existing methods unchanged

### MCP Server (`internal/mcp/server/server.go`)
- **Impact**: Two new tools registered
- **Change**: Additive - no changes to existing tools

### All 14 Skills (`plugins/go-ent/skills/`)
- **Impact**: Content refactored to v2 format
- **Change**: Format migration with content preservation

### Build Process
- **Impact**: New `make skill-validate`, `make skill-sync`, `make skill-quality` targets
- **Change**: Additive - existing targets unchanged

## Alternatives Considered

### 1. Keep Simple Markdown Format
**Why not chosen**: Fails to leverage research-backed patterns that improve Claude performance by 15-20% (per Anthropic research). No path to quality improvement or consistency.

### 2. Enforce v2 Format Immediately (Breaking Change)
**Why not chosen**: Breaks existing deployments. Template-first approach with gradual migration reduces risk and allows iterative validation.

### 3. Use JSON/YAML Instead of XML Tags
**Why not chosen**: Anthropic explicitly recommends XML tags for Claude prompts. Research shows 15-20% performance improvement with XML structure vs unstructured formats.

### 4. External Validation Tool (Separate Binary)
**Why not chosen**: Increases deployment complexity. Integrating validation into the existing MCP server follows the project's plugin architecture and enables runtime validation.

### 5. Automated Migration Tool
**Why not chosen**: Skill content requires human judgment for role definition, constraint specification, and example selection. Template-first approach provides clear reference while maintaining quality control.

## Success Criteria

### Phase 1: Infrastructure ✅
- `go test ./internal/skill/...` passes
- Parser handles both v1 and v2 format detection
- Validator produces meaningful issues with line numbers
- Quality scorer produces scores 0-100

### Phase 2: MCP Tooling ✅
- `skill_validate` tool works via Claude Code MCP
- `skill_quality` tool lists all skills with scores

### Phase 3: Template ✅
- go-code skill validates strict mode with 0 errors
- Quality score >= 90
- All XML sections present and well-formed

### Phase 4: Migration ✅
- All 14 skills validate with 0 errors in strict mode
- Average quality score >= 80
- All skills synced to `.claude/skills/ent/`

### Phase 5/6: Complete ✅
- `make skill-validate` passes
- Documentation complete
- Migration guide published

## Implementation Notes

### Quality Scoring Rubric

| Component | Weight | Criteria |
|-----------|--------|----------|
| Frontmatter | 20 pts | name, description, version, tags present |
| Structure | 30 pts | All required XML sections well-formed |
| Content | 30 pts | Examples present, edge cases covered |
| Triggers | 20 pts | Auto-activation keywords extracted |

### Validation Rules

1. `validateFrontmatter` - Required fields present (name, description)
2. `validateVersion` - Semantic version format (if present)
3. `validateXMLTags` - Well-formed XML sections (balanced tags)
4. `validateRoleSection` - `<role>` tag present and non-empty
5. `validateInstructionsSection` - `<instructions>` tag present
6. `validateExamples` - If `<examples>` present, contains `<example>` children with `<input>` and `<output>`
7. `validateConstraints` - `<constraints>` has list items
8. `validateEdgeCases` - `<edge_cases>` handles at least 2 scenarios
9. `validateOutputFormat` - `<output_format>` present for structured outputs

### Migration Strategy

**Template Creation**:
1. Extract key patterns from current go-code skill
2. Define role based on Go expertise
3. Convert sections to XML-tagged format
4. Add 2-3 input/output examples
5. Define constraints from CLAUDE.md guidelines
6. Specify edge case handlers (delegation to other skills)
7. Validate achieves quality score >= 90

**Skill Migration**:
- Use go-code as reference template
- Preserve all existing content value
- Enhance with XML structure
- Add missing sections (role, examples, edge cases)
- Validate before considering complete

## Timeline Estimate

**Total**: ~34 hours

- Phase 1: 10h (infrastructure)
- Phase 2: 3.5h (MCP tools)
- Phase 3: 3.5h (template)
- Phase 4: 13h (migration)
- Phase 5: 1.5h (tooling)
- Phase 6: 3h (docs)

## Dependencies

- No external dependencies
- Uses existing validation pattern from `internal/spec/validator.go`
- Follows existing parser pattern from `internal/skill/parser.go`

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Skills don't improve Claude performance | High | Validate template first, A/B test if needed |
| Migration introduces errors | Medium | Strict validation gates, manual review |
| Backward incompatibility | Medium | Parser detects version, supports both formats |
| Quality scoring too subjective | Low | Use research-backed criteria, iterate based on feedback |
| Migration effort underestimated | Low | Template-first validates approach, can pause migration |
