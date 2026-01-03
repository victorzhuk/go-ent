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
- [ ] 1.1.1 Create `cmd/goent/templates/embed.go` with `//go:embed **/*.tmpl` directive
- [ ] 1.1.2 Export `TemplateFS embed.FS` variable
- [ ] 1.1.3 Verify `make build` copies templates before embedding

### T1.2: Create template engine
- **Story**: specs/cli-build/spec.md#Template Processing
- **Files**: cmd/goent/internal/template/engine.go
- **Depends**: T1.1
- **Parallel**: No
- [ ] 1.2.1 Create `template/engine.go` with `Engine` struct
- [ ] 1.2.2 Implement `NewEngine(fs embed.FS)` constructor
- [ ] 1.2.3 Implement `Process(templatePath string, vars TemplateVars, outputPath string) error`
- [ ] 1.2.4 Implement `ProcessAll(templateDir string, vars TemplateVars, outputDir string) error`
- [ ] 1.2.5 Define `TemplateVars` struct with ModulePath, ProjectName, GoVersion

### T1.3: Convert templates to Go template syntax
- **Story**: specs/cli-build/spec.md#Template Processing
- **Files**: templates/**/*.tmpl
- **Depends**: T1.1
- **Parallel**: Yes (with T1.2)
- [ ] 1.3.1 Convert `{{MODULE_PATH}}` to `{{.ModulePath}}` in all templates
- [ ] 1.3.2 Convert `{{PROJECT_NAME}}` to `{{.ProjectName}}` in all templates
- [ ] 1.3.3 Add `{{.GoVersion}}` where appropriate (go.mod files)
- [ ] 1.3.4 Test template syntax is valid with `template.ParseFS`

### T1.4: Create goent_generate tool
- **Story**: specs/mcp-tools/spec.md#Project Generation Tool
- **Files**: cmd/goent/internal/tools/generate.go
- **Depends**: T1.2, T1.3
- **Parallel**: No
- [ ] 1.4.1 Define `GenerateInput` struct with path, module_path, project_type, project_name
- [ ] 1.4.2 Define `inputSchema` for MCP tool registration
- [ ] 1.4.3 Implement `generateHandler` function
- [ ] 1.4.4 Support `project_type: "standard"` using `templates/*.tmpl`
- [ ] 1.4.5 Support `project_type: "mcp"` using `templates/mcp/*.tmpl`
- [ ] 1.4.6 Validate target directory doesn't exist or is empty
- [ ] 1.4.7 Register tool in `register.go`

## Phase 2: Validation Tool

### T2.1: Create validation rule framework
- **Story**: specs/mcp-tools/spec.md#Spec Validation Tool
- **Files**: cmd/goent/internal/spec/validator.go, cmd/goent/internal/spec/rules.go
- **Depends**: T1.4
- **Parallel**: No
- [ ] 2.1.1 Define `ValidationError` and `ValidationWarning` types
- [ ] 2.1.2 Define `ValidationContext` struct with current file, line, content
- [ ] 2.1.3 Define `ValidationRule` function type
- [ ] 2.1.4 Implement `Validator` struct with rule collection
- [ ] 2.1.5 Implement rules for spec format validation:
  - [ ] `validateRequirementHeader` - checks `### Requirement:` format
  - [ ] `validateScenarioHeader` - checks `#### Scenario:` format
  - [ ] `validateRequirementHasScenario` - each requirement has >= 1 scenario
  - [ ] `validateDeltaOperations` - ADDED/MODIFIED/REMOVED/RENAMED are valid

### T2.2: Create goent_spec_validate tool
- **Story**: specs/mcp-tools/spec.md#Spec Validation Tool
- **Files**: cmd/goent/internal/tools/validate.go
- **Depends**: T2.1
- **Parallel**: No
- [ ] 2.2.1 Define `ValidateInput` struct with type, id, strict
- [ ] 2.2.2 Define `inputSchema` for MCP tool registration
- [ ] 2.2.3 Implement `validateHandler` function
- [ ] 2.2.4 Support `type: "spec"` validation
- [ ] 2.2.5 Support `type: "change"` validation (includes all files in change dir)
- [ ] 2.2.6 Implement `strict` mode (warnings become errors)
- [ ] 2.2.7 Return structured validation report
- [ ] 2.2.8 Register tool in `register.go`

## Phase 3: Archive Tool

### T3.1: Create spec merger
- **Story**: specs/mcp-tools/spec.md#Change Archive Tool
- **Files**: cmd/goent/internal/spec/merger.go
- **Depends**: None
- **Parallel**: Yes (with T3.2)
- [ ] 3.1.1 Implement `ParseDeltaSpec(content string) (*DeltaSpec, error)`
- [ ] 3.1.2 Implement `MergeDeltas(baseSpec, deltaSpec string) (string, error)`
- [ ] 3.1.3 Handle ADDED: append new requirements
- [ ] 3.1.4 Handle MODIFIED: replace existing requirements
- [ ] 3.1.5 Handle REMOVED: delete requirements (with comment)
- [ ] 3.1.6 Handle RENAMED: update requirement name

### T3.2: Create archiver
- **Story**: specs/mcp-tools/spec.md#Change Archive Tool
- **Files**: cmd/goent/internal/spec/archiver.go
- **Depends**: None
- **Parallel**: Yes (with T3.1)
- [ ] 3.2.1 Implement `Archive(changeID string, skipSpecs bool) error`
- [ ] 3.2.2 Generate archive path with date prefix: `YYYY-MM-DD-{changeID}`
- [ ] 3.2.3 Move change directory to archive
- [ ] 3.2.4 Implement dry-run mode

### T3.3: Create goent_spec_archive tool
- **Story**: specs/mcp-tools/spec.md#Change Archive Tool
- **Files**: cmd/goent/internal/tools/archive.go
- **Depends**: T3.1, T3.2
- **Parallel**: No
- [ ] 3.3.1 Define `ArchiveInput` struct with id, skip_specs, dry_run
- [ ] 3.3.2 Define `inputSchema` for MCP tool registration
- [ ] 3.3.3 Implement `archiveHandler` function
- [ ] 3.3.4 Validate change before archive (call validator)
- [ ] 3.3.5 Merge deltas into specs (unless skip_specs)
- [ ] 3.3.6 Move change to archive directory
- [ ] 3.3.7 Register tool in `register.go`

## Phase 4: Plugin Configuration Fix

### T4.1: Fix plugin.json path
- **Story**: N/A (configuration fix)
- **Files**: plugins/go-ent/.claude-plugin/plugin.json
- **Depends**: None
- **Parallel**: Independent
- [ ] 4.1.1 Change absolute path to relative path `./dist/goent`
- [ ] 4.1.2 Test plugin installation in Claude Code
- [ ] 4.1.3 Update README if installation instructions change

## Phase 5: Testing and Documentation

### T5.1: Add unit tests
- **Files**: cmd/goent/internal/template/*_test.go, cmd/goent/internal/spec/*_test.go
- **Depends**: All implementation tasks
- [ ] 5.1.1 Add tests for template engine
- [ ] 5.1.2 Add tests for validation rules
- [ ] 5.1.3 Add tests for spec merger
- [ ] 5.1.4 Add tests for archiver
- [ ] 5.1.5 Verify `make test` passes

### T5.2: Update input schemas for existing tools
- **Story**: specs/mcp-tools/spec.md#MCP Tool Input Schemas
- **Files**: cmd/goent/internal/tools/*.go
- **Depends**: None
- [ ] 5.2.1 Add inputSchema to goent_spec_init
- [ ] 5.2.2 Add inputSchema to goent_spec_create
- [ ] 5.2.3 Add inputSchema to goent_spec_update
- [ ] 5.2.4 Add inputSchema to goent_spec_delete
- [ ] 5.2.5 Add inputSchema to goent_spec_list
- [ ] 5.2.6 Add inputSchema to goent_spec_show
- [ ] 5.2.7 Add inputSchema to all registry tools
- [ ] 5.2.8 Add inputSchema to all workflow tools
- [ ] 5.2.9 Add inputSchema to all loop tools
