# Tasks: Add Metrics and Performance Monitoring

## 1. Metrics Storage
- [ ] 1.1 Create `internal/metrics/store.go`
- [ ] 1.2 Define `Metric` struct with all fields
- [ ] 1.3 Implement in-memory ring buffer (last 1000 entries)
- [ ] 1.4 Add persistent storage option (SQLite or JSON file)
- [ ] 1.5 Implement retention policy (delete old entries)
- [ ] 1.6 Add thread-safe read/write with RWMutex

## 2. Metrics Collection
- [ ] 2.1 Create `internal/metrics/collector.go`
- [ ] 2.2 Implement pre-execution hook
- [ ] 2.3 Implement post-execution hook
- [ ] 2.4 Add async write to avoid blocking
- [ ] 2.5 Extract token counts from MCP responses
- [ ] 2.6 Measure execution duration
- [ ] 2.7 Track success/failure status

## 3. Aggregation and Statistics
- [ ] 3.1 Create `internal/metrics/aggregator.go`
- [ ] 3.2 Implement average calculation
- [ ] 3.3 Implement percentile calculation (p50, p95, p99)
- [ ] 3.4 Add time-based grouping (hour, day, week)
- [ ] 3.5 Add filtering by tool name
- [ ] 3.6 Add filtering by session ID
- [ ] 3.7 Calculate success rate percentage

## 4. Export Capabilities
- [ ] 4.1 Create `internal/metrics/exporter.go`
- [ ] 4.2 Implement JSON export
- [ ] 4.3 Implement CSV export
- [ ] 4.4 Implement Prometheus format export
- [ ] 4.5 Add HTTP endpoint for Prometheus scraping (optional)
- [ ] 4.6 Add file export with timestamp

## 5. MCP Tools Implementation
- [ ] 5.1 Implement `metrics_show` tool
- [ ] 5.2 Implement `metrics_summary` tool
- [ ] 5.3 Implement `metrics_export` tool
- [ ] 5.4 Implement `metrics_reset` tool (testing only)
- [ ] 5.5 Add formatted table output
- [ ] 5.6 Add chart/graph ASCII visualization (optional)
- [ ] 5.7 Register tools in `register.go`

## 6. Integration
- [ ] 6.1 Add metrics collector middleware to MCP server
- [ ] 6.2 Hook into tool execution pipeline
- [ ] 6.3 Add metrics config to project settings
- [ ] 6.4 Implement opt-out mechanism
- [ ] 6.5 Add logging for metrics system status
- [ ] 6.6 Handle metrics store errors gracefully

## 7. Testing
- [ ] 7.1 Test metrics collection accuracy
- [ ] 7.2 Test retention policy (old entries deleted)
- [ ] 7.3 Test aggregation calculations
- [ ] 7.4 Test export formats (JSON, CSV, Prometheus)
- [ ] 7.5 Test filtering and querying
- [ ] 7.6 Test concurrent writes (thread-safety)
- [ ] 7.7 Performance test (overhead <1ms)

## 8. Validation
- [ ] 8.1 Measure token reduction for simple tasks
- [ ] 8.2 Verify "70-90% reduction" claim
- [ ] 8.3 Benchmark tool execution times
- [ ] 8.4 Validate discovery search accuracy
- [ ] 8.5 Test Prometheus integration

## 9. Documentation
- [ ] 9.1 Document metrics schema
- [ ] 9.2 Add usage examples for each tool
- [ ] 9.3 Document export formats
- [ ] 9.4 Add privacy and opt-out instructions
- [ ] 9.5 Create dashboard examples
- [ ] 9.6 Document Prometheus integration
