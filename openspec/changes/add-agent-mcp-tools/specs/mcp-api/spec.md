## ADDED Requirements

### Requirement: Tool Renaming for Cleaner API

The system SHALL rename all MCP tools from `go_ent_*` prefix to domain-specific prefixes.

#### Scenario: Spec tools renamed
- **WHEN** MCP server starts
- **THEN** tools use `spec_*` prefix instead of `go_ent_spec_*`
- **AND** `spec_init`, `spec_list`, `spec_show`, `spec_create`, `spec_update`, `spec_delete` are registered

#### Scenario: Registry tools renamed
- **WHEN** MCP server starts
- **THEN** tools use `registry_*` prefix instead of `go_ent_registry_*`
- **AND** `registry_list`, `registry_update`, `registry_sync`, `registry_next`, `registry_deps` are registered

#### Scenario: Workflow tools renamed
- **WHEN** MCP server starts
- **THEN** tools use `workflow_*` prefix instead of `go_ent_workflow_*`
- **AND** `workflow_start`, `workflow_status`, `workflow_approve` are registered

#### Scenario: Generation tools renamed
- **WHEN** MCP server starts
- **THEN** `go_ent_generate` is renamed to `project_generate`

#### Scenario: Backward compatibility
- **WHEN** tool renaming is complete
- **THEN** old tool names are NOT supported (breaking change v3.0)
- **AND** migration guide documents the changes

### Requirement: Agent Execution Tools

The system SHALL provide MCP tools for executing tasks via agents.

#### Scenario: agent_execute
- **WHEN** `agent_execute` is called with `task: "implement feature"`, `agent: "dev"`
- **THEN** agent is selected (or auto-selected if not specified)
- **AND** task is executed via selected agent
- **AND** execution_id is returned for tracking

#### Scenario: agent_execute with auto-selection
- **WHEN** `agent_execute` is called without `agent` parameter
- **THEN** system analyzes task complexity
- **AND** appropriate agent is auto-selected
- **AND** selection rationale is included in response

#### Scenario: agent_status
- **WHEN** `agent_status` is called with `execution_id`
- **THEN** current status is returned (pending, running, completed, failed)
- **AND** progress information is included if available

#### Scenario: agent_list
- **WHEN** `agent_list` is called
- **THEN** all available agents are returned
- **AND** each agent includes: name, role, capabilities, model

### Requirement: Skill Management Tools

The system SHALL provide MCP tools for managing and discovering skills.

#### Scenario: skill_list
- **WHEN** `skill_list` is called
- **THEN** all registered skills are returned
- **AND** each skill includes: name, description, auto-activation rules

#### Scenario: skill_list with filter
- **WHEN** `skill_list` is called with `context: "go-api"`
- **THEN** only skills matching the context are returned

#### Scenario: skill_info
- **WHEN** `skill_info` is called with `name: "go-code"`
- **THEN** detailed skill information is returned
- **AND** full skill prompt is included

### Requirement: Runtime Management Tools

The system SHALL provide MCP tools for runtime discovery and status.

#### Scenario: runtime_list
- **WHEN** `runtime_list` is called
- **THEN** all configured runtimes are returned (Claude Code, OpenCode, CLI)
- **AND** each runtime includes availability status

#### Scenario: runtime_status
- **WHEN** `runtime_status` is called with `runtime: "opencode"`
- **THEN** runtime health status is returned
- **AND** version information is included

### Requirement: Agent Delegation

The system SHALL support delegating tasks between agents.

#### Scenario: agent_delegate
- **WHEN** `agent_delegate` is called with `from_agent: "lead"`, `to_agent: "dev"`, `task: "..."`
- **THEN** task is delegated to target agent
- **AND** delegation chain is tracked
- **AND** execution_id is returned

#### Scenario: delegation rules validation
- **WHEN** delegation violates rules (e.g., dev cannot delegate to architect)
- **THEN** error is returned explaining the constraint

### Requirement: Tool Schema Consistency

The system SHALL maintain consistent schema patterns across all MCP tools.

#### Scenario: Standard parameters
- **WHEN** any tool is called
- **THEN** common parameters follow conventions:
  - `path` for project directory
  - `id` for resource identifiers
  - `name` for lookups
  - `status` for filtering

#### Scenario: Standard responses
- **WHEN** tool returns data
- **THEN** responses use consistent formats:
  - Lists return array of objects
  - Single items return object
  - Errors include actionable messages
