# Proposal: Add Metrics and Performance Monitoring

## Overview

Implement comprehensive metrics collection and monitoring system to track token usage, tool execution performance, agent selection accuracy, and system health for go-ent MCP server.

## Rationale

### Problem

- No visibility into token consumption per tool/session
- Cannot measure tool discovery effectiveness
- No performance metrics for optimization decisions
- Difficult to validate claims (e.g., "70-90% token reduction")
- No system health monitoring

### Solution

- **Metrics Store**: Time-series storage for tool/session metrics
- **Token Tracking**: Track input/output tokens per tool call
- **Performance Monitoring**: Measure tool execution duration and success rates
- **Dashboard MCP Tool**: Query and visualize metrics
- **Export Capabilities**: Prometheus, JSON, CSV formats

## Key Components

### Metrics Categories

| Category | Metrics | Purpose |
|----------|---------|---------|
| **Token Usage** | Input tokens, output tokens, total tokens | Cost tracking, optimization |
| **Tool Performance** | Execution time, success rate, error rate | Performance tuning |
| **Discovery** | Search queries, tool loads, activation rate | Tool discovery effectiveness |
| **Agent Selection** | Complexity scores, model choices, skill matches | Agent system accuracy |

### Implementation Files

1. `internal/metrics/store.go` - Time-series metrics storage
2. `internal/metrics/collector.go` - Metrics collection middleware
3. `internal/metrics/exporter.go` - Export to various formats
4. `internal/metrics/aggregator.go` - Summary calculations

### New MCP Tools

| Tool | Description |
|------|-------------|
| `metrics_show` | Display metrics for session/tool/time period |
| `metrics_summary` | Aggregate statistics (avg, p50, p95, p99) |
| `metrics_export` | Export metrics to file (JSON, CSV, Prometheus) |
| `metrics_reset` | Clear metrics (for testing) |

## Dependencies

- Requires: None (Phase 5.2 - Independent)
- Blocks: None
- Complements: add-tool-discovery (validates token reduction)

## Success Criteria

- [ ] Token tracking per tool call
- [ ] Performance metrics (duration, success rate)
- [ ] Metrics dashboard accessible via MCP
- [ ] Export to Prometheus format
- [ ] Historical data retention (configurable)
- [ ] Validation: Can prove token reduction claims

## Impact

### Performance

- **Overhead**: <1ms per tool call (async collection)
- **Storage**: ~1KB per metric entry
- **Retention**: Configurable (default: 7 days)

### Use Cases

- **Cost Optimization**: Identify high-token tools
- **Performance Tuning**: Find slow operations
- **Discovery Validation**: Measure tool search accuracy
- **Capacity Planning**: Track usage trends

## Architecture

```
Metrics System
├── Collector (middleware)
│   ├── Pre-execution hook
│   ├── Post-execution hook
│   └── Async write to store
├── Store (time-series)
│   ├── In-memory (last 1000 entries)
│   ├── Persistent (SQLite or file)
│   └── Retention policy
├── Aggregator
│   ├── Statistics (avg, min, max, percentiles)
│   ├── Time-based grouping (hour, day, week)
│   └── Tool/session filtering
└── Exporter
    ├── Prometheus scrape endpoint
    ├── JSON export
    └── CSV export
```

## Metric Schema

```go
type Metric struct {
    SessionID     string
    ToolName      string
    TokensIn      int
    TokensOut     int
    Duration      time.Duration
    Success       bool
    ErrorMsg      string
    Timestamp     time.Time
    Metadata      map[string]string
}
```

## Migration

**For Users:**
1. Metrics collection enabled by default (opt-out via config)
2. No performance impact (<1ms overhead)
3. Data stored locally, never sent externally

**For Operators:**
1. Configure retention period in config
2. Optionally expose Prometheus endpoint
3. Use `metrics_export` for external analysis

## Privacy

- **Local Only**: Metrics stored locally, never transmitted
- **Opt-Out**: Can disable metrics collection via config
- **No PII**: Only tool names, tokens, durations tracked
- **Anonymized**: Session IDs are ephemeral UUIDs
