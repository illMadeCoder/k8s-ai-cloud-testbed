# Observability Cost Management Tutorial

Learn to analyze and optimize observability costs in Kubernetes.

## Overview

Observability can become one of the largest infrastructure costs. This tutorial teaches you to:
- Understand cardinality and its impact on Prometheus
- Analyze log volume and optimize Loki costs
- Implement retention and sampling strategies
- Build cost analysis dashboards

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      COST ANALYSIS                               │
│                                                                  │
│                    Grafana Dashboard                             │
│                    (cost metrics)                                │
│                          │                                       │
│         ┌────────────────┼────────────────┐                     │
│         │                │                │                     │
│         ▼                ▼                ▼                     │
│    ┌─────────┐     ┌─────────┐     ┌─────────┐                 │
│    │PROMETHEUS│     │  LOKI   │     │ STORAGE │                 │
│    │tsdb_*   │     │ volume  │     │  disk   │                 │
│    └────▲────┘     └────▲────┘     └─────────┘                 │
│         │               │                                       │
│    ServiceMonitor   Promtail                                    │
│         │               │                                       │
│    ┌────┴────┐     ┌────┴────┐                                 │
│    │Cardinality│   │   Log    │                                 │
│    │Generator │   │Generator │                                 │
│    └──────────┘   └──────────┘                                 │
│                                                                  │
│              COST SIMULATION APPS                                │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

- Completed `prometheus-tutorial`
- Completed `loki-tutorial` (recommended)

## Quick Start

```bash
# Deploy the tutorial
task kind:conduct -- observability-cost-tutorial

# Wait for pods
kubectl get pods -n observability-cost -w

# Access Grafana
kubectl port-forward -n observability-cost svc/kube-prometheus-stack-grafana 3000:80 &
open http://localhost:3000  # admin/admin
```

## Tutorial Modules

### Module 1: Observability Cost Drivers

**What drives observability costs:**

| Component | Cost Driver | Impact |
|-----------|-------------|--------|
| Prometheus | Cardinality (unique series) | Memory, storage, query time |
| Loki | Log volume (bytes/sec) | Storage, ingestion CPU |
| Tempo | Trace volume | Storage, sampling overhead |

**Rule of thumb:**
- 1M time series ≈ 1-2GB RAM for Prometheus
- 10KB/sec logs ≈ 26GB/month storage in Loki

### Module 2: Metrics Cardinality Analysis

**What is cardinality?**
The number of unique time series. Each unique combination of metric name + labels = 1 series.

```
http_requests_total{method="GET", status="200"}   # Series 1
http_requests_total{method="GET", status="500"}   # Series 2
http_requests_total{method="POST", status="200"}  # Series 3
```

**Find your cardinality:**
```promql
# Total series in Prometheus
prometheus_tsdb_head_series

# Top 10 metrics by series count
topk(10, count by (__name__)({__name__=~".+"}))

# Series per job
count by (job) ({__name__=~".+"})
```

**Cardinality explosion example:**
```
# SAFE: Bounded labels (4 methods × 5 statuses × 4 endpoints = 80 series)
http_requests_total{method, status, endpoint}

# DANGEROUS: Unbounded labels (4 × 5 × 1M users = 20M series!)
http_requests_total{method, status, user_id}
```

### Module 3: Label Best Practices

**Safe labels (bounded cardinality):**
- `method`: GET, POST, PUT, DELETE, PATCH
- `status`: 200, 201, 400, 404, 500
- `endpoint`: /api/users, /api/orders (limited set)
- `region`: us-east-1, us-west-2, eu-west-1
- `env`: prod, staging, dev

**Dangerous labels (unbounded cardinality):**
- `user_id`: Unique per user (millions)
- `request_id`: Unique per request (infinite)
- `email`: Unique per user
- `timestamp`: Changes every second
- `ip_address`: Millions of unique values

**Best practices:**
1. Never use user/request IDs as metric labels
2. Use histograms instead of per-user gauges
3. Aggregate at collection time, not query time
4. Use recording rules for expensive queries

### Module 4: Log Volume Analysis

**Log cost factors:**
1. **Volume**: Bytes ingested per second
2. **Retention**: How long logs are kept
3. **Indexing**: Label cardinality in Loki

**Compare logging strategies:**
```bash
# Verbose logger (~1KB/sec)
kubectl logs -n observability-cost -l log-level=verbose --tail=5

# Efficient logger (~100B/sec)
kubectl logs -n observability-cost -l log-level=efficient --tail=5
```

**Log optimization strategies:**
1. Drop DEBUG logs in production
2. Sample high-volume logs (1 in 100)
3. Use structured logging (JSON)
4. Avoid logging PII
5. Set aggressive retention for noisy sources

**Promtail drop example:**
```yaml
pipeline_stages:
  - drop:
      expression: ".*DEBUG.*"
      drop_counter_reason: debug_dropped
```

### Module 5: Retention & Sampling

**Prometheus retention:**
```yaml
prometheusSpec:
  retention: 15d        # Time-based
  retentionSize: 50GB   # Size-based (whichever hits first)
```

**Loki retention:**
```yaml
limits_config:
  retention_period: 168h  # 7 days
```

**Sampling strategies:**

| Strategy | Pros | Cons |
|----------|------|------|
| No sampling | Complete data | Highest cost |
| Head sampling | Predictable | Miss rare events |
| Tail sampling | Catch errors | Higher latency |
| Probabilistic | Simple | May miss patterns |

### Module 6: Cost Dashboard

The included dashboard shows:
1. **Total Time Series**: Current cardinality
2. **TSDB Size**: Prometheus storage usage
3. **Memory Usage**: Prometheus memory consumption
4. **Log Ingestion Rate**: Loki bytes/sec
5. **Cardinality Trend**: Series growth over time

## Cost Optimization Checklist

### Metrics
- [ ] Monitor `prometheus_tsdb_head_series` for growth
- [ ] Alert when cardinality exceeds threshold
- [ ] Remove unused metrics (check `prometheus_tsdb_head_series_created_total`)
- [ ] Use recording rules for expensive aggregations
- [ ] Consider downsampling historical data

### Logs
- [ ] Drop DEBUG logs in production
- [ ] Set retention per log source
- [ ] Sample high-volume logs
- [ ] Use efficient log formats
- [ ] Avoid logging in hot paths

### General
- [ ] Set resource limits on observability components
- [ ] Use object storage for long-term retention
- [ ] Implement cost allocation per namespace
- [ ] Review and prune unused dashboards/alerts

## Exercises

### Exercise 1: Find Cardinality Explosion

```promql
# Query to find the "bad" high-cardinality metric
count(bad_requests_total)

# Compare to the "good" bounded metric
count(good_requests_total)
```

### Exercise 2: Calculate Storage Cost

```promql
# Current TSDB size
prometheus_tsdb_head_chunks_storage_size_bytes

# Growth rate (bytes/hour)
rate(prometheus_tsdb_head_chunks_storage_size_bytes[1h]) * 3600
```

### Exercise 3: Log Volume by Source

```logql
# In Loki - volume by app
sum by (app) (bytes_over_time({namespace="observability-cost"}[1h]))
```

## Troubleshooting

### Prometheus OOM

```bash
# Check memory usage
kubectl top pod -n observability-cost -l app.kubernetes.io/name=prometheus

# Find cardinality culprits
curl http://localhost:9090/api/v1/status/tsdb | jq '.data.seriesCountByMetricName | sort_by(-.value) | .[0:10]'
```

### Loki disk full

```bash
# Check storage
kubectl exec -n observability-cost loki-0 -- df -h /var/loki

# Force compaction
kubectl exec -n observability-cost loki-0 -- wget -qO- http://localhost:3100/compactor/ring
```

## Resources

- [Prometheus Cardinality](https://prometheus.io/docs/practices/naming/#labels)
- [Loki Best Practices](https://grafana.com/docs/loki/latest/best-practices/)
- [FinOps for Observability](https://www.finops.org/)
- [Google SRE - Monitoring](https://sre.google/sre-book/monitoring-distributed-systems/)
