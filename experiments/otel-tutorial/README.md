# OpenTelemetry & Distributed Tracing Tutorial

Learn distributed tracing with OpenTelemetry, OTel Collector, and Grafana Tempo.

## Overview

This tutorial covers:
- OpenTelemetry fundamentals (SDK, Collector, OTLP)
- Trace anatomy (spans, attributes, events, links)
- W3C Trace Context propagation
- Tempo and TraceQL queries
- Service dependency maps
- Correlating traces with logs and metrics

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        VISUALIZATION                            │
│                                                                 │
│                    Grafana (Tempo Datasource)                   │
│                    ├── Trace Search                             │
│                    ├── TraceQL Queries                          │
│                    └── Service Graph                            │
│                              │                                  │
│                              ▼                                  │
│                    ┌─────────────────┐                          │
│                    │      TEMPO      │                          │
│                    │ (trace backend) │                          │
│                    └────────▲────────┘                          │
│                             │                                   │
│                    ┌────────┴────────┐                          │
│                    │  OTEL COLLECTOR │                          │
│                    │   (receiver +   │                          │
│                    │    processor)   │                          │
│                    └────────▲────────┘                          │
│                             │ OTLP                              │
│         ┌───────────────────┼───────────────────┐               │
│         │                   │                   │               │
│    ┌────┴────┐        ┌─────┴─────┐       ┌─────┴─────┐        │
│    │  user   │ ────▶  │   order   │ ────▶ │  payment  │        │
│    │ service │  HTTP  │  service  │  HTTP │  service  │        │
│    └─────────┘        └───────────┘       └───────────┘        │
│                                                                 │
│              MULTI-SERVICE DEMO APP (OTel SDK)                  │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

- Completed `prometheus-tutorial` (metrics fundamentals)
- Completed `loki-tutorial` (optional, for logs-traces correlation)

## Quick Start

```bash
# Deploy the tutorial
task kind:conduct -- otel-tutorial

# Wait for all pods to be ready
kubectl get pods -n otel-tutorial -w

# Access services
kubectl port-forward -n otel-tutorial svc/kube-prometheus-stack-grafana 3000:80 &
kubectl port-forward -n otel-tutorial svc/user-service 8080:8080 &

# Open Grafana (admin/admin)
open http://localhost:3000
```

## Generating Traces

```bash
# Get a user (creates trace across user-service only)
curl http://localhost:8080/api/users/1

# Create an order (creates distributed trace across all 3 services)
curl -X POST http://localhost:8080/api/users/1/orders \
  -H "Content-Type: application/json" \
  -d '{"items": [{"product_id": "prod-001", "quantity": 2}]}'
```

## Tutorial Modules

### Module 1: OpenTelemetry Fundamentals

**Objectives:**
- Understand OTel architecture (SDK → Collector → Backend)
- Learn about OTLP (OpenTelemetry Protocol)
- Explore the three signals: traces, metrics, logs

**Key Concepts:**

| Component | Role |
|-----------|------|
| OTel SDK | Instrument applications, create spans |
| OTel Collector | Receive, process, export telemetry |
| OTLP | Wire protocol for sending telemetry |
| Backend (Tempo) | Store and query traces |

**Explore:**
```bash
# View OTel Collector config
kubectl get configmap -n otel-tutorial otel-collector-opentelemetry-collector -o yaml

# Check collector is receiving traces
kubectl logs -n otel-tutorial -l app.kubernetes.io/name=opentelemetry-collector --tail=50
```

### Module 2: Trace Anatomy

**Objectives:**
- Understand trace structure (trace → spans → events)
- Learn about span attributes and status
- Explore parent-child relationships

**Trace Structure:**
```
Trace (trace_id: abc123)
├── Span: HTTP GET /api/users/1 (parent)
│   ├── Span: db.query users (child)
│   └── Span: HTTP POST /api/orders (child)
│       ├── Span: validate.inventory (child)
│       └── Span: HTTP POST /api/payments (child)
│           ├── Span: fraud.check (child)
│           └── Span: payment.process (child)
```

**Span Attributes:**
- `service.name`: Which service created the span
- `http.method`, `http.url`: HTTP details
- `http.status_code`: Response status
- `db.system`, `db.statement`: Database details
- Custom attributes for business context

### Module 3: Context Propagation

**Objectives:**
- Understand W3C Trace Context headers
- See how trace IDs flow across services
- Learn about baggage propagation

**W3C Trace Context Headers:**
```
traceparent: 00-{trace_id}-{span_id}-{flags}
tracestate: vendor1=value1,vendor2=value2
```

**Example:**
```bash
# Make request with custom trace ID (for testing)
curl -H "traceparent: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01" \
  http://localhost:8080/api/users/1
```

### Module 4: Tempo & TraceQL

**Objectives:**
- Query traces using TraceQL
- Filter by service, duration, status
- Use service graph visualization

**TraceQL Examples:**
```
# Find all traces from user-service
{resource.service.name="user-service"}

# Find slow traces (>500ms)
{resource.service.name=~".*-service"} | duration > 500ms

# Find error traces
{status=error}

# Find traces with specific span name
{name="HTTP POST"}

# Find traces calling payment-service with status OK
{resource.service.name="payment-service" && status=ok}

# Structural query: find traces where user-service calls payment-service
{resource.service.name="user-service"} >> {resource.service.name="payment-service"}
```

**Exercise:**
1. Open Grafana → Explore → Select Tempo
2. Run: `{resource.service.name="user-service"}`
3. Click a trace to see the waterfall view
4. Identify the call chain across services

### Module 5: Span Analysis

**Objectives:**
- Analyze span waterfall views
- Identify latency bottlenecks
- Understand span breakdown

**What to Look For:**
- **Gap Analysis**: Large gaps between spans indicate waiting time
- **Deep Nesting**: Many levels of child spans may indicate N+1 problems
- **Long Spans**: Identify which operation takes the most time
- **Error Spans**: Red spans indicate failures

**Exercise:**
1. Generate traffic with several requests
2. Find a trace with multiple spans
3. In the waterfall view, identify:
   - Which service takes the longest?
   - Which operation within that service is slowest?
   - Are there any gaps or waiting time?

### Module 6: Error Tracing

**Objectives:**
- Trace error propagation across services
- Understand span status codes
- Use exception spans for debugging

**Span Status:**
- `OK`: Operation succeeded
- `ERROR`: Operation failed
- `UNSET`: Status not explicitly set

**Finding Errors:**
```
# TraceQL for errors
{status=error}

# Find traces with specific error message
{span.error.message=~".*timeout.*"}
```

### Module 7: Metrics from Traces (Span Metrics)

**Objectives:**
- Understand span metrics (RED metrics from traces)
- Query span-derived metrics in Prometheus

The OTel Collector generates metrics from spans:
- `otel_span_duration_count`: Request count
- `otel_span_duration_sum`: Total duration
- `otel_span_duration_bucket`: Duration histogram

**Prometheus Queries:**
```promql
# Request rate by service
sum(rate(otel_span_duration_count{service_name=~".*-service"}[1m])) by (service_name)

# p95 latency by service
histogram_quantile(0.95, sum(rate(otel_span_duration_bucket[1m])) by (le, service_name))

# Error rate
sum(rate(otel_span_duration_count{status_code="ERROR"}[1m])) by (service_name)
```

### Module 8: Logs ↔ Traces Correlation

**Objectives:**
- Understand trace_id injection in logs
- Jump from logs to traces
- Use Grafana's integrated view

**How It Works:**
1. Application logs include `trace_id` field
2. Query logs in Loki with trace_id
3. Click "View trace" to jump to Tempo

**Loki Query (if Loki is deployed):**
```
{app="user-service"} | json | trace_id != ""
```

### Module 9: Service Dependency Maps

**Objectives:**
- Understand service graph generation
- Visualize service dependencies
- Identify critical paths

**View Service Graph:**
1. Grafana → Explore → Tempo
2. Click "Service Graph" tab
3. See visualization: `user-service → order-service → payment-service`

The service graph is generated from trace data showing which services call which.

## Components Deployed

| Component | Purpose | Access |
|-----------|---------|--------|
| OTel Collector | Receive, process, export traces | Internal only |
| Tempo | Trace storage and query | Via Grafana |
| Grafana | Visualization, TraceQL | Port 3000 |
| user-service | Entry point, demo app | Port 8080 |
| order-service | Middle tier, demo app | Internal |
| payment-service | Backend, demo app | Internal |

## Troubleshooting

### No traces appearing in Tempo

```bash
# Check OTel Collector is running
kubectl get pods -n otel-tutorial -l app.kubernetes.io/name=opentelemetry-collector

# Check collector logs for errors
kubectl logs -n otel-tutorial -l app.kubernetes.io/name=opentelemetry-collector

# Check Tempo is running
kubectl get pods -n otel-tutorial -l app.kubernetes.io/name=tempo

# Verify demo app is sending traces
kubectl logs -n otel-tutorial -l app=user-service | grep -i trace
```

### Demo app not responding

```bash
# Check all pods are running
kubectl get pods -n otel-tutorial

# Check service endpoints
kubectl get svc -n otel-tutorial

# Check demo app logs
kubectl logs -n otel-tutorial -l app=user-service
kubectl logs -n otel-tutorial -l app=order-service
kubectl logs -n otel-tutorial -l app=payment-service
```

## Next Steps

After completing this tutorial:
1. **tracing-comparison**: Compare Tempo vs Jaeger
2. **Auto-instrumentation**: Explore OTel auto-instrumentation agents
3. **Sampling**: Configure head/tail sampling for production
4. **Exemplars**: Link metrics to traces via exemplars

## Resources

- [OpenTelemetry Docs](https://opentelemetry.io/docs/)
- [TraceQL Reference](https://grafana.com/docs/tempo/latest/traceql/)
- [Tempo Documentation](https://grafana.com/docs/tempo/latest/)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
