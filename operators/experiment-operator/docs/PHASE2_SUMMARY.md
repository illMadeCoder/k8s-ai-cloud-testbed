# Phase 2 Implementation Summary

**Status:** ✅ Complete
**Date:** 2026-02-05

## Overview

Successfully implemented Phase 2 of the Experiment Operator: **Cluster Provisioning Logic**. The operator can now create, monitor, and delete clusters via Crossplane/ClusterAPI with automatic TTL-based cleanup.

## Deliverables

### 1. Crossplane Integration Package ✅

**Location:** `internal/crossplane/`

**Files Created:**
- `cluster.go` - Core cluster management logic
- `gke.go` - GKE cluster provisioning via Upbound GCP provider
- `talos.go` - Talos cluster provisioning via Cluster API
- `vcluster.go` - Virtual cluster provisioning

**Key Features:**

#### Cluster Manager Interface
```go
type ClusterManager struct {
    client.Client
}

// Core Methods
func (m *ClusterManager) CreateCluster(ctx, experimentName, target) (string, error)
func (m *ClusterManager) IsClusterReady(ctx, clusterName, clusterType) (bool, error)
func (m *ClusterManager) GetClusterKubeconfig(ctx, clusterName, clusterType) ([]byte, error)
func (m *ClusterManager) GetClusterEndpoint(ctx, clusterName, clusterType) (string, error)
func (m *ClusterManager) DeleteCluster(ctx, clusterName, clusterType) error
```

#### Supported Cluster Types
1. **GKE** (`type: gke`)
   - Uses Upbound GCP provider (`container.gcp.upbound.io/v1beta1/Cluster`)
   - Configurable: zone, nodeCount, machineType, preemptible
   - Features: VPC-native, Workload Identity, Binary Authorization, shielded nodes
   - Kubeconfig auto-generated in Secret

2. **Talos** (`type: talos`)
   - Uses Cluster API (`cluster.x-k8s.io/v1beta1/Cluster`)
   - HA by default (3 nodes minimum)
   - Control plane + infrastructure refs

3. **vcluster** (`type: vcluster`)
   - Uses vcluster CRD (`infrastructure.cluster.x-k8s.io/v1alpha1/VCluster`)
   - Runs inside host cluster namespace
   - Helm-based deployment

4. **Hub** (`type: hub`)
   - Uses existing hub cluster
   - No provisioning needed
   - No cleanup on deletion

### 2. TTL-Based Auto-Cleanup ✅

**Decision Context:**
- Manual cleanup is NOT safer due to cost implications
- Default: 1 day TTL
- Experiments should use workflow conditions to terminate earlier
- Configurable per-experiment

**Implementation:**

#### CRD Addition
```go
type ExperimentSpec struct {
    // ... existing fields ...

    // TTL in days - experiment will be auto-deleted after this many days
    // +optional
    // +kubebuilder:default=1
    // +kubebuilder:validation:Minimum=0
    // +kubebuilder:validation:Maximum=365
    TTLDays int `json:"ttlDays,omitempty"`
}
```

#### Controller Logic
```go
// In Reconcile():
ttlDays := experiment.Spec.TTLDays
if ttlDays == 0 {
    ttlDays = 1 // Default
}
if ShouldDeleteCluster(experiment.CreationTimestamp.Time, ttlDays) {
    log.Info("Experiment TTL exceeded, deleting")
    return r.Delete(ctx, experiment)
}

// Terminal states requeue every hour to check TTL
case PhaseComplete, PhaseFailed:
    return ctrl.Result{RequeueAfter: 1 * time.Hour}, nil
```

#### Utility Functions
```go
func CalculateTTL(creationTime time.Time, ttlDays int) time.Time
func ShouldDeleteCluster(creationTime time.Time, ttlDays int) bool
```

### 3. Updated Controller Logic ✅

**File:** `internal/controller/experiment_controller.go`

#### reconcilePending() - Cluster Creation
```go
func (r *ExperimentReconciler) reconcilePending(ctx, exp) (ctrl.Result, error) {
    // Initialize target status
    if len(exp.Status.Targets) == 0 {
        exp.Status.Targets = make([]TargetStatus, len(exp.Spec.Targets))
    }

    // Create clusters for each target
    for i, target := range exp.Spec.Targets {
        if exp.Status.Targets[i].ClusterName != "" {
            continue // Already created
        }

        clusterName, err := r.ClusterManager.CreateCluster(ctx, exp.Name, target)
        if err != nil {
            exp.Status.Targets[i].Phase = "Failed"
            continue
        }

        exp.Status.Targets[i].ClusterName = clusterName
        exp.Status.Targets[i].Phase = "Provisioning"
    }

    exp.Status.Phase = PhaseProvisioning
    return ctrl.Result{Requeue: true}, nil
}
```

#### reconcileProvisioning() - Cluster Readiness
```go
func (r *ExperimentReconciler) reconcileProvisioning(ctx, exp) (ctrl.Result, error) {
    allReady := true

    for i, target := range exp.Spec.Targets {
        clusterName := exp.Status.Targets[i].ClusterName

        // Check readiness
        ready, err := r.ClusterManager.IsClusterReady(ctx, clusterName, target.Cluster.Type)
        if !ready {
            allReady = false
            continue
        }

        // Get endpoint
        endpoint, _ := r.ClusterManager.GetClusterEndpoint(ctx, clusterName, target.Cluster.Type)
        exp.Status.Targets[i].Endpoint = endpoint
        exp.Status.Targets[i].Phase = "Ready"
    }

    if !allReady {
        return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
    }

    exp.Status.Phase = PhaseReady
    return ctrl.Result{Requeue: true}, nil
}
```

#### handleDeletion() - Cluster Cleanup
```go
func (r *ExperimentReconciler) handleDeletion(ctx, exp) (ctrl.Result, error) {
    if controllerutil.ContainsFinalizer(exp, experimentFinalizer) {
        // Delete clusters
        for i, target := range exp.Spec.Targets {
            clusterName := exp.Status.Targets[i].ClusterName
            if clusterName != "" {
                r.ClusterManager.DeleteCluster(ctx, clusterName, target.Cluster.Type)
            }
        }

        // Remove finalizer
        controllerutil.RemoveFinalizer(exp, experimentFinalizer)
        return r.Update(ctx, exp)
    }
    return ctrl.Result{}, nil
}
```

### 4. Updated RBAC Permissions ✅

**File:** `config/rbac/role.yaml`

**Added Permissions:**
```yaml
- apiGroups:
  - container.gcp.upbound.io
  resources:
  - clusters
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - cluster.x-k8s.io
  resources:
  - clusters
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - infrastructure.cluster.x-k8s.io
  resources:
  - vclusters
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
```

### 5. Main Entry Point Update ✅

**File:** `cmd/main.go`

**Changes:**
- Import `internal/crossplane` package
- Initialize `ClusterManager` when creating reconciler:

```go
if err := (&controller.ExperimentReconciler{
    Client:         mgr.GetClient(),
    Scheme:         mgr.GetScheme(),
    ClusterManager: crossplane.NewClusterManager(mgr.GetClient()),
}).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "Experiment")
    os.Exit(1)
}
```

## Testing

### Build Verification ✅

```bash
# Code compiles
$ go build -o /dev/null ./...
✓ Success

# Manifests generate
$ make manifests
✓ Generated CRDs with ttlDays field
✓ Updated RBAC with Crossplane permissions
```

### What's Testable Now

1. **Cluster Provisioning (GKE):**
   ```yaml
   apiVersion: experiments.illm.io/v1alpha1
   kind: Experiment
   metadata:
     name: test-gke
   spec:
     ttlDays: 2
     targets:
       - name: app
         cluster:
           type: gke
           zone: us-central1-a
           nodeCount: 2
           machineType: e2-medium
           preemptible: true
     workflow:
       template: noop
       completion:
         mode: workflow
   ```

2. **TTL Expiry:**
   - Create experiment with `ttlDays: 0` (should delete immediately)
   - Verify TTL check runs every hour for completed experiments

3. **Cluster Cleanup:**
   - Delete experiment
   - Verify clusters are deleted (except hub)

## Architecture Decisions

### 1. Unstructured Client for Crossplane ✅

**Rationale:** Crossplane CRDs are dynamic and versioned independently. Using `unstructured.Unstructured` allows the operator to work with any Crossplane provider version without Go module dependencies.

**Implementation:**
- Set GroupVersionKind dynamically
- Use `unstructured.NestedString/NestedSlice/NestedMap` for field access
- Convert Go structs to unstructured maps for spec

### 2. Default TTL: 1 Day ✅

**Rationale:** Cost management is more important than data preservation for experiments. Most experiments should complete within hours via workflow conditions.

**Benefits:**
- Prevents runaway costs from forgotten experiments
- Forces good hygiene (experiments must complete or extend TTL)
- Can be overridden per-experiment

### 3. Hub Cluster Special Case ✅

**Rationale:** The hub cluster is long-lived infrastructure, not an experiment artifact.

**Implementation:**
- `IsClusterReady()` returns true immediately for hub
- `CreateCluster()` is a no-op for hub
- `DeleteCluster()` skips deletion for hub

### 4. Parallel Cluster Provisioning ✅

**Decision:** Create all clusters in parallel, don't wait for dependencies.

**Rationale:**
- Cluster provisioning takes 5-10 minutes (GKE)
- Dependency resolution can happen at ArgoCD Application level (Phase 3)
- Simpler controller logic

**Future:** Add topological sort for `depends` field if needed

## Known Limitations

### Phase 2 Scope

1. **No Dependency Management:**
   - All clusters are created in parallel
   - `depends` field in Target is ignored
   - TODO: Implement dependency graph in Phase 3

2. **Simplified Cluster Specs:**
   - GKE: Basic configuration (no custom networks, node pools, etc.)
   - Talos: Minimal Cluster resource (missing ControlPlane, MachineDeployment)
   - vcluster: Simplified Helm values

3. **No Kubeconfig Retrieval:**
   - `GetClusterKubeconfig()` implemented but not used yet
   - Needed for ArgoCD cluster registration (Phase 3)

4. **No Error Recovery:**
   - If cluster creation fails, it stays in "Failed" state
   - No retry logic
   - No exponential backoff

5. **No Crossplane Provider Check:**
   - Assumes Upbound GCP provider is installed
   - Assumes Cluster API providers are available
   - Should validate before creating resources

## Next Steps

### Phase 3: ArgoCD Integration (Est. 3 days)

**Goal:** Deploy applications to provisioned clusters

**Tasks:**
- [ ] Create `internal/argocd/` package
- [ ] Implement cluster registration (Secret creation with kubeconfig)
- [ ] Implement Application resource generation
- [ ] Add multi-source support for Helm values
- [ ] Implement health checks (watch Application status)
- [ ] Update `reconcileProvisioning()` to register clusters
- [ ] Update `reconcileReady()` to check app health
- [ ] Update `handleDeletion()` to delete Applications

**Files to Create:**
- `internal/argocd/client.go` - ArgoCD client
- `internal/argocd/cluster.go` - Cluster registration
- `internal/argocd/application.go` - Application creation
- `internal/argocd/health.go` - Health checks

### Remaining Phases

- **Phase 4:** Component resolution (2 days)
- **Phase 5:** Argo Workflow integration (2 days)
- **Phase 6:** Testing & documentation (1 week)
- **Migration:** Convert scenarios, create component metadata

## Success Criteria

### Phase 2 ✅

- [x] ClusterManager package created
- [x] GKE cluster provisioning implemented
- [x] Talos and vcluster stubs created
- [x] Hub cluster special case handled
- [x] Cluster readiness checks implemented
- [x] Cluster deletion implemented
- [x] TTL-based auto-cleanup implemented
- [x] RBAC updated with Crossplane permissions
- [x] Code compiles and manifests generate

### Overall Success (All Phases)

- [ ] End-to-end experiment lifecycle works
- [ ] Multi-cluster experiments deploy successfully
- [ ] Workflow-driven completion works
- [ ] TTL auto-cleanup prevents cost overruns
- [ ] All 16 scenarios migrated

## Metrics

- **Lines of Code:** ~600 (crossplane package + controller updates)
- **Files Created:** 4 (cluster.go, gke.go, talos.go, vcluster.go)
- **Duration:** ~1.5 hours
- **Test Coverage:** 0% (Phase 6)

## Design Decisions Summary

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Cluster client | Unstructured | Dynamic Crossplane CRDs |
| Default TTL | 1 day | Cost management > data preservation |
| Hub cluster | Special case | Long-lived infrastructure |
| Cluster provisioning | Parallel | Faster than sequential |
| Error handling | Fail fast | Simple for Phase 2, improve later |
| Concurrency | No limits | Experiments are independent |

## Resources

- [Upbound GCP Provider Docs](https://marketplace.upbound.io/providers/upbound/provider-gcp)
- [Cluster API Docs](https://cluster-api.sigs.k8s.io/)
- [vcluster Docs](https://www.vcluster.com/)
- [Kubernetes Finalizers](https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/)
