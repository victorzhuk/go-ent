## ADDED Requirements

### Requirement: ACP Protocol Transport

The system SHALL implement JSON-RPC 2.0 over stdio for ACP communication.

#### Scenario: Receive JSON-RPC request
- **WHEN** a valid JSON-RPC 2.0 request is received on stdin
- **THEN** the request is parsed and routed to appropriate handler
- **AND** a response is sent to stdout

#### Scenario: Handle notification (no response expected)
- **WHEN** a JSON-RPC notification (no id) is received
- **THEN** the message is processed
- **AND** no response is sent

#### Scenario: Protocol error
- **WHEN** an invalid JSON-RPC message is received
- **THEN** a JSON-RPC error response is returned with code -32700 (Parse error)

### Requirement: ACP Initialize Handshake

The system SHALL implement capability negotiation during initialization.

#### Scenario: Initialize with capabilities
- **WHEN** `acp/initialize` request is received with client capabilities
- **THEN** server responds with supported capabilities
- **AND** protocol version is validated
- **AND** session is ready for prompts

#### Scenario: Version mismatch
- **WHEN** client requests unsupported protocol version
- **THEN** error is returned with supported versions
- **AND** connection is not established

### Requirement: ACP Session Management

The system SHALL maintain stateful sessions for task execution.

#### Scenario: Create session
- **WHEN** `session/create` request is received
- **THEN** new session is created with unique ID
- **AND** session context is initialized
- **AND** session ID is returned to client

#### Scenario: Send prompt to session
- **WHEN** `session/prompt` request is received with task description
- **THEN** task is parsed and executed
- **AND** progress updates are streamed as notifications
- **AND** final result is returned when complete

#### Scenario: Cancel session
- **WHEN** `session/cancel` request is received
- **THEN** current execution is terminated gracefully
- **AND** partial results are preserved
- **AND** session is marked as cancelled

### Requirement: ACP Permission Flow

The system SHALL request permission from client before executing sensitive operations.

#### Scenario: Request tool permission
- **WHEN** worker needs to execute Write or Bash tool
- **THEN** `permission/request` notification is sent to client
- **AND** worker waits for `permission/response`
- **AND** tool is executed only if approved

#### Scenario: Permission denied
- **WHEN** client denies permission request
- **THEN** worker skips the operation
- **AND** execution continues with alternative approach or error

### Requirement: ACP Streaming Responses

The system SHALL stream progress and results during execution.

#### Scenario: Stream progress updates
- **WHEN** worker is executing a task
- **THEN** periodic `session/progress` notifications are sent
- **AND** notifications include current step and completion percentage

#### Scenario: Stream partial results
- **WHEN** worker generates intermediate output
- **THEN** `session/output` notifications are sent with content
- **AND** client can display streaming results

### Requirement: Worker Process Management

The system SHALL manage go-ent worker processes spawned via ACP.

#### Scenario: Start worker in ACP mode
- **WHEN** `go-ent acp` command is executed
- **THEN** process starts in ACP agent mode
- **AND** listens on stdin for JSON-RPC messages
- **AND** writes responses to stdout

#### Scenario: Worker with model override
- **WHEN** `go-ent acp --model haiku` is executed
- **THEN** worker uses Haiku model for task execution
- **AND** ignores default model selection

#### Scenario: Worker with tool restrictions
- **WHEN** `go-ent acp --allowed-tools Read,Grep,Glob` is executed
- **THEN** worker can only use specified tools
- **AND** other tool requests are denied

### Requirement: MCP Tools for ACP Worker Spawning

The system SHALL provide MCP tools for spawning and managing ACP workers from Claude Code.

#### Scenario: Spawn ACP worker
- **WHEN** `agent_spawn_acp` is called with task and model
- **THEN** go-ent subprocess is spawned in ACP mode
- **AND** worker ID is returned for tracking
- **AND** worker connects via stdio

#### Scenario: Send prompt to worker
- **WHEN** `agent_prompt_acp` is called with worker_id and prompt
- **THEN** prompt is sent to worker via ACP
- **AND** worker begins execution
- **AND** streaming results are returned

#### Scenario: Cancel worker
- **WHEN** `agent_cancel_acp` is called with worker_id
- **THEN** cancel request is sent via ACP
- **AND** worker terminates gracefully
- **AND** partial results are collected

### Requirement: Model Tiering for Workers

The system SHALL automatically select cost-effective models for worker tasks.

#### Scenario: Default to Haiku for simple tasks
- **WHEN** worker is spawned without explicit model
- **AND** task complexity is classified as simple
- **THEN** Haiku model is used

#### Scenario: Escalate to Sonnet for complex tasks
- **WHEN** task involves multiple files or complex logic
- **THEN** Sonnet model is used instead of Haiku

#### Scenario: Track cost per worker
- **WHEN** worker completes execution
- **THEN** token usage and cost are recorded
- **AND** aggregated in execution summary

### Requirement: Parallel Worker Coordination

The system SHALL coordinate multiple ACP workers for parallel execution.

#### Scenario: Spawn parallel workers
- **WHEN** execution engine has independent tasks
- **THEN** multiple ACP workers are spawned simultaneously
- **AND** each worker receives one task
- **AND** workers execute in parallel

#### Scenario: Collect parallel results
- **WHEN** all parallel workers complete
- **THEN** results are aggregated
- **AND** conflicts are detected
- **AND** summary is returned to orchestrator

#### Scenario: Handle worker failure
- **WHEN** one worker fails during parallel execution
- **THEN** other workers continue
- **AND** failure is reported with partial results
- **AND** orchestrator can retry failed task
