# Tasks: Add ACP Proxy Mode for OpenCode Worker Orchestration

## Dependencies
- Extends: add-execution-engine (runtime abstraction)
- Requires: add-background-agents (async spawning infrastructure)
- External: OpenCode installed on system (`opencode` binary)

## Relationship with add-background-agents

This proposal builds ON TOP of add-background-agents:

| Tool Type | Proposal | Use Case |
|-----------|----------|----------|
| `go_ent_agent_*` | add-background-agents | Lightweight internal agents via direct API |
| `worker_*` | **this proposal** | Heavy OpenCode workers via ACP/CLI |

Internal agents (Haiku) → quick exploration, analysis, validation
OpenCode workers (GLM/Kimi) → bulk implementation, multi-file changes

## 1. Worker Manager Core

- [ ] 1.1 Create `internal/worker/manager.go` - Worker lifecycle management
- [ ] 1.2 Create `internal/worker/worker.go` - OpenCode worker abstraction
- [ ] 1.3 Create `internal/worker/pool.go` - Worker pool with concurrency limits
- [ ] 1.4 Create `internal/worker/config.go` - Load provider configs
- [ ] 1.5 Implement worker health monitoring and timeout handling

## 2. OpenCode ACP Communication

- [ ] 2.1 Create `internal/opencode/acp.go` - ACP client for OpenCode
- [ ] 2.2 Implement JSON-RPC 2.0 transport over stdio
- [ ] 2.3 Implement `acp/initialize` handshake with OpenCode
- [ ] 2.4 Implement `session/create` for new tasks
- [ ] 2.5 Implement `session/prompt` to send work to OpenCode
- [ ] 2.6 Implement streaming response handling
- [ ] 2.7 Implement `session/cancel` for graceful termination

## 3. OpenCode CLI Communication

- [ ] 3.1 Create `internal/opencode/cli.go` - CLI wrapper for OpenCode
- [ ] 3.2 Implement `opencode -p "prompt" -f json` execution
- [ ] 3.3 Parse JSON output from CLI mode
- [ ] 3.4 Handle CLI errors and timeouts
- [ ] 3.5 Support custom config files via `--config` flag

## 4. Direct Provider API (for simple tasks)

- [ ] 4.1 Create `internal/provider/anthropic.go` - Direct Anthropic API
- [ ] 4.2 Create `internal/provider/openai_compat.go` - OpenAI-compatible APIs
- [ ] 4.3 Implement streaming responses from direct API
- [ ] 4.4 Add rate limiting and retry logic

## 5. Task Router

- [ ] 5.1 Create `internal/router/router.go` - Task-to-provider routing
- [ ] 5.2 Create `internal/router/rules.go` - Routing rule definitions
- [ ] 5.3 Load routing rules from `.goent/routing.yaml`
- [ ] 5.4 Implement complexity-based routing (simple → CLI, complex → ACP)
- [ ] 5.5 Implement context-size routing (large → Kimi 128K)
- [ ] 5.6 Implement cost-based routing with budget constraints
- [ ] 5.7 Add manual provider override capability

## 6. Provider Configuration

- [ ] 6.1 Create `internal/config/providers.go` - Provider config loader
- [ ] 6.2 Support multiple OpenCode config files (per provider)
- [ ] 6.3 Validate provider connectivity on startup
- [ ] 6.4 Support environment variable substitution in configs
- [ ] 6.5 Add provider cost tracking configuration

## 7. MCP Tools for Claude Code

- [ ] 7.1 Add MCP tool `worker_spawn` - Spawn OpenCode worker
  - Parameters: provider, task, method (acp/cli/api)
  - Returns: worker_id

- [ ] 7.2 Add MCP tool `worker_prompt` - Send prompt to ACP worker
  - Parameters: worker_id, prompt, context_files
  - Returns: streaming results

- [ ] 7.3 Add MCP tool `worker_status` - Check worker status
  - Parameters: worker_id
  - Returns: status, progress, current_step

- [ ] 7.4 Add MCP tool `worker_output` - Get worker output
  - Parameters: worker_id, since_last
  - Returns: output text

- [ ] 7.5 Add MCP tool `worker_cancel` - Cancel worker
  - Parameters: worker_id
  - Returns: partial_results

- [ ] 7.6 Add MCP tool `worker_list` - List active workers
  - Returns: workers with status

- [ ] 7.7 Add MCP tool `provider_list` - List configured providers
  - Returns: providers with capabilities

- [ ] 7.8 Add MCP tool `provider_recommend` - Get optimal provider for task
  - Parameters: task_description, context_size
  - Returns: provider, method, rationale

## 8. Result Aggregation

- [ ] 8.1 Create `internal/aggregator/aggregator.go` - Collect parallel results
- [ ] 8.2 Implement conflict detection (multiple workers editing same file)
- [ ] 8.3 Implement result merging strategy
- [ ] 8.4 Generate execution summary with per-provider stats
- [ ] 8.5 Track cost per worker per provider

## 9. Integration with Existing Systems

- [ ] 9.1 Integrate with OpenSpec registry for task tracking
- [ ] 9.2 Update task status in tasks.md after worker completion
- [ ] 9.3 Integrate with context memory for pattern learning
- [ ] 9.4 Add hooks for pre/post worker execution

## 10. Testing

- [ ] 10.1 Unit tests for worker manager
- [ ] 10.2 Unit tests for ACP communication
- [ ] 10.3 Unit tests for CLI communication
- [ ] 10.4 Integration tests for parallel workers
- [ ] 10.5 Test provider failover scenarios
- [ ] 10.6 Benchmark: ACP vs CLI vs API performance
- [ ] 10.7 Test with actual OpenCode + GLM/Kimi providers
