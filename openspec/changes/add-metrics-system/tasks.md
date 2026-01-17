# Tasks: Add Metrics and Performance Monitoring

## 1. Metrics Storage
- [x] 1.1 Create `internal/metrics/store.go` ✓ 2026-01-15
- [x] 1.2 Define `Metric` struct with all fields ✓ 2026-01-15
- [x] 1.3 Implement in-memory ring buffer (last 1000 entries) ✓ 2026-01-15
- [x] 1.4 Add persistent storage option (SQLite or JSON file) ✓ 2026-01-15
- [x] 1.5 Implement retention policy (delete old entries) ✓ 2026-01-15
- [x] 1.6 Add thread-safe read/write with RWMutex ✓ 2026-01-15

## 2. Metrics Collection
- [x] 2.1 Create `internal/metrics/collector.go` ✓ 2026-01-15
- [x] 2.2 Implement pre-execution hook ✓ 2026-01-15
- [x] 2.3 Implement post-execution hook ✓ 2026-01-15
- [x] 2.4 Add async write to avoid blocking ✓ 2026-01-15
- [x] 2.5 Extract token counts from MCP responses ✓ 2026-01-15
- [x] 2.6 Measure execution duration ✓ 2026-01-15
- [x] 2.7 Track success/failure status ✓ 2026-01-15

## 3. Aggregation and Statistics
- [x] 3.1 Create `internal/metrics/aggregator.go` ✓ 2026-01-15
- [x] 3.2 Implement average calculation ✓ 2026-01-15
- [x] 3.3 Implement percentile calculation (p50, p95, p99) ✓ 2026-01-15
- [x] 3.4 Add time-based grouping (hour, day, week) ✓ 2026-01-15
- [x] 3.5 Add filtering by tool name ✓ 2026-01-15
- [x] 3.6 Add filtering by session ID ✓ 2026-01-15
- [x] 3.7 Calculate success rate percentage ✓ 2026-01-15

## 4. Export Capabilities
- [x] 4.1 Create `internal/metrics/exporter.go` ✓ 2026-01-15
- [x] 4.2 Implement JSON export ✓ 2026-01-15
- [x] 4.3 Implement CSV export ✓ 2026-01-15
- [x] 4.4 Implement Prometheus format export ✓ 2026-01-15
- [x] 4.5 Add HTTP endpoint for Prometheus scraping (optional) ✓ 2026-01-15 (skipped - optional)
- [x] 4.6 Add file export with timestamp ✓ 2026-01-15

## 5. MCP Tools Implementation
- [x] 5.1 Implement `metrics_show` tool ✓ 2026-01-17
- [x] 5.2 Implement `metrics_summary` tool ✓ 2026-01-17
- [x] 5.3 Implement `metrics_export` tool ✓ 2026-01-17
- [x] 5.4 Implement `metrics_reset` tool (testing only) ✓ 2026-01-17
- [x] 5.5 Add formatted table output ✓ 2026-01-17
- [ ] 5.6 Add chart/graph ASCII visualization (optional)
- [ ] 5.7 Register tools in `register.go`

## 6. Integration
- [x] 6.1 Add metrics collector middleware to MCP server ✓ 2026-01-17
- [x] 6.2 Hook into tool execution pipeline ✓ 2026-01-17
- [ ] 6.3 Add metrics config to project settings
- [x] 6.4 Implement opt-out mechanism ✓ 2026-01-17
- [x] 6.5 Add logging for metrics system status ✓ 2026-01-17
- [x] 6.6 Handle metrics store errors gracefully ✓ 2026-01-17

## 7. Testing
- [x] 7.1 Test metrics collection accuracy ✓ 2026-01-15
- [x] 7.2 Test retention policy (old entries deleted) ✓
- [x] 7.3 Test aggregation calculations ✓ 2026-01-15
- [x] 7.4 Test export formats (JSON, CSV, Prometheus) ✓ 2026-01-15
- [x] 7.5 Test filtering and querying ✓
- [x] 7.6 Test concurrent writes (thread-safety) ✓
- [x] 7.7 Performance test (overhead <1ms) ✓ 2026-01-15

## 8. Validation
- [ ] 8.1 Measure token reduction for simple tasks
- [ ] 8.2 Verify "70-90% reduction" claim
- [ ] 8.3 Benchmark tool execution times
- [ ] 8.4 Validate discovery search accuracy
- [ ] 8.5 Test Prometheus integration

## 9. Documentation
- [x] 9.1 Document metrics schema ✓ 2026-01-17
- [x] 9.2 Add usage examples for each tool ✓ 2026-01-17
- [x] 9.3 Document export formats ✓ 2026-01-17
- [ ] 9.4 Add privacy and opt-out instructions
- [ ] 9.5 Create dashboard examples
- [ ] 9.6 Document Prometheus integration
