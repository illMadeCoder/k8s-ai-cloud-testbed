# Experiment Operator

A Kubernetes operator for orchestrating multi-cluster experiment deployments with ArgoCD and Argo Workflows.

## Overview

The Experiment Operator automates the lifecycle of cloud-native experiments:
- Provisions clusters via Crossplane (GKE, Talos, vcluster)
- Deploys applications via ArgoCD with dependency management
- Runs validation and load generation via Argo Workflows
- Provides workflow-driven lifecycle management

## Quick Start

### Prerequisites
- Kubernetes cluster with kubectl access
- ArgoCD installed
- Argo Workflows installed
- Crossplane installed (for cluster provisioning)

### Installation

```bash
# Install CRDs
make install

# Deploy the operator
make deploy IMG=ghcr.io/illmadecoder/experiment-operator:latest

# Verify deployment
kubectl get pods -n experiment-operator-system
```

### Create an Experiment

```bash
kubectl create namespace experiments
kubectl apply -f config/samples/experiments_v1alpha1_experiment.yaml
```

### Monitor Progress

```bash
# Watch experiment status
kubectl get experiments -n experiments -w

# View detailed status
kubectl describe experiment gateway-tutorial -n experiments

# Check operator logs
kubectl logs -n experiment-operator-system deployment/experiment-operator-controller-manager -f
```

## Development

### Build

```bash
# Build manager binary
make build

# Build and push container image
make docker-build docker-push IMG=ghcr.io/illmadecoder/experiment-operator:latest
```

### Testing

```bash
# Run tests
make test

# Install CRDs in test cluster
make install

# Run locally (requires kubeconfig)
make run
```

### Generate Manifests

```bash
# After modifying types or markers
make manifests generate
```

## Architecture

See the implementation plan in `/docs/experiment-operator-implementation-plan.md` for detailed architecture.

### Phase Status

- ✅ **Phase 1**: Operator scaffolding and basic controller
- 🚧 **Phase 2**: Cluster provisioning (in progress)
- ⏳ **Phase 3**: ArgoCD integration
- ⏳ **Phase 4**: Component resolution
- ⏳ **Phase 5**: Argo Workflow integration
- ⏳ **Phase 6**: Testing and documentation

## License

Apache 2.0
