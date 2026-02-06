# ADR-009: Time-Series Database Selection

## Status

Accepted

## Context

The observability stack requires a time-series database (TSDB) for metrics storage. Multiple options exist with different architectures, trade-offs, and operational characteristics.

### Requirements

| Requirement | Priority | Notes |
|-------------|----------|-------|
| PromQL compatibility | Must | Existing dashboards and alerts use PromQL |
| Kubernetes-native | Must | Easy deployment, operator support |
| Resource efficient | High | Kind ~8GB RAM, N100 home lab |
| Long-term storage | Medium | Thanos/Mimir for retention beyond 15 days |
| Multi-cluster support | Medium | For home lab, not Kind |
| Learning value | Medium | Resume relevance |

## Options Considered

### Option 1: Prometheus TSDB

**What it is:** The original PromQL database, embedded in Prometheus server.

**Architecture:**
- Local TSDB stored on disk
- Write-ahead log (WAL) for crash recovery
- Block-based storage (2-hour blocks, compacted over time)
- Single-node by design

**Pros:**
- De facto standard for Kubernetes monitoring
- Zero additional complexity
- Massive ecosystem (exporters, dashboards, alerts)
- Well-documented, battle-tested

**Cons:**
- Single-node limits (vertical scaling only)
- 15-day default retention (configurable but limited by disk)
- No native HA (requires Thanos/Mimir for redundancy)
- Federation has scaling limits

**Best for:** Single cluster, standard retention needs, simplicity

**Resource usage:** ~500MB-2GB RAM depending on cardinality

### Option 2: Victoria Metrics

**What it is:** Drop-in Prometheus replacement with better performance.

**Architecture:**
- Custom storage format (not Prometheus TSDB blocks)
- Merge-tree inspired compression
- Single-binary or cluster mode
- vmagent (scraper) + vmselect + vminsert + vmstorage

**Pros:**
- 5-10x better compression than Prometheus
- Lower memory usage for same cardinality
- PromQL compatible (with MetricsQL extensions)
- Simpler long-term retention (built-in)
- Handles high cardinality better

**Cons:**
- Smaller community than Prometheus
- MetricsQL differences can surprise
- Less ecosystem integration (some tools assume Prometheus)
- Different operational model

**Best for:** Resource-constrained environments, high cardinality, long retention

**Resource usage:** ~200MB-1GB RAM (2-5x less than Prometheus for same data)

### Option 3: Grafana Mimir

**What it is:** Grafana's horizontally-scalable, multi-tenant TSDB.

**Architecture:**
- Distributed components (distributor, ingester, querier, compactor, store-gateway)
- Object storage backend (S3/GCS required)
- Multi-tenant by design
- Evolved from Cortex

**Pros:**
- Horizontally scalable (handle any load)
- Multi-tenant (team isolation)
- Grafana-native integration
- Long-term storage built-in

**Cons:**
- Complex to operate (many components)
- Requires object storage
- Overkill for single cluster
- Higher resource baseline

**Best for:** Large scale, multi-team, cloud-native deployments

**Resource usage:** 2-4GB+ RAM minimum (distributed system overhead)

### Option 4: Thanos (with Prometheus)

**What it is:** Not a TSDB replacement, but extends Prometheus with long-term storage.

**Architecture:**
- Sidecar uploads Prometheus blocks to S3
- Store Gateway serves historical data
- Query component federates across Sidecars + Store
- Compactor handles downsampling

**Pros:**
- Extends existing Prometheus (not replace)
- Long-term retention via object storage
- Global query across multiple Prometheus
- Downsampling for cost efficiency

**Cons:**
- Eventual consistency (queries may miss recent data)
- Adds operational complexity
- Requires object storage
- Multiple components to manage

**Best for:** Keep Prometheus, add long-term storage and multi-cluster

**Resource usage:** ~500MB-1GB for components (on top of Prometheus)

### Option 5: InfluxDB

**What it is:** Purpose-built TSDB with its own query language.

**Architecture:**
- Time-structured merge tree (TSM)
- Tags (indexed) vs fields (not indexed)
- Flux query language (InfluxDB 2.x)
- Retention policies and continuous queries

**Pros:**
- Purpose-built for time-series
- Strong ecosystem for IoT
- Good documentation
- Flux is powerful (but different)

**Cons:**
- Different query language (not PromQL)
- Not Kubernetes-native ecosystem
- Can't reuse Prometheus dashboards/alerts
- OSS vs Cloud licensing complexity

**Best for:** Greenfield projects, IoT, teams without PromQL investment

**Resource usage:** ~500MB-2GB RAM

## Technical Comparison

### Architecture Summary

| Solution | Storage Model | Scalability | HA Model | Query Language |
|----------|---------------|-------------|----------|----------------|
| **Prometheus** | Local blocks | Vertical | External (Thanos/Mimir) | PromQL |
| **Victoria Metrics** | Merge-tree | Vertical + Cluster | Built-in cluster mode | MetricsQL (PromQL compat) |
| **Mimir** | Object storage | Horizontal | Distributed | PromQL |
| **Thanos** | Prometheus + S3 | Horizontal | Distributed | PromQL |
| **InfluxDB** | TSM | Vertical + Enterprise | Enterprise only | Flux/InfluxQL |

### Resource Comparison (10k active series)

| Solution | RAM | CPU | Storage Efficiency |
|----------|-----|-----|-------------------|
| **Prometheus** | ~500MB | Low | Baseline |
| **Victoria Metrics** | ~150MB | Low | 5-10x better |
| **Mimir** | ~2GB+ | Medium | Similar to Prometheus |
| **Thanos** | ~1GB+ | Low-Medium | Prometheus + S3 overhead |

### What Matters for This Lab

| Concern | Best Option | Notes |
|---------|-------------|-------|
| Learning standard tooling | Prometheus | Industry standard, transferable skills |
| Resource efficiency | Victoria Metrics | Best for Kind/N100 constraints |
| Long-term retention | Thanos or Victoria Metrics | Both support S3 backend |
| Multi-cluster | Thanos or Mimir | Distributed query layer |
| Simplicity | Prometheus | Single binary, no dependencies |

## Decision

**Use Prometheus TSDB** as the primary metrics database, with **Thanos** for long-term storage when needed.

### Why Prometheus + Thanos

| Factor | Reasoning |
|--------|-----------|
| **Industry standard** | Every Kubernetes environment uses Prometheus; transferable skills |
| **Ecosystem** | kube-prometheus-stack, ServiceMonitors, PodMonitors, PrometheusRules |
| **Incremental complexity** | Start simple (Prometheus), add Thanos when retention matters |
| **Learning progression** | Understand the baseline before alternatives |

### TSDB Comparison Tutorial

Victoria Metrics and Mimir deserve exploration as alternatives:
- Genuinely more efficient (measurable difference on N100)
- Drop-in replacement shows portability
- Different architectural trade-offs worth understanding

**Implemented:** See `experiments/tsdb-comparison/` - hands-on tutorial comparing all three TSDBs under identical workload with cardinality scaling from 1k to 50k series.

### Environment Strategy

| Environment | Solution | Rationale |
|-------------|----------|-----------|
| **Kind (tutorials)** | Prometheus | Standard tooling, learn the baseline |
| **Kind (VM comparison)** | Victoria Metrics | See the efficiency difference |
| **Talos (home lab)** | Prometheus + Thanos | Long-term retention, HA |
| **Cloud** | Managed (Grafana Cloud, Amazon Managed Prometheus) | Operational simplicity |

## Consequences

### Positive
- Learning industry-standard tooling
- Massive ecosystem of exporters, dashboards, alerts
- Clear upgrade path (Thanos when needed)
- Skills transfer to any Kubernetes environment

### Negative
- Higher resource usage than Victoria Metrics
- Requires Thanos for long-term retention (added complexity)
- Single-node limitations without federation

### Migration Path
- PromQL dashboards work with Victoria Metrics, Thanos, Mimir
- Can swap backends without changing queries
- Victoria Metrics accepts Prometheus remote_write

## References

- [Prometheus Storage](https://prometheus.io/docs/prometheus/latest/storage/)
- [Victoria Metrics Comparison](https://docs.victoriametrics.com/single-server-victoriametrics/#prominent-features)
- [Thanos Architecture](https://thanos.io/tip/thanos/design.md/)
- [Mimir Architecture](https://grafana.com/docs/mimir/latest/references/architecture/)
- [Victoria Metrics vs Prometheus Benchmarks](https://victoriametrics.github.io/benchmark/)
- [TSDB Comparison Tutorial](../../experiments/tsdb-comparison/README.md) - Hands-on comparison in this lab
