# illm-k8s-ai-lab Roadmap

A learning-focused Kubernetes experiment roadmap for **Cloud Architect**, **Platform Architect**, and **Solutions Architect** roles. Tutorial-style with hands-on exercises; benchmarks come after fundamentals.

## How to Use This Roadmap

- **Sequential but flexible** - Phases build on each other, but skip ahead if you have experience
- **Each phase has experiments** - Found in `experiments/scenarios/<topic>/`
- **ADRs are continuous** - Document decisions as you make them, not as a separate phase (49+ ADRs throughout)
- **FinOps is integrated** - Cost considerations in every relevant phase, not just Phase 1.6 and 11.6
- **Appendices are optional** - Deep dives for specialized needs (MLOps, Compliance)
- **Run experiments manually first** - Automation comes from understanding, not before it
- **AI assistance emerges organically** - We'll add AI tooling based on actual pain points discovered while running experiments

### Cross-Cutting Concerns

| Concern | Where It Appears |
|---------|-----------------|
| **ADRs** | Every phase - document decisions as you make them |
| **FinOps** | Phase 1.6 (foundation), 4.7, 8.5, 9.6, 10.6, 11.6 (full), 15.5 |
| **SLOs/SLAs** | Phase 4.2 (foundation), 7.6 (deployments), 12.4 (chaos), Appendix A.6 (contracts) |
| **Security** | Phase 3 (foundation), 6 (network), plus security sections throughout |
| **Testing** | Phase 2.5 (foundation), plus validation in every experiment |

| | |
|---|---|
| **Target** | ~90 experiments across 16 phases + appendices |
| **Environment** | Kind (local), Talos on N100 (home lab), AKS/EKS (cloud) |
| **Focus** | Portfolio-ready experiments with ADRs |

**Principles:**
- Supply chain security from day one (Phase 2)
- Security foundations before features (Phase 3)
- Build system complexity, then chaos test it (Phase 12)
- Workflow automation after manual understanding (Phase 13)
- ADRs mandatory for technology decisions (continuous, not a phase)
- Runbooks accompany operational components

---

## Phases

| Phase | Topic | Status | Details |
|-------|-------|--------|---------|
| 1 | [Platform Bootstrap & GitOps](docs/roadmap/phase-01-platform-bootstrap.md) | In Progress | Hub, orchestrator, Argo Workflows, Talos, GitLab CI |
| 2 | [CI/CD & Supply Chain](docs/roadmap/phase-02-cicd-supply-chain.md) | Not Started | Image building, scanning, SBOM, signing |
| 3 | [Security Foundations](docs/roadmap/phase-03-security.md) | Not Started | Secrets, RBAC, admission control, policy |
| 4 | [Observability](docs/roadmap/phase-04-observability.md) | Not Started | Metrics, logging, tracing, dashboards |
| 5 | [Traffic Management](docs/roadmap/phase-05-traffic-management.md) | Not Started | Ingress, load balancing, DNS |
| 6 | [Network Security](docs/roadmap/phase-06-network-security.md) | Not Started | Network policies, firewalls, DDoS, WAF |
| 7 | [Deployment Strategies](docs/roadmap/phase-07-deployment-strategies.md) | Not Started | Blue-green, canary, progressive delivery |
| 8 | [Data & Storage](docs/roadmap/phase-08-data-storage.md) | Not Started | Persistent volumes, operators, backup |
| 9 | [Service Mesh](docs/roadmap/phase-09-service-mesh.md) | Not Started | Istio, Linkerd, mTLS, traffic policies |
| 10 | [Messaging & Events](docs/roadmap/phase-10-messaging.md) | Not Started | Kafka, RabbitMQ, NATS, CloudEvents |
| 11 | [Autoscaling](docs/roadmap/phase-11-autoscaling.md) | Not Started | HPA, VPA, KEDA, cluster autoscaling |
| 12 | [Chaos Engineering](docs/roadmap/phase-12-chaos-engineering.md) | Not Started | Fault injection, resilience testing |
| 13 | [Workflow Orchestration](docs/roadmap/phase-13-workflow-orchestration.md) | Not Started | Advanced Argo patterns, events, Tekton |
| 14 | [Developer Experience](docs/roadmap/phase-14-developer-experience.md) | Not Started | Backstage, golden paths, self-service |
| 15 | [Advanced Benchmarks](docs/roadmap/phase-15-advanced-benchmarks.md) | Not Started | Database, messaging, mesh comparisons |

### The Capstone

| Phase | Topic | Status | Details |
|-------|-------|--------|---------|
| 16 | [Web Serving Architecture](docs/roadmap/phase-16-web-serving-finale.md) | Not Started | Threading models, HTTP/2/3, gRPC, GraphQL, proxies, runtimes |

*The crown jewel - after mastering all infrastructure layers, examine what actually serves the traffic.*

---

## Progression

```
Foundation          Traffic & Releases       System Complexity
─────────────────   ──────────────────────   ─────────────────────────
1. Platform         5. Traffic Management    8.  Data & Storage
2. CI/CD            6. Network Security      9.  Service Mesh
3. Security         7. Deployment            10. Messaging
4. Observability                             11. Autoscaling

Validate            Platform Engineering     Synthesis
─────────────────   ──────────────────────   ─────────────────────────
12. Chaos           13. Workflow             15. Advanced Benchmarks
                    14. Developer Experience 16. Web Serving (capstone)
```

---

## Quick Links

- [Learning Path Summary](docs/roadmap/learning-path-summary.md)
- [Notes](docs/roadmap/notes.md)
- [GitOps Patterns](docs/gitops-patterns.md)

### Appendices (Optional Deep Dives)

- [A: MLOps & AI Infrastructure](docs/roadmap/appendix-mlops.md) - Kubeflow, KServe, vector DBs, GPU scheduling
- [B: Identity & Authentication](docs/roadmap/appendix-identity-auth.md) - Password security, JWT/JWE, OAuth flows, OIDC, API keys/PATs, IdP deployment
- [C: PKI & Certificate Management](docs/roadmap/appendix-pki-certs.md) - X.509, TLS, step-ca, cert-manager, mTLS, SPIFFE
- [D: Compliance & Security Operations](docs/roadmap/appendix-compliance-soc.md) - SOC, PCI-DSS, HIPAA/PHI, hardening
- [E: Distributed Systems Fundamentals](docs/roadmap/appendix-distributed-systems.md) - Consensus, CAP, distributed transactions, clocks, replication
- [F: API Design & Contracts](docs/roadmap/appendix-api-design.md) - REST, GraphQL, gRPC, versioning, OpenAPI, contract testing
- [G: Container & Runtime Internals](docs/roadmap/appendix-container-internals.md) - Namespaces, cgroups, OCI, runtimes, security primitives
- [H: Performance Engineering](docs/roadmap/appendix-performance-engineering.md) - Profiling, load testing, capacity planning, latency optimization
- [I: Event-Driven Architecture](docs/roadmap/appendix-event-driven.md) - Event sourcing, CQRS, Saga, outbox pattern, schema evolution
- [J: Database Internals](docs/roadmap/appendix-database-internals.md) - Storage engines, indexing, query optimization, replication, sharding
- [K: SRE Practices & Incident Management](docs/roadmap/appendix-sre-practices.md) - SLOs, on-call, incident response, post-mortems, toil reduction
- [L: Multi-Cloud & Disaster Recovery](docs/roadmap/appendix-multicloud-dr.md) - Multi-cloud strategy, DR planning, failover, geographic distribution

---

## Current Focus

**Phase 1: Platform Bootstrap**
- [x] Create `hub/` directory structure
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
