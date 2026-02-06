# illm-k8s-ai-lab

A Kubernetes learning lab with GitOps, supply chain security, and observability. 17 hands-on experiments, 12 ADRs, and a 16-phase roadmap running on Kind, Talos (home lab), and cloud via Crossplane.

## Highlights

| Category | What's Here |
|----------|-------------|
| **GitOps** | ArgoCD app-of-apps, sync waves, multi-source, Image Updater |
| **Supply Chain** | SLSA Level 2: Cosign keyless signing, SBOM attestation, Kyverno admission |
| **Observability** | Prometheus, Loki, Tempo, Grafana + Pyrra SLOs with error budgets |
| **Platforms** | Kind (local), Talos (home lab hardware), AKS/EKS (via Crossplane) |
| **Secrets** | OpenBao + External Secrets Operator, internal PKI |

## Project Structure

```
platform/
├── hub/                # Control plane (runs on Kind locally)
│   ├── apps/           # ArgoCD Applications (GitOps root)
│   ├── values/         # Helm values for hub components
│   ├── manifests/      # Raw K8s manifests
│   ├── cluster/        # Kind cluster provisioning
│   └── bootstrap/      # Initial ArgoCD setup
└── targets/            # Workload clusters managed by hub
    └── talos/          # Home lab (N100 hardware)

experiments/             # 17 runnable tutorials (one dir per experiment)

components/              # Reusable infra (50+ components)

docs/
├── adrs/               # Architecture Decision Records
└── roadmap/            # 16 phases + 12 appendices
```

## Architecture

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                         Hub Cluster (Kind)                                   │
│                                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   ArgoCD    │  │   OpenBao   │  │  Crossplane │  │      Kyverno        │  │
│  │   (GitOps)  │  │  (Secrets)  │  │   (Infra)   │  │ (Policy/Admission)  │  │
│  └──────┬──────┘  └─────────────┘  └──────┬──────┘  └─────────────────────┘  │
│         │                                 │                                  │
│         │  ┌─────────────┐  ┌─────────────┐                                  │
│         │  │   MetalLB   │  │ k8s_gateway │                                  │
│         │  │ (LoadBal)   │  │    (DNS)    │                                  │
│         │  └─────────────┘  └─────────────┘                                  │
└─────────┼─────────────────────────────────┼──────────────────────────────────┘
          │                                 │
          │    Deploys workloads            │    Provisions clusters
          │                                 │
          └────────────┬────────────────────┴────────────────────┐
                       │                                         │
                       ▼                                         ▼
┌───────────────────────────────────────────────┐  ┌───────────────────────────┐
│            Target Cluster(s)                  │  │    Load Gen Cluster(s)    │
│         Talos (on-prem) / AKS / EKS           │  │      Talos / AKS / EKS    │
│                                               │  │                           │
│  ┌─────────────────────────────────────────┐  │  │  ┌─────────────────────┐  │
│  │            Observability                │  │  │  │    k6 / Locust      │  │
│  │  Prometheus, Loki, Tempo, Grafana       │  │  │  │    Load Tests       │  │
│  │  Pyrra (SLOs + error budgets)           │  │  │  │                     │  │
│  └─────────────────────────────────────────┘  │  │  └──────────┬──────────┘  │
│  ┌─────────────────────────────────────────┐  │  │             │             │
│  │            Applications                 │  │  │             │             │
│  │  Demo apps, Argo Workflows              │◄─┼──┼─────────────┘             │
│  │  metrics, logs, traces                  │  │  │        HTTP traffic       │
│  └─────────────────────────────────────────┘  │  │                           │
└───────────────────────────────────────────────┘  └───────────────────────────┘
```

## Quick Start

```bash
# Prerequisites: Docker, kubectl, task (go-task.dev), helm

task hub:bootstrap                      # Create cluster + GitOps
task hub:conduct -- prometheus-tutorial # Run an experiment
task hub:down -- prometheus-tutorial
task hub:destroy
```

## Experiments

Run `task hub:conduct -- <name>` to deploy any scenario:

| Scenario | What You Learn |
|----------|----------------|
| **prometheus-tutorial** | Custom metrics, ServiceMonitor, PromQL, RED dashboards |
| **loki-tutorial** | LogQL queries, label indexing, Promtail pipelines |
| **otel-tutorial** | Distributed tracing, OTLP, span correlation, TraceQL |
| **slo-tutorial** | Error budgets, multi-burn-rate alerts, Pyrra |
| **tsdb-comparison** | Prometheus vs Victoria Metrics (resource efficiency) |
| **logging-comparison** | Loki vs Elasticsearch (trade-offs) |
| **tracing-comparison** | Tempo vs Jaeger |
| **cicd-fundamentals** | GitHub Actions + Cosign + SBOM + Image Updater |
| **seaweedfs-tutorial** | O(1) lookups, Haystack architecture |
| **observability-cost-tutorial** | Cardinality, log volume, retention tuning |

[Full list →](experiments/)

## Architecture Decisions

12 ADRs document real trade-offs:

| ADR | Decision |
|-----|----------|
| [007](docs/adrs/ADR-007-supply-chain-security.md) | Sigstore ecosystem for supply chain (Cosign + Syft + Rekor) |
| [011](docs/adrs/ADR-011-observability-architecture.md) | Three-pillar observability with metric↔log↔trace correlation |
| [012](docs/adrs/ADR-012-crossplane-experiment-abstraction.md) | Single `ExperimentCluster` CRD for all platforms |

[All ADRs →](docs/adrs/)

## Roadmap

16 phases from platform bootstrap to chaos engineering:

| Phase | Status |
|-------|--------|
| 1. Platform Bootstrap (ArgoCD, OpenBao, Crossplane) | Complete |
| 2. CI/CD & Supply Chain Security | Complete |
| 3. Observability (metrics, logs, traces, SLOs) | In Progress |
| 4-16. Traffic, Deployments, Security, Mesh, Chaos... | Planned |

[Full roadmap →](docs/roadmap.md)

## Supply Chain Security (SLSA Level 2)

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   GitHub    │     │   Trivy     │     │    Syft     │     │   Cosign    │
│   Actions   │────►│   (scan)    │────►│   (SBOM)    │────►│   (sign)    │
│             │     │             │     │             │     │  keyless    │
└─────────────┘     └─────────────┘     └─────────────┘     └──────┬──────┘
                                                                   │
                    ┌──────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────┐     ┌─────────────────────────────┐
│              GHCR                   │     │      Rekor                  │
│  ┌───────────┐  ┌────────────────┐  │     │   (transparency log)        │
│  │   Image   │  │  Attestations  │  │     │                             │
│  │           │  │  - signature   │  │     │   Public record of          │
│  │           │  │  - SBOM        │  │     │   all signatures            │
│  └───────────┘  └────────────────┘  │     └─────────────────────────────┘
└──────────────────┬──────────────────┘
                   │
                   ▼
┌─────────────────────────────────────┐     ┌─────────────────────────────┐
│      ArgoCD Image Updater           │     │         Kyverno             │
│                                     │     │                             │
│   Detects new image, updates        │────►│   Verifies signature        │
│   deployment manifests              │     │   before admission          │
└─────────────────────────────────────┘     └─────────────────────────────┘
```

## License

MIT
