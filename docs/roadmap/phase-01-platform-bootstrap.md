## Phase 1: Platform Bootstrap & GitOps Foundation

*Establish the hub cluster and GitOps foundation. The hub provides persistent services (secrets, registry, GitOps) that experiments consume.*

### 1.1 Document Current GitOps Patterns

**Goal:** Capture and understand the existing GitOps architecture before building new experiments

**Learning objectives:**
- Understand app-of-apps pattern implementation
- Document multi-source Helm + Git integration
- Map the components structure

**Current Patterns to Document:**
- [x] **App-of-Apps Hierarchy:**
  - [x] Core app-of-apps (`core-app-of-apps.yaml`) managing platform components
  - [x] Stack applications (ELK, Loki stacks)
  - [x] Experiment-specific applications
- [x] **Multi-Source Applications:**
  - [x] Helm chart + Git values file pattern
  - [x] Directory-based selective sync (include/exclude patterns)
  - [x] `$values` reference for external values files
- [x] **Sync Strategies:**
  - [x] Sync wave ordering for dependencies
  - [x] Retry policies with exponential backoff
  - [x] ignoreDifferences for CRDs and webhooks
  - [x] ServerSideApply and RespectIgnoreDifferences
- [x] **Experiment GitOps Pattern:**
  - [x] Per-experiment ArgoCD Applications
  - [x] Label-based organization (experiment, cluster)
  - [x] Workflow integration with ArgoCD sync
- [x] Create `docs/gitops-patterns.md` documenting all patterns
- [x] Create architecture diagram of app-of-apps hierarchy

---

### 1.2 Hub Cluster & Experiment Architecture

**Goal:** Establish a persistent hub cluster that provisions experiment infrastructure

*The hub is always local (Kind for dev, Talos for home lab) - zero cloud cost for persistent infrastructure. The hub provisions experiment clusters (including an orchestrator) in parallel via Crossplane or Kind, then the orchestrator runs the experiment.*

**Learning objectives:**
- Understand the hub → experiment cluster pattern
- Design idempotent, environment-agnostic cluster bootstrap
- Establish GitOps flow for experiment provisioning

**Architecture:**
```
┌─────────────────────────────────────────────────────────────────────┐
│  Hub Cluster (always local: Kind for dev, Talos for home lab)      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐   │
│  │   ArgoCD    │ │    Argo     │ │  Crossplane │ │  OpenBao    │   │
│  │   (root)    │ │  Workflows  │ │ (provisions)│ │  (secrets)  │   │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘   │
│        │                               │                            │
│        │ task hub:up                  │ provisions in PARALLEL     │
└────────┼───────────────────────────────┼────────────────────────────┘
         │                               │
         │         ┌─────────────────────┼─────────────────────┐
         │         ▼                     ▼                     ▼
         │  ┌─────────────┐       ┌─────────────┐       ┌─────────────┐
         │  │ Orchestrator│       │   Target    │       │   Loadgen   │
         │  │ (ArgoCD +   │       │  Cluster    │       │   Cluster   │
         │  │  Workflows) │       │             │       │             │
         │  └──────┬──────┘       └─────────────┘       └─────────────┘
         │         │                     ▲                     ▲
         │         │    deploys apps     │                     │
         │         └─────────────────────┴─────────────────────┘
         │
         └── orchestrator runs experiment workflow
```

**Experiment Lifecycle:**
```bash
task hub:up -- http-baseline
  ├── 1. Hub provisions ALL clusters in PARALLEL:
  │     ├── http-baseline-orchestrator (Kind or cloud via Crossplane)
  │     ├── http-baseline-target
  │     └── http-baseline-loadgen
  ├── 2. Hub bootstraps ArgoCD + Argo Workflows on orchestrator
  ├── 3. Hub registers target/loadgen with orchestrator's ArgoCD
  ├── 4. Orchestrator's ArgoCD deploys workloads to targets
  ├── 5. Orchestrator's Argo Workflows runs experiment
  ├── 6. Results collected to Hub (future: MinIO)
  └── 7. Cleanup: orchestrator + targets destroyed, hub remains
```

**Hub Core Services (MVP, in dependency order):**
1. **ArgoCD** - Root GitOps, deploys everything else
2. **Argo Workflows** - Experiment orchestration
3. **OpenBao** - Secrets for all tiers
4. **Crossplane** - Provisions cloud resources for experiments
5. **Cloudflare Tunnel** - Webhook delivery to hub behind NAT (see ADR-003)

**Hub Extended Services (Later):**
- [ ] **Container Registry** - Harbor for experiment images
- [ ] **Private CA** - step-ca for internal certificates (see [Appendix: PKI & Certs](appendix-pki-certs.md))
- [ ] **Identity Provider** - Keycloak or Dex for SSO (see [Appendix: Identity & Auth](appendix-identity-auth.md))
- [ ] **Artifact Storage** - MinIO for results, Helm charts, backups

**GitOps Flow (webhook-triggered):**
```
git push → GitHub webhook → Cloudflare Tunnel → ArgoCD → sync
```
- GitHub webhook configured to `https://hub.yourdomain.com/api/webhook`
- Cloudflare Tunnel routes webhook to ArgoCD (no inbound firewall needed)
- ArgoCD ApplicationSet auto-discovers experiments in `experiments/` directory
- Fallback: ArgoCD polls Git every 3 min if webhook unavailable

**Hub Environments:**

| Environment | Purpose | Provisions Targets Via |
|-------------|---------|------------------------|
| Kind hub | Dev/test for Talos hub | Kind clusters (local) |
| Talos hub | Production home lab | Kind, Crossplane (cloud) |

*No cloud hub - the hub is always local (zero cloud cost for persistent infra).*

**Directory Structure:**
```
platform/
├── hub/                               # Control plane (runs on Kind locally)
│   ├── apps/                          # ArgoCD Applications
│   │   ├── argocd.yaml                # ArgoCD self-manages
│   │   ├── argo-workflows.yaml        # Experiment orchestration
│   │   ├── metallb.yaml               # LoadBalancer
│   │   └── dns-stack.yaml             # k8s_gateway for DNS
│   ├── values/                        # Helm values for hub components
│   ├── manifests/                     # Raw K8s manifests
│   ├── cluster/                       # Kind cluster provisioning
│   │   └── Taskfile.yaml              # task hub:bootstrap, hub:up, hub:tutorial, etc.
│   └── bootstrap/                     # Initial ArgoCD setup
│       ├── argocd-values-kind.yaml    # ArgoCD + app-of-apps reference
│       └── hub-application.yaml       # Root app-of-apps
└── targets/                           # Workload clusters managed by hub
    └── talos/                         # Home lab (N100 hardware)
        └── cluster/                   # Talos provisioning

experiments/<experiment-name>/
├── orchestrator/                      # Orchestrator cluster config
│   ├── cluster.yaml                   # Cluster definition
│   └── argocd/                        # Apps for orchestrator
├── target/                            # Target cluster config
│   ├── cluster.yaml
│   └── argocd/
└── workflow/
    └── experiment.yaml                # Argo Workflow (runs on orchestrator)
```

**Bootstrap (one command per environment):**
```bash
# Kind hub (dev)
task hub:bootstrap

# Talos hub (future)
task talos:bootstrap
```

**Tasks (in order):**
1. [x] Create `platform/` directory structure
2. [x] Create ArgoCD bootstrap values with app-of-apps reference
3. [x] Create Kind app-of-apps with ArgoCD self-management
4. [x] Add MetalLB to Kind app-of-apps
5. [x] Add dns-stack to Kind app-of-apps
6. [x] Add Argo Workflows to Kind app-of-apps
7. [ ] Test Kind hub bootstrap end-to-end
8. [ ] Deploy OpenBao via ArgoCD
9. [ ] Deploy Crossplane via ArgoCD
10. [ ] Configure Cloudflare Tunnel for webhook delivery
11. [ ] Create Talos app-of-apps
12. [ ] **ADR:** Document hub cluster pattern

---

### 1.3 Home Lab Cluster with Talos Linux

**Goal:** Deploy the hub cluster on bare metal using Talos Linux

*The N100 becomes your production hub cluster. Immutable OS, declarative config, production patterns.*

**Learning objectives:**
- Understand Talos Linux as an immutable, secure Kubernetes OS
- Manage infrastructure declaratively (no SSH, API-driven)
- Use Ansible for initial provisioning/PXE boot setup
- Practice GitOps for both OS and cluster configuration

**Hardware:**
- [x] GMKtec NucBox G3 (ordered)
  - [x] Intel N100 (4C/4T, up to 3.4GHz, 6W TDP)
  - [x] 16GB DDR4 RAM, 512GB NVMe SSD
  - [x] 2.5GbE ethernet, WiFi 6, BT 5.2
- [ ] Network switch (5-port gigabit, ~$20) - optional for single node
- [ ] Ethernet cable

**Tasks:**
- [ ] Create `experiments/talos-home-lab/`
- [ ] Ansible for initial setup:
  - [ ] Inventory file for home lab nodes
  - [ ] Playbook to prepare USB/PXE boot media
  - [ ] Talos image customization (extensions, config)
  - [ ] Network boot infrastructure (optional, for multi-node)
- [ ] Talos Linux fundamentals:
  - [ ] Understand machine config vs cluster config
  - [ ] Generate configs with `talosctl gen config`
  - [ ] Apply configs with `talosctl apply-config`
  - [ ] Bootstrap cluster with `talosctl bootstrap`
  - [ ] No SSH - all management via talosctl API
- [ ] Talos configuration as code:
  - [ ] Machine configs in Git
  - [ ] Patch files for node-specific overrides
  - [ ] Encrypted secrets with sops/age
- [ ] Networking:
  - [ ] Static IPs via machine config
  - [ ] MetalLB for LoadBalancer services
  - [ ] Cilium CNI (default in Talos)
  - [ ] Ingress controller (Contour)
- [ ] Storage:
  - [ ] Local path provisioner
  - [ ] Mayastor or OpenEBS (optional, for multi-node)
- [ ] ArgoCD bootstrap:
  - [ ] Deploy ArgoCD to Talos cluster
  - [ ] Connect to same Git repo as Kind
  - [ ] Multi-cluster GitOps pattern
- [ ] Cluster operations:
  - [ ] Upgrade Talos OS declaratively
  - [ ] Upgrade Kubernetes version
  - [ ] Add worker nodes
  - [ ] Backup and restore etcd
- [ ] Document Talos patterns and home lab architecture
- [ ] **ADR:** Document Talos vs K3s vs kubeadm for bare metal
- [ ] **ADR:** Document Ansible role in Talos workflow

---

### 1.4 Crossplane Fundamentals

**Goal:** Master Crossplane for cloud resource provisioning

*Crossplane runs on the hub cluster and provisions cloud resources for experiments.*

**Learning objectives:**
- Understand Crossplane architecture (providers, XRDs, compositions)
- Create and use Composite Resource Definitions
- Build reusable compositions for common patterns

**Tasks:**
- [ ] Deploy Crossplane to hub cluster
- [ ] Install AWS and Azure providers (credentials from OpenBao via Phase 3)
- [ ] Verify existing XRDs: Database, ObjectStorage, Queue, Cache
- [ ] Test claims provision real cloud resources
- [ ] Document XRD authoring patterns
- [ ] **ADR:** Document Crossplane abstraction patterns for app teams

---

### 1.5 FinOps Foundation & Cost Tagging

**Goal:** Establish cost visibility foundation and tagging strategy

*Foundation only - full FinOps implementation in Phase 9.6 after observability and multi-tenancy.*

**Learning objectives:**
- Understand Kubernetes cost allocation concepts
- Implement resource tagging strategy
- Establish cost attribution foundations

**Tasks:**
- [ ] Create `experiments/finops-foundation/`
- [ ] Define tagging strategy:
  - [ ] Required labels: `team`, `project`, `environment`, `cost-center`
  - [ ] Document label standards
  - [ ] Create label validation policy (enforced in Phase 3.5 Policy & Governance)
- [ ] Implement Crossplane cost tags:
  - [ ] Azure resource tags via compositions
  - [ ] AWS resource tags via compositions
  - [ ] Tag inheritance patterns
- [ ] Deploy OpenCost (lightweight):
  - [ ] Basic namespace-level cost visibility
  - [ ] Understand cost model fundamentals
- [ ] Document cost allocation strategy
- [ ] **ADR:** Document cost tagging and allocation approach

---

### 1.7 Argo Workflows Fundamentals

**Goal:** Set up Argo Workflows for experiment orchestration

*Argo Workflows runs on both the hub (for provisioning) and orchestrator (for experiment execution). This section covers hub deployment - orchestrator bootstrap happens during `hub:up`.*

**Learning objectives:**
- Understand Argo Workflows role in experiment lifecycle
- Deploy and configure Argo Workflows on hub
- Create workflow templates for experiment provisioning

**Argo Workflows in the Architecture:**

| Location | Purpose |
|----------|---------|
| Hub | Provisions experiment clusters, bootstraps orchestrator |
| Orchestrator | Runs actual experiment workflow (k6, etc.) |

**Tasks:**
- [x] Deploy Argo Workflows to hub via ArgoCD
- [ ] Basic workflow patterns:
  - [ ] Parameter passing between steps
  - [ ] Exit handlers for cleanup
  - [ ] Retry policies
- [ ] Hub provisioning workflow:
  - [ ] Create Kind clusters in parallel
  - [ ] Bootstrap ArgoCD on orchestrator
  - [ ] Register targets with orchestrator
  - [ ] Trigger orchestrator workflow
- [ ] Orchestrator experiment workflow:
  - [ ] Wait for apps to sync
  - [ ] Run load test (k6)
  - [ ] Collect results
- [ ] Integration with `task hub:up`:
  - [ ] Submit workflow to hub
  - [ ] Wait for completion
  - [ ] Report results
- [ ] Verify with simple test workflow
- [ ] **ADR:** Document Argo Workflows vs Tekton for experiment orchestration

---

