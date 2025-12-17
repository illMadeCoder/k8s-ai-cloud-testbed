# Experiment Template

Copy this folder to create a new experiment.

## Structure

```
my-experiment/
  target/                    # Each folder = a cluster
    cluster.yaml             # Cluster config (size, provider)
    argocd/                  # ArgoCD apps deployed to this cluster
    crossplane/              # Crossplane claims for this cluster
  loadgen/                   # Optional: add more clusters as needed
    cluster.yaml
    argocd/
    crossplane/
  workflow/
    experiment.yaml          # Argo Workflow (runs on orchestrator)
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

GitLab CI will discover and provision it via Terraform.
