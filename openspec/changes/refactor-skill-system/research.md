# Research: Skill System Refactoring

## Research Question

How should we structure Claude skills to maximize effectiveness, and what gap exists between current implementation and research-backed best practices?

## Current System Analysis

### Architecture

**Skill Loading Flow**:
```
MCP Server → Registry.Load(skillsPath)
           → Parser.ParseSkillFile(SKILL.md)
           → Extract frontmatter (name, description)
           → Extract triggers from "Auto-activates for:"
           → Store SkillMeta in registry
```

**File Structure**:
- **Source**: `plugins/go-ent/skills/` (14 skills: 5 core, 9 Go-specific)
- **Runtime**: `.claude/skills/ent/` (synced via plugin system)
- **Format**: SKILL.md with YAML frontmatter + markdown body

**Current Metadata** (`internal/skill/parser.go:13-18`):
```go
type SkillMeta struct {
    Name        string
    Description string
    Triggers    []string   // Extracted from description
    FilePath    string
}
```

### Example Current Format

**File**: `plugins/go-ent/skills/go/go-code/SKILL.md`

```yaml
---
name: go-code
description: "Modern Go implementation patterns, error handling, concurrency. Auto-activates for: writing Go code, implementing features, refactoring, error handling, configuration."
---

# Go Code Patterns

## Bootstrap Pattern
[Code block]

## Error Handling
[Code block]

## Concurrency
[Code block]
```

**Characteristics**:
- 235 lines, plain markdown
- No XML structure
- No role definition
- Code examples without input/output pairing
- No explicit edge case handling
- No constraints section

### Skill Inventory

| Category | Count | Examples |
|----------|-------|----------|
| Core (language-agnostic) | 5 | arch-core, debug-core, review-core, security-core, api-design |
| Go-specific | 9 | go-code, go-test, go-arch, go-db, go-api, go-sec, go-perf, go-review, go-ops |
| **Total** | **14** | |

## Research Guide Analysis

### Source

`docs/research/SKILL.md` - 254-line comprehensive guide synthesizing:
- Anthropic official documentation
- Peer-reviewed research (1,500+ papers)
- Validated community patterns

### Key Findings

#### 1. XML Tags Show 15-20% Performance Improvement

**From research**:
> "XML tags are Claude's recommended primary method for structuring prompts, with research showing **15-20% performance improvement** when properly implemented."

**Recommended structure**:
```xml
<role>Expert persona</role>
<instructions>Task guidelines</instructions>
<constraints>Boundaries</constraints>
<edge_cases>Special handling</edge_cases>
<examples>Input/output pairs</examples>
<output_format>Expected structure</output_format>
```

#### 2. Role Assignment is Most Powerful

**From research**:
> "Role prompting is what Anthropic calls 'the most powerful way to use system prompts with Claude.'"

**Pattern**:
```xml
<role>
You are a senior data scientist with 10+ years experience.
Provide evidence-based answers and cite sources.
When uncertain, say so explicitly rather than speculating.
</role>
```

**Current skills**: ❌ No role definitions

#### 3. Examples Should Use Input/Output Pairs

**From research**:
> "Anthropic's official guidance recommends **3-5 diverse, relevant examples** to show Claude exactly what you want."

**Pattern**:
```xml
<examples>
<example>
<input>User request or scenario</input>
<output>Expected response format</output>
</example>
</examples>
```

**Current skills**: ❌ Code snippets only, no input/output pairing

#### 4. Edge Cases Must Be Explicit

**From research**:
> "Include edge cases in your examples—Anthropic specifically recommends 'challenging examples and edge cases to help Claude understand exactly what you're looking for.'"

**Pattern**:
```xml
<edge_cases>
If input is unclear: Ask specific clarifying questions
If information is missing: State assumptions explicitly
If request is out of scope: Delegate to appropriate skill
</edge_cases>
```

**Current skills**: ❌ No edge case sections

#### 5. Constraints Prevent Ambiguity

**From research**:
> "Ambiguity is the root cause of most prompt failures. Replace subjective terms with concrete specifications."

**Pattern**:
```xml
<constraints>
- ZERO comments explaining WHAT code does
- No verbose AI-style naming
- No magic numbers
- Private by default
</constraints>
```

**Current skills**: ❌ No constraints sections

## Gap Analysis

### Structural Gaps

| Research Pattern | Current Skills | Gap | Impact |
|------------------|----------------|-----|--------|
| XML tags for structure | Plain markdown | Missing | 15-20% performance loss |
| `<role>` with expert persona | None | Missing | Less focused responses |
| `<instructions>` section | Headers only | Weak | Ambiguous guidance |
| `<constraints>` boundaries | None | Missing | Inconsistent output |
| `<edge_cases>` handling | None | Missing | Poor failure modes |
| `<examples>` with I/O pairs | Code snippets | Weak | Less clear expectations |
| `<output_format>` spec | None | Missing | Format inconsistency |

### Metadata Gaps

| Field | Current | Needed | Purpose |
|-------|---------|--------|---------|
| Version | ❌ | ✅ | Track evolution, compatibility |
| Author | ❌ | ✅ | Attribution, maintenance |
| Tags | ❌ | ✅ | Categorization, discovery |
| AllowedTools | ❌ | ✅ | Security boundary |
| QualityScore | ❌ | ✅ | Effectiveness measurement |

### Infrastructure Gaps

| Capability | Current | Needed | Purpose |
|------------|---------|--------|---------|
| Structure validation | ❌ | ✅ | Enforce consistency |
| Content quality scoring | ❌ | ✅ | Measure effectiveness |
| Version detection | ❌ | ✅ | Support migration |
| MCP validation tools | ❌ | ✅ | Runtime verification |

## Technology Evaluation

### XML vs JSON vs YAML for Structure

| Aspect | XML Tags | JSON | YAML |
|--------|----------|------|------|
| Claude performance | ✅ 15-20% improvement (Anthropic research) | ❌ No evidence | ❌ No evidence |
| Readability | ✅ Clear boundaries | ⚠️ Verbose | ✅ Clean |
| Nested structure | ✅ Native | ✅ Native | ✅ Native |
| Anthropic recommendation | ✅ Explicit | ❌ Not mentioned | ❌ Not mentioned |

**Recommendation**: XML tags per Anthropic research

### Validation Approach Options

#### Option A: External Validation Tool
**Pros**:
- Independent binary
- Could validate across projects

**Cons**:
- Deployment complexity
- Not integrated with MCP server
- Requires separate installation

#### Option B: Integrated MCP Tools (Chosen)
**Pros**:
- Follows plugin architecture
- Runtime validation available
- Exposed via Claude Code naturally
- Uses existing patterns (`internal/spec/validator.go`)

**Cons**:
- Coupled to MCP server

**Decision**: Option B - integrate into MCP server following existing validation patterns

### Migration Strategy Options

#### Option A: Automated Migration Tool
**Pros**:
- Fast migration
- Consistent transformation

**Cons**:
- Cannot infer role definitions
- Cannot generate meaningful examples
- Cannot determine proper constraints
- Risk of losing content nuance

#### Option B: Template-First Manual Migration (Chosen)
**Pros**:
- Quality control per skill
- Human judgment for role/examples/constraints
- Template provides clear reference
- Iterative validation

**Cons**:
- More time investment
- Requires understanding of each skill domain

**Decision**: Option B - create exemplary go-code template, then manual migration with validation gates

### Quality Scoring Approach

**Rubric Design**:

| Component | Weight | Rationale |
|-----------|--------|-----------|
| Frontmatter completeness | 20% | Foundation for discovery and versioning |
| Structure compliance | 30% | Correlates with 15-20% performance gain |
| Content quality | 30% | Examples and edge cases drive clarity |
| Trigger clarity | 20% | Enables auto-activation |

**Scoring Formula**:
```
Score = (FrontmatterScore × 0.2) +
        (StructureScore × 0.3) +
        (ContentScore × 0.3) +
        (TriggerScore × 0.2)
```

**Thresholds**:
- **90-100**: Excellent - all patterns implemented
- **80-89**: Good - minor gaps
- **70-79**: Acceptable - needs improvement
- **<70**: Poor - requires refactoring

## Validation Rules Research

### Existing Pattern: Spec Validator

**File**: `internal/spec/validator.go`

The project has an existing validation pattern for OpenSpec files:

```go
type Validator struct {
    rules []ValidationRule
}

type ValidationRule func(ctx *ValidationContext) []ValidationIssue
```

**Rules Pattern**:
- Each rule is a function that checks one aspect
- Returns list of issues with severity (error, warning, info)
- Context provides file path, content, parsed structure

**Recommendation**: Follow this pattern for skill validation

### Proposed Validation Rules

1. **validateFrontmatter**: name and description required
2. **validateVersion**: semantic version format if present
3. **validateXMLTags**: balanced tags, proper nesting
4. **validateRoleSection**: present and non-empty
5. **validateInstructionsSection**: present with content
6. **validateExamples**: at least 2 examples with input/output
7. **validateConstraints**: list of specific boundaries
8. **validateEdgeCases**: handles unclear/missing/out-of-scope
9. **validateOutputFormat**: specified for structured outputs

## Research-Backed Template

### Synthesized from Anthropic Documentation

```xml
---
name: skill-name
description: What it does and when to use it. Auto-activates for: trigger1, trigger2.
version: 2.0.0
author: go-ent
tags: [category, domain]
---

# Skill Name

<role>
You are [specific expert role] with expertise in [domain].
[Behavioral guidelines and constraints from project standards]
</role>

<instructions>
When [triggering condition]:
1. [Step with clear action verb]
2. [Step with clear action verb]
3. [Step with clear action verb]

[Output format requirements]
</instructions>

<constraints>
- [What to include - concrete specification]
- [What to exclude - concrete specification]
- [Boundaries - concrete limits]
</constraints>

<edge_cases>
If input is unclear: [specific handling]
If information is missing: [specific handling]
If request is out of scope: [specific delegation or rejection]
</edge_cases>

<examples>
<example>
<input>[Representative user request or scenario]</input>
<output>[Expected response following instructions]</output>
</example>
<example>
<input>[Edge case or boundary condition]</input>
<output>[Appropriate handling demonstration]</output>
</example>
</examples>

<output_format>
[Exact specification of expected output structure]
[Format constraints and requirements]
</output_format>

## Reference Content
[Existing tables, code blocks, and quick reference material]
```

## Key Insights

1. **XML structure correlates with performance**: 15-20% improvement is significant enough to justify refactoring effort

2. **Role definition is critical**: Anthropic calls it "most powerful" system prompt technique - current skills lack this entirely

3. **Examples drive consistency**: Input/output pairs show Claude exactly what's expected - current code snippets don't provide this

4. **Validation enables quality**: Without validation, skills will drift from best practices over time

5. **Template-first reduces risk**: Validating approach on one skill before migrating 14 reduces chance of wasted effort

6. **Backward compatibility is essential**: Supporting both v1 and v2 formats during migration prevents breaking existing deployments

## Recommendations

### Immediate Actions

1. ✅ Extend parser for v2 frontmatter (version, author, tags)
2. ✅ Create validator with 9 research-backed rules
3. ✅ Implement quality scorer using rubric
4. ✅ Build MCP tools for validation and quality reporting

### Template Creation

1. ✅ Select go-code as template (highest impact, clear domain)
2. ✅ Refactor to v2 format with all XML sections
3. ✅ Validate achieves quality score >= 90
4. ✅ Use as reference for remaining 13 skills

### Migration Approach

1. ✅ Tier 1: High-value Go skills (go-arch, go-test, go-db, go-api)
2. ✅ Tier 2: Specialized Go skills (go-sec, go-perf, go-review, go-ops)
3. ✅ Tier 3: Core skills (api-design, arch-core, debug-core, review-core, security-core)

### Success Metrics

- All skills achieve quality score >= 80
- Zero validation errors in strict mode
- Template (go-code) achieves score >= 90
- Migration completed without breaking existing functionality

## References

- [Anthropic Documentation](https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering) - Official prompt engineering guidance
- `docs/research/SKILL.md` - Project's comprehensive skill authoring guide
- `docs/research/AGENT.md` - Multi-agent architecture patterns
- `internal/spec/validator.go` - Existing validation pattern to follow
