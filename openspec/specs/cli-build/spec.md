# cli-build Specification

## Purpose
TBD - created by archiving change add-build-infrastructure. Update Purpose after archive.
## Requirements
### Requirement: CLI Build System
The CLI tool SHALL build successfully from the project root using Go build tools.

#### Scenario: Build from root succeeds
- **WHEN** running `make build` from project root
- **THEN** CLI binary is created at `dist/go-ent`
- **AND** binary is executable

#### Scenario: Direct Go build succeeds
- **WHEN** running `go build ./cmd/go-ent`
- **THEN** binary is created successfully
- **AND** no build errors occur

### Requirement: Template Embedding
The CLI tool SHALL embed all project templates including dotfiles at build time using Go embed, with templates located in `/internal/templates/`.

**What changed**: Consolidated templates from `/templates/` and `/cmd/go-ent/templates/` into `/internal/templates/` and fixed embedding of dotfiles (`.gitignore.tmpl`, `.golangci.yml.tmpl`) using explicit embed directives.

#### Scenario: Templates embedded in binary
- **WHEN** CLI is built
- **THEN** all template files from `/internal/templates/` directory are embedded in the binary
- **AND** embedded files include dotfiles (`.gitignore.tmpl`, `.golangci.yml.tmpl`)
- **AND** embedded files are accessible via `embed.FS`
- **AND** no build-time copying is required

#### Scenario: Dotfiles properly embedded
- **WHEN** checking embedded files after build
- **THEN** `.gitignore.tmpl` is included in binary
- **AND** `.golangci.yml.tmpl` is included in binary
- **AND** all other dotfile templates (`.*.tmpl`) are included
- **AND** embed directive uses explicit paths for dotfiles

#### Scenario: Embedded templates accessible at runtime
- **WHEN** `go_ent_generate` tool is executed
- **THEN** CLI can read all template files from embedded filesystem
- **AND** dotfile templates are accessible
- **AND** template files are processed and written to target project directory
- **AND** generated projects include `.gitignore` and `.golangci.yml` files

#### Scenario: Template structure consolidated
- **WHEN** examining project structure
- **THEN** templates exist only at `/internal/templates/`
- **AND** no `/templates/` directory at project root
- **AND** no `/cmd/go-ent/templates/` copy directory
- **AND** embed directive references `/internal/templates/` directly

### Requirement: Template Organization
Templates SHALL be organized in `/internal/templates/` with proper subdirectory structure and explicit embedding directives.

#### Scenario: Template directory structure
- **WHEN** examining `/internal/templates/` structure
- **THEN** root-level templates exist (e.g., `CLAUDE.md.tmpl`, `Makefile.tmpl`)
- **AND** dotfile templates exist (e.g., `.gitignore.tmpl`, `.golangci.yml.tmpl`)
- **AND** subdirectories exist (`build/`, `cmd/`, `deploy/`, `internal/`, `mcp/`)
- **AND** each subdirectory contains appropriate `.tmpl` files

#### Scenario: Embed directives
- **WHEN** reviewing `/internal/templates/embed.go`
- **THEN** root templates embedded with `//go:embed *.tmpl`
- **AND** dotfiles explicitly embedded: `//go:embed .gitignore.tmpl .golangci.yml.tmpl`
- **AND** subdirectories embedded with specific patterns (e.g., `//go:embed mcp/*.tmpl mcp/**/*.tmpl`)
- **AND** no glob patterns like `**/*.tmpl` used (they miss dotfiles)

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

### Requirement: Build Artifacts
The CLI build process SHALL produce clean, versioned artifacts without requiring template preparation steps.

**What changed**: Removed `prepare-templates` Makefile target and template copying from build process.

#### Scenario: Build output location
- **WHEN** `make build` is executed
- **THEN** binary is built directly without template copying
- **AND** no intermediate template directories are created
- **AND** binary is placed in `dist/` directory
- **AND** binary is named `goent`

#### Scenario: Clean build
- **WHEN** `make clean` is executed
- **THEN** `dist/` directory is removed
- **AND** no `/cmd/go-ent/templates/` artifacts remain
- **AND** build artifacts are fully cleaned

### Requirement: Go Module Structure
The CLI SHALL follow clean code principles with no unused directories or files in the project structure.

**What changed**: Removed empty `/cmd/go-ent/internal/resources/` directory that had no purpose or usage.

#### Scenario: No unused directories
- **WHEN** examining `/cmd/go-ent/internal/` structure
- **THEN** only active packages with code exist
- **AND** no empty directories are present
- **AND** all directories serve a documented purpose

#### Scenario: Clean package structure
- **WHEN** listing directories in `/cmd/go-ent/internal/`
- **THEN** only `tools/` and `server/` directories exist (after refactoring)
- **AND** each directory contains Go source files
- **AND** no placeholder or empty packages exist

### Requirement: Makefile Interface
A Makefile SHALL provide standard build targets for CLI development.

#### Scenario: Make targets available
- **WHEN** running `make help` or viewing Makefile
- **THEN** targets include: `build`, `test`, `lint`, `fmt`, `clean`, `validate-plugin`

#### Scenario: Build target
- **WHEN** running `make build`
- **THEN** CLI is compiled and output to `dist/go-ent`

#### Scenario: Clean target
- **WHEN** running `make clean`
- **THEN** `dist/` directory is removed
- **AND** build artifacts are cleaned up

### Requirement: Code Quality Checks
CLI code SHALL be validated with linting tools before commits.

#### Scenario: Linting configuration exists
- **WHEN** checking project root
- **THEN** `.golangci.yml` configuration file exists
- **AND** linters include: errcheck, gosimple, govet, staticcheck, gofmt, goimports

#### Scenario: Lint check passes
- **WHEN** running `make lint`
- **THEN** golangci-lint executes successfully
- **AND** no linting errors are reported for compliant code

### Requirement: Testing Infrastructure
CLI code SHALL have test coverage with race detection enabled.

#### Scenario: Test target exists
- **WHEN** running `make test`
- **THEN** Go tests execute with `-race` and `-cover` flags
- **AND** test results are reported

#### Scenario: Test files exist
- **WHEN** checking `cmd/go-ent/` directory
- **THEN** `main_test.go` file exists
- **AND** test cases cover core functionality

