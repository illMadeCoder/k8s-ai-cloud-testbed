# TSDB Comparison: The Cardinality Crisis

**Time**: 45-60 minutes
**Difficulty**: Intermediate
**Prerequisites**: Basic Prometheus knowledge

## The Mission

The USS Kubernetes is experiencing a sensor data explosion. Thousands of new sensors are coming online, and the observability system is struggling. Engineering has been tasked with evaluating three time-series database options:

1. **Prometheus** - The current standard
2. **Victoria Metrics** - The efficient alternative
3. **Mimir** - The distributed powerhouse

Your mission: Deploy all three, stress-test them, and determine which is best for different scenarios.

## Learning Objectives

By the end of this tutorial, you will:

1. Understand the architectural differences between TSDBs
2. Measure real resource usage under identical workloads
3. Know when to choose Prometheus, Victoria Metrics, or Mimir
4. Be able to articulate the trade-offs in job interviews

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Target Cluster                               │
│                                                                      │
│  ┌──────────────────┐                                               │
│  │   Cardinality    │     All three TSDBs scrape the same metrics   │
│  │   Generator      │◄────────────────────────────────────────────┐ │
│  │   (1k-50k series)│                                             │ │
│  └──────────────────┘                                             │ │
│           │                                                        │ │
│           ▼                                                        │ │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐            │ │
│  │  Prometheus  │  │   Victoria   │  │    Mimir     │            │ │
│  │    (local)   │  │   Metrics    │  │ (distributed)│            │ │
│  │              │  │  (efficient) │  │              │            │ │
│  └──────────────┘  └──────────────┘  └──────┬───────┘            │ │
│                                              │                    │ │
│  ┌──────────────────────────────────────────┼────────────────────┘ │
│  │                    Grafana               │                      │
│  │  3 datasources, comparison dashboard     │                      │
│  └──────────────────────────────────────────┘                      │
│                                              │                      │
│                                        ┌─────┴─────┐               │
│                                        │   MinIO   │               │
│                                        │ (S3 for   │               │
│                                        │  Mimir)   │               │
│                                        └───────────┘               │
└─────────────────────────────────────────────────────────────────────┘
```

## Part 1: Deploy the Tutorial

```bash
# Deploy the TSDB comparison tutorial
task kind:up -- tsdb-comparison

# Wait for all components to be ready (3-5 minutes)
kubectl get pods -A -w
```

Once deployed, you'll have:
- Prometheus in `monitoring` namespace
- Victoria Metrics in `victoria-metrics` namespace
- Mimir in `mimir` namespace
- MinIO in `minio` namespace (S3 for Mimir)
- Cardinality Generator in `tsdb-comparison` namespace

## Part 2: Explore the TSDBs

### Access the UIs

After deployment, get the service IPs:

```bash
# Grafana (all dashboards)
kubectl get svc -n monitoring prometheus-grafana -o jsonpath='{.status.loadBalancer.ingress[0].ip}'

# Prometheus UI
kubectl get svc -n monitoring prometheus-kube-prometheus-prometheus -o jsonpath='{.status.loadBalancer.ingress[0].ip}'

# MinIO Console (for Mimir storage)
kubectl get svc -n minio minio-external -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

**Grafana credentials**: admin / admin

### Explore the Comparison Dashboard

1. Open Grafana
2. Navigate to Dashboards → TSDB Comparison
3. You'll see side-by-side panels for:
   - Memory usage
   - CPU usage
   - Active series count
   - Architecture explanations

## Part 3: The Cardinality Test

### Understanding Cardinality

**Cardinality** = number of unique time series.

Each unique combination of metric name + labels = one series:
```
sensor_reading{deck="1", section="a", type="temp"}  # Series 1
sensor_reading{deck="1", section="a", type="humidity"}  # Series 2
sensor_reading{deck="1", section="b", type="temp"}  # Series 3
# ... and so on
```

High cardinality is the #1 cause of TSDB performance issues.

### Scale the Cardinality

```bash
# Check current cardinality
kubectl exec -n tsdb-comparison deploy/benchmark-runner -- \
  curl -s http://cardinality-generator:8080/cardinality | jq .

# Set to LOW (1,000 series) - baseline
kubectl exec -n tsdb-comparison deploy/benchmark-runner -- \
  curl -s -X POST "http://cardinality-generator:8080/cardinality?level=low"

# Wait 2 minutes, then check Grafana dashboard

# Set to MEDIUM (10,000 series)
kubectl exec -n tsdb-comparison deploy/benchmark-runner -- \
  curl -s -X POST "http://cardinality-generator:8080/cardinality?level=medium"

# Wait 2 minutes, observe changes in Grafana

# Set to HIGH (50,000 series) - stress test
kubectl exec -n tsdb-comparison deploy/benchmark-runner -- \
  curl -s -X POST "http://cardinality-generator:8080/cardinality?level=high"

# Wait 2 minutes, observe the divergence
```

### What to Observe

At each cardinality level, note:

1. **Memory Usage**
   - Prometheus: Grows linearly with cardinality
   - Victoria Metrics: Grows slower (better compression)
   - Mimir: Baseline overhead + linear growth

2. **CPU Usage**
   - All three increase with cardinality
   - Victoria Metrics typically lowest

3. **Query Performance**
   - Run the query benchmark (see Part 4)

## Part 4: Run the Benchmarks

### Full Comparison Benchmark

```bash
# Run the full benchmark (takes ~5 minutes)
kubectl exec -n tsdb-comparison deploy/benchmark-runner -- \
  /scripts/benchmark.sh
```

This will:
1. Set cardinality to low, medium, high
2. Measure memory, CPU, query latency at each level
3. Output a comparison table

### Query Performance Benchmark

```bash
# Run query latency benchmark
kubectl exec -n tsdb-comparison deploy/benchmark-runner -- \
  /scripts/query-benchmark.sh
```

## Part 5: Architecture Deep-Dive

### Why Victoria Metrics Uses Less Memory

**Prometheus Architecture:**
- Each metric → separate file on disk
- Index in memory: `{labels} → offset`
- 2-hour blocks, compacted over time
- Trade-off: Simple but memory-intensive

**Victoria Metrics Architecture:**
- Merge-tree storage engine (like ClickHouse)
- Better compression (varint encoding, delta-of-delta)
- In-memory cache is smarter
- Trade-off: More complex code, eventual consistency

**The Math:**
- Prometheus: ~500 bytes/series in memory
- Victoria Metrics: ~100-200 bytes/series
- At 50k series: 25MB vs 5-10MB difference

### Why Mimir Needs Object Storage

**Mimir Architecture:**
- Distributed: ingesters, queriers, compactors
- Data flows: write → ingester → S3 → store-gateway → read
- Multi-tenant by design
- Unlimited horizontal scaling

**The Trade-off:**
- More operational complexity
- Higher baseline resource usage
- But: scales to millions of series across multiple clusters

### Decision Framework

| Scenario | Best Choice | Why |
|----------|-------------|-----|
| Single cluster, <100k series | Prometheus | Simple, battle-tested |
| Resource-constrained (N100, RPi) | Victoria Metrics | 3-5x less RAM |
| Need >1 year retention | Victoria Metrics or Mimir | Built-in long-term storage |
| Multi-cluster, multi-team | Mimir | Native federation, multi-tenant |
| High cardinality (>1M series) | Victoria Metrics | Handles it better |
| Grafana-centric stack | Mimir | Grafana Labs integration |

## Part 6: Cost Modeling

### Translating to Cloud Costs

Resource usage translates directly to cloud costs:

| TSDB | RAM @ 50k series | EC2 equivalent | Monthly cost* |
|------|------------------|----------------|---------------|
| Prometheus | ~800MB | t3.small (2GB) | ~$15 |
| Victoria Metrics | ~300MB | t3.micro (1GB) | ~$8 |
| Mimir | ~1.2GB (with MinIO) | t3.medium (4GB) | ~$30 |

*Approximate, varies by region and usage

### When Each Makes Sense

**Prometheus**:
- You have one cluster
- Retention needs are <30 days
- Team knows Prometheus

**Victoria Metrics**:
- Running on limited hardware
- High cardinality workloads
- Want drop-in Prometheus replacement

**Mimir**:
- Multiple teams need isolated tenants
- Unlimited retention required
- Already using Grafana Cloud patterns

## Part 7: Cleanup

```bash
# Tear down the tutorial
task kind:down -- tsdb-comparison
```

## Summary

You've learned:

1. **Architecture matters**: Same PromQL, very different internals
2. **Measure, don't assume**: Run benchmarks on YOUR workload
3. **Trade-offs everywhere**:
   - Prometheus: Simple but limited
   - Victoria Metrics: Efficient but different
   - Mimir: Powerful but complex
4. **Cardinality is king**: High cardinality kills all TSDBs

## Next Steps

- Try the [SeaweedFS tutorial](../seaweedfs-tutorial/) - object storage for long-term metrics
- Explore [Thanos](../thanos-tutorial/) - extend Prometheus with S3 backend
- Read [ADR-009](../../../docs/adrs/ADR-009-tsdb-selection.md) - our TSDB decision rationale

## References

- [Prometheus Storage](https://prometheus.io/docs/prometheus/latest/storage/)
- [Victoria Metrics Architecture](https://docs.victoriametrics.com/single-server-victoriametrics/#architecture-overview)
- [Mimir Architecture](https://grafana.com/docs/mimir/latest/references/architecture/)
- [High Cardinality in Prometheus](https://www.robustperception.io/cardinality-is-key)
