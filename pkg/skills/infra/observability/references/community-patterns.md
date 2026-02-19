## Three Pillars
1. **Logs**: Structured events with context (what happened)
2. **Metrics**: Aggregated numerical measurements (how much/how fast)
3. **Traces**: Request flow across services (where time is spent)

## Structured Logging
```json
{
  "timestamp": "2025-01-15T10:30:00Z",
  "level": "error",
  "message": "Failed to process order",
  "service": "order-service",
  "trace_id": "abc123",
  "span_id": "def456",
  "order_id": "ord-789",
  "error": "payment gateway timeout",
  "duration_ms": 5003
}
```
- Use structured JSON in production
- Include trace/span IDs for correlation
- Log at appropriate levels: DEBUG, INFO, WARN, ERROR
- Never log sensitive data (PII, credentials, tokens)
- Use sampling for high-volume debug logs

## OpenTelemetry
- Use OTel SDK as the standard instrumentation layer
- Auto-instrument HTTP clients, databases, message queues
- Add custom spans for business-critical operations
- Export to Jaeger, Tempo, or cloud providers
- Propagate context through HTTP headers (`traceparent`)

## Metrics (Prometheus)
- **Counter**: monotonically increasing (requests_total, errors_total)
- **Gauge**: current value (active_connections, queue_size)
- **Histogram**: distribution (request_duration_seconds, response_size_bytes)
- Use labels for dimensions, but keep cardinality low
- Follow naming conventions: `<namespace>_<name>_<unit>`

## Alerting
- Alert on symptoms (latency, errors), not causes
- Use SLOs (Service Level Objectives) to define thresholds
- Multi-window alerting: short window for severity, long window for burn rate
- Include runbooks in alert metadata
- Avoid alert fatigue — every alert should be actionable

## Health Checks
- `/healthz` — liveness: is the process running?
- `/readyz` — readiness: can it serve traffic?
- Check critical dependencies in readiness (DB, cache)
- Return structured JSON with component status
