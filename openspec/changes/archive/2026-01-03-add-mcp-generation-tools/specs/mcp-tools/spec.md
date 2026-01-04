## ADDED Requirements

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
