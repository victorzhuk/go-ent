# MCP Tools Spec Delta

## MODIFIED Requirements

### Requirement: MCP Server Naming
The MCP server SHALL be named `go-ent` to match the project repository name.

**Previous**: MCP server was named `goent-spec`
**Reason for Change**: Unify naming with repository and remove redundant `-spec` suffix

#### Scenario: MCP server reports correct name
- **WHEN** MCP server initialization occurs
- **THEN** the server SHALL report name as "go-ent"
- **AND** the server version SHALL match the binary version

### Requirement: MCP Tool Naming Convention
All MCP tools SHALL use the `go_ent_` prefix (with underscore separator) for consistency with MCP protocol conventions.

**Previous**: Tools used `go_ent_` prefix (without hyphen)
**Reason for Change**: Align tool namespacing with unified `go-ent` project naming

#### Scenario: Tool names use correct prefix
- **WHEN** MCP tools are registered
- **THEN** all tool names SHALL start with `go_ent_`
- **AND** no tool names SHALL use `go_ent_` prefix

#### Scenario: Spec management tools renamed
- **GIVEN** the previous tool names:
  - `go_ent_spec_init` → `go_ent_spec_init`
  - `go_ent_spec_list` → `go_ent_spec_list`
  - `go_ent_spec_show` → `go_ent_spec_show`
  - `go_ent_spec_create` → `go_ent_spec_create`
  - `go_ent_spec_update` → `go_ent_spec_update`
  - `go_ent_spec_delete` → `go_ent_spec_delete`
  - `go_ent_spec_validate` → `go_ent_spec_validate`
  - `go_ent_spec_archive` → `go_ent_spec_archive`
- **WHEN** a client calls any tool
- **THEN** the tool SHALL be accessible by its new `go_ent_` prefixed name
- **AND** old `go_ent_` names SHALL NOT be registered
