# ADR-012: Crossplane Experiment Abstraction

## Status

Accepted

## Context

This lab runs experiments across multiple infrastructure providers:

- **Kind**: Local ephemeral clusters for quick iteration
- **Talos**: Persistent hardware home lab cluster
- **Azure AKS**: Cloud Kubernetes for production-like testing
- **AWS EKS**: Multi-cloud comparison

Currently, each provider has separate tooling:
- `task kind:up` / `task kind:down` (imperative, Taskfile-driven)
- `task talos:up` / `task talos:down` (targets existing cluster)
- Cloud clusters would require additional Terraform or CLI tooling

This fragmentation leads to:
1. **Inconsistent UX** - Different commands per provider
2. **Duplicated logic** - Each provider reimplements cluster registration
3. **No single source of truth** - Cluster state scattered across tools
4. **Manual ArgoCD integration** - Cluster secrets created imperatively

## Decision

Use **Crossplane as the universal infrastructure abstraction** for all experiment clusters. A single `ExperimentCluster` Custom Resource Definition (XRD) provides a declarative API that works across all providers.

### Architecture

```
Hub Cluster
├── Crossplane
│   ├── XRD: ExperimentCluster (illm.io/v1alpha1)
│   ├── Compositions:
│   │   ├── vcluster     (local ephemeral - replaces Kind)
│   │   ├── talos        (reference existing hardware)
│   │   ├── aks          (Azure)
│   │   └── eks          (AWS)
│   └── Providers:
│       ├── provider-helm        (deploy vcluster)
│       ├── provider-kubernetes  (create secrets, RBAC)
│       ├── provider-azure       (AKS)
│       └── provider-aws         (EKS)
└── ArgoCD
    └── Auto-registers clusters from Crossplane-created secrets
```

### User Experience

```bash
# All providers use the same interface
kubectl apply -f - <<EOF
apiVersion: illm.io/v1alpha1
kind: ExperimentCluster
metadata:
  name: otel-tutorial
spec:
  provider: vcluster    # or: talos, aks, eks
  scenario: otel-tutorial
  size: small
  compositionSelector:
    matchLabels:
      provider: vcluster  # Must match spec.provider
EOF

# Check status
kubectl get experimentclusters
# NAME            PROVIDER   PHASE   SERVER                              AGE
# otel-tutorial   vcluster   Ready   https://otel-tutorial.exp-otel...   45s

# Cleanup
kubectl delete experimentcluster otel-tutorial
```

### XRD Schema

```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: CompositeResourceDefinition
metadata:
  name: xexperimentclusters.illm.io
spec:
  group: illm.io
  names:
    kind: XExperimentCluster
    plural: xexperimentclusters
  claimNames:
    kind: ExperimentCluster
    plural: experimentclusters
  versions:
    - name: v1alpha1
      served: true
      referenceable: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              required: [provider, scenario]
              properties:
                provider:
                  type: string
                  enum: [vcluster, talos, aks, eks]
                  description: Infrastructure provider
                scenario:
                  type: string
                  description: Experiment scenario name
                size:
                  type: string
                  enum: [small, medium, large]
                  default: small
            status:
              type: object
              properties:
                clusterName:
                  type: string
                server:
                  type: string
                phase:
                  type: string
                  enum: [Provisioning, Ready, Failed, Deleting]
```

## Provider Compositions

### vcluster (Replaces Kind)

**Why vcluster over Kind:**

| Factor | Kind | vcluster |
|--------|------|----------|
| Provisioning | `kind create` (Docker) | Helm chart (Kubernetes-native) |
| Crossplane support | Requires Docker socket | Native via provider-helm |
| Startup time | ~60-90s | ~30-45s |
| Resource isolation | Full (separate containers) | Shared (virtual namespaces) |
| Hub cluster access | Network bridge needed | Same cluster networking |

**Composition pipeline:**

```
1. Create namespace (exp-{name})
       │
       ▼
2. Deploy vcluster via Helm Release
       │
       ▼
3. Wait for vcluster kubeconfig secret (vc-{name})
       │
       ▼
4. Transform kubeconfig → ArgoCD cluster secret
       │
       ▼
5. (Optional) Create ArgoCD Application for scenario
```

### Talos (Reference Existing)

Talos is persistent hardware - no provisioning needed. The composition:

1. Creates an ExternalSecret referencing OpenBao credentials
2. ExternalSecret populates ArgoCD cluster secret
3. Cluster becomes available for targeting

### AKS / EKS (Future)

Use Crossplane's native cloud providers:
- `provider-azure-containerservice` for AKS
- `provider-aws-eks` for EKS

These compositions create full cluster infrastructure (resource groups, VPCs, node pools) declaratively.

## ArgoCD Integration

### Cluster Secret Format

ArgoCD discovers clusters via labeled secrets:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cluster-{experiment-name}
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: cluster  # Required
    experiment: {scenario-name}               # For filtering
    provider: {vcluster|talos|aks|eks}        # For identification
type: Opaque
stringData:
  name: "{experiment-name}"
  server: "https://{api-server-url}"
  config: |
    {
      "tlsClientConfig": {
        "insecure": false,
        "caData": "..."
      }
    }
```

### Application Deployment Options

**Option A: Composition-embedded apps**
- ArgoCD Application created as part of composition
- Atomic: cluster + apps created/deleted together
- Best for self-contained experiments

**Option B: ApplicationSet with cluster selector**
- ApplicationSet watches for clusters with specific labels
- Decoupled: apps deploy automatically when cluster appears
- Best for baseline apps (observability, networking)

## Comparison

| Approach | Pros | Cons |
|----------|------|------|
| **Current (Taskfile per provider)** | Simple, no dependencies | Inconsistent UX, duplicated logic |
| **Terraform** | Declarative, mature | Separate state, not K8s-native |
| **Cluster API** | K8s-native, cloud-focused | Heavy, no vcluster support |
| **Crossplane** | K8s-native, extensible, GitOps-friendly | Learning curve, more abstractions |

### Why Crossplane

| Factor | Reasoning |
|--------|-----------|
| **Kubernetes-native** | Claims are CRs, managed by kubectl, GitOps-compatible |
| **Extensible** | Single XRD, multiple compositions per provider |
| **Existing investment** | Already deployed for data services (Database, Cache, Queue XRDs) |
| **Unified state** | Cluster state in etcd, not scattered across tools |
| **ArgoCD integration** | Compositions can create ArgoCD secrets directly |

## Implementation

### Required Providers

| Provider | Purpose | Version |
|----------|---------|---------|
| `provider-helm` | Deploy vcluster charts | v0.20.0 |
| `provider-kubernetes` | Create secrets, namespaces | v0.16.0 |
| `function-patch-and-transform` | Standard composition patching | v0.9.2 |
| `function-go-templating` | Kubeconfig transformation | v0.11.1 |

### File Structure

```
components/infrastructure/
├── crossplane-providers/
│   ├── provider-helm.yaml
│   ├── provider-kubernetes.yaml
│   └── functions.yaml
├── crossplane-xrds/
│   └── definitions/
│       └── experiment-cluster/
│           ├── definition.yaml
│           ├── composition-vcluster.yaml
│           └── composition-talos.yaml
```

### Taskfile Integration

The Taskfile becomes a thin wrapper:

```yaml
# platform/hub/cluster/Taskfile.yaml  # or experiment-specific Taskfile
tasks:
  up:
    desc: "Deploy experiment: task exp:up -- <scenario> <provider>"
    cmds:
      - |
        kubectl apply -f - <<EOF
        apiVersion: illm.io/v1alpha1
        kind: ExperimentCluster
        metadata:
          name: {{.SCENARIO}}
        spec:
          provider: {{.PROVIDER}}
          scenario: {{.SCENARIO}}
          size: {{.SIZE | default "small"}}
        EOF
      - kubectl wait experimentcluster/{{.SCENARIO}} --for=condition=Ready --timeout=300s

  down:
    desc: "Teardown experiment"
    cmds:
      - kubectl delete experimentcluster {{.SCENARIO}} --ignore-not-found
```

## Consequences

### Positive

- **Unified UX** - Same commands for all providers
- **Declarative** - Cluster state in Git, reconciled by Crossplane
- **Extensible** - Add new providers by creating compositions
- **ArgoCD-native** - Clusters auto-register, apps auto-deploy
- **GitOps-friendly** - ExperimentCluster claims can be committed to Git
- **Learning opportunity** - Explore Crossplane abstraction limits

### Negative

- **Abstraction overhead** - More indirection vs. direct `kind create`
- **Debugging complexity** - Issues may be in XRD, composition, or provider
- **vcluster trade-offs** - Less isolated than Kind (shared hub resources)
- **Provider maturity** - vcluster composition is custom, not upstream

### Trade-offs

| Trade-off | Mitigation |
|-----------|------------|
| vcluster shared resources | Size limits in composition, hub resource quotas |
| Debugging complexity | Status conditions, events, composition revision history |
| Learning curve | ADR documentation, example claims |

## Known Limitations

1. **Composition selection** - Claims must include `compositionSelector.matchLabels.provider` matching `spec.provider`. The Taskfile wrapper handles this automatically.
2. **Kubeconfig transformation** - vcluster YAML kubeconfig must be converted to ArgoCD JSON format
3. **Cleanup ordering** - ArgoCD apps must delete before vcluster; may need finalizer tuning
4. **Secret readiness** - ArgoCD secret creation must wait for vcluster secret
5. **No native TTL** - Crossplane doesn't auto-delete after time; use `kubectl delete`

## Migration Path

1. **Phase 1**: vcluster + Talos compositions (local development)
2. **Phase 2**: AKS/EKS compositions (cloud testing)
3. **Phase 3**: Retire `task kind:*` commands (vcluster replaces Kind)
4. **Phase 4**: Evaluate Crossplane for other lab resources (databases, queues on real cloud)

## References

- [Crossplane Documentation](https://docs.crossplane.io/)
- [vcluster Documentation](https://www.vcluster.com/docs/)
- [ArgoCD Cluster Registration](https://argo-cd.readthedocs.io/en/stable/operator-manual/declarative-setup/#clusters)
- [Crossplane Composition Functions](https://docs.crossplane.io/latest/concepts/composition-functions/)
- [Existing XRDs: Database, Cache, Queue](../../components/infrastructure/crossplane-xrds/)

## Decision Date

2026-01-12
