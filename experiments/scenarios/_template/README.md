# Experiment Template

Copy this folder to create a new experiment.

## Two Execution Patterns

### Hub Pattern (Simple)
For simple experiments, the workflow runs directly on the hub cluster.
No orchestrator cluster needed.

```
my-experiment/
  target/
    cluster.yaml
    argocd/
  workflow/
    experiment.yaml     # Runs on hub
```

### Orchestrator Pattern (Complex)
For complex experiments requiring isolated experiment management.
Add an `orchestrator/` folder and the workflow runs there instead.

```
my-experiment/
  orchestrator/         # Special: gets ArgoCD + Argo Workflows
    cluster.yaml
  target/
    cluster.yaml
    argocd/
  loadgen/
    cluster.yaml
    argocd/
  workflow/
    experiment.yaml     # Runs on orchestrator, not hub
```

**Benefits of orchestrator pattern:**
- Parallel cluster provisioning (hub creates all clusters at once)
- Isolated experiment state (workflow history separate from hub)
- Closer to production pattern (orchestrator manages experiment lifecycle)

## Structure

```
my-experiment/
  orchestrator/              # Optional: if present, runs the workflow
    cluster.yaml
  target/                    # Each folder = a cluster
    cluster.yaml             # Cluster config (size, provider)
    argocd/                  # ArgoCD apps deployed to this cluster
    crossplane/              # Crossplane claims for this cluster
  loadgen/                   # Optional: add more clusters as needed
    cluster.yaml
    argocd/
    crossplane/
  workflow/
    experiment.yaml          # Argo Workflow (runs on orchestrator or hub)
```

## cluster.yaml

```yaml
size: small | medium | large
provider: azure | aws
# Optional overrides:
# nodes: 3
# vm_size: Standard_D4s_v3
```

## Adding a cluster

Just create a new folder with `cluster.yaml`:

```bash
mkdir -p my-experiment/newcluster/{argocd,crossplane}
echo "size: small" > my-experiment/newcluster/cluster.yaml
```

The hub provisions all clusters in parallel, then bootstraps the orchestrator
(if present), registers clusters with ArgoCD, and runs the workflow.
