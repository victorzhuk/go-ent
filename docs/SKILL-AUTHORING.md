# Skill Authoring Guide (v2 Format)

This guide explains how to create high-quality skills using the v2 format for go-ent's plugin system.

## Overview

The v2 skill format provides structured, validated, and high-quality skill definitions with automatic quality scoring. Skills in v2 format include:

- **Required XML sections**: `<role>`, `<instructions>`, `<constraints>`, `<edge_cases>`, `<examples>`, `<output_format>`
- **Enhanced frontmatter**: `version`, `author`, `tags` fields
- **Validation**: Automatic checking for required sections and content
- **Quality scoring**: 0-100 scale with detailed breakdown
- **MCP tools**: `skill_validate` and `skill_quality` for inspection

## Complete Skill Template

Here's a complete template for a v2 skill:

```markdown
---
name: your-skill-name
description: "Skill description. Auto-activates for: trigger1, trigger2, trigger3"
version: "2.0.0"
author: "your-name"
tags: ["category", "keyword", "topic"]
---

# Skill Title

<role>
Expert persona definition with domain expertise and behavioral guidelines.
</role>

<instructions>

## Pattern 1

Code or content example with explanation.

**Why this pattern**:
- Reason 1
- Reason 2

## Pattern 2

Another example with clear explanation.

**Rules**:
- Rule 1
- Rule 2

</instructions>

<constraints>
- Include specific patterns or approaches
- Include required output format elements
- Exclude anti-patterns or discouraged practices
- Exclude certain implementation details
- Bound to specific architectural principles
</constraints>

<edge_cases>
If input is unclear: Ask clarifying questions before proceeding.

If context is missing: Request additional information about architecture, patterns, or integration.

If performance concerns arise: Delegate to performance skill for profiling and optimization.

If architecture questions emerge: Delegate to architecture skill for system design guidance.

If testing requirements are needed: Delegate to testing skill for test coverage strategies.
</edge_cases>

<examples>
<example>
<input>Example user request or input</input>
<output>
Expected output or response
</output>
</example>

<example>
<input>Another example request</input>
<output>
Another expected response
</output>
</example>
</examples>

<output_format>
Provide output following these guidelines:

1. **Format requirement 1**: Specific format instruction
2. **Format requirement 2**: Another format instruction
3. **Quality criteria**: What makes output high-quality

Focus on practical, actionable guidance with minimal abstractions.
</output_format>
```

## Frontmatter Fields

| Field      | Required | Description                               | Example                          |
|------------|----------|-------------------------------------------|----------------------------------|
| `name`     | Yes      | Skill identifier (lowercase, hyphens)      | `go-code`                        |
| `description` | Yes   | What skill does + auto-activation triggers  | `"Modern Go patterns. Auto-activates for: writing code, implementing features"` |
| `version`  | No       | Semantic version (recommended for v2)      | `"2.0.0"`                        |
| `author`   | No       | Attribution                               | `"go-ent"`                       |
| `tags`     | No       | Categorization array (YAML list)          | `["go", "code", "implementation"]` |

### Triggers

Auto-activation triggers are extracted from the `description` field:

- Format: `"description text. Auto-activates for: trigger1, trigger2, trigger3"`
- Alternative: `"description text. Activates when: trigger1, trigger2"`
- Minimum: 1 trigger required
- Recommended: 3+ triggers for better activation

## XML Sections

### `<role>` - Expert Persona Definition

Define the AI's expertise and behavioral guidelines:

```xml
<role>
Expert Go developer focused on clean architecture, patterns, and idioms. Prioritize SOLID, DRY, KISS, YAGNI principles with production-grade quality, maintainability, and performance.
</role>
```

**Purpose**: Sets the persona and expertise level
**Content**: Expert identity, principles to follow, quality expectations
**Best practices**:
- 1-2 sentences defining expertise
- Include behavioral guidelines (what to prioritize)
- Mention key principles or standards
- Keep it concise and focused

### `<instructions>` - Core Knowledge and Patterns

Provide detailed, actionable guidance:

```xml
<instructions>

## Pattern Name

```go
func example() {
    // Code example
}
```

**Why this pattern**:
- Reason 1
- Reason 2

## Another Pattern

Explanation with code blocks and rules.

**Rules**:
- Rule 1
- Rule 2

</instructions>
```

**Purpose**: Core knowledge and patterns
**Content**: Code examples, explanations, rules, patterns
**Format**: Markdown with code blocks, lists, emphasis
**Best practices**:
- Use code blocks with language tags
- Include "Why this pattern" sections
- Use bullet lists for rules
- Group related patterns together

### `<constraints>` - Boundaries and Requirements

Define what to include and exclude:

```xml
<constraints>
- Include clean, idiomatic Go code following standard conventions
- Include proper error wrapping with context using `%w` verb
- Include context propagation as first parameter throughout layers
- Exclude magic numbers (use named constants instead)
- Exclude global mutable state (pass dependencies explicitly)
- Exclude panic in production code (use error handling instead)
- Bound to clean layered architecture: Transport → UseCase → Domain ← Repository
</constraints>
```

**Purpose**: Set clear boundaries and requirements
**Content**: Include rules, exclude rules, architectural boundaries
**Format**: Bullet list starting with "Include" or "Exclude"
**Best practices**:
- Start each line with "Include" or "Exclude"
- Cover both positive and negative constraints
- Mention architectural boundaries
- Be specific about what's allowed/disallowed

### `<edge_cases>` - Edge Case Handling

Document 5+ scenarios with handling instructions:

```xml
<edge_cases>
If input is unclear or ambiguous: Ask clarifying questions to understand the specific requirement before proceeding with implementation.

If context is missing for a feature: Request additional information about architecture decisions, existing patterns, or integration points.

If performance concerns arise: Delegate to go-perf skill for profiling, optimization strategies, and benchmarking guidance.

If architecture questions emerge: Delegate to go-arch skill for system design, layer boundaries, and structural decisions.

If testing requirements are needed: Delegate to go-test skill for test coverage, table-driven tests, and mocking strategies.

If security considerations are relevant: Delegate to go-sec skill for authentication, authorization, and input validation patterns.
</edge_cases>
```

**Purpose**: Handle edge cases and delegations
**Content**: 5+ scenarios with "If X: Y" format
**Format**: Each scenario on separate line
**Best practices**:
- Use "If X: Y" format consistently
- Include delegation scenarios
- Cover common edge cases
- Be specific about handling actions

### `<examples>` - Input/Output Pairs

Provide 2-3 concrete examples:

```xml
<examples>
<example>
<input>Refactor main() to use bootstrap pattern with graceful shutdown</input>
<output>
```go
func main() {
    if err := run(context.Background(), os.Getenv, os.Stdout, os.Stderr); err != nil {
        slog.Error("fatal", "error", err)
        os.Exit(1)
    }
}
```
</example>

<example>
<input>Fix error handling in this function - it's not wrapping errors properly</input>
<output>
```go
// Before
func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    return nil, err
}

// After
func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    if err != nil {
        return fmt.Errorf("query user %s: %w", id, err)
    }
}
```
</example>

<example>
<input>Implement repository pattern with proper error handling and domain mapping</input>
<output>
```go
package userrepo

import (
    "context"
    "fmt"
)

type repository struct {
    db *sqlx.DB
}

func New(db *sqlx.DB) *repository {
    return &repository{db: db}
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
    // Implementation
}
```
</example>
</examples>
```

**Purpose**: Demonstrate skill application
**Content**: 2-3 examples with `<input>` and `<output>` tags
**Format**: Realistic user requests and responses
**Best practices**:
- Use realistic user requests
- Include complete, runnable code
- Show before/after comparisons when helpful
- Cover different use cases

### `<output_format>` - Expected Output Structure

Define expected output format:

```xml
<output_format>
Provide production-ready Go code following established patterns:

1. **Code Structure**: Clean, idiomatic Go with proper package organization
2. **Naming**: Short, natural variable names (cfg, repo, ctx, req, resp)
3. **Error Handling**: Wrapped errors with lowercase context using `%w`
4. **Context**: Always first parameter, propagated through all layers
5. **Interfaces**: Minimal interfaces at consumer side, return structs

Focus on practical implementation with minimal abstractions unless complexity demands it.
</output_format>
```

**Purpose**: Guide output structure and format
**Content**: Format requirements, structure expectations, emphasis
**Format**: Clear, actionable guidelines
**Best practices**:
- Number key requirements
- Use emphasis for important points
- Focus on practical guidance
- Mention quality criteria

## Validation Rules

Skills are validated using 9 rules. All rules run in both non-strict and strict modes.

### Rule 1: validateFrontmatter

Checks required frontmatter fields.

**What it checks**:
- `name` field is present (error if missing)
- `description` field is present (error if missing)
- `version` field for v2 skills (warning in non-strict, error in strict)

**Examples**:

Good:
```yaml
---
name: go-code
description: "Modern Go patterns. Auto-activates for: writing code, implementing features"
version: "2.0.0"
author: "go-ent"
tags: ["go", "code"]
---
```

Bad (missing version):
```yaml
---
name: my-skill
description: "Does something"
---
```

Bad (missing name):
```yaml
---
description: "Does something"
---
```

**How to fix**: Add missing fields to frontmatter. For v2 skills, include `version`, `author`, and `tags`.

---

### Rule 2: validateVersion

Checks semantic version format.

**What it checks**:
- Version field matches `v1.0.0` or `1.0.0` format (semver)
- Only runs if `version` field is present

**Examples**:

Good:
```yaml
version: "2.0.0"
version: "1.2.3"
version: "v3.4.5"
```

Bad:
```yaml
version: "2.0"
version: "latest"
version: "v2"
```

**How to fix**: Use semantic versioning: `MAJOR.MINOR.PATCH`

---

### Rule 3: validateXMLTags

Checks for balanced XML tags.

**What it checks**:
- All XML tags have matching open/close tags
- No duplicate top-level tags
- Only checks v2 skills

**Examples**:

Good:
```xml
<role>...</role>
<instructions>...</instructions>
```

Bad (unbalanced):
```xml
<role>...
<!-- Missing </role> -->
```

Bad (duplicate):
```xml
<role>...</role>
<role>...</role>
```

**How to fix**: Ensure every `<tag>` has a matching `</tag>` and no duplicates.

---

### Rule 4: validateRoleSection

Checks `<role>` section presence and content.

**What it checks**:
- `<role>` section exists (warning in non-strict, error in strict for v2)
- `<role>` tag is closed with `</role>`
- Role section is not empty
- Role section has at least 2 lines of content (warning)

**Examples**:

Good:
```xml
<role>
Expert Go developer focused on clean architecture, patterns, and idioms.
Prioritize SOLID, DRY, KISS, YAGNI principles.
</role>
```

Bad (missing):
```xml
<!-- No <role> section -->
```

Bad (empty):
```xml
<role>

</role>
```

Bad (too short):
```xml
<role>
Expert.
</role>
```

**How to fix**: Add `<role>` section with 2+ lines defining expertise and behavioral guidelines.

---

### Rule 5: validateInstructionsSection

Checks `<instructions>` section presence.

**What it checks**:
- `<instructions>` section exists (warning in non-strict, error in strict for v2)
- `<instructions>` tag is closed with `</instructions>`

**Examples**:

Good:
```xml
<instructions>
## Pattern 1
Code example...
</instructions>
```

Bad (missing):
```xml
<!-- No <instructions> section -->
```

Bad (unclosed):
```xml
<instructions>
## Pattern 1
Code example...
<!-- Missing </instructions> -->
```

**How to fix**: Add `<instructions>` section with patterns, examples, and rules.

---

### Rule 6: validateExamples

Checks `<examples>` section structure.

**What it checks**:
- `<examples>` tag is closed
- `<examples>` contains at least one `<example>` tag (warning)
- Each `<example>` has `<input>` and `<output>` tags (error)

**Examples**:

Good:
```xml
<examples>
<example>
<input>User request</input>
<output>Response</output>
</example>
</examples>
```

Bad (no input/output):
```xml
<examples>
<example>
Just text without tags
</example>
</examples>
```

Bad (no examples):
```xml
<examples>
<!-- No <example> tags -->
</examples>
```

**How to fix**: Ensure each `<example>` has `<input>` and `<output>` tags with proper nesting.

---

### Rule 7: validateConstraints

Checks `<constraints>` section format.

**What it checks**:
- `<constraints>` tag is closed
- Constraints items use list format (start with `- `) (warning)
- Constraints section is not empty (warning)

**Examples**:

Good:
```xml
<constraints>
- Include clean code patterns
- Exclude anti-patterns
- Bound to specific principles
</constraints>
```

Bad (no list format):
```xml
<constraints>
Include clean code patterns.
Exclude anti-patterns.
</constraints>
```

Bad (empty):
```xml
<constraints>

</constraints>
```

**How to fix**: Use bullet list format starting with `- ` for each constraint.

---

### Rule 8: validateEdgeCases

Checks `<edge_cases>` section scenarios.

**What it checks**:
- `<edge_cases>` tag is closed
- At least 2 scenarios using 'if', 'when', or 'should' keywords (warning)

**Examples**:

Good:
```xml
<edge_cases>
If input is unclear: Ask clarifying questions.
If context is missing: Request additional information.
When performance is a concern: Delegate to performance skill.
Should security arise: Delegate to security skill.
```
</edge_cases>
```

Bad (no scenarios):
```xml
<edge_cases>
No scenarios defined.
</edge_cases>
```

**How to fix**: Add scenarios using "If X: Y" or "When X: Y" format.

---

### Rule 9: validateOutputFormat

Checks `<output_format>` section for v2 skills.

**What it checks**:
- `<output_format>` section exists (warning in non-strict, error in strict for v2)
- `<output_format>` tag is closed
- Output format section is not empty (warning)

**Examples**:

Good:
```xml
<output_format>
Provide production-ready code following these guidelines:

1. **Structure**: Clean, idiomatic code
2. **Naming**: Short, natural variable names
3. **Errors**: Wrapped with context

Focus on practical implementation.
</output_format>
```

Bad (empty):
```xml
<output_format>

</output_format>
```

**How to fix**: Add `<output_format>` section with specific output guidelines.

---

## Strict vs Non-Strict Mode

**Non-strict mode** (default):
- Allows warnings for some missing sections
- Valid if no errors (warnings are ignored)
- Good for initial drafts

**Strict mode**:
- Treats warnings as errors
- All sections must be complete
- Valid only if zero issues
- Required for production skills

Enable strict mode:
```bash
make skill-validate strict=true
# or
Use skill_validate with skill_id="go-code", strict=true
```

## Quality Scoring Rubric

Quality scores range from 0-100 and are computed automatically:

### Frontmatter (20 points)

| Component   | Points | Criteria                                  |
|------------|--------|-------------------------------------------|
| `name`     | 5      | Non-empty skill name                        |
| `description` | 5   | Non-empty description                       |
| `version`  | 5      | Version field present                        |
| `tags`     | 5      | Tags array has at least one element         |

**Max: 20 points**

### Structure (30 points)

| Section       | Points | Criteria                            |
|---------------|--------|-------------------------------------|
| `<role>`      | 10     | Role section present                  |
| `<instructions>` | 10  | Instructions section present          |
| `<examples>`  | 10     | Examples section present              |

**Max: 30 points**

### Content (30 points)

| Component      | Points | Criteria                                   |
|----------------|--------|--------------------------------------------|
| Examples count | 15     | 2+ examples (10 points for 1 example)      |
| `<edge_cases>` | 15    | Edge cases section present                  |

**Max: 30 points**

### Triggers (20 points)

| Trigger Count | Points | Calculation                        |
|---------------|--------|-----------------------------------|
| 0             | 0      | No triggers                        |
| 1             | 6.67   | 1 × 6.67                          |
| 2             | 13.33  | 2 × 6.67                          |
| 3+            | 20     | Full points                        |

**Max: 20 points**

### Total Score

**Max: 100 points**

### Thresholds

| Score Range    | Quality Level           | Action                           |
|---------------|------------------------|----------------------------------|
| ≥ 90          | Excellent              | Template quality, ready for reference |
| 80 - 89       | Good                  | Acceptable for production         |
| < 80          | Needs improvement      | Add sections, examples, triggers |

**Target**: ≥ 80 for production skills, ≥ 90 for template/reference skills.

## Good vs Bad Patterns

### Good Pattern Example

```markdown
---
name: go-code
description: "Modern Go implementation patterns. Auto-activates for: writing Go code, implementing features, refactoring, error handling, configuration"
version: "2.0.0"
author: "go-ent"
tags: ["go", "code", "implementation"]
---

# Go Code Patterns

<role>
Expert Go developer focused on clean architecture, patterns, and idioms. Prioritize SOLID, DRY, KISS, YAGNI principles with production-grade quality, maintainability, and performance.
</role>

<instructions>

## Bootstrap Pattern

```go
func main() {
    if err := run(context.Background(), os.Getenv, os.Stdout, os.Stderr); err != nil {
        slog.Error("fatal", "error", err)
        os.Exit(1)
    }
}
```

**Why this pattern**:
- Testable (injectable dependencies)
- Graceful shutdown (30s timeout)
- Proper signal handling

## Error Handling

```go
if err != nil {
    return fmt.Errorf("query user %s: %w", id, err)
}
```

**Rules**:
- Always wrap with context
- Lowercase, no trailing punctuation
- Use `%w` for wrapping

</instructions>

<constraints>
- Include clean, idiomatic Go code following standard conventions
- Include proper error wrapping with context using `%w` verb
- Exclude magic numbers (use named constants instead)
- Exclude global mutable state (pass dependencies explicitly)
- Bound to clean layered architecture: Transport → UseCase → Domain ← Repository
</constraints>

<edge_cases>
If input is unclear: Ask clarifying questions before proceeding.

If context is missing: Request additional information about architecture decisions.

If performance concerns arise: Delegate to go-perf skill for profiling.

If architecture questions emerge: Delegate to go-arch skill for system design.

If testing requirements are needed: Delegate to go-test skill for test coverage.
</edge_cases>

<examples>
<example>
<input>Refactor main() to use bootstrap pattern with graceful shutdown</input>
<output>
```go
func main() {
    if err := run(context.Background(), os.Getenv, os.Stdout, os.Stderr); err != nil {
        slog.Error("fatal", "error", err)
        os.Exit(1)
    }
}

func run(ctx context.Context, getenv func(string) string, stdout, stderr io.Writer) error {
    cfg, err := config.LoadFromEnv(getenv)
    if err != nil {
        return fmt.Errorf("config: %w", err)
    }
    // ... rest of implementation
}
```
</example>

<example>
<input>Fix error handling in this function</input>
<output>
```go
// Before
func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
    return nil, err
}

// After
func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
    if err != nil {
        return fmt.Errorf("query user %s: %w", id, err)
    }
}
```
</example>
</examples>

<output_format>
Provide production-ready Go code following established patterns:

1. **Code Structure**: Clean, idiomatic Go with proper package organization
2. **Naming**: Short, natural variable names (cfg, repo, ctx, req, resp)
3. **Error Handling**: Wrapped errors with lowercase context using `%w`
4. **Context**: Always first parameter, propagated through all layers

Focus on practical implementation with minimal abstractions.
</output_format>
```

**Why this is good**:
- Complete frontmatter with all fields
- Clear, concise role definition
- Rich instructions with multiple patterns
- Specific constraints (include/exclude)
- 5 edge case scenarios
- 2 realistic examples with input/output
- Clear output format guidelines
- **Score**: ~95/100 (excellent)

---

### Bad Pattern Example

```markdown
---
name: my-skill
description: "Does some stuff"
---

# My Skill

This skill helps with things.

## Some Instructions

Write good code.

## Edge Cases

If something is wrong, try to fix it.

## Examples

Example 1: Do this
Example 2: Do that
```

**Why this is bad**:
- Missing `version`, `author`, `tags` fields
- No XML tags (v1 format, but v2 expected)
- Vague description ("does some stuff")
- No triggers in description
- Role section missing
- Instructions are too generic
- No constraints section
- Edge cases section doesn't use "If X: Y" format
- Examples lack `<input>`/`<output>` tags
- No output format section
- **Score**: ~35/100 (needs major improvement)

**How to fix**:
1. Add frontmatter fields (`version`, `author`, `tags`)
2. Add triggers to description
3. Wrap sections in XML tags
4. Add `<role>` section with expertise definition
5. Expand `<instructions>` with specific patterns
6. Add `<constraints>` with include/exclude rules
7. Format edge cases as "If X: Y"
8. Add `<examples>` with proper tags
9. Add `<output_format>` section

---

## Migration Guide

### Step-by-Step Process

#### Step 1: Use go-code as Template

Copy the go-code skill as a starting point:

```bash
mkdir -p plugins/go-ent/skills/your-category/your-skill
cp plugins/go-ent/skills/go/go-code/SKILL.md plugins/go-ent/skills/your-category/your-skill/SKILL.md
```

#### Step 2: Update Frontmatter

Edit the frontmatter with your skill's information:

```yaml
---
name: your-skill-name
description: "Skill description. Auto-activates for: trigger1, trigger2, trigger3"
version: "2.0.0"
author: "your-name"
tags: ["category", "keyword", "topic"]
---
```

**Important**:
- `name`: lowercase with hyphens, max 64 characters
- `description`: what skill does + auto-activation triggers
- `version`: semantic version (e.g., `2.0.0`)
- `author`: attribution (e.g., "go-ent" or your name)
- `tags`: array of category keywords

#### Step 3: Update `<role>` Section

Define the expert persona:

```xml
<role>
Expert [domain] focused on [specialty]. Prioritize [principles] with [quality goals].
</role>
```

**Tips**:
- Keep it concise (1-2 sentences)
- Include domain expertise
- Mention principles to follow
- Define quality expectations

#### Step 4: Update `<instructions>` Section

Add your skill's core patterns and guidance:

```xml
<instructions>

## Pattern 1

Code or content example with explanation.

**Why this pattern**:
- Reason 1
- Reason 2

## Pattern 2

Another example with clear explanation.

**Rules**:
- Rule 1
- Rule 2

</instructions>
```

**Tips**:
- Use code blocks with language tags
- Include "Why this pattern" sections
- Group related patterns
- Use bullet lists for rules

#### Step 5: Update `<constraints>` Section

Define boundaries and requirements:

```xml
<constraints>
- Include specific patterns or approaches
- Include required output format elements
- Exclude anti-patterns or discouraged practices
- Exclude certain implementation details
- Bound to specific architectural principles
</constraints>
```

**Tips**:
- Start each line with "Include" or "Exclude"
- Cover both positive and negative constraints
- Mention architectural boundaries
- Be specific about what's allowed/disallowed

#### Step 6: Update `<edge_cases>` Section

Add 5+ edge case scenarios:

```xml
<edge_cases>
If input is unclear: Ask clarifying questions before proceeding.

If context is missing: Request additional information about architecture.

If performance concerns arise: Delegate to performance skill.

If architecture questions emerge: Delegate to architecture skill.

If testing requirements are needed: Delegate to testing skill.
</edge_cases>
```

**Tips**:
- Use "If X: Y" format consistently
- Include delegation scenarios
- Cover common edge cases
- Be specific about handling actions
- Target 5+ scenarios

#### Step 7: Update `<examples>` Section

Add 2-3 concrete examples:

```xml
<examples>
<example>
<input>Example user request</input>
<output>
```go
// Code example
```
</output>
</example>

<example>
<input>Another example request</input>
<output>
```go
// Another code example
```
</output>
</example>
</examples>
```

**Tips**:
- Use realistic user requests
- Include complete, runnable code
- Show before/after comparisons when helpful
- Cover different use cases
- Target 2-3 examples

#### Step 8: Update `<output_format>` Section

Define expected output format:

```xml
<output_format>
Provide output following these guidelines:

1. **Format requirement 1**: Specific instruction
2. **Format requirement 2**: Another instruction
3. **Quality criteria**: What makes output high-quality

Focus on practical, actionable guidance.
</output_format>
```

**Tips**:
- Number key requirements
- Use emphasis for important points
- Focus on practical guidance
- Mention quality criteria

#### Step 9: Validate with Strict Mode

Run validation in strict mode:

```bash
make skill-validate strict=true
```

Or use MCP tool:
```
Use skill_validate with skill_id="your-skill", strict=true
```

**Fix any validation errors** before proceeding.

#### Step 10: Check Quality Score

Generate quality report:

```bash
make skill-quality
```

Or use MCP tool:
```
Use skill_quality with skill_id="your-skill", threshold=80
```

**Quality targets**:
- ≥ 90: Template quality (recommended for reference skills)
- ≥ 80: Good quality (acceptable for production)
- < 80: Needs improvement

**If score < 80**:
- Add missing frontmatter fields (version, author, tags)
- Ensure all XML sections are present
- Add more examples (target 2-3)
- Add more edge cases (target 5+)
- Add more triggers in description (target 3+)

#### Step 11: Test with Real Work

Test the skill with actual work:

1. Trigger the skill with a relevant task
2. Verify skill content appears in context
3. Check output quality and relevance
4. Adjust if needed based on results

### Migration Checklist

- [ ] Copied go-code as template
- [ ] Updated frontmatter (name, description, version, author, tags)
- [ ] Updated `<role>` section with expert persona
- [ ] Updated `<instructions>` section with patterns
- [ ] Updated `<constraints>` section with include/exclude rules
- [ ] Updated `<edge_cases>` section with 5+ scenarios
- [ ] Updated `<examples>` section with 2-3 input/output pairs
- [ ] Updated `<output_format>` section with guidelines
- [ ] Validated with strict mode (`make skill-validate strict=true`)
- [ ] Quality score ≥ 80 (≥ 90 for templates)
- [ ] Tested with real work
- [ ] Skill triggers correctly
- [ ] Output quality meets expectations

### Backward Compatibility Notes

**v1 format** (no XML tags) still works:
- Detected by absence of `<role>` and `<instructions>` tags
- Loaded as legacy format
- No validation or quality scoring
- Can still be used, but won't benefit from v2 features

**v2 format**:
- Detected by presence of `<role>` or `<instructions>` tags
- Fully validated and scored
- Enhanced metadata (version, author, tags)
- Required for new skills

**Migration path**:
- Existing v1 skills can continue to work
- Migrate to v2 to get validation and quality scoring
- No breaking changes for existing skills

## Best Practices from Research

Based on research from `docs/research/SKILL.md`, here are proven practices for high-performance skills:

### 1. Use XML Tags for Structure

XML tags improve performance by **15-20%** when properly implemented.

**Best practices**:
- Use meaningful tag names that match content
- Nest tags for hierarchical content
- Reference tags explicitly in instructions
- Maintain consistent naming throughout

**Example**:
```xml
<role>...</role>
<instructions>
<examples>...</examples>
</instructions>
```

### 2. Provide Specific, Actionable Instructions

Ambiguity is the root cause of most skill failures.

**Best practices**:
- State everything explicitly; assume nothing
- Replace subjective terms with concrete specifications
- Test your skill by asking: "Could two people interpret this differently?"

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

### 3. Include Rich Examples with Input/Output

Examples are most valuable when you need consistent formatting or domain-specific output patterns.

**Best practices**:
- Provide 3-5 diverse, relevant examples
- Include edge cases in your examples
- Show boundary conditions and unexpected inputs
- Use the `<example>` tag with `<input>` and `<output>` subtags

**Example**:
```xml
<examples>
<example>
<input>Refactor main() to use bootstrap pattern</input>
<output>
```go
func main() {
    if err := run(context.Background(), os.Getenv, os.Stdout, os.Stderr); err != nil {
        slog.Error("fatal", "error", err)
        os.Exit(1)
    }
}
```
</output>
</example>
</examples>
```

### 4. Document Clear Constraints and Edge Cases

Explicit constraints prevent the skill from taking incorrect actions.

**Best practices**:
- Use bullet lists starting with "Include" or "Exclude"
- Cover both positive and negative constraints
- Mention architectural boundaries
- Document 5+ edge case scenarios

**Example**:
```xml
<constraints>
- Include clean, idiomatic Go code
- Exclude magic numbers (use named constants)
- Exclude global mutable state
</constraints>

<edge_cases>
If input is unclear: Ask clarifying questions.
If context is missing: Request additional information.
If performance concerns arise: Delegate to performance skill.
</edge_cases>
```

### 5. Use Concise, Well-Structured Prompts

Over-prompting causes attention dilution—16K tokens with RAG outperformed 128K monolithic prompts.

**Best practices**:
- Every token consumes attention budget
- Challenge each instruction: "Does Claude really need this?"
- Remove until model misbehaves, not add until it behaves
- Focus on communication, not cleverness

### 6. Enable Appropriate Reasoning for Task Complexity

For Claude 4.x with internal reasoning, explicit chain-of-thought provides minimal benefit (2.9-3.1%) while adding 20-80% time cost.

**Best practices**:
- Use structured CoT for complex multi-step tasks where visibility into reasoning matters
- Don't over-engineer simpler skills
- Consider if reasoning visibility is needed

**When to use CoT**:
- Complex multi-step tasks
- Tasks where intermediate steps matter
- When debugging or troubleshooting
- When you need to see the reasoning process

### 7. Optimize Context Window Strategy

Position matters in the context window. Put longform data at top, instructions at end. This improves response quality by up to 30%.

**Best practices**:
- For skills with long content, use a "scratchpad" technique
- Have Claude extract relevant quotes into a thinking section
- Keep system prompts concise to leave room for conversation history

**Formula**: `System_Tokens + History_Tokens + User_Input_Tokens ≤ Model_Window`

## Quick Reference Template

Here's a minimal v2 template you can copy-paste:

```markdown
---
name: your-skill-name
description: "Skill description. Auto-activates for: trigger1, trigger2, trigger3"
version: "2.0.0"
author: "your-name"
tags: ["category", "keyword"]
---

# Skill Title

<role>
Expert [domain] focused on [specialty]. Prioritize [principles].
</role>

<instructions>

## Pattern 1

Code or content example.

**Why this pattern**:
- Reason 1
- Reason 2

## Pattern 2

Another example.

</instructions>

<constraints>
- Include specific patterns
- Exclude anti-patterns
- Bound to principles
</constraints>

<edge_cases>
If input is unclear: Ask clarifying questions.

If context is missing: Request additional information.

If [situation]: [action].

If [situation]: [action].

If [situation]: [action].
</edge_cases>

<examples>
<example>
<input>User request</input>
<output>Expected response</output>
</example>

<example>
<input>Another request</input>
<output>Another response</output>
</example>
</examples>

<output_format>
Provide output following these guidelines:

1. **Requirement 1**: Specific instruction
2. **Requirement 2**: Another instruction

Focus on practical guidance.
</output_format>
```

## Validation and Quality Commands

### Validate Skills

```bash
# Validate all skills (non-strict)
make skill-validate

# Validate all skills (strict mode)
make skill-validate strict=true

# Validate specific skill via MCP
Use skill_validate with skill_id="go-code", strict=true
```

### Quality Report

```bash
# Generate quality report for all skills
make skill-quality

# Generate quality report with custom threshold
Use skill_quality with threshold=90

# Check specific skill
Use skill_quality with skill_id="go-code"
```

### Quality Report Example

```
Skill Quality Report
==================

go-code: Score 95/100 ✓
  Frontmatter: 20/20
  Structure: 30/30
  Content: 30/30
  Triggers: 15/20

go-arch: Score 88/100 ✓
  Frontmatter: 20/20
  Structure: 30/30
  Content: 25/30 (edge_cases missing 1 case)
  Triggers: 13/20

my-new-skill: Score 65/100 ✗
  Frontmatter: 15/20 (version missing)
  Structure: 20/30 (examples missing)
  Content: 15/30 (edge_cases missing)
  Triggers: 15/20

Summary: 2/3 skills meet quality threshold (≥80)
```

## Resources

- **Development Guide**: `docs/DEVELOPMENT.md`
- **Research Guide**: `docs/research/SKILL.md`
- **Example Skills**: `plugins/go-ent/skills/*/SKILL.md`
- **Template Skill**: `plugins/go-ent/skills/go/go-code/SKILL.md`
- **Validation Code**: `internal/skill/validator.go`, `internal/skill/rules.go`
- **Scoring Code**: `internal/skill/scorer.go`

## Troubleshooting

### Validation Fails

**Problem**: Validation errors in strict mode

**Solutions**:
- Check all 9 validation rules above
- Ensure all XML sections are present
- Verify tags are balanced and properly nested
- Check frontmatter has required fields
- Use `make skill-validate` to see specific errors

### Low Quality Score

**Problem**: Quality score < 80

**Solutions**:
- Add missing frontmatter fields (version, author, tags)
- Ensure all XML sections are present
- Add more examples (target 2-3)
- Add more edge cases (target 5+)
- Add more triggers in description (target 3+)
- Check `make skill-quality` for detailed breakdown

### Skill Doesn't Activate

**Problem**: Skill doesn't auto-activate for expected tasks

**Solutions**:
- Check description includes "Auto-activates for:" or "Activates when:"
- Ensure triggers are specific and relevant
- Add more triggers (3+ recommended)
- Verify trigger language matches user queries
- Test skill with specific trigger words

### Examples Don't Help

**Problem**: Examples don't guide output effectively

**Solutions**:
- Use realistic user requests as inputs
- Include complete, runnable code in outputs
- Show before/after comparisons for refactoring
- Cover different use cases and scenarios
- Ensure examples demonstrate key patterns

---

**Version**: 2.0.0
**Last Updated**: 2025-01-18
