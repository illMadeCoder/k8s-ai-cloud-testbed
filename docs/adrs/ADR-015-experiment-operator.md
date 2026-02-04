# ADR-015: Experiment Operator

## Status

Proposed

## Context

The lab runs experiments across multiple infrastructure providers. The gateway-tutorial is the first experiment running off-cluster on GKE provisioned via Crossplane. Setting it up required manual steps:

1. Create `GKECluster` claim (Crossplane provisions VPC, subnet, cluster, node pool)
2. Register cluster in ArgoCD (create labeled Secret with kubeconfig)
3. Install Tailscale operator on GKE
4. Deploy Envoy Gateway via ArgoCD Application
5. Deploy Grafana Alloy for metrics collection
6. Create Tailscale egress service for metrics backhaul to hub Mimir
7. Create Tailscale-exposed Mimir service on hub
8. Configure Alloy remote-write with `X-Scope-OrgID` header

This manual orchestration is error-prone (9 commits to fix the metrics pipeline alone) and not reproducible. Each new off-cluster experiment would repeat the same setup.

### Current State

- **15 experiment scenarios** under `experiments/scenarios/`
- **Crossplane v2.1.3** with Pipeline mode compositions (`function-patch-and-transform`, `function-go-templating`, `function-auto-ready`)
- **Existing XRDs**: `XGKECluster` (GKE provisioning), `XObjectStorage` (multi-cloud storage)
- **ADR-012** designed `XExperimentCluster` for unified cluster provisioning but did not address app deployment or observability
- **RBAC for native resource composition** already in place (`crossplane-compose-native-resources` ClusterRole grants Crossplane access to Secrets, ArgoCD Applications, Argo Workflows)
- **ArgoCD multi-source pattern** used by all experiments (Helm charts + Git values + Git manifests)
- **Argo Workflows** deployed on hub for validation workflows

## Decision

Use a **Crossplane `XExperiment` XRD** that creates the full experiment environment from a single claim: cluster infrastructure, ArgoCD registration, app deployment, and observability pipeline.

This elevates ADR-012's `XExperimentCluster` (cluster-only) to `XExperiment` (cluster + apps + observability).

### Architecture

```
kubectl apply -f claim.yaml
        |
        v
  Crossplane Composition (Pipeline mode)
  +------------------------------------------------------+
  |  step 1: function-patch-and-transform                |
  |    +-- GKE Cluster (VPC + Subnet + Cluster + NP)     |
  |    +-- ArgoCD cluster Secret (native K8s resource)   |
  |    +-- ArgoCD Application (scenario apps)            |
  |                                                       |
  |  step 2: function-go-templating                      |
  |    +-- Kubeconfig transform (conn secret -> ArgoCD   |
  |        JSON format)                                  |
  |                                                       |
  |  step 3: function-auto-ready                         |
  +------------------------------------------------------+
        |                    |                |
        v                    v                v
    GCP (GKE)         ArgoCD deploys      Hub Mimir/Loki
    real cluster       apps to target      receives telemetry
```

### XRD Schema

```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: CompositeResourceDefinition
metadata:
  name: xexperiments.illm.io
spec:
  group: illm.io
  names:
    kind: XExperiment
    plural: xexperiments
  claimNames:
    kind: Experiment
    plural: experiments
  connectionSecretKeys:
    - kubeconfig
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
              required: [scenario]
              properties:
                scenario:
                  type: string
                  description: "Scenario directory under experiments/scenarios/"
                cluster:
                  type: object
                  properties:
                    provider:
                      type: string
                      enum: [gke, vcluster, talos]
                      default: gke
                    zone:
                      type: string
                      default: us-central1-a
                    machineType:
                      type: string
                      default: e2-medium
                    nodeCount:
                      type: integer
                      default: 2
                    preemptible:
                      type: boolean
                      default: true
                observability:
                  type: object
                  properties:
                    enabled:
                      type: boolean
                      default: true
                    transport:
                      type: string
                      enum: [direct, tailscale]
                      default: tailscale
                ttlMinutes:
                  type: integer
                  default: 0
                  description: "Auto-delete after N minutes. 0 = no TTL."
            status:
              type: object
              properties:
                phase:
                  type: string
                clusterName:
                  type: string
                endpoint:
                  type: string
      additionalPrinterColumns:
        - name: Scenario
          type: string
          jsonPath: .spec.scenario
        - name: Provider
          type: string
          jsonPath: .spec.cluster.provider
        - name: Phase
          type: string
          jsonPath: .status.phase
        - name: Age
          type: date
          jsonPath: .metadata.creationTimestamp
```

### User Experience

```bash
# Deploy full experiment environment
kubectl apply -f experiments/scenarios/gateway-tutorial/claim.yaml

# Watch progress
kubectl get experiments -n experiments -w
# NAME               SCENARIO           PROVIDER   PHASE          AGE
# gateway-tutorial   gateway-tutorial   gke        Provisioning   10s
# gateway-tutorial   gateway-tutorial   gke        Ready          9m

# Check what was created
kubectl get applications -n argocd -l experiment=gateway-tutorial

# Tear down everything (cluster, apps, secrets)
kubectl delete experiment gateway-tutorial -n experiments
```

### Composition Design

The GKE composition creates these resources from a single claim:

**1. GKE Infrastructure** (inline from existing `xgkeclusters.gcp.illm.io` composition):
- VPC Network (`compute.gcp.upbound.io/v1beta1 Network`)
- Subnetwork with pod/service secondary ranges
- GKE Cluster (zonal, VPC-native, no default node pool, deletion protection off)
- Node Pool (preemptible, shielded instances, auto-repair/upgrade)

**2. ArgoCD Cluster Secret** (Crossplane v2 native resource composition):
- Secret in `argocd` namespace with `argocd.argoproj.io/secret-type: cluster` label
- `stringData.name` = experiment name (ArgoCD uses this as the cluster display name)
- `stringData.server` = GKE API endpoint (from cluster status)
- `stringData.config` = JSON with TLS client config (transformed from kubeconfig via `function-go-templating`)

**3. ArgoCD Application** (Crossplane v2 native resource composition):
- Points to `experiments/scenarios/{scenario}/target/argocd/` as directory source
- Uses `directory.recurse: true` to discover all Application manifests in the scenario
- Labels with `experiment: {scenario}` for filtering

### Kubeconfig Handling

The GKE Crossplane provider writes a connection secret with a `kubeconfig` key (YAML format). ArgoCD requires a JSON `config` field with `tlsClientConfig`. The `function-go-templating` pipeline step transforms between these formats, extracting `certificate-authority-data`, `client-certificate-data`, and `client-key-data` from the kubeconfig.

### Destination Name Compatibility

Existing scenario `app.yaml` files use `destination.name: target` (hardcoded). For v1, the composition creates the ArgoCD cluster secret with `stringData.name: target`, matching existing conventions. This limits concurrency to one experiment at a time, which is acceptable for a home lab.

For multi-experiment support in the future, `function-go-templating` can patch `destination.name` in child Application manifests at deploy time, or scenarios can migrate to claim-based destination names.

### Observability Integration

The composition optionally deploys telemetry collection on the target cluster:

- **Tailscale transport** (for GKE/AKS/EKS): Alloy pushes metrics through Tailscale egress to hub Mimir/Loki. Reuses the pattern from the manual gateway experiment (egress service with `tailscale.com/tailnet-fqdn` annotation).
- **Direct transport** (for vcluster/same-network): Alloy pushes directly to hub services via Kubernetes DNS.
- **Tenant isolation**: `X-Scope-OrgID: {experiment-name}` header on all writes to Mimir/Loki. Each experiment's telemetry is queryable through a dedicated Grafana datasource.

### Cleanup Ordering

When the Experiment claim is deleted, Crossplane deletes all composed resources. If the GKE cluster deletes before the ArgoCD Application, the Application's `resources-finalizer.argocd.argoproj.io` finalizer hangs (can't delete resources on a dead cluster).

Mitigation: the composition sets `finalizers: []` on the ArgoCD Application, removing the resource cleanup finalizer. Orphaned resources on a deleted cluster are acceptable — they disappear with the cluster.

### File Structure

```
experiments/components/infrastructure/crossplane-xrds/definitions/
  experiment/
    definition.yaml              # XExperiment XRD
    composition-gke.yaml         # GKE provider composition
    composition-vcluster.yaml    # vcluster composition (Phase 2)

experiments/scenarios/gateway-tutorial/
  claim.yaml                     # Experiment claim (new entry point)
  target/argocd/app.yaml         # Existing ArgoCD apps (unchanged)
  manifests/                     # Existing manifests (unchanged)
  workflow/experiment.yaml       # Existing validation workflow (unchanged)
```

## Alternatives Considered

### Go Operator (Kubebuilder)

Custom CRDs with a Go controller that orchestrates cluster provisioning, ArgoCD registration, app deployment, and validation.

**Rejected because:** Introduces Go compilation, operator lifecycle management (OLM), and a code maintenance surface that a solo developer cannot sustain for 15+ experiments. The existing 1100-line Taskfile already contains the orchestration logic — moving it to Go does not reduce complexity, it relocates it behind a compilation wall.

### Pure Crossplane Compositions

All orchestration handled by composition functions (`function-patch-and-transform`, `function-kcl`).

**Rejected because:** Compositions are static resource DAGs. They cannot orchestrate conditional waits ("wait for cluster ready, then deploy apps") or implement retry/timeout logic. The gateway experiment requires sequential steps with readiness gates.

### Argo Workflows as the Operator

No new CRD — use parameterized `WorkflowTemplate` as the abstraction. Workflow steps handle provisioning, deployment, and validation.

**Rejected because:** Loses the declarative `kubectl get experiments` UX and Kubernetes-native status reporting. Workflows are imperative (run-to-completion), not reconciled. The Crossplane XRD provides the declarative API; Argo Workflows handles the parts that need orchestration (validation).

### Pure Argo Workflows + Argo Events

Argo Events watches for Git changes, triggers WorkflowTemplates that set up experiments.

**Rejected because:** Adds another operator (Argo Events) with sensor/trigger abstractions. Argo Events is not deployed on the hub and would add operational overhead for a problem the Crossplane composition already solves.

## Implementation Phases

### Phase 1 (MVP): GKE Experiment via XRD

- `XExperiment` XRD definition
- GKE composition (cluster + ArgoCD secret + ArgoCD app)
- Gateway-tutorial `claim.yaml`
- Single-experiment concurrency (`destination.name: target`)

### Phase 2: vcluster Composition

- vcluster composition for local experiments (replaces Kind)
- Faster iteration cycle (~30-45s startup vs 8-10min for GKE)
- Taskfile becomes thin wrapper: `kubectl apply/delete` claim

### Phase 3: Validation Workflow Integration

- Argo Workflow as a composed resource (auto-runs on experiment deploy)
- Reusable WorkflowTemplate library (wait-for-apps, connectivity-test, k6-benchmark)
- Status patching: Workflow completion updates `status.validationResult`

### Phase 4: Multi-Experiment Support

- Parameterized `destination.name` via `function-go-templating`
- Concurrent experiments on different clusters
- Per-experiment Grafana datasource auto-provisioning

### Phase 5: TTL + Cost Controls

- Composed Argo Workflow with suspend step implements TTL
- Kyverno policy blocks Experiment creation when budget exceeded
- Cloud cost tagging via composition labels

## Consequences

### Positive

- **Single entry point**: `kubectl apply -f claim.yaml` replaces 8+ manual steps
- **Declarative**: Experiment state in Git, reconciled by Crossplane
- **Reproducible**: Same claim always produces same environment
- **Observable**: `kubectl get experiments` shows status, Grafana shows metrics
- **Extensible**: New providers (AKS, EKS) are new compositions, not new tooling
- **Cleanup is atomic**: `kubectl delete experiment` removes everything
- **No custom code**: Leverages existing Crossplane, ArgoCD, and Argo Workflows
- **Incremental**: Existing experiments and Taskfile workflows continue working

### Negative

- **Composition complexity**: GKE composition will be ~300+ lines (VPC + Subnet + Cluster + NodePool + Secret + Application)
- **Debugging indirection**: Issues may be in XRD, composition function, provider, or ArgoCD
- **Kubeconfig transformation**: YAML-to-JSON conversion via `function-go-templating` is non-trivial to debug
- **Single-experiment limitation**: v1 `destination.name: target` convention prevents concurrent experiments
- **Status reporting gaps**: Crossplane cannot observe ArgoCD Application health or Argo Workflow completion in composition patches

### Trade-offs

| Trade-off | Mitigation |
|-----------|------------|
| Large composition | Split into well-commented sections; follows existing GKE composition pattern |
| Kubeconfig transform debugging | Test with `crossplane render` CLI before deploying |
| Single-experiment concurrency | Acceptable for home lab; Phase 4 adds multi-experiment |
| Status gaps | Use `kubectl get applications` and `argo get @latest` for detailed status |
| Cleanup ordering | Remove ArgoCD finalizer in composition; orphaned resources deleted with cluster |

## References

- [ADR-012: Crossplane Experiment Abstraction](ADR-012-crossplane-experiment-abstraction.md) (predecessor, cluster-only)
- [ADR-013: Crossplane v2 Upgrade](ADR-013-crossplane-v2-upgrade.md) (enables native resource composition)
- [ADR-005: Experiment Lifecycle](ADR-005-experiment-lifecycle.md) (current workflow-based lifecycle)
- [Crossplane Composition Functions](https://docs.crossplane.io/latest/concepts/composition-functions/)
- [Crossplane v2 Native Resource Composition](https://docs.crossplane.io/v2.1/concepts/server-side-compositions/)
- RBAC: `experiments/components/infrastructure/crossplane-providers/core/rbac-native-resources.yaml`
- GKE composition pattern: `experiments/components/infrastructure/crossplane-xrds/definitions/gke-cluster/composition.yaml`
- Gateway observability pipeline: `experiments/components/infrastructure/gateway-observability/`

## Decision Date

2026-02-03
