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

### 1.2 Hub Cluster & Three-Tier Architecture

**Goal:** Establish a persistent hub cluster and define the three-tier experiment architecture

*The hub hosts shared services. Each experiment gets its own orchestrator cluster (with ArgoCD + Argo Workflows) that deploys to target clusters. This provides complete isolation between experiments.*

**Learning objectives:**
- Understand the three-tier cluster pattern (Hub → Orchestrator → Target)
- Design idempotent, environment-agnostic cluster bootstrap
- Establish GitOps flow across cluster tiers

**Three-Tier Architecture:**
```
┌─────────────────────────────────────────────────────────────────────┐
│  Hub Cluster (always running - your N100, Kind, or cloud)          │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐                   │
│  │   ArgoCD    │ │  OpenBao    │ │  Registry   │                   │
│  │   (root)    │ │  (secrets)  │ │  (Harbor)   │                   │
│  └─────────────┘ └─────────────┘ └─────────────┘                   │
│        │                                                            │
│        │ deploys orchestrator via: task exp:run experiment=X       │
└────────┼────────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Orchestrator Cluster (per-experiment, ephemeral)                   │
│  ┌─────────────┐ ┌─────────────┐                                   │
│  │   ArgoCD    │ │    Argo     │  ← Pulls secrets from Hub OpenBao │
│  │ (exp-local) │ │  Workflows  │  ← Pulls images from Hub Registry │
│  └─────────────┘ └─────────────┘                                   │
│        │                                                            │
│        │ deploys workloads to target(s)                            │
└────────┼────────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Target Cluster(s) (per-experiment, ephemeral)                      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐                   │
│  │   App A     │ │   App B     │ │  Load Gen   │                   │
│  └─────────────┘ └─────────────┘ └─────────────┘                   │
└─────────────────────────────────────────────────────────────────────┘
```

**Experiment Lifecycle:**
```bash
task exp:run experiment=http-baseline
  ├── 1. Create orchestrator cluster (Kind/K3s/cloud)
  ├── 2. Bootstrap ArgoCD on orchestrator (pulls from Hub)
  ├── 3. Orchestrator ArgoCD deploys target cluster(s)
  ├── 4. Orchestrator ArgoCD deploys workloads to targets
  ├── 5. Argo Workflows runs experiment workflow
  ├── 6. Collect results to Hub (MinIO/artifacts)
  └── 7. (optional) task exp:teardown experiment=http-baseline
```

**Hub Core Services (MVP, in dependency order):**
1. **ArgoCD** - Root GitOps, deploys everything else
2. **OpenBao** - Secrets for all tiers (needed before cloudflared)
3. **Cloudflare Tunnel** - Webhook delivery to hub behind NAT (see ADR-003)
4. **Container Registry** - Harbor or distribution/registry

**Hub Extended Services (Later):**
- [ ] **Private CA** - step-ca for internal certificates
- [ ] **Identity Provider** - Keycloak or Dex for SSO
- [ ] **Artifact Storage** - MinIO for results, Helm charts, backups
- [ ] **DNS** - CoreDNS or external-dns for service discovery

**GitOps Flow (webhook-triggered):**
```
git push → GitHub webhook → Cloudflare Tunnel → ArgoCD → sync
```
- GitHub webhook configured to `https://hub.yourdomain.com/api/webhook`
- Cloudflare Tunnel routes webhook to ArgoCD (no inbound firewall needed)
- ArgoCD ApplicationSet auto-discovers experiments in `experiments/scenarios/` directory
- Fallback: ArgoCD polls Git every 3 min if webhook unavailable

**Bootstrap Requirements:**
- [ ] Single command bootstrap: `task hub:bootstrap OVERLAY=<env>`
- [ ] Portable across deployment environments:
  - [ ] Kind - laptop development
  - [ ] K3s - lightweight home lab
  - [ ] Talos - immutable home lab (N100)
  - [ ] AKS/EKS/GKE - cloud-hosted hub (for those without local hardware)
- [ ] Idempotent - can re-run safely
- [ ] GitOps from first deployment (ArgoCD self-manages after bootstrap)
- [ ] Same Git repo, different overlays - only cluster-specific config differs

**Directory Structure:**
```
hub/
├── Taskfile.yaml                      # Convenience tasks (optional)
├── bootstrap/
│   ├── argocd-values-kind.yaml        # ArgoCD + app-of-apps reference
│   ├── argocd-values-talos.yaml
│   └── argocd-values-cloud.yaml
└── app-of-apps/
    ├── kind/                          # Kind-specific services
    │   ├── kustomization.yaml
    │   ├── argocd.yaml                # ArgoCD self-manages
    │   ├── dns-stack.yaml             # CoreDNS + etcd + ExternalDNS
    │   └── values/
    ├── talos/                         # Talos-specific services
    │   ├── kustomization.yaml
    │   ├── argocd.yaml
    │   ├── dns-stack.yaml
    │   ├── metallb.yaml               # LoadBalancer for bare metal
    │   └── values/
    └── cloud/                         # Cloud-specific services
        ├── kustomization.yaml
        ├── argocd.yaml
        ├── external-dns.yaml          # Route53/CloudDNS/Azure DNS
        └── values/

experiments/scenarios/
└── <experiment-name>/
    ├── orchestrator/                  # Per-experiment orchestrator (ephemeral)
    └── target/                        # Target cluster workloads
```

**Adaptor Layer (provides consistent capabilities across environments):**
| Capability | Kind | Talos | Cloud |
|------------|------|-------|-------|
| LoadBalancer | MetalLB | MetalLB | native |
| DNS | k8s_gateway (CoreDNS plugin) | k8s_gateway | ExternalDNS → cloud DNS |

**Bootstrap (one command per environment):**
```bash
# Kind
kind create cluster --name hub
helm install argocd argo/argo-cd -n argocd --create-namespace -f hub/bootstrap/argocd-values-kind.yaml
# ArgoCD deploys: MetalLB → dns-stack → other services (all GitOps)

# Talos (cluster already exists)
helm install argocd argo/argo-cd -n argocd --create-namespace -f hub/bootstrap/argocd-values-talos.yaml

# Cloud (cluster already exists)
helm install argocd argo/argo-cd -n argocd --create-namespace -f hub/bootstrap/argocd-values-cloud.yaml
```

**Tasks (in order):**
1. [x] Create `hub/` directory structure
2. [x] Create ArgoCD bootstrap values with app-of-apps reference
3. [x] Create Kind app-of-apps with ArgoCD self-management
4. [x] Add MetalLB to Kind app-of-apps (LoadBalancer capability)
5. [x] Add dns-stack to Kind app-of-apps (k8s_gateway for DNS)
6. [ ] Test Kind hub bootstrap
6. [ ] Deploy OpenBao via ArgoCD
7. [ ] Configure Cloudflare Tunnel for webhook delivery:
   - [ ] Create tunnel in Cloudflare Zero Trust (manual or Terraform)
   - [ ] Store tunnel token in OpenBao
   - [ ] Deploy cloudflared via ArgoCD (after OpenBao)
   - [ ] Configure GitHub webhook to tunnel URL
8. [ ] Deploy Harbor via ArgoCD
9. [ ] Create ArgoCD ApplicationSet for experiment auto-discovery
10. [ ] Create Talos app-of-apps (add MetalLB)
11. [ ] Create Cloud app-of-apps (ExternalDNS for cloud provider)
12. [ ] **ADR:** Document hub cluster pattern and adaptor layer

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
- [ ] Create `experiments/scenarios/talos-home-lab/`
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

### 1.4 GitLab CI Setup (Cloud IaC)

**Goal:** Establish GitLab CI for Terraform state management and IaC orchestration

*Used when experiments need cloud resources (AKS, EKS, cloud databases, etc.). Dormant until needed.*

**Learning objectives:**
- Understand GitLab CI pipelines for Terraform
- Configure GitLab-managed Terraform state backend
- Set up GitOps workflow for infrastructure changes

**Tasks:**
- [x] Create GitLab account and connect GitHub repo (mirror)
- [x] Create `.gitlab-ci.yml` for Terraform pipelines
- [x] Configure GitLab CI variables for cloud credentials (Azure, AWS)
- [ ] Document GitLab CI workflow patterns
- [x] **ADR:** Document why GitLab CI over Spacelift/Terraform Cloud (see `docs/adrs/ADR-001-gitlab-ci-for-iac.md`)

---

### 1.5 Crossplane Fundamentals

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
- [ ] **ADR:** Document Crossplane vs Terraform for app teams

---

### 1.6 FinOps Foundation & Cost Tagging

**Goal:** Establish cost visibility foundation and tagging strategy

*Foundation only - full FinOps implementation in Phase 9.6 after observability and multi-tenancy.*

**Learning objectives:**
- Understand Kubernetes cost allocation concepts
- Implement resource tagging strategy
- Establish cost attribution foundations

**Tasks:**
- [ ] Create `experiments/scenarios/finops-foundation/`
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

**Goal:** Set up Argo Workflows as the experiment execution engine

*Argo Workflows runs on the orchestrator cluster and executes experiment pipelines. This covers basic setup - advanced patterns come later in Phase 14.*

**Learning objectives:**
- Understand Argo Workflows role in experiment execution
- Deploy and configure Argo Workflows
- Create basic experiment workflow templates

**Tasks:**
- [ ] Deploy Argo Workflows to orchestrator:
  - [ ] Install via Helm chart
  - [ ] Configure artifact repository (MinIO/S3)
  - [ ] Set up RBAC for workflow execution
  - [ ] Configure resource quotas for workflows
- [ ] Basic workflow patterns:
  - [ ] Simple sequential workflow (deploy → test → cleanup)
  - [ ] Parameter passing (target URL, duration, users)
  - [ ] Artifact collection (test results, logs)
  - [ ] Exit handlers for cleanup
- [ ] Create base experiment template:
  - [ ] WorkflowTemplate for standard experiment structure
  - [ ] Parameterized inputs (experiment name, config)
  - [ ] Standard outputs (results location, status)
- [ ] Integration with experiment lifecycle:
  - [ ] Taskfile triggers workflow submission
  - [ ] Wait for completion
  - [ ] Retrieve results
- [ ] Verify with simple test workflow
- [ ] **ADR:** Document Argo Workflows vs Tekton for experiment orchestration

---

