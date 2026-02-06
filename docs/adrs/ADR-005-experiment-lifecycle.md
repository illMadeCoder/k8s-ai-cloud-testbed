# ADR-005: Experiment Lifecycle with Automatic Cleanup

## Status

Accepted

## Context

The platform runs experiments on ephemeral Kind clusters provisioned on-demand. Each experiment requires:

1. **Cluster provisioning**: Create a Kind cluster for the target workload
2. **Resource deployment**: Deploy ArgoCD Applications for the experiment
3. **Test execution**: Run Argo Workflows to conduct the experiment
4. **Cleanup**: Remove all resources when the experiment completes

Without proper cleanup, orphaned resources accumulate:
- ArgoCD Applications remain after experiments finish
- Kind clusters consume resources indefinitely
- ConfigMaps and secrets clutter namespaces

**Problems to solve:**
1. Automate cleanup after experiment completion (success or failure)
2. Support idempotent re-runs (cleanup before new run)
3. Maintain clear separation between orchestrator and target clusters

## Decision

**Use Argo Workflows `onExit` handlers for Kubernetes resource cleanup, combined with Taskfile for cluster lifecycle.**

### Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                      EXPERIMENT LIFECYCLE                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  1. task exp:conduct:<name>                                          │
│     │                                                                 │
│     ├── Create Kind cluster (target)                                 │
│     ├── Register cluster in ArgoCD (secret)                         │
│     ├── Apply ArgoCD Applications                                    │
│     └── Submit Argo Workflow                                         │
│                                                                       │
│  2. Workflow executes                                                │
│     │                                                                 │
│     ├── wait-for-app: Poll until ArgoCD app healthy                 │
│     ├── k6-test: Run load test against target                       │
│     └── report-results: Log completion                               │
│                                                                       │
│  3. Workflow exits (success or failure)                              │
│     │                                                                 │
│     └── onExit: cleanup template                                     │
│         ├── Delete ArgoCD Applications                               │
│         ├── Delete k6 ConfigMap                                      │
│         └── Delete cluster secret                                    │
│                                                                       │
│  4. task exp:teardown:<name>                                         │
│     │                                                                 │
│     └── Delete Kind cluster                                          │
│                                                                       │
└─────────────────────────────────────────────────────────────────────┘
```

### Workflow Pattern

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: experiment-name-exp-
  namespace: argo-workflows
spec:
  serviceAccountName: argo-workflow
  entrypoint: run-experiment
  onExit: cleanup                    # Always runs, even on failure
  arguments:
    parameters:
      - name: experiment-name
        value: "experiment-name"

  templates:
    - name: run-experiment
      steps:
        - - name: wait-for-app
            template: wait-argocd-healthy
        - - name: test
            template: k6-test
        - - name: report
            template: complete

    - name: cleanup
      container:
        image: bitnami/kubectl:latest
        command: [sh, -c]
        args:
          - |
            EXP="{{workflow.parameters.experiment-name}}"
            kubectl delete application ${EXP}-target -n argocd --ignore-not-found=true
            kubectl delete application ${EXP}-loadgen -n argocd --ignore-not-found=true
            kubectl delete configmap k6-scripts -n argo-workflows --ignore-not-found=true
            kubectl delete secret cluster-target -n argocd --ignore-not-found=true
```

### RBAC Requirements

The workflow service account needs permissions to manage resources across namespaces:

```yaml
# ClusterRole: argocd-app-manager
rules:
  - apiGroups: ["argoproj.io"]
    resources: ["applications"]
    verbs: ["get", "list", "watch", "delete"]
  - apiGroups: [""]
    resources: ["secrets", "configmaps"]
    verbs: ["get", "list", "delete"]

# ClusterRole: pod-reader (for wait-for-app)
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list", "watch"]
```

### Taskfile Integration

```yaml
# Taskfile.yml
exp:conduct:http-baseline:
  desc: "Run http-baseline experiment (provision → test → cleanup)"
  cmds:
    - task: exp:deploy:http-baseline    # Create cluster, register, apply apps
    - task: exp:run:http-baseline       # Submit workflow
    - task: exp:teardown:http-baseline  # Delete Kind cluster
```

### Resource Ownership

| Resource | Created By | Deleted By |
|----------|------------|------------|
| Kind cluster | Taskfile | Taskfile |
| Cluster secret | Taskfile | Workflow cleanup |
| ArgoCD Applications | Taskfile | Workflow cleanup |
| k6 ConfigMap | ArgoCD | Workflow cleanup |
| Target deployments | ArgoCD | ArgoCD (Application deleted) |

## Consequences

### Positive

- **No orphaned resources**: Cleanup runs regardless of experiment outcome
- **Idempotent**: Re-running cleans up first via `--ignore-not-found`
- **Observable**: Workflow logs show cleanup progress
- **Separation of concerns**: Workflows handle K8s resources, Taskfile handles clusters

### Negative

- **Partial cleanup on Taskfile interrupt**: If `task exp:conduct` is killed mid-workflow, cleanup may not run
- **RBAC complexity**: Service account needs broad permissions
- **Two cleanup mechanisms**: Workflow onExit + Taskfile teardown

### Trade-offs

| Approach | Auto Cleanup | Cluster Cleanup | Complexity |
|----------|--------------|-----------------|------------|
| Manual | No | No | Low |
| Workflow only | Yes (K8s) | No | Medium |
| **Workflow + Taskfile** | Yes (K8s) | Yes | Higher |
| Operator pattern | Yes | Yes | Highest |

## Experiment Structure

```
experiments/<experiment-name>/
├── loadgen/
│   └── argocd/
│       ├── app.yaml         # ArgoCD Application for k6 ConfigMap
│       └── k6-scripts.yaml  # k6 test scripts ConfigMap
├── target/
│   └── argocd/
│       ├── app.yaml         # ArgoCD Application for target workload
│       └── nodeport-service.yaml  # Cross-cluster access
└── workflow/
    └── experiment.yaml      # Argo Workflow with onExit cleanup
```

### Labels Convention

ArgoCD Applications use consistent labels for filtering:

```yaml
labels:
  experiment: <experiment-name>
  cluster: target|loadgen
```

## Files

```
components/
└── workflows/
    └── argo-workflows/
        └── argocd-reader-rbac.yaml   # RBAC for cleanup

experiments/
├── hello-app/
│   ├── loadgen/argocd/...
│   ├── target/argocd/...
│   └── workflow/experiment.yaml
└── http-baseline/
    ├── loadgen/argocd/...
    ├── target/argocd/...
    └── workflow/experiment.yaml
```

## References

- [Argo Workflows Exit Handlers](https://argo-workflows.readthedocs.io/en/latest/walk-through/exit-handlers/)
- [ArgoCD Application Finalizers](https://argo-cd.readthedocs.io/en/stable/user-guide/app_deletion/)
- [Kind Cluster Lifecycle](https://kind.sigs.k8s.io/docs/user/quick-start/)

## Decision Date

2025-12-30
