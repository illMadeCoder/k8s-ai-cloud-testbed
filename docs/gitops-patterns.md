# GitOps Patterns in illm-k8s-lab

This document captures the ArgoCD/GitOps patterns used in this project for deploying and managing Kubernetes workloads.

## Overview

The project uses a layered GitOps approach:

| Layer | Tool | Responsibility |
|-------|------|----------------|
| Cloud Infrastructure | GitLab CI + Terraform | VNets, subnets, AKS/EKS clusters, IAM |
| Kubernetes Platform | ArgoCD | Core platform components, observability, infrastructure |
| Applications | ArgoCD | Experiment workloads, demo apps |
| Workflows | Argo Workflows | Experiment orchestration, load testing |

ArgoCD serves as the GitOps engine for all Kubernetes resources, while GitLab CI manages cloud infrastructure provisioning via Terraform (see [ADR-001](adrs/ADR-001-gitlab-ci-for-iac.md)).

---

## App-of-Apps Pattern

### Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         ArgoCD (orchestrator cluster)                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │              core-infrastructure (App-of-Apps)                   │   │
│  │              platform/apps/core-infrastructure.yaml │   │
│  └───────────────────────────┬─────────────────────────────────────┘   │
│                              │                                          │
│       ┌──────────────────────┼──────────────────────┐                  │
│       │                      │                      │                  │
│       ▼                      ▼                      ▼                  │
│  ┌─────────┐           ┌──────────┐          ┌───────────┐            │
│  │  Core   │           │Observ-   │          │Infra-     │            │
│  │         │           │ability   │          │structure  │            │
│  ├─────────┤           ├──────────┤          ├───────────┤            │
│  │cert-mgr │           │prometheus│          │crossplane │            │
│  │gateway  │           │otel      │          │vault      │            │
│  │argocd   │           │minio     │          │keda       │            │
│  └─────────┘           └──────────┘          └───────────┘            │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    Stack Applications                            │   │
│  │                    platform/apps/stack-*            │   │
│  ├─────────────────────────────┬───────────────────────────────────┤   │
│  │       stack-elk             │          stack-loki               │   │
│  │  (eck-operator + elk-stack) │    (loki + promtail)              │   │
│  └─────────────────────────────┴───────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                  Experiment Applications                         │   │
│  │                  experiments/{name}/{cluster}/argocd/        │   │
│  ├─────────────────┬─────────────────┬─────────────────────────────┤   │
│  │  http-baseline  │   hello-app     │    multi-cloud-demo         │   │
│  │  (target/loadgen)│   (target)     │    (target + crossplane)    │   │
│  └─────────────────┴─────────────────┴─────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Root Application

**File:** `platform/apps/core-infrastructure.yaml`

The root Application uses multi-source configuration to selectively deploy platform components:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: core-infrastructure
  namespace: argocd
  labels:
    layer: core
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  sources:
    # Each source deploys a specific component category
    - repoURL: https://github.com/illMadeCoder/illm-k8s-lab.git
      targetRevision: HEAD
      path: lab/components/core/cert-manager
      directory:
        recurse: false
        include: 'cert-manager*.yaml'
    # ... additional sources for gateway, observability, etc.
  destination:
    server: https://kubernetes.default.svc
    namespace: argocd
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

### Stack Applications

Stacks group related components that are deployed together:

**File:** `platform/apps/stack-elk.yaml`
- Deploys ECK Operator + ELK Stack as a unit
- Used when full Elasticsearch logging is needed

**File:** `platform/apps/stack-loki.yaml`
- Deploys Loki + Promtail for lightweight logging
- Alternative to ELK for simpler use cases

---

## Multi-Source Configuration

### Helm Chart + Git Values Pattern

Combine external Helm charts with repository-managed values using the `$values` reference:

```yaml
spec:
  sources:
    # Source 1: External Helm chart
    - repoURL: https://charts.jetstack.io
      chart: cert-manager
      targetRevision: v1.19.1
      helm:
        releaseName: cert-manager
        valueFiles:
          - $values/lab/components/core/cert-manager/values.yaml

    # Source 2: Git repository (creates $values reference)
    - repoURL: https://github.com/illMadeCoder/illm-k8s-lab.git
      targetRevision: HEAD
      ref: values
```

**How it works:**
1. `ref: values` creates a named reference to the Git repository
2. `$values/path/to/file.yaml` resolves to that path in the referenced repo
3. Enables separation of chart definition from deployment values

**Examples:**
- `lab/components/core/cert-manager/cert-manager.yaml`
- `lab/components/infrastructure/vault/vault.yaml`
- `lab/components/observability/prometheus-stack/kube-prometheus-stack.yaml`

### Directory-Based Selective Sync

Control which files are synced from a directory:

**Include Pattern** (only sync matching files):
```yaml
directory:
  recurse: false
  include: 'cert-manager*.yaml'
```

**Exclude Pattern** (sync all except matching):
```yaml
directory:
  exclude: 'demo-app.yaml'
```

**Multiple Patterns** (brace expansion):
```yaml
directory:
  include: '{*-httproute.yaml,*-rbac.yaml}'
```

**Recursive Sync** (include subdirectories):
```yaml
directory:
  recurse: true
```

---

## Sync Strategies

### Sync Waves (Deployment Ordering)

Sync waves control the order in which resources are deployed. Lower numbers deploy first.

```yaml
metadata:
  annotations:
    argocd.argoproj.io/sync-wave: "0"
```

**Wave ordering used in this project:**

| Wave | Purpose | Examples |
|------|---------|----------|
| 0 | Initial setup, CRDs | Crossplane providers |
| 1 | Dependent on wave 0 | Provider configurations |
| 2 | Requires providers healthy | Crossplane XRDs |
| 3 | Final configuration | Environment configs |

**Example: Crossplane deployment chain**

```
Wave 0: provider-aws.yaml, provider-azure.yaml
    ↓
Wave 1: provider-config/*.yaml
    ↓
Wave 2: xrds.yaml (CompositeResourceDefinitions)
    ↓
Wave 3: environment-config.yaml
```

### Retry Policies

Configure exponential backoff for applications with external dependencies:

```yaml
syncPolicy:
  retry:
    limit: 5
    backoff:
      duration: 5s      # Initial retry delay
      factor: 2         # Multiply delay each retry
      maxDuration: 3m   # Maximum delay cap
```

**Standard configuration:** 5 retries, 5s→3m

**Extended configuration (complex components):** 10 retries, 10s→5m
- Used for: Crossplane XRDs, Crossplane Providers

### ignoreDifferences

Prevent false out-of-sync states caused by server-side modifications:

**For CRDs:**
```yaml
ignoreDifferences:
  - group: apiextensions.k8s.io
    kind: CustomResourceDefinition
    jqPathExpressions:
      - .metadata.annotations
      - .metadata.managedFields
      - .spec.validation
```

**For Webhooks:**
```yaml
ignoreDifferences:
  - group: admissionregistration.k8s.io
    kind: MutatingWebhookConfiguration
    jqPathExpressions:
      - .metadata.managedFields
      - .webhooks[].clientConfig.caBundle
  - group: admissionregistration.k8s.io
    kind: ValidatingWebhookConfiguration
    jqPathExpressions:
      - .metadata.managedFields
      - .webhooks[].clientConfig.caBundle
```

**For Crossplane resources:**
```yaml
ignoreDifferences:
  - group: apiextensions.crossplane.io
    kind: CompositeResourceDefinition
    jqPathExpressions:
      - .status
  - group: pkg.crossplane.io
    kind: Provider
    jqPathExpressions:
      - .status
```

### ServerSideApply

Use server-side apply for better handling of CRDs and webhooks:

```yaml
syncOptions:
  - ServerSideApply=true
```

**Components using ServerSideApply:**
- Vault, Crossplane, KEDA
- MinIO, ELK Stack, ECK Operator
- K6 Operator, RabbitMQ, Strimzi
- Chaos Mesh

### RespectIgnoreDifferences

Enable the `ignoreDifferences` configuration:

```yaml
syncOptions:
  - RespectIgnoreDifferences=true
```

Required alongside `ignoreDifferences` configuration.

---

## Experiment GitOps Pattern

### Directory Structure

Each experiment follows a consistent layout:

```
experiments/{experiment-name}/
├── target/                         # Target cluster (app under test)
│   ├── argocd/
│   │   └── app.yaml               # ArgoCD Application
│   ├── crossplane/                # Optional: cloud resources
│   │   └── claims.yaml
│   └── k6/                        # Optional: load test scripts
├── loadgen/                       # Load generator cluster
│   ├── argocd/
│   │   ├── app.yaml
│   │   └── k6-scripts.yaml        # ConfigMap with k6 scripts
│   └── cluster.yaml
└── workflow/
    └── experiment.yaml            # Argo Workflow definition
```

### Label Conventions

All experiment Applications use consistent labels:

```yaml
metadata:
  labels:
    experiment: http-baseline      # Experiment name
    cluster: target                # target | loadgen | orchestrator
```

### Multi-Cluster Targeting

Experiments deploy to registered clusters using `destination.name`:

```yaml
spec:
  destination:
    name: target                   # Cluster registered in ArgoCD
    namespace: demo
```

For orchestrator-local resources, use the in-cluster server:

```yaml
spec:
  destination:
    server: https://kubernetes.default.svc
    namespace: argo-workflows
```

### Example: http-baseline

**Target Application** (`experiments/http-baseline/target/argocd/app.yaml`):
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: http-baseline-target
  namespace: argocd
  labels:
    experiment: http-baseline
    cluster: target
spec:
  project: default
  sources:
    - repoURL: https://github.com/illMadeCoder/illm-k8s-lab.git
      targetRevision: HEAD
      path: lab/components/apps/demo-app/k8s
      directory:
        recurse: false
        include: 'deployment.yaml'
    - repoURL: https://github.com/illMadeCoder/illm-k8s-lab.git
      targetRevision: HEAD
      path: experiments/http-baseline/target/argocd
      directory:
        recurse: false
        include: 'nodeport-service.yaml'
  destination:
    name: target
    namespace: demo
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

---

## Component Organization

### Directory Structure

```
components/
├── apps/              # Business applications
│   ├── demo-app/
│   └── hello-app/
├── chaos/             # Chaos engineering
│   └── chaos-mesh/
├── core/              # Platform foundations
│   ├── argocd/
│   ├── cert-manager/
│   └── gateway-api/
├── infrastructure/    # Cloud/resource management
│   ├── crossplane/
│   ├── crossplane-providers/
│   ├── crossplane-xrds/
│   ├── keda/
│   └── vault/
├── messaging/         # Event/message brokers
│   ├── kafka/
│   ├── rabbitmq/
│   └── strimzi-operator/
├── observability/     # Monitoring & logging
│   ├── eck-operator/
│   ├── elk-stack/
│   ├── loki/
│   ├── otel-collector/
│   ├── prometheus-stack/
│   └── promtail/
├── storage/           # Data persistence
│   └── minio/
├── testing/           # Load testing
│   └── k6/
└── workflows/         # Orchestration
    └── argo-workflows/
```

### Component File Organization

Each component follows this pattern:

```
{component}/
├── {component}.yaml       # ArgoCD Application
├── values.yaml            # Helm values (if Helm-based)
├── k8s/                   # Raw Kubernetes manifests
│   ├── deployment.yaml
│   ├── service.yaml
│   └── httproute.yaml
└── {additional}.yaml      # Component-specific resources
    # e.g., issuers.yaml, rbac.yaml, httproute.yaml
```

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Application name | `{component}` or `{category}-{component}` | `cert-manager`, `kube-prometheus-stack` |
| Application file | `{component}.yaml` | `cert-manager.yaml` |
| Values file | `values.yaml` | `values.yaml` |
| K8s manifests dir | `k8s/` | `k8s/deployment.yaml` |
| Additional resources | `{component}-{type}.yaml` | `kibana-httproute.yaml` |

---

## Pattern Reference

### Sync Options by Component

| Component | ServerSideApply | RespectIgnoreDifferences | ignoreDifferences | Retry |
|-----------|-----------------|--------------------------|-------------------|-------|
| cert-manager | - | - | - | Standard |
| gateway-api | - | - | - | Standard |
| prometheus-stack | - | Yes | CRDs, Webhooks | Standard |
| crossplane | Yes | Yes | CRDs, Webhooks | Extended |
| crossplane-providers | - | Yes | Provider status | Extended |
| crossplane-xrds | - | Yes | XRD/Composition status | Extended |
| vault | Yes | Yes | StatefulSet VCTs | Standard |
| minio | Yes | - | - | Standard |
| elk-stack | Yes | - | - | Standard |
| eck-operator | Yes | - | - | Standard |
| k6-operator | Yes | - | - | Standard |
| rabbitmq | Yes | - | - | Standard |
| strimzi | Yes | - | - | Standard |
| chaos-mesh | Yes | - | - | Standard |

### When to Use Each Pattern

| Pattern | Use When |
|---------|----------|
| Multi-source (Helm + Git) | External Helm chart with custom values |
| Directory include | Only specific files from a directory |
| Directory exclude | Most files except a few |
| Sync waves | Dependencies between resources |
| Retry policies | External dependencies, CRD installation |
| ignoreDifferences | Server modifies resources (CRDs, webhooks, status) |
| ServerSideApply | Complex CRDs, webhook configurations |
| RespectIgnoreDifferences | Using ignoreDifferences config |

---

## Related Documentation

- [ADR-001: GitLab CI for IaC Orchestration](adrs/ADR-001-gitlab-ci-for-iac.md)
- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [ArgoCD Application Specification](https://argo-cd.readthedocs.io/en/stable/user-guide/application-specification/)
