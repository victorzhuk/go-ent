# Hooks System

The hooks system in go-ent provides extensibility through lifecycle event handlers. Hooks can execute commands or suggest agent invocations at key points in the workflow.

## Overview

go-ent supports two types of hooks:

1. **Tool Hooks**: Execute before/after MCP tool calls (e.g., running `goimports` after file edits)
2. **OpenSpec Lifecycle Hooks**: Trigger at workflow milestones (e.g., when a change is created or archived)

## Hook Types

### Command Hooks

Execute shell commands with environment variables:

```yaml
type: command
command: |
  echo "Task completed: $TASK_NUM"
  go test ./...
```

**Features:**
- 10-second timeout
- Environment variable substitution
- JSON input via stdin (for tool hooks)
- Non-zero exit blocks pre-hooks, logs warnings for post-hooks

### Agent Hooks

Log suggestions for agent invocation (no auto-execution):

```yaml
type: agent
agent: reviewer
prompt: "Review code before archiving"
```

**Output example:**
```
💡 Suggestion: Run /ent:reviewer - Review code before archiving
```

**Design rationale:** Agent hooks don't auto-invoke to avoid external dependencies and keep implementation simple.

## Configuration

### File Locations

1. **Embedded defaults**: `pkg/hooks/hooks.json` (built into binary)
2. **Custom config**: Pass file path when creating registry
3. **Agent-specific**: Define in agent YAML metadata

### Default Tool Hooks

The embedded `pkg/hooks/hooks.json` provides:

**PreToolUse:**
- Block dangerous Bash commands (`rm -rf /`, `chmod 777`, etc.)

**PostToolUse:**
- Auto-format Go files with `goimports` after Edit/Write

**Stop:**
- Show modified files and suggest next steps

### OpenSpec Lifecycle Hooks

Defined in `pkg/hooks/openspec.yaml`:

| Event | Trigger Point | Default Action |
|-------|--------------|----------------|
| `onChangeCreated` | After `openspec_new_change` | Log change created |
| `onTasksReady` | When tasks.md is complete | Suggest planner review |
| `onTaskStarted` | When `registry_start_task` is called | Log task started |
| `onTaskCompleted` | When `registry_mark_done` is called | Log completion, suggest tests |
| `beforeArchive` | Before `openspec_archive` | Suggest reviewer check |
| `afterArchive` | After `openspec_archive` | Log archive complete |

## Usage

### Custom Hook Configuration

Create a hooks configuration file:

**hooks.yaml:**
```yaml
tool:
  PreToolUse:
    - matcher: "Bash"
      hooks:
        - type: command
          command: echo "Running bash command..."

  PostToolUse:
    - matcher: "Edit|Write"
      hooks:
        - type: command
          command: |
            read INPUT
            FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path')
            [[ "$FILE" == *.go ]] && goimports -w "$FILE"

openspec:
  onChangeCreated:
    type: command
    command: echo "Change created: $CHANGE_ID"

  beforeArchive:
    type: agent
    agent: reviewer
    prompt: "Final review before archiving $CHANGE_ID"
```

**hooks.json:**
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "echo 'Running bash...'"
          }
        ]
      }
    ]
  },
  "openspec": {
    "onChangeCreated": {
      "type": "command",
      "command": "echo 'Change created'"
    }
  }
}
```

### Agent-Specific Hooks

Define hooks in agent metadata (`pkg/agents/meta/*.yaml`):

```yaml
name: coder
description: Go code implementation
model: sonnet
hooks:
  PostToolUse:
    - matcher: "Edit|Write"
      hooks:
        - type: command
          command: golangci-lint run --new-from-rev=HEAD~1
```

## Tool Name Patterns

Hook matchers use **regular expressions**:

| Pattern | Matches |
|---------|---------|
| `Bash` | Exact match: Bash tool only |
| `Edit\|Write` | Either Edit or Write |
| `.*Edit.*` | Any tool containing "Edit" |
| `^Bash$` | Exact Bash (anchored) |
| `` (empty) | Matches all tools |

## Environment Variables

### Command Hooks

Available in all command hooks:

| Variable | Description | Example |
|----------|-------------|---------|
| `TOOL_NAME` | Name of the MCP tool | `Edit`, `Bash` |

### OpenSpec Hooks

Available in OpenSpec lifecycle hooks:

| Variable | Description | Events |
|----------|-------------|--------|
| `CHANGE_ID` | Change identifier | All |
| `TASK_NUM` | Task number | `onTaskStarted`, `onTaskCompleted` |

### JSON Input (Tool Hooks Only)

Tool hooks receive JSON via stdin:

```json
{
  "tool_name": "Edit",
  "tool_input": {
    "file_path": "main.go",
    "old_string": "...",
    "new_string": "..."
  }
}
```

**Example usage:**
```bash
read INPUT
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path')
echo "Edited file: $FILE"
```

## Implementation Details

### MCP Middleware

Hooks are implemented as MCP protocol middleware:

```go
s.AddReceivingMiddleware(createHookMiddleware(hookRegistry))
```

**Execution flow:**
1. Request arrives at MCP server
2. PreToolUse hooks execute (can block)
3. Tool handler executes
4. PostToolUse hooks execute (non-blocking)
5. Response returned

### Hook Execution

**Pre-hooks:**
- Execute sequentially
- First error blocks tool execution
- Return error to client

**Post-hooks:**
- Execute after tool completes
- Errors logged, don't affect response
- Always run, even if tool failed

## Testing

Run hook tests:

```bash
go test ./internal/hooks/... -v
```

Test custom configuration:

```bash
# Create test config
cat > /tmp/hooks.yaml <<EOF
openspec:
  onChangeCreated:
    type: command
    command: echo "Test hook"
EOF

# Test loading
go test ./internal/hooks/... -run TestRegistry_LoadFromFile
```

## Security Considerations

1. **Command Injection**: Hooks execute shell commands with substituted variables. Sanitize inputs in production.
2. **Timeout**: All commands have a 10-second timeout to prevent hangs.
3. **Permissions**: Hooks run with the same permissions as the MCP server process.

**Default safety hooks:**
- Block `rm -rf /`
- Block `chmod 777`
- Block `dd if=/dev`

## Troubleshooting

### Hooks Not Firing

1. Check hook registry initialization:
   ```
   INFO  initialized hook registry
   ```

2. Verify pattern matches tool name:
   ```go
   // Test pattern matching
   executor.MatchTool("Edit|Write", "Edit") // true
   ```

3. Check logs for hook execution errors

### Hook Command Fails

- Commands timeout after 10 seconds
- Check command syntax (must be valid bash)
- Verify environment variables are set
- Test command standalone: `bash -c 'YOUR_COMMAND'`

### JSON Parsing Issues

For tool hooks, verify JSON structure:
```bash
# Test parsing
echo '{"tool_name":"Edit","tool_input":{...}}' | jq .
```

## Examples

### Auto-format on Save

```yaml
tool:
  PostToolUse:
    - matcher: "Edit|Write"
      hooks:
        - type: command
          command: |
            read INPUT
            FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path')
            case "$FILE" in
              *.go) goimports -w "$FILE" ;;
              *.py) black "$FILE" ;;
              *.rs) rustfmt "$FILE" ;;
            esac
```

### Block Dangerous Operations

```yaml
tool:
  PreToolUse:
    - matcher: "Bash"
      hooks:
        - type: command
          command: |
            read INPUT
            CMD=$(echo "$INPUT" | jq -r '.tool_input.command')
            if echo "$CMD" | grep -q "rm -rf"; then
              echo "Blocked: rm -rf" >&2
              exit 2
            fi
```

### Workflow Notifications

```yaml
openspec:
  onTaskCompleted:
    type: command
    command: |
      curl -X POST https://slack.com/api/chat.postMessage \
        -H "Authorization: Bearer $SLACK_TOKEN" \
        -d "text=Task $TASK_NUM completed in $CHANGE_ID"

  afterArchive:
    type: command
    command: |
      gh pr create --title "Complete $CHANGE_ID" \
        --body "Change archived and ready for review"
```

## Architecture

```
┌─────────────────────────────────────────┐
│         MCP Server                      │
│  ┌────────────────────────────────────┐ │
│  │   Hook Middleware                  │ │
│  │   ┌──────────┐   ┌──────────────┐ │ │
│  │   │ PreHooks │ → │ Tool Handler │ │ │
│  │   └──────────┘   └──────────────┘ │ │
│  │         ↓              ↓           │ │
│  │   ┌──────────┐   ┌──────────────┐ │ │
│  │   │PostHooks │ ← │   Response   │ │ │
│  │   └──────────┘   └──────────────┘ │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
           ↓
    ┌─────────────┐
    │Hook Registry│
    └─────────────┘
           ↓
    ┌─────────────┐
    │Hook Executor│
    └─────────────┘
```

## Future Enhancements

Potential improvements (not yet implemented):

1. **Async execution**: Run post-hooks in goroutines
2. **Hook marketplace**: Share and discover community hooks
3. **Conditional hooks**: Execute based on runtime conditions
4. **Hook chaining**: Dependencies between hooks
5. **Metrics**: Track hook execution times and success rates
6. **Auto-invoke agents**: Optional auto-execution for agent hooks

## See Also

- [MCP Tools](./MCP_TOOLS.md) - Tool catalog
- [OpenSpec Workflow](./OPENSPEC.md) - Workflow lifecycle
- [Agent System](./AGENTS_AND_SKILLS.md) - Agent metadata
