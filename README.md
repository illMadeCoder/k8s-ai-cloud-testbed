# k8s-ai-cloud-testbed

AI-driven Kubernetes benchmarking lab. A custom operator orchestrates experiments
across on-prem and cloud clusters, collects metrics, and publishes results — all
from a terminal conversation with Claude.

## How It Works

```
Terminal (Claude Code)
    │
    │  "Run the TSDB comparison experiment"
    ▼
┌──────────────────────────────────────────────────┐
│  Experiment CRD applied to hub cluster           │
│                                                  │
│  Experiment Operator reconciles:                 │
│    1. Crossplane  → provisions GKE cluster       │
│    2. ArgoCD      → deploys components           │
│    3. Workflows   → runs validation + load gen   │
│    4. Metrics     → collects Prometheus queries   │
│    5. Storage     → writes results to SeaweedFS  │
└──────────────────────────────────────────────────┘
```

```bash
# Deploy an experiment
kubectl create -f experiments/tsdb-comparison/experiment.yaml

# Watch the lifecycle
kubectl get experiments -n experiments -w
# NAME                    PHASE          AGE
# tsdb-comparison-x7k2q   Provisioning   10s
# tsdb-comparison-x7k2q   Running        9m
# tsdb-comparison-x7k2q   Complete       25m
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│              Hub Cluster — Talos Linux (on-prem, N100)          │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │          Experiment Operator (Go, Kubebuilder)            │  │
│  │                                                           │  │
│  │  Drives the full lifecycle:                               │  │
│  │  Pending → Provisioning → Ready → Running → Complete      │  │
│  └──────┬────────────────┬────────────────┬──────────────────┘  │
│         │                │                │                     │
│         ▼                ▼                ▼                     │
│    Crossplane v2     ArgoCD          Argo Workflows             │
│    (provisions       (deploys 38     (validation,               │
│     GKE/EKS/AKS)     apps)           load gen)                 │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │ VictoriaMetrics │ SeaweedFS S3 │ Kyverno + Cosign         │  │
│  │ (metrics hub)   │ (results)    │ (policy + signing)       │  │
│  │ OpenBao         │ Loki / Tempo │                          │  │
│  │ (secrets)       │ (logs/traces)│                          │  │
│  └───────────────────────────────────────────────────────────┘  │
└────────────────────────────────┬────────────────────────────────┘
                                 │
             Crossplane provisions, ArgoCD deploys
                                 │
            ┌────────────────────┼────────────────────┐
            ▼                    ▼                     ▼
   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
   │     GKE      │    │     EKS      │    │     AKS      │
   │ (preemptible)│    │    (spot)    │    │    (spot)    │
   │              │    │              │    │              │
   │  Experiment  │    │  Experiment  │    │  Experiment  │
   │  workloads   │    │  workloads   │    │  workloads   │
   │      │       │    │      │       │    │      │       │
   │  Alloy agent │    │  Alloy agent │    │  Alloy agent │
   │  (4 metrics) │    │  (4 metrics) │    │  (4 metrics) │
   └──────┬───────┘    └──────┬───────┘    └──────┬───────┘
          │                   │                    │
          └─── Tailscale (encrypted) ─── remote write ──► hub VictoriaMetrics
```

**Metrics backhaul** — The operator auto-injects a Grafana Alloy agent onto each
target cluster. Alloy scrapes cAdvisor, filters to four key metrics (CPU, memory,
network RX/TX), stamps them with `{experiment="<name>"}`, and remote-writes to
hub VictoriaMetrics through a Tailscale WireGuard tunnel. No monitoring stack needed
on the target — the hub collects everything.

## Getting Started

### Prerequisites

| Requirement | Purpose |
|-------------|---------|
| **Talos Linux cluster** | Hub cluster (tested on Intel N100 mini-PC, 16 GB RAM) |
| **Terminal machine** | Any machine with network access to the Talos node (WSL, Linux, macOS) |
| **kubectl** | Kubernetes CLI |
| **talosctl** | Talos Linux management |
| **helm** | Initial ArgoCD install only — ArgoCD manages everything after bootstrap |
| **GCP/AWS/Azure credentials** | For Crossplane to provision experiment clusters (optional for hub-only use) |

### 1. Fork and Configure

ArgoCD syncs from Git, so it needs a repo you control.

```bash
# Fork this repo on GitHub, then clone your fork
git clone https://github.com/<your-username>/k8s-ai-cloud-testbed.git
cd k8s-ai-cloud-testbed

# Update the ArgoCD root app to point to your fork
sed -i 's|illMadeCoder|<your-username>|g' platform/bootstrap/hub-application.yaml
git commit -am "point ArgoCD at my fork" && git push
```

All 38 ArgoCD apps reference this repo for values and manifests. If you want
ArgoCD to sync your own changes, update the `repoURL` in `platform/apps/` too.

### 2. Bootstrap the Hub

The hub cluster runs all control-plane services. Bootstrap is a three-step
process — after that, ArgoCD manages everything from Git.

```bash
# Install kube-vip (LoadBalancer support) and local-path-provisioner (storage)
kubectl apply -f platform/bootstrap/kube-vip.yaml
kubectl apply -f platform/bootstrap/local-path-provisioner.yaml

# Install ArgoCD via Helm (one-time — ArgoCD self-manages after this)
helm repo add argo https://argoproj.github.io/argo-helm
helm install argocd argo/argo-cd \
  --namespace argocd --create-namespace \
  --values platform/bootstrap/argocd-values-talos.yaml

# Apply the root app-of-apps — ArgoCD discovers and syncs all 38 child apps
kubectl apply -f platform/bootstrap/hub-application.yaml
```

ArgoCD reads `platform/apps/` from your fork and deploys the full stack:
Crossplane, Kyverno, OpenBao, VictoriaMetrics, SeaweedFS, Argo Workflows,
the experiment operator, and everything else. Sync waves ensure correct ordering.

```bash
# Watch apps come up (~10 min for full convergence)
kubectl get applications -n argocd -w
```

### 3. Initialize Secrets

OpenBao stores all secrets centrally. External Secrets Operator syncs them to
Kubernetes secrets automatically.

```bash
# Initialize OpenBao (first time only)
kubectl exec -n openbao openbao-0 -- bao operator init \
  -key-shares=1 -key-threshold=1 -format=json > ~/.illmlab/openbao-keys.json

# Unseal
kubectl exec -n openbao openbao-0 -- bao operator unseal \
  $(jq -r '.unseal_keys_b64[0]' ~/.illmlab/openbao-keys.json)

# Store cloud credentials for Crossplane (GCP example)
kubectl exec -n openbao openbao-0 -- bao kv put secret/cloud/gcp \
  credentials=@~/.secrets/gcp-credentials.json
```

### 4. Run an Experiment

```bash
# Create (generateName gives a unique suffix each time)
kubectl create -f experiments/hello-app/experiment.yaml

# Watch: Pending → Provisioning → Ready → Running → Complete
kubectl get experiments -n experiments -w

# Check results
kubectl get experiments -n experiments -o wide   # Shows ResultsURL column
```

The operator handles the full lifecycle — provisioning cloud clusters, deploying
components via ArgoCD, running workflows, collecting metrics, and storing results
in SeaweedFS S3.

## Experiment Model

Two custom CRDs define the experiment system: **Experiment** (namespaced) and
**Component** (cluster-scoped).

### Operator Lifecycle

```
kubectl create -f experiment.yaml
         │
         ▼
    ┌─────────┐     Crossplane provisions    ┌──────────────┐
    │ Pending  │ ──────────────────────────►  │ Provisioning │
    └─────────┘     GKE/EKS/AKS cluster      └──────┬───────┘
                                                     │
                    ArgoCD deploys components         │
                    + injects Alloy agent             ▼
                                               ┌───────────┐
                    Argo Workflows runs         │   Ready    │
                    validation + load gen       └─────┬─────┘
                                                      │
                         ┌────────────────────────────┘
                         ▼
                    ┌───────────┐
                    │  Running  │ ◄── manual mode: stays here until user tears down
                    └─────┬─────┘
                          │  workflow completes (or user deletes)
                          ▼
                    ┌───────────┐     1. Collect metrics (target Prometheus → hub VM fallback)
                    │ Complete  │     2. Store summary.json → SeaweedFS S3          (always)
                    └─────┬─────┘     3. Commit to benchmark site via GitHub API   (publish: true)
                          │           4. Launch AI analyzer Job (Claude Code)       (publish: true)
                          │           5. Delete cloud cluster + ArgoCD apps
                          ▼
                       Cleaned
```

At any point, unrecoverable errors transition to `Failed`. Steps 3–4 are
best-effort and non-fatal — a GitHub API or analyzer failure won't block cleanup.

### Experiment Spec

```yaml
apiVersion: experiments.illm.io/v1alpha1
kind: Experiment
metadata:
  generateName: tsdb-comparison-      # K8s appends a unique suffix
  namespace: experiments
spec:
  description: "Prometheus vs VictoriaMetrics"
  tags: ["comparison", "observability", "metrics"]
  publish: true                        # publish results to benchmark site + AI analysis
  ttlDays: 7                           # auto-delete after N days (default: 1)

  targets:
    - name: prometheus                 # deploy target name
      cluster:
        type: gke                      # gke | hub
        preemptible: true              # use preemptible/spot nodes
        # zone: us-central1-a          # GCP zone (optional)
        # machineType: e2-standard-4   # node machine type (optional)
        # nodeCount: 3                 # node count (optional)
        # diskSizeGb: 50              # disk size (optional)
      components:
        - app: kube-prometheus-stack    # resolves to Component CRD by name
          params:                       # override component parameters
            grafana.enabled: "true"
        - app: metrics-app
      observability:
        enabled: true                  # inject Alloy + Tailscale for metrics backhaul
        transport: tailscale           # tailscale | direct
      # depends: [other-target]        # wait for another target to be ready first

  workflow:
    template: experiment-lifecycle      # Argo WorkflowTemplate name
    completion:
      mode: workflow                   # workflow: auto-complete | manual: stay running
    # params:                          # extra workflow parameters
    #   duration: "10m"

  metrics:                             # custom PromQL queries (optional — defaults provided)
    - name: p99_latency
      query: histogram_quantile(0.99, rate(http_duration_seconds_bucket[5m]))
      type: range                      # instant (bar chart) | range (line chart)
      unit: seconds
      description: "P99 request latency"

  # tutorial:                          # interactive mode — exposes kubeconfig + service endpoints
  #   exposeKubeconfig: true
  #   services:
  #     - name: grafana
  #       target: prometheus
  #       service: kube-prometheus-stack-grafana
  #       namespace: tsdb-comparison
```

### Completion Modes

| Mode | Behavior | Use case |
|------|----------|----------|
| `workflow` | Auto-completes when Argo Workflow finishes | Comparisons, benchmarks |
| `manual` | Stays `Running` until user deletes the experiment | Tutorials, interactive demos |

### Component CRD

Reusable building blocks referenced by `spec.targets[].components[].app`:

```yaml
apiVersion: experiments.illm.io/v1alpha1
kind: Component
metadata:
  name: kube-prometheus-stack          # referenced by app: field
spec:
  sources:
    helm:
      repo: https://prometheus-community.github.io/helm-charts
      chart: kube-prometheus-stack
  parameters: { retention: 24h }
  observability: { serviceMonitor: true }
```

42 components across 8 categories in `components/`.

## Project Structure

```
operators/experiment-operator/       Custom Go operator (Kubebuilder)
  internal/
    controller/                      Reconciliation loop
    crossplane/                      GKE/EKS/AKS cluster provisioning
    argocd/                          Application lifecycle + cluster registration
    components/                      ComponentRef → Git/Helm source resolution
    workflow/                        Argo Workflows integration
    metrics/                         Prometheus query collection
    storage/                         SeaweedFS S3 results storage
components/                          42 reusable components (8 categories)
experiments/                         15 experiment scenarios
platform/                            Hub cluster infra (38 ArgoCD apps)
docs/adrs/                           17 Architecture Decision Records
docs/roadmap/                        10 phases, ~50 target experiments
.github/workflows/                   SLSA build pipelines
```

## Supply Chain Security — SLSA Level 2

```
GitHub Actions ──► Trivy scan ──► Syft SBOM ──► Cosign keyless sign
                                                       │
                   ┌───────────────────────────────────┘
                   ▼
     ┌──────────────────────────────┐    ┌─────────────────────────┐
     │            GHCR              │    │         Rekor           │
     │  ┌────────┐  ┌───────────┐  │    │   (transparency log)    │
     │  │ Image  │  │Attestation│  │    │                         │
     │  │        │  │ sig + SBOM│  │    │   Public record of      │
     │  └────────┘  └───────────┘  │    │   all signatures        │
     └──────────────┬───────────────┘    └─────────────────────────┘
                    │
                    ▼
     ┌──────────────────────────────┐
     │          Kyverno             │
     │  Verifies Cosign signature   │
     │  before pod admission        │
     └──────────────────────────────┘
```

## Experiments

**Deployable now:** hello-app, prometheus-tutorial, loki-tutorial

**Scenarios defined:**

| Scenario | Focus |
|----------|-------|
| tsdb-comparison | Prometheus vs VictoriaMetrics (resource efficiency) |
| gateway-comparison | Envoy Gateway vs Kong vs Traefik |
| logging-comparison | Loki vs Elasticsearch (trade-offs) |
| tracing-comparison | Tempo vs Jaeger |
| otel-tutorial | Distributed tracing, OTLP, span correlation |
| slo-tutorial | Error budgets, multi-burn-rate alerts, Pyrra |
| observability-cost-tutorial | Cardinality, log volume, retention tuning |
| seaweedfs-tutorial | O(1) lookups, Haystack architecture |
| cicd-fundamentals | GitHub Actions + Cosign + SBOM |

15 scenarios across observability, traffic, security, and cost engineering.
Target: ~50 experiments across 10 phases.

## Roadmap

| Phase | Topic | Status |
|-------|-------|--------|
| 1 | Platform Bootstrap & GitOps | Complete |
| 2 | CI/CD & Supply Chain Security | Complete |
| 3 | Observability | In Progress |
| 4 | Traffic Management | In Progress |
| 5 | Data Persistence | Planned |
| 6 | Security & Policy | Planned |
| 7 | Service Mesh | Planned |
| 8 | Messaging & Events | Planned |
| 9 | Autoscaling & Resources | Planned |
| 10 | Performance & Cost Engineering | Planned |

[Full roadmap →](docs/roadmap/)

## Architecture Decisions

| ADR | Decision |
|-----|----------|
| [015](docs/adrs/ADR-015-experiment-operator.md) | Go operator over Crossplane XRD for experiment orchestration |
| [017](docs/adrs/ADR-017-benchmark-results-site.md) | GitHub Pages + Astro + Vega-Lite for benchmark results site |
| [016](docs/adrs/ADR-016-hub-metrics-backend.md) | VictoriaMetrics Single over Mimir for hub metrics backhaul |
| [007](docs/adrs/ADR-007-supply-chain-security.md) | Sigstore ecosystem for supply chain security |
| [013](docs/adrs/ADR-013-crossplane-v2-upgrade.md) | Crossplane v2 upgrade — Pipeline mode compositions |

17 ADRs document trade-offs with data. [All ADRs →](docs/adrs/)

## License

MIT
