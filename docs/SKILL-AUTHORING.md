# Skill Authoring Guide (v4 Format)

Complete guide for creating high-quality skills using the v4 format (Markdown sections + flat YAML triggers) for go-ent's plugin system.

---

## Table of Contents

- [Overview](#overview)
- [Complete Skill Template](#complete-skill-template)
- [Frontmatter Fields](#frontmatter-fields)
- [Markdown Sections](#markdown-sections)
- [Validation](#validation)
- [Good vs Bad Examples](#good-vs-bad-examples)
- [Best Practices](#best-practices)
- [Quick Reference](#quick-reference)

---

## Overview

The v4 skill format is the **only supported format** in go-ent and uses:

- **YAML frontmatter** - Minimal metadata with flat trigger array
- **Markdown sections** - `## Role`, `## Instructions`, `## Examples`
- **Flat triggers** - Simple string array in frontmatter
- **No XML tags** - Clean Markdown only
- **No legacy formats** - v1, v2, v3 are not supported

### v4 Format Characteristics

| Feature | v4 Format |
|---------|-----------|
| Metadata | Minimal YAML frontmatter |
| Triggers | Flat array: `triggers: [str1, str2]` |
| Role | `## Role` Markdown heading |
| Instructions | `## Instructions` Markdown heading |
| Examples | `## Examples` Markdown heading |
| References | `## References` Markdown heading (optional) |
| Validation | Format detection, section presence |

---

## Complete Skill Template

Here's a complete template for a v4 skill:

```markdown
---
name: your-skill-name
description: Brief skill description for trigger matching decisions
triggers:
  - keyword1
  - keyword2
  - pattern phrase
---

## Role

One-line expert persona definition with domain expertise.

## Instructions

Detailed, actionable instructions for skill execution.

### Pattern 1

Code or content example with clear explanation.

**Why this pattern**:
- Reason 1
- Reason 2

### Pattern 2

Another example with specific rules.

**Rules**:
- Rule 1
- Rule 2

## Examples

<example>
<input>User request example</input>
<output>Expected response or code</output>
</example>

<example>
<input>Another user request</input>
<output>Another expected output</output>
</example>

## References

- [Reference name](references/file.md)
- [Another reference](references/another.md)
```

---

## Frontmatter Fields

### Required Fields

| Field | Description | Example |
|-------|-------------|---------|
| `name` | Skill identifier (lowercase, hyphens) | `go-code` |
| `description` | Brief summary for invocation decisions | `Go coding patterns and best practices` |
| `triggers` | Flat array of trigger strings | `["go code", "golang", "go patterns"]` |

### Optional Fields

None. v4 format is minimal - only name, description, and triggers.

### Triggers (v4 Format)

Triggers are a **flat array of strings** in frontmatter:

```yaml
triggers:
  - error handling
  - error wrapping
  - fmt.Errorf
```

**Best practices**:
- 3+ keywords recommended
- Use specific, relevant phrases
- Match user terminology
- Include variations of key terms

**Examples**:
```yaml
# Good triggers
triggers:
  - go code
  - golang patterns
  - idiomatic go
  - go best practices

# Also valid (single line)
triggers: ["go code", "golang", "idiomatic go"]
```

---

## Markdown Sections

### `## Role` - Expert Persona (Required)

Define the AI's expertise in one clear sentence:

```markdown
## Role

Expert Go developer specializing in clean architecture, idiomatic patterns, and production-grade code quality.
```

**Purpose**: Sets the persona and expertise level
**Content**: One-line expert identity
**Best practices**:
- Single sentence
- Include domain expertise
- Mention key principles or focus
- Keep concise and focused

---

### `## Instructions` - Core Knowledge (Required)

Provide detailed, actionable guidance with code examples:

```markdown
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

### Repository Pattern

Use constructor injection for dependencies:

```go
type Repository struct {
    db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
    return &Repository{db: db}
}
```

**Rules**:
- Inject dependencies via constructor
- Use pointer receivers
- Keep fields private
```

**Purpose**: Core knowledge and patterns
**Content**: Code examples, explanations, rules, patterns
**Format**: Markdown with code blocks, lists, subsections
**Best practices**:
- Use code blocks with language tags
- Include "Why this pattern" sections
- Group related patterns with subsections (###)
- Be specific and actionable

---

### `## Examples` - Usage Examples (Required)

Provide 2+ concrete examples using XML-style example tags:

```markdown
## Examples

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

func run(ctx context.Context, getenv func(string) string, stdout, stderr io.Writer) error {
    // Application logic here
    return nil
}
```
</output>
</example>

<example>
<input>Fix error handling in repository function</input>
<output>
```go
// Before
func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
    return nil, err
}

// After
func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
    var user User
    if err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Name); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    return &user, nil
}
```
</output>
</example>
```

**Purpose**: Demonstrate skill application with realistic scenarios
**Content**: 2+ examples with clear input/output
**Format**: `<example><input>...</input><output>...</output></example>`
**Best practices**:
- Use realistic user requests
- Include complete, runnable code
- Show before/after comparisons when helpful
- Cover different use cases

---

### `## References` - Reference Links (Optional)

Link to reference documentation files:

```markdown
## References

- [Go conventions](references/conventions.md)
- [Error patterns](references/errors.md)
- [Best practices](references/best-practices.md)
```

**Purpose**: Link to supporting documentation
**Content**: Markdown links to reference files
**Format**: Unordered list of links
**Best practices**:
- Use relative paths
- Link to actual reference files in skill directory
- Keep link text descriptive

---

## Validation

Skills are validated for v4 format compliance.

### Validation Checks

1. **Format detection** - Verifies v4 format (flat triggers + Markdown sections)
2. **Required frontmatter** - `name`, `description`, `triggers` must be present
3. **Required sections** - `## Role`, `## Instructions`, `## Examples` must exist
4. **Trigger format** - Triggers must be array of strings

### Running Validation

```bash
# Validate all skills
make build && ./bin/ent

# Parser will reject non-v4 skills automatically
# Error: "unsupported skill format: {version}, only v4 is supported"
```

### Validation Errors

| Error | Cause | Fix |
|-------|-------|-----|
| "missing name" | No `name:` in frontmatter | Add `name: skill-id` |
| "missing description" | No `description:` in frontmatter | Add `description: "..."` |
| "missing triggers" | No `triggers:` in frontmatter | Add `triggers: [...]` |
| "unsupported skill format" | Not v4 format | Ensure flat triggers and `## Role/Instructions/Examples` |

---

## Good vs Bad Examples

### Good Example (v4 Format)

```markdown
---
name: go-error
description: Error handling patterns and wrapping strategies for Go
triggers:
  - error handling
  - error wrapping
  - fmt.Errorf
  - error context
---

## Role

Expert Go error handling engineer specializing in error design patterns, wrapping strategies, and production-grade error management.

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

## Examples

<example>
<input>Add error handling to database query</input>
<output>
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
</output>
</example>

<example>
<input>Fix error propagation through layers</input>
<output>
```go
// Repository layer
func (r *repo) FindByID(ctx context.Context, id string) (*User, error) {
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
```
</output>
</example>
```

**Why this is good**:
- Complete frontmatter with flat triggers
- Clear one-line role definition
- Rich instructions with code examples and "why" explanations
- 2 realistic examples with clear input/output
- Uses Markdown sections (no XML)
- Proper v4 format

---

### Bad Example (Needs Improvement)

```markdown
---
name: my-skill
description: Does stuff
triggers:
  - stuff
---

This skill helps with things.

Write good code.

Example: Do this
```

**Why this is bad**:
- Vague description ("Does stuff")
- Single generic trigger
- No Markdown sections (missing ## Role, ## Instructions, ## Examples)
- Instructions are too generic
- Examples lack structure (no `<example>` tags)
- Not v4 format compliant

**How to fix**:
1. Add specific description
2. Add 3+ specific trigger keywords
3. Add `## Role` section with expertise definition
4. Add `## Instructions` with specific patterns and code examples
5. Add `## Examples` section with `<example>` tags
6. Use code blocks for examples
7. Be specific and actionable

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
- 2-3 diverse examples minimum
- Include edge cases
- Show complete, runnable code
- Use realistic scenarios
- Use `<example><input>...</input><output>...</output></example>` format

### 3. Use 3+ Trigger Keywords

More triggers = better activation matching.

**Example**:
```yaml
triggers:
  - error handling
  - error wrapping
  - fmt.Errorf
  - error context
```

### 4. Keep Instructions Focused

Every token consumes attention budget.

**Best practices**:
- Challenge each instruction: "Is this needed?"
- Focus on critical patterns only
- Use code examples over prose
- Be concise but complete

### 5. One-Line Role Definition

The role should be a single, clear sentence.

**Good**:
```markdown
## Role

Expert Go developer specializing in clean architecture, idiomatic patterns, and production-grade code quality.
```

**Bad**:
```markdown
## Role

You are an expert. You know Go. You write code. You follow best practices. You care about quality.
```

---

## Quick Reference

### Minimal v4 Template

```markdown
---
name: skill-name
description: Brief description for trigger matching
triggers:
  - trigger1
  - trigger2
  - trigger3
---

## Role

Expert [domain] specializing in [specialty].

## Instructions

### Pattern 1

Code or content example.

**Why this pattern**:
- Reason 1
- Reason 2

## Examples

<example>
<input>User request</input>
<output>
```lang
// Code example
```
</output>
</example>

<example>
<input>Another request</input>
<output>
```lang
// Another example
```
</output>
</example>
```

### File Naming Convention

All skills must be in a file named `SKILL.md`:

```
pkg/skills/
├── core/
│   └── api-design/
│       └── SKILL.md
├── go/
│   └── go-code/
│       └── SKILL.md
└── ent/
    └── ent-openspec/
        └── SKILL.md
```

---

## Resources

- **Example Skills**: `pkg/skills/*/SKILL.md`
- **Reference Skills**: `pkg/skills/ent/*/SKILL.md`
- **Parser Code**: `internal/skill/parser.go`

---

## Troubleshooting

### Skill Doesn't Load

**Problem**: "unsupported skill format" error

**Solutions**:
- Ensure frontmatter has `name`, `description`, `triggers`
- Ensure `triggers` is a flat array of strings
- Ensure `## Role`, `## Instructions`, `## Examples` sections exist
- Remove any XML tags (v2 format)
- Remove object-style triggers (v3 format)

### Skill Doesn't Activate

**Problem**: Skill doesn't auto-activate

**Solutions**:
- Check `triggers` array in frontmatter
- Ensure keywords match user terminology
- Add more keywords (3+ recommended)
- Test with specific trigger words

### Examples Don't Help

**Problem**: Examples don't guide output effectively

**Solutions**:
- Use realistic user requests
- Include complete, runnable code
- Show before/after for refactoring
- Cover different use cases
- Use proper `<example><input>...</input><output>...</output></example>` format

---

**Version**: 4.0.0
**Last Updated**: 2026-01-31
