# Gateway Cluster Observability

Collects Envoy Gateway metrics from the GKE gateway cluster and ships them to
the hub's Mimir via Tailscale.

## Architecture

```
GKE Cluster                              Hub Cluster
┌──────────────────┐                    ┌──────────────────┐
│ Envoy Gateway    │                    │ Mimir            │
│   :19001/metrics │                    │   :8080/api/v1   │
└────────┬─────────┘                    └────────▲─────────┘
         │                                       │
         ▼                                       │
┌──────────────────┐    Tailscale    ┌──────────────────┐
│ Grafana Alloy    │──────────────────│ mimir-tailscale  │
│   scrape + push  │                  │   (TS exposed)   │
└──────────────────┘                  └──────────────────┘
```

## Deployment

### Hub side (one-time)
```bash
# Expose Mimir via Tailscale
kubectl --context talos-hub apply -f hub-mimir-service.yaml
```

### GKE side
```bash
# Create namespace
kubectl create namespace observability

# Create Tailscale egress to hub Mimir
kubectl apply -f gke-mimir-egress.yaml

# Install Alloy
helm upgrade --install alloy grafana/alloy \
  --namespace observability \
  -f alloy-values.yaml
```

## Metrics

- **envoy_gateway**: Controller metrics (reconcile times, errors)
- **envoy_proxy**: Data plane metrics (request counts, latencies, connections)

Query in Grafana with tenant `gateway-lab`:
```promql
rate(envoy_http_downstream_rq_total[5m])
```
