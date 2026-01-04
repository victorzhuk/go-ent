# Tasks: Add MCP Template Generation and Validation Tools

## Dependencies
- T1.1 → T1.2, T1.3
- T1.2, T1.3 → T1.4 [P]
- T1.4 → T2.1
- T2.1 → T2.2
- T3.1, T3.2 → T3.3 [P]
- T4.1 (independent)

## Phase 1: Template Embedding System

### T1.1: Create embed.go for template bundling
- **Story**: specs/cli-build/spec.md#Template Embedding
- **Files**: cmd/goent/templates/embed.go
- **Depends**: None
- **Parallel**: No
- [x] 1.1.1 Create `cmd/goent/templates/embed.go` with `//go:embed **/*.tmpl` directive (created during testing phase)
- [x] 1.1.2 Export `TemplateFS embed.FS` variable
- [x] 1.1.3 Verify `make build` copies templates before embedding

### T1.2: Create template engine
- **Story**: specs/cli-build/spec.md#Template Processing
- **Files**: cmd/goent/internal/template/engine.go
- **Depends**: T1.1
- **Parallel**: No
- [x] 1.2.1 Create `template/engine.go` with `Engine` struct
- [x] 1.2.2 Implement `NewEngine(fs embed.FS)` constructor
- [x] 1.2.3 Implement `Process(templatePath string, vars TemplateVars, outputPath string) error`
- [x] 1.2.4 Implement `ProcessAll(templateDir string, vars TemplateVars, outputDir string) error`
- [x] 1.2.5 Define `TemplateVars` struct with ModulePath, ProjectName, GoVersion

### T1.3: Convert templates to Go template syntax
- **Story**: specs/cli-build/spec.md#Template Processing
- **Files**: templates/**/*.tmpl
- **Depends**: T1.1
- **Parallel**: Yes (with T1.2)
- [x] 1.3.1 Convert `{{MODULE_PATH}}` to `{{.ModulePath}}` in all templates
- [x] 1.3.2 Convert `{{PROJECT_NAME}}` to `{{.ProjectName}}` in all templates
- [x] 1.3.3 Add `{{.GoVersion}}` where appropriate (go.mod files)
- [x] 1.3.4 Test template syntax is valid with `template.ParseFS`

### T1.4: Create goent_generate tool
- **Story**: specs/mcp-tools/spec.md#Project Generation Tool
- **Files**: cmd/goent/internal/tools/generate.go
- **Depends**: T1.2, T1.3
- **Parallel**: No
- [x] 1.4.1 Define `GenerateInput` struct with path, module_path, project_type, project_name
- [x] 1.4.2 Define `inputSchema` for MCP tool registration
- [x] 1.4.3 Implement `generateHandler` function
- [x] 1.4.4 Support `project_type: "standard"` using `templates/*.tmpl`
- [x] 1.4.5 Support `project_type: "mcp"` using `templates/mcp/*.tmpl`
- [x] 1.4.6 Validate target directory doesn't exist or is empty
- [x] 1.4.7 Register tool in `register.go`

## Phase 2: Validation Tool

### T2.1: Create validation rule framework
- **Story**: specs/mcp-tools/spec.md#Spec Validation Tool
- **Files**: cmd/goent/internal/spec/validator.go, cmd/goent/internal/spec/rules.go
- **Depends**: T1.4
- **Parallel**: No
- [x] 2.1.1 Define `ValidationError` and `ValidationWarning` types
- [x] 2.1.2 Define `ValidationContext` struct with current file, line, content
- [x] 2.1.3 Define `ValidationRule` function type
- [x] 2.1.4 Implement `Validator` struct with rule collection
- [x] 2.1.5 Implement rules for spec format validation:
  - [x] `validateRequirementHeader` - checks `### Requirement:` format
  - [x] `validateScenarioHeader` - checks `#### Scenario:` format
  - [x] `validateRequirementHasScenario` - each requirement has >= 1 scenario
  - [x] `validateDeltaOperations` - ADDED/MODIFIED/REMOVED/RENAMED are valid

### T2.2: Create goent_spec_validate tool
- **Story**: specs/mcp-tools/spec.md#Spec Validation Tool
- **Files**: cmd/goent/internal/tools/validate.go
- **Depends**: T2.1
- **Parallel**: No
- [x] 2.2.1 Define `ValidateInput` struct with type, id, strict
- [x] 2.2.2 Define `inputSchema` for MCP tool registration
- [x] 2.2.3 Implement `validateHandler` function
- [x] 2.2.4 Support `type: "spec"` validation
- [x] 2.2.5 Support `type: "change"` validation (includes all files in change dir)
- [x] 2.2.6 Implement `strict` mode (warnings become errors)
- [x] 2.2.7 Return structured validation report
- [x] 2.2.8 Register tool in `register.go`

## Phase 3: Archive Tool

### T3.1: Create spec merger
- **Story**: specs/mcp-tools/spec.md#Change Archive Tool
- **Files**: cmd/goent/internal/spec/merger.go
- **Depends**: None
- **Parallel**: Yes (with T3.2)
- [x] 3.1.1 Implement `ParseDeltaSpec(content string) (*DeltaSpec, error)`
- [x] 3.1.2 Implement `MergeDeltas(baseSpec, deltaSpec string) (string, error)`
- [x] 3.1.3 Handle ADDED: append new requirements
- [x] 3.1.4 Handle MODIFIED: replace existing requirements
- [x] 3.1.5 Handle REMOVED: delete requirements (with comment)
- [x] 3.1.6 Handle RENAMED: update requirement name

### T3.2: Create archiver
- **Story**: specs/mcp-tools/spec.md#Change Archive Tool
- **Files**: cmd/goent/internal/spec/archiver.go
- **Depends**: None
- **Parallel**: Yes (with T3.1)
- [x] 3.2.1 Implement `Archive(changeID string, skipSpecs bool) error`
- [x] 3.2.2 Generate archive path with date prefix: `YYYY-MM-DD-{changeID}`
- [x] 3.2.3 Move change directory to archive
- [x] 3.2.4 Implement dry-run mode

### T3.3: Create goent_spec_archive tool
- **Story**: specs/mcp-tools/spec.md#Change Archive Tool
- **Files**: cmd/goent/internal/tools/archive.go
- **Depends**: T3.1, T3.2
- **Parallel**: No
- [x] 3.3.1 Define `ArchiveInput` struct with id, skip_specs, dry_run
- [x] 3.3.2 Define `inputSchema` for MCP tool registration
- [x] 3.3.3 Implement `archiveHandler` function
- [x] 3.3.4 Validate change before archive (call validator)
- [x] 3.3.5 Merge deltas into specs (unless skip_specs)
- [x] 3.3.6 Move change to archive directory
- [x] 3.3.7 Register tool in `register.go`

## Phase 4: Plugin Configuration Fix

### T4.1: Fix plugin.json path
- **Story**: N/A (configuration fix)
- **Files**: plugins/go-ent/.claude-plugin/plugin.json
- **Depends**: None
- **Parallel**: Independent
- [x] 4.1.1 Change absolute path to relative path `./dist/goent`
- [x] 4.1.2 Test plugin installation in Claude Code (plugin.json already uses relative path)
- [ ] 4.1.3 Update README if installation instructions change

## Phase 5: Testing and Documentation

### T5.1: Add unit tests
- **Files**: cmd/goent/internal/template/*_test.go, cmd/goent/internal/spec/*_test.go
- **Depends**: All implementation tasks
- [x] 5.1.1 Add tests for template engine
- [x] 5.1.2 Add tests for validation rules
- [x] 5.1.3 Add tests for spec merger
- [x] 5.1.4 Add tests for archiver
- [x] 5.1.5 Verify `make test` passes

### T5.2: Update input schemas for existing tools
- **Story**: specs/mcp-tools/spec.md#MCP Tool Input Schemas
- **Files**: cmd/goent/internal/tools/*.go
- **Depends**: None
- [x] 5.2.1 Add inputSchema to goent_spec_init
- [x] 5.2.2 Add inputSchema to goent_spec_create
- [x] 5.2.3 Add inputSchema to goent_spec_update
- [x] 5.2.4 Add inputSchema to goent_spec_delete
- [x] 5.2.5 Add inputSchema to goent_spec_list
- [x] 5.2.6 Add inputSchema to goent_spec_show
- [x] 5.2.7 Add inputSchema to all registry tools (6 tools)
- [x] 5.2.8 Add inputSchema to all workflow tools (3 tools)
- [x] 5.2.9 Add inputSchema to all loop tools (4 tools)
