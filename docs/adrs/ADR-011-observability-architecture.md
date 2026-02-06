# ADR-011: Observability Architecture

## Status

Accepted

## Context

This lab requires a complete observability stack covering the three pillars (metrics, logs, traces) plus supporting infrastructure (object storage, SLOs). Individual decisions are documented in:

- [ADR-008: Object Storage](ADR-008-object-storage.md) - SeaweedFS
- [ADR-009: TSDB Selection](ADR-009-tsdb-selection.md) - Prometheus + Thanos
- [ADR-010: Logging Stack](ADR-010-logging-stack-selection.md) - Loki (with ELK comparison)

This ADR provides the holistic view of how components integrate across layers and signals.

## Architecture Overview

### Signal Types (Three Pillars + SLOs)

| Signal | Purpose | Query Pattern | Cardinality |
|--------|---------|---------------|-------------|
| **Metrics** | What is happening (aggregates) | Time-series queries, aggregations | Low-medium (labels) |
| **Logs** | Why it happened (context) | Full-text or label search | High (every event) |
| **Traces** | How it flowed (causality) | Trace ID lookup, service graphs | Medium (sampled) |
| **SLOs** | Are we meeting objectives | Error budget burn rate | Derived from metrics |

### Layers

| Layer | Function | Components |
|-------|----------|------------|
| **Instrumentation** | Emit signals from applications | OTel SDK, Prometheus client libs, structured logging |
| **Collection** | Gather signals from sources | Prometheus scrape, Fluent Bit, OTel Collector |
| **Processing** | Transform, enrich, route | Fluent Bit pipelines, OTel processors |
| **Storage** | Persist signals | Prometheus TSDB, Loki, Tempo, Elasticsearch |
| **Long-term Storage** | Retain beyond local capacity | Thanos + SeaweedFS, Loki chunks |
| **Query** | Retrieve and analyze | PromQL, LogQL, TraceQL, KQL |
| **Visualization** | Present to humans | Grafana, Kibana |
| **Alerting** | Notify on conditions | Alertmanager, Grafana Alerting |

## Component Map

### By Signal and Layer

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                 VISUALIZATION                                    │
│                                                                                  │
│    ┌─────────────────────────────────────┐    ┌─────────────────────────────┐   │
│    │              GRAFANA                │    │          KIBANA             │   │
│    │  • Metrics dashboards               │    │  • Log exploration          │   │
│    │  • Log panels (Loki)                │    │  • Full-text search         │   │
│    │  • Trace views (Tempo)              │    │  • ELK-native alerts        │   │
│    │  • SLO dashboards (Sloth/Pyrra)     │    │                             │   │
│    │  • Unified alerting                 │    │                             │   │
│    └─────────────────────────────────────┘    └─────────────────────────────┘   │
│                        PRIMARY                         ELK COMPARISON            │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
        ┌───────────────────────────────┼───────────────────────────────┐
        │                               │                               │
        ▼                               ▼                               ▼
┌───────────────────┐     ┌─────────────────────────┐     ┌───────────────────┐
│      METRICS      │     │          LOGS           │     │      TRACES       │
├───────────────────┤     ├─────────────────────────┤     ├───────────────────┤
│                   │     │                         │     │                   │
│  ┌─────────────┐  │     │  ┌─────────┐ ┌───────┐  │     │  ┌─────────────┐  │
│  │ Prometheus  │  │     │  │  Loki   │ │  ES   │  │     │  │    Tempo    │  │
│  │    TSDB     │  │     │  │         │ │       │  │     │  │             │  │
│  └──────┬──────┘  │     │  └────┬────┘ └───┬───┘  │     │  └──────┬──────┘  │
│         │         │     │       │          │      │     │         │         │
│  ┌──────▼──────┐  │     │       │          │      │     │         │         │
│  │   Thanos    │  │     │       │          │      │     │         │         │
│  │  (Query +   │  │     │       │          │      │     │         │         │
│  │   Store)    │  │     │       │          │      │     │         │         │
│  └──────┬──────┘  │     │       │          │      │     │         │         │
│         │         │     │       │          │      │     │         │         │
│  ┌──────▼──────┐  │     │  ┌────▼────┐     │      │     │  ┌──────▼──────┐  │
│  │  Victoria   │  │     │  │  Loki   │     │      │     │  │   Jaeger    │  │
│  │  Metrics    │  │     │  │ chunks  │     │      │     │  │    (alt)    │  │
│  │ (compare)   │  │     │  └────┬────┘     │      │     │  └─────────────┘  │
│  └─────────────┘  │     │       │          │      │     │                   │
└────────┬──────────┘     └───────┼──────────┼──────┘     └─────────┬─────────┘
         │                        │          │                      │
         │                        │     ES local                    │
         ▼                        ▼     storage                     ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              OBJECT STORAGE                                      │
│                                                                                  │
│                              ┌─────────────┐                                     │
│                              │  SeaweedFS  │                                     │
│                              │   (S3 API)  │                                     │
│                              └─────────────┘                                     │
│                                     │                                            │
│           ┌─────────────────────────┼─────────────────────────┐                  │
│           │                         │                         │                  │
│    ┌──────▼──────┐           ┌──────▼──────┐           ┌──────▼──────┐          │
│    │thanos-blocks│           │ loki-chunks │           │tempo-traces │          │
│    │  (metrics)  │           │   (logs)    │           │  (traces)   │          │
│    └─────────────┘           └─────────────┘           └─────────────┘          │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        ▲
                                        │
┌─────────────────────────────────────────────────────────────────────────────────┐
│                               COLLECTION                                         │
│                                                                                  │
│  ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────────────────┐   │
│  │   Prometheus    │   │   Fluent Bit    │   │   OpenTelemetry Collector   │   │
│  │    (scrape)     │   │   (DaemonSet)   │   │        (Deployment)         │   │
│  │                 │   │                 │   │                             │   │
│  │ • ServiceMonitor│   │ • Tail logs     │   │ • OTLP receiver             │   │
│  │ • PodMonitor    │   │ • K8s metadata  │   │ • Batch processor           │   │
│  │ • Probe         │   │ • Parse JSON    │   │ • Tempo exporter            │   │
│  │                 │   │ • Multi-output  │   │ • Prometheus exporter       │   │
│  │                 │   │   (Loki + ES)   │   │ • Loki exporter (optional)  │   │
│  └────────┬────────┘   └────────┬────────┘   └──────────────┬──────────────┘   │
│           │                     │                           │                   │
│           │     ┌───────────────┴───────────────┐           │                   │
│           │     │                               │           │                   │
│           ▼     ▼                               ▼           ▼                   │
│      Prometheus  Loki                     Elasticsearch   Tempo                 │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        ▲
                                        │
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            INSTRUMENTATION                                       │
│                                                                                  │
│  ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────────────────┐   │
│  │ Prometheus libs │   │ Structured JSON │   │     OpenTelemetry SDK       │   │
│  │                 │   │    logging      │   │                             │   │
│  │ • promhttp      │   │                 │   │ • Auto-instrumentation      │   │
│  │ • client_golang │   │ • zerolog       │   │ • Manual spans              │   │
│  │ • micrometer    │   │ • zap           │   │ • Trace context propagation │   │
│  │                 │   │ • slog          │   │ • Baggage                   │   │
│  └─────────────────┘   └─────────────────┘   └─────────────────────────────┘   │
│                                                                                  │
│                              APPLICATION CODE                                    │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Collector Strategy

### Options Evaluated

| Collector | Metrics | Logs | Traces | Multi-Output | Resource Usage |
|-----------|:-------:|:----:|:------:|:------------:|----------------|
| **Promtail** | ✗ | ✓ | ✗ | Loki only | ~30MB |
| **Fluent Bit** | ✓ | ✓ | ✗ | ✓ (Loki, ES, S3, etc.) | ~15MB |
| **Fluentd** | ✓ | ✓ | ✗ | ✓ | ~100MB |
| **OpenTelemetry Collector** | ✓ | ✓ | ✓ | ✓ (OTLP ecosystem) | ~50-100MB |
| **Grafana Alloy** | ✓ | ✓ | ✓ | ✓ (Grafana native) | ~50MB |
| **Vector** | ✓ | ✓ | ✓ | ✓ | ~30MB |

### Decision: Layered Collection

```
Signal Flow:

  Metrics ──► Prometheus scrape ──────────────────────────► Prometheus
                                                                │
                                                                ▼
  Logs ────► Fluent Bit ──┬──► Loki (label-indexed)        Thanos ──► SeaweedFS
                          │
                          └──► Elasticsearch (full-text)

  Traces ──► OTel Collector ──► Tempo ──────────────────────────────► SeaweedFS
                │
                └──► Prometheus (span metrics)
```

**Rationale:**

| Signal | Collector | Why |
|--------|-----------|-----|
| **Metrics** | Prometheus native scrape | ServiceMonitor/PodMonitor ecosystem, pull model |
| **Logs** | Fluent Bit | Lightweight, multi-output for Loki vs ELK comparison |
| **Traces** | OpenTelemetry Collector | OTLP standard, span-to-metrics conversion |

**Future consideration:** Grafana Alloy could unify all three once the stack is stable, but starting with specialized collectors provides clearer learning boundaries.

## Query Languages

| Signal | Language | Syntax Family | Strengths |
|--------|----------|---------------|-----------|
| **Metrics** | PromQL | Functional | Aggregations, rate calculations, histograms |
| **Metrics** | MetricsQL | PromQL superset | WITH expressions, rollup functions |
| **Logs (Loki)** | LogQL | PromQL-inspired | Label filtering, log metrics |
| **Logs (ES)** | KQL/Lucene | Search syntax | Full-text, fuzzy matching |
| **Traces** | TraceQL | Span-oriented | Structural queries across spans |

### Query Correlation

```
┌─────────────────────────────────────────────────────────────────┐
│                     GRAFANA EXPLORE                             │
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │   Metrics   │◄──►│    Logs     │◄──►│   Traces    │         │
│  │   PromQL    │    │   LogQL     │    │  TraceQL    │         │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘         │
│         │                  │                  │                 │
│         │    Exemplars     │    Trace ID      │                 │
│         └──────────────────┼──────────────────┘                 │
│                            │                                    │
│                    Derived Labels                               │
│              (service, namespace, pod)                          │
└─────────────────────────────────────────────────────────────────┘
```

**Correlation mechanisms:**

| From | To | Mechanism |
|------|-----|-----------|
| Metrics → Traces | Exemplars | PromQL exemplar support, histogram buckets link to trace IDs |
| Logs → Traces | Trace ID field | `traceID` label in structured logs |
| Traces → Logs | Span metadata | Query logs by service + time window |
| Traces → Metrics | Span metrics | OTel Collector generates `calls`, `duration` metrics |

## Storage Strategy

### By Retention Tier

| Tier | Duration | Storage | Resolution | Use Case |
|------|----------|---------|------------|----------|
| **Hot** | 0-2 hours | Prometheus WAL | Full | Real-time dashboards |
| **Warm** | 2h-15 days | Prometheus TSDB | Full | Recent troubleshooting |
| **Cold** | 15d-1 year | Thanos + SeaweedFS | Downsampled (5m, 1h) | Capacity planning, trends |
| **Archive** | 1+ years | SeaweedFS (compressed) | 1h | Compliance, audits |

### Object Storage Buckets

| Bucket | Contents | Retention | Size Estimate |
|--------|----------|-----------|---------------|
| `thanos-blocks` | Prometheus TSDB blocks | 1 year | ~10GB/month |
| `loki-chunks` | Compressed log chunks | 30 days | ~50GB/month |
| `tempo-traces` | Trace spans | 7 days | ~5GB/month |
| `velero-backups` | Cluster backups | 30 days | ~2GB/backup |

## Comparison Tutorials

The architecture supports side-by-side comparisons for learning:

### TSDB Comparison (Implemented)

See `experiments/tsdb-comparison/`

| Aspect | Prometheus | Victoria Metrics | Mimir |
|--------|------------|------------------|-------|
| **Architecture** | Single binary | Single or cluster | Distributed |
| **Compression** | Baseline | 5-10x better | Similar to Prometheus |
| **RAM (10k series)** | ~500MB | ~150MB | ~2GB |
| **HA Model** | External (Thanos) | Built-in cluster | Built-in |

### Logging Comparison (Planned)

See `experiments/logging-comparison/`

| Aspect | Loki | Elasticsearch |
|--------|------|---------------|
| **Index strategy** | Labels only | Full-text on all fields |
| **Query speed (content)** | Slow (grep) | Fast (inverted index) |
| **Query speed (labels)** | Fast | Fast |
| **Storage efficiency** | 10-100x better | Baseline |
| **RAM requirement** | ~300MB | ~2GB+ (JVM) |
| **Best for** | Label-based K8s debugging | Full-text search, analytics |

## SLO Integration

SLOs are derived from metrics, not a separate signal:

```
┌─────────────────────────────────────────────────────────────────┐
│                         SLO LAYER                               │
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │    Sloth    │    │    Pyrra    │    │   Grafana   │         │
│  │ (generator) │───►│ (dashboard) │───►│  (alerts)   │         │
│  └──────┬──────┘    └─────────────┘    └─────────────┘         │
│         │                                                       │
│         │ Generates PrometheusRules                             │
│         ▼                                                       │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ slo:sli_error:ratio_rate5m                                  ││
│  │ slo:error_budget:remaining                                  ││
│  │ slo:burn_rate:ratio_rate1h                                  ││
│  └─────────────────────────────────────────────────────────────┘│
│                         │                                       │
│                         ▼                                       │
│                    Prometheus                                   │
└─────────────────────────────────────────────────────────────────┘
```

## Environment Strategy

| Environment | Metrics | Logs | Traces | Object Storage |
|-------------|---------|------|--------|----------------|
| **Kind (tutorials)** | Prometheus | Loki + ES (compare) | Tempo | Filesystem or SeaweedFS |
| **Talos (home lab)** | Prometheus + Thanos | Loki | Tempo | SeaweedFS (persistent) |
| **Cloud (production)** | Managed Prometheus or Mimir | Loki or CloudWatch | Tempo or X-Ray | Native S3/GCS |

## Decision Summary

| Domain | Decision | Rationale |
|--------|----------|-----------|
| **Object Storage** | SeaweedFS | Apache 2.0, lightweight, Haystack architecture |
| **Metrics TSDB** | Prometheus + Thanos | Industry standard, extensible |
| **Metrics Alternative** | Victoria Metrics (comparison) | Learn efficiency trade-offs |
| **Logs Primary** | Loki | Grafana native, label-indexed, lightweight |
| **Logs Alternative** | Elasticsearch (comparison) | Learn full-text indexing trade-offs |
| **Log Collector** | Fluent Bit | Multi-output, lightweight, CNCF |
| **Traces** | Tempo | Grafana native, SeaweedFS backend |
| **Trace Collector** | OpenTelemetry Collector | OTLP standard |
| **SLOs** | Sloth + Pyrra | Generate recording rules, dashboards |
| **Visualization** | Grafana (primary), Kibana (ELK) | Single pane of glass |

## Consequences

### Positive

- **Unified visualization** - Grafana for metrics, logs, traces in one UI
- **Consistent storage** - SeaweedFS backend for all long-term data
- **Learning comparisons** - Side-by-side tutorials (TSDB, logging)
- **Resource efficient** - Stack runs on Kind (~8GB) and N100 home lab
- **Standard protocols** - PromQL, OTLP, S3 API are transferable skills

### Negative

- **Multiple query languages** - PromQL, LogQL, TraceQL, KQL (for ELK comparison)
- **Comparison complexity** - Running both Loki and ELK increases resource usage
- **No single collector** - Prometheus + Fluent Bit + OTel Collector (could unify with Alloy later)

### Migration Path

1. **Collector unification** - Grafana Alloy can replace Promtail + OTel Collector
2. **Cloud migration** - All components use standard protocols (S3, OTLP, PromQL)
3. **Scale out** - Prometheus → Mimir, single Loki → Loki cluster

## References

- [ADR-008: Object Storage Selection](ADR-008-object-storage.md)
- [ADR-009: TSDB Selection](ADR-009-tsdb-selection.md)
- [ADR-010: Logging Stack Selection](ADR-010-logging-stack-selection.md)
- [Phase 3: Observability Roadmap](../roadmap/phase-03-observability.md)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Grafana LGTM Stack](https://grafana.com/oss/lgtm-stack/)
- [Three Pillars of Observability](https://www.oreilly.com/library/view/distributed-systems-observability/9781492033431/ch04.html)
