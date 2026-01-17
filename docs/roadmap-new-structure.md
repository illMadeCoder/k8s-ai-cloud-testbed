# illm-k8s-ai-lab Roadmap (Consolidated)

**Updated:** 2026-01-17
**Structure:** 9 core phases + 18 appendices

A focused Kubernetes learning lab for **Cloud Architect**, **Platform Architect**, and **Solutions Architect** roles. Core phases cover essential production infrastructure patterns. Appendices provide advanced specialization topics.

---

## Core Learning Path

Complete these 9 phases for a portfolio-ready Kubernetes lab:

| Phase | Topic | Status | Scope |
|-------|-------|--------|-------|
| **1** | [Platform Bootstrap & GitOps](roadmap/phase-01-platform-bootstrap.md) | ✅ Complete | Hub cluster, ArgoCD, Crossplane, OpenBao, Argo Workflows |
| **2** | [CI/CD & Supply Chain](roadmap/phase-02-cicd-supply-chain.md) | ✅ Complete | GitHub Actions, Cosign, SBOM, Kyverno, Image Updater |
| **3** | [Observability](roadmap/phase-03-observability.md) | 🚧 In Progress | Prometheus, Loki, Tempo, Grafana, Pyrra SLOs, SeaweedFS |
| **4** | Traffic Management | 📋 Planned | Gateway API, ingress, routing, load balancing |
| **5** | Data & Persistence | 📋 Planned | PostgreSQL, Redis, backup/DR, schema migration |
| **6** | Security & Policy | 📋 Planned | TLS, secrets (ESO+OpenBao), RBAC, admission control, NetworkPolicy |
| **7** | Service Mesh | 📋 Planned | Istio, Linkerd, Cilium comparison |
| **8** | Messaging & Events | 📋 Planned | Kafka, RabbitMQ, NATS, CloudEvents patterns |
| **9** | Autoscaling & Resources | 📋 Planned | HPA, KEDA, VPA, cluster autoscaling |

---

## Phase Progression

```
Foundation              Infrastructure         Advanced
─────────────────       ────────────────────   ─────────────────────
1. Platform             4. Traffic Mgmt        7. Service Mesh
2. CI/CD                5. Data & Storage      8. Messaging
3. Observability        6. Security            9. Autoscaling
```

**Dependencies:**
```
Phase 1 (Platform)
   ↓
Phase 2 (CI/CD)
   ↓
Phase 3 (Observability) ←─────────┐
   ↓                               │
Phase 4 (Traffic Management)       │
   ↓                               │
Phase 5 (Data & Persistence)       │
   ↓                               │
Phase 6 (Security & Policy) ───────┘
   ↓
Phase 7 (Service Mesh)
   ↓
Phase 8 (Messaging & Events)
   ↓
Phase 9 (Autoscaling & Resources)
```

---

## Advanced Topics (Appendices)

Dive deeper into specialized areas after completing core phases:

### Cloud & Platform Engineering
- **[A: MLOps & AI Infrastructure](roadmap/appendix-mlops.md)** - Kubeflow, KServe, GPU scheduling
- **[G: Deployment Strategies](roadmap/appendix-deployment-strategies.md)** - Rolling, blue-green, canary, feature flags
- **[P: Chaos Engineering](roadmap/appendix-chaos-engineering.md)** - Pod/network/infra chaos, SLO impact
- **[Q: Advanced Workflow Patterns](roadmap/appendix-advanced-workflows.md)** - Argo Events, Tekton, GitOps workflows
- **[R: Internal Developer Platforms](roadmap/appendix-idp.md)** - Backstage, self-service, golden paths

### Security & Compliance
- **[B: Identity & Authentication](roadmap/appendix-identity-auth.md)** - OAuth, OIDC, JWT, password security
- **[C: PKI & Certificate Management](roadmap/appendix-pki-certs.md)** - X.509, TLS, step-ca, cert-manager, SPIFFE
- **[D: Compliance & Security Operations](roadmap/appendix-compliance-soc.md)** - SOC, PCI-DSS, HIPAA, zero trust
- **[O: SLSA Framework Deep Dive](roadmap/appendix-slsa.md)** - SLSA Levels 1-4, provenance, Sigstore

### Architecture & Design
- **[E: Distributed Systems Fundamentals](roadmap/appendix-distributed-systems.md)** - CAP, consensus, replication
- **[F: API Design & Contracts](roadmap/appendix-api-design.md)** - REST, GraphQL, gRPC, OpenAPI
- **[H: gRPC & HTTP/2 Patterns](roadmap/appendix-grpc.md)** - Protocol details, streaming, load balancing
- **[K: Event-Driven Architecture](roadmap/appendix-event-driven.md)** - Event sourcing, CQRS, Saga patterns

### Performance & Operations
- **[I: Container & Runtime Internals](roadmap/appendix-container-internals.md)** - Namespaces, cgroups, OCI runtimes
- **[J: Performance Engineering](roadmap/appendix-performance-engineering.md)** - Profiling, load testing, optimization
- **[L: Database Internals](roadmap/appendix-database-internals.md)** - Storage engines, indexing, sharding
- **[M: SRE Practices & Incident Management](roadmap/appendix-sre-practices.md)** - On-call, incident response, toil
- **[S: Web Serving Internals](roadmap/appendix-web-serving-internals.md)** - Threading models, HTTP/2/3, runtimes

### Multi-Cloud
- **[N: Multi-Cloud & Disaster Recovery](roadmap/appendix-multicloud-dr.md)** - Multi-cloud strategy, DR planning

---

## Principles

- **Core = Portfolio Demonstrations** - 9 phases show production-ready infrastructure skills
- **Appendices = Specialization** - Deep dives for career-specific needs
- **Sequential but flexible** - Phases build on each other, skip ahead if experienced
- **ADRs continuous** - Document decisions as you make them (50+ throughout)
- **Validation first** - Every experiment must be played through before moving on
- **Supply chain security from day one** - SLSA Level 2 from Phase 2 onward

---

## Environment

| Cluster | Purpose | Platform | Status |
|---------|---------|----------|--------|
| **Hub** | Control plane | Talos (N100 home lab) | ✅ Running |
| **Target** | Workloads | Talos / AKS / EKS | 📋 Planned |
| **LoadGen** | Load testing | Talos / AKS / EKS | 📋 Planned |

---

## Current Focus (Phase 3)

**Observability - 60% Complete**

Completed:
- ✅ Prometheus + Grafana (metrics-app, RED dashboards)
- ✅ Victoria Metrics comparison
- ✅ SeaweedFS object storage

Backlog (needs validation):
- [ ] Loki tutorial
- [ ] Elasticsearch tutorial
- [ ] Logging comparison
- [ ] OpenTelemetry tutorial
- [ ] Tempo tutorial
- [ ] Jaeger tutorial
- [ ] Tracing comparison
- [ ] Pyrra SLOs
- [ ] Observability cost management

**Next:** Validate all 9 backlog experiments (2 weeks)

---

## Quick Start

```bash
# Prerequisites: Docker, kubectl, task, helm

task hub:bootstrap                      # Create hub cluster
task hub:conduct -- prometheus-tutorial # Run an experiment
task hub:down -- prometheus-tutorial    # Cleanup
task hub:destroy                        # Destroy cluster
```

---

## Progress Metrics

| Metric | Value |
|--------|-------|
| Phases complete | 2 / 9 (22%) |
| Phases in progress | 1 / 9 (11%) |
| Experiments validated | 8 / ~50 (16%) |
| ADRs documented | 13 |
| Appendices available | 18 |

---

## Timeline

**Estimated completion:** 4-5 months

- Phase 3 validation: 2 weeks
- Roadmap restructure: 1 week
- Phases 4-9: 3-4 months

---

## What Changed (2026-01-17 Consolidation)

**Before:**
- 16 core phases + 12 appendices
- ~80-90 experiments
- 10-12 months estimated

**After:**
- **9 core phases + 18 appendices**
- **~45-50 core experiments**
- **4-5 months to portfolio-ready**

**Moved to Appendices:**
- Phase 5 (Deployment Strategies) → Appendix G
- Phase 12 (Chaos Engineering) → Appendix P
- Phase 13 (Advanced Workflows) → Appendix Q
- Phase 14 (Backstage IDP) → Appendix R
- Phase 15 (Benchmarks) → Deleted (redistributed inline)
- Phase 16 (Web Serving) → Appendix S
- gRPC deep dive → Appendix H

**Consolidated:**
- Phase 7 + 8 → Phase 6 (Security & Policy)

**Result:**
- 44% reduction in core scope
- 50% fewer core experiments
- 6-7 months saved
- All content preserved for specialization

---

## Resources

- [Strategic Review (2026-01)](strategic-review-2026-01.md)
- [Consolidation Analysis](roadmap-consolidation-analysis.md)
- [GitOps Patterns](gitops-patterns.md)
- [ADRs](adrs/)
- [Experiments](../experiments/scenarios/)
