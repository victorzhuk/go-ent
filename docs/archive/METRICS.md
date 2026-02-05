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

## Usage Examples

### metrics_show - View Raw Metrics

**Basic usage:**
```
mcp__go_ent__metrics_show()
```
Returns all metrics (last 100 by default).

**Filter by tool name:**
```
mcp__go_ent__metrics_show(tool_name="go_ent_agent_spawn")
```
Shows only agent_spawn tool metrics.

**Filter by time range:**
```
mcp__go_ent__metrics_show(
    start_time="2026-01-01T00:00:00Z",
    end_time="2026-01-17T23:59:59Z"
)
```
Shows metrics for the given date range.

**Limit results:**
```
mcp__go_ent__metrics_show(limit=50)
```
Returns at most 50 most recent metrics.

**Export to JSON format:**
```
mcp__go_ent__metrics_show(format="json")
```
Returns metrics in JSON instead of table.

**Example output:**
```
## Metrics Table

| Tool           | Session                        | Tokens In | Tokens Out | Duration | Success |
|----------------|--------------------------------|-----------|------------|-----------|----------|
| agent_spawn    | 550e8400-e29b-41d4-a716-446655440 | 1234      | 567        | 1.2s      | ✓       |
| agent_status   | 550e8400-e29b-41d4-a716-446655440 | 560       | 234        | 0.3s      | ✓       |
```

### metrics_summary - View Aggregated Statistics

**Overall summary (no grouping):**
```
mcp__go_ent__metrics_summary(group_by="none")
```
Shows overall statistics across all metrics with percentiles.

**Group by tool name:**
```
mcp__go_ent__metrics_summary(group_by="tool")
```
Shows one row per tool with per-tool statistics.

**Group by time buckets:**
```
mcp__go_ent__metrics_summary(group_by="hour")
```
Shows metrics grouped by hour.

```
mcp__go_ent__metrics_summary(group_by="day")
```
Shows metrics grouped by day.

**Combined filters:**
```
mcp__go_ent__metrics_summary(
    group_by="tool",
    start_time="2026-01-15T00:00:00Z",
    end_time="2026-01-17T23:59:59Z"
)
```
Shows per-tool statistics for a specific time range.

**Export to JSON:**
```
mcp__go_ent__metrics_summary(group_by="tool", format="json")
```
Returns summary in JSON format instead of table.

**Example output (group_by="tool"):**
```
## Summary by Tool

┌─────────────────────────────┬─────────┬──────────┬─────────────┬─────────────┬──────────────┐
│ Tool                    │  Count  │ Success% │ Avg Duration │ Avg Tokens │ Total Cost   │
├─────────────────────────────┼─────────┼──────────┼─────────────┼─────────────┼──────────────┤
│ agent_spawn             │     42  │   95.2%  │      1.2s   │        120  │     $0.45    │
│ agent_status            │     85  │  100.0%  │      0.3s   │         40  │     $0.10    │
│ agent_list              │     23  │  100.0%  │      0.1s   │         12  │     $0.03    │
└─────────────────────────────┴─────────┴──────────┴─────────────┴─────────────┴──────────────┘

## Overall Statistics

- **Total executions**: 150
- **Success rate**: 96.7%
- **Total cost**: $0.58
- **Avg duration**: 0.63s
```

### metrics_export - Export Metrics to File

**Export to JSON with default filename:**
```
mcp__go_ent__metrics_export(format="json")
```
Creates `metrics_20260117_150405.json` in current directory.

**Export to CSV with custom filename:**
```
mcp__go_ent__metrics_export(
    format="csv",
    filename="january_metrics"
)
```
Creates `january_metrics.csv`.

**Export to Prometheus format:**
```
mcp__go_ent__metrics_export(format="prometheus")
```
Creates `metrics_20260117_150405.prom` with Prometheus scrape format.

**Filter and export:**
```
mcp__go_ent__metrics_export(
    format="json",
    tool_name="go_ent_agent_spawn",
    start_time="2026-01-15T00:00:00Z",
    end_time="2026-01-17T23:59:59Z"
)
```
Exports only agent_spawn metrics for the time range.

**Limit records:**
```
mcp__go_ent__metrics_export(format="csv", limit=1000)
```
Exports at most 1000 most recent metrics.

**Example return value:**
```json
{
  "success": true,
  "filename": "metrics_20260117_150405.json",
  "format": "json",
  "records": 42,
  "file_size_bytes": 8192,
  "path": "/home/user/project/metrics_20260117_150405.json"
}
```

### metrics_reset - Clear All Metrics

**Basic usage:**
```
mcp__go_ent__metrics_reset()
```
Clears all metrics from the store.

**Example return value:**
```json
{
  "success": true,
  "message": "All metrics cleared from store",
  "previous_count": 42,
  "current_count": 0
}
```

**Note**: This operation cannot be undone. Use with caution, especially in production.

## Export Formats

Metrics can be exported to three formats: JSON, CSV, and Prometheus.

### JSON Format

**Description**: Structured JSON format with human-readable indentation.

**Use case**: Programmatic access, API integration, data analysis.

**File extension**: `.json`

**Example output**:
```json
[
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
]
```

**Format details**:
- Array of metric objects
- 2-space indentation
- All fields from schema included
- Timestamp in RFC3339 format with nanosecond precision

**When to use**:
- Loading metrics into other applications
- Manual inspection with pretty printing
- API responses or web UIs

### CSV Format

**Description**: Spreadsheet-compatible CSV format with header row.

**Use case**: Spreadsheet analysis, Excel/Google Sheets import.

**File extension**: `.csv`

**Example output**:
```csv
SessionID,ToolName,TokensIn,TokensOut,Duration,Success,ErrorMsg,Timestamp
550e8400-e29b-41d4-a716-446655440000,go_ent_agent_spawn,1234,567,1.234567s,true,,2026-01-17T15:30:45Z
550e8400-e29b-41d4-a716-446655440001,agent_status,560,234,0.345678s,true,,2026-01-17T15:35:12Z
```

**Format details**:
- Header row with field names
- One metric per line
- Comma-separated values
- Duration in Go's string format (e.g., "1.234567s")
- Empty fields use empty string

**When to use**:
- Opening in Excel, Google Sheets, Numbers
- Data analysis with spreadsheet formulas
- Generating reports

### Prometheus Format

**Description**: Prometheus scrape format for time-series monitoring systems.

**Use case**: Prometheus monitoring dashboards, Grafana integration.

**File extension**: `.prom`

**Example output**:
```prom
# HELP tool_tokens_total Total tokens consumed per tool
# TYPE tool_tokens_total gauge
# HELP tool_duration_seconds Tool execution duration in seconds
# TYPE tool_duration_seconds gauge
# HELP tool_success_total Total tool execution success count
# TYPE tool_success_total counter

tool_tokens_total{session="550e8400-e29b-41d4-a716-446655440000",tool="go_ent_agent_spawn",success="1"} 1801
tool_duration_seconds{session="550e8400-e29b-41d4-a716-446655440000",tool="go_ent_agent_spawn",success="1"} 1.235
tool_success_total{session="550e8400-e29b-41d4-a716-446655440000",tool="go_ent_agent_spawn",success="1"} 1
```

**Format details**:
- HELP comments for metric descriptions
- TYPE comments for metric types (gauge, counter)
- Labels: session, tool, success status
- Timestamps are handled by Prometheus scrape time
- No explicit timestamps in data

**When to use**:
- Integrating with Prometheus monitoring
- Building Grafana dashboards
- Alerting on tool performance
- Production monitoring stacks

### Format Comparison

| Format | Pros | Cons | Best For |
|---------|-------|-------|----------|
| JSON | Human-readable, structured, tools support | Large file size | APIs, automation |
| CSV | Spreadsheet-compatible, small size | No nesting, no types | Data analysis, reports |
| Prometheus | Time-series optimized, monitoring | Limited fields, labels only | Production monitoring, dashboards |

### Exporting Metrics

Use the `metrics_export` tool:

```bash
# Export to JSON
mcp__go_ent__metrics_export(format="json", filename="my_metrics")

# Export to CSV
mcp__go_ent__metrics_export(format="csv", filename="report_january")

# Export to Prometheus format
mcp__go_ent__metrics_export(format="prometheus", filename="metrics_for_prometheus")
```

See [Usage Examples](#usage-examples) for more details.

### File Naming

**With custom filename**: `{filename}.{extension}`
- `my_metrics.json`
- `report_january.csv`
- `metrics_for_prometheus.prom`

**With timestamp**: `metrics_{timestamp}.{extension}`
- `metrics_20260117_150405.json`
- `metrics_20260117_150405.csv`
- Timestamp format: `YYYYMMDD_HHMMSS`

### Filtering Before Export

You can filter metrics before export:

```bash
# Export only agent_spawn metrics
mcp__go_ent__metrics_export(
    format="json",
    tool_name="go_ent_agent_spawn"
)

# Export last 7 days
mcp__go_ent__metrics_export(
    format="csv",
    start_time="2026-01-10T00:00:00Z",
    end_time="2026-01-17T23:59:59Z"
)

# Export up to 1000 records
mcp__go_ent__metrics_export(format="prometheus", limit=1000)
```

See [Usage Examples](#usage-examples) for filter documentation.

## Dashboard Examples

This section provides ready-to-use examples for visualizing go-ent metrics in monitoring tools.

### Grafana Dashboard

Use these panel configurations to build a Grafana dashboard for go-ent metrics.

**Panel 1: Tool Execution Rate**
```json
{
  "title": "Tool Execution Rate",
  "targets": [
    {
      "expr": "sum(rate(tool_success_total{instance=\"go-ent\"}[5m])) by (tool)",
      "legendFormat": "{{tool}}"
    }
  ],
  "description": "Tool executions per second (5m rate)",
  "type": "graph"
}
```

**Panel 2: Average Duration by Tool**
```json
{
  "title": "Average Tool Duration",
  "targets": [
    {
      "expr": "avg(tool_duration_seconds{instance=\"go-ent\"}) by (tool)",
      "legendFormat": "{{tool}}"
    }
  ],
  "description": "Average execution time per tool",
  "type": "graph"
}
```

**Panel 3: Success Rate Gauge**
```json
{
  "title": "Success Rate",
  "targets": [
    {
      "expr": "avg(success_rate{instance=\"go-ent\"}) by (tool)",
      "legendFormat": "{{tool}}"
    }
  ],
  "description": "Success percentage by tool",
  "type": "gauge"
}
```

### Prometheus Queries

Use these PromQL queries in Prometheus or Grafana to analyze go-ent metrics.

**Total token usage:**
```promql
sum(tool_tokens_total{instance="go-ent"})
```

**Top 5 tools by execution count:**
```promql
topk(5, sum(rate(tool_success_total{instance="go-ent"}[1h])) by (tool))
```

**Average duration (last hour):**
```promql
avg(tool_duration_seconds{instance="go-ent"}[1h:])
```

**Success rate (last 24h):**
```promql
sum(increase(tool_success_total{instance="go-ent",success="1"}[24h])) / sum(increase(tool_success_total{instance="go-ent"}[24h])) * 100
```

**Error rate by tool:**
```promql
sum(rate(tool_success_total{instance="go-ent",success="0"}[5m])) by (tool)
```

### Alert Examples

Use these alert definitions in Prometheus to monitor go-ent performance.

**High error rate alert:**
```yaml
groups:
  - name: go-ent_metrics
    rules:
      - alert: HighErrorRate
        expr: |
          sum(rate(tool_success_total{instance="go-ent",success="0"}[5m])) > 0.1
        for: 5m
        annotations:
          summary: "High error rate detected for go-ent tools"
        labels:
          severity: warning
```

**Tool timeout alert:**
```yaml
- alert: SlowToolExecution
  expr: avg(tool_duration_seconds{instance="go-ent"}) by (tool) > 30
  for: 5m
  annotations:
    summary: "Tool {{ $labels.tool }} taking longer than 30s average"
  labels:
    severity: critical
```

### Quick Start Guide

Follow these steps to set up monitoring for go-ent metrics.

**1. Export metrics to Prometheus format:**
```bash
mcp__go_ent__metrics_export(format="prometheus", filename="metrics_for_prometheus")
```

**2. Start Prometheus with metrics file:**
```bash
prometheus --config.file=prometheus.yml
```

**3. Import dashboard into Grafana:**
- Create new dashboard
- Paste JSON configuration from examples above
- Set Prometheus data source

**Example prometheus.yml:**
```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'go-ent-metrics'
    file_sd_configs:
      - files:
        - 'data/metrics_for_prometheus.prom'
```

**Tips:**
- Use `file_sd_configs` for static metrics files
- Set appropriate `scrape_interval` based on metrics update frequency
- Configure Grafana to auto-refresh every 15-30 seconds for near real-time monitoring
- Use dashboard variables to filter by tool name or time range

## Common Patterns

**ISO 8601 timestamps:**
Always use ISO 8601 format (RFC3339) for time filters:
```
2026-01-17T15:30:45Z
```

**Filter combinations:**
You can combine multiple filters:
```
mcp__go_ent__metrics_summary(
    group_by="hour",
    tool_name="agent_spawn",
    start_time="2026-01-01T00:00:00Z",
    end_time="2026-01-31T23:59:59Z"
)
```

**Format options:**
All tools support format parameter:
- `table` (default): Human-readable ASCII table
- `json`: Machine-readable JSON
- `csv`: Spreadsheet-compatible CSV

## Tips

- **Default limits**: metrics_show defaults to 100, metrics_export defaults to 1000
- **Time zones**: Use `Z` suffix for UTC or include timezone offset
- **Empty results**: Tools return "No metrics found" when filters match nothing
- **Export location**: Files are created in current working directory

### Viewing Metrics

Use the metrics tools to view collected data:

```
/go-ent:metrics:show --format summary
/go-ent:metrics:export --format json
```

See `docs/DEVELOPMENT.md` for more details on the metrics system.
