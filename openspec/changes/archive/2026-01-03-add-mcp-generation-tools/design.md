# Design: MCP Template Generation and Validation Tools

## Context

The go-ent project aims to be an enterprise Go development toolkit with:
- Spec-driven development workflows (OpenSpec)
- MCP-based tooling for Claude Code integration
- Project scaffolding from templates

Currently, the MCP server handles spec/task management but lacks generation, validation, and archive capabilities. Templates exist but are not embedded or processed.

### Constraints
- Must work as stdio MCP server (no HTTP)
- Templates must be embedded in binary (single-file distribution)
- Must integrate with existing tool registration pattern
- Claude Code plugin must work when installed from marketplace

### Stakeholders
- Developers using go-ent to scaffold new projects
- AI assistants using MCP tools for spec-driven workflows
- Claude Code marketplace for plugin distribution

## Goals / Non-Goals

### Goals
- Embed templates in binary using Go 1.16+ embed
- Create `go_ent_generate` tool for project scaffolding
- Create `go_ent_spec_validate` tool for spec validation
- Create `go_ent_spec_archive` tool for change archival
- Fix plugin.json for marketplace compatibility

### Non-Goals
- Adding CLI subcommands (MCP-only for now)
- Custom template syntax (use Go text/template)
- Template hot-reloading (embedded only)
- GUI or interactive prompts

## Decisions

### Decision 1: Template Embedding Strategy

**Choice**: Use `//go:embed` with templates copied to `cmd/go-ent/templates/` at build time

**Rationale**:
- Go embed requires files to be in/below the package directory
- Templates live at project root for easy editing
- Makefile `prepare-templates` already copies them

**Structure**:
```
cmd/go-ent/
├── templates/          # Copied from root at build time
│   ├── embed.go        # //go:embed directive
│   ├── go.mod.tmpl
│   └── ...
└── internal/
    └── tools/
        └── generate.go # Uses embedded FS
```

**Alternatives considered**:
- Symlinks: Don't work with go:embed
- Move templates to cmd/go-ent/: Breaks clean separation
- External template loading: Requires file distribution

### Decision 2: Template Engine

**Choice**: Use Go `text/template` with simple variable map

**Rationale**:
- Standard library, no dependencies
- Supports conditionals and loops if needed later
- `.tmpl` extension already used

**Variables**:
```go
type TemplateVars struct {
    ModulePath  string // e.g., "github.com/user/project"
    ProjectName string // e.g., "my-project"
    GoVersion   string // e.g., "1.23"
}
```

**Template syntax**:
```
module {{.ModulePath}}

go {{.GoVersion}}
```

Note: Current templates use `{{MODULE_PATH}}` style. Will convert to Go template syntax `{{.ModulePath}}`.

### Decision 3: Project Types

**Choice**: Support two project archetypes via `project_type` parameter

| Type | Description | Templates Used |
|------|-------------|----------------|
| `standard` | Web service with clean architecture | `templates/*.tmpl` |
| `mcp` | MCP server plugin | `templates/mcp/*.tmpl` |

**Rationale**:
- Templates already organized this way
- Clear separation of concerns
- Extensible to more types later

### Decision 4: Validation Rules

**Choice**: Implement validation as composable rule functions

```go
type ValidationRule func(ctx *ValidationContext) []ValidationError

var specRules = []ValidationRule{
    validateRequirementHasScenario,
    validateScenarioFormat,
    validateDeltaOperations,
    validateCrossReferences,
}
```

**Validation categories**:
1. **Structural**: Required files exist, correct directories
2. **Format**: Scenario headers use `####`, requirements use `### Requirement:`
3. **Semantic**: Delta ops reference existing requirements, no orphaned tasks
4. **Cross-reference**: Task references match requirement IDs

### Decision 5: Archive Process

**Choice**: Atomic archive with spec merging

**Process**:
1. Validate change passes `--strict`
2. Read all delta specs from `changes/{id}/specs/`
3. Merge deltas into main specs in `specs/`
4. Move `changes/{id}/` to `changes/archive/YYYY-MM-DD-{id}/`
5. Write updated specs

**Rollback**: If any step fails, no changes are made (read-then-write pattern)

### Decision 6: Plugin Path Resolution

**Choice**: Use relative path from plugin installation directory

**Current** (broken):
```json
"mcp": {
  "command": "/home/zhuk/Projects/own/go-ent/dist/go-ent"
}
```

**Fixed**:
```json
"mcp": {
  "command": "./dist/go-ent"
}
```

Or if Claude Code supports it:
```json
"mcp": {
  "command": "${PLUGIN_DIR}/dist/go-ent"
}
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Template syntax change breaks existing templates | Convert all templates in same PR, test thoroughly |
| Archive corrupts specs on partial failure | Use read-then-write pattern, validate before write |
| Plugin path may not resolve correctly | Test with actual Claude Code installation |
| Validation too strict blocks valid specs | Start with warnings, promote to errors iteratively |

## Migration Plan

1. **Phase 1**: Add embed.go and generate tool (no breaking changes)
2. **Phase 2**: Add validate tool (no breaking changes)
3. **Phase 3**: Add archive tool (no breaking changes)
4. **Phase 4**: Convert template syntax (update all .tmpl files)
5. **Phase 5**: Fix plugin.json path

All phases are additive. Rollback is simply reverting the commit.

## Open Questions

1. Should `go_ent_generate` create a git repository? (Tentative: no, let user init)
2. Should validation return warnings vs errors? (Tentative: both, with severity)
3. Should archive support `--dry-run`? (Tentative: yes, for safety)

## Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                      MCP Server                              │
│  cmd/go-ent/main.go                                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   tools/register.go                          │
│  Register all tools with MCP server                         │
└─────────────────────────────────────────────────────────────┘
         │              │              │              │
         ▼              ▼              ▼              ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│ generate.go │ │ validate.go │ │ archive.go  │ │ crud.go     │
│ NEW         │ │ NEW         │ │ NEW         │ │ existing    │
└──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └─────────────┘
       │               │               │
       ▼               ▼               ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│ templates/  │ │ spec/       │ │ spec/       │
│ embed.go    │ │ validator.go│ │ archiver.go │
│ engine.go   │ │ rules.go    │ │ merger.go   │
└─────────────┘ └─────────────┘ └─────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│                   Embedded Templates                         │
│  //go:embed **/*.tmpl                                       │
│  var TemplateFS embed.FS                                    │
└─────────────────────────────────────────────────────────────┘
```
