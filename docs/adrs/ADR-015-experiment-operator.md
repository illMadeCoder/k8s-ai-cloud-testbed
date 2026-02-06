# ADR-015: Experiment Operator

## Status

Accepted (Revised 2026-02-05)

**Revision note:** Originally proposed a Crossplane XRD approach. Revised to a Go operator (Kubebuilder) after discovering that Crossplane compositions cannot orchestrate multi-target deployments with dependency ordering, conditional workflow submission, or fine-grained status reporting across ArgoCD and Argo Workflows.

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

- **15 experiment scenarios** under `experiments/`
- **Crossplane v2.1.3** with Pipeline mode compositions (`function-patch-and-transform`, `function-go-templating`, `function-auto-ready`)
- **Existing XRDs**: `XGKECluster` (GKE provisioning), `XObjectStorage` (multi-cloud storage)
- **ADR-012** designed `XExperimentCluster` for unified cluster provisioning but did not address app deployment or observability
- **RBAC for native resource composition** already in place (`crossplane-compose-native-resources` ClusterRole grants Crossplane access to Secrets, ArgoCD Applications, Argo Workflows)
- **ArgoCD multi-source pattern** used by all experiments (Helm charts + Git values + Git manifests)
- **Argo Workflows** deployed on hub for validation workflows

### Why the Crossplane XRD Approach Was Abandoned

The original ADR proposed a Crossplane `XExperiment` XRD. During implementation planning, several limitations emerged that made a pure Crossplane approach insufficient:

1. **Multi-target orchestration**: Experiments need multiple targets (e.g., app cluster + loadgen on hub) with dependency ordering. Crossplane compositions produce a flat resource DAG with no sequencing control.
2. **Conditional workflow submission**: Workflows should only run after all ArgoCD apps are healthy. Compositions cannot express "wait for external resource status, then create another resource."
3. **Status aggregation**: The operator needs to aggregate status from Crossplane clusters, ArgoCD Applications, and Argo Workflows into a single `Experiment.status`. Composition status patches cannot observe cross-resource health.
4. **Component resolution**: Resolving `app: nginx` to a specific Git source + Helm parameters requires lookup logic (Component CRDs with fallback). Compositions cannot do dynamic lookups.

## Decision

Use a **Go operator** (Kubebuilder/controller-runtime) with `Experiment` and `Component` CRDs. The operator orchestrates cluster provisioning via Crossplane, app deployment via ArgoCD, and validation via Argo Workflows through a phase-based reconciliation loop.

### Architecture

```
kubectl apply -f experiments/gateway-tutorial.yaml
        |
        v
  Experiment Operator (Go, controller-runtime)
  +----------------------------------------------------------+
  |  Phase: Pending                                          |
  |    Create Crossplane cluster resources per target         |
  |                                                          |
  |  Phase: Provisioning                                     |
  |    Wait for clusters → register with ArgoCD              |
  |    Create ArgoCD Applications (resolved from Components) |
  |                                                          |
  |  Phase: Ready                                            |
  |    Wait for ArgoCD apps to be Healthy+Synced             |
  |    Submit Argo Workflow for validation                    |
  |                                                          |
  |  Phase: Running                                          |
  |    Watch Argo Workflow status                             |
  |    Succeeded → Complete | Failed/Error → Failed          |
  |                                                          |
  |  Phase: Complete/Failed                                  |
  |    TTL check → auto-delete when expired                  |
  +----------------------------------------------------------+
        |              |              |              |
        v              v              v              v
    Crossplane      ArgoCD        Argo           Component
    (clusters)    (applications)  Workflows       CRDs
```

### CRD Schema

**Experiment** (`experiments.illm.io/v1alpha1`): Namespaced resource defining targets, components, and workflow.

**Component** (`experiments.illm.io/v1alpha1`): Cluster-scoped resource defining reusable component sources with Helm config and parameters.

### User Experience

```bash
# Deploy full experiment
kubectl apply -f experiments/gateway-tutorial.yaml

# Watch progress
kubectl get experiments -w
# NAME               PHASE          WORKFLOW   AGE
# gateway-tutorial   Provisioning              10s
# gateway-tutorial   Ready                     9m
# gateway-tutorial   Running        Running    10m
# gateway-tutorial   Complete       Succeeded  25m

# Tear down (operator cleans up clusters, apps, workflows)
kubectl delete experiment gateway-tutorial
```

### Project Structure

```
operators/experiment-operator/
├── api/v1alpha1/
│   ├── experiment_types.go       # Experiment CRD
│   └── component_types.go        # Component CRD (cluster-scoped)
├── internal/
│   ├── controller/               # Reconciliation loop
│   ├── crossplane/               # Cluster provisioning (GKE, Talos, vcluster)
│   ├── argocd/                   # Application management, cluster registration
│   ├── components/               # Component resolution (CR lookup + fallback)
│   └── workflow/                 # Argo Workflow submission and status
├── cmd/main.go
└── config/                       # CRDs, RBAC, deployment manifests
```

---

## Design Decisions

### DD-1: Graceful Degradation vs Fail-Fast on Missing Resources

**Decision:** Graceful degradation with fallbacks.

**Current behavior:**

| Missing resource | Behavior | Rationale |
|---|---|---|
| Component CR not found | Fall back to convention path (`components/{type}/{name}`) | Backward compatibility - experiments work without Component CRs |
| WorkflowTemplate not found | Create inline suspend workflow (manual `argo resume`) | Allows experiments to run without pre-created templates |
| No components resolved for target | Deploy ArgoCD example guestbook app | **Bug - should fail. Fix pending.** |

**Trade-off:** Silent fallbacks risk confusing behavior (user expects Bitnami nginx, gets convention path). But requiring Component CRs for every app would block adoption. The fallback logs a warning.

**Revisit when:** Component CRs exist for all apps and the fallback path is no longer needed. At that point, switch to fail-fast.

### DD-2: Deletion Strategy (Best-Effort vs Blocking)

**Decision:** Best-effort deletion with finalizer removal.

**Current behavior:** `handleDeletion` attempts to delete ArgoCD apps, unregister clusters, delete Crossplane resources, and delete Argo Workflows. If any step fails, it logs the error and continues. The finalizer is always removed.

**Trade-off:** Orphaned cloud resources may incur cost. But blocking deletion on a failed provider (e.g., GCP API down) means stuck experiments that can't be cleaned up at all. Since TTL auto-deletes experiments, the window for orphaned resources is bounded.

**Revisit when:** Cost monitoring is in place. Could add a hybrid approach: retry N times, then remove finalizer and emit a Kubernetes Event or create a `bd` issue for manual follow-up.

### DD-3: One ArgoCD Application per Target vs per Component

**Decision:** One Application per target (all components as multi-source).

**Current behavior:** All components for a target are bundled into a single ArgoCD Application with multiple sources. If one component fails, the entire target shows as unhealthy.

**Trade-off:** Simpler to manage (fewer objects, single health check). But loses granularity - can't independently sync, rollback, or debug individual components.

**Revisit when:** Experiments have 5+ components per target and per-component visibility becomes necessary.

### DD-4: Cluster Naming and Uniqueness

**Decision:** Deterministic names (`{experiment}-{target}`), no random suffix.

**Current behavior:** Cluster names are `{experiment.Name}-{target.Name}`. Two experiments with the same name in different namespaces would collide.

**Trade-off:** Predictable names make debugging easier (`kubectl get cluster gateway-tutorial-app`). Collision risk is low since experiment names are unique within their namespace and the lab runs few concurrent experiments.

**Revisit when:** Multi-namespace experiments are needed. Add namespace prefix or UUID suffix at that point.

### DD-5: Status Polling vs Watches

**Decision:** Polling with fixed intervals.

**Current behavior:**
- Cluster readiness: poll every 10s
- ArgoCD app health: poll every 15s
- Argo Workflow status: poll every 15s
- TTL check for terminal experiments: every 1h

**Trade-off:** Polling is simple and has no setup complexity. Watches would react faster and use fewer API calls, but require setting up secondary resource watches in `SetupWithManager` for unstructured types (Crossplane, ArgoCD, Argo Workflow CRDs) which adds complexity.

**Revisit when:** Running 10+ concurrent experiments. At that point, polling generates significant API load and watches become worth the complexity.

### DD-6: Unstructured Clients vs Typed API Clients

**Decision:** Use `unstructured.Unstructured` for all external resources.

**Current behavior:** Crossplane clusters, ArgoCD Applications, and Argo Workflows are all managed via `unstructured.Unstructured` with string-based field paths. No Go dependency on their type packages.

**Trade-off:** No compile-time type safety (typos in field paths are runtime errors). But avoids version coupling - the operator works regardless of which ArgoCD or Crossplane version is installed, and the go.mod stays lean.

**Revisit when:** A field path typo causes a production incident, or when the external APIs stabilize enough that version coupling is acceptable.

### DD-7: TTL Scope (Experiment-Level)

**Decision:** Single TTL applies to the entire experiment, not per-cluster.

**Current behavior:** `spec.ttlDays` (default: 1, max: 365) triggers auto-deletion of the entire experiment after the TTL expires. All clusters, apps, and workflows are cleaned up together.

**Trade-off:** Simple and aligned with cost control intent. But doesn't allow "keep the app cluster, delete the loadgen" scenarios.

**Revisit when:** Experiments need heterogeneous lifetimes across targets.

### DD-8: Hub Cluster Convention

**Decision:** `cluster.type: hub` means "use the cluster the operator runs on."

**Current behavior:** Hub targets skip cluster provisioning and deletion. The ArgoCD server URL is set to `https://kubernetes.default.svc`. No kubeconfig is needed.

**Trade-off:** Simple and correct for a single-hub topology. Would break if the operator runs on a different cluster than the hub.

**Revisit when:** Multi-hub or remote-hub topologies are needed.

### DD-9: Workflow Namespace

**Decision:** All Argo Workflows are created in the `argo` namespace.

**Current behavior:** Hardcoded to `argo` (the default Argo Workflows namespace). Workflows are named `{experiment}-validation`.

**Trade-off:** Simple, single namespace. But workflows from different experiments could interfere. Labels (`experiments.illm.io/experiment`) provide isolation at the query level.

**Revisit when:** RBAC isolation between experiments is needed, or when workflow naming collisions occur.

### DD-10: Target Dependency Ordering

**Decision:** Dependencies declared but not yet enforced.

**Current behavior:** Targets have a `depends` field but all clusters are created in parallel regardless. The dependency graph is not traversed.

**Trade-off:** Simplifies the initial implementation. Most experiments have independent targets or natural ordering through cluster readiness timing. But an experiment with `loadgen.depends: [app]` could submit the loadgen workflow before the app is healthy.

**Revisit when:** Implementing Phase 6 (tests). This is the highest-priority enhancement.

---

## Known Issues

These are implementation bugs, not design decisions:

1. **Guestbook placeholder**: When no components resolve for a target, deploys the ArgoCD example guestbook app instead of failing. Should return an error.
2. **GKE PROJECT_ID**: Binary authorization config references `PROJECT_ID` as a literal string, never substituted.
3. **Talos provisioning incomplete**: Only creates the Cluster resource, not TalosControlPlane or worker resources. Talos clusters will not actually provision.
4. **Hardcoded Development logging**: `zap.Options{Development: true}` is hardcoded in `cmd/main.go`.
5. **Helm parameter ordering**: Iterating over `map[string]string` produces non-deterministic parameter order in ArgoCD Applications.

---

## Alternatives Considered

### Crossplane XExperiment XRD (Original ADR-015 Proposal)

Single Crossplane XRD with composition pipeline: `function-patch-and-transform` + `function-go-templating` + `function-auto-ready`.

**Rejected because:** Compositions are static resource DAGs. Cannot orchestrate conditional waits ("wait for cluster ready, then deploy apps"), multi-target dependency ordering, or status aggregation across ArgoCD and Argo Workflows. The gateway experiment requires sequential steps with readiness gates that compositions cannot express.

**What it was good for:** Single-cluster, single-app experiments with no validation workflow. The `XGKECluster` XRD remains useful for cluster-only provisioning.

### Pure Crossplane Compositions

All orchestration handled by composition functions (`function-patch-and-transform`, `function-kcl`).

**Rejected for same reasons as above.** Additionally, `function-kcl` adds a KCL language dependency that increases cognitive load.

### Argo Workflows as the Operator

No new CRD - use parameterized `WorkflowTemplate` as the abstraction.

**Rejected because:** Loses the declarative `kubectl get experiments` UX and Kubernetes-native status reporting. Workflows are imperative (run-to-completion), not reconciled.

### Pure Argo Workflows + Argo Events

Argo Events watches for Git changes, triggers WorkflowTemplates.

**Rejected because:** Adds another operator (Argo Events) with sensor/trigger abstractions not deployed on the hub.

## Implementation Phases

### Phase 1: Operator Scaffolding - Complete

- Kubebuilder project at `operators/experiment-operator/`
- Experiment and Component CRD definitions
- Basic controller with phase transitions and finalizer pattern
- RBAC manifests

### Phase 2: Cluster Provisioning - Complete

- Crossplane integration (GKE, Talos, vcluster, hub)
- Cluster readiness checks, endpoint/kubeconfig retrieval
- TTL-based auto-deletion

### Phase 3: ArgoCD Integration - Complete

- Cluster registration (Secret creation)
- Application creation with multi-source support
- Health checks (Healthy + Synced)

### Phase 4: Component Resolution - Complete

- Component CRD (cluster-scoped) with sources, parameters, observability metadata
- Resolver: Component CR lookup with convention-based fallback
- Parameter merging (ComponentRef.params > CR defaults)

### Phase 5: Argo Workflow Integration - Complete

- Workflow submission from WorkflowTemplate references
- Inline fallback workflow (suspend step) when template missing
- Status watching (Pending/Running/Succeeded/Failed/Error)
- Workflow cleanup on experiment deletion

### Phase 6: Testing & Documentation - Pending

- Unit tests for controller logic and component resolution
- Integration tests for full experiment lifecycle
- Fix known issues listed above

### Migration - Pending

- Create Component CRs for existing apps
- Convert scenarios to Experiment CRs
- Deprecate Taskfile orchestration

## Consequences

### Positive

- **Single entry point**: `kubectl apply -f experiment.yaml` replaces 8+ manual steps
- **Declarative**: Experiment state in Git, reconciled by controller
- **Multi-target**: Supports experiments across multiple clusters with dependency ordering
- **Component reuse**: Component CRDs define apps once, reference many times
- **Observable**: `kubectl get experiments` shows phase, workflow status, age
- **Cost-safe**: TTL auto-deletion prevents forgotten clusters (default: 1 day)
- **Cleanup is atomic**: `kubectl delete experiment` removes clusters, apps, and workflows

### Negative

- **Go maintenance surface**: Operator requires Go compilation, container builds, deployment
- **Debugging indirection**: Issues may be in operator, Crossplane, ArgoCD, or Argo Workflows
- **No type safety**: Unstructured clients mean field path typos are runtime errors
- **Polling overhead**: Fixed-interval status checks could become expensive at scale

### Trade-offs

| Trade-off | Current default | Mitigation |
|-----------|----------------|------------|
| Fallback vs fail-fast | Graceful degradation | Log warnings; switch to fail-fast after migration |
| Deletion: best-effort vs blocking | Best-effort | TTL bounds orphan window; add retry+alerting later |
| One app per target vs per component | Per target | Switch to per-component when granularity needed |
| Polling vs watches | Polling (10-15s) | Switch to watches at 10+ concurrent experiments |
| Unstructured vs typed clients | Unstructured | Switch to typed if runtime errors become a problem |

## References

- [ADR-012: Crossplane Experiment Abstraction](ADR-012-crossplane-experiment-abstraction.md) (predecessor, cluster-only)
- [ADR-013: Crossplane v2 Upgrade](ADR-013-crossplane-v2-upgrade.md) (enables native resource composition)
- [ADR-005: Experiment Lifecycle](ADR-005-experiment-lifecycle.md) (current workflow-based lifecycle)
- Operator source: `operators/experiment-operator/`
- Example Component CR: `components/apps/nginx/component.yaml`

## Decision Date

2026-02-03 (original), 2026-02-05 (revised to Go operator)
