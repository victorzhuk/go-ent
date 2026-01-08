# Tasks: Add ACP Agent Mode with Multi-Provider Workers

## Dependencies
- Extends: add-execution-engine (runtime abstraction)
- Requires: add-background-agents (async spawning infrastructure)

## 1. Multi-Provider Backend System

- [ ] 1.1 Create `internal/provider/provider.go` - Provider interface definition
- [ ] 1.2 Create `internal/provider/anthropic.go` - Anthropic (Haiku/Sonnet/Opus)
- [ ] 1.3 Create `internal/provider/openai.go` - OpenAI compatible (GLM, Kimi, DeepSeek)
- [ ] 1.4 Create `internal/provider/config.go` - Load providers.yaml configuration
- [ ] 1.5 Create `internal/provider/registry.go` - Provider registry and discovery
- [ ] 1.6 Implement provider health checks and failover

## 2. ACP Protocol Core

- [ ] 2.1 Create `internal/acp/transport.go` - JSON-RPC 2.0 over stdio
- [ ] 2.2 Create `internal/acp/messages.go` - Request/Response/Notification types
- [ ] 2.3 Create `internal/acp/handler.go` - Message router and dispatcher
- [ ] 2.4 Implement initialize handshake with capability + provider negotiation
- [ ] 2.5 Implement session lifecycle (create, update, close)

## 3. ACP Methods Implementation

- [ ] 3.1 Implement `acp/initialize` - Handshake with provider info
- [ ] 3.2 Implement `acp/authenticate` - API key validation per provider
- [ ] 3.3 Implement `session/create` - New conversation session
- [ ] 3.4 Implement `session/prompt` - Send task to worker
- [ ] 3.5 Implement `session/cancel` - Cancel running task
- [ ] 3.6 Implement `permission/request` - Ask client for tool approval
- [ ] 3.7 Implement streaming responses with progress updates

## 4. Worker Execution

- [ ] 4.1 Create `internal/acp/worker.go` - Worker process lifecycle
- [ ] 4.2 Implement task parsing from ACP prompt
- [ ] 4.3 Connect worker to configured provider backend
- [ ] 4.4 Execute task using internal tools (Read/Write/Edit/Bash/AST)
- [ ] 4.5 Stream results back via ACP notifications
- [ ] 4.6 Handle provider-specific errors and rate limits

## 5. Provider-Aware Task Routing

- [ ] 5.1 Create `internal/routing/router.go` - Task-to-provider routing
- [ ] 5.2 Create `internal/routing/rules.go` - Routing rule definitions
- [ ] 5.3 Implement complexity-based routing (simple→Haiku, bulk→GLM)
- [ ] 5.4 Implement context-size routing (large files→Kimi K2)
- [ ] 5.5 Implement cost-based routing with budget constraints
- [ ] 5.6 Add manual override capability

## 6. CLI Entry Point

- [ ] 6.1 Add `go-ent acp` subcommand to start in ACP mode
- [ ] 6.2 Add `--provider` flag for explicit provider selection
- [ ] 6.3 Add `--model` flag for explicit model selection
- [ ] 6.4 Add `--allowed-tools` flag for permission scoping
- [ ] 6.5 Add `--timeout` flag for execution limits
- [ ] 6.6 Add `--context` flag for initial context injection

## 7. Worker Pool Management

- [ ] 7.1 Create `internal/pool/pool.go` - Heterogeneous worker pool
- [ ] 7.2 Implement spawn with provider selection
- [ ] 7.3 Add result aggregation for parallel workers
- [ ] 7.4 Integrate cost tracking per worker per provider
- [ ] 7.5 Add worker health monitoring
- [ ] 7.6 Implement rate limit awareness per provider

## 8. MCP Tools (for Claude Code Orchestrator)

- [ ] 8.1 Create `internal/acp/client.go` - ACP client for spawning workers
- [ ] 8.2 Add MCP tool `worker_spawn` - Spawn worker with provider/model
- [ ] 8.3 Add MCP tool `worker_prompt` - Send prompt to worker
- [ ] 8.4 Add MCP tool `worker_status` - Check worker status
- [ ] 8.5 Add MCP tool `worker_output` - Read streaming results
- [ ] 8.6 Add MCP tool `worker_cancel` - Cancel worker
- [ ] 8.7 Add MCP tool `worker_list` - List active workers
- [ ] 8.8 Add MCP tool `provider_list` - List available providers
- [ ] 8.9 Add MCP tool `provider_status` - Check provider health

## 9. Testing

- [ ] 9.1 Unit tests for provider abstraction
- [ ] 9.2 Unit tests for ACP protocol handling
- [ ] 9.3 Integration tests for worker lifecycle
- [ ] 9.4 Test multi-provider parallel execution
- [ ] 9.5 Test provider failover
- [ ] 9.6 Test rate limit handling
- [ ] 9.7 Benchmark: provider performance comparison
