# Skill Authoring Guide (v3 Format)

Complete guide for creating high-quality skills using the v3 format (Markdown sections + YAML frontmatter) for go-ent's plugin system.

---

## Table of Contents

- [Overview](#overview)
- [Complete Skill Template](#complete-skill-template)
- [Frontmatter Fields](#frontmatter-fields)
- [Markdown Sections](#markdown-sections)
- [Validation Rules](#validation-rules)
- [Quality Scoring](#quality-scoring)
- [Good vs Bad Examples](#good-vs-bad-examples)
- [Migration from v2](#migration-from-v2)
- [Best Practices](#best-practices)
- [Quick Reference](#quick-reference)

---

## Overview

The v3 skill format aligns with Claude Code official patterns and uses:

- **YAML frontmatter** - Metadata including triggers
- **Markdown sections** - `## Role`, `## Instructions`, etc. (no XML tags)
- **Trigger definitions** - In frontmatter, not in body
- **Automatic validation** - Schema validation and quality scoring
- **Backward compatibility** - v2 skills (XML tags) still work

### v3 vs v2 Format

| Feature | v2 Format | v3 Format |
|---------|-----------|-----------|
| Metadata | YAML frontmatter | YAML frontmatter |
| Triggers | XML `<triggers>` in body | YAML `triggers:` in frontmatter |
| Role | `<role>` XML tag | `## Role` Markdown heading |
| Instructions | `<instructions>` XML tag | `## Instructions` Markdown heading |
| Constraints | `<constraints>` XML tag | `## Constraints` Markdown heading |
| Edge Cases | `<edge_cases>` XML tag | `## Edge Cases` Markdown heading |
| Examples | `<examples>` XML tag | `## Examples` Markdown heading |
| Output Format | `<output_format>` XML tag | `## Output Format` Markdown heading |

---

## Complete Skill Template

Here's a complete template for a v3 skill:

```markdown
---
name: your-skill-name
description: "Skill description for activation decisions"
version: "1.0.0"
disable-model-invocation: false
user-invocable: true
allowed-tools:
  - Read
  - Grep
  - Edit
triggers:
  keywords:
    - trigger1
    - trigger2
    - trigger3
  file_pattern: "*.go"
  weight: 0.8
---

## Role

Expert persona definition with domain expertise and behavioral guidelines.

## Instructions

### Pattern 1

Code or content example with explanation.

**Why this pattern**:
- Reason 1
- Reason 2

### Pattern 2

Another example with clear explanation.

**Rules**:
- Rule 1
- Rule 2

## Constraints

- Include specific patterns or approaches
- Include required output format elements
- Exclude anti-patterns or discouraged practices
- Exclude certain implementation details

## Edge Cases

If input is unclear: Ask clarifying questions before proceeding.

If context is missing: Request additional information about architecture.

If performance concerns arise: Delegate to performance skill.

If architecture questions emerge: Delegate to architecture skill.

If testing requirements are needed: Delegate to testing skill.

## Examples

### Example 1: Feature Implementation

**Input**: Example user request

**Output**:
```go
// Expected code example
```

### Example 2: Refactoring

**Input**: Another example request

**Output**:
```go
// Another code example
```

## Output Format

Provide output following these guidelines:

1. **Format requirement 1**: Specific format instruction
2. **Format requirement 2**: Another format instruction
3. **Quality criteria**: What makes output high-quality

Focus on practical, actionable guidance with minimal abstractions.
```

---

## Frontmatter Fields

### Required Fields

| Field | Description | Example |
|-------|-------------|---------|
| `name` | Skill identifier (lowercase, hyphens) | `go-code` |
| `description` | Brief summary for invocation decisions | `"Modern Go patterns for implementation"` |

### Optional Fields (Claude Code)

| Field | Description | Example |
|-------|-------------|---------|
| `version` | Semantic version | `"1.0.0"` |
| `disable-model-invocation` | Prevent automatic invocation | `false` |
| `user-invocable` | Allow manual invocation | `true` |
| `allowed-tools` | Tools skill can use | `["Read", "Grep"]` |
| `triggers` | Activation triggers (see below) | See Triggers section |

### Triggers (v3 Format)

Triggers moved from body to frontmatter:

```yaml
triggers:
  keywords:
    - error handling
    - error wrapping
    - fmt.Errorf
  file_pattern: "*.go"
  weight: 0.8
```

**Fields**:
- `keywords` - Array of trigger phrases
- `file_pattern` - Glob pattern (e.g., "*.go", "*.ts")
- `weight` - Activation weight 0.0-1.0 (default: 0.5)

**Best practices**:
- 3+ keywords recommended
- Use specific, relevant phrases
- Match user terminology
- Weight 0.8+ for high priority

---

## Markdown Sections

### `## Role` - Expert Persona Definition

Define the AI's expertise and behavioral guidelines:

```markdown
## Role

Expert Go developer focused on clean architecture, patterns, and idioms.
Prioritize SOLID, DRY, KISS, YAGNI principles with production-grade quality,
maintainability, and performance.
```

**Purpose**: Sets the persona and expertise level
**Content**: Expert identity, principles to follow, quality expectations
**Best practices**:
- 1-2 sentences defining expertise
- Include behavioral guidelines
- Mention key principles or standards
- Keep concise and focused

---

### `## Instructions` - Core Knowledge and Patterns

Provide detailed, actionable guidance:

```markdown
## Instructions

### Pattern Name

```go
func example() {
    // Code example
}
```

**Why this pattern**:
- Reason 1
- Reason 2

### Another Pattern

Explanation with code blocks and rules.

**Rules**:
- Rule 1
- Rule 2
```

**Purpose**: Core knowledge and patterns
**Content**: Code examples, explanations, rules, patterns
**Format**: Markdown with code blocks, lists, emphasis
**Best practices**:
- Use code blocks with language tags
- Include "Why this pattern" sections
- Use bullet lists for rules
- Group related patterns together
- Use subsections (###) for organization

---

### `## Constraints` - Boundaries and Requirements

Define what to include and exclude:

```markdown
## Constraints

- Include clean, idiomatic Go code following standard conventions
- Include proper error wrapping with context using `%w` verb
- Include context propagation as first parameter throughout layers
- Exclude magic numbers (use named constants instead)
- Exclude global mutable state (pass dependencies explicitly)
- Exclude panic in production code (use error handling instead)
```

**Purpose**: Set clear boundaries and requirements
**Content**: Include rules, exclude rules
**Format**: Bullet list starting with "Include" or "Exclude"
**Best practices**:
- Start each line with "Include" or "Exclude"
- Cover both positive and negative constraints
- Be specific about what's allowed/disallowed
- Use code formatting for technical terms

---

### `## Edge Cases` - Edge Case Handling

Document 5+ scenarios with handling instructions:

```markdown
## Edge Cases

If input is unclear or ambiguous: Ask clarifying questions to understand
the specific requirement before proceeding with implementation.

If context is missing for a feature: Request additional information about
architecture decisions, existing patterns, or integration points.

If performance concerns arise: Delegate to go-perf skill for profiling,
optimization strategies, and benchmarking guidance.

If architecture questions emerge: Delegate to go-arch skill for system
design, layer boundaries, and structural decisions.

If testing requirements are needed: Delegate to go-test skill for test
coverage, table-driven tests, and mocking strategies.
```

**Purpose**: Handle edge cases and delegations
**Content**: 5+ scenarios with "If X: Y" format
**Format**: Each scenario on separate paragraph
**Best practices**:
- Use "If X: Y" format consistently
- Include delegation scenarios
- Cover common edge cases
- Be specific about handling actions

---

### `## Examples` - Input/Output Pairs

Provide 2-3 concrete examples using subsections:

```markdown
## Examples

### Example 1: Bootstrap Pattern

**Input**: Refactor main() to use bootstrap pattern with graceful shutdown

**Output**:
```go
func main() {
    if err := run(context.Background(), os.Getenv, os.Stdout, os.Stderr); err != nil {
        slog.Error("fatal", "error", err)
        os.Exit(1)
    }
}
```

### Example 2: Error Handling

**Input**: Fix error handling in this function

**Output**:
```go
// Before
func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
    return nil, err
}

// After
func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
    if err != nil {
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    return &user, nil
}
```
```

**Purpose**: Demonstrate skill application
**Content**: 2-3 examples with clear input/output
**Format**: Use subsections (###) with **Input**/**Output** labels
**Best practices**:
- Use realistic user requests
- Include complete, runnable code
- Show before/after comparisons when helpful
- Cover different use cases

---

### `## Output Format` - Expected Output Structure

Define expected output format:

```markdown
## Output Format

Provide production-ready Go code following established patterns:

1. **Code Structure**: Clean, idiomatic Go with proper package organization
2. **Naming**: Short, natural variable names (cfg, repo, ctx, req, resp)
3. **Error Handling**: Wrapped errors with lowercase context using `%w`
4. **Context**: Always first parameter, propagated through all layers
5. **Interfaces**: Minimal interfaces at consumer side, return structs

Focus on practical implementation with minimal abstractions unless
complexity demands it.
```

**Purpose**: Guide output structure and format
**Content**: Format requirements, structure expectations, emphasis
**Format**: Numbered list with bold headings
**Best practices**:
- Number key requirements
- Use bold for categories
- Focus on practical guidance
- Mention quality criteria

---

## Validation Rules

Skills are validated using format-aware rules. v3 skills are validated for Markdown sections instead of XML tags.

### Key Validation Checks

1. **Frontmatter validation** - Required: `name`, `description`
2. **Version format** - If present, must be semantic (1.0.0)
3. **Markdown sections** - Checks for `## Role`, `## Instructions`, etc.
4. **Trigger format** - Validates `triggers:` in frontmatter
5. **Content quality** - Section length, example count, edge cases

### Strict vs Non-Strict Mode

**Non-strict mode** (default):
- Allows warnings for some missing sections
- Valid if no errors (warnings ignored)
- Good for initial drafts

**Strict mode**:
- Treats warnings as errors
- All sections must be complete
- Valid only if zero issues
- Required for production skills

Enable strict mode:
```bash
make skill-validate strict=true
# or via MCP
Use skill_validate with skill_id="go-code", strict=true
```

---

## Quality Scoring

Quality scores range from 0-100 and are computed automatically:

### Scoring Breakdown

| Category | Points | Criteria |
|----------|--------|----------|
| **Frontmatter** | 20 | name, description, version, triggers |
| **Structure** | 30 | Role, Instructions, Examples sections |
| **Content** | 30 | Example count (2+), Edge cases (5+) |
| **Triggers** | 20 | Keyword count (3+ for full points) |
| **Total** | 100 | Maximum possible score |

### Quality Thresholds

| Score Range | Quality Level | Action |
|-------------|--------------|--------|
| ≥ 90 | Excellent | Template quality, ready for reference |
| 80 - 89 | Good | Acceptable for production |
| < 80 | Needs improvement | Add sections, examples, triggers |

**Target**: ≥ 80 for production skills, ≥ 90 for template/reference skills.

---

## Good vs Bad Examples

### Good Example (v3 Format)

```markdown
---
name: go-error
description: "Error handling patterns for Go"
version: "1.0.0"
user-invocable: true
triggers:
  keywords:
    - error handling
    - error wrapping
    - fmt.Errorf
    - error context
  file_pattern: "*.go"
  weight: 0.8
---

## Role

Expert Go error handling engineer specializing in error design patterns,
wrapping strategies, and production-grade error management.

## Instructions

### Error Wrapping Pattern

Always wrap errors with context using `%w`:

```go
if err != nil {
    return fmt.Errorf("query user %s: %w", id, err)
}
```

**Why this pattern**:
- Preserves error chain for errors.Is/As
- Adds operation context
- Enables error tracing

### Lowercase Messages

Error messages should be lowercase without trailing punctuation:

```go
// Good
return fmt.Errorf("query user: %w", err)

// Bad
return fmt.Errorf("Query user: %w", err)  // uppercase
return fmt.Errorf("query user.: %w", err) // trailing period
```

## Constraints

- Include proper error wrapping with %w
- Include lowercase error messages
- Include operation context in errors
- Exclude unwrapped errors
- Exclude uppercase error messages

## Edge Cases

If error is already wrapped: Don't double-wrap, check if it's already wrapped.

If error is domain error: Return typed error directly without wrapping.

If multiple errors occur: Use multierror package or return first critical error.

If error needs translation: Map internal errors to domain errors at boundary.

If logging is needed: Log at origin, not during propagation.

## Examples

### Example 1: Repository Error

**Input**: Add error handling to database query

**Output**:
```go
func (r *Repository) GetUser(id string) (*User, error) {
    var u User
    if err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Name); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    return &u, nil
}
```

### Example 2: Multi-Layer Wrapping

**Input**: Fix error propagation through layers

**Output**:
```go
// Repository layer
func (r *repo) FindByID(ctx context.Context, id string) (*User, error) {
    // ... db query ...
    if err != nil {
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    return user, nil
}

// UseCase layer
func (uc *useCase) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := uc.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get user: %w", err)
    }
    return user, nil
}

// Transport layer
func (h *handler) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.uc.GetUser(r.Context(), id)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            http.Error(w, "user not found", http.StatusNotFound)
            return
        }
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(user)
}
```

## Output Format

Provide error handling following these guidelines:

1. **Wrapping**: Use `%w` for all error propagation
2. **Context**: Add operation context to every wrapped error
3. **Messages**: Lowercase, no trailing punctuation
4. **Domain Errors**: Return typed errors at boundaries
5. **Checking**: Use errors.Is/As for type checking

Focus on production-grade error handling with full traceability.
```

**Why this is good**:
- Complete frontmatter with triggers in YAML
- Clear role definition
- Rich instructions with code examples
- Specific constraints (include/exclude)
- 5 edge case scenarios
- 2 realistic examples with clear input/output
- Clear output format guidelines
- Uses Markdown sections (no XML)
- **Score**: ~95/100 (excellent)

---

### Bad Example (Needs Improvement)

```markdown
---
name: my-skill
description: "Does some stuff"
---

This skill helps with things.

Write good code.

If something is wrong, try to fix it.

Example 1: Do this
Example 2: Do that
```

**Why this is bad**:
- No `version` field
- Vague description ("does some stuff")
- No triggers
- No Markdown sections
- Role section missing
- Instructions are too generic
- No constraints section
- Edge cases don't use "If X: Y" format
- Examples lack structure (no ## Examples section)
- No output format section
- **Score**: ~25/100 (needs major improvement)

**How to fix**:
1. Add `version`, `triggers` to frontmatter
2. Add specific description
3. Add `## Role` section with expertise definition
4. Add `## Instructions` with specific patterns
5. Add `## Constraints` with include/exclude rules
6. Add `## Edge Cases` with "If X: Y" format (5+ scenarios)
7. Add `## Examples` section with subsections
8. Add `## Output Format` section
9. Use code blocks for examples
10. Add 3+ trigger keywords

---

## Migration from v2

### Automated Migration

Use the migration script:

```bash
go run scripts/migrate-skills.go
```

The script:
1. Detects v2 skills (XML tags)
2. Extracts `<triggers>` from body
3. Moves triggers to frontmatter
4. Converts XML tags to Markdown sections
5. Validates migrated skills

### Manual Migration Steps

If migrating manually:

1. **Extract triggers** - Move from `<triggers>` in body to `triggers:` in frontmatter
2. **Convert tags to headings**:
   - `<role>` → `## Role`
   - `<instructions>` → `## Instructions`
   - `<constraints>` → `## Constraints`
   - `<edge_cases>` → `## Edge Cases`
   - `<examples>` → `## Examples`
   - `<output_format>` → `## Output Format`
3. **Update examples** - Use subsections (###) with **Input**/**Output** labels
4. **Validate** - Run `make skill-validate strict=true`

### Example Migration

**Before (v2)**:
```markdown
---
name: go-error
description: "Error handling. Auto-activates for: error handling, wrapping"
version: "2.0.0"
---

<triggers>
  keywords:
    - "error handling"
  weight: 0.8
</triggers>

<role>
Expert Go error handling engineer.
</role>

<instructions>
## Pattern
Code here
</instructions>
```

**After (v3)**:
```markdown
---
name: go-error
description: "Error handling patterns for Go"
version: "2.0.0"
triggers:
  keywords:
    - error handling
    - error wrapping
  weight: 0.8
---

## Role

Expert Go error handling engineer.

## Instructions

### Pattern

Code here
```

---

## Best Practices

### 1. Use Specific, Actionable Instructions

Ambiguity causes failures. Be explicit.

**Bad**:
```
Write professional code.
```

**Good**:
```
Write clean, idiomatic Go following SOLID principles.
Include proper error wrapping with context using %w.
Use short variable names (cfg, repo, ctx) in small scopes.
```

### 2. Provide Rich Examples

Examples demonstrate patterns effectively.

**Best practices**:
- 2-3 diverse examples
- Include edge cases
- Show complete, runnable code
- Use realistic scenarios

### 3. Document Clear Constraints

Explicit constraints prevent incorrect actions.

**Example**:
```markdown
## Constraints

- Include clean, idiomatic Go code
- Include proper error wrapping with %w
- Exclude magic numbers (use named constants)
- Exclude global mutable state
```

### 4. Use 3+ Trigger Keywords

More triggers = better activation.

**Example**:
```yaml
triggers:
  keywords:
    - error handling
    - error wrapping
    - fmt.Errorf
    - error context
  weight: 0.8
```

### 5. Keep Instructions Concise

Every token consumes attention budget.

**Best practices**:
- Challenge each instruction: "Is this needed?"
- Remove until model misbehaves
- Focus on communication, not cleverness

---

## Quick Reference

### Minimal v3 Template

```markdown
---
name: skill-name
description: "Brief description"
version: "1.0.0"
triggers:
  keywords:
    - trigger1
    - trigger2
    - trigger3
  weight: 0.8
---

## Role

Expert [domain] focused on [specialty]. Prioritize [principles].

## Instructions

### Pattern 1

Code or content example.

**Why this pattern**:
- Reason 1
- Reason 2

## Constraints

- Include specific patterns
- Exclude anti-patterns

## Edge Cases

If input is unclear: Ask clarifying questions.

If context is missing: Request additional information.

If [situation]: [action].

## Examples

### Example 1

**Input**: User request

**Output**:
```go
// Code example
```

### Example 2

**Input**: Another request

**Output**:
```go
// Another example
```

## Output Format

Provide output following these guidelines:

1. **Requirement 1**: Specific instruction
2. **Requirement 2**: Another instruction

Focus on practical guidance.
```

### Validation Commands

```bash
# Validate all skills (non-strict)
make skill-validate

# Validate all skills (strict mode)
make skill-validate strict=true

# Validate specific skill via MCP
Use skill_validate with skill_id="go-code", strict=true

# Generate quality report
make skill-quality

# Check specific skill quality
Use skill_quality with skill_id="go-code", threshold=80
```

---

## Resources

- **[AGENTS_AND_SKILLS.md](./AGENTS_AND_SKILLS.md)** - v3 agent and skill architecture
- **[MIGRATION_V3.md](./MIGRATION_V3.md)** - Migration from v2 to v3
- **[CLAUDE_CODE_COMPATIBILITY.md](./CLAUDE_CODE_COMPATIBILITY.md)** - Claude Code alignment
- **Example Skills**: `plugins/go-ent/skills/*/SKILL.md`
- **Reference Skills**: `plugins/go-ent/skills/ent/*/SKILL.md`
- **Validation Code**: `internal/skill/validator.go`, `internal/skill/rules.go`
- **Scoring Code**: `internal/skill/scorer.go`

---

## Troubleshooting

### Validation Fails

**Problem**: Validation errors in strict mode

**Solutions**:
- Check for Markdown sections (not XML tags)
- Ensure triggers in frontmatter (not body)
- Verify frontmatter has required fields
- Use `make skill-validate` to see specific errors

### Low Quality Score

**Problem**: Quality score < 80

**Solutions**:
- Add missing frontmatter fields (version, triggers)
- Ensure all Markdown sections are present
- Add more examples (target 2-3)
- Add more edge cases (target 5+)
- Add more triggers (target 3+)
- Check `make skill-quality` for breakdown

### Skill Doesn't Activate

**Problem**: Skill doesn't auto-activate

**Solutions**:
- Check `triggers.keywords` in frontmatter
- Ensure keywords match user terminology
- Add more keywords (3+ recommended)
- Increase weight (0.8+ for priority)
- Test with specific trigger words

### Examples Don't Help

**Problem**: Examples don't guide output effectively

**Solutions**:
- Use realistic user requests
- Include complete, runnable code
- Show before/after for refactoring
- Cover different use cases
- Use clear subsections with Input/Output labels

---

**Version**: 3.0.0
**Last Updated**: 2026-01-28
