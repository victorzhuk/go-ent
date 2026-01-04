# mcp-tools Specification

## Purpose
MCP (Model Context Protocol) tools for project scaffolding, spec validation, and change management.
## Requirements
### Requirement: Project Generation Tool

The MCP server SHALL provide a `goent_generate` tool that creates new Go projects from embedded templates.

#### Scenario: Generate standard project
- **WHEN** `goent_generate` is called with `project_type: "standard"`, `path: "/tmp/myproject"`, `module_path: "github.com/user/myproject"`
- **THEN** a complete Go project structure is created at the specified path
- **AND** all template placeholders are replaced with provided values
- **AND** the project builds successfully with `go build ./...`

#### Scenario: Generate MCP server project
- **WHEN** `goent_generate` is called with `project_type: "mcp"`, `path: "/tmp/mymcp"`, `module_path: "github.com/user/mymcp"`
- **THEN** an MCP server project structure is created at the specified path
- **AND** the project includes MCP server boilerplate code
- **AND** the project builds successfully with `go build ./...`

#### Scenario: Target directory exists
- **WHEN** `goent_generate` is called with a path that already exists and contains files
- **THEN** the tool returns an error
- **AND** no files are modified

#### Scenario: Invalid project type
- **WHEN** `goent_generate` is called with an unknown `project_type`
- **THEN** the tool returns an error listing valid project types

### Requirement: Spec Validation Tool

The MCP server SHALL provide a `goent_spec_validate` tool that validates specs and change proposals.

#### Scenario: Validate change proposal
- **WHEN** `goent_spec_validate` is called with `type: "change"`, `id: "add-feature"`
- **THEN** the tool validates the change directory structure
- **AND** validates proposal.md, tasks.md, and spec deltas exist
- **AND** validates all requirements have at least one scenario
- **AND** returns a validation report with errors and warnings

#### Scenario: Validate spec
- **WHEN** `goent_spec_validate` is called with `type: "spec"`, `id: "auth"`
- **THEN** the tool validates the spec file format
- **AND** validates scenario headers use `#### Scenario:` format
- **AND** validates requirement headers use `### Requirement:` format
- **AND** returns a validation report

#### Scenario: Strict validation mode
- **WHEN** `goent_spec_validate` is called with `strict: true`
- **THEN** warnings are treated as errors
- **AND** validation fails if any issues are found

#### Scenario: Validation passes
- **WHEN** a valid spec or change is validated
- **THEN** the tool returns success status
- **AND** reports "No issues found"

#### Scenario: Validation fails with details
- **WHEN** an invalid spec or change is validated
- **THEN** the tool returns failure status
- **AND** lists each issue with file path and line number
- **AND** categorizes issues as error or warning

### Requirement: Change Archive Tool

The MCP server SHALL provide a `goent_spec_archive` tool that archives completed changes and updates specs.

#### Scenario: Archive change successfully
- **WHEN** `goent_spec_archive` is called with `id: "add-feature"`
- **THEN** the change is validated first
- **AND** delta specs are merged into main specs in `specs/`
- **AND** the change directory is moved to `changes/archive/YYYY-MM-DD-add-feature/`
- **AND** a success message confirms the archive

#### Scenario: Archive with skip-specs option
- **WHEN** `goent_spec_archive` is called with `id: "refactor-tooling"`, `skip_specs: true`
- **THEN** the change directory is moved to archive
- **AND** main specs are NOT modified
- **AND** this is useful for tooling-only changes

#### Scenario: Archive fails validation
- **WHEN** `goent_spec_archive` is called for a change that fails validation
- **THEN** the archive is aborted
- **AND** no files are moved or modified
- **AND** validation errors are returned

#### Scenario: Dry run archive
- **WHEN** `goent_spec_archive` is called with `dry_run: true`
- **THEN** the tool reports what would be changed
- **AND** no files are actually moved or modified

### Requirement: Template Embedding

The MCP server binary SHALL embed all project templates at build time.

#### Scenario: Templates accessible at runtime
- **WHEN** the goent binary is executed
- **THEN** all template files from `templates/` are accessible via embedded filesystem
- **AND** templates can be read without external file dependencies

#### Scenario: Template variable substitution
- **WHEN** templates are processed during generation
- **THEN** `{{.ModulePath}}` is replaced with the module path
- **AND** `{{.ProjectName}}` is replaced with the project name
- **AND** `{{.GoVersion}}` is replaced with the Go version
- **AND** no unsubstituted placeholders remain in output

### Requirement: MCP Tool Input Schemas

All MCP tools SHALL define JSON input schemas for better client compatibility.

#### Scenario: Tool has input schema
- **WHEN** MCP client requests tool list
- **THEN** each tool includes an `inputSchema` property
- **AND** schema defines required and optional parameters
- **AND** schema includes parameter types and descriptions

#### Scenario: Invalid input rejected
- **WHEN** a tool is called with parameters not matching its schema
- **THEN** the tool returns a clear error message
- **AND** error indicates which parameters are invalid

### Requirement: Generation Configuration
The system SHALL support optional `generation.yaml` configuration file in the `openspec/` directory.

#### Scenario: Parse generation config
- **WHEN** `generation.yaml` exists in `openspec/` directory
- **THEN** the system parses and applies configuration settings

#### Scenario: Use defaults when no config
- **WHEN** `generation.yaml` does not exist
- **THEN** the system uses built-in default settings

#### Scenario: Override archetype
- **WHEN** component specifies `archetype` in generation.yaml
- **THEN** that archetype is used instead of auto-detected one

### Requirement: Project Archetypes
The system SHALL provide built-in project archetypes with predefined template sets.

#### Scenario: Standard archetype
- **WHEN** archetype is `standard`
- **THEN** templates for web service with clean architecture are selected

#### Scenario: MCP archetype
- **WHEN** archetype is `mcp`
- **THEN** templates for MCP server plugin are selected

#### Scenario: Custom archetype
- **WHEN** user defines custom archetype in generation.yaml
- **THEN** the custom template set is used for generation

### Requirement: Archetype Discovery
The system SHALL provide a tool to list available project archetypes.

#### Scenario: List all archetypes
- **WHEN** `goent_list_archetypes` is called without filter
- **THEN** all built-in and custom archetypes are returned with metadata

#### Scenario: Filter archetypes
- **WHEN** `goent_list_archetypes` is called with type filter
- **THEN** only matching archetypes are returned

### Requirement: Spec Analysis
The system SHALL analyze spec files to identify patterns and components.

#### Scenario: Detect CRUD pattern
- **WHEN** spec contains create/read/update/delete scenarios
- **THEN** CRUD pattern is identified with high confidence

#### Scenario: Detect API pattern
- **WHEN** spec contains endpoint/request/response language
- **THEN** API pattern is identified

#### Scenario: Return analysis confidence
- **WHEN** spec is analyzed
- **THEN** confidence score (0.0-1.0) is included in results

### Requirement: Archetype Selection
The system SHALL map spec analysis to recommended archetypes.

#### Scenario: Auto-select archetype
- **WHEN** no archetype is specified in generation.yaml
- **THEN** archetype is selected based on spec analysis

#### Scenario: Explicit override
- **WHEN** archetype is specified in generation.yaml
- **THEN** explicit archetype takes precedence over auto-selection

### Requirement: Component Generation
The system SHALL generate component scaffolds from spec and templates.

#### Scenario: Generate from spec
- **WHEN** `goent_generate_component` is called with spec path
- **THEN** component scaffold is generated using selected templates

#### Scenario: Mark extension points
- **WHEN** component is generated
- **THEN** extension points are marked with `@generate:` comments

#### Scenario: Include prompt context
- **WHEN** extension points are generated
- **THEN** relevant spec requirements are included as context

### Requirement: Spec-Driven Generation
The system SHALL generate complete projects from spec files.

#### Scenario: Full project generation
- **WHEN** `goent_generate_from_spec` is called
- **THEN** all identified components are generated

#### Scenario: Component integration
- **WHEN** multiple components are generated
- **THEN** integration points between components are created

### Requirement: AI Prompt Templates
The system SHALL provide prompt templates for AI-assisted code generation.

#### Scenario: Load prompt template
- **WHEN** prompt template is requested by type
- **THEN** template is loaded from `prompts/` directory

#### Scenario: Substitute variables
- **WHEN** prompt template is populated
- **THEN** spec content and requirements are substituted

### Requirement: Extension Points
The system SHALL mark extension points in generated code for AI completion.

#### Scenario: Constructor extension
- **WHEN** `@generate:constructor` marker is present
- **THEN** AI can generate dependency injection code

#### Scenario: Methods extension
- **WHEN** `@generate:methods` marker is present
- **THEN** AI can generate business logic methods from spec

