# ACP Protocol Research Findings

**Date**: 2026-01-10
**Researcher**: Claude Sonnet 4.5
**Purpose**: Verify OpenCode ACP implementation against add-acp-agent-mode proposal assumptions

---

## Executive Summary

✅ **VERIFIED**: OpenCode supports ACP via `opencode acp` command
⚠️ **CORRECTED**: Configuration mechanism differs from proposal (no `--config` flag)
⚠️ **CORRECTED**: JSON-RPC method names differ from proposal assumptions
✅ **CONFIRMED**: JSON-RPC 2.0 over stdio transport
✅ **CONFIRMED**: Supports multiple AI providers (GLM, Kimi, DeepSeek, Anthropic)

---

## Detailed Findings

### 1. ACP Command Existence

**Proposal Assumption:**
```bash
opencode acp --config ~/.opencode-glm.json
```

**Actual Reality:**
```bash
opencode acp  # No --config flag exists
```

**Configuration Instead:**
- `OPENCODE_CONFIG` environment variable for config file path
- `OPENCODE_CONFIG_DIR` environment variable for config directory
- `OPENCODE_CONFIG_CONTENT` environment variable for inline JSON
- Default location: `~/.config/opencode/opencode.json`

**Sources:**
- [OpenCode CLI Documentation](https://opencode.ai/docs/cli/)
- [OpenCode ACP Support](https://opencode.ai/docs/acp/)

---

### 2. JSON-RPC Methods

**Proposal Assumptions:**
- `acp/initialize` - Initialize handshake
- `session/prompt` - Send task to agent
- `session/cancel` - Cancel worker

**Actual ACP Protocol Methods:**

| Category | Method | Purpose |
|----------|--------|---------|
| **Initialization** | `initialize` | Establishes connection, negotiates protocol version |
| | `authenticate` | Authenticates client using specified method |
| | `session/new` | Creates new conversation session |
| | `session/load` | Resumes existing session (optional) |
| **Session Ops** | `session/prompt` | ✅ **CORRECT** - Processes user message |
| | `session/cancel` | ✅ **CORRECT** - Cancels ongoing operations |
| | `session/set_mode` | Switches agent operating mode |
| | `session/update` | Streams real-time progress notifications |
| **Client Requests** | `fs/read_text_file` | Agent requests file access |
| | `fs/write_text_file` | Agent writes files |
| | `terminal/create` | Execute command |
| | `terminal/output` | Get command output |
| | `terminal/kill` | Terminate command |
| | `session/request_permission` | Request user authorization |

**Key Differences:**
- ❌ No `acp/initialize` method (just `initialize`)
- ✅ `session/prompt` is correct
- ✅ `session/cancel` is correct
- ⚠️ Must call `session/new` before `session/prompt`

**Sources:**
- [Agent Client Protocol Schema](https://github.com/agentclientprotocol/agent-client-protocol)
- [ACP Documentation](https://agentclientprotocol.com/)

---

### 3. Provider Configuration

**Proposal Assumption:**
Multiple provider-specific config files:
- `~/.opencode-glm.json`
- `~/.opencode-kimi.json`
- `~/.opencode-deepseek.json`

**Actual Reality:**
Single `opencode.json` with multiple providers:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "moonshot": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "Moonshot AI",
      "options": {
        "baseURL": "https://api.moonshot.cn/v1"
      },
      "models": {
        "kimi-k2": {
          "name": "Kimi K2"
        }
      }
    },
    "deepseek": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "DeepSeek",
      "options": {
        "baseURL": "https://api.deepseek.com/v1"
      }
    },
    "anthropic": {
      "options": {
        "baseURL": "https://api.anthropic.com/v1"
      }
    }
  }
}
```

**Model Selection:**
- Use `/models` command in OpenCode TUI to select provider/model
- Use `--model provider/model` flag in CLI mode
- Cannot switch provider dynamically in ACP mode (session-bound)

**Sources:**
- [OpenCode Providers Documentation](https://opencode.ai/docs/providers/)
- [Kimi K2 Setup Guide](https://medium.com/@connectshefeek/opencode-kimi-k2-model-getting-started-524b90b7a5d7)

---

### 4. Communication Transport

**Proposal Assumption:**
JSON-RPC 2.0 over stdio

**Actual Reality:**
✅ **CONFIRMED** - nd-JSON (newline-delimited JSON) over stdin/stdout

**Message Format:**
```json
// Request
{"jsonrpc":"2.0","id":"1","method":"session/prompt","params":{"message":"Fix the bug"}}

// Response
{"jsonrpc":"2.0","id":"1","result":{"content":"..."}}

// Notification (no id)
{"jsonrpc":"2.0","method":"session/update","params":{"type":"agent_message_chunk","content":"..."}}
```

**Sources:**
- [OpenCode ACP Documentation](https://opencode.ai/docs/acp/)
- [ACP Protocol Schema](https://raw.githubusercontent.com/agentclientprotocol/agent-client-protocol/main/schema/schema.json)

---

### 5. Streaming Support

**Proposal Assumption:**
Streaming responses via ACP

**Actual Reality:**
✅ **CONFIRMED** - Via `session/update` notifications

**Update Types:**
- `user_message_chunk` - User input streaming
- `agent_message_chunk` - Model response streaming
- `agent_thought_chunk` - Internal reasoning visibility
- `tool_call` - New tool execution initiated
- `tool_call_update` - Tool execution results/status
- `plan` - Agent execution plans
- `available_commands_update` - Commands ready/changed
- `current_mode_update` - Active mode changes

**Sources:**
- [ACP Protocol Schema](https://github.com/agentclientprotocol/agent-client-protocol)

---

### 6. CLI Mode (Non-ACP)

**Proposal Assumption:**
```bash
opencode -p "prompt" -f json --config ~/.opencode-glm.json
```

**Actual Reality:**
```bash
opencode run --model provider/model --prompt "prompt"
# OR
opencode run -c "session-id" --prompt "prompt"  # continue session
```

**Key Differences:**
- ❌ No `-p` flag (use `--prompt` or `-m` for message)
- ❌ No `-f json` flag (output format not configurable)
- ❌ No `--config` flag (use `OPENCODE_CONFIG` env var)
- ✅ Use `--model provider/model` to select provider

**Sources:**
- [OpenCode CLI Documentation](https://opencode.ai/docs/cli/)

---

### 7. Direct API Mode

**Proposal Assumption:**
Bypass OpenCode and call provider APIs directly

**Actual Reality:**
✅ **POSSIBLE** - Can implement direct API calls to:
- Anthropic API (Claude models)
- OpenAI-compatible APIs (GLM, Kimi, DeepSeek via their base URLs)

**Recommendation:**
For simple queries, direct API calls are faster than spawning OpenCode process.

---

## Impact on Proposal

### Critical Changes Required

| Proposal Section | Change Required |
|------------------|----------------|
| **Worker spawning** | Replace `--config` flag with `OPENCODE_CONFIG` env var |
| **ACP handshake** | Use `initialize` + `session/new`, not `acp/initialize` |
| **Provider switching** | Cannot switch providers in running ACP session |
| **Multi-provider** | Use single `opencode.json`, select provider via `--model` in CLI mode |
| **CLI execution** | Update command syntax (no `-p`, `-f` flags) |

### Design Implications

1. **Provider Selection Strategy:**
   - ACP mode: Session-bound provider (set at spawn)
   - CLI mode: Per-request provider via `--model` flag
   - API mode: Direct provider API calls

2. **Configuration Management:**
   - Single `opencode.json` with all providers
   - Worker spawns with `OPENCODE_CONFIG` env var pointing to config
   - Provider selection via `--model` flag (CLI) or session config (ACP)

3. **Worker Pool Architecture:**
   - Cannot reuse ACP workers for different providers
   - Each worker is provider-specific (set at session creation)
   - CLI mode more flexible but higher startup overhead

---

## Recommendations

### 1. Update Proposal Protocol Details

Replace assumed method names:
- ❌ `acp/initialize` → ✅ `initialize`
- ✅ `session/prompt` (correct)
- ✅ `session/cancel` (correct)
- ➕ Add `session/new` (required before prompts)

### 2. Update Configuration Strategy

**From (proposed):**
```go
func (m *WorkerManager) SpawnACP(provider string, task *Task) (*OpenCodeWorker, error) {
    configPath := m.configs[provider]  // ~/.opencode-glm.json
    cmd := exec.Command("opencode", "acp", "--config", configPath)
}
```

**To (actual):**
```go
func (m *WorkerManager) SpawnACP(provider string, task *Task) (*OpenCodeWorker, error) {
    configPath := m.configs["opencode"]  // Single opencode.json
    cmd := exec.Command("opencode", "acp")
    cmd.Env = append(os.Environ(),
        fmt.Sprintf("OPENCODE_CONFIG=%s", configPath),
    )
}
```

### 3. Provider Selection Mechanism

**Option A: Pre-select in config (recommended for ACP)**
Create provider-specific configs:
```json
// opencode-glm.json
{
  "provider": { "moonshot": {...} },
  "defaultModel": "moonshot/glm-4"
}

// opencode-kimi.json
{
  "provider": { "moonshot": {...} },
  "defaultModel": "moonshot/kimi-k2"
}
```

**Option B: CLI mode for dynamic selection**
Use CLI mode with `--model` flag when provider flexibility needed.

### 4. Update CLI Command Syntax

**From (proposed):**
```bash
opencode -p "prompt" -f json --config ~/.opencode-glm.json
```

**To (actual):**
```bash
OPENCODE_CONFIG=~/.opencode-glm.json opencode run --model moonshot/glm-4 --prompt "prompt"
```

---

## Sources

- [OpenCode ACP Support](https://opencode.ai/docs/acp/)
- [OpenCode CLI Documentation](https://opencode.ai/docs/cli/)
- [OpenCode Providers](https://opencode.ai/docs/providers/)
- [Agent Client Protocol](https://agentclientprotocol.com/)
- [ACP GitHub Repository](https://github.com/agentclientprotocol/agent-client-protocol)
- [ACP Protocol Schema](https://raw.githubusercontent.com/agentclientprotocol/agent-client-protocol/main/schema/schema.json)
- [OpenCode Kimi K2 Setup](https://medium.com/@connectshefeek/opencode-kimi-k2-model-getting-started-524b90b7a5d7)

---

## Next Steps

1. ✅ Update `design.md` with correct ACP protocol details
2. ✅ Update `specs/acp-protocol/spec.md` with actual method names
3. ✅ Update `tasks.md` with correct implementation approach
4. ✅ Remove Phase 3 (Dynamic MCP Discovery) to separate proposal
5. ✅ Add dependency blocker notes
