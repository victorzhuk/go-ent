## ADDED Requirements

### Requirement: CLI Mode Detection

The system SHALL detect whether it's being invoked in CLI mode or MCP mode.

#### Scenario: MCP mode via stdin
- **WHEN** go-ent is invoked without arguments
- **AND** stdin is connected to pipe
- **THEN** MCP server mode is activated
- **AND** JSON-RPC protocol is used

#### Scenario: CLI mode with arguments
- **WHEN** go-ent is invoked with subcommand (e.g., `go-ent run`)
- **THEN** CLI mode is activated
- **AND** subcommand is executed directly

#### Scenario: Help text available
- **WHEN** `go-ent --help` is invoked
- **THEN** CLI command overview is displayed
- **AND** MCP mode is not started

### Requirement: Run Command

The system SHALL execute tasks via CLI with agent selection.

#### Scenario: Basic task execution
- **WHEN** `go-ent run "add user endpoint"` is invoked
- **THEN** task is analyzed and complexity is determined
- **AND** appropriate agent is selected
- **AND** task executes with progress output

#### Scenario: Explicit agent selection
- **WHEN** `go-ent run --agent dev "implement feature"` is invoked
- **THEN** Dev agent is used regardless of complexity
- **AND** agent auto-selection is bypassed

#### Scenario: Strategy selection
- **WHEN** `go-ent run --strategy parallel "task1,task2,task3"` is invoked
- **THEN** tasks are executed in parallel
- **AND** results are aggregated at the end

#### Scenario: Budget limit
- **WHEN** `go-ent run --budget 0.50 "complex task"` is invoked
- **THEN** execution stops if budget would be exceeded
- **AND** cost estimate is shown before proceeding

#### Scenario: Dry run mode
- **WHEN** `go-ent run --dry-run "task"` is invoked
- **THEN** agent is selected and plan is shown
- **AND** no actual execution occurs

### Requirement: Status Command

The system SHALL show execution status for running and completed tasks.

#### Scenario: Show current status
- **WHEN** `go-ent status` is invoked
- **THEN** active executions are listed
- **AND** each shows: execution_id, agent, status, progress

#### Scenario: Show specific execution
- **WHEN** `go-ent status <execution-id>` is invoked
- **THEN** detailed status is shown
- **AND** output log is included

#### Scenario: No active executions
- **WHEN** `go-ent status` is invoked with no active tasks
- **THEN** message "No active executions" is shown

### Requirement: Agent Commands

The system SHALL provide CLI commands for agent discovery and information.

#### Scenario: List agents
- **WHEN** `go-ent agent list` is invoked
- **THEN** all available agents are displayed
- **AND** each shows: name, role, model, capabilities

#### Scenario: Agent details
- **WHEN** `go-ent agent info dev` is invoked
- **THEN** detailed Dev agent information is shown
- **AND** includes: role description, typical tasks, model tier

### Requirement: Skill Commands

The system SHALL provide CLI commands for skill discovery.

#### Scenario: List skills
- **WHEN** `go-ent skill list` is invoked
- **THEN** all registered skills are displayed
- **AND** each shows: name, description, auto-activation rules

#### Scenario: Skill details
- **WHEN** `go-ent skill info go-code` is invoked
- **THEN** full skill prompt is displayed
- **AND** activation context rules are shown

### Requirement: Spec Management Commands

The system SHALL provide CLI commands for OpenSpec management.

#### Scenario: Initialize OpenSpec
- **WHEN** `go-ent spec init` is invoked
- **THEN** openspec folder structure is created
- **AND** project.yaml is initialized

#### Scenario: List specs
- **WHEN** `go-ent spec list` is invoked
- **THEN** all specs are listed
- **AND** each shows: ID, description, path

#### Scenario: Show spec
- **WHEN** `go-ent spec show config-system` is invoked
- **THEN** spec content is displayed

#### Scenario: List changes
- **WHEN** `go-ent spec list --type change` is invoked
- **THEN** active changes are listed
- **AND** each shows: ID, status, description

### Requirement: Config Commands

The system SHALL provide CLI commands for configuration management.

#### Scenario: Show config
- **WHEN** `go-ent config show` is invoked
- **THEN** current configuration is displayed in YAML format

#### Scenario: Set config value
- **WHEN** `go-ent config set agent.default dev` is invoked
- **THEN** configuration value is updated
- **AND** confirmation message is shown

#### Scenario: Initialize config
- **WHEN** `go-ent config init` is invoked
- **THEN** default .go-ent/config.yaml is created
- **AND** prompts for required values (project name, etc.)

### Requirement: Global Flags

The system SHALL support global flags across all CLI commands.

#### Scenario: Verbose output
- **WHEN** any command is invoked with `--verbose` flag
- **THEN** detailed logging is enabled
- **AND** debug information is shown

#### Scenario: Config file override
- **WHEN** any command is invoked with `--config /path/to/config.yaml`
- **THEN** specified config file is used instead of default

#### Scenario: JSON output
- **WHEN** any command is invoked with `--format json`
- **THEN** output is formatted as JSON
- **AND** suitable for programmatic parsing

### Requirement: Error Handling

The system SHALL provide clear error messages in CLI mode.

#### Scenario: Invalid command
- **WHEN** `go-ent invalid-command` is invoked
- **THEN** error message shows available commands
- **AND** exit code is non-zero

#### Scenario: Missing required argument
- **WHEN** required argument is missing
- **THEN** error explains what's required
- **AND** usage example is shown

#### Scenario: OpenSpec not initialized
- **WHEN** spec command is invoked without openspec folder
- **THEN** error suggests running `go-ent spec init`
- **AND** exit code indicates error
