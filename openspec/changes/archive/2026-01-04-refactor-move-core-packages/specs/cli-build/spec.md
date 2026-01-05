# cli-build Specification Delta

## MODIFIED Requirements

### Requirement: Go Module Structure
The CLI SHALL follow Go project layout best practices with internal packages properly organized at the project root.

**What changed**: Moved reusable internal packages from `/cmd/go-ent/internal/` to `/internal/` to follow golang-standards/project-layout conventions.

#### Scenario: Root module structure
- **WHEN** checking module structure
- **THEN** root `go.mod` exists with module `github.com/victorzhuk/go-ent`
- **AND** CLI entry point is at `cmd/go-ent/main.go`
- **AND** reusable internal packages are at `/internal/`
- **AND** CLI-specific code remains in `/cmd/go-ent/internal/`

#### Scenario: Internal packages at project root
- **WHEN** examining project structure
- **THEN** core domain logic packages exist at `/internal/spec/`
- **AND** template engine exists at `/internal/template/`
- **AND** code generation logic exists at `/internal/generation/`
- **AND** these packages have no MCP dependencies

#### Scenario: CLI-specific packages remain under cmd
- **WHEN** examining `/cmd/go-ent/internal/` structure
- **THEN** MCP tool handlers exist at `/cmd/go-ent/internal/tools/`
- **AND** MCP server factory exists at `/cmd/go-ent/internal/server/`
- **AND** these packages depend on MCP SDK

#### Scenario: Import paths follow new structure
- **WHEN** importing internal packages in tools
- **THEN** spec package imported as `github.com/victorzhuk/go-ent/internal/spec`
- **AND** template package imported as `github.com/victorzhuk/go-ent/internal/template`
- **AND** generation package imported as `github.com/victorzhuk/go-ent/internal/generation`
- **AND** all imports compile successfully
