# Spec Delta: Templates

## ADDED Requirements

### Requirement: MCP Server Template Support
The system SHALL provide Model Context Protocol (MCP) server project templates.

#### Scenario: User selects MCP template type
- **GIVEN** user initiates project scaffolding
- **WHEN** user specifies `template_type=mcp`
- **THEN** system generates MCP server structure from `templates/mcp/` directory
- **AND** generated `go.mod` includes `github.com/modelcontextprotocol/go-sdk v1.2.0`
- **AND** generated `main.go` uses stdio transport for MCP communication

#### Scenario: MCP server binary naming
- **GIVEN** MCP project is scaffolded
- **WHEN** user runs `make build`
- **THEN** binary name SHALL be `{{PROJECT_NAME}}` without `srv_` prefix
- **AND** binary location SHALL be `bin/{{PROJECT_NAME}}`

#### Scenario: MCP server structure
- **GIVEN** MCP template is generated
- **WHEN** examining project structure
- **THEN** directory SHALL contain:
  - `cmd/server/main.go` with run() pattern
  - `internal/server/server.go` with MCP SDK setup
  - Tool registration functions
  - Signal handling (SIGTERM, SIGINT, SIGQUIT)

### Requirement: Distroless Docker Images
The system SHALL generate Dockerfile templates using distroless base images with minimal bash support.

#### Scenario: Builder stage uses Debian Trixie
- **GIVEN** Dockerfile template is generated
- **WHEN** examining builder stage
- **THEN** base image SHALL be `golang:1.25.5-trixie`
- **AND** builder SHALL install `bash-static` package
- **AND** builder SHALL accept `VERSION` and `VCS_REF` build arguments

#### Scenario: Runtime stage uses distroless
- **GIVEN** Dockerfile template is generated
- **WHEN** examining runtime stage
- **THEN** base image SHALL be `gcr.io/distroless/static-debian13:nonroot`
- **AND** runtime SHALL include `/usr/bin/bash` from bash-static
- **AND** runtime SHALL run as non-root user by default

#### Scenario: Entrypoint flexibility with bash
- **GIVEN** Docker image is built from template
- **WHEN** container starts
- **THEN** ENTRYPOINT SHALL be `["/usr/bin/bash"]`
- **AND** CMD SHALL be `["-c", "/{{PROJECT_NAME}} serve"]`
- **AND** users SHALL be able to override CMD with custom arguments

#### Scenario: Binary location in Docker image
- **GIVEN** Dockerfile template is generated
- **WHEN** examining COPY instructions
- **THEN** binary SHALL be copied from `bin/{{PROJECT_NAME}}` in builder
- **AND** binary SHALL be placed at `/{{PROJECT_NAME}}` in runtime image

### Requirement: Build Metadata in Templates
The system SHALL include VERSION and VCS_REF build metadata in generated Makefiles and Dockerfiles.

#### Scenario: Makefile includes version variables
- **GIVEN** Makefile template is generated
- **WHEN** examining version configuration
- **THEN** Makefile SHALL define `VERSION` from git tags or "dev"
- **AND** Makefile SHALL define `VCS_REF` from git commit hash or "unknown"
- **AND** docker target SHALL pass `--build-arg VERSION=$(VERSION)`
- **AND** docker target SHALL pass `--build-arg VCS_REF=$(VCS_REF)`

#### Scenario: Dockerfile accepts build args
- **GIVEN** Dockerfile template is generated
- **WHEN** examining ARG declarations
- **THEN** Dockerfile SHALL declare `ARG VERSION=local`
- **AND** Dockerfile SHALL declare `ARG VCS_REF=unknown`
- **AND** VERSION SHALL be passed to `make build` command

### Requirement: Binary Output Directory
The system SHALL generate Makefiles that output binaries to `bin/` directory.

#### Scenario: Build target output location
- **GIVEN** Makefile template is generated
- **WHEN** user runs `make build`
- **THEN** binary SHALL be created at `bin/{{PROJECT_NAME}}`
- **AND** `bin/` directory SHALL be created if it doesn't exist

#### Scenario: Clean target removes bin directory
- **GIVEN** Makefile template is generated
- **WHEN** user runs `make clean`
- **THEN** `bin/` directory and contents SHALL be removed

## MODIFIED Requirements

### Requirement: Go Version in Templates
The system SHALL generate templates using Go 1.25.5.

#### Scenario: go.mod specifies Go 1.25.5
- **GIVEN** go.mod template is generated
- **WHEN** examining go directive
- **THEN** version SHALL be `go 1.25.5`

#### Scenario: Dockerfile uses Go 1.25.5 builder
- **GIVEN** Dockerfile template is generated
- **WHEN** examining builder base image
- **THEN** image SHALL be `golang:1.25.5-trixie`

### Requirement: Linter Configuration Compatibility
The system SHALL generate golangci-lint configuration compatible with Go 1.25.

#### Scenario: golangci.yml specifies Go version
- **GIVEN** .golangci.yml template is generated
- **WHEN** examining run configuration
- **THEN** `run.go` SHALL be set to `"1.25"`
- **AND** `run.go-version` SHALL be set to `"1.25.5"`

#### Scenario: Go 1.25 compatible linters enabled
- **GIVEN** .golangci.yml template is generated
- **WHEN** examining enabled linters
- **THEN** `copyloopvar` linter SHALL be enabled
- **AND** deprecated linters SHALL be removed

## REMOVED Requirements

None.

## RENAMED Requirements

None.
