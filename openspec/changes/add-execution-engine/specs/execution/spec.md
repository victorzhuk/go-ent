## ADDED Requirements

### Requirement: Runner Interface

The system SHALL provide an abstracted Runner interface for executing tasks across different runtimes.

#### Scenario: Runner factory
- **WHEN** execution engine needs a runner for `runtime: "claude"`
- **THEN** ClaudeRunner instance is created
- **AND** runner is initialized with configuration

#### Scenario: Runtime selection
- **WHEN** task specifies `runtime: "opencode"`
- **THEN** OpenCodeRunner is selected
- **AND** task executes via OpenCode native API

#### Scenario: CLI runtime fallback
- **WHEN** no runtime is specified AND not in MCP mode
- **THEN** CLI runtime is used
- **AND** task executes standalone

### Requirement: Claude Code MCP Runner

The system SHALL execute tasks via Claude Code using MCP protocol.

#### Scenario: MCP prompt execution
- **WHEN** ClaudeRunner executes task
- **THEN** agent role prompt is built
- **AND** relevant skills are injected
- **AND** task description is included
- **AND** MCP message is sent to Claude Code

#### Scenario: Streaming responses
- **WHEN** Claude Code returns streaming response
- **THEN** output is captured incrementally
- **AND** progress is tracked

#### Scenario: MCP error handling
- **WHEN** MCP communication fails
- **THEN** error is captured with context
- **AND** partial output is preserved

### Requirement: OpenCode Native Runner

The system SHALL execute tasks via OpenCode native API (required for v3.0).

#### Scenario: OpenCode API authentication
- **WHEN** OpenCodeRunner initializes
- **THEN** API key from config is used
- **AND** connection is validated

#### Scenario: Native API execution
- **WHEN** task is executed via OpenCode
- **THEN** OpenCode API `/execute` endpoint is called
- **AND** task prompt and configuration are sent
- **AND** response is parsed from JSON

#### Scenario: OpenCode provider selection
- **WHEN** OpenCodeRunner executes with specific provider
- **THEN** provider config is passed to OpenCode
- **AND** OpenCode uses the specified AI backend (GLM, Kimi, etc.)

### Requirement: CLI Standalone Runner

The system SHALL execute tasks in standalone CLI mode.

#### Scenario: CLI execution
- **WHEN** CLIRunner executes task
- **THEN** prompt template is rendered
- **AND** task executes without MCP protocol
- **AND** output is written to stdout

#### Scenario: CLI with template rendering
- **WHEN** agent role requires specific prompt format
- **THEN** template is populated with task details
- **AND** rendered prompt is used for execution

### Requirement: Execution Strategies

The system SHALL support multiple execution strategies.

#### Scenario: Single strategy
- **WHEN** execution strategy is "single"
- **THEN** single agent executes the entire task
- **AND** result is returned directly

#### Scenario: Multi strategy (conversation)
- **WHEN** execution strategy is "multi"
- **THEN** first agent (e.g., Architect) designs
- **AND** second agent (e.g., Dev) implements
- **AND** agents share conversation context

#### Scenario: Parallel strategy
- **WHEN** execution strategy is "parallel"
- **AND** tasks have dependency graph
- **THEN** independent tasks execute simultaneously
- **AND** dependent tasks wait for prerequisites
- **AND** results are aggregated

#### Scenario: Strategy auto-selection
- **WHEN** no strategy specified AND task is complex
- **THEN** Multi strategy is selected automatically
- **AND** Architect designs before Dev implements

### Requirement: Dependency Graph Execution

The system SHALL execute tasks respecting dependencies in parallel strategy.

#### Scenario: Topological sort
- **WHEN** dependency graph is provided
- **THEN** tasks are sorted topologically
- **AND** execution order respects dependencies

#### Scenario: Parallel execution of independent tasks
- **WHEN** tasks T1, T2 have no dependencies
- **THEN** T1 and T2 execute in parallel
- **AND** both results are collected

#### Scenario: Wait for dependencies
- **WHEN** task T3 depends on T1, T2
- **THEN** T3 waits until T1 and T2 complete
- **AND** T3 receives outputs from T1 and T2 as context

#### Scenario: Dependency failure handling
- **WHEN** task T1 fails AND T3 depends on T1
- **THEN** T3 is skipped
- **AND** T3 status is marked as "skipped_dependency_failed"

### Requirement: Budget Tracking

The system SHALL track and enforce spending limits during execution.

#### Scenario: Track token usage
- **WHEN** task executes
- **THEN** input and output tokens are counted
- **AND** cost is calculated based on model pricing

#### Scenario: Budget limit enforcement
- **WHEN** execution would exceed budget
- **THEN** warning is shown before proceeding
- **AND** execution can be cancelled by user

#### Scenario: Cost accumulation
- **WHEN** multiple tasks execute in parallel
- **THEN** costs are accumulated across all tasks
- **AND** total cost is reported at end

#### Scenario: Per-model pricing
- **WHEN** calculating cost
- **THEN** model-specific pricing is used
  - Opus: $15/$75 per 1M tokens (in/out)
  - Sonnet: $3/$15 per 1M tokens
  - Haiku: $0.25/$1.25 per 1M tokens

### Requirement: Result Aggregation

The system SHALL collect and aggregate results from executions.

#### Scenario: Single execution result
- **WHEN** single strategy completes
- **THEN** result includes: output, status, token_usage, cost

#### Scenario: Multi execution result
- **WHEN** multi strategy completes
- **THEN** result includes outputs from all agents in chain
- **AND** each agent's contribution is preserved

#### Scenario: Parallel execution result
- **WHEN** parallel strategy completes
- **THEN** result includes outputs from all tasks
- **AND** dependency order is preserved in output

#### Scenario: Execution metadata
- **WHEN** any execution completes
- **THEN** metadata is included:
  - execution_id, agent(s) used, runtime(s), duration
  - total_tokens (input + output), total_cost
  - status (completed/failed/partial)

### Requirement: Error Recovery

The system SHALL handle execution errors gracefully.

#### Scenario: Runtime failure recovery
- **WHEN** runtime fails during execution
- **THEN** partial output is preserved
- **AND** error details are captured
- **AND** retry is possible with different runtime

#### Scenario: Timeout handling
- **WHEN** execution exceeds timeout
- **THEN** execution is terminated
- **AND** partial results are returned
- **AND** status is marked as "timeout"

#### Scenario: Graceful degradation
- **WHEN** OpenCode runtime is unavailable
- **THEN** fallback to Claude Code runtime is attempted
- **AND** user is notified of fallback

### Requirement: Execution Context Management

The system SHALL manage context across execution steps.

#### Scenario: Context preservation in multi-agent
- **WHEN** Architect completes design
- **THEN** design output is included in Dev's context
- **AND** Dev can reference Architect's decisions

#### Scenario: Context limits
- **WHEN** context exceeds model limits
- **THEN** context is summarized or truncated
- **AND** most relevant information is preserved

#### Scenario: File context injection
- **WHEN** task references specific files
- **THEN** file contents are included in context
- **AND** file paths are clearly marked
