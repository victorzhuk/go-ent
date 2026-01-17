# Metrics Collection

go-ent collects anonymous metrics about tool usage for performance optimization and development insights.

## Opt-Out

### Via Config File

Add to `.go-ent/config.yaml`:

```yaml
metrics:
  enabled: false
```

### Via Environment Variable

```bash
export GOENT_METRICS_ENABLED=false
```

### Data Collected

When enabled, go-ent collects:

- Tool name (e.g., `go-ent:execute`)
- Execution duration
- Token usage (estimated)
- Success/failure status
- Error messages (if any)
- Timestamp
- Session ID (anonymous UUID)

### Metrics Schema

Each metric entry contains the following fields:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `session_id` | string | Unique identifier for a session (UUID v7). Correlates metrics from the same tool call chain. | `"550e8400-e29b-41d4-a716-446655440000"` |
| `tool_name` | string | Name of the MCP tool that was executed. | `"go_ent_agent_spawn"` |
| `tokens_in` | number | Estimated input tokens (tokens in prompt/response). Set to 0 if not available. | `1234` |
| `tokens_out` | number | Estimated output tokens (tokens in response). Set to 0 if not available. | `567` |
| `duration` | string | Execution duration in RFC3339 format (nanosecond precision). | `"1.234567s"` |
| `success` | boolean | Whether the tool execution succeeded. | `true` |
| `error_msg` | string | Error message if execution failed. Empty string on success. | `""` or `"timeout: tool not responding"` |
| `timestamp` | string | When the metric was recorded (RFC3339 format). | `"2026-01-17T15:30:45.123456789Z"` |
| `metadata` | object | Additional context data (key-value pairs). Optional field, usually empty. | `{}` or `{"user_id": "123"}` |

#### Example Metric Entry

```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "tool_name": "go_ent_agent_spawn",
  "tokens_in": 1234,
  "tokens_out": 567,
  "duration": "1.234567s",
  "success": true,
  "error_msg": "",
  "timestamp": "2026-01-17T15:30:45.123456789Z",
  "metadata": {}
}
```

#### Field Constraints

- **session_id**: Valid UUID v7 format
- **tool_name**: Non-empty string, valid MCP tool name
- **tokens_in/tokens_out**: Non-negative integers (0 if unavailable)
- **duration**: Non-negative duration (RFC3339)
- **success**: Boolean value
- **error_msg**: Empty on success, non-empty on failure
- **timestamp**: RFC3339 format, not zero
- **metadata**: Object with string keys and values

#### Storage Format

Metrics are stored in JSON format at `data/metrics.json`. Each entry is a JSON object with the schema above. The file is rewritten on each metric addition.

#### Retention Policy

Old metrics are automatically removed based on retention period (default: 7 days). Only metrics within the retention window are stored.

### Data NOT Collected

go-ent does NOT store:

- Personal data (names, emails, etc.)
- Code content or source files
- File paths or project names
- API keys or credentials
- User-provided input data

### Privacy

All metrics are stored locally in `data/metrics.json`. No data is transmitted to external servers. Retention period is 7 days by default.

### Viewing Metrics

Use the metrics tools to view collected data:

```
/go-ent:metrics:show --format summary
/go-ent:metrics:export --format json
```

See `docs/DEVELOPMENT.md` for more details on the metrics system.
