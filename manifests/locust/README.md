# Locust Load Testing Setup

Distributed Locust deployment with multiple test scenarios for 2-tier architecture experiments.

## Architecture

- **1 Master**: Web UI + orchestration (port 8089)
- **3 Workers**: Execute load tests (scalable)
- **Gateway Access**: Via Envoy Gateway at `locust.local`

## Test Scenarios

### 1. Steady Load (`steady_load.py`)
Constant load with 75% reads, 25% writes.
```bash
kubectl set env deployment/locust-master -n locust LOCUST_LOCUSTFILE=/locust/steady_load.py
kubectl set env deployment/locust-worker -n locust LOCUST_LOCUSTFILE=/locust/steady_load.py
kubectl rollout restart deployment/locust-master deployment/locust-worker -n locust
```

### 2. Spike Test (`spike_test.py`)
Sudden burst load for testing auto-scaling and resilience.
```bash
kubectl set env deployment/locust-master -n locust LOCUST_LOCUSTFILE=/locust/spike_test.py
kubectl set env deployment/locust-worker -n locust LOCUST_LOCUSTFILE=/locust/spike_test.py
kubectl rollout restart deployment/locust-master deployment/locust-worker -n locust
```

### 3. Ramp Up (`ramp_up.py`)
Gradual load increase: 10 → 50 → 100 → 200 users over 4 minutes.
```bash
kubectl set env deployment/locust-master -n locust LOCUST_LOCUSTFILE=/locust/ramp_up.py
kubectl set env deployment/locust-worker -n locust LOCUST_LOCUSTFILE=/locust/ramp_up.py
kubectl rollout restart deployment/locust-master deployment/locust-worker -n locust
```

### 4. API + Storage Mixed (`api_storage_mixed.py`)
Realistic workload: 50% reads, 20% writes, 20% updates, 10% deletes.
```bash
kubectl set env deployment/locust-master -n locust LOCUST_LOCUSTFILE=/locust/api_storage_mixed.py
kubectl set env deployment/locust-worker -n locust LOCUST_LOCUSTFILE=/locust/api_storage_mixed.py
kubectl rollout restart deployment/locust-master deployment/locust-worker -n locust
```

### 5. Health Check (`health_check.py`)
Focused health endpoint testing.
```bash
kubectl set env deployment/locust-master -n locust LOCUST_LOCUSTFILE=/locust/health_check.py
kubectl set env deployment/locust-worker -n locust LOCUST_LOCUSTFILE=/locust/health_check.py
kubectl rollout restart deployment/locust-master deployment/locust-worker -n locust
```

## Access

- **Web UI**: http://locust.local:32266 (via Gateway)
- **Direct**: `kubectl port-forward -n locust svc/locust-master 8089:8089`

## Scaling Workers

```bash
kubectl scale deployment/locust-worker -n locust --replicas=5
```

## Custom Test Scripts

Edit the ConfigMap:
```bash
kubectl edit configmap locust-scripts -n locust
```

Then restart deployments to pick up changes.

## Production Migration Notes

For Azure App Gateway:
1. Update HTTPRoute to use proper hostname
2. Configure TLS/SSL certificates
3. Adjust worker replicas based on node count
4. Set resource limits based on cluster capacity
