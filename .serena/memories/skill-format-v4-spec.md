# Skill Format v4 Specification

## Overview

Minimal skill format for go-ent that prioritizes simplicity, maintainability, and industrial best practices.

## File Structure

```
skills/{category}/{name}/
├── SKILL.md              # Required: Core skill definition
├── references/           # Optional: Extended content
│   ├── *.md             # Reference documents
│   └── patterns/        # Pattern library
└── scripts/             # Optional: Helper scripts
    └── *.sh, *.py, etc
```

## SKILL.md Format

### Frontmatter (YAML)

```yaml
---
name: string              # Required: lowercase, hyphens, max 64 chars
description: string       # Required: max 256 chars, explains WHAT and WHEN
triggers:                 # Required: array of trigger phrases
  - string               # Each trigger: lowercase, 2-4 words recommended
---
```

**Field Details:**

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | Yes | `^[a-z][a-z0-9-]{0,63}$` |
| `description` | string | Yes | Max 256 chars, no "Auto-activates for:" prefix |
| `triggers` | []string | Yes | 1-10 items, each 1-5 words |

### Body (Markdown)

Required sections marked with ✓

```markdown
## Role                    ✓

Expert persona definition (1-3 sentences).

## Instructions            ✓

### Pattern Name

Content with code examples.

**Why this pattern**:
- Reason 1
- Reason 2

### Another Pattern

More content...

## Examples                ✓

### Example 1: Description

**Input**: User request

**Output**:
```language
// Code or response
```
```

**Section Requirements:**

| Section | Required | Min Length | Max Length | Content |
|---------|----------|------------|------------|---------|
| Role | ✓ | 50 chars | 500 chars | Persona definition |
| Instructions | ✓ | 200 chars | 10KB | Patterns, rules, guidance |
| Examples | ✓ | 1 example | 5 examples | Input/output pairs |

### Optional: References Section

If skill has extended content in `references/`:

```markdown
## References

- [Constraints](references/constraints.md)
- [Edge Cases](references/edge-cases.md)
- [Advanced Patterns](references/patterns/)
```

## References/ Directory

### Structure Rules

```
references/
├── *.md                    # Flat reference files
├── patterns/              # Pattern library (flat)
│   └── *.md
└── scripts/               # Helper scripts
    └── *
```

### Validation Rules

1. **Max Depth**: 1 subdirectory level (`references/patterns/` allowed, `references/a/b/` not allowed)
2. **Markdown Files**: Must have single `# Header` as first line
3. **No Frontmatter**: Reference files are pure content
4. **Max Size**: 50KB per file
5. **Naming**: lowercase, hyphens, `.md` extension

### Reference File Template

```markdown
# Reference Title

Content here. No frontmatter, no YAML.

## Subsections Allowed

Use ## for subsections within reference.
```

## Matching Algorithm

### Trigger Matching

```go
func matchSkill(skill SkillMeta, query string, fileExt string) (bool, float64) {
    queryLower := strings.ToLower(query)
    matches := 0
    
    // Keyword matching
    for _, trigger := range skill.Triggers {
        if strings.Contains(queryLower, trigger) {
            matches++
        }
    }
    
    // Score = match ratio
    score := float64(matches) / float64(len(skill.Triggers))
    
    // Category bonus (file extension matches skill category)
    if matches > 0 && categoryMatches(fileExt, skill.Category) {
        score += 0.1
    }
    
    // Return: matched (any trigger hit), score (0.0-1.1)
    return matches > 0, math.Min(score, 1.0)
}
```

### Category Inference

Category derived from path: `skills/{category}/{name}/`

| Path Pattern | Category | File Extensions |
|--------------|----------|-----------------|
| `skills/go/*` | go | .go, .mod, .sum |
| `skills/core/*` | core | .md, .txt, .yaml |
| `skills/ent/*` | ent | .md, .yaml |

## Validation Rules

### Essential Rules (5 total)

1. **validateFrontmatter**
   - Required fields present: name, description, triggers
   - No unknown fields
   - YAML syntax valid

2. **validateNameFormat**
   - Matches `^[a-z][a-z0-9-]{0,63}$`
   - Not empty, not too long

3. **validateRoleSection**
   - `## Role` heading exists
   - Content length 50-500 chars
   - Not empty after heading

4. **validateInstructionsSection**
   - `## Instructions` heading exists
   - Content length 200-10KB
   - Has at least one subsection (###) or code block

5. **validateExamplesSection**
   - `## Examples` heading exists
   - At least one example (### Example N:)
   - Each example has **Input** and **Output**

### Optional Rules (Warnings Only)

6. **validateReferences**
   - If `references/` exists, structure is valid
   - No deep nesting
   - No frontmatter in reference files

7. **validateTriggerQuality**
   - 3+ triggers recommended
   - Triggers are specific (not generic words like "code", "write")

## Migration from v3

### Automated Conversion

```bash
# Convert single skill
ent skill migrate v3-to-v4 --input go-error/SKILL.md --output go-error-v4/

# Convert all skills
ent skill migrate v3-to-v4 --all --backup
```

### Conversion Rules

| v3 Element | v4 Conversion |
|------------|---------------|
| `triggers.keywords` | Flatten to `triggers` array |
| `triggers.file_pattern` | Remove (category inference) |
| `triggers.weight` | Remove (computed at runtime) |
| `version`, `author`, `license` | Remove |
| `category` | Infer from path |
| `quality_score` | Remove |
| `## Constraints` | Move to `references/constraints.md` if >5 items |
| `## Edge Cases` | Move to `references/edge-cases.md` if >5 items |
| `## Output Format` | Merge into Instructions as subsection |
| `<role>` XML tags | Convert to `## Role` |
| `<instructions>` XML | Convert to `## Instructions` |

## Examples

### Minimal Valid Skill

```markdown
---
name: go-hello
description: Basic Go patterns for beginners
triggers:
  - go basics
  - golang intro
---

## Role

Go programming assistant focused on beginner-friendly explanations.

## Instructions

### Hello World

Start with a simple main package:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

## Examples

### Example 1: First Program

**Input**: Create a hello world program

**Output**:
```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```
```

### Full Skill with References

```markdown
---
name: go-error
description: Error handling patterns for Go
triggers:
  - error handling
  - error wrapping
  - fmt.Errorf
  - go errors
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

### Lowercase Messages

Error messages should be lowercase without trailing punctuation.

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

## References

- [Complete Constraints](references/constraints.md)
- [Edge Cases](references/edge-cases.md)
```

## Version History

| Version | Date | Changes |
|---------|------|---------|
| v4.0.0 | 2026-01-30 | Initial simplified format |
| v3.0.0 | 2026-01-15 | Markdown sections + YAML frontmatter |
| v2.0.0 | 2025-12-01 | XML tags in Markdown |
| v1.0.0 | 2025-11-01 | Basic frontmatter only |
