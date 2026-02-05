# Template Authoring Guide

This guide explains how to create custom skill templates for go-ent's template system.

## Overview

Templates provide a fast, structured way to create new skills. A template is a pre-built skill definition with placeholders that get replaced with user input during skill generation.

### Template Benefits

- **Speed**: Generate complete skills in seconds
- **Consistency**: Enforce structure and best practices
- **Quality**: Built-in validation and scoring
- **Reusability**: Share templates across teams

## Template Structure

Every template is a directory with two required files:

```
my-template/
├── template.md    # Skill template with placeholders
└── config.yaml    # Template metadata and prompts
```

### File: template.md

The skill template in v2 format with `${PLACEHOLDER}` markers for dynamic content.

### File: config.yaml

Template metadata defining:
- Template identity (name, category, description)
- Author and version
- User prompts for collecting placeholder values

## Complete Example

Here's a complete template example for a Go API skill:

### template.md

```markdown
---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
version: "${VERSION}"
author: "${AUTHOR}"
tags: [${TAGS}]
---

# ${SKILL_NAME}

<role>
Expert Go developer focused on REST API design and implementation.
Prioritize clean architecture, proper error handling, and HTTP best practices.
</role>

<instructions>

## API Endpoint Structure

```go
package api

import (
    "context"
    "net/http"
)

type Handler struct {
    useCase UseCaseInterface
}

func NewHandler(useCase UseCaseInterface) *Handler {
    return &Handler{useCase: useCase}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    var req CreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }
    
    resp, err := h.useCase.Create(ctx, req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(resp)
}
```

## Error Responses

```go
func (h *Handler) handleError(w http.ResponseWriter, err error) {
    if errors.Is(err, ErrNotFound) {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    http.Error(w, "internal error", http.StatusInternalServerError)
}
```

</instructions>

<constraints>
- Use standard library net/http unless framework is specified
- Include context propagation from request
- Wrap errors with context before returning
- Use status codes appropriately (200, 201, 400, 404, 500)
- Validate input before processing
- Keep handlers thin (delegate to use cases)
- Use JSON for request/response bodies
- Include proper Content-Type headers
- Handle panic recovery in middleware
- Log request details for debugging
</constraints>

<edge_cases>
If input is invalid: Return 400 Bad Request with error details.

If resource not found: Return 404 Not Found with clear message.

If unauthorized: Return 401 Unauthorized or 403 Forbidden.

If rate limited: Return 429 Too Many Requests with Retry-After header.

If service unavailable: Return 503 Service Unavailable.

If request context cancelled: Handle gracefully without logging errors.
</edge_cases>

<examples>
<example>
<input>Create an endpoint to create a user</input>
<output>
```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    var req struct {
        Name  string 'json:"name"'
        Email string 'json:"email"'
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }
    
    user, err := h.useCase.CreateUser(ctx, req.Name, req.Email)
    if err != nil {
        h.handleError(w, err)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
```
</output>
</example>

<example>
<input>Create an endpoint to get user by ID</input>
<output>
```go
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    id := mux.Vars(r)["id"]
    user, err := h.useCase.GetUser(ctx, id)
    if err != nil {
        h.handleError(w, err)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```
</output>
</example>
</examples>

<output_format>
Provide production-ready Go API code:

1. **Handlers**: Thin, delegate to use cases
2. **Error Handling**: Appropriate status codes, wrapped errors
3. **Input Validation**: Decode and validate before processing
4. **Context**: Always propagate from request
5. **Response**: JSON with proper Content-Type headers

Focus on HTTP best practices and clean architecture.
</output_format>
```

### config.yaml

```yaml
name: go-api
category: go
description: REST API implementation patterns for Go
author: go-ent
version: 1.0.0
prompts:
  - key: SKILL_NAME
    prompt: Skill name (e.g., go-api)
    default: my-api-skill
    required: true
  - key: DESCRIPTION
    prompt: Brief description of what this skill does
    default: Go REST API patterns
    required: true
  - key: VERSION
    prompt: Skill version
    default: 1.0.0
    required: true
  - key: AUTHOR
    prompt: Author name or organization
    default: go-ent
    required: true
  - key: TAGS
    prompt: Comma-separated tags (e.g., go,api,rest)
    default: go,api,rest
    required: true
```

## Placeholder Syntax

Placeholders use `${PLACEHOLDER}` format and are replaced during skill generation.

### Standard Placeholders

| Placeholder      | Description              | Default Value   | Required |
|-----------------|--------------------------|----------------|----------|
| `${SKILL_NAME}` | Name of the skill       | `my-skill`     | Yes      |
| `${DESCRIPTION}` | Skill description       | (empty)        | Yes      |
| `${VERSION}`     | Skill version           | `1.0.0`        | Yes      |
| `${AUTHOR}`      | Author name             | (empty)        | Yes      |
| `${TAGS}`        | Comma-separated tags    | (empty)        | Yes      |

### Custom Placeholders

Add custom placeholders in `config.yaml` prompts:

```yaml
prompts:
  - key: SKILL_NAME
    prompt: Skill name
    default: my-skill
    required: true
  
  - key: DOMAIN
    prompt: What domain does this skill cover?
    default: general
    required: true
  
  - key: FRAMEWORK
    prompt: Which framework to use?
    default: standard-library
    required: false
```

Use in template.md:

```markdown
<role>
Expert ${DOMAIN} developer focused on ${FRAMEWORK} patterns.
</role>
```

### Placeholder Behavior

- **Required placeholders**: Must be provided by user or have default
- **Missing placeholders**: Kept as-is (not replaced)
- **Empty values**: Replaced with empty string or default
- **Case-sensitive**: `${SKILL_NAME}` != `${skill_name}`

## Config Schema

### Required Fields

| Field      | Type   | Description                         | Example          |
|------------|--------|-------------------------------------|------------------|
| `name`     | string | Template identifier                  | `go-api`         |
| `category` | string | Template category                    | `go`             |
| `description` | string | Brief template description         | `REST API patterns` |

### Optional Fields

| Field      | Type   | Description                         | Example          |
|------------|--------|-------------------------------------|------------------|
| `author`   | string | Template author                      | `go-ent`         |
| `version`  | string | Semantic version                    | `1.0.0`          |
| `prompts`  | array  | User prompts for placeholders       | (see below)      |

### Prompt Schema

Each prompt object in `prompts` array:

| Field     | Type    | Description                              | Example          |
|-----------|---------|------------------------------------------|------------------|
| `key`     | string  | Placeholder name without `${}`             | `SKILL_NAME`     |
| `prompt`  | string  | Text shown to user in interactive mode    | `Skill name`     |
| `default` | string  | Default value if user skips input         | `my-skill`       |
| `required` | boolean | Whether placeholder must have value       | `true`           |

### Complete Config Example

```yaml
name: python-web
category: python
description: Web framework patterns for Python
author: your-name
version: 1.0.0
prompts:
  - key: SKILL_NAME
    prompt: Skill name (e.g., python-django)
    default: python-web
    required: true
  
  - key: DESCRIPTION
    prompt: What does this skill do?
    default: Python web patterns
    required: true
  
  - key: VERSION
    prompt: Skill version
    default: 1.0.0
    required: true
  
  - key: AUTHOR
    prompt: Who created this skill?
    default: your-name
    required: true
  
  - key: TAGS
    prompt: Comma-separated tags (e.g., python,web,django)
    default: python,web
    required: true
  
  - key: FRAMEWORK
    prompt: Which framework? (django, flask, fastapi)
    default: standard-library
    required: false
```

## Template Best Practices

### 1. Start Simple

Begin with a minimal template, then expand:

```markdown
---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
version: "${VERSION}"
author: "${AUTHOR}"
tags: [${TAGS}]
---

# ${SKILL_NAME}

<role>
Expert [domain] focused on [specialty].
</role>

<instructions>
## Pattern 1
Code example.
</instructions>

<constraints>
- Include specific patterns
- Exclude anti-patterns
</constraints>

<edge_cases>
If input unclear: Ask clarifying questions.
</edge_cases>

<examples>
<example>
<input>User request</input>
<output>Response</output>
</example>
</examples>

<output_format>
Provide output following guidelines.
</output_format>
```

### 2. Use Meaningful Placeholders

Good placeholder names are descriptive and consistent:

```yaml
prompts:
  - key: SKILL_NAME        # Good - Clear, consistent
  - key: SKILL_DESCRIPTION # Good - Descriptive
  - key: DOMAIN           # Good - Clear purpose
  - key: FRAMEWORK        # Good - Specific
```

Avoid vague names:
```yaml
prompts:
  - key: value1   # Bad - Unclear
  - key: param    # Bad - Too generic
  - key: thing    # Bad - No context
```

### 3. Provide Sensible Defaults

Defaults help users get started faster:

```yaml
prompts:
  - key: VERSION
    prompt: Skill version
    default: 1.0.0    # Good default
  
  - key: AUTHOR
    prompt: Author name
    default: go-ent    # Good default for project
  
  - key: FRAMEWORK
    prompt: Which framework?
    default: stdlib    # Clear, explicit
```

### 4. Structure Prompts Logically

Group related prompts together:

```yaml
prompts:
  # Core metadata
  - key: SKILL_NAME
    prompt: Skill name
    default: my-skill
    required: true
  
  - key: DESCRIPTION
    prompt: Description
    required: true
  
  # Versioning
  - key: VERSION
    prompt: Version
    default: 1.0.0
    required: true
  
  - key: AUTHOR
    prompt: Author
    required: true
  
  # Categorization
  - key: TAGS
    prompt: Tags
    required: true
  
  # Domain-specific
  - key: LANGUAGE
    prompt: Programming language
    default: go
    required: false
```

### 5. Validate Template Before Publishing

Always test your template:

```bash
# 1. Test template loading
go-ent skill show-template my-template

# 2. Generate a skill from template
go-ent skill new test-skill --template my-template

# 3. Validate generated skill
go-ent skill validate test-skill

# 4. Check quality score
go-ent skill quality test-skill
```

### 6. Document Edge Cases

Include comprehensive edge case coverage:

```markdown
<edge_cases>
If input is unclear: Ask clarifying questions.

If context is missing: Request additional information.

If performance concerns arise: Suggest optimizations.

If security issues detected: Highlight and recommend fixes.

If architecture questions emerge: Delegate to architecture skill.

If testing needed: Delegate to testing skill.

If error handling complex: Provide detailed patterns.

If API design needed: Follow REST principles.
</edge_cases>
```

### 7. Provide Real Examples

Use realistic examples that match your skill's purpose:

```markdown
<examples>
<example>
<input>Create a simple REST endpoint for user creation</input>
<output>
```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }
    
    user, err := h.useCase.CreateUser(ctx, req.Name, req.Email)
    if err != nil {
        h.handleError(w, err)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
```
</output>
</example>
</examples>
```

### 8. Follow v2 Skill Format

Ensure template.md is valid v2 format:

```markdown
---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
version: "${VERSION}"
author: "${AUTHOR}"
tags: [${TAGS}]
---

<role>
Expert persona definition.
</role>

<instructions>
Patterns and examples.
</instructions>

<constraints>
- Include requirements
- Exclude anti-patterns
</constraints>

<edge_cases>
If scenario: Handling approach.
</edge_cases>

<examples>
<example>
<input>Request</input>
<output>Response</output>
</example>
</examples>

<output_format>
Output guidelines.
</output_format>
```

## Validation Rules

Templates must pass these validations when added with `go-ent skill add-template`:

### 1. Structure Validation

- Template directory must exist
- `template.md` must be present
- `config.yaml` must be present

### 2. Config Validation

- `config.yaml` must be valid YAML
- Required fields present: `name`, `category`, `description`
- `name` must not be empty
- `category` must not be empty
- `prompts` array must be valid (if present)
- Each prompt must have `key`, `prompt`, `required`

### 3. Skill Validation

`template.md` must pass v2 skill validation:

- Frontmatter has required fields
- All XML tags are balanced
- Required sections present: `<role>`, `<instructions>`, `<constraints>`, `<edge_cases>`, `<examples>`, `<output_format>`
- Examples have `<input>` and `<output>` tags

### Validation Example

```bash
$ go-ent skill add-template /path/to/my-template

Template structure valid
Config file valid
Skill validation failed: Missing <role> section

Fix the template.md file before adding.
```

## Quality Scoring

Templates generate skills that are scored on 0-100 scale:

### Scoring Components

| Component      | Points | Criteria                              |
|----------------|--------|---------------------------------------|
| Frontmatter    | 20     | name, description, version, tags      |
| Structure      | 30     | role, instructions, examples          |
| Content        | 30     | examples count, edge_cases            |
| Triggers       | 20     | number of triggers in description     |

### Score Thresholds

| Score  | Quality   | Action                         |
|--------|-----------|--------------------------------|
| >= 90   | Excellent | Template quality, ready for use |
| 80-89  | Good      | Acceptable for production       |
| < 80   | Needs work | Add sections, examples         |

### Improving Quality

To increase score:

1. **Add frontmatter fields**: Include `version`, `author`, `tags`
2. **Add XML sections**: Ensure all required sections present
3. **Add examples**: Target 2-3 examples with input/output
4. **Add edge cases**: Target 5+ scenarios
5. **Add triggers**: Include 3+ triggers in description

Example:
```yaml
---
name: ${SKILL_NAME}
description: "${DESCRIPTION}. Auto-activates for: trigger1, trigger2, trigger3"
version: "${VERSION}"
author: "${AUTHOR}"
tags: [${TAGS}]
---
```

## Adding Custom Templates

### Step 1: Create Template Directory

```bash
mkdir -p my-custom-template
cd my-custom-template
```

### Step 2: Create config.yaml

```yaml
name: my-custom
category: custom
description: Custom skill template for specific domain
author: your-name
version: 1.0.0
prompts:
  - key: SKILL_NAME
    prompt: Skill name
    default: my-skill
    required: true
  - key: DESCRIPTION
    prompt: Description
    default: Custom skill
    required: true
  - key: VERSION
    prompt: Version
    default: 1.0.0
    required: true
  - key: AUTHOR
    prompt: Author
    default: go-ent
    required: true
  - key: TAGS
    prompt: Tags
    default: custom
    required: true
```

### Step 3: Create template.md

```markdown
---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
version: "${VERSION}"
author: "${AUTHOR}"
tags: [${TAGS}]
---

# ${SKILL_NAME}

<role>
Expert in your domain focused on best practices.
</role>

<instructions>
## Pattern 1

Code or content example.

**Why this pattern**:
- Reason 1
- Reason 2
</instructions>

<constraints>
- Include specific patterns
- Exclude anti-patterns
</constraints>

<edge_cases>
If input unclear: Ask clarifying questions.
</edge_cases>

<examples>
<example>
<input>User request</input>
<output>Response</output>
</example>
</examples>

<output_format>
Provide output following guidelines.
</output_format>
```

### Step 4: Validate Template

```bash
# Check template structure
go-ent skill show-template my-custom

# Add to user templates
go-ent skill add-template /path/to/my-custom-template
```

### Step 5: Test Template

```bash
# Generate a skill
go-ent skill new test-skill --template my-custom

# Validate generated skill
go-ent skill validate test-skill

# Check quality
go-ent skill quality test-skill
```

## Template Locations

### Built-in Templates

Location: `plugins/go-ent/templates/skills/`

These are distributed with go-ent and updated with releases.

### User Templates

Location: `~/.go-ent/templates/skills/`

Custom templates added by users for personal use.

### Environment Variables

Override default locations:

```bash
# Custom templates directory
export GO_ENT_TEMPLATE_DIR=/path/to/custom/templates

# Custom skills output directory
export GO_ENT_SKILLS_DIR=/path/to/custom/skills
```

## CLI Commands

### list-templates

List all available templates:

```bash
# List all templates
go-ent skill list-templates

# Filter by category
go-ent skill list-templates --category go

# Show only built-in
go-ent skill list-templates --built-in

# Show only custom
go-ent skill list-templates --custom
```

### show-template

View template details:

```bash
# Show built-in template
go-ent skill show-template go-basic

# Show custom template
go-ent skill show-template my-custom
```

Output includes:
- Template metadata
- Configuration prompts
- Template preview (first 20 lines)

### add-template

Add custom template to registry:

```bash
# Add to user templates (default)
go-ent skill add-template /path/to/template

# Add to built-in directory
go-ent skill add-template /path/to/template \
  --built-in /path/to/go-ent/plugins/go-ent/templates/skills/
```

Validates template before adding.

### new

Generate skill from template:

```bash
# Interactive mode
go-ent skill new go-payment

# Non-interactive mode
go-ent skill new go-payment \
  --template go-complete \
  --description "Payment processing" \
  --category go \
  --author "Your Name" \
  --tags "payment,api"
```

## Common Issues

### Template Not Found

**Problem**: Template doesn't appear in list

**Solution**:
- Check template directory exists
- Verify `config.yaml` and `template.md` are present
- Run `go-ent skill show-template <name>` to debug

### Validation Fails

**Problem**: `go-ent skill add-template` fails validation

**Solution**:
- Check `config.yaml` is valid YAML
- Verify required fields present
- Validate `template.md` passes v2 skill validation
- Check for missing XML tags or sections

### Placeholders Not Replaced

**Problem**: Generated skill still has `${PLACEHOLDER}` text

**Solution**:
- Ensure placeholder key matches config prompt key
- Check placeholder is defined in config.yaml prompts
- Verify prompt `required` field is set correctly

### Low Quality Score

**Problem**: Generated skill scores < 80

**Solution**:
- Add missing frontmatter fields
- Ensure all XML sections present
- Add more examples (target 2-3)
- Add more edge cases (target 5+)
- Add more triggers in description (target 3+)

## Resources

- **Skill Authoring Guide**: `docs/SKILL-AUTHORING.md`
- **Development Guide**: `docs/DEVELOPMENT.md`
- **Built-in Templates**: `plugins/go-ent/templates/skills/*/`
- **Example Skills**: `plugins/go-ent/skills/*/SKILL.md`

## Support

For help with template creation:

1. Check existing built-in templates for examples
2. Refer to `docs/SKILL-AUTHORING.md` for v2 format details
3. Use `go-ent skill show-template` to inspect templates
4. Test with `go-ent skill new` before publishing
5. Validate with `go-ent skill validate` and quality scoring

---
**Version**: 1.0.0  
**Last Updated**: 2025-01-18
