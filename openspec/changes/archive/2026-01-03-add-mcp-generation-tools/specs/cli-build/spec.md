## MODIFIED Requirements

### Requirement: Template Embedding
The CLI tool SHALL embed all project templates at build time using Go embed.

#### Scenario: Templates embedded in binary
- **WHEN** CLI is built
- **THEN** all template files from `templates/` directory are embedded in the binary
- **AND** embedded files are accessible via `embed.FS`

#### Scenario: Embedded templates accessible at runtime
- **WHEN** `go_ent_generate` tool is executed
- **THEN** CLI can read template files from embedded filesystem
- **AND** template files are processed and written to target project directory

#### Scenario: Build copies templates first
- **WHEN** `make build` is executed
- **THEN** templates are copied from root `templates/` to `cmd/go-ent/templates/`
- **AND** `//go:embed` directive includes copied templates
- **AND** binary contains all template files

### Requirement: Template Processing
The CLI tool SHALL process template files and replace placeholders with project-specific values.

#### Scenario: Go template syntax
- **WHEN** templates are processed during project generation
- **THEN** Go `text/template` syntax is used (e.g., `{{.ModulePath}}`)
- **AND** template variables are substituted correctly

#### Scenario: Module path replacement
- **WHEN** templates are processed during project generation
- **THEN** all occurrences of `{{.ModulePath}}` are replaced with the actual module path
- **AND** no template placeholders remain in output files

#### Scenario: Project name replacement
- **WHEN** templates are processed during project generation
- **THEN** all occurrences of `{{.ProjectName}}` are replaced with the actual project name
- **AND** the project name is derived from the module path if not specified
