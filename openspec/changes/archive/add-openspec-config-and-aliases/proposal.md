# Change: Add OpenSpec Schema Support

## Why

go-ent currently uses custom `/ent:*` commands. The standard OpenSpec approach uses:
1. `openspec init` - Scaffolds OpenSpec structure and generates skills
2. Standard `/opsx:*` commands - Provided by auto-generated skills
3. Schema definitions - Define artifact types and templates

We should provide a go-ent schema that works with standard OpenSpec CLI.

## What Changes

### ADDED: OpenSpec Schema for Go

Create `openspec/schemas/go-ent/schema.yaml`:

```yaml
name: go-ent
description: Go project with agent-driven workflows

artifacts:
  - id: proposal
    generates: proposal.md
    requires: []
    template: templates/proposal.md
    
  - id: specs
    generates: specs/**/spec.md
    requires: [proposal]
    template: templates/spec.md
    
  - id: design
    generates: design.md
    requires: [proposal]
    template: templates/design.md
    
  - id: tasks
    generates: tasks.md
    requires: [specs, design]
    template: templates/tasks.md
```

### ADDED: Templates

Create templates in `openspec/schemas/go-ent/templates/`:
- `proposal.md` - Go-specific proposal template
- `spec.md` - Specification with ADDED/MODIFIED/REMOVED format
- `design.md` - Design document with Go patterns
- `tasks.md` - Task list with verification steps

### ADDED: `ent init` Command

Simple command to scaffold OpenSpec in a Go project:

```bash
# Create openspec/config.yaml for this project
ent init

# Output: openspec/config.yaml created
# User then runs: openspec init
```

This creates:
```yaml
schema: go-ent
context: |
  ## Go Project
  Module: github.com/user/project
  Go version: 1.23
```

### REMOVED: Custom /ent:* Commands (Future)

Once OpenSpec workflow is working:
- `/ent:plan` → use `/opsx:new`
- `/ent:apply` → use `/opsx:apply`
- `/ent:archive` → use `/opsx:archive`

## Impact

- **Affected specs**: openspec-schema (NEW)
- **Affected code**:
  - `openspec/schemas/go-ent/schema.yaml` (NEW)
  - `openspec/schemas/go-ent/templates/*.md` (NEW)
  - `internal/cli/init.go` (NEW - simple init command)
- **New dependency**: Users need `openspec` CLI installed
- **Migration**: `/ent:*` commands remain during transition

## Alternatives Considered

1. **Keep /ent:* commands only**: Rejected - not standard OpenSpec
2. **Generate everything from agents**: Rejected - over-engineered
3. **✅ Simple schema + init command**: Selected - clean, standard, maintainable

## Success Criteria

- [ ] `openspec schemas` lists `go-ent` schema
- [ ] `openspec init --schema go-ent` works
- [ ] `/opsx:*` commands work with go-ent schema
- [ ] `ent init` creates basic openspec/config.yaml
- [ ] Templates include Go-specific guidance
