# Demo App

Simple HTTP echo application for testing the K8s lab environment.

## Overview

- **Image**: `hashicorp/http-echo:1.0`
- **Purpose**: Testing Gateway API, ArgoCD, Prometheus, and Backstage integration
- **Namespace**: `demo`

## Components

### Deployment
- 2 replicas for high availability
- Resource limits: 100m CPU, 64Mi memory
- Prometheus annotations for metrics scraping

### Service
- ClusterIP service on port 80
- Routes to container port 8080
- Named port `http-backend` for ServiceMonitor

### HTTPRoute (Gateway API)
- Path: `/demo`
- Exposed via `lab-gateway`
- Cross-namespace access via ReferenceGrant

### ServiceMonitor (Prometheus)
- Scrapes metrics every 30s
- Path: `/metrics` (note: http-echo doesn't expose real metrics, but setup is correct)

### Catalog Info (Backstage)
- Component: `demo-app`
- System: `k8s-lab`
- Annotations for ArgoCD and Kubernetes plugin integration

## Access

### Via Gateway API
```bash
# Port forward the gateway
kubectl port-forward svc/envoy-default-lab-gateway-cb012393 -n envoy-gateway-system 8080:80

# Access the app
curl http://localhost:8080/demo
```

### Direct access
```bash
# Port forward the service
kubectl port-forward svc/demo-app -n demo 8081:80

# Access directly
curl http://localhost:8081
```

## ArgoCD Integration

The app is managed by ArgoCD via the app-of-apps pattern.

### View in ArgoCD
```bash
# List applications
kubectl get applications -n argocd

# Get app details
kubectl describe application demo-app -n argocd

# Or via UI
kubectl port-forward svc/argocd-server -n argocd 8080:443
# Open: https://localhost:8080
```

### Sync status
```bash
# Check sync status
kubectl get application demo-app -n argocd

# Force sync
kubectl patch application demo-app -n argocd --type merge -p '{"operation":{"initiatedBy":{"username":"admin"},"sync":{"syncStrategy":{"hook":{}}}}}'
```

## Prometheus Integration

### View ServiceMonitor
```bash
kubectl get servicemonitor demo-app -n demo
kubectl describe servicemonitor demo-app -n demo
```

### Check if Prometheus is scraping
```bash
# Port forward Prometheus
kubectl port-forward svc/kube-prometheus-stack-prometheus -n observability 9090:9090

# Open Prometheus UI: http://localhost:9090
# Go to Status > Targets and search for "demo-app"
```

## Backstage Integration

### Register in Backstage

1. Port forward Backstage:
   ```bash
   kubectl port-forward svc/backstage -n backstage 7007:7007
   ```

2. Open http://localhost:7007

3. Click "Create" → "Register Existing Component"

4. Enter URL:
   ```
   https://github.com/illMadeCoder/illm-k8s-lab/blob/main/apps/demo-app/catalog-info.yaml
   ```

5. Click "Analyze" → "Import"

### View in Backstage

After registration, you'll see:
- **Component Overview**: Basic info, links, docs
- **ArgoCD Tab**: Sync status, health, last sync time
- **Kubernetes Tab**: Pod status, logs, resource usage
- **CI/CD Tab**: Deployment history

## Monitoring in Grafana

1. Port forward Grafana:
   ```bash
   kubectl port-forward svc/kube-prometheus-stack-grafana -n observability 3000:80
   ```

2. Open http://localhost:3000 (admin/admin)

3. Go to Dashboards → Kubernetes

4. Search for "demo" namespace or "demo-app" pods

## Troubleshooting

### App not responding
```bash
# Check pods
kubectl get pods -n demo

# Check logs
kubectl logs -n demo -l app=demo-app

# Check service
kubectl get svc demo-app -n demo
```

### ArgoCD out of sync
```bash
# Check diff
kubectl describe application demo-app -n argocd

# Manual sync
kubectl patch application demo-app -n argocd --type merge -p '{"spec":{"syncPolicy":{"automated":{"prune":true,"selfHeal":true}}}}'
```

### Prometheus not scraping
```bash
# Check ServiceMonitor
kubectl get servicemonitor -n demo

# Check Prometheus targets
kubectl port-forward svc/kube-prometheus-stack-prometheus -n observability 9090:9090
# Visit: http://localhost:9090/targets
```
