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

### Requirement: Multi-Provider Backend Support

The system SHALL support multiple AI providers for worker execution.

#### Scenario: Configure provider
- **WHEN** providers.yaml contains provider configuration
- **THEN** provider is registered with API key and base URL
- **AND** provider models are available for selection

#### Scenario: Supported providers
- **WHEN** worker requests a provider
- **THEN** system supports: Anthropic (Haiku/Sonnet/Opus), Z.AI (GLM 4.7), Moonshot (Kimi K2), DeepSeek, Alibaba (Qwen3)

#### Scenario: Provider health check
- **WHEN** provider is selected for task
- **THEN** health check is performed before spawning
- **AND** unhealthy providers are skipped

#### Scenario: Provider failover
- **WHEN** primary provider fails or is rate-limited
- **THEN** task is retried with fallback provider
- **AND** failure is logged for monitoring

### Requirement: Provider-Aware Task Routing

The system SHALL route tasks to optimal providers based on task characteristics.

#### Scenario: Route by complexity
- **WHEN** task complexity is simple (lint, format)
- **THEN** cheap fast provider is selected (Haiku, GLM 4.7)

#### Scenario: Route by context size
- **WHEN** task requires large context (>50K tokens)
- **THEN** long-context provider is selected (Kimi K2 - 128K)

#### Scenario: Route by task type
- **WHEN** task type is code-heavy refactoring
- **THEN** code-optimized provider is selected (DeepSeek)

#### Scenario: Explicit provider override
- **WHEN** `--provider` flag is specified
- **THEN** specified provider is used regardless of routing rules

#### Scenario: Cost-based routing
- **WHEN** budget constraint is active
- **THEN** cheapest capable provider is selected

### Requirement: Model Tiering for Workers

The system SHALL automatically select cost-effective models for worker tasks.

#### Scenario: Default to cheap model for bulk tasks
- **WHEN** worker is spawned without explicit model
- **AND** task is bulk implementation
- **THEN** GLM 4.7 or Haiku is used based on availability

#### Scenario: Escalate to Sonnet for complex tasks
- **WHEN** task involves multiple files or complex logic
- **THEN** Sonnet model is used

#### Scenario: Use Opus only for orchestrator
- **WHEN** task is research, planning, or review
- **THEN** task remains with Claude Code orchestrator (Opus)
- **AND** not delegated to worker

#### Scenario: Track cost per worker per provider
- **WHEN** worker completes execution
- **THEN** token usage and cost are recorded per provider
- **AND** aggregated in execution summary with provider breakdown

### Requirement: Parallel Worker Coordination

The system SHALL coordinate multiple ACP workers for parallel execution.

#### Scenario: Spawn parallel workers
- **WHEN** execution engine has independent tasks
- **THEN** multiple ACP workers are spawned simultaneously
- **AND** each worker receives one task
- **AND** workers execute in parallel

#### Scenario: Heterogeneous swarm
- **WHEN** multiple tasks with different characteristics exist
- **THEN** workers with different providers are spawned
- **AND** example: Task 1 → GLM 4.7, Task 2 → Kimi K2, Task 3 → Haiku
- **AND** all workers execute in parallel

#### Scenario: Collect parallel results
- **WHEN** all parallel workers complete
- **THEN** results are aggregated
- **AND** conflicts are detected
- **AND** summary is returned to orchestrator
- **AND** per-provider statistics included

#### Scenario: Handle worker failure
- **WHEN** one worker fails during parallel execution
- **THEN** other workers continue
- **AND** failure is reported with partial results
- **AND** orchestrator can retry failed task with different provider

### Requirement: MCP Provider Management Tools

The system SHALL provide MCP tools for managing providers from Claude Code.

#### Scenario: List providers
- **WHEN** `provider_list` is called
- **THEN** all configured providers are returned
- **AND** each provider includes: name, models, status, cost info

#### Scenario: Check provider status
- **WHEN** `provider_status` is called with provider name
- **THEN** health check is performed
- **AND** rate limit status is returned
- **AND** recent error count is included

#### Scenario: Get provider recommendation
- **WHEN** `provider_recommend` is called with task description
- **THEN** optimal provider is recommended based on routing rules
- **AND** rationale is included in response
