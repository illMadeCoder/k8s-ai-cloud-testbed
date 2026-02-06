# SLOs & Error Budgets Tutorial

Learn Service Level Objectives (SLOs) and error budget management using Pyrra.

## Overview

This tutorial covers:
- SLI/SLO/SLA hierarchy
- Defining SLIs from Prometheus metrics
- Pyrra ServiceLevelObjective CRDs
- Error budget calculations and tracking
- Multi-window, multi-burn-rate alerting

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                       VISUALIZATION                              │
│                                                                  │
│       Grafana Dashboard              Pyrra UI                    │
│       (error budgets)              (SLO management)              │
│              │                          │                        │
│              └──────────┬───────────────┘                        │
│                         │                                        │
│              ┌──────────▼──────────┐                            │
│              │     PROMETHEUS      │                            │
│              │  - SLI metrics      │                            │
│              │  - Recording rules  │                            │
│              │  - Burn-rate alerts │                            │
│              └──────────▲──────────┘                            │
│                         │                                        │
│                   ServiceMonitor                                 │
│                         │                                        │
│              ┌──────────┴──────────┐                            │
│              │     METRICS-APP     │                            │
│              │  http_requests_*    │                            │
│              │  request_duration_* │                            │
│              └─────────────────────┘                            │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

- Completed `prometheus-tutorial` (metrics fundamentals)

## Quick Start

```bash
# Deploy the tutorial
task kind:conduct -- slo-tutorial

# Wait for all pods to be ready
kubectl get pods -n slo-tutorial -w

# Access services
kubectl port-forward -n slo-tutorial svc/kube-prometheus-stack-grafana 3000:80 &
kubectl port-forward -n slo-tutorial svc/pyrra 9099:9099 &
kubectl port-forward -n slo-tutorial svc/metrics-app 8080:8080 &

# Open UIs
open http://localhost:3000  # Grafana (admin/admin)
open http://localhost:9099  # Pyrra
```

## Tutorial Modules

### Module 1: SLO Fundamentals

**Key Concepts:**

| Term | Definition | Example |
|------|------------|---------|
| **SLI** | Service Level Indicator | % of requests returning 2xx |
| **SLO** | Service Level Objective | 99.9% of requests succeed |
| **SLA** | Service Level Agreement | Contract with penalties |
| **Error Budget** | Allowed failures | 0.1% = 43.2 min/month |

**Error Budget Calculation:**
```
Error Budget = 100% - SLO Target

For 99.9% SLO (30-day window):
  Budget = 0.1% = 0.001
  Minutes = 30 days × 24 hours × 60 min × 0.001
  = 43.2 minutes of downtime allowed
```

### Module 2: Defining SLIs

**Good SLIs have these properties:**
- Directly measure user experience
- Can be calculated from existing metrics
- Are aggregatable across time and instances

**Availability SLI:**
```promql
# Success rate = successful requests / total requests
sum(rate(http_requests_total{status!~"5.."}[5m]))
/
sum(rate(http_requests_total[5m]))
```

**Latency SLI:**
```promql
# % requests under threshold = fast requests / total requests
sum(rate(request_duration_seconds_bucket{le="0.5"}[5m]))
/
sum(rate(request_duration_seconds_count[5m]))
```

### Module 3: Pyrra SLO Definitions

Pyrra uses Kubernetes CRDs to define SLOs:

**Availability SLO (99.9%):**
```yaml
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: metrics-app-availability
spec:
  target: "99.9"
  window: 30d
  indicator:
    ratio:
      errors:
        metric: http_requests_total{job="metrics-app",status=~"5.."}
      total:
        metric: http_requests_total{job="metrics-app"}
```

**Latency SLO (99% < 500ms):**
```yaml
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: metrics-app-latency
spec:
  target: "99"
  window: 30d
  indicator:
    latency:
      success:
        metric: request_duration_seconds_bucket{job="metrics-app",le="0.5"}
      total:
        metric: request_duration_seconds_count{job="metrics-app"}
```

### Module 4: Error Budget Math

**Burn Rate:**
The rate at which you're consuming your error budget.

```
Burn Rate = Actual Error Rate / Allowed Error Rate

If SLO = 99.9% (0.1% allowed errors)
And actual error rate = 1%
Then burn rate = 1% / 0.1% = 10x

At 10x burn rate, you'll exhaust your monthly budget in 3 days.
```

**Budget Consumption Over Time:**
| Burn Rate | Time to Exhaust 30-day Budget |
|-----------|-------------------------------|
| 1x | 30 days |
| 2x | 15 days |
| 6x | 5 days |
| 14.4x | 50 hours |
| 36x | 20 hours |

### Module 5: Multi-Window Alerting

Google SRE's multi-window, multi-burn-rate alerting strategy:

| Alert Type | Long Window | Short Window | Burn Rate | Action |
|------------|-------------|--------------|-----------|--------|
| Page (critical) | 1h | 5m | 14.4x | Wake someone up |
| Page (critical) | 6h | 30m | 6x | Wake someone up |
| Ticket (warning) | 3d | 6h | 1x | Create ticket |

**Why multiple windows?**
- Long window: Ensures significant budget consumption
- Short window: Confirms the issue is ongoing
- Both must fire to alert

**Exercise:**
```bash
# Generate errors to consume budget
for i in {1..50}; do
  curl http://localhost:8080/error
  sleep 1
done

# Watch alerts in Prometheus:
kubectl port-forward -n slo-tutorial svc/kube-prometheus-stack-prometheus 9090:9090
# Navigate to Alerts tab
```

### Module 6: SLO Dashboard

**Grafana Dashboard Panels:**
1. **Current SLI Gauges** - Real-time availability and latency
2. **Error Budget Remaining** - % of budget left this window
3. **SLI Over Time** - Trend with SLO threshold line
4. **Request Rate by Status** - Success vs error breakdown

**Pyrra UI Features:**
- SLO list with compliance status
- Error budget burn-down visualization
- Alert status integration
- Historical compliance trends

## Exercises

### Exercise 1: Consume Error Budget

```bash
# Generate normal traffic
for i in {1..100}; do curl -s http://localhost:8080/; done

# Generate errors (30% error rate)
for i in {1..30}; do curl -s http://localhost:8080/error; done

# Check the dashboard - error budget should decrease
```

### Exercise 2: Trigger Slow Responses

```bash
# Hit the slow endpoint (500-2000ms responses)
for i in {1..50}; do curl -s http://localhost:8080/slow; done

# This will impact the latency SLO
```

### Exercise 3: View Generated Alerts

```bash
# See the PrometheusRules Pyrra created
kubectl get prometheusrules -n slo-tutorial -o yaml

# The rules implement multi-window, multi-burn-rate alerting
```

## Components Deployed

| Component | Purpose | Access |
|-----------|---------|--------|
| Prometheus | Metrics collection, SLI calculation | Port 9090 |
| Pyrra | SLO management, rule generation | Port 9099 |
| Grafana | Dashboards, visualization | Port 3000 |
| metrics-app | Sample application with metrics | Port 8080 |

## Troubleshooting

### SLOs not appearing in Pyrra

```bash
# Check SLO CRDs exist
kubectl get servicelevelobjectives -n slo-tutorial

# Check Pyrra logs
kubectl logs -n slo-tutorial -l app.kubernetes.io/name=pyrra

# Verify Prometheus connectivity
kubectl exec -n slo-tutorial deploy/pyrra -- \
  wget -qO- http://kube-prometheus-stack-prometheus:9090/api/v1/status/config
```

### Metrics not appearing

```bash
# Check ServiceMonitor
kubectl get servicemonitors -n slo-tutorial

# Verify metrics-app is exposing metrics
kubectl port-forward -n slo-tutorial svc/metrics-app 8080:8080
curl http://localhost:8080/metrics | head -20
```

## Alternative Tools

While this tutorial uses Pyrra, other SLO tools exist:

| Tool | Approach | UI |
|------|----------|-----|
| **Pyrra** | K8s CRDs + operator | Built-in web UI |
| **Sloth** | YAML → PrometheusRules | None (use Grafana) |
| **OpenSLO** | Vendor-neutral spec | Varies by implementation |

## Resources

- [Google SRE Book - SLOs](https://sre.google/sre-book/service-level-objectives/)
- [Pyrra Documentation](https://github.com/pyrra-dev/pyrra)
- [The Art of SLOs](https://sre.google/workbook/implementing-slos/)
- [Multi-Window, Multi-Burn-Rate Alerts](https://sre.google/workbook/alerting-on-slos/)
