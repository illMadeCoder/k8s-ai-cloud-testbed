# ADR-010: Logging Stack Selection

## Status

Accepted

## Context

The observability stack requires centralized logging for troubleshooting, debugging, and audit purposes. Multiple options exist with fundamentally different architectures and trade-offs.

### Requirements

| Requirement | Priority | Notes |
|-------------|----------|-------|
| Kubernetes-native | Must | Easy deployment, understands pods/namespaces |
| Grafana integration | High | Single pane of glass with metrics |
| Resource efficient | High | Kind ~8GB RAM, N100 home lab |
| Structured log support | High | JSON logs from modern apps |
| Full-text search | Medium | Nice to have, not critical for most K8s debugging |
| Long-term retention | Medium | S3/SeaweedFS backend for cost-effective storage |
| Learning value | Medium | Resume relevance |

## Options Considered

### Option 1: Grafana Loki

**What it is:** Log aggregation system designed to be cost-effective and easy to operate. Indexes only labels (metadata), not log content.

**Architecture:**
- Promtail/Alloy agents collect logs from nodes
- Logs stored as compressed chunks in object storage
- Only labels are indexed (namespace, pod, container, app)
- Queries filter by labels first, then grep content

**Pros:**
- 10-100x cheaper than full-text indexing (less storage, CPU)
- Native Grafana integration (same UI as metrics)
- LogQL similar to PromQL (consistent query experience)
- Lightweight - runs well on Kind/N100
- Uses same object storage as Thanos/Tempo (SeaweedFS)

**Cons:**
- No full-text index (grep-style search on content)
- Slow for queries without good label selectivity
- Less powerful for log analytics
- Smaller ecosystem than ELK

**Best for:** Teams already using Grafana, resource-constrained environments, label-based debugging

**Resource usage:** ~200-500MB RAM for single-node

### Option 2: ELK Stack (Elasticsearch, Logstash, Kibana)

**What it is:** The original log aggregation stack. Full-text indexes every log line.

**Architecture:**
- Filebeat/Fluentd/Logstash collect and parse logs
- Elasticsearch indexes and stores (inverted index on all fields)
- Kibana for visualization and search
- Optional: Elastic APM for traces

**Pros:**
- Powerful full-text search (find any string instantly)
- Rich query language (KQL, Lucene)
- Kibana has excellent log exploration UX
- Mature ecosystem, extensive documentation
- Elastic APM adds tracing (competing with Tempo/Jaeger)

**Cons:**
- Resource hungry (JVM-based, needs heap)
- Complex to operate (shards, replicas, ILM)
- Expensive storage (indexes everything)
- Separate UI from Grafana (context switching)
- AGPL license concerns for Elasticsearch

**Best for:** Teams needing full-text search, existing ELK expertise, complex log analytics

**Resource usage:** 2-4GB+ RAM minimum (JVM heap)

### Option 3: EFK Stack (Elasticsearch, Fluentd, Kibana)

**What it is:** ELK variant using Fluentd instead of Logstash.

**Architecture:**
- Same as ELK but Fluentd for collection/parsing
- Fluentd is CNCF project, cloud-native
- Often paired with Fluent Bit (lightweight forwarder)

**Pros:**
- Fluentd more Kubernetes-native than Logstash
- Fluent Bit very lightweight for node agents
- Better plugin ecosystem for K8s
- Same powerful Elasticsearch backend

**Cons:**
- Same resource/cost issues as ELK
- Same operational complexity
- Configuration can be verbose

**Best for:** Kubernetes environments wanting ELK power with better collection

**Resource usage:** 2-4GB+ RAM (same as ELK)

### Option 4: Vector + ClickHouse

**What it is:** Modern alternative using Vector (Rust-based collector) with ClickHouse (column-store database).

**Architecture:**
- Vector agents collect and transform logs
- ClickHouse stores in columnar format
- Query via SQL
- Grafana plugin for visualization

**Pros:**
- Very high performance (Rust + columnar storage)
- SQL queries (familiar for many)
- Excellent for log analytics and aggregations
- Cost-effective at scale

**Cons:**
- Less mature ecosystem
- ClickHouse operational complexity
- No native Kubernetes integration like Loki
- Requires more custom setup

**Best for:** High-volume logging, analytics-heavy use cases, SQL preference

**Resource usage:** ~500MB-1GB RAM

### Option 5: Cloud Managed (CloudWatch, Stackdriver, Azure Monitor)

**What it is:** Cloud provider logging services.

**Pros:**
- Zero operational overhead
- Integrated with cloud services
- Auto-scaling

**Cons:**
- Vendor lock-in
- Can be expensive at scale
- Not portable to on-prem/Kind
- Less learning value

**Best for:** Production cloud workloads, teams wanting simplicity

## Technical Comparison

### Architecture Summary

| Solution | Index Strategy | Storage Model | Query Language | UI |
|----------|----------------|---------------|----------------|-----|
| **Loki** | Labels only | Chunks in S3 | LogQL | Grafana |
| **ELK/EFK** | Full-text | Inverted index | KQL/Lucene | Kibana |
| **Vector + ClickHouse** | Columnar | Column store | SQL | Grafana |
| **Cloud Managed** | Varies | Managed | Varies | Native |

### Resource Comparison

| Solution | RAM (minimum) | Storage Efficiency | Query Speed (full-text) |
|----------|---------------|-------------------|------------------------|
| **Loki** | ~300MB | Excellent (10-100x) | Slow (grep) |
| **ELK** | ~2GB+ | Poor (indexes all) | Fast |
| **ClickHouse** | ~500MB | Good | Fast (SQL) |

### What Matters for This Lab

| Concern | Best Option | Notes |
|---------|-------------|-------|
| Resource efficiency | Loki | Runs on Kind/N100 |
| Grafana integration | Loki | Same UI as metrics |
| Full-text search | ELK | If you need it |
| Learning standard tooling | Both | Loki growing, ELK established |
| Object storage reuse | Loki | Uses SeaweedFS like Thanos |

## Decision

**Use Grafana Loki** as the primary logging stack.

### Why Loki

| Factor | Reasoning |
|--------|-----------|
| **Grafana integration** | Single pane of glass - logs next to metrics |
| **Resource efficiency** | Runs comfortably on Kind and N100 |
| **LogQL** | Consistent query experience with PromQL |
| **Object storage** | Reuses SeaweedFS from Section 3.2 |
| **Cost model** | Only indexes labels, not content |

### When to Consider ELK

ELK/EFK makes sense when:
- Full-text search is critical (security logs, audit)
- Team has existing ELK expertise
- Complex log analytics required
- Kibana features needed

### Logging Comparison Tutorial

Similar to TSDB comparison, a hands-on tutorial could demonstrate:
- Loki vs ELK resource usage under same log volume
- Query performance differences (label vs full-text)
- When label-based wins vs when full-text wins

**Potential:** See `experiments/logging-comparison/` (future)

### Environment Strategy

| Environment | Solution | Rationale |
|-------------|----------|-----------|
| **Kind (tutorials)** | Loki | Resource efficient, Grafana native |
| **Talos (home lab)** | Loki | Consistent with Kind, SeaweedFS backend |
| **Cloud (production)** | Evaluate | ELK if full-text needed, else Loki |

## Consequences

### Positive
- Unified observability in Grafana (metrics + logs)
- Low resource footprint
- Simple operations (no JVM tuning, shard management)
- Cost-effective storage with SeaweedFS

### Negative
- No instant full-text search
- Requires good label hygiene
- Less powerful for log analytics
- Smaller community than ELK

### Migration Path
- Logs can be dual-shipped to both Loki and ELK during evaluation
- Fluentd/Fluent Bit can output to multiple destinations
- Grafana can query both (Loki + Elasticsearch datasources)

## Log Collection Strategy

Regardless of backend, use **Promtail** or **Grafana Alloy** for collection:
- Native Kubernetes service discovery
- Automatic pod/namespace labels
- Pipeline stages for parsing

For ELK compatibility, **Fluent Bit** is the lightweight alternative:
- CNCF project
- Multi-output support
- Lower resource usage than Fluentd

## References

- [Grafana Loki Documentation](https://grafana.com/docs/loki/latest/)
- [Loki vs Elasticsearch](https://grafana.com/blog/2020/05/12/an-only-slightly-opinionated-comparison-of-elasticsearch-and-grafana-loki/)
- [ELK Stack Documentation](https://www.elastic.co/guide/index.html)
- [Vector Documentation](https://vector.dev/docs/)
- [ClickHouse for Logs](https://clickhouse.com/docs/en/guides/developer/working-with-logs)
- [ADR-008: Object Storage Selection](ADR-008-object-storage.md) - SeaweedFS for Loki chunks
- [ADR-009: TSDB Selection](ADR-009-tsdb-selection.md) - Similar decision pattern
