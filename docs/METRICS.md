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
