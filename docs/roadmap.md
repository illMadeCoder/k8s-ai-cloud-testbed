# illm-k8s-ai-lab Roadmap

A **benchmarking-focused** Kubernetes experiment lab for **Cloud Architect**, **Platform Architect**, and **Solutions Architect** roles. Deploy, measure, cost-attribute each technology in the ecosystem.

## How to Use This Roadmap

- **Sequential but flexible** - Phases build on each other, but skip ahead if you have experience
- **Each phase has experiments** - Found in `experiments/scenarios/<topic>/`
- **ADRs are continuous** - Document decisions as you make them (49+ ADRs throughout)
- **FinOps in every phase** - Cost considerations integrated, not bolted on
- **Appendices are optional** - Deep dives for specialized needs
- **Run experiments manually first** - Automation comes from understanding

### Cross-Cutting Concerns

| Concern | Where It Appears |
|---------|-----------------|
| **ADRs** | Every phase - document decisions as you make them |
| **FinOps** | Every phase - cost per component, cost per transaction |
| **SLOs/SLAs** | Phase 3 (foundation), Phase 10 (capstone) |
| **Security** | Phase 6 (consolidated security & policy) |
| **Testing** | Phase 2 (foundation), plus validation in every experiment |

| | |
|---|---|
| **Target** | ~50-55 core experiments across 10 phases + appendices |
| **Environment** | Kind (local), Talos on N100 (home lab), AKS/EKS (cloud) |
| **Focus** | Benchmarking each technology with data-driven decisions |
| **Timeline** | 5-6 months (core), 6-7 months (with priority appendices) |

**Principles:**
- Supply chain security from day one (Phase 2)
- Build something before securing it (Data before Security)
- Benchmark everything - measure before recommending
- ADRs mandatory for technology decisions
- Runbooks accompany operational components

---

## Phases (10 Core)

*Consolidated from 16 phases. Focus on benchmarking each technology.*

| Phase | Topic | Status | Details |
|-------|-------|--------|---------|
| 1 | [Platform Bootstrap & GitOps](docs/roadmap/phase-01-platform-bootstrap.md) | Complete | Hub, orchestrator, Argo Workflows, Crossplane |
| 2 | [CI/CD & Supply Chain](docs/roadmap/phase-02-cicd-supply-chain.md) | Complete | Image building, scanning, SBOM, signing |
| 3 | [Observability](docs/roadmap/phase-03-observability.md) | In Progress | Prometheus, Loki, Tempo, Grafana, SLOs |
| 4 | [Traffic Management](docs/roadmap/phase-04-traffic-management.md) | Not Started | Gateway API, ingress comparison |
| 5 | [Data & Persistence](docs/roadmap/phase-05-data-persistence.md) | Not Started | PostgreSQL, Redis, backup, benchmark |
| 6 | [Security & Policy](docs/roadmap/phase-06-security-policy.md) | Not Started | TLS, secrets (ESO), RBAC, Kyverno, NetworkPolicy |
| 7 | [Service Mesh](docs/roadmap/phase-07-service-mesh.md) | Not Started | Istio vs Linkerd vs Cilium, overhead benchmark |
| 8 | [Messaging & Events](docs/roadmap/phase-08-messaging-events.md) | Not Started | Kafka vs RabbitMQ vs NATS, throughput benchmark |
| 9 | [Autoscaling & Resources](docs/roadmap/phase-09-autoscaling-resources.md) | Not Started | HPA, VPA, KEDA, cluster autoscaling |
| 10 | [Performance & Cost Engineering](docs/roadmap/phase-10-performance-cost-engineering.md) | Not Started | Runtime comparison, full stack benchmark, cost per transaction |

### The Capstone (Phase 10)

*Synthesize everything: Runtime comparison (Go/Rust/.NET/Node/Bun), full stack composition benchmark, cost per transaction. Data-driven system engineering.*

---

## Progression

```
Foundation              Stateful + Security        Complexity
───────────────────     ───────────────────────    ───────────────────────
1. Platform Bootstrap   5. Data & Persistence      7. Service Mesh
2. CI/CD & Supply Chain 6. Security & Policy       8. Messaging & Events
3. Observability                                   9. Autoscaling
4. Traffic Management

                        Capstone
                        ───────────────────────
                        10. Performance & Cost
                            (Full stack benchmark)
```

---

## Quick Links

- [Learning Path Summary](docs/roadmap/learning-path-summary.md)
- [Notes](docs/roadmap/notes.md)
- [GitOps Patterns](docs/gitops-patterns.md)

### Appendices (Optional Deep Dives)

**Priority Appendices (After Core Phases)**
- [N: Deployment Strategies](docs/roadmap/appendix-n-deployment-strategies.md) - Blue-green, canary, progressive delivery, feature flags
- [O: gRPC & HTTP/2 Patterns](docs/roadmap/appendix-o-grpc-http2.md) - Protocol deep dive, load balancing, observability
- [T: eBPF & Advanced Metrics](docs/roadmap/appendix-t-ebpf-metrics.md) - I/O tracing, Pixie, Parca, Tetragon
- [U: Chi Observability Stack](docs/roadmap/appendix-u-chi-observability.md) - Traffic as energy flow, USE Method, service mesh philosophy

**Specialized Topics (Moved from Core)**
- [P: Chaos Engineering](docs/roadmap/appendix-p-chaos-engineering.md) - Fault injection, resilience testing
- [Q: Advanced Workflow Patterns](docs/roadmap/appendix-q-advanced-workflows.md) - Argo Events, Tekton, advanced CI/CD
- [R: Internal Developer Platforms](docs/roadmap/appendix-r-developer-platforms.md) - Backstage, golden paths, self-service
- [S: Web Serving Internals](docs/roadmap/appendix-s-web-serving-internals.md) - Threading models, HTTP/2/3, proxies, runtimes

**Reference Appendices**
- [A: MLOps & AI Infrastructure](docs/roadmap/appendix-mlops.md) - Kubeflow, KServe, vector DBs, GPU scheduling
- [B: Identity & Authentication](docs/roadmap/appendix-identity-auth.md) - OAuth, OIDC, JWT/JWE, API keys, IdP
- [C: PKI & Certificate Management](docs/roadmap/appendix-pki-certs.md) - X.509, mTLS, SPIFFE
- [D: Compliance & Security Operations](docs/roadmap/appendix-compliance-soc.md) - SOC, PCI-DSS, HIPAA/PHI
- [E: Distributed Systems Fundamentals](docs/roadmap/appendix-distributed-systems.md) - Consensus, CAP, clocks
- [F: API Design & Contracts](docs/roadmap/appendix-api-design.md) - REST, GraphQL, OpenAPI
- [G: Container & Runtime Internals](docs/roadmap/appendix-container-internals.md) - Namespaces, cgroups, OCI
- [H: Performance Engineering](docs/roadmap/appendix-performance-engineering.md) - Profiling, load testing
- [I: Event-Driven Architecture](docs/roadmap/appendix-event-driven.md) - Event sourcing, CQRS, Saga
- [J: Database Internals](docs/roadmap/appendix-database-internals.md) - Storage engines, indexing, sharding, distributed databases, sagas, K8s operators, benchmarking
- [K: SRE Practices & Incident Management](docs/roadmap/appendix-sre-practices.md) - On-call, post-mortems
- [L: Multi-Cloud & Disaster Recovery](docs/roadmap/appendix-multicloud-dr.md) - DR planning, failover
- [M: SLSA Framework Deep Dive](docs/roadmap/appendix-slsa.md) - Provenance, attestations, Sigstore

---

## Current Focus

**Phase 1: Platform Bootstrap**
- [x] Create `platform/hub/` directory structure
- [x] Create ArgoCD bootstrap values with app-of-apps reference
- [x] Create Kind app-of-apps with ArgoCD self-management
- [x] Add MetalLB to Kind app-of-apps
- [x] Add dns-stack to Kind app-of-apps
- [x] Add Argo Workflows to Kind app-of-apps
- [x] Test Kind hub bootstrap end-to-end
- [x] Update `kind:conduct` for orchestrator pattern (parallel provisioning)
- [x] Add OpenBao to Kind app-of-apps
- [x] Add Crossplane to Kind app-of-apps
- [x] Test full hub bootstrap with all components
- [x] Configure webhooks for instant sync (smee.io relay)

**Phase 1 Complete** - Hub bootstraps with ArgoCD, Argo Workflows, OpenBao, Crossplane, MetalLB, dns-stack, webhook relay.

See [Phase 1](docs/roadmap/phase-01-platform-bootstrap.md) for full details.

---

## Recent Progress (2025-12-30)

### Experiment Lifecycle

- [x] Implement full experiment lifecycle (deploy → run → cleanup)
- [x] Add `onExit` cleanup handlers to Argo Workflows
- [x] Configure RBAC for workflow cleanup operations
- [x] Test hello-app and http-baseline experiments end-to-end
- [x] Document pattern in [ADR-005](docs/adrs/ADR-005-experiment-lifecycle.md)

### TLS & Certificates

- [x] Implement TLS persistence via OpenBao (ADR-004)
- [x] Configure External Secrets Operator for TLS sync
- [x] Temporarily disable Let's Encrypt (rate limit exhaustion)
- [x] Add `ignoreDifferences` for ESO-managed secrets
- [x] Re-enable Let's Encrypt after rate limit reset (issuers ready)

### ArgoCD Fixes

- [x] Fix metallb OutOfSync (CRD annotation drift)
- [x] Fix cert-manager-config Degraded (remove deleted file references)
- [x] Fix external-secrets-config Degraded (remove circular PushSecret)
- [x] All apps now Synced and Healthy

---

## Phase 2.1: CI/CD Fundamentals (2025-12-31)

### Directory Refactoring

- [x] Flatten `experiments/components/components/` to `experiments/components/`
- [x] Update 30 path references across YAML files and docs

### CI/CD Pipeline

- [x] Create `cicd-sample` Go app with version endpoint
- [x] Create GitHub Actions workflow (`.github/workflows/cicd-sample.yaml`)
- [x] Trivy vulnerability scanning → GitHub Security tab
- [x] Auto-update K8s manifest with new image SHA
- [x] Document in [ADR-006](docs/adrs/ADR-006-cicd-pipeline.md)
- [x] Create [cicd-fundamentals experiment](experiments/scenarios/cicd-fundamentals/)

---

## Phase 2.2-2.3: Supply Chain Security (2026-01-01)

### Image Signing & SBOM

- [x] Add keyless Cosign signing to CI pipeline
- [x] Generate SBOM with Syft (SPDX format)
- [x] Attest SBOM to image with Cosign
- [x] Add manifest validation with kubeconform
- [x] Document in [ADR-007](docs/adrs/ADR-007-supply-chain-security.md)

### Admission Control (Kyverno)

- [x] Deploy Kyverno via ArgoCD
- [x] Create ClusterPolicy for image signature verification
- [x] Verify signed images from ghcr.io/illmadecoder/* pass
- [x] Policy in Audit mode (ready for Enforce when needed)

### ArgoCD Image Updater

- [x] Add ArgoCD Image Updater to hub
- [x] Configure cicd-sample app with image updater annotations
- [x] Add Kustomize support for Image Updater compatibility
- [x] Remove git-push from CI pipeline (CI decoupled from cluster)
- [x] Test end-to-end flow: CI → GHCR → Image Updater → ArgoCD sync
- [x] Use SHA-based continuous deployment (not semver - apps are per-component, not repo-level)

### Auto-Detection & Dependency Updates

- [x] Create generic `build-components.yaml` workflow (auto-detects changed Dockerfiles)
- [x] Matrix-based parallel builds for multiple app changes
- [x] Flat image naming: `ghcr.io/illmadecoder/{app-name}`
- [x] Configure Renovate for automated dependency updates
- [x] Remove Dependabot (Renovate supports glob patterns, scales to hundreds of apps)

**Phase 2 Complete** - CI/CD pipeline with supply chain security, auto-detection builds, continuous deployment via Image Updater, Renovate for dependency updates.

---

## Phase 3: Observability (In Progress)

**Architecture:** See [ADR-011: Observability Architecture](docs/adrs/ADR-011-observability-architecture.md) for the holistic view of how metrics, logs, traces, and storage integrate.

### Phase 3.1: Prometheus & Grafana

### Observability Foundation

- [x] Fix prometheus-stack component path references
- [x] Create `metrics-app` with custom Prometheus metrics (Counter, Gauge, Histogram, Summary)
- [x] Create ServiceMonitor for scrape discovery
- [x] Create `prometheus-tutorial` experiment scenario
- [x] Document PromQL patterns in experiment README
- [x] Test experiment end-to-end with `task kind:conduct -- prometheus-tutorial`
- [x] Build Grafana RED dashboard for metrics-app

### TSDB Comparison (Phase 3.1 addition)

- [x] Create TSDB comparison tutorial (Prometheus vs Victoria Metrics)
- [x] Create cardinality-generator app
- [x] Deploy and test comparison end-to-end
- [x] Document in ADR-009

### Next Steps

- [x] Re-enable Let's Encrypt after rate limit reset (issuers ready)
- [x] Set up Talos home lab cluster
- [x] Phase 3.2: SeaweedFS Object Storage
- [x] Phase 3.3: Logging Comparison (Loki vs ELK)
  - [x] Create `loki-tutorial` experiment (Loki + Promtail + Grafana)
  - [x] Create `elk-tutorial` experiment (Elasticsearch + Kibana + Fluent Bit)
  - [x] Create `logging-comparison` tutorial experiment
  - [ ] **BACKLOG**: Play through loki-tutorial, elk-tutorial, logging-comparison
- [x] Phase 3.4: OpenTelemetry & Distributed Tracing
  - [x] Create Tempo component (ArgoCD app + Helm values)
  - [x] Create Jaeger component (ArgoCD app + Helm values)
  - [x] Create `otel-demo` multi-service app (user → order → payment)
  - [x] Create `otel-tutorial` experiment (OTel Collector + Tempo + Grafana)
  - [x] Create `tracing-comparison` experiment (Tempo vs Jaeger)
  - [ ] **BACKLOG**: Play through otel-tutorial, tracing-comparison
- [x] Phase 3.5: SLOs & Error Budgets (Pyrra)
  - [x] Create Pyrra component (ArgoCD app + Helm values)
  - [x] Create `slo-tutorial` experiment (Pyrra + error budgets + multi-burn-rate alerts)
  - [ ] **BACKLOG**: Play through slo-tutorial
- [x] Phase 3.6: Observability Cost Management
  - [x] Create `observability-cost-tutorial` experiment (cardinality, log volume, retention)
  - [ ] **BACKLOG**: Play through observability-cost-tutorial

---

## Hub Infrastructure Updates (2026-01)

### Istio Service Mesh for Hub

Replaced Traefik with Istio for hub ingress and mTLS between services. This is infrastructure setup, not Phase 7 experiments (which will compare meshes).

- [x] Deploy Istio via ArgoCD (istio-base, istio-cni, istio-istiod, istio-ingress)
- [x] Configure CNI plugin for Talos compatibility (no iptables in containers)
- [x] Replace Traefik with Istio ingress gateway
- [x] Configure Tailscale LoadBalancer for external access
- [x] Create VirtualServices for path-based routing (/grafana, /mimir, /loki, /tempo, /argocd, /openbao, /kiali)
- [x] Enable sidecar injection for namespaces (observability, seaweedfs, argocd, openbao)
- [x] Configure mTLS with PeerAuthentication (STRICT) and DestinationRules (ISTIO_MUTUAL)
- [x] Deploy Kiali for service graph visualization
- [x] Configure Tempo tracing integration
- [x] Document in [ADR-014](docs/adrs/ADR-014-service-mesh-istio.md)

**Lessons Learned:**
- Kyverno excluded from mesh - init containers need K8s API access before sidecar starts
- `pilot.cni.enabled: true` is the critical setting (not just `global.istio_cni.enabled`)
- Services without sidecars need `mode: DISABLE` in DestinationRules

### SeaweedFS Fix

- [x] Fixed idx volume persistence (was emptyDir, now PVC) - see [ADR-008](docs/adrs/ADR-008-object-storage.md)
