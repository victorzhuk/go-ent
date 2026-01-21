## ADDED Requirements

### Requirement: go-ent as ACP Proxy

The system SHALL act as an ACP proxy between Claude Code and OpenCode workers (NOT as a worker itself).

#### Scenario: Proxy role
- **GIVEN** Claude Code delegates a task via MCP
- **WHEN** go-ent receives the task
- **THEN** go-ent spawns an OpenCode worker to execute the task
- **AND** go-ent does NOT execute the task itself

#### Scenario: Worker management
- **WHEN** go-ent spawns an OpenCode worker
- **THEN** go-ent manages the worker lifecycle
- **AND** go-ent collects results from the worker
- **AND** go-ent returns results to Claude Code

### Requirement: OpenCode ACP Communication

The system SHALL communicate with OpenCode workers via ACP protocol.

#### Scenario: Spawn ACP worker
- **WHEN** `worker_spawn` is called with `method: "acp"`, provider, and model
- **THEN** go-ent executes `opencode acp` with `OPENCODE_CONFIG` env var
- **AND** go-ent establishes JSON-RPC 2.0 connection over stdin/stdout
- **AND** worker_id is returned for tracking

#### Scenario: ACP initialize handshake
- **WHEN** OpenCode subprocess starts
- **THEN** go-ent sends `initialize` request (JSON-RPC method)
- **AND** capabilities are negotiated (protocol version, features)
- **AND** go-ent sends `authenticate` if required
- **AND** go-ent sends `session/new` with provider/model selection
- **AND** session is ready for `session/prompt` requests

#### Scenario: Send prompt via ACP
- **WHEN** `worker_prompt` is called with worker_id and prompt
- **THEN** go-ent sends `session/prompt` to OpenCode
- **AND** OpenCode executes the task with its configured AI provider
- **AND** streaming responses are forwarded to Claude Code

#### Scenario: Cancel ACP worker
- **WHEN** `worker_cancel` is called with worker_id
- **THEN** go-ent sends `session/cancel` to OpenCode
- **AND** OpenCode terminates gracefully
- **AND** partial results are collected

### Requirement: OpenCode CLI Communication

The system SHALL support CLI communication for quick one-shot tasks.

#### Scenario: Execute via CLI
- **WHEN** `worker_spawn` is called with `method: "cli"`, provider, and model
- **THEN** go-ent executes `opencode run --model <provider/model> --prompt "<prompt>"`
- **AND** sets `OPENCODE_CONFIG` environment variable
- **AND** waits for completion
- **AND** parses output

#### Scenario: CLI with config
- **WHEN** OpenCode CLI is executed
- **THEN** `OPENCODE_CONFIG` env var points to config file path
- **AND** provider/model is selected via `--model provider/model` flag
- **AND** OpenCode uses the specified AI provider and model

#### Scenario: CLI error handling
- **WHEN** OpenCode CLI returns non-zero exit code
- **THEN** error is captured
- **AND** partial output is preserved
- **AND** error is reported to Claude Code

### Requirement: Direct Provider API

The system SHALL support direct API calls for simple tasks (bypassing OpenCode).

#### Scenario: Direct API for Anthropic
- **WHEN** `worker_spawn` is called with `method: "api"`, `provider: "haiku"`
- **THEN** go-ent makes direct API call to Anthropic
- **AND** no OpenCode process is spawned
- **AND** response is returned immediately

#### Scenario: Direct API for OpenAI-compatible
- **WHEN** provider uses OpenAI-compatible API (GLM, Kimi, DeepSeek)
- **THEN** go-ent can make direct API calls
- **AND** `base_url` from provider config is used

### Requirement: OpenCode Configuration with Multiple Providers

The system SHALL use a single OpenCode configuration file with multiple providers.

#### Scenario: Single config with multiple providers
- **GIVEN** `~/.config/opencode/opencode.json` contains multiple provider definitions
- **WHEN** worker is spawned with specific provider/model
- **THEN** `OPENCODE_CONFIG` env var points to the config file
- **AND** provider/model is selected via `--model` flag (CLI) or session config (ACP)

#### Scenario: GLM provider selection
- **GIVEN** `opencode.json` has `moonshot/glm-4` configured
- **WHEN** provider "moonshot" and model "glm-4" are selected
- **THEN** OpenCode uses GLM 4.7 for task execution

#### Scenario: Kimi provider selection
- **GIVEN** `opencode.json` has `moonshot/kimi-k2` configured
- **WHEN** provider "moonshot" and model "kimi-k2" are selected
- **THEN** OpenCode uses Kimi K2 (128K context)
- **AND** large context tasks can be handled

#### Scenario: DeepSeek provider selection
- **GIVEN** `opencode.json` has `deepseek/deepseek-coder` configured
- **WHEN** provider "deepseek" and model "deepseek-coder" are selected
- **THEN** OpenCode uses DeepSeek Coder for code-heavy tasks

### Requirement: Task Routing

The system SHALL route tasks to optimal provider based on task characteristics.

#### Scenario: Route by complexity
- **WHEN** task is simple (lint, format, trivial fix)
- **THEN** router selects `method: "cli"` or `method: "api"`
- **AND** cheap fast provider is selected (Haiku, GLM)

#### Scenario: Route by context size
- **WHEN** task requires context > 50K tokens
- **THEN** router selects Kimi K2 provider (128K context)
- **AND** `method: "acp"` is used for streaming

#### Scenario: Route by task type
- **WHEN** task is code-heavy refactoring
- **THEN** router selects DeepSeek or Sonnet
- **AND** `method: "acp"` is used for complex execution

#### Scenario: Explicit provider override
- **WHEN** `provider` parameter is explicitly specified
- **THEN** specified provider is used regardless of routing rules

#### Scenario: Route based on rules file
- **GIVEN** `.goent/routing.yaml` contains routing rules
- **WHEN** task is routed
- **THEN** rules are evaluated in order
- **AND** first matching rule determines provider and method

### Requirement: MCP Tools for Worker Management

The system SHALL expose MCP tools for Claude Code to manage OpenCode workers.

#### Scenario: worker_spawn
- **WHEN** `worker_spawn` is called with provider and task
- **THEN** OpenCode worker is spawned (ACP, CLI, or API based on method)
- **AND** worker_id is returned

#### Scenario: worker_prompt
- **WHEN** `worker_prompt` is called with worker_id and prompt
- **THEN** prompt is sent to the ACP worker
- **AND** streaming results are returned

#### Scenario: worker_status
- **WHEN** `worker_status` is called with worker_id
- **THEN** current status is returned (running, completed, failed)
- **AND** progress percentage and current step are included

#### Scenario: worker_output
- **WHEN** `worker_output` is called with worker_id
- **THEN** accumulated output from worker is returned
- **AND** `since_last` flag returns only new output

#### Scenario: worker_cancel
- **WHEN** `worker_cancel` is called with worker_id
- **THEN** worker is terminated gracefully
- **AND** partial results are preserved

#### Scenario: worker_list
- **WHEN** `worker_list` is called
- **THEN** all active workers are returned with status

#### Scenario: provider_list
- **WHEN** `provider_list` is called
- **THEN** all configured providers are returned
- **AND** each includes: name, method, capabilities, cost estimate

#### Scenario: provider_recommend
- **WHEN** `provider_recommend` is called with task description
- **THEN** optimal provider is recommended
- **AND** rationale explains the recommendation

### Requirement: Parallel Worker Execution

The system SHALL support multiple OpenCode workers running in parallel.

#### Scenario: Spawn parallel workers
- **WHEN** Claude Code spawns multiple workers
- **THEN** workers execute simultaneously
- **AND** each worker is independent

#### Scenario: Heterogeneous swarm
- **WHEN** multiple tasks with different characteristics exist
- **THEN** workers with different providers are spawned
- **AND** example: Worker 1 (GLM), Worker 2 (Kimi), Worker 3 (DeepSeek)

#### Scenario: Result aggregation
- **WHEN** all workers complete
- **THEN** go-ent aggregates results
- **AND** conflicts are detected (same file edited by multiple workers)
- **AND** summary is returned to Claude Code

#### Scenario: Worker failure handling
- **WHEN** one worker fails
- **THEN** other workers continue
- **AND** failure is reported to Claude Code
- **AND** Claude Code can retry with different provider

### Requirement: Cost Tracking

The system SHALL track costs per worker per provider.

#### Scenario: Track worker cost
- **WHEN** worker completes execution
- **THEN** token usage is recorded
- **AND** cost is calculated based on provider pricing

#### Scenario: Cost aggregation
- **WHEN** execution with multiple workers completes
- **THEN** total cost is calculated
- **AND** breakdown by provider is included
- **AND** cost is returned to Claude Code

#### Scenario: Budget enforcement
- **WHEN** budget limit is configured
- **AND** execution would exceed budget
- **THEN** warning is returned
- **AND** cheaper provider is suggested

### Requirement: Provider Health and Failover

The system SHALL handle provider failures gracefully.

#### Scenario: Provider health check
- **WHEN** provider is selected
- **THEN** optional health check can be performed
- **AND** unhealthy providers are skipped

#### Scenario: Provider failover
- **WHEN** OpenCode worker fails due to provider issue (rate limit, timeout)
- **THEN** go-ent can retry with fallback provider
- **AND** failure is logged

#### Scenario: Rate limit awareness
- **WHEN** provider returns rate limit error
- **THEN** go-ent tracks rate limit status
- **AND** routes subsequent tasks to other providers
