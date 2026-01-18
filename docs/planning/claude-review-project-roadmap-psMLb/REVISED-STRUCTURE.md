# REVISED: Core vs Appendix Split

**Date:** 2026-01-18
**Status:** REVISED based on feedback

---

## Key Change: Fundamentals in Core, Deep Dives in Appendices

**Feedback:** Chi and eBPF should be appendix topics, but we still need service mesh and network observability fundamentals as major phases.

**Revised Approach:**
- **Core phases:** Fundamentals/101 level - what you need to know
- **Priority appendices:** Advanced techniques and philosophies - next level mastery

---

## REVISED 10 Core Phases

| # | Title | Core Content (Fundamentals) | What's NOT in Core |
|---|-------|----------------------------|-------------------|
| **1** | Platform Bootstrap & GitOps ✅ | ArgoCD, Crossplane, OpenBao, Argo Workflows | ✅ Already right-sized |
| **2** | CI/CD & Supply Chain ✅ | GitHub Actions, Cosign, SBOM, Kyverno, Image Updater | ✅ Already right-sized |
| **3** | Observability 🚧 | Prometheus, Loki, Tempo, Grafana, basic metrics | Advanced: eBPF tools → Appendix |
| **4** | Traffic Management | Gateway API, Ingress, basic routing, gateway comparison | Advanced: gRPC deep dive → Appendix |
| **5** | Data & Persistence | PostgreSQL, Redis, backup, basic benchmark | Advanced: Deep DB internals → Appendix |
| **6** | Security & Policy | mTLS, RBAC, NetworkPolicy basics, admission control | Advanced: Zero trust, compliance → Appendix |
| **7** | Service Mesh Fundamentals | Deploy mesh, mTLS, basic observability, sidecar overhead | Advanced: Chi philosophy → Appendix |
| **8** | Messaging & Events | Kafka, RabbitMQ, NATS basics, simple benchmark | Advanced: Event sourcing patterns → Appendix |
| **9** | Autoscaling & Resources | HPA, VPA, KEDA basics, cluster autoscaling | ✅ Right-sized |
| **10** | Performance & Cost Engineering | Runtime comparison, cost per transaction, basic full-stack | Advanced: Deep profiling → Appendix |

---

## What Changes from Previous Proposal

### Phase 3: Observability (SIMPLIFIED)

**KEEP in Core:**
- ✅ Prometheus + Grafana (metrics)
- ✅ Loki (logging)
- ✅ Tempo (tracing)
- ✅ Basic cost per metric/log/trace
- ✅ TSDB comparison (Prometheus vs VictoriaMetrics)
- ✅ Logging comparison (Loki vs ELK)
- ✅ Tracing comparison (Tempo vs Jaeger)

**MOVE to Appendix T: eBPF & Advanced Metrics** (NEW)
- biosnoop, biotop (block I/O tracing)
- tcptop, tcpretrans (network I/O tracing)
- cachestat, vfsstat (filesystem tracing)
- Pixie (auto-instrumentation)
- Parca (continuous profiling)
- Tetragon (runtime security observability)

**Rationale:** You can learn observability fundamentals with Prometheus/Loki/Tempo. eBPF is the "next level" for deep system introspection.

---

### Phase 7: Service Mesh (SIMPLIFIED)

**KEEP in Core:**
- ✅ Service mesh fundamentals (why mesh?)
- ✅ Deploy Istio, Linkerd, Cilium
- ✅ mTLS basics (identity verification)
- ✅ Basic service-to-service observability
- ✅ Sidecar overhead measurement
- ✅ Mesh comparison (feature set, overhead)

**MOVE to Appendix U: Chi Observability Stack** (NEW)
- Chi energy flow philosophy
- 4-phase lab (Glass Window, Gauge, Valve & Armor, Federation)
- USE Method (Utilization, Saturation, Errors)
- Multi-cluster federation
- Advanced mesh patterns

**Rationale:** You can learn service mesh fundamentals with basic Istio/Linkerd deployment. Chi is an advanced mental model for mastery.

---

## NEW Priority Appendices (Top Tier)

These are the "graduate level" appendices - do these FIRST after core:

### Appendix T: eBPF & Advanced System Metrics ⭐ PRIORITY
**Source:** Was going to be Phase 3.6, now appendix
**When to use:** After Phase 3 core, when you need deep system visibility
**Content:**
- eBPF fundamentals (how it works)
- Block I/O metrics: biosnoop, biotop
- Network I/O metrics: tcptop, tcpretrans, tcplife
- Filesystem metrics: vfsstat, cachestat
- Pixie (auto-instrumented observability)
- Parca (continuous profiling)
- Tetragon (runtime security + performance)
- Integration with Prometheus/Grafana

**Lab:** `ebpf-advanced-metrics` experiment

---

### Appendix U: Chi Observability Stack ⭐ PRIORITY
**Source:** Was going to be Phase 7 enhancement, now appendix
**When to use:** After Phase 7 core, when you want to master service mesh
**Content:**
- Traffic as energy flow philosophy
- Phase 1: Glass Window (Hubble/Pixie visualization)
- Phase 2: Gauge (USE Method dashboards)
- Phase 3: Valve & Armor (Linkerd advanced routing)
- Phase 4: Federation (multi-cluster trust boundaries)

**Labs:**
- `chi-glass-window` (flow visualization)
- `chi-gauge-saturation` (USE Method)
- `chi-valve-smart-routing` (EWMA)
- `chi-armor-identity` (mTLS deep dive)
- `chi-federation` (multi-cluster)

---

### Appendix G: Deployment Strategies ⭐ PRIORITY
**Source:** Phase 5 (moved to appendix as planned)
**When to use:** After Phase 4, before production deployments
**Content:**
- Rolling updates (optimization)
- Blue-green deployments
- Canary with Argo Rollouts
- Feature flags (OpenFeature)
- SLO-based deployment gates

---

### Other Appendices (Do After Priority)

**Appendix H:** gRPC & HTTP/2 Patterns
**Appendix P:** Chaos Engineering
**Appendix Q:** Advanced Workflow Patterns
**Appendix R:** Internal Developer Platforms
**Appendix S:** Web Serving Internals
... (12 existing appendices)

---

## Revised Core Phase Content

### Phase 3: Observability (Core Fundamentals)

**Sub-phases:**
1. Prometheus & Grafana (metrics)
2. TSDB Comparison (Prometheus vs VictoriaMetrics)
3. Loki & Logging (log aggregation)
4. Logging Comparison (Loki vs ELK)
5. Tempo & Distributed Tracing (spans and traces)
6. Tracing Comparison (Tempo vs Jaeger)
7. Pyrra & SLOs (error budgets, multi-burn-rate alerts)
8. Cost Management (cardinality, retention, cost per metric/log/trace)

**Experiments:** 9 total (all fundamentals)

**FinOps:** Cost per metric, cost per GB logs, cost per trace

**What's NOT here:** eBPF, Pixie, Parca, Tetragon → See Appendix T

---

### Phase 7: Service Mesh Fundamentals

**Sub-phases:**
1. Service Mesh Decision Framework (why mesh?)
2. Istio Deployment (control plane, data plane, basic mTLS)
3. Linkerd Deployment (lightweight alternative)
4. Cilium Service Mesh (eBPF-based)
5. Mesh Comparison (features, overhead, complexity)
6. Basic Network Observability (service map, golden signals)

**What you learn:**
- Why service meshes exist
- How to deploy and configure a mesh
- mTLS for service-to-service encryption
- Basic traffic management and observability
- How to measure sidecar overhead
- How to choose between mesh options

**Experiments:**
- `mesh-istio-basics` (deploy Istio, verify mTLS)
- `mesh-linkerd-basics` (deploy Linkerd, compare)
- `mesh-cilium-basics` (deploy Cilium service mesh)
- `mesh-comparison` (overhead benchmark)
- `mesh-observability` (service map, metrics)

**FinOps:** Mesh overhead cost (sidecar tax)

**What's NOT here:**
- Chi energy flow philosophy → Appendix U
- USE Method (saturation metrics) → Appendix U
- Multi-cluster federation → Appendix U
- Advanced routing (EWMA) → Appendix U

---

## Recommended Learning Path

### Minimum Viable (Core Only)
```
Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5 →
Phase 6 → Phase 7 → Phase 8 → Phase 9 → Phase 10

Result: Portfolio-ready, fundamentals mastered
Timeline: 5-6 months
```

### With Priority Appendices (Recommended)
```
Phase 1-2 (Platform + CI/CD)
   ↓
Phase 3 (Observability fundamentals)
   ↓
[Appendix T: eBPF & Advanced Metrics] ← Go deeper on observability
   ↓
Phase 4-6 (Traffic, Data, Security)
   ↓
Phase 7 (Service Mesh fundamentals)
   ↓
[Appendix U: Chi Observability Stack] ← Master service mesh
   ↓
[Appendix G: Deployment Strategies] ← Production patterns
   ↓
Phase 8-10 (Messaging, Autoscaling, Performance)

Result: Portfolio-ready + mastery of key topics
Timeline: 6-7 months
```

### Full Mastery
```
Core 10 phases + All 18 appendices as needed

Result: Subject matter expert level
Timeline: 8-10 months
```

---

## What This Fixes

### Problem with Previous Proposal
- Phase 3 was getting too heavy (Prometheus + Loki + Tempo + eBPF + Pixie + Parca)
- Phase 7 was mixing fundamentals (basic mesh) with philosophy (Chi)
- Hard to know what's "must learn" vs "nice to have"

### Solution
**Core = Fundamentals you need**
- Can complete in 5-6 months
- Portfolio-ready
- Clear what you need to know

**Priority Appendices = Mastery topics**
- Do these AFTER core to level up
- Optional but highly valuable
- Clear progression path

**Other Appendices = Specialization**
- Do as career/project requires
- gRPC for API teams
- Chaos for SRE roles
- Backstage for platform engineering

---

## Revised Impact

| Metric | Before | After (Core) | After (Core + Priority) |
|--------|--------|--------------|------------------------|
| Phases | 16 | 10 | 10 + 3 priority appendices |
| Experiments | 80-90 | 45-50 | 60-65 |
| Timeline | 10-12 months | 5-6 months | 6-7 months |
| Outcome | Overwhelming | Achievable | Mastery |

---

## Updated Appendix Structure

### Priority Appendices (Do These First) ⭐

| Letter | Title | When to Use |
|--------|-------|-------------|
| **T** | eBPF & Advanced System Metrics | After Phase 3, for deep observability |
| **U** | Chi Observability Stack | After Phase 7, for service mesh mastery |
| **G** | Deployment Strategies | After Phase 4, before production |

### Specialized Appendices (As Needed)

| Letter | Title | When to Use |
|--------|-------|-------------|
| **H** | gRPC & HTTP/2 Patterns | For API-heavy architectures |
| **P** | Chaos Engineering | For SRE/reliability focus |
| **Q** | Advanced Workflow Patterns | For CI/CD automation |
| **R** | Internal Developer Platforms | For platform engineering |
| **S** | Web Serving Internals | For performance engineering |

### Reference Appendices (Existing)

A-F, I-O: MLOps, Identity, PKI, Compliance, Distributed Systems, API Design, Containers, Performance, Event-Driven, Databases, SRE, Multi-Cloud, SLSA

---

## Migration from Previous Proposal

| Previous | Revised | Rationale |
|----------|---------|-----------|
| Phase 3 includes eBPF | Phase 3 = basics only, eBPF → Appendix T | Don't overwhelm Phase 3 |
| Phase 7 includes Chi | Phase 7 = basics only, Chi → Appendix U | Fundamentals first, philosophy later |
| 10 core phases | Still 10 core phases | ✅ No change |
| Phase 10 capstone | Still Phase 10 capstone | ✅ No change |
| 18 appendices | 18 appendices (reordered by priority) | Better learning path |

---

## Questions for Approval

1. ✅ or ❌ **10 core phases** focusing on fundamentals?
2. ✅ or ❌ **Move eBPF to Appendix T** (priority appendix after Phase 3)?
3. ✅ or ❌ **Move Chi to Appendix U** (priority appendix after Phase 7)?
4. ✅ or ❌ **Phase 7 covers service mesh fundamentals** (deploy, mTLS, basic observability)?
5. ✅ or ❌ **Recommended path: Core → Priority Appendices → Specialization**?

---

## Summary

**Core (5-6 months):**
- Fundamentals only
- Deployable, measurable, cost-conscious
- Portfolio-ready

**+ Priority Appendices (6-7 months):**
- eBPF for deep system visibility
- Chi for service mesh mastery
- Deployment strategies for production readiness

**+ Specialization (8-10 months):**
- Pick appendices based on role/interest
- gRPC, Chaos, Workflows, IDP, etc.

**This structure makes the core achievable while preserving the deep content as optional mastery topics.**
