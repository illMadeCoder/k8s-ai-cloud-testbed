# illm-k8s-ai-lab

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
    │
    ▼
Benchmark Results Site (GitHub Pages, Astro + Vega-Lite)
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
┌─────────────────────────────────────────────────────────────────────┐
│                   Hub Cluster — Talos Linux (on-prem, N100)         │
│                                                                     │
│   ┌───────────────────────────────────────────────────────────┐     │
│   │              Experiment Operator (Go, Kubebuilder)        │     │
│   │                                                           │     │
│   │   Drives the full lifecycle:                              │     │
│   │   Pending → Provisioning → Ready → Running → Complete     │     │
│   └────┬──────────────┬──────────────┬───────────────────┘     │
│        │              │              │                         │
│        ▼              ▼              ▼                         │
│   Crossplane v2   ArgoCD        Argo Workflows                │
│   (provisions     (deploys 38   (validation,                  │
│    GKE/EKS/AKS)   apps)         load gen)                     │
│                                                               │
│   ┌─────────────────────────────────────────────────────┐     │
│   │  VictoriaMetrics │ SeaweedFS S3 │ Kyverno + Cosign  │     │
│   │  (metrics hub)   │ (results)    │ (policy + signing) │     │
│   │  OpenBao         │ Loki / Tempo │                    │     │
│   │  (secrets)       │ (logs/traces)│                    │     │
│   └─────────────────────────────────────────────────────┘     │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                Crossplane provisions, ArgoCD deploys
                                │
              ┌─────────────────┼─────────────────┐
              ▼                 ▼                  ▼
     ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
     │     GKE      │  │     EKS      │  │     AKS      │
     │  (spot/pre-  │  │  (spot)      │  │  (spot)      │
     │   emptible)  │  │              │  │              │
     │              │  │              │  │              │
     │ Apps + Obs + │  │ Apps + Obs + │  │ Apps + Obs + │
     │ Load Gen     │  │ Load Gen     │  │ Load Gen     │
     └──────────────┘  └──────────────┘  └──────────────┘
              │                 │                  │
              └─────── metrics flow back ─────────┘
                        to VictoriaMetrics
```

## Experiment Model

Two custom CRDs define the experiment system:

**Experiment** (namespaced) — what to run:
```yaml
spec:
  targets:
    - name: target
      cluster: { provider: gke, preemptible: true }
      components:
        - app: prometheus        # resolves to Component CRD
        - app: hello-app
  workflow: { template: experiment-lifecycle }
  metrics:
    queries:
      - name: p99_latency
        query: histogram_quantile(0.99, rate(http_duration_seconds_bucket[5m]))
```

**Component** (cluster-scoped) — reusable building blocks:
```yaml
spec:
  sources:
    helm: { repo: https://prometheus-community.github.io/helm-charts, chart: prometheus }
  parameters: { retention: 24h }
  observability: { serviceMonitor: true }
```

Lifecycle: `Pending → Provisioning → Ready → Running → Complete → Results (S3 + charts)`

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

## Quick Start

```bash
# Deploy an experiment
kubectl create -f experiments/hello-app/experiment.yaml

# Watch: Pending → Provisioning → Ready → Running → Complete
kubectl get experiments -n experiments -w

# Check results
kubectl get experiments -n experiments -o wide
```

## License

MIT
