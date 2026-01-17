# Metrics System - Delta Spec

## ADDED Requirements

### Requirement: Token Usage Tracking

The system SHALL track input and output tokens for every tool execution.

#### Scenario: Record token counts
- **WHEN** tool executes
- **THEN** capture input token count from request
- **AND** capture output token count from response
- **AND** store with timestamp and tool name

#### Scenario: Query token usage by tool
- **WHEN** user requests token metrics for a tool
- **THEN** return total tokens, average tokens, token distribution
- **AND** include breakdown by input vs output

---

### Requirement: Performance Metrics Collection

The system SHALL measure execution time and success rate for all tools.

#### Scenario: Measure execution duration
- **WHEN** tool begins execution
- **THEN** record start timestamp
- **WHEN** tool completes execution
- **THEN** calculate duration
- **AND** store duration with metric entry

#### Scenario: Track success rate
- **WHEN** tool execution succeeds
- **THEN** record success=true
- **WHEN** tool execution fails
- **THEN** record success=false with error message
- **AND** calculate success rate as percentage

---

### Requirement: Metrics Storage and Retention

The system SHALL store metrics with configurable retention policy.

#### Scenario: Store metrics in memory
- **WHEN** metric is collected
- **THEN** append to in-memory ring buffer
- **AND** retain last 1000 entries
- **AND** evict oldest entries when buffer full

#### Scenario: Persist metrics to disk
- **WHEN** persistent storage is enabled
- **THEN** write metrics to SQLite database or JSON file
- **AND** apply retention policy (default: 7 days)
- **AND** delete entries older than retention period

---

### Requirement: Metrics Aggregation

The system SHALL compute aggregate statistics over metrics.

#### Scenario: Calculate summary statistics
- **WHEN** user requests metrics summary
- **THEN** calculate average, min, max for duration and tokens
- **AND** calculate p50, p95, p99 percentiles
- **AND** group by tool name or time period

#### Scenario: Time-based grouping
- **WHEN** user requests metrics grouped by time
- **THEN** support grouping by hour, day, or week
- **AND** compute aggregates for each time bucket

---

### Requirement: Metrics Export

The system SHALL export metrics in multiple formats.

#### Scenario: Export as JSON
- **WHEN** user exports metrics to JSON
- **THEN** return array of metric objects
- **AND** include all fields (session, tool, tokens, duration)

#### Scenario: Export as CSV
- **WHEN** user exports metrics to CSV
- **THEN** generate CSV with header row
- **AND** one row per metric entry

#### Scenario: Export as Prometheus
- **WHEN** user exports metrics to Prometheus format
- **THEN** generate Prometheus text format
- **AND** include counters for tokens and durations
- **AND** include gauges for success rates

---

### Requirement: Metrics Dashboard Tool

The system SHALL provide `metrics_show` MCP tool for querying metrics.

#### Scenario: Show metrics for session
- **WHEN** user calls `metrics_show` with session ID
- **THEN** return all metrics for that session
- **AND** include total tokens and duration

#### Scenario: Show metrics for tool
- **WHEN** user calls `metrics_show` with tool name
- **THEN** return aggregated metrics for that tool
- **AND** include call count, success rate, avg tokens, avg duration

#### Scenario: Show metrics for time period
- **WHEN** user calls `metrics_show` with start/end time
- **THEN** return metrics within time range
- **AND** group by tool or session as requested

---

### Requirement: Metrics Privacy and Opt-Out

The system SHALL respect user privacy and provide opt-out mechanism.

#### Scenario: Local storage only
- **WHEN** metrics are collected
- **THEN** store locally only (never transmit)
- **AND** do not include personally identifiable information

#### Scenario: Opt-out of metrics
- **WHEN** user disables metrics in config
- **THEN** do not collect any metrics
- **AND** do not write to metrics store
- **AND** metrics tools return empty results

---

### Requirement: Low Overhead Collection

The system SHALL collect metrics with minimal performance impact.

#### Scenario: Async collection
- **WHEN** metric is collected
- **THEN** write to store asynchronously
- **AND** do not block tool execution
- **AND** overhead is less than 1ms per tool call

#### Scenario: Graceful degradation
- **WHEN** metrics store is unavailable
- **THEN** log warning
- **AND** continue tool execution without metrics
- **AND** do not fail tool execution
