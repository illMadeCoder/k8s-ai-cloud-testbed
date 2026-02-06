# Tempo vs Jaeger - Tracing Backend Comparison

Compare Grafana Tempo and Jaeger for distributed tracing.

## Overview

This scenario deploys both Tempo and Jaeger to receive the same traces via dual-write from the OTel Collector, allowing direct comparison.

## Architecture

```
┌───────────────────────────────────────────────────────────────┐
│                       VISUALIZATION                           │
│                                                               │
│     Grafana Explore                    Jaeger UI              │
│     (Tempo datasource)                 (port 16686)           │
│           │                                │                  │
│           ▼                                ▼                  │
│    ┌─────────────┐                 ┌─────────────┐           │
│    │    TEMPO    │                 │   JAEGER    │           │
│    │  TraceQL    │                 │  All-in-One │           │
│    └──────▲──────┘                 └──────▲──────┘           │
│           │                               │                  │
│           └───────────┬───────────────────┘                  │
│                       │                                      │
│            ┌──────────┴──────────┐                           │
│            │   OTEL COLLECTOR    │                           │
│            │   (dual-write)      │                           │
│            └──────────▲──────────┘                           │
│                       │ OTLP                                 │
│    ┌──────────────────┼──────────────────┐                   │
│    │                  │                  │                   │
│    ▼                  ▼                  ▼                   │
│ user-service → order-service → payment-service               │
│                                                               │
│              OTel Demo App                                    │
└───────────────────────────────────────────────────────────────┘
```

## Quick Start

```bash
# Deploy the comparison environment
task kind:conduct -- tracing-comparison

# Wait for pods
kubectl get pods -n tracing-comparison -w

# Access Grafana (Tempo)
kubectl port-forward -n tracing-comparison svc/kube-prometheus-stack-grafana 3000:80 &

# Access Jaeger UI
kubectl port-forward -n tracing-comparison svc/jaeger-query 16686:16686 &

# Generate traces
curl http://localhost:8080/api/users/1
```

## Comparison

### Architecture

| Aspect | Tempo | Jaeger |
|--------|-------|--------|
| **Model** | Block-oriented storage | Collector + Query + UI |
| **Storage** | Local/S3/GCS (like Loki) | Memory, Cassandra, ES |
| **Deployment** | Single binary or microservices | All-in-one or distributed |
| **Protocol** | OTLP, Jaeger, Zipkin | OTLP, Jaeger, Zipkin |

### Query Capabilities

| Feature | Tempo (TraceQL) | Jaeger |
|---------|-----------------|--------|
| Service filter | `{resource.service.name="x"}` | Dropdown |
| Duration filter | `{duration > 500ms}` | Min/Max Duration |
| Error filter | `{status=error}` | Tags: error=true |
| Tag search | `{span.http.method="GET"}` | Tag key=value |
| Structural query | `{svc=A} >> {svc=B}` | Not supported |
| Regex | `{name=~"HTTP.*"}` | Limited |

### UI/UX

| Aspect | Tempo | Jaeger |
|--------|-------|--------|
| **Interface** | Grafana Explore | Dedicated UI |
| **Trace View** | Waterfall in Grafana | Dedicated trace view |
| **Service Graph** | Via Grafana datasource | Built-in dependency graph |
| **Compare Traces** | Grafana panel | Side-by-side comparison |
| **Deep Linking** | Grafana URLs | Jaeger URLs |

### Integration

| Integration | Tempo | Jaeger |
|-------------|-------|--------|
| **Metrics** | Native (Prometheus) | Via OTel |
| **Logs** | Native (Loki) | Via tags |
| **Alerting** | Grafana Alerting | External |
| **Dashboards** | Grafana | External |

## Exercises

### Exercise 1: Query the Same Trace

1. Generate a trace:
   ```bash
   curl -X POST http://localhost:8080/api/users/1/orders \
     -H "Content-Type: application/json" \
     -d '{"items": [{"product_id": "p1", "quantity": 1}]}'
   ```

2. Find it in Tempo (Grafana Explore):
   ```
   {resource.service.name="user-service"} | duration > 100ms
   ```

3. Copy the trace ID and find it in Jaeger UI

4. Compare the visualizations

### Exercise 2: Structural Query (Tempo only)

Try finding traces where user-service calls payment-service:

```
{resource.service.name="user-service"} >> {resource.service.name="payment-service"}
```

This query follows the trace structure. Note that Jaeger doesn't have this capability.

### Exercise 3: Service Graph Comparison

1. **Tempo**: In Grafana, go to Explore → Tempo → Service Graph tab
2. **Jaeger**: In Jaeger UI, go to System Architecture / DAG

Compare how each visualizes the service dependencies.

### Exercise 4: Error Analysis

1. If the demo app generates errors, find them:
   - Tempo: `{status=error}`
   - Jaeger: Tags → error=true

2. Compare how each displays error information

## When to Use Each

### Choose Tempo When:
- You're already using Grafana for metrics/logs
- You want unified observability (LGTM stack)
- You need TraceQL for complex queries
- You want logs ↔ traces ↔ metrics correlation

### Choose Jaeger When:
- You want a dedicated tracing UI
- You're in a CNCF-native environment
- You need flexible storage backends (Cassandra, ES)
- Your team prefers the Jaeger interface

## Resources

- [Tempo Documentation](https://grafana.com/docs/tempo/latest/)
- [TraceQL Reference](https://grafana.com/docs/tempo/latest/traceql/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Jaeger vs Tempo Comparison](https://grafana.com/blog/2021/04/13/how-to-migrate-from-jaeger-to-grafana-tempo/)
