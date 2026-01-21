# Tasks: Add ACP Proxy Mode for OpenCode Worker Orchestration

## 🚀 IN PROGRESS

**Implementation Started**: 2026-01-20

**Dependencies Resolved**:
1. **execution-engine-v2** - ✅ Runtime abstraction available (v1 features complete)
2. **add-background-agents** - ✅ Async spawning infrastructure complete

**Progress**: Phase 2 & 3 (ACP & CLI Communication) complete, 19/19 tasks complete, Phase 4: 4/4 tasks complete, Phase 5: 7/7 tasks complete, Phase 6: 6/6 tasks complete, Phase 7: 8/8 tasks complete, Phase 8: 5/5 tasks complete, Phase 9: 4/4 tasks complete

---

## Dependencies
- **Requires**: execution-engine-v2 (runtime abstraction) - ✅ COMPLETED (v1 features available)
- **Requires**: add-background-agents (async spawning infrastructure) - ✅ COMPLETED
- **External**: OpenCode installed on system (`opencode` binary)

## Relationship with add-background-agents

This proposal builds ON TOP of add-background-agents:

| Tool Type | Proposal | Use Case |
|-----------|----------|----------|
| `go_ent_agent_*` | add-background-agents | Lightweight internal agents via direct API |
| `worker_*` | **this proposal** | Heavy OpenCode workers via ACP/CLI |

Internal agents (Haiku) → quick exploration, analysis, validation
OpenCode workers (GLM/Kimi) → bulk implementation, multi-file changes

## Phase 3 Separated

Dynamic MCP Discovery features moved to **add-dynamic-mcp-discovery** proposal:
- `mcp_find`, `mcp_add`, `mcp_remove`, `mcp_active` tools
- Docker MCP Gateway integration
- MCP routing rules engine

---

## 1. Worker Manager Core

- [x] 1.1 Create `internal/worker/manager.go` - Worker lifecycle management ✓ 2026-01-15
- [x] 1.2 Create `internal/worker/worker.go` - OpenCode worker abstraction ✓ 2026-01-17
- [x] 1.3 Create `internal/worker/pool.go` - Worker pool with concurrency limits ✓ 2026-01-20
- [ ] 1.4 Create `internal/worker/config.go` - Load provider configs
- [x] 1.5 Implement worker health monitoring and timeout handling ✓ 2026-01-20

## 2. OpenCode ACP Communication

- [x] 2.1 Create `internal/opencode/acp.go` - ACP client for OpenCode ✓ 2026-01-20
- [x] 2.2 Implement JSON-RPC 2.0 transport over stdin/stdout (nd-JSON format) ✓ 2026-01-20
- [x] 2.3 Implement `initialize` handshake with capability negotiation ✓ 2026-01-20
- [x] 2.4 Implement `authenticate` request (if required by OpenCode) ✓ 2026-01-20
- [x] 2.5 Implement `session/new` to create session with provider/model ✓ 2026-01-20
- [x] 2.6 Implement `session/prompt` to send work to OpenCode ✓ 2026-01-20
- [x] 2.7 Implement streaming response handling via `session/update` notifications ✓ 2026-01-20
- [x] 2.8 Implement `session/cancel` for graceful termination ✓ 2026-01-20
- [x] 2.9 Handle client requests from OpenCode (`fs/read_text_file`, `fs/write_text_file`, `terminal/*`) ✓ 2026-01-20

## 3. OpenCode CLI Communication

- [x] 3.1 Create `internal/opencode/cli.go` - CLI wrapper for OpenCode ✓ 2026-01-20
- [x] 3.2 Implement `opencode run --model <provider/model> --prompt "<prompt>"` execution ✓ 2026-01-21
- [x] 3.3 Set `OPENCODE_CONFIG` environment variable for config path ✓ 2026-01-21
- [x] 3.4 Parse output from CLI mode ✓ 2026-01-21
- [x] 3.5 Handle CLI errors and timeouts ✓ 2026-01-21

## 4. Direct Provider API (for simple tasks)

- [x] 4.1 Create `internal/provider/anthropic.go` - Direct Anthropic API ✓ 2026-01-20
- [x] 4.2 Create `internal/provider/openai_compat.go` - OpenAI-compatible APIs ✓ 2026-01-20
- [x] 4.3 Implement streaming responses from direct API ✓ 2026-01-21
- [x] 4.4 Add rate limiting and retry logic ✓ 2026-01-21

## 5. Task Router

- [ ] 5.1 Create `internal/router/router.go` - Task-to-provider routing
- [x] 5.2 Create `internal/router/rules.go` - Routing rule definitions ✓ 2026-01-20
- [x] 5.3 Load routing rules from `.goent/routing.yaml` ✓ 2026-01-20
- [x] 5.4 Implement complexity-based routing (simple → CLI, complex → ACP) ✓ 2026-01-21
- [x] 5.5 Implement context-size routing (large → Kimi 128K) ✓ 2026-01-21
- [x] 5.6 Implement cost-based routing with budget constraints ✓ 2026-01-20
- [x] 5.7 Add manual provider override capability ✓ 2026-01-20

## 6. Provider Configuration

- [x] 6.1 Create `internal/config/providers.go` - Provider config loader ✓ 2026-01-20
- [x] 6.2 Support single OpenCode config file with multiple providers ✓ 2026-01-20
- [x] 6.3 Load provider/model mappings from `.goent/providers.yaml` ✓ 2026-01-21 (implemented in 6.1)
- [x] 6.4 Validate provider connectivity on startup ✓ 2026-01-20
- [x] 6.5 Support environment variable substitution in configs ✓ 2026-01-20
- [x] 6.6 Add provider cost tracking configuration ✓ 2026-01-20

## 7. MCP Tools for Claude Code

- [x] 7.1 Add MCP tool `worker_spawn` - Spawn OpenCode worker ✓ 2026-01-20
  - Parameters: provider, task, method (acp/cli/api)
  - Returns: worker_id

- [x] 7.2 Add MCP tool `worker_prompt` - Send prompt to ACP worker ✓ 2026-01-20

- [x] 7.3 Add MCP tool `worker_status` - Check worker status ✓ 2026-01-20
  - Parameters: worker_id
  - Returns: status, progress, current_step

- [x] 7.4 Add MCP tool `worker_output` - Get worker output ✓ 2026-01-20
   - Parameters: worker_id, since_last
   - Returns: output text

- [x] 7.5 Add MCP tool `worker_cancel` - Cancel worker ✓ 2026-01-20
   - Parameters: worker_id
   - Returns: partial_results

- [x] 7.6 Add MCP tool `worker_list` - List active workers ✓ 2026-01-20
  - Returns: workers with status

- [x] 7.7 Add MCP tool `provider_list` - List configured providers ✓ 2026-01-20
  - Returns: providers with capabilities

- [x] 7.8 Add MCP tool `provider_recommend` - Get optimal provider for task ✓ 2026-01-20
  - Parameters: task_description, context_size
  - Returns: provider, method, rationale

## 8. Result Aggregation

- [x] 8.1 Create `internal/aggregator/aggregator.go` - Collect parallel results ✓ 2026-01-20
- [x] 8.2 Implement conflict detection (multiple workers editing same file) ✓ 2026-01-20
- [x] 8.3 Implement result merging strategy ✓ 2026-01-20
- [x] 8.4 Generate execution summary with per-provider stats ✓ 2026-01-20
- [x] 8.5 Track cost per worker per provider ✓ 2026-01-20

## 9. Integration with Existing Systems

- [x] 9.1 Integrate with OpenSpec registry for task tracking ✓ 2026-01-20
- [x] 9.2 Update task status in tasks.md after worker completion ✓ 2026-01-20
- [x] 9.3 Integrate with context memory for pattern learning ✓ 2026-01-20
- [x] 9.4 Add hooks for pre/post worker execution ✓ 2026-01-20

## 10. Testing

- [ ] 10.1 Unit tests for worker manager
- [ ] 10.2 Unit tests for ACP communication
- [ ] 10.3 Unit tests for CLI communication
- [ ] 10.4 Integration tests for parallel workers
- [ ] 10.5 Test provider failover scenarios
- [ ] 10.6 Benchmark: ACP vs CLI vs API performance
- [ ] 10.7 Test with actual OpenCode + GLM/Kimi providers
