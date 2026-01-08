# Tasks: Add ACP Agent Mode

## Dependencies
- Extends: add-execution-engine (runtime abstraction)
- Requires: add-background-agents (async spawning infrastructure)

## 1. ACP Protocol Core

- [ ] 1.1 Create `internal/acp/transport.go` - JSON-RPC 2.0 over stdio
- [ ] 1.2 Create `internal/acp/messages.go` - Request/Response/Notification types
- [ ] 1.3 Create `internal/acp/handler.go` - Message router and dispatcher
- [ ] 1.4 Implement initialize handshake with capability negotiation
- [ ] 1.5 Implement session lifecycle (create, update, close)

## 2. ACP Methods Implementation

- [ ] 2.1 Implement `acp/initialize` - Handshake and capability exchange
- [ ] 2.2 Implement `acp/authenticate` - Optional API key validation
- [ ] 2.3 Implement `session/create` - New conversation session
- [ ] 2.4 Implement `session/prompt` - Send task to worker
- [ ] 2.5 Implement `session/cancel` - Cancel running task
- [ ] 2.6 Implement `permission/request` - Ask client for tool approval
- [ ] 2.7 Implement streaming responses with progress updates

## 3. Worker Execution

- [ ] 3.1 Create `internal/acp/worker.go` - Worker process lifecycle
- [ ] 3.2 Implement task parsing from ACP prompt
- [ ] 3.3 Integrate with existing agent selector for model selection
- [ ] 3.4 Execute task using internal tools (Read/Write/Edit/Bash/AST)
- [ ] 3.5 Stream results back via ACP notifications
- [ ] 3.6 Handle errors and partial completion

## 4. CLI Entry Point

- [ ] 4.1 Add `go-ent acp` subcommand to start in ACP mode
- [ ] 4.2 Add `--model` flag for explicit model selection
- [ ] 4.3 Add `--allowed-tools` flag for permission scoping
- [ ] 4.4 Add `--timeout` flag for execution limits
- [ ] 4.5 Add `--context` flag for initial context injection

## 5. Integration with Execution Engine

- [ ] 5.1 Add ACP runner to execution engine
- [ ] 5.2 Implement worker pool management
- [ ] 5.3 Add result aggregation for parallel workers
- [ ] 5.4 Integrate cost tracking per worker
- [ ] 5.5 Add worker health monitoring

## 6. Client SDK (for Claude Code)

- [ ] 6.1 Create `internal/acp/client.go` - ACP client for spawning workers
- [ ] 6.2 Add MCP tool `agent_spawn_acp` - Spawn go-ent ACP worker
- [ ] 6.3 Add MCP tool `agent_prompt_acp` - Send prompt to ACP worker
- [ ] 6.4 Add MCP tool `agent_stream_acp` - Read streaming results
- [ ] 6.5 Add MCP tool `agent_cancel_acp` - Cancel ACP worker

## 7. Testing

- [ ] 7.1 Unit tests for ACP protocol handling
- [ ] 7.2 Integration tests for worker lifecycle
- [ ] 7.3 Test parallel worker execution
- [ ] 7.4 Test error handling and recovery
- [ ] 7.5 Benchmark: single vs parallel execution
