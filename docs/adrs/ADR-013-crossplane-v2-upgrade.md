# ADR-013: Crossplane v2 Upgrade

## Status

Accepted (2026-01-28 — Crossplane v2.1.3 deployed on hub cluster)

## Context

This lab currently runs **Crossplane v1.18.1** to manage the `ExperimentCluster` abstraction (see [ADR-012](./ADR-012-crossplane-experiment-abstraction.md)). Crossplane v2.0 was released with significant architectural changes that address several limitations we've encountered:

### Current Pain Points (v1.x)

1. **Wrapper Overhead**: Native K8s resources (ConfigMaps, Secrets, ArgoCD Applications) require `provider-kubernetes` Object wrappers
2. **Cluster-Scoped XRs**: All composite resources are cluster-scoped, limiting namespace isolation
3. **CRD Bloat**: 143 CRDs installed on hub cluster, most unused
4. **Manual ArgoCD Integration**: Cluster registration and app deployment require separate steps
5. **No Day-2 Operations**: Backups, credential rotation, upgrades are imperative

### Crossplane v2 Capabilities

| Feature | v1.x | v2.x |
|---------|------|------|
| **XR Scope** | Cluster-scoped only | Namespace-scoped option |
| **Composition Content** | Crossplane MRs only | Any Kubernetes resource |
| **CRD Management** | All CRDs from providers | Selective via MRDs |
| **Composition Logic** | Native patch-and-transform | Composition functions (gRPC) |
| **Day-2 Operations** | Not supported | Declarative Operations |
| **XR Structure** | Mixed fields | `spec.crossplane.*` for machinery |

## Decision

Upgrade from **Crossplane v1.18.1 to v2.x** to enable:

1. **Namespace-scoped ExperimentClusters** - Full isolation per experiment
2. **Direct K8s resource composition** - ArgoCD Applications in compositions without wrappers
3. **Selective CRD installation** - Only install resources we use via MRDs
4. **Single-claim environments** - One claim creates infra + registers cluster + deploys apps

### Architecture Changes

```
CURRENT (v1.18.1)
─────────────────

  ExperimentCluster claim
         │
         ▼
  ┌─────────────────────────────────────────────────────────┐
  │ Crossplane Composition                                  │
  │                                                         │
  │   ┌─────────────────┐   ┌─────────────────┐            │
  │   │ ResourceGroup   │   │ KubernetesCluster│            │
  │   │ (Azure MR)      │   │ (Azure MR)       │            │
  │   └─────────────────┘   └─────────────────┘            │
  │                                                         │
  │   ┌─────────────────────────────────────────┐          │
  │   │ Object (provider-kubernetes wrapper)    │ ◄── Extra│
  │   │   └─ manifest:                          │    layer │
  │   │        kind: Secret (ArgoCD cluster)    │          │
  │   └─────────────────────────────────────────┘          │
  └─────────────────────────────────────────────────────────┘
         │
         ▼
  [Manual] ArgoCD Application created separately


PROPOSED (v2.x)
───────────────

  ExperimentCluster claim (namespaced)
         │
         ▼
  ┌─────────────────────────────────────────────────────────┐
  │ Crossplane Composition (can compose ANY resource)       │
  │                                                         │
  │   ┌─────────────────┐   ┌─────────────────┐            │
  │   │ ResourceGroup   │   │ KubernetesCluster│            │
  │   │ (Azure MR)      │   │ (Azure MR)       │            │
  │   └─────────────────┘   └─────────────────┘            │
  │                                                         │
  │   ┌─────────────────┐   ┌─────────────────┐            │
  │   │ Secret          │   │ Application     │ ◄── Direct!│
  │   │ (native K8s)    │   │ (ArgoCD CR)     │    No wrap │
  │   └─────────────────┘   └─────────────────┘            │
  └─────────────────────────────────────────────────────────┘
         │
         ▼
  ArgoCD deploys workloads (created by same composition)
```

### Composition Example (v2 Style)

```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: xexperimentclusters.aks.illm.io
spec:
  compositeTypeRef:
    apiVersion: illm.io/v1alpha1
    kind: XExperimentCluster

  resources:
    # Cloud infrastructure
    - name: resource-group
      base:
        apiVersion: azure.upbound.io/v1beta1
        kind: ResourceGroup
        ...

    - name: aks-cluster
      base:
        apiVersion: containerservice.azure.upbound.io/v1beta1
        kind: KubernetesCluster
        ...

    # ArgoCD cluster registration - DIRECT, no wrapper!
    - name: cluster-secret
      base:
        apiVersion: v1
        kind: Secret
        metadata:
          namespace: argocd
          labels:
            argocd.argoproj.io/secret-type: cluster
        type: Opaque
      patches:
        - fromFieldPath: status.atProvider.kubeConfig
          toFieldPath: stringData.config
          transforms:
            - type: string
              string:
                type: Convert
                convert: ToJSON

    # ArgoCD Application - also DIRECT!
    - name: argocd-app
      base:
        apiVersion: argoproj.io/v1alpha1
        kind: Application
        metadata:
          namespace: argocd
        spec:
          source:
            repoURL: https://github.com/illMadeCoder/k8s-ai-cloud-testbed.git
          destination:
            name: ""  # Patched to cluster name
      patches:
        - fromFieldPath: spec.scenario
          toFieldPath: spec.source.path
          transforms:
            - type: string
              string:
                fmt: "experiments/%s/target/argocd"
        - fromFieldPath: status.clusterName
          toFieldPath: spec.destination.name
```

### ManagedResourceDefinitions (MRDs)

Instead of installing all 143 CRDs from provider packages:

```yaml
# Only install what we actually use
apiVersion: pkg.crossplane.io/v1alpha1
kind: ManagedResourceDefinition
metadata:
  name: aks-resources
spec:
  provider: upbound-provider-family-azure
  resources:
    - group: containerservice.azure.upbound.io
      kinds: [KubernetesCluster, KubernetesClusterNodePool]
    - group: azure.upbound.io
      kinds: [ResourceGroup]
    - group: network.azure.upbound.io
      kinds: [VirtualNetwork, Subnet]
```

**Impact**: ~143 CRDs → ~20 CRDs (only those we use)

### MRAP-in-Composition Pattern

Instead of installing all cloud provider CRDs at hub startup (which consumes ~600MB per provider family), we embed MRDs inside compositions. CRDs only activate when an experiment needs them.

```
BEFORE: Static Provider Installation (~2.4GB total)
─────────────────────────────────────────────────
Hub Startup:
  └── Install all providers → All CRDs installed → 600MB × 4 families

AFTER: MRAP-in-Composition (Dynamic Activation)
───────────────────────────────────────────────
Hub Startup:
  └── Install provider-families only → ~30MB each (ghost CRDs)

ExperimentCluster claim (AKS):
  └── Composition activates:
      ├── MRD: azure-containerservice (KubernetesCluster, NodePool)
      ├── MRD: azure-network (VirtualNetwork, Subnet)
      └── MRD: azure-resources (ResourceGroup)

      Total: ~29MB instead of ~600MB
```

**Implementation**: Compositions include MRD resources that Crossplane creates alongside infrastructure:

```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: xexperimentclusters.aks.illm.io
spec:
  compositeTypeRef:
    apiVersion: illm.io/v1alpha1
    kind: XExperimentCluster

  resources:
    # Activate only the CRDs this composition needs
    - name: mrd-azure-containerservice
      base:
        apiVersion: pkg.crossplane.io/v1alpha1
        kind: ManagedResourceDefinition
        spec:
          providerRef:
            name: provider-family-azure
          resources:
            - group: containerservice.azure.upbound.io
              kinds: [KubernetesCluster, KubernetesClusterNodePool]

    - name: mrd-azure-network
      base:
        apiVersion: pkg.crossplane.io/v1alpha1
        kind: ManagedResourceDefinition
        spec:
          providerRef:
            name: provider-family-azure
          resources:
            - group: network.azure.upbound.io
              kinds: [VirtualNetwork, Subnet]

    # Now the actual infrastructure (CRDs activated above)
    - name: resource-group
      base:
        apiVersion: azure.upbound.io/v1beta1
        kind: ResourceGroup
        # ...
```

**Benefits**:
- Memory: ~29MB per active experiment vs ~600MB per provider family
- Isolation: Each experiment activates only what it needs
- Cleanup: MRDs deleted when experiment is deleted → CRDs deactivate
- Multi-cloud: Can run AKS, EKS, GKE experiments simultaneously with minimal overhead

**Trade-off**: First experiment using a cloud takes ~30s longer (CRD activation). Subsequent experiments using same CRDs are instant.

### Namespace-Scoped Resources

```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: CompositeResourceDefinition
metadata:
  name: xexperimentclusters.illm.io
spec:
  group: illm.io
  names:
    kind: XExperimentCluster
  claimNames:
    kind: ExperimentCluster
    namespaced: true           # NEW in v2: XR lives in namespace
```

User experience:

```bash
# Create experiment in a namespace
kubectl apply -n experiments -f claim.yaml

# List experiments (namespace-scoped)
kubectl get experimentclusters -n experiments

# Each namespace is isolated
kubectl get experimentclusters -A
NAMESPACE      NAME                  PROVIDER   STATUS
experiments    gateway-comparison    aks        Ready
team-alpha     load-testing          eks        Ready
team-beta      security-audit        aks        Provisioning
```

### RBAC for Native Resources

Crossplane v2 needs explicit permission to manage non-MR resources:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: crossplane-compose-argocd
  labels:
    rbac.crossplane.io/aggregate-to-crossplane: "true"
rules:
  - apiGroups: ["argoproj.io"]
    resources: ["applications", "applicationsets"]
    verbs: ["*"]
  - apiGroups: [""]
    resources: ["secrets", "configmaps"]
    verbs: ["*"]
```

## ArgoCD Integration Pattern

### The Chain (Not Circular)

```
Git                          Hub Cluster                    Target Cluster
────                         ───────────                    ──────────────

scenarios/gateway/
├── claim.yaml ─────────────► ArgoCD syncs claim
│                                    │
│                                    ▼
│                             ExperimentCluster
│                                    │
│                                    │ Crossplane reconciles
│                                    ▼
│                             ┌─────────────────────┐
│                             │ • AKS Cluster       │──► Azure
│                             │ • Cluster Secret    │
│                             │ • ArgoCD App ───────┼──────────────┐
│                             └─────────────────────┘              │
│                                                                  │
└── target/argocd/ ◄───────────────────────────────────────────────┘
    ├── app.yaml                    ArgoCD deploys apps
    ├── nginx.yaml                  to target cluster
    └── envoy.yaml                        │
                                          ▼
                                   ┌─────────────────┐
                                   │ Deployments     │
                                   │ Services        │
                                   │ HTTPRoutes      │
                                   └─────────────────┘

RESULT: Full ArgoCD visibility into workloads
```

### What Shows in ArgoCD UI

```
Applications:
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                          │
│   gateway-infra           Synced ✓   Healthy ✓   (hub cluster)          │
│   └── ExperimentCluster   Synced ✓   Ready ✓                            │
│                                                                          │
│   gateway-apps            Synced ✓   Healthy ✓   (aks-gateway cluster)  │ ◄── Created by
│   ├── Deployment/echo-v1  Synced ✓   Healthy ✓   3/3 pods              │     Crossplane!
│   ├── Deployment/echo-v2  Synced ✓   Healthy ✓   3/3 pods              │
│   ├── Service/echo-v1     Synced ✓                                      │
│   └── HTTPRoute/...       Synced ✓                                      │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

## Day-2 Operations (Future)

Crossplane v2 introduces declarative Operations:

```yaml
apiVersion: apiextensions.crossplane.io/v1alpha1
kind: Operation
metadata:
  name: rotate-azure-credentials
  namespace: experiments
spec:
  trigger:
    schedule: "0 0 * * 0"  # Weekly
  compositeRef:
    apiVersion: illm.io/v1alpha1
    kind: XExperimentCluster
    name: gateway-comparison
  pipeline:
    - step: rotate
      functionRef:
        name: fn-credential-rotate
```

**Use cases:**
- Credential rotation
- Backup/restore
- Kubernetes version upgrades
- Certificate renewal

## Experiment Infrastructure Architecture

### Results Storage

Three-tier storage strategy for experiment results:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Hub Cluster                                     │
│                                                                              │
│  ┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────┐  │
│  │     PostgreSQL      │    │     SeaweedFS       │    │       Git       │  │
│  │    (Structured)     │    │   (Raw Artifacts)   │    │   (Insights)    │  │
│  ├─────────────────────┤    ├─────────────────────┤    ├─────────────────┤  │
│  │ • Run metadata      │    │ • k6 results JSON   │    │ • README.md     │  │
│  │ • Metrics summaries │    │ • Grafana snapshots │    │ • Final report  │  │
│  │ • Budget tracking   │    │ • Pod logs archive  │    │ • Agent insights│  │
│  │ • JSONB queryable   │    │ • Heap dumps        │    │ • Recommendations│ │
│  └─────────────────────┘    └─────────────────────┘    └─────────────────┘  │
│           ▲                          ▲                          ▲           │
│           │                          │                          │           │
│           └──────────────────────────┴──────────────────────────┘           │
│                                      │                                       │
│                           Argo Workflow writes results                       │
└─────────────────────────────────────────────────────────────────────────────┘
```

**PostgreSQL**: Structured, queryable experiment data
- Run metadata (experiment, start/end, provider, cost)
- Metrics summaries (p50/p95/p99 latency, throughput)
- Budget tracking and cost accumulation
- JSONB columns for flexible schema evolution

**SeaweedFS (S3-compatible)**: Raw artifacts
- k6 load test JSON output
- Grafana dashboard snapshots
- Compressed pod logs
- Profiling data, heap dumps

**Git**: Human-readable insights
- Final experiment report (markdown)
- AI-generated analysis and recommendations
- Committed to `experiments/results/{experiment}/{run-id}/`

### Observability Architecture

Grafana Agent on target clusters pushes metrics to hub:

```
Target Cluster                              Hub Cluster
──────────────                              ───────────
┌─────────────────────┐                     ┌─────────────────────┐
│   Experiment Apps   │                     │       Mimir         │
│  ┌───────────────┐  │                     │  (metrics storage)  │
│  │ nginx-ingress │──┼─┐                   └──────────▲──────────┘
│  │ traefik       │  │ │                              │
│  │ envoy-gateway │  │ │    Push                      │
│  └───────────────┘  │ │                   ┌──────────┴──────────┐
│                     │ ├──────────────────►│   Grafana Agent     │
│  ┌───────────────┐  │ │   remote_write    │   (on hub)          │
│  │ Grafana Agent │──┼─┘                   └─────────────────────┘
│  │   (~100MB)    │  │
│  └───────────────┘  │                     ┌─────────────────────┐
│                     │                     │      Grafana        │
└─────────────────────┘                     │  (all experiments)  │
                                            └─────────────────────┘
```

**Benefits:**
- Lightweight: ~100MB vs ~1.5GB for full Prometheus+Grafana stack
- Faster experiment startup (no observability deployment)
- Single Grafana dashboard sees ALL experiments
- Metrics persist after experiment teardown

**Grafana Agent Config (target cluster):**
```yaml
metrics:
  wal_directory: /tmp/wal
  global:
    scrape_interval: 15s
    remote_write:
      - url: http://mimir.hub-cluster:9009/api/v1/push
        headers:
          X-Scope-OrgID: experiments
```

### AI Agent Integration

Claude (running the experiment) monitors and writes the final report:

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Claude Session                                │
│                                                                      │
│   1. kubectl apply -f claim.yaml        # Start experiment          │
│                     │                                                │
│                     ▼                                                │
│   2. Watch: kubectl get experimentcluster -w                        │
│              Crossplane provisions → Argo Workflow runs             │
│                     │                                                │
│                     ▼                                                │
│   3. Monitor progress:                                              │
│      - argo get @latest                                             │
│      - kubectl logs -f workflow-pod                                 │
│      - Query PostgreSQL for interim results                         │
│                     │                                                │
│                     ▼                                                │
│   4. Workflow completes:                                            │
│      - Fetch results from PostgreSQL                                │
│      - Fetch artifacts from SeaweedFS                               │
│      - Analyze metrics, compare against budgets                     │
│                     │                                                │
│                     ▼                                                │
│   5. Write report to Git:                                           │
│      experiments/results/gateway-comparison/run-001/README.md       │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

**No separate AI agent service needed** - Claude orchestrates the entire lifecycle.

### Budget Tracking

Two-tier budget enforcement:

```yaml
# Per-experiment (in XRD spec)
apiVersion: illm.io/v1alpha1
kind: ExperimentCluster
metadata:
  name: gateway-comparison
  namespace: experiments
spec:
  provider: aks
  budget:
    maxCostUsd: 50        # Hard limit for this experiment
    alertThresholdPct: 80 # Alert at 80% consumed
```

```yaml
# Global limits (ConfigMap on hub)
apiVersion: v1
kind: ConfigMap
metadata:
  name: experiment-budget-limits
  namespace: crossplane-system
data:
  azure.monthlyLimitUsd: "200"
  aws.monthlyLimitUsd: "200"
  gcp.monthlyLimitUsd: "200"
  global.monthlyLimitUsd: "500"
```

**Enforcement flow:**
1. Composition function checks budget ConfigMap before provisioning
2. If global limit exceeded → reject claim with status condition
3. Cost accumulates in PostgreSQL per experiment run
4. Argo Workflow step checks budget mid-run, can abort if exceeded

## Migration Path

### Phase 1: Preparation (Pre-Upgrade)

1. **Audit existing compositions** - Identify provider-kubernetes Object wrappers
2. **Update package references** - Ensure fully-qualified registry names
3. **Test in isolated environment** - Validate compositions work with v2
4. **Document rollback procedure** - Crossplane supports in-place downgrade

### Phase 2: Upgrade

```bash
# Upgrade Crossplane
helm upgrade crossplane crossplane-stable/crossplane \
  --namespace crossplane-system \
  --version 2.1.0

# Verify
kubectl get pods -n crossplane-system
crossplane version
```

### Phase 3: Composition Updates

1. **Remove Object wrappers** - Replace with direct K8s resources
2. **Add RBAC ClusterRoles** - Grant Crossplane permission for native resources
3. **Update XRDs** - Add `namespaced: true` to claim names
4. **Test each composition** - Verify resources create correctly

### Phase 4: MRD Implementation

1. **Identify used CRDs** - `kubectl get crd | grep upbound`
2. **Create MRDs** - Define only needed resources
3. **Update provider configs** - Enable MRD-based installation
4. **Verify CRD reduction** - Confirm unused CRDs removed

### Rollback Procedure

```bash
# If issues arise, downgrade is supported
helm upgrade crossplane crossplane-stable/crossplane \
  --namespace crossplane-system \
  --version 1.18.1

# Legacy compositions continue to work
# v1-style XRs and MRs are backward compatible
```

## Comparison: Crossplane vs Alternatives

### Helm/Kustomize

| Aspect | Helm/Kustomize | Crossplane v2 |
|--------|----------------|---------------|
| Execution | Client-side, once | Server-side, continuous |
| Drift detection | No | Yes (reconciles) |
| Abstraction | Partial (values) | Full (XRD hides implementation) |
| Cloud resources | No | Yes (via providers) |
| Package ecosystem | Huge | Growing |

**Verdict**: Not a replacement. Crossplane can compose Helm releases within compositions.

### ArgoCD

| Aspect | ArgoCD | Crossplane v2 |
|--------|--------|---------------|
| Source of truth | Git | Kubernetes CR |
| Primary job | Deploy manifests | Abstract & compose |
| Git integration | Native | None |
| UI | Yes | No |

**Verdict**: Complementary. ArgoCD delivers claims, Crossplane expands them.

## Consequences

### Positive

- **Single-claim environments** - One CR creates infra + apps + monitoring
- **Full namespace isolation** - Experiments don't see each other's resources
- **Reduced CRD footprint** - Only install what we use
- **Native K8s composition** - No more Object wrappers
- **ArgoCD visibility preserved** - Compositions create ArgoCD Apps
- **Day-2 automation** - Declarative operations for maintenance tasks

### Negative

- **Migration effort** - Existing compositions need updates
- **Learning curve** - New concepts (MRDs, Operations, namespaced XRs)
- **Composition functions required** - Native patch-and-transform deprecated
- **Ecosystem maturity** - v2 is newer, fewer examples available

### Trade-offs

| Trade-off | Mitigation |
|-----------|------------|
| Migration complexity | Phased approach, test in isolation first |
| Composition function learning | Start with `function-patch-and-transform` (compatible) |
| Fewer community examples | Document our patterns in lab |

## Implementation Checklist

### Crossplane v2 Upgrade
- [ ] Review Crossplane v2 upgrade guide
- [ ] Audit current compositions for Object wrappers
- [ ] Test upgrade in isolated Kind cluster
- [ ] Create RBAC ClusterRoles for native resources
- [ ] Update `composition-aks.yaml` to include ArgoCD App
- [ ] Update `composition-vcluster.yaml` similarly
- [ ] Create MRDs for Azure, AWS, GCP resources
- [ ] Update XRD to support namespaced claims
- [ ] Test full experiment lifecycle
- [ ] Update Taskfile wrappers if needed

### Hub Infrastructure
- [ ] Deploy PostgreSQL on hub cluster (experiment results DB)
- [ ] Deploy SeaweedFS on hub cluster (S3-compatible artifacts)
- [ ] Deploy Mimir on hub cluster (metrics aggregation)
- [ ] Create experiment-budget-limits ConfigMap
- [ ] Create PostgreSQL schema for experiment results

### Observability
- [ ] Create Grafana Agent Helm values for target clusters
- [ ] Configure remote_write to hub Mimir
- [ ] Create multi-experiment Grafana dashboard

### Workflow Templates
- [ ] Create `tutorial-interactive` workflow template
- [ ] Create `benchmark-k6` workflow template
- [ ] Create `benchmark-comparison` workflow template
- [ ] Create `soak-test` workflow template

### Documentation
- [ ] Document new patterns in lab
- [ ] Create experiment results directory structure

## References

- [Crossplane v2 What's New](https://docs.crossplane.io/latest/whats-new/)
- [Crossplane v2 Upgrade Guide](https://docs.crossplane.io/latest/guides/upgrade-to-crossplane-v2/)
- [Crossplane 2.0 Announcement](https://blog.crossplane.io/announcing-crossplane-2-0/)
- [ADR-012: Crossplane Experiment Abstraction](./ADR-012-crossplane-experiment-abstraction.md)
- [Composition Functions](https://docs.crossplane.io/latest/concepts/composition-functions/)

## Decision Date

2026-01-17 (Proposed)
