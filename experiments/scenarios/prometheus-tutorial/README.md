# Prometheus Tutorial Experiment

**Phase 3.1: Prometheus & Grafana Deep Dive**

Learn Prometheus metrics collection, PromQL queries, and Grafana dashboards with a hands-on experiment.

## Learning Objectives

- Understand Prometheus architecture (scraping, TSDB, federation)
- Write effective PromQL queries
- Build actionable Grafana dashboards
- Configure alerting pipelines

## Components Deployed

| Component | Purpose |
|-----------|---------|
| **kube-prometheus-stack** | Prometheus, Grafana, Alertmanager, exporters |
| **metrics-app** | Sample Go app with custom Prometheus metrics |

## Custom Metrics Exposed

The `metrics-app` exposes these Prometheus metrics:

| Metric | Type | Description |
|--------|------|-------------|
| `http_requests_total` | Counter | Total HTTP requests by method, path, status |
| `active_connections` | Gauge | Current number of active connections |
| `request_duration_seconds` | Histogram | Request duration with latency buckets |
| `response_size_bytes` | Summary | Response size with quantiles |
| `items_processed_total` | Counter | Business metric: items processed |

## Endpoints

The `metrics-app` provides these endpoints:

| Endpoint | Behavior |
|----------|----------|
| `/` | Normal response (10-100ms delay) |
| `/slow` | Slow response (500-2000ms delay) |
| `/error` | Random errors (30% failure rate) |
| `/process` | Business logic (processes 1-10 items) |
| `/metrics` | Prometheus metrics |
| `/health` | Liveness probe |
| `/ready` | Readiness probe |

## Running the Tutorial

```bash
# From repo root
task kind:conduct -- prometheus-tutorial
```

This will:
1. Create a Kind cluster with kube-prometheus-stack and metrics-app
2. Wait for LoadBalancer IPs to be assigned
3. Display access URLs and tutorial instructions
4. Wait for you to explore (the tutorial stays running)
5. **Press Ctrl+C when done** - this triggers automatic cleanup

## What You'll Learn

- Navigate Prometheus UI and execute PromQL queries
- Understand the four metric types: Counter, Gauge, Histogram, Summary
- Build queries for rate, percentiles, and error rates
- Explore pre-built Grafana dashboards
- Optionally create your own dashboards

## Accessing Services

Services are exposed via LoadBalancer (MetalLB locally, cloud LB in production):

| Service | URL | Credentials |
|---------|-----|-------------|
| **Prometheus** | Shown during tutorial | None |
| **Grafana** | Shown during tutorial | admin / admin |

### Port-Forward Alternative

If LoadBalancer IPs aren't available:

```bash
kubectl --context kind-prometheus-tutorial-target port-forward -n observability svc/kube-prometheus-stack-prometheus 9090:9090
kubectl --context kind-prometheus-tutorial-target port-forward -n observability svc/kube-prometheus-stack-grafana 3000:80
```

## PromQL Tutorial Queries

Try these queries in Prometheus UI or Grafana:

### Counter Queries

```promql
# Total requests (instant)
http_requests_total

# Request rate per second (last 5 minutes)
rate(http_requests_total[5m])

# Request rate by path
sum(rate(http_requests_total[5m])) by (path)

# Error rate (5xx responses)
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))
```

### Histogram Queries

```promql
# 50th percentile latency (median)
histogram_quantile(0.50, rate(request_duration_seconds_bucket[5m]))

# 95th percentile latency
histogram_quantile(0.95, rate(request_duration_seconds_bucket[5m]))

# 99th percentile latency by path
histogram_quantile(0.99, sum(rate(request_duration_seconds_bucket[5m])) by (path, le))

# Average latency
rate(request_duration_seconds_sum[5m]) / rate(request_duration_seconds_count[5m])
```

### Gauge Queries

```promql
# Current active connections
active_connections

# Max connections over last hour
max_over_time(active_connections[1h])
```

### Aggregation Examples

```promql
# Total requests by status code
sum(http_requests_total) by (status)

# Average request rate across all pods
avg(rate(http_requests_total[5m]))

# Top 5 slowest endpoints
topk(5, histogram_quantile(0.99, rate(request_duration_seconds_bucket[5m])))
```

### Alert-worthy Queries

```promql
# High error rate (> 5%)
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) > 0.05

# High latency (p99 > 1 second)
histogram_quantile(0.99, rate(request_duration_seconds_bucket[5m])) > 1

# No requests (absence detection)
absent(http_requests_total{job="metrics-app"})
```

## Building RED Dashboards

The RED method measures:
- **R**ate: Request rate (requests/second)
- **E**rrors: Error rate (errors/second or %)
- **D**uration: Latency (p50, p90, p99)

```promql
# Rate panel
sum(rate(http_requests_total[5m]))

# Errors panel
sum(rate(http_requests_total{status=~"5.."}[5m]))

# Duration panel (p50, p90, p99 as separate queries)
histogram_quantile(0.50, rate(request_duration_seconds_bucket[5m]))
histogram_quantile(0.90, rate(request_duration_seconds_bucket[5m]))
histogram_quantile(0.99, rate(request_duration_seconds_bucket[5m]))
```

## Files Structure

```
prometheus-tutorial/
├── README.md                           # This file
├── target/
│   ├── cluster.yaml                    # Cluster definition
│   └── argocd/
│       ├── app.yaml                    # ArgoCD Application
│       └── prometheus-values.yaml      # Helm values for kube-prometheus-stack
└── workflow/
    └── experiment.yaml                 # Argo Workflow
```

## Related

- [Phase 3.1: Prometheus & Grafana](../../../docs/roadmap/phase-03-observability.md)
- [metrics-app source](../../components/apps/metrics-app/)
- [kube-prometheus-stack component](../../components/observability/prometheus-stack/)

## Troubleshooting

### Prometheus not scraping metrics-app

Check ServiceMonitor exists and labels match:
```bash
kubectl get servicemonitor -n demo
kubectl describe servicemonitor metrics-app -n demo
```

Check Prometheus targets:
```bash
# In Prometheus UI: Status → Targets
# Or via API:
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.labels.job=="metrics-app")'
```

### Metrics not appearing

Ensure the app is exposing metrics:
```bash
kubectl port-forward -n demo svc/metrics-app 8080:80
curl http://localhost:8080/metrics
```

### Grafana login issues

Default credentials are `admin` / `admin`. If changed:
```bash
kubectl get secret -n observability kube-prometheus-stack-grafana -o jsonpath='{.data.admin-password}' | base64 -d
```
