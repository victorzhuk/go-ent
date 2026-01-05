# CLI Build Capability

## ADDED Requirements

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
The CLI tool SHALL embed all project templates at build time using Go embed.

#### Scenario: Templates embedded in binary
- **WHEN** CLI is built
- **THEN** all template files from `templates/` directory are embedded in the binary
- **AND** embedded files are accessible via `embed.FS`

#### Scenario: Embedded templates accessible at runtime
- **WHEN** CLI command `init` is executed
- **THEN** CLI can read template files from embedded filesystem
- **AND** template files are copied to target project directory

### Requirement: Template Processing
The CLI tool SHALL process template files and replace placeholders with project-specific values.

#### Scenario: Project name replacement
- **WHEN** templates are processed during project initialization
- **THEN** all occurrences of `{{PROJECT_NAME}}` are replaced with the actual project name
- **AND** no `{{PROJECT_NAME}}` placeholders remain in output files

#### Scenario: Module path replacement
- **WHEN** templates are processed during project initialization
- **THEN** all occurrences of `{{MODULE_PATH}}` are replaced with the actual module path
- **AND** no `{{MODULE_PATH}}` placeholders remain in output files

### Requirement: Build Artifacts
The CLI build process SHALL produce clean, versioned artifacts in a dedicated output directory.

#### Scenario: Build output location
- **WHEN** `make build` is executed
- **THEN** binary is placed in `dist/` directory
- **AND** binary is named `goent`

#### Scenario: Clean build
- **WHEN** `make clean && make build` is executed
- **THEN** all previous build artifacts are removed
- **AND** fresh binary is created

### Requirement: Go Module Structure
The CLI SHALL be part of the root Go module to enable template embedding.

#### Scenario: Root module exists
- **WHEN** checking module structure
- **THEN** root `go.mod` exists with module `github.com/victorzhuk/go-ent`
- **AND** CLI source is located at `cmd/go-ent/main.go`

#### Scenario: Template path resolution
- **WHEN** Go embed directive processes `//go:embed ../../templates/*`
- **THEN** templates are found at project root `templates/` directory
- **AND** all template files are successfully embedded

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
