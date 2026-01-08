# Troubleshooting & Common Gotchas

Known issues encountered during experiment development and their solutions.

## Helm Chart Issues

### kube-prometheus-stack CRD Size Limit

**Symptom**: ArgoCD sync fails with:
```
metadata.annotations: Too long: must have at most 262144 bytes
```

**Cause**: kube-prometheus-stack CRDs are large and exceed kubectl's annotation limit for last-applied-configuration.

**Solution**: Use `skipCrds: true` in Helm values and apply CRDs separately:
```yaml
# In ArgoCD Application helm config
helm:
  skipCrds: true
```

Apply CRDs manually with server-side apply:
```bash
kubectl apply --server-side --force-conflicts -f \
  https://raw.githubusercontent.com/prometheus-community/helm-charts/main/charts/kube-prometheus-stack/charts/crds/crds/crd-servicemonitors.yaml
```

**Affected Charts**: kube-prometheus-stack, cert-manager (large installations)

### Victoria Metrics Minimum Retention Period

**Symptom**: Victoria Metrics pod crashes with:
```
-retentionPeriod cannot be smaller than a day; got 2h
```

**Cause**: Victoria Metrics requires minimum 1 day retention period.

**Solution**: Set `retentionPeriod: 1d` or higher:
```yaml
server:
  retentionPeriod: 1d  # Minimum required
```

### Mimir Distributed Chart Has No Monolithic Mode

**Symptom**: Deploying mimir-distributed with `deploymentMode: monolithic` still creates all distributed components.

**Cause**: The mimir-distributed Helm chart doesn't support true monolithic mode despite documentation suggesting otherwise.

**Solution**: Either:
1. Use all distributed components (resource-heavy, not suitable for Kind)
2. Use Victoria Metrics instead for resource-constrained environments
3. Run Mimir externally or skip in tutorials

---

## Port Conflicts

### Node Exporter DaemonSet Conflicts

**Symptom**: Node exporter pods stuck in `Pending` with:
```
0/1 nodes are available: 1 node(s) didn't have free ports for the requested pod ports
```

**Cause**: Multiple prometheus stacks deploying node-exporter DaemonSets that compete for host port 9100.

**Solutions**:

1. **Disable node-exporter** in secondary prometheus installations:
```yaml
nodeExporter:
  enabled: false
```

2. **Use different host ports** (requires custom configuration)

3. **Single shared observability stack** for the cluster

**Best Practice**: Only enable node-exporter in one prometheus installation per cluster.

---

## ArgoCD Issues

### Conflicting Applications

**Symptom**: Pods constantly terminating/recreating, ArgoCD apps showing "Unknown" sync status, resources with `requiresPruning: true`.

**Cause**: Multiple ArgoCD Applications deploying similar resources (e.g., two prometheus operators).

**Solutions**:

1. **One cluster per tutorial** - use separate Kind clusters for each experiment
2. **Delete conflicting apps** before deploying new ones:
```bash
kubectl delete application conflicting-app -n argocd
```
3. **Check for existing deployments** before creating new experiments:
```bash
argocd app list
kubectl get prometheus -A
```

### ArgoCD Server CrashLoopBackOff

**Symptom**: `argocd-server` pod in CrashLoopBackOff, CLI commands fail with connection refused.

**Cause**: Various (memory, configuration, cluster state)

**Solution**: Delete the pod to trigger recreation:
```bash
kubectl delete pod -n argocd -l app.kubernetes.io/name=argocd-server
```

If persistent, check logs:
```bash
kubectl logs -n argocd -l app.kubernetes.io/name=argocd-server --previous
```

---

## Container Image Issues

### Benchmark/Debug Pods Missing Tools

**Symptom**: Scripts fail with `jq: not found` or similar.

**Cause**: Minimal images like `curlimages/curl` don't include tools like `jq`, `awk`, etc.

**Solution**: Use `nicolaka/netshoot` for debug/benchmark pods:
```yaml
containers:
  - name: runner
    image: nicolaka/netshoot:latest
    command: ["sleep", "infinity"]
```

`netshoot` includes: curl, jq, dig, nslookup, tcpdump, netstat, iperf, etc.

---

## Kind Cluster Issues

### Insufficient Resources for Multiple TSDBs

**Symptom**: Pods OOMKilled or stuck in Pending.

**Cause**: Running 3 TSDBs + supporting infrastructure exceeds default Kind resources.

**Solution**: Use medium or large cluster size:
```yaml
# cluster.yaml
size: medium  # or large
```

Resource estimates for TSDB comparison:
| Component | RAM |
|-----------|-----|
| Prometheus | ~500MB |
| Victoria Metrics | ~200MB |
| Mimir (minimal) | ~500MB |
| MinIO | ~200MB |
| Grafana | ~200MB |
| **Total** | ~1.6GB |

---

## Checklist Before Creating New Experiments

1. **Check existing deployments** on target cluster:
   ```bash
   kubectl get prometheus -A
   kubectl get daemonsets -A | grep node-exporter
   argocd app list -l cluster=target
   ```

2. **Check for port conflicts** if using host ports

3. **Verify helm chart defaults** - don't assume sensible minimums (e.g., Victoria Metrics retention)

4. **Use `skipCrds: true`** for large CRD helm charts

5. **Choose appropriate debug image** (`nicolaka/netshoot` for scripts needing tools)

---

## Related Documentation

- [GitOps Patterns](gitops-patterns.md)
- [ADR-009: TSDB Selection](adrs/ADR-009-tsdb-selection.md)
