# Plugin Development Guide

This guide walks through creating custom plugins for go-ent, including skills, agents, and packaging for the marketplace.

## Overview

Plugins are self-contained packages that extend go-ent with:
- **Skills**: Knowledge bases that auto-activate for specific tasks
- **Agents**: Specialized AI assistants with defined capabilities
- **Rules**: Event-driven automation

## Plugin Directory Structure

A typical plugin follows this structure:

```
my-plugin/
├── plugin.yaml          # Plugin manifest (required)
├── skills/              # Optional: Skill definitions
│   └── skill-name/
│       └── SKILL.md
├── agents/              # Optional: Agent definitions
│   └── agent-name.md
└── rules/               # Optional: Rule definitions
    └── rule-name.yaml
```

## Quick Start: Create a Simple Skill Plugin

### Step 1: Create Plugin Directory

```bash
mkdir -p my-plugin/skills/my-skill
cd my-plugin
```

### Step 2: Create the Manifest

Create `plugin.yaml`:

```yaml
name: my-skill-plugin
version: 1.0.0
description: A sample skill plugin demonstrating the basics
author: Your Name <you@example.com>

skills:
  - name: my-skill
    path: skills/my-skill/SKILL.md
```

### Step 3: Create the Skill Definition

Create `skills/my-skill/SKILL.md`:

```markdown
---
name: my-skill
description: "My custom skill. Auto-activates for: custom logic, my domain."
---

# My Custom Skill

This skill provides knowledge about my specific domain.

## When This Activates

This skill auto-activates when you:
- Work with my domain terminology
- Implement custom logic patterns
- Need specific knowledge about my domain

## Pattern 1: My Custom Pattern

```go
// Example code pattern
func CustomFunction(input string) (string, error) {
    if input == "" {
        return "", fmt.Errorf("input cannot be empty")
    }
    return strings.ToUpper(input), nil
}
```

**Why this pattern**:
- Handles empty input gracefully
- Returns consistent error format
- Transforms input predictably

## Pattern 2: My Architecture

```go
type MyService struct {
    config Config
    client Client
}

func NewService(cfg Config, client Client) *MyService {
    return &MyService{
        config: cfg,
        client: client,
    }
}

func (s *MyService) Process(ctx context.Context, data Data) error {
    // Implementation
    return nil
}
```

**Key points**:
- Dependency injection for testability
- Context propagation for cancellation
- Clear interface boundaries
```

### Step 4: Test Locally

Place your plugin in go-ent's plugins directory:

```bash
cp -r my-plugin ~/Projects/own/go-ent/plugins/
```

Restart go-ent (or rebuild MCP server if needed):

```bash
cd ~/Projects/own/go-ent
make build-mcp
# Restart Claude Code
```

Verify the skill loads by typing something that should trigger it (e.g., "Implement custom logic for my domain").

## Creating Skills

Skills provide specialized knowledge that auto-activates based on context. Use the SKILL.md format:

```markdown
---
name: skill-name
description: "Skill description. Auto-activates for: trigger1, trigger2, trigger3."
---

# Skill Name

Comprehensive documentation for this skill...

## When This Activates

Describe when this skill should activate:
- Trigger condition 1
- Trigger condition 2

## Pattern 1: Code Pattern

```go
// Example code
```

**Why this pattern**: Explanation of rationale

## Pattern 2: Architecture Decision

```markdown
# Design Pattern Name

## Overview
Description

## Components
List of components

## Data Flow
Step-by-step flow
```

## Best Practices

1. **Be specific in description**: Include clear activation triggers
2. **Provide working examples**: Copy-pasteable code snippets
3. **Explain rationale**: Help users understand why patterns work
4. **Use consistent formatting**: Markdown with code blocks
5. **Keep it focused**: One skill = one domain of knowledge
```

## Creating Agents

Agents are specialized AI assistants with defined tools and capabilities.

### Agent Definition Format

Create `agents/agent-name.md`:

```markdown
---
name: my-agent
description: "Brief description of this agent's purpose"
model: claude-3-sonnet
color: "#3B82F6"
skills:
  - my-skill
  - go-code
tools:
  file-read: true
  file-write: true
  glob: true
  grep: true
---

# My Agent

You are a specialized agent for [specific domain].

## Responsibilities

- Responsibility 1
- Responsibility 2
- Responsibility 3

## Workflow

1. First, analyze the request
2. Check dependencies and constraints
3. Implement solution following best practices
4. Validate and test

## Tools Available

- `file-read`: Read file contents
- `file-write`: Write or modify files
- `glob`: Find files by pattern
- `grep`: Search file contents

## Skills Available

- `my-skill`: Domain-specific knowledge
- `go-code`: Go implementation patterns

## Examples

### Example 1

**Request**: Implement feature X

**Approach**:
1. Parse requirements
2. Design architecture
3. Implement following patterns from my-skill
4. Write tests
5. Verify

### Example 2

**Request**: Fix bug Y

**Approach**:
1. Reproduce issue
2. Identify root cause
3. Apply fix
4. Add regression test
5. Verify fix

## Capabilities

- Can use my-skill for domain knowledge
- Has access to file operations
- Follows go-code patterns for implementation
- Understands project structure
```

### Agent Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique agent identifier (e.g., `my-agent`) |
| `description` | string | Yes | Brief description of agent's purpose |
| `model` | string | Yes | Claude model to use (`claude-3-opus`, `claude-3-sonnet`, `claude-3-haiku`) |
| `color` | string | No | Hex color code for UI (e.g., `#3B82F6`) |
| `skills` | array | No | List of skill names this agent uses |
| `tools` | object | No | Map of tool names to boolean (if true, tool is available) |

## Creating Rules

Rules define event-driven automation.

### Rule Definition Format

Create `rules/rule-name.yaml`:

```yaml
name: my-rule
description: "Description of what this rule does"
priority: high

trigger:
  event: file_modified
  pattern: "*.go"
  path: "internal/domain/"

conditions:
  - type: content_matches
    pattern: "TODO|FIXME|XXX"

actions:
  - type: reject
    message: "TODOs not allowed in domain layer"

  - type: comment
    message: "Consider replacing TODO with actual implementation"

  - type: modify
    target: file
    operation: add_line
    position: end
    content: "// Code reviewed: no TODOs found"
```

### Rule Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique rule identifier |
| `description` | string | Yes | What this rule does |
| `priority` | string | No | Rule priority (`low`, `medium`, `high`) |
| `trigger` | object | Yes | Event trigger configuration |
| `conditions` | array | No | Conditions to evaluate |
| `actions` | array | Yes | Actions to execute when conditions match |

### Trigger Types

- `file_modified`: Triggered when a file is modified
- `file_created`: Triggered when a file is created
- `file_deleted`: Triggered when a file is deleted

### Condition Types

- `content_matches`: File content matches regex pattern
- `path_matches`: File path matches pattern
- `size_exceeds`: File size exceeds limit

### Action Types

- `reject`: Reject the operation
- `comment`: Add a comment
- `modify`: Modify the file
- `notify`: Send notification

## Testing Plugins Locally

### Method 1: Direct Copy

```bash
# Copy your plugin to plugins directory
cp -r my-plugin ~/Projects/own/go-ent/plugins/

# Restart go-ent
cd ~/Projects/own/go-ent
make build-mcp
# Restart Claude Code
```

### Method 2: Symlink (Development)

```bash
# Symlink for faster iteration
ln -s /path/to/my-plugin ~/Projects/own/go-ent/plugins/my-plugin

# Changes reflect immediately (no rebuild needed)
```

### Verification

1. Check plugin loads: List installed plugins in go-ent
2. Test skill activation: Type something that should trigger your skill
3. Test agent: Invoke your agent and verify it responds correctly
4. Check logs: Look for any loading errors

## Packaging Plugins for Marketplace

### Step 1: Prepare Plugin Archive

```bash
# Create a zip archive of your plugin
cd my-plugin
zip -r ../my-plugin-1.0.0.zip .
cd ..
```

**Important**: The archive should contain plugin files at the root, not a parent directory.

### Step 2: Validate Manifest

```bash
# Use go-ent to validate the manifest
go-ent plugin validate my-plugin/plugin.yaml
```

### Step 3: Test Install

```bash
# Test installing from local archive
go-ent plugin install local my-plugin-1.0.0.zip
```

### Step 4: Submit to Marketplace

The marketplace submission process includes:

1. **Register Plugin** (via marketplace web UI or API):
   - Plugin name (unique)
   - Description
   - Category
   - Tags

2. **Upload Archive**:
   - Upload the zip file
   - System validates manifest automatically

3. **Review Process**:
   - Automatic validation
   - Manual review (if needed)
   - Approval or rejection with feedback

4. **Publish**:
   - Plugin becomes available to users
   - Version management through marketplace

### Versioning

Follow semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Breaking changes (incompatible API changes)
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

Update `version` in `plugin.yaml` for each release.

### Changelog

Include a `CHANGELOG.md` in your plugin:

```markdown
# Changelog

## [1.0.0] - 2026-01-17

### Added
- Initial release
- Skill for custom patterns
- Agent for domain-specific tasks

### Changed
- N/A

### Fixed
- N/A

### Deprecated
- N/A

### Removed
- N/A
```

## Complete Example: Data Validation Plugin

Here's a complete example plugin with skills, agents, and rules.

### Directory Structure

```
data-validation-plugin/
├── plugin.yaml
├── skills/
│   └── validation/
│       └── SKILL.md
├── agents/
│   └── validator.md
└── rules/
    └── enforce-validation.yaml
```

### plugin.yaml

```yaml
name: data-validation-plugin
version: 1.0.0
description: Comprehensive data validation patterns and tools
author: Data Team <data@example.com>

skills:
  - name: validation
    path: skills/validation/SKILL.md

agents:
  - name: validator
    path: agents/validator.md

rules:
  - name: enforce-validation
    path: rules/enforce-validation.yaml
```

### skills/validation/SKILL.md

```markdown
---
name: validation
description: "Data validation patterns. Auto-activates for: validation, input validation, data integrity, schema validation."
---

# Data Validation Patterns

## When This Activates

This skill activates when you:
- Implement input validation
- Validate data integrity
- Define schema constraints
- Create validation rules

## Pattern 1: Struct Validation

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=3,max=100"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"gte=0,lte=150"`
}

func (r *CreateUserRequest) Validate() error {
    if err := validator.New().Struct(r); err != nil {
        return fmt.Errorf("validate request: %w", err)
    }
    return nil
}
```

**Why this pattern**:
- Uses struct tags for declarative validation
- Centralized validation logic
- Easy to extend with custom validators
- Standard library compatible (github.com/go-playground/validator/v10)

## Pattern 2: Domain Validation

```go
func (u *User) ChangeEmail(newEmail string) error {
    if newEmail == "" {
        return fmt.Errorf("email cannot be empty")
    }

    if u.Email == newEmail {
        return ErrEmailUnchanged
    }

    if !isValidEmail(newEmail) {
        return ErrInvalidEmail
    }

    u.Email = newEmail
    return nil
}
```

**Why this pattern**:
- Domain-specific validation in domain layer
- Returns domain errors
- Encapsulates validation logic
- Clear error messages

## Pattern 3: Schema Validation

```go
type Schema struct {
    Fields []Field
}

type Field struct {
    Name     string
    Type     string
    Required bool
    Min      int
    Max      int
}

func (s *Schema) Validate(data map[string]any) error {
    for _, field := range s.Fields {
        value, exists := data[field.Name]

        if field.Required && !exists {
            return fmt.Errorf("field %s is required", field.Name)
        }

        if exists {
            if err := s.validateField(field, value); err != nil {
                return fmt.Errorf("field %s: %w", field.Name, err)
            }
        }
    }
    return nil
}

func (s *Schema) validateField(field Field, value any) error {
    // Type-specific validation
    switch field.Type {
    case "string":
        str, ok := value.(string)
        if !ok {
            return ErrInvalidType
        }
        if len(str) < field.Min {
            return fmt.Errorf("too short (min %d)", field.Min)
        }
        if len(str) > field.Max {
            return fmt.Errorf("too long (max %d)", field.Max)
        }
    // ... other types
    }
    return nil
}
```

**Why this pattern**:
- Flexible schema definition
- Run-time validation
- Extensible field types
- Clear error reporting
```

### agents/validator.md

```markdown
---
name: validator
description: "Specializes in implementing data validation logic"
model: claude-3-sonnet
color: "#10B981"
skills:
  - validation
  - go-code
tools:
  file-read: true
  file-write: true
  glob: true
  grep: true
---

# Data Validation Agent

You are a specialist in implementing robust data validation.

## Responsibilities

- Design validation strategies
- Implement validation logic
- Create custom validators
- Ensure data integrity
- Handle validation errors gracefully

## Workflow

1. **Analyze Requirements**
   - Identify what needs validation
   - Determine validation rules
   - Check domain constraints

2. **Choose Validation Approach**
   - Struct validation for inputs
   - Domain validation for business logic
   - Schema validation for dynamic data

3. **Implement**
   - Follow patterns from validation skill
   - Use go-code for Go implementation
   - Write validation tests

4. **Validate**
   - Test with valid inputs
   - Test with invalid inputs
   - Verify error messages

## Tools Available

- `file-read`: Read existing code
- `file-write`: Create/modify validation code
- `glob`: Find files to validate
- `grep`: Search for existing patterns

## Skills Available

- `validation`: Validation patterns and best practices
- `go-code`: Go implementation patterns

## Examples

### Example 1: Input Validation

**Request**: Add validation to CreateUserRequest

**Approach**:
1. Analyze request structure
2. Add validation tags
3. Implement Validate() method
4. Add tests for valid/invalid cases

### Example 2: Domain Validation

**Request**: Ensure email uniqueness in User domain

**Approach**:
1. Add domain validation method
2. Return domain errors
3. Document constraints
4. Test edge cases

### Example 3: Custom Validator

**Request**: Create validator for phone numbers

**Approach**:
1. Define validation function
2. Register with validator library
3. Add to struct tags
4. Test various formats
```

### rules/enforce-validation.yaml

```yaml
name: enforce-validation
description: "Reject files without validation in request structs"
priority: high

trigger:
  event: file_modified
  pattern: "*_request.go"
  path: "internal/transport/"

conditions:
  - type: content_matches
    pattern: "type.*Request struct"

  - type: content_not_matches
    pattern: "func \(.*Request\) Validate\(\) error"

actions:
  - type: reject
    message: "Request structs must have Validate() method"

  - type: comment
    message: "Add Validate() method to ensure data integrity"

  - type: notify
    channel: "code-quality"
    message: "Request file without validation detected"
```

## Best Practices

### Plugin Design

1. **Single Purpose**: One plugin = one domain of functionality
2. **Clear Naming**: Use descriptive names for skills and agents
3. **Good Documentation**: Explain why patterns work, not just how
4. **Test Coverage**: Include tests for critical paths
5. **Version Carefully**: Use semantic versioning

### Skill Development

1. **Activation Triggers**: Be specific about when skill activates
2. **Working Examples**: Provide copy-pasteable code
3. **Rationale**: Explain why patterns work
4. **Consistency**: Use consistent formatting and structure

### Agent Development

1. **Clear Scope**: Define what agent is responsible for
2. **Right Tools**: Include only necessary tools
3. **Relevant Skills**: Connect to relevant skills
4. **Examples**: Provide concrete examples of workflows

### Rule Development

1. **Specific Triggers**: Target specific events
2. **Clear Actions**: Define exactly what to do
3. **Prioritize**: Use appropriate priority level
4. **Test**: Verify rules work as expected

## Troubleshooting

### Plugin Not Loading

**Symptoms**: Plugin doesn't appear in list

**Solutions**:
1. Check `plugin.yaml` is valid YAML
2. Verify manifest validation passes
3. Check file paths in manifest are correct
4. Review logs for error messages

### Skill Not Activating

**Symptoms**: Skill content doesn't appear in context

**Solutions**:
1. Check `description` field includes activation triggers
2. Verify trigger keywords are relevant
3. Test with different wording
4. Check skill file is in correct location

### Agent Not Available

**Symptoms**: Agent doesn't appear in available agents

**Solutions**:
1. Check agent file exists at specified path
2. Verify agent YAML frontmatter is valid
3. Check agent name doesn't conflict with existing agents
4. Restart go-ent to reload

### Rule Not Triggering

**Symptoms**: Rule actions don't execute

**Solutions**:
1. Check trigger event matches expected event
2. Verify conditions are correctly defined
3. Check rule priority (higher rules execute first)
4. Review logs for rule evaluation

## Resources

- **Manifest Reference**: `PLUGIN_MANIFEST.md`
- **Marketplace Guide**: `PLUGIN_MARKETPLACE.md`
- **Development Guide**: `DEVELOPMENT.md`
- **Plugin Manager Code**: `internal/plugin/manager.go`
- **E2E Tests**: `internal/plugin/e2e_test.go`
