## ADDED Requirements

### Requirement: Background Agent Spawning

The system SHALL provide an MCP tool to spawn agents that execute asynchronously in the background.

#### Scenario: Spawn background exploration agent
- **WHEN** `go_ent_agent_spawn` is called with `task: "explore authentication patterns"`, `agent_type: "explore"`, `background: true`
- **THEN** the agent is spawned asynchronously
- **AND** an `agent_id` is returned immediately
- **AND** the calling agent continues execution without blocking

#### Scenario: Spawn with model override
- **WHEN** `go_ent_agent_spawn` is called with `model: "haiku"` parameter
- **THEN** the background agent uses the specified model
- **AND** cost-optimized model is used for background research

#### Scenario: Spawn with timeout
- **WHEN** `go_ent_agent_spawn` is called with `timeout_ms: 300000`
- **THEN** the agent automatically terminates after 5 minutes if not completed

### Requirement: Background Agent Status Monitoring

The system SHALL provide an MCP tool to check the status of running background agents.

#### Scenario: Check agent status
- **WHEN** `go_ent_agent_status` is called with `agent_id: "abc123"`
- **THEN** the current status is returned (running, completed, failed, killed)
- **AND** progress information is included if available

#### Scenario: Agent completed
- **WHEN** agent has finished execution
- **THEN** status shows `completed`
- **AND** final output is available via `go_ent_agent_output`

#### Scenario: Agent failed
- **WHEN** agent encountered an error
- **THEN** status shows `failed`
- **AND** error details are included in response

### Requirement: Background Agent Output Retrieval

The system SHALL provide an MCP tool to retrieve output from background agents.

#### Scenario: Get full output
- **WHEN** `go_ent_agent_output` is called with `agent_id: "abc123"`
- **THEN** all output from the agent is returned

#### Scenario: Get filtered output
- **WHEN** `go_ent_agent_output` is called with `filter: "error|warning"`
- **THEN** only lines matching the regex pattern are returned

#### Scenario: Get incremental output
- **WHEN** `go_ent_agent_output` is called with `since_last: true`
- **THEN** only new output since last retrieval is returned

### Requirement: Background Agent Termination

The system SHALL provide an MCP tool to terminate running background agents.

#### Scenario: Kill running agent
- **WHEN** `go_ent_agent_kill` is called with `agent_id: "abc123"`
- **THEN** the agent is terminated gracefully
- **AND** status is updated to `killed`
- **AND** partial output remains available

#### Scenario: Kill non-existent agent
- **WHEN** `go_ent_agent_kill` is called with invalid agent_id
- **THEN** an error is returned indicating agent not found

### Requirement: Background Agent Registry

The system SHALL maintain a registry of all spawned background agents.

#### Scenario: List all agents
- **WHEN** `go_ent_agent_list` is called
- **THEN** all background agents are returned with their status
- **AND** agents are sorted by spawn time (newest first)

#### Scenario: Filter by status
- **WHEN** `go_ent_agent_list` is called with `status: "running"`
- **THEN** only running agents are returned

#### Scenario: Automatic cleanup
- **WHEN** session terminates
- **THEN** all background agents are terminated gracefully
- **AND** resources are released

### Requirement: Model Tiering for Background Agents

The system SHALL route background agent tasks to cost-optimized models by default.

#### Scenario: Default to cheaper model
- **WHEN** background agent is spawned without explicit model
- **THEN** Haiku model is used for exploration/research tasks
- **AND** Sonnet is used for implementation tasks

#### Scenario: Explicit model override
- **WHEN** `model` parameter is specified
- **THEN** the specified model is used regardless of task type
