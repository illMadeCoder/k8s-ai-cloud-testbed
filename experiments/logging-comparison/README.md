# Logging Stack Comparison: Loki vs Elasticsearch

**Difficulty**: Intermediate | **Duration**: 45-60 minutes

Compare two popular Kubernetes logging solutions with measurable, reproducible results.

## Overview

This tutorial deploys both logging stacks side-by-side:

| Stack | Components | Architecture |
|-------|------------|--------------|
| **Loki** | Loki + Promtail | Label-indexed, stores chunks |
| **ELK** | Elasticsearch + Fluent Bit + Kibana | Full-text inverted index |

```
┌──────────────────────────────────────────────────────────────────────┐
│                         Target Cluster                                │
│  ┌──────────────┐                                                    │
│  │    log-      │     ┌──────────────┐                               │
│  │  generator   │────▶│    Loki      │  (label-indexed)              │
│  │   (stdout)   │     │  + Promtail  │                               │
│  │              │     └──────────────┘                               │
│  │              │     ┌──────────────┐                               │
│  │              │────▶│Elasticsearch │  (full-text index)            │
│  │              │     │  + Fluent Bit│                               │
│  └──────────────┘     └──────────────┘                               │
│                                                                       │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐                 │
│  │   Grafana   │   │   Kibana    │   │ Comparison  │                 │
│  │  (unified)  │   │  (ES native)│   │  Dashboard  │                 │
│  └─────────────┘   └─────────────┘   └─────────────┘                 │
└──────────────────────────────────────────────────────────────────────┘
```

## Learning Objectives

By the end of this tutorial, you will:

1. **Understand architecture differences** - Label-indexed vs full-text indexing
2. **Measure resource usage** - Memory, CPU, and storage under load
3. **Compare query performance** - Which queries each system excels at
4. **Make informed decisions** - When to choose Loki vs Elasticsearch

## Prerequisites

- Completed: Kind cluster basics
- **Recommended**: [loki-tutorial](../loki-tutorial/) - Learn LogQL basics first
- **Recommended**: [elk-tutorial](../elk-tutorial/) - Learn KQL/Lucene basics first
- ~6GB RAM available (Elasticsearch needs 2GB+)

> **Note**: This tutorial focuses on comparing the two stacks, not teaching query basics.
> Complete the individual tutorials first for the best learning experience.

## Part 1: Deploy the Tutorial

### 1.1 Start the Cluster

```bash
task kind:up -- logging-comparison
```

This creates a medium-sized Kind cluster and deploys:
- Loki (single binary mode) + Promtail
- Elasticsearch (single node) + Fluent Bit + Kibana
- Grafana with both datasources
- Log generator workload

### 1.2 Wait for Pods

```bash
# Watch all pods come up
kubectl get pods -A -w

# Or wait for specific namespaces
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=loki -n loki --timeout=300s
kubectl wait --for=condition=ready pod -l app=elasticsearch-master -n elasticsearch --timeout=300s
```

### 1.3 Access the UIs

```bash
# Get service IPs
kubectl get svc -A | grep LoadBalancer
```

| Service | Default Port | Purpose |
|---------|--------------|---------|
| Grafana | 3000 | Unified dashboards (admin/admin) |
| Kibana | 5601 | Elasticsearch native UI |
| Log Generator | 8080 | Control API |

## Part 2: Explore the Baseline

### 2.1 View the Comparison Dashboard

Open Grafana and navigate to: **Dashboards > Logging Stack Comparison**

You'll see:
- Side-by-side memory usage (Loki vs Elasticsearch)
- Side-by-side CPU usage
- Total resource comparison stats
- Log generator status

### 2.2 Check Initial Log Generator Config

```bash
# Get log-generator IP
GENERATOR_IP=$(kubectl get svc log-generator -n logging-comparison -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

# Check current configuration
curl http://${GENERATOR_IP}:8080/config
```

Default: `medium` rate (100 logs/sec), `medium` cardinality (20 services)

### 2.3 View Logs in Each System

**In Grafana (Loki):**
1. Go to Explore
2. Select "Loki" datasource
3. Query: `{app="log-generator"}`

**In Kibana (Elasticsearch):**
1. Open Kibana at port 5601
2. Go to Discover
3. Create index pattern: `logs-*`
4. View incoming logs

## Part 3: Scale Up - Observe Divergence

### 3.1 Increase Log Volume

```bash
# Scale to high: 1000 logs/sec
curl -X POST "http://${GENERATOR_IP}:8080/config?rate=high&cardinality=high"
```

### 3.2 Watch Resource Usage

In the Grafana dashboard, observe:
- **Loki**: Modest memory increase (~300-400MB total)
- **Elasticsearch**: Significant memory increase (~1.5-2GB+ total)

### 3.3 Why the Difference?

| Aspect | Loki | Elasticsearch |
|--------|------|---------------|
| **Indexing** | Only labels | All fields |
| **Storage** | Compressed chunks | Inverted index + source |
| **Memory** | Minimal for ingestion | JVM heap for indexing |
| **Query** | Grep through chunks | Index lookup |

## Part 4: Query Performance Comparison

### 4.1 Run the Query Benchmark

```bash
# Exec into benchmark runner
kubectl exec -it benchmark-runner -n logging-comparison -- bash

# Run query benchmark
/scripts/query-benchmark.sh
```

### 4.2 Test Types

**Test 1: Label-based filtering** (Loki advantage)
- Find all errors from service-1
- Loki uses label index → fast
- Elasticsearch uses filter → also fast

**Test 2: Full-text search** (Elasticsearch advantage)
- Find specific trace_id
- Loki greps through chunks → slower
- Elasticsearch uses inverted index → fast

**Test 3: Aggregations** (Depends on query)
- Count errors by service
- Loki: metric queries efficient
- Elasticsearch: aggregations efficient

### 4.3 Manual Query Examples

**Loki (LogQL):**
```logql
# Find errors
{app="log-generator"} |= "error"

# Parse JSON and filter
{app="log-generator"} | json | level="error" | service="service-5"

# Metric: errors per minute
sum(rate({app="log-generator"} |= "error" [1m]))
```

**Elasticsearch (KQL in Kibana):**
```
# Find errors
level:error

# Filter by service
level:error AND service:service-5

# Full-text search
message:"connection refused"
```

## Part 5: Understanding the Architecture

### 5.1 Loki Architecture

```
Logs → Promtail → Loki
         │          │
    1. Reads stdout │
    2. Parses JSON  │
    3. Extracts     │
       labels       │
                    │
              ┌─────┴─────┐
              │ Label     │
              │ Index     │──── Small (just labels)
              │           │
              │ Chunks    │
              │ (gzip)    │──── Compressed log text
              └───────────┘
```

**Key insight**: Loki doesn't index log content. It stores compressed chunks and greps through them at query time.

### 5.2 Elasticsearch Architecture

```
Logs → Fluent Bit → Elasticsearch
           │              │
    1. Reads stdout       │
    2. Parses JSON        │
    3. Enriches with      │
       K8s metadata       │
                          │
                    ┌─────┴─────┐
                    │ Inverted  │
                    │ Index     │──── Large (all terms)
                    │           │
                    │ _source   │
                    │ (JSON)    │──── Original document
                    └───────────┘
```

**Key insight**: Elasticsearch indexes EVERY field, making any search fast but requiring more resources.

## Part 6: Decision Framework

### When to Use Loki

✅ **Good fit:**
- Known query patterns (filter by namespace, pod, level)
- Cost-conscious environments
- Grafana-centric observability
- Simple deployment requirements
- Container/Kubernetes-native logging

❌ **Poor fit:**
- Heavy full-text search requirements
- Complex ad-hoc queries
- Compliance requiring searchable archives

### When to Use Elasticsearch

✅ **Good fit:**
- Full-text search requirements
- Complex query patterns discovered at runtime
- Need to search any field efficiently
- Existing Elastic ecosystem investment
- Log analytics and visualizations

❌ **Poor fit:**
- Resource-constrained environments
- Simple label-based filtering only
- Tight Grafana integration needed

### Cost Comparison (Rough)

At 100GB/day ingestion:

| Aspect | Loki | Elasticsearch |
|--------|------|---------------|
| **Memory** | 2-4 GB | 16-32 GB |
| **Storage** | ~30 GB | ~100 GB |
| **CPU** | Low | Medium-High |
| **Complexity** | Simple | Complex |

## Part 7: Cleanup

### 7.1 Scale Down Log Generator

```bash
curl -X POST "http://${GENERATOR_IP}:8080/config?rate=low"
```

### 7.2 Destroy the Cluster

```bash
task kind:down -- logging-comparison
```

## Summary

| Feature | Loki | Elasticsearch |
|---------|------|---------------|
| **Architecture** | Label-index + chunks | Full inverted index |
| **Memory Usage** | Low (~300MB) | High (~2GB+) |
| **Label Queries** | Fast | Fast |
| **Full-text Search** | Slow (grep) | Fast (indexed) |
| **Best For** | Known patterns, K8s | Ad-hoc search, analytics |
| **Complexity** | Simple | Complex |
| **Grafana Integration** | Native | Plugin |

**Bottom line**: Use Loki for cost-effective logging with known query patterns. Use Elasticsearch when you need powerful full-text search and complex analytics.

## Next Steps

- [ADR-010: Logging Stack Selection](../../docs/adrs/ADR-010-logging-stack-selection.md) - Our decision rationale
- [Prometheus Tutorial](../prometheus-tutorial/) - Metrics observability
- [TSDB Comparison](../tsdb-comparison/) - Similar comparison for metrics backends

## Troubleshooting

### Elasticsearch OOM

If Elasticsearch crashes with OOM:
```bash
# Check pod status
kubectl describe pod -n elasticsearch elasticsearch-master-0

# Elasticsearch needs 2GB minimum. Ensure cluster has resources:
kubectl top nodes
```

### Loki Not Receiving Logs

Check Promtail:
```bash
# Promtail logs
kubectl logs -n loki -l app.kubernetes.io/name=promtail

# Check Loki is healthy
curl http://${LOKI_IP}:3100/ready
```

### No Logs in Elasticsearch

Check Fluent Bit:
```bash
# Fluent Bit logs
kubectl logs -n elasticsearch -l app.kubernetes.io/name=fluent-bit

# Check ES index
curl http://${ES_IP}:9200/_cat/indices?v
```
