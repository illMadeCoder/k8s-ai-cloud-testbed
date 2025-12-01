# illm-k8s-lab

A GitOps-based Kubernetes experimentation platform for running reproducible infrastructure experiments with automated lifecycle management.

## Overview

This lab provides a framework for:
- **Defining experiments** as code (infrastructure + applications + load tests)
- **Running locally** on minikube for development
- **Deploying to production** on Azure AKS with multi-cluster support
- **Automated lifecycle**: deploy → run → collect results → destroy

## Quick Start

### Prerequisites

- [minikube](https://minikube.sigs.k8s.io/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Task](https://taskfile.dev/) (task runner)
- [Helm](https://helm.sh/)
- [Argo CLI](https://argoproj.github.io/argo-workflows/cli/) (for workflows)
- [Terraform](https://terraform.io/) (for prod deployments)
- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/) (for AKS)

### Local Setup (Minikube)

```bash
# Start minikube
minikube start --memory=8192 --cpus=4

# Bootstrap ArgoCD
task bootstrap

# Deploy core infrastructure (k6, observability)
task deploy:core

# Run an experiment
task exp:run:full NAME=http-baseline
```

### Production Setup (Azure AKS)

```bash
# Login to Azure
az login

# Run full experiment (creates infra → runs test → destroys infra)
task exp:run:full:prod NAME=http-baseline USERS=100 DURATION=300s
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Experiment                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────────┐     ┌──────────────┐     ┌──────────────┐   │
│   │   Terraform  │────▶│   ArgoCD     │────▶│    Argo      │   │
│   │   (infra)    │     │   (apps)     │     │   Workflow   │   │
│   └──────────────┘     └──────────────┘     └──────────────┘   │
│          │                    │                    │            │
│          ▼                    ▼                    ▼            │
│   ┌──────────────┐     ┌──────────────┐     ┌──────────────┐   │
│   │ AKS Clusters │     │  Demo App    │     │   k6 Load    │   │
│   │ (target,     │     │  Monitoring  │     │    Test      │   │
│   │  loadgen)    │     │  k6 Scripts  │     │              │   │
│   └──────────────┘     └──────────────┘     └──────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Project Structure

```
illm-k8s-lab/
├── argocd-apps/           # Shared ArgoCD applications
│   ├── demo-app/          # Sample application
│   ├── k6/                # Load testing infrastructure
│   └── observability/     # Prometheus, Grafana, Loki
│
├── experiments/           # Experiment definitions
│   └── <experiment-name>/
│       ├── argocd/              # ArgoCD applications (one per cluster)
│       │   ├── target.yaml      # Apps for "target" cluster
│       │   └── loadgen.yaml     # Apps for "loadgen" cluster
│       ├── terraform/           # Infrastructure per environment
│       │   └── prod/
│       │       ├── main.tf
│       │       ├── variables.tf
│       │       └── outputs.tf
│       ├── workflow/            # Argo Workflow definitions
│       │   └── experiment.yaml  # Main experiment workflow
│       └── k6/                  # Load test scripts
│           └── baseline.js
│
├── terraform-modules/     # Reusable Terraform modules
│   └── aks/               # Azure Kubernetes Service
│
└── Taskfile.yaml          # Task runner commands
```

## Conventions

### ArgoCD Files → Cluster Matching

ArgoCD application files are named after the cluster they deploy to:

| File | Deploys To |
|------|------------|
| `argocd/target.yaml` | Cluster named "target" |
| `argocd/loadgen.yaml` | Cluster named "loadgen" |
| `argocd/worker.yaml` | Cluster named "worker" |

This convention enables automatic deployment - the task iterates through clusters and finds the matching ArgoCD file.

### Terraform Cluster Definitions

Clusters are defined in `terraform/prod/main.tf`:

```hcl
variable "clusters" {
  default = {
    target = {
      vm_size    = "Standard_D4s_v3"
      node_count = 3
      min_nodes  = 2
      max_nodes  = 10
    }
    loadgen = {
      vm_size    = "Standard_D2s_v3"
      node_count = 2
      min_nodes  = 1
      max_nodes  = 5
    }
  }
}
```

Cluster names here **must match** ArgoCD filenames.

### Core vs Experiment Infrastructure

**Core infrastructure** (deployed once, persists):
- k6 namespace and scripts
- Observability stack (Prometheus, Grafana, Loki)
- Gateway API and Cert Manager
- ArgoCD itself

**Experiment infrastructure** (per-experiment):
- Target application being tested
- Experiment-specific configurations

Deploy core infrastructure with:

```bash
task deploy:core
```

## Experiment Lifecycle

### Full Automated Run

A single command handles the entire lifecycle:

```bash
task exp:run:full:prod NAME=http-baseline USERS=50 DURATION=120s
```

This executes:

1. **Deploy Infrastructure** (`terraform apply`)
   - Creates AKS clusters defined in `terraform/prod/`
   - Writes kubeconfig files for each cluster

2. **Deploy Applications** (ArgoCD)
   - Applies `argocd/{cluster}.yaml` to each cluster
   - Waits for sync and healthy status

3. **Run Experiment** (Argo Workflow)
   - Submits workflow to loadgen cluster
   - Workflow waits for apps to be ready
   - Runs k6 load test against target cluster
   - Collects and stores results

4. **Destroy Infrastructure** (`terraform destroy`)
   - Automatically tears down all clusters
   - No lingering cloud resources

### Manual Steps

For more control, run steps individually:

```bash
# Deploy only
task exp:deploy:prod NAME=http-baseline

# View kubeconfigs
task exp:kubeconfig:prod NAME=http-baseline

# Connect to specific cluster
KUBECONFIG=experiments/http-baseline/terraform/prod/kubeconfig-target kubectl get pods

# Run test manually
task exp:run:prod NAME=http-baseline TARGET=http://<ip> USERS=10

# Destroy when done
task exp:undeploy:prod NAME=http-baseline
```

### Local Development (Minikube)

Deploys **all** ArgoCD files to a single cluster:

```bash
# Prerequisites
task deploy:core                              # Deploy core infra (k6, observability)

# Deploy experiment
task exp:deploy:minikube NAME=http-baseline   # Deploys all argocd/*.yaml

# Run load test
task exp:run USERS=10 DURATION=60s

# Teardown
task exp:undeploy:minikube NAME=http-baseline
```

## Creating a New Experiment

1. **Create directory structure:**
   ```bash
   mkdir -p experiments/my-experiment/{argocd,terraform/prod,workflow,k6}
   ```

2. **Define clusters** in `terraform/prod/main.tf`:
   ```hcl
   variable "clusters" {
     default = {
       server = { ... }  # Your cluster names
       client = { ... }
     }
   }
   ```

3. **Create matching ArgoCD apps:**
   ```bash
   # argocd/server.yaml - apps for server cluster
   # argocd/client.yaml - apps for client cluster
   ```

4. **Create workflow** in `workflow/experiment.yaml`

5. **Add load test scripts** in `k6/`

6. **Test locally first:**
   ```bash
   task exp:deploy:minikube NAME=my-experiment
   task exp:run
   ```

7. **Run in production:**
   ```bash
   task exp:run:full:prod NAME=my-experiment
   ```

## Available Tasks

| Task | Description |
|------|-------------|
| **Bootstrap** | |
| `task bootstrap` | Install ArgoCD on cluster |
| `task deploy:core` | Deploy core infrastructure |
| **Local (Minikube)** | |
| `task exp:deploy:minikube NAME=x` | Deploy experiment apps (all argocd files) |
| `task exp:run` | Run k6 load test |
| `task exp:run:full NAME=x` | Full lifecycle (deploy→run→undeploy) |
| `task exp:undeploy:minikube NAME=x` | Remove experiment |
| **Production (AKS)** | |
| `task exp:deploy:prod NAME=x` | Create AKS clusters + deploy apps |
| `task exp:plan:prod NAME=x` | Preview terraform changes |
| `task exp:run:prod NAME=x TARGET=url` | Run load test on prod |
| `task exp:run:full:prod NAME=x` | Full lifecycle with auto-destroy |
| `task exp:undeploy:prod NAME=x` | Destroy AKS clusters |
| `task exp:kubeconfig:prod NAME=x` | List kubeconfig files |
| **Utilities** | |
| `task exp:list` | List available experiments |
| `task exp:status` | Show k6 pod status |
| `task exp:clean` | Clean up k6 pods |
| `task status` | Show ArgoCD application status |
| `task argocd:ui` | Port-forward ArgoCD UI |

## License

MIT
