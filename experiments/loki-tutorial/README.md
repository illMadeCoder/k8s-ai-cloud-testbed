# Loki Tutorial - LogQL Mastery

Learn log aggregation with Grafana Loki and LogQL.

## Quick Start

```bash
task kind:conduct -- loki-tutorial
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         GRAFANA                             │
│                    (Visualization)                          │
│                          │                                  │
│                     LogQL queries                           │
│                          ▼                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                       LOKI                           │   │
│  │              (Log Storage & Query)                   │   │
│  │                                                      │   │
│  │  • Label-indexed (not full-text)                     │   │
│  │  • Chunks stored on filesystem                       │   │
│  │  • LogQL query language                              │   │
│  └─────────────────────────────────────────────────────┘   │
│                          ▲                                  │
│                     Push logs                               │
│                          │                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                    PROMTAIL                          │   │
│  │               (Log Collection)                       │   │
│  │                                                      │   │
│  │  • DaemonSet on each node                            │   │
│  │  • Tails container logs                              │   │
│  │  • Adds Kubernetes labels                            │   │
│  └─────────────────────────────────────────────────────┘   │
│                          ▲                                  │
│                    Container logs                           │
│                          │                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                  LOG GENERATOR                       │   │
│  │              (Demo Application)                      │   │
│  │                                                      │   │
│  │  • Structured JSON logs                              │   │
│  │  • Configurable rate & cardinality                   │   │
│  │  • Multiple log levels (debug/info/warn/error)       │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## What You'll Learn

### Module 1: Label Selectors

Labels are the foundation of LogQL. Loki indexes labels, not log content.

```logql
# Basic selector
{app="logging-app"}

# Multiple labels
{app="logging-app", namespace="demo"}

# Regex matching
{app=~"logging.*"}

# Not equal
{namespace!="kube-system"}
```

### Module 2: Line Filters

Filter log content after selecting by labels.

```logql
# Contains (case-sensitive)
{app="logging-app"} |= "error"

# Does not contain
{app="logging-app"} != "debug"

# Regex filter
{app="logging-app"} |~ "service-[0-9]+"

# Case-insensitive
{app="logging-app"} |~ "(?i)error"
```

### Module 3: JSON Parser

Parse structured logs to filter on fields.

```logql
# Parse JSON
{app="logging-app"} | json

# Filter on parsed field
{app="logging-app"} | json | level="error"

# Numeric comparison
{app="logging-app"} | json | status >= 500

# Multiple filters
{app="logging-app"} | json | level="error" | service="service-1"

# Format output
{app="logging-app"} | json | line_format "{{.service}}: {{.message}}"
```

### Module 4: Label Filters

After parsing, filter on extracted labels.

```logql
# String equality
{app="logging-app"} | json | level="error"

# Numeric comparison
{app="logging-app"} | json | duration_ms > 1000

# Regex on extracted field
{app="logging-app"} | json | endpoint=~"/api/v1/.*"
```

### Module 5: Metric Queries

Convert logs to metrics for dashboards and alerts.

```logql
# Count logs per minute
sum(count_over_time({app="logging-app"} [1m]))

# Errors by service
sum by (service) (count_over_time({app="logging-app"} | json | level="error" [5m]))

# Error rate percentage
sum(count_over_time({app="logging-app"} | json | level="error" [5m]))
/
sum(count_over_time({app="logging-app"} [5m]))
* 100

# P95 response time from logs
quantile_over_time(0.95, {app="logging-app"} | json | unwrap duration_ms [5m]) by ()

# Bytes processed
sum(bytes_over_time({app="logging-app"} [1h]))
```

## Log Generator Output

The demo app produces structured JSON logs:

```json
{
  "timestamp": "2024-01-15T10:23:45.123456789Z",
  "level": "error",
  "service": "service-3",
  "endpoint": "/api/v1/orders",
  "method": "POST",
  "status": 500,
  "duration_ms": 1234,
  "trace_id": "abc123def456...",
  "message": "Database connection failed"
}
```

Control the generator via API:

```bash
# Get current config
curl http://$APP_IP:8080/config

# Set log rate (low=10/s, medium=100/s, high=1000/s)
curl -X POST http://$APP_IP:8080/config \
  -H "Content-Type: application/json" \
  -d '{"rate": "high"}'

# Set cardinality (low=5 services, medium=20, high=100)
curl -X POST http://$APP_IP:8080/config \
  -H "Content-Type: application/json" \
  -d '{"cardinality": "medium"}'
```

## Dashboard

The tutorial includes a Grafana dashboard with:

| Panel | Query | Purpose |
|-------|-------|---------|
| Log Volume | `sum(count_over_time({app="logging-app"} [1m]))` | Current ingestion rate |
| Error Rate | Error count / Total count * 100 | Error percentage |
| P95 Duration | `quantile_over_time(0.95, ... unwrap duration_ms)` | Performance from logs |
| Logs by Level | `sum by (level) (count_over_time(...))` | Volume breakdown |
| Live Logs | `{app="logging-app"} \| json` | Streaming log view |

## Loki vs Full-Text Search

| Aspect | Loki | Elasticsearch |
|--------|------|---------------|
| **Indexing** | Labels only | Full content |
| **Query speed (labels)** | Fast | Fast |
| **Query speed (content)** | Slow (grep) | Fast (inverted index) |
| **Storage** | 10-100x smaller | Large (indexes everything) |
| **RAM usage** | ~300MB | ~2GB+ (JVM) |
| **Best for** | K8s debugging, known patterns | Ad-hoc search, analytics |

## Next Steps

- **elk-tutorial**: Learn Elasticsearch + Kibana for comparison
- **logging-comparison**: Run both stacks side-by-side

## References

- [Loki Documentation](https://grafana.com/docs/loki/latest/)
- [LogQL Documentation](https://grafana.com/docs/loki/latest/query/)
- [Promtail Configuration](https://grafana.com/docs/loki/latest/send-data/promtail/)
- [ADR-010: Logging Stack Selection](../../docs/adrs/ADR-010-logging-stack-selection.md)
- [ADR-011: Observability Architecture](../../docs/adrs/ADR-011-observability-architecture.md)
