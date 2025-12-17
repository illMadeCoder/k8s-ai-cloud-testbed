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
├── images/                # Image source code (Dockerfiles + source)
│   ├── _template/         # Template for new images
│   └── <image-name>/
│       ├── Dockerfile     # Container build definition
│       ├── src/           # Application source code
│       └── manifests/     # Kubernetes manifests
│
├── lab/                   # Experiments and workload catalog
│   ├── experiments/       # Experiment definitions
│   └── components/  # Reusable ArgoCD components
│   ├── _template/         # Template for new experiments
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
│           └── test.js
│
├── terraform-modules/     # Reusable Terraform modules
│   └── aks/               # Azure Kubernetes Service
│
├── .github/workflows/     # CI/CD pipelines
│   └── build-images.yaml  # Build and push images to GHCR
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
KUBECONFIG=experiments/scenarios/http-baseline/terraform/prod/kubeconfig-target kubectl get pods

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

## Images

The `images/` directory contains source code for building container images. Each image has a Dockerfile, source code, and Kubernetes manifests.

### Creating an Image

1. **Copy the template:**
   ```bash
   cp -r images/_template images/my-app
   ```

2. **Update the files:**
   - `Dockerfile` - Build instructions for your image
   - `src/` - Your application source code
   - `manifests/deployment.yaml` - Update image path: `ghcr.io/<owner>/illm-k8s-lab/my-app:latest`
   - `manifests/service.yaml` - Update app name
   - `manifests/httproute.yaml` - Update hostname

3. **Create ArgoCD Application** in `argocd-apps/my-app/my-app.yaml`:
   ```yaml
   apiVersion: argoproj.io/v1alpha1
   kind: Application
   metadata:
     name: my-app
     namespace: argocd
   spec:
     project: default
     source:
       repoURL: https://github.com/<owner>/illm-k8s-lab.git
       targetRevision: HEAD
       path: images/my-app/manifests
     destination:
       server: https://kubernetes.default.svc
       namespace: my-app
     syncPolicy:
       automated:
         prune: true
         selfHeal: true
       syncOptions:
         - CreateNamespace=true
   ```

4. **Push to trigger build:**
   ```bash
   git add images/my-app argocd-apps/my-app
   git commit -m "Add my-app image"
   git push
   ```

The GitHub Actions workflow automatically builds and pushes images to GHCR when changes are pushed to `images/<image-name>/`.

### Image Structure

```
images/my-app/
├── Dockerfile              # Multi-stage build (builder + runtime)
├── src/
│   └── main.go             # Application source code
└── manifests/
    ├── deployment.yaml     # Deployment with health probes
    ├── service.yaml        # ClusterIP service
    └── httproute.yaml      # Gateway API HTTPRoute (optional)
```

## Creating a New Experiment

### Using the Template

1. **Copy the template:**
   ```bash
   cp -r experiments/scenarios/_template experiments/scenarios/my-experiment
   ```

2. **Replace placeholders** in all files:
   | Placeholder | Description | Example |
   |-------------|-------------|---------|
   | `EXPERIMENT_NAME` | Experiment identifier | `latency-test` |
   | `APP_NAME` | Application being tested | `hello-app` |
   | `APP_NAMESPACE` | Namespace for the app | `hello-app` |
   | `TARGET_SERVICE` | Service name to test | `hello-app` |

3. **Customize the k6 test** in `k6/k6-scripts.yaml` for your test scenario

4. **Test locally:**
   ```bash
   task exp:deploy:minikube NAME=my-experiment
   kubectl create -f experiments/scenarios/my-experiment/workflow/experiment.yaml
   task exp:undeploy:minikube NAME=my-experiment
   ```

5. **Run full lifecycle:**
   ```bash
   task exp:run:full NAME=my-experiment
   ```

### Template Files

| File | Purpose |
|------|---------|
| `argocd/target.yaml` | ArgoCD Application - deploys app + k6 scripts |
| `k6/k6-scripts.yaml` | ConfigMap with k6 test script |
| `workflow/experiment.yaml` | Argo Workflow - wait → load test → report |

### Manual Setup (Without Template)

1. **Create directory structure:**
   ```bash
   mkdir -p experiments/scenarios/my-experiment/{argocd,terraform/prod,workflow,k6}
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
