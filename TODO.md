# illm-k8s-lab TODO

## Overview

A learning-focused Kubernetes experiment roadmap for **Cloud Architect**, **Platform Architect**, and **Solutions Architect** roles. Tutorial-style with hands-on exercises; benchmarks come after fundamentals.

| | |
|---|---|
| **Target** | ~73 experiments across 16 phases |
| **Environment** | Kind (local), Raspberry Pi (home lab), AKS/EKS (cloud) |
| **Focus** | Portfolio-ready experiments with ADRs |

**Principles:**
- Supply chain security from day one (Phase 2)
- Security foundations before features (Phase 3)
- Tutorials first, benchmarks later (Phase 12)
- Workflow automation last, informed by manual runs (Phase 14)
- ADRs mandatory for technology decisions
- Runbooks accompany operational components

---

## Phase 1: Platform Bootstrap & GitOps Foundation

*Get the multi-cloud GitOps foundation running. Document existing patterns first, then build on them.*

### 1.1 Document Current GitOps Patterns

**Goal:** Capture and understand the existing GitOps architecture before building new experiments

**Learning objectives:**
- Understand app-of-apps pattern implementation
- Document multi-source Helm + Git integration
- Map the workload-catalog structure

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

### 1.2 Raspberry Pi Cluster & Ansible

**Goal:** Build a home lab Kubernetes cluster on Raspberry Pi, provisioned with Ansible

*Bridge between Kind (single-node dev) and cloud (production). Real hardware, real networking, real constraints.*

**Learning objectives:**
- Understand Ansible for bare-metal/VM provisioning
- Configure multi-node Kubernetes on physical hardware
- Handle real networking (DHCP, DNS, load balancing)
- Practice Day 2 operations on physical infrastructure

**Tasks:**
- [ ] Create `experiments/raspberry-pi-cluster/`
- [ ] Hardware setup:
  - [ ] 3+ Raspberry Pi 4/5 nodes (1 control plane, 2+ workers)
  - [ ] Network switch, power supply
  - [ ] SD cards or USB SSDs
- [ ] Ansible fundamentals:
  - [ ] Inventory file (static and dynamic)
  - [ ] Playbooks for base OS configuration
  - [ ] Roles for reusable configuration (k8s-common, k8s-control-plane, k8s-worker)
  - [ ] Variables and templates (Jinja2)
  - [ ] Handlers for service restarts
  - [ ] Vault for secrets (ansible-vault)
- [ ] OS provisioning:
  - [ ] Ubuntu Server or Raspberry Pi OS (64-bit)
  - [ ] SSH key distribution
  - [ ] Hostname, timezone, locale configuration
  - [ ] Package updates and hardening
- [ ] Kubernetes installation:
  - [ ] Container runtime (containerd)
  - [ ] kubeadm, kubelet, kubectl
  - [ ] Control plane initialization
  - [ ] Worker node join
  - [ ] CNI installation (Cilium or Calico)
- [ ] Networking:
  - [ ] Static IPs or DHCP reservations
  - [ ] MetalLB for LoadBalancer services
  - [ ] Local DNS (Pi-hole or CoreDNS)
  - [ ] Ingress controller (Contour/nginx)
- [ ] Storage:
  - [ ] Local path provisioner
  - [ ] NFS for shared storage (optional)
- [ ] ArgoCD bootstrap:
  - [ ] Deploy ArgoCD to Pi cluster
  - [ ] Connect to same Git repo as Kind
  - [ ] Multi-cluster GitOps pattern
- [ ] Document Ansible patterns and Pi cluster architecture
- [ ] **ADR:** Document Ansible vs other config management (Puppet, Chef, Salt)

---

### 1.3 Spacelift Setup (Cloud IaC)

**Goal:** Establish Spacelift for Terraform state management and IaC orchestration

**Learning objectives:**
- Understand Spacelift stacks, contexts, and policies
- Configure Kind for multi-cluster experiments
- Set up GitOps workflow for infrastructure changes

**Tasks:**
- [ ] Create Spacelift account and connect GitHub repo
- [ ] Create `spacelift-root` stack (administrative=true)
- [ ] Configure cloud credential contexts (Azure, AWS) in Spacelift UI
- [ ] Set up Kind cluster with sufficient resources
- [ ] Verify Spacelift can plan/apply to local state
- [ ] Document Spacelift workflow patterns
- [x] **ADR:** Document why Spacelift over Terraform Cloud/Atlantis (see `docs/adrs/ADR-001-spacelift-for-iac-orchestration.md`)

---

### 1.4 Crossplane Fundamentals

**Goal:** Master Crossplane for cloud resource provisioning

**Learning objectives:**
- Understand Crossplane architecture (providers, XRDs, compositions)
- Create and use Composite Resource Definitions
- Build reusable compositions for common patterns

**Tasks:**
- [ ] Deploy Crossplane to Kind cluster
- [ ] Install AWS and Azure providers (credentials via ESO from Phase 3)
- [ ] Verify existing XRDs: Database, ObjectStorage, Queue, Cache
- [ ] Test claims provision real cloud resources
- [ ] Document XRD authoring patterns
- [ ] **ADR:** Document Crossplane vs Terraform for app teams

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

## Phase 2: CI/CD & Supply Chain Security

*You need to build and deploy images before anything else. Supply chain security is baked in from day one.*

### 2.1 CI/CD Pipeline Fundamentals

**Goal:** Establish image building and registry workflows

**Learning objectives:**
- Understand CI/CD pipeline patterns for Kubernetes
- Configure container registry workflows
- Implement GitOps image update patterns

**Tasks:**
- [ ] Create `experiments/cicd-fundamentals/`
- [ ] GitHub Actions pipeline:
  - [ ] Build multi-arch container images
  - [ ] Push to GitHub Container Registry (GHCR)
  - [ ] Tag strategies (semver, git SHA, branch)
  - [ ] Cache optimization (layer caching, buildx)
- [ ] GitLab CI comparison:
  - [ ] Create `.gitlab-ci.yml` equivalent
  - [ ] Deploy GitLab Runner on Kubernetes (Helm chart)
  - [ ] Compare GitLab Registry vs GHCR
- [ ] ArgoCD Image Updater:
  - [ ] Automatic image tag updates
  - [ ] Write-back strategies (Git vs annotation)
  - [ ] Semver constraints
- [ ] Document CI/CD pipeline patterns
- [ ] **ADR:** Document CI platform selection (GitHub Actions vs GitLab CI)

---

### 2.2 Container Image Security

**Goal:** Secure the container image supply chain

**Learning objectives:**
- Understand container image vulnerabilities
- Implement scanning and policy gates
- Generate and verify SBOMs

**Tasks:**
- [ ] Create `experiments/image-security/`
- [ ] Image scanning with Trivy:
  - [ ] Integrate into CI pipeline
  - [ ] Vulnerability severity thresholds (block on critical/high)
  - [ ] Scan base images vs application layers
  - [ ] Secret detection in images
- [ ] SBOM generation:
  - [ ] Generate SBOMs with Syft
  - [ ] SPDX vs CycloneDX formats
  - [ ] Attach SBOMs to images (OCI artifacts)
  - [ ] SBOM storage and retrieval
- [ ] Base image management:
  - [ ] Renovate/Dependabot for base image updates
  - [ ] Distroless and minimal base images
  - [ ] Multi-stage build patterns
- [ ] Document image security patterns

---

### 2.3 Image Signing & Verification

**Goal:** Implement cryptographic image verification

**Learning objectives:**
- Understand Sigstore ecosystem (Cosign, Fulcio, Rekor)
- Sign images in CI pipelines
- Verify signatures at admission

**Tasks:**
- [ ] Create `experiments/image-signing/`
- [ ] Image signing with Cosign:
  - [ ] Keyless signing (OIDC/Fulcio)
  - [ ] Key-based signing (for air-gapped)
  - [ ] Sign images in GitHub Actions
  - [ ] Transparency log (Rekor) integration
- [ ] Attestation creation:
  - [ ] SLSA provenance attestations
  - [ ] Vulnerability scan attestations
  - [ ] SBOM attestations
- [ ] Admission verification:
  - [ ] Kyverno image verification policies
  - [ ] Require signatures from trusted keys
  - [ ] Verify attestations at admission
  - [ ] Policy exceptions for system images
- [ ] SLSA compliance:
  - [ ] Understand SLSA levels (1-4)
  - [ ] Implement SLSA Level 2+ pipeline
  - [ ] Document provenance chain
- [ ] Document signing and verification patterns
- [ ] **ADR:** Document supply chain security strategy

---

### 2.4 Registry & Artifact Management

**Goal:** Manage container registries and OCI artifacts

**Learning objectives:**
- Understand OCI registry concepts
- Implement registry security and policies
- Manage Helm charts as OCI artifacts

**Tasks:**
- [ ] Create `experiments/registry-management/`
- [ ] Registry options:
  - [ ] GHCR configuration and access
  - [ ] Harbor deployment (self-hosted option)
  - [ ] Azure Container Registry / ECR via Crossplane
- [ ] Registry security:
  - [ ] Image pull secrets management (via ESO)
  - [ ] Registry allowlists (Kyverno/OPA)
  - [ ] Content trust policies
- [ ] OCI artifacts:
  - [ ] Helm charts in OCI registries
  - [ ] ArgoCD with OCI Helm charts
  - [ ] Policy bundles as OCI artifacts
- [ ] Image lifecycle:
  - [ ] Tag retention policies
  - [ ] Garbage collection
  - [ ] Image promotion workflows (dev → staging → prod)
- [ ] Document registry patterns

---

## Phase 3: Security Foundations

*Security first - TLS, certificates, identity, secrets, and policies are prerequisites for everything else.*

### 3.1 Bootstrap Credentials & External Secrets

**Goal:** Establish credential management foundation for cloud deployments

*This phase unblocks all cloud-dependent work by setting up External Secrets Operator to sync credentials from cloud secret managers.*

**Learning objectives:**
- Understand External Secrets Operator architecture
- Configure cloud secret store providers (Azure Key Vault, AWS Secrets Manager)
- Bootstrap credential flow for Crossplane and Spacelift

**Tasks:**
- [ ] Store bootstrap credentials in cloud secret managers:
  - [ ] Azure Key Vault: Create vault, store SP credentials (ARM_CLIENT_ID, etc.)
  - [ ] AWS Secrets Manager: Store IAM credentials (AWS_ACCESS_KEY_ID, etc.)
- [ ] Deploy External Secrets Operator via ArgoCD
- [ ] Create ClusterSecretStore for Azure Key Vault
- [ ] Create ClusterSecretStore for AWS Secrets Manager
- [ ] Create ExternalSecrets for Crossplane provider credentials
- [ ] Verify secrets sync to `crossplane-system` namespace
- [ ] Document bootstrap credential flow

---

### 3.2 Vault & Secrets Management

**Goal:** Centralized secrets management with Vault + External Secrets Operator

*Vault becomes the central secrets store, with ESO providing GitOps-friendly secret synchronization.*

**Learning objectives:**
- Understand Vault architecture (secrets engines, auth methods, policies)
- Integrate Vault with External Secrets Operator
- Establish secrets management patterns for experiments

**Tasks:**
- [ ] Deploy Vault via ArgoCD (already in workload-catalog)
- [ ] Configure Kubernetes auth method
- [ ] Create policies for experiments (per-namespace isolation)
- [ ] Create ClusterSecretStore pointing to Vault
- [ ] Migrate bootstrap secrets to Vault (optional, for consolidation)
- [ ] Create example ExternalSecret using Vault backend
- [ ] Document Vault + ESO patterns
- [ ] **ADR:** Document secrets management architecture (see `docs/adrs/ADR-002-secrets-management.md`)

---

### 3.3 cert-manager & TLS Automation

**Goal:** Automate TLS certificate lifecycle in Kubernetes

**Learning objectives:**
- Understand PKI fundamentals and certificate lifecycle
- Configure cert-manager issuers (self-signed, ACME, private CA)
- Automate certificate renewal and monitoring

**Tasks:**
- [ ] Create `experiments/cert-manager-tutorial/`
- [ ] Deploy cert-manager via ArgoCD
- [ ] Configure Issuers:
  - [ ] SelfSigned (for development)
  - [ ] Let's Encrypt staging (ACME HTTP-01)
  - [ ] Let's Encrypt production
  - [ ] Private CA (for internal mTLS)
- [ ] Create Certificate resources:
  - [ ] Ingress/Gateway TLS termination
  - [ ] Wildcard certificate
  - [ ] Short-lived certificate (test renewal)
- [ ] Implement DNS-01 challenge (Azure DNS or Route53 via Crossplane)
- [ ] Set up certificate expiry alerting
- [ ] Test failure scenarios (issuer down, challenge failure)
- [ ] Document certificate patterns for different use cases

---

### 3.4 Advanced Vault Patterns

**Goal:** Dynamic credentials, PKI, and advanced secret injection patterns

*Builds on Phase 3.2 (Vault + ESO basics) with production-grade patterns.*

**Learning objectives:**
- Implement dynamic database credentials
- Configure PKI secrets engine for certificate generation
- Compare secret injection methods (Agent vs CSI vs ESO)

**Tasks:**
- [ ] Create `experiments/vault-advanced/`
- [ ] Configure advanced auth methods:
  - [ ] AppRole (for CI/CD pipelines)
  - [ ] JWT/OIDC (for external identity providers)
- [ ] Set up dynamic secrets engines:
  - [ ] Database engine (dynamic PostgreSQL creds with TTL)
  - [ ] PKI engine (dynamic certificates for mTLS)
- [ ] Implement secret injection comparison:
  - [ ] Vault Agent Sidecar (file-based injection)
  - [ ] Vault CSI Provider (volume-based injection)
  - [ ] External Secrets Operator (Kubernetes Secret sync)
- [ ] Test secret rotation workflows:
  - [ ] Automatic credential rotation
  - [ ] Application restart-free rotation
- [ ] Implement audit logging and monitoring
- [ ] Configure Vault HA (Raft storage) for production
- [ ] Document injection method trade-offs
- [ ] **ADR:** Compare Vault injection methods (Agent vs CSI vs ESO)

---

### 3.5 Policy & Governance

**Goal:** Implement policy-as-code for compliance and operational guardrails

**Learning objectives:**
- Understand admission controllers and policy engines
- Implement organizational policies at scale
- Create audit trails for compliance

**Tasks:**
- [ ] Create `experiments/policy-governance-tutorial/`
- [ ] Deploy policy engine:
  - [ ] Kyverno (Kubernetes-native) OR
  - [ ] OPA Gatekeeper (Rego-based)
- [ ] Implement policy categories:
  - [ ] **Security policies:**
    - [ ] Require non-root containers
    - [ ] Disallow privileged pods
    - [ ] Enforce resource limits
    - [ ] Require specific labels (owner, team, cost-center)
  - [ ] **Supply chain policies (from Phase 2):**
    - [ ] Require signed images
    - [ ] Verify image attestations
    - [ ] Restrict image registries (allowlist)
  - [ ] **Operational policies:**
    - [ ] Require probes (liveness, readiness)
    - [ ] Enforce image pull policy
    - [ ] Require PodDisruptionBudgets for production
  - [ ] **Networking policies:**
    - [ ] Require NetworkPolicy for namespaces
    - [ ] Restrict LoadBalancer services
    - [ ] Enforce ingress annotations
- [ ] Policy lifecycle:
  - [ ] Audit mode vs enforce mode
  - [ ] Policy exceptions and exemptions
  - [ ] Policy versioning in Git
  - [ ] Policy testing (CI validation)
- [ ] Compliance reporting:
  - [ ] Policy violation dashboards
  - [ ] Audit log integration
  - [ ] Compliance score tracking
- [ ] Multi-tenancy policies:
  - [ ] Namespace quotas enforcement
  - [ ] Cross-namespace restrictions
  - [ ] Tenant isolation validation
- [ ] Document policy patterns
- [ ] **ADR:** Document policy engine selection (Kyverno vs OPA)

---

### 3.6 Network Policies & Pod Security

**Goal:** Implement defense-in-depth with network segmentation and pod security

**Learning objectives:**
- Write effective NetworkPolicy resources
- Understand Pod Security Standards (PSS)
- Implement least-privilege pod configurations

**Tasks:**
- [ ] Create `experiments/network-security-tutorial/`
- [ ] Deploy Calico or Cilium CNI (for NetworkPolicy support)
- [ ] Implement NetworkPolicy patterns:
  - [ ] Default deny all ingress/egress
  - [ ] Allow specific service-to-service communication
  - [ ] Allow egress to specific external CIDRs
  - [ ] Namespace isolation
- [ ] Configure Pod Security:
  - [ ] Pod Security Admission (PSA) labels
  - [ ] Restricted security context
  - [ ] Read-only root filesystem
  - [ ] Non-root containers
- [ ] Test policy enforcement (verify blocked traffic)
- [ ] Document security baseline for all experiments

---

### 3.7 Identity & Access Management

**Goal:** Integrate external identity providers with Kubernetes RBAC

**Learning objectives:**
- Understand OIDC authentication flow
- Configure Kubernetes API server for external IdP
- Implement just-in-time access patterns

**Tasks:**
- [ ] Create `experiments/identity-tutorial/`
- [ ] Deploy identity provider:
  - [ ] **Auth0** (work requirement) OR
  - [ ] Keycloak (self-hosted alternative)
  - [ ] Azure AD / Entra ID integration
- [ ] Configure OIDC authentication:
  - [ ] API server OIDC flags (Kind/EKS/AKS)
  - [ ] kubectl OIDC plugin (kubelogin)
  - [ ] Group claims mapping
- [ ] Implement RBAC patterns:
  - [ ] ClusterRole/Role definitions
  - [ ] RoleBinding to OIDC groups
  - [ ] Namespace-scoped access
  - [ ] Read-only vs admin personas
- [ ] Service account best practices:
  - [ ] Workload identity (Azure/AWS IAM)
  - [ ] Token projection and audiences
  - [ ] Cross-namespace service account access
- [ ] Implement audit logging:
  - [ ] API server audit policy
  - [ ] Who did what, when
- [ ] Document identity patterns and onboarding workflow
- [ ] **ADR:** Document identity federation architecture

---

### 3.8 Multi-Tenancy Security Foundations

**Goal:** Establish tenant isolation patterns using security primitives

*This covers security isolation - production-scale multi-tenancy (quotas, scheduling) comes in Phase 9.5*

**Learning objectives:**
- Implement namespace-based tenant isolation
- Apply policies for tenant boundaries
- Integrate RBAC with tenant model

**Tasks:**
- [ ] Create `experiments/multi-tenancy-security/`
- [ ] Namespace isolation model:
  - [ ] Create tenant namespaces (tenant-a, tenant-b)
  - [ ] Apply default NetworkPolicies (deny all, allow within tenant)
  - [ ] Implement namespace labels for tenant identification
- [ ] RBAC for tenants (using Phase 3.7 identity):
  - [ ] Map OIDC groups to tenant namespaces
  - [ ] Tenant-admin vs tenant-developer roles
  - [ ] Cross-tenant access restrictions
- [ ] Policy enforcement (using Phase 3.5 Kyverno/OPA):
  - [ ] Prevent cross-namespace resource references
  - [ ] Enforce tenant labels on all resources
  - [ ] Restrict cluster-scoped resource creation
  - [ ] Validate resource names include tenant prefix
- [ ] Secrets isolation:
  - [ ] Vault namespaces per tenant (if using Vault)
  - [ ] Prevent secret access across tenants
- [ ] ArgoCD tenant isolation:
  - [ ] ArgoCD Projects per tenant
  - [ ] Repository restrictions per project
  - [ ] Destination namespace restrictions
- [ ] Validation testing:
  - [ ] Verify tenant A cannot access tenant B resources
  - [ ] Verify network isolation works
  - [ ] Verify RBAC boundaries hold
- [ ] Document tenant security patterns

---

## Phase 4: Observability Stack

*You need to see what's happening before you can improve it. These skills are used in every subsequent experiment.*

### 4.1 Prometheus & Grafana Deep Dive

**Goal:** Master metrics collection, PromQL, alerting, and dashboards

**Learning objectives:**
- Understand Prometheus architecture (scraping, TSDB, federation)
- Write effective PromQL queries
- Build actionable Grafana dashboards
- Configure alerting pipelines

**Tasks:**
- [ ] Create `experiments/prometheus-tutorial/`
- [ ] Deploy kube-prometheus-stack via ArgoCD
- [ ] Build sample app with custom metrics:
  - [ ] Counter (http_requests_total)
  - [ ] Gauge (active_connections)
  - [ ] Histogram (request_duration_seconds)
  - [ ] Summary (response_size_bytes)
- [ ] Create ServiceMonitor for scrape discovery
- [ ] Write PromQL tutorial queries:
  - [ ] rate() and irate() for counters
  - [ ] Aggregations (sum, avg, max by labels)
  - [ ] histogram_quantile() for percentiles
  - [ ] absent() for missing metric alerts
  - [ ] predict_linear() for capacity planning
- [ ] Build Grafana dashboards:
  - [ ] RED metrics (Rate, Errors, Duration)
  - [ ] USE metrics (Utilization, Saturation, Errors)
  - [ ] Dashboard variables and templating
- [ ] Configure alerting:
  - [ ] PrometheusRule CRDs
  - [ ] Alertmanager routing and silences
  - [ ] Alert grouping and inhibition
- [ ] Document PromQL patterns and anti-patterns

---

### 4.2 SLOs & Error Budgets

**Goal:** Implement Service Level Objectives for reliability-driven operations

*SLOs are taught early because they're used throughout: canary analysis, deployment decisions, capacity planning.*

**Learning objectives:**
- Understand SLI/SLO/SLA hierarchy
- Implement error budget tracking
- Use SLOs to drive architectural decisions

**Tasks:**
- [ ] Create `experiments/slo-tutorial/`
- [ ] Deploy SLO tooling:
  - [ ] Sloth (SLO generator for Prometheus)
  - [ ] Pyrra (SLO dashboards and alerts)
- [ ] Define SLIs for demo application:
  - [ ] Availability SLI (successful requests / total requests)
  - [ ] Latency SLI (requests < threshold / total requests)
  - [ ] Throughput SLI (if applicable)
- [ ] Create SLO specifications:
  - [ ] 99.9% availability (43.8 min/month error budget)
  - [ ] 99% latency < 200ms
  - [ ] Multi-window, multi-burn-rate alerts
- [ ] Error budget tracking:
  - [ ] Error budget remaining dashboard
  - [ ] Burn rate visualization
  - [ ] Budget depletion forecasting
- [ ] SLO-driven alerting:
  - [ ] Fast burn alerts (immediate action)
  - [ ] Slow burn alerts (trending toward breach)
  - [ ] Error budget exhaustion alerts
- [ ] Operational integration:
  - [ ] SLO review process
  - [ ] Error budget policy (freeze deploys when exhausted)
  - [ ] SLO-based incident prioritization
- [ ] Document SLO patterns and anti-patterns
- [ ] **ADR:** Document SLO strategy and target selection

---

### 4.3 MinIO Object Storage

**Goal:** Deploy S3-compatible object storage as foundation for observability backends

*MinIO is taught here because Loki, Thanos, Tempo, Velero, and Argo Workflows all need object storage.*

**Learning objectives:**
- Understand MinIO architecture
- Configure for observability use cases
- Establish storage foundation for later phases

**Tasks:**
- [ ] Create `experiments/minio-tutorial/`
- [ ] Deploy MinIO operator
- [ ] Create MinIO tenant:
  - [ ] Single node (development)
  - [ ] Multi-node distributed (HA)
- [ ] Configure:
  - [ ] Buckets and policies
  - [ ] Access keys and IAM
  - [ ] Lifecycle rules
  - [ ] Versioning
- [ ] Create buckets for observability:
  - [ ] `loki-chunks` - for log storage
  - [ ] `thanos-blocks` - for metrics long-term storage
  - [ ] `tempo-traces` - for trace storage
  - [ ] `velero-backups` - for cluster backups (Phase 10)
  - [ ] `argo-artifacts` - for workflow artifacts (Phase 13)
- [ ] Monitoring:
  - [ ] MinIO metrics in Prometheus
  - [ ] Storage capacity dashboards
  - [ ] Alert on bucket growth
- [ ] Document object storage patterns

---

### 4.4 Loki & Log Aggregation

**Goal:** Centralized logging with Loki and LogQL

*Requires: Phase 4.3 (MinIO) for log chunk storage*

**Learning objectives:**
- Understand Loki's label-based architecture (vs full-text indexing)
- Write effective LogQL queries
- Correlate logs with metrics in Grafana

**Tasks:**
- [ ] Create `experiments/loki-tutorial/`
- [ ] Deploy Loki stack (Loki + Promtail)
- [ ] Configure Loki storage:
  - [ ] Point to MinIO bucket from Phase 4.3
  - [ ] Configure retention policies
- [ ] Build app with structured JSON logging
- [ ] Configure Promtail pipelines:
  - [ ] Label extraction (namespace, pod, container)
  - [ ] JSON field parsing
  - [ ] Regex extraction
  - [ ] Drop/keep filtering
  - [ ] Multiline log handling
- [ ] Write LogQL tutorial:
  - [ ] Label matchers and line filters
  - [ ] Parser expressions (json, pattern, regexp)
  - [ ] Metric queries (rate, count_over_time)
  - [ ] Unwrap for numeric fields
- [ ] Build log dashboards:
  - [ ] Log panel with live tail
  - [ ] Log volume over time
  - [ ] Error log filtering
- [ ] Set up log-based alerts (error rate threshold)
- [ ] Correlate logs ↔ metrics in Grafana (split view)
- [ ] Document logging best practices

---

### 4.5 OpenTelemetry & Distributed Tracing

**Goal:** End-to-end observability with traces, connecting metrics and logs

*Requires: Phase 4.1 (Prometheus), Phase 4.3 (MinIO), Phase 4.4 (Loki)*

**Learning objectives:**
- Understand OpenTelemetry architecture (SDK, Collector, backends)
- Instrument applications for distributed tracing
- Correlate traces ↔ metrics ↔ logs

**Tasks:**
- [ ] Create `experiments/opentelemetry-tutorial/`
- [ ] Deploy OpenTelemetry Collector
- [ ] Deploy Tempo (using MinIO for storage) or Jaeger as trace backend
- [ ] Build multi-service demo app (3+ services):
  - [ ] Service A → Service B → Service C
  - [ ] Each service instrumented with OTel SDK
- [ ] Implement tracing:
  - [ ] Auto-instrumentation (HTTP, gRPC, DB)
  - [ ] Manual span creation
  - [ ] Span attributes and events
  - [ ] Context propagation (W3C Trace Context)
- [ ] Configure Collector:
  - [ ] OTLP receiver
  - [ ] Batch processor
  - [ ] Exporters (Tempo/Jaeger, Prometheus)
- [ ] Connect the three pillars:
  - [ ] Exemplars (metrics → traces)
  - [ ] Trace ID in logs (logs → traces)
  - [ ] Service graph from traces
- [ ] Build trace-aware dashboards:
  - [ ] Service dependency map
  - [ ] Latency breakdown by span
  - [ ] Error trace exploration
- [ ] Document sampling strategies (head vs tail)

---

### 4.6 Thanos for Multi-Cluster Metrics

**Goal:** Long-term metrics storage and global query view across clusters

*Requires: Phase 4.1 (Prometheus), Phase 4.3 (MinIO)*

**Learning objectives:**
- Understand Thanos architecture (Sidecar, Store, Query, Compactor)
- Implement multi-cluster metrics aggregation
- Configure long-term retention with object storage

**Tasks:**
- [ ] Create `experiments/thanos-tutorial/`
- [ ] Deploy Thanos components:
  - [ ] Sidecar (alongside Prometheus)
  - [ ] Store Gateway (for object storage queries)
  - [ ] Query (global query layer)
  - [ ] Compactor (downsampling and retention)
- [ ] Configure object storage:
  - [ ] Use MinIO bucket from Phase 4.3
  - [ ] Retention policies (raw, 5m, 1h downsampling)
- [ ] Multi-cluster setup:
  - [ ] Prometheus + Sidecar per cluster
  - [ ] Central Query component
  - [ ] External labels for cluster identification
- [ ] Query patterns:
  - [ ] Cross-cluster queries
  - [ ] Historical data queries
  - [ ] Deduplication strategies
- [ ] Grafana integration:
  - [ ] Thanos Query as datasource
  - [ ] Multi-cluster dashboards
- [ ] Compare with alternatives:
  - [ ] Thanos vs Cortex vs Mimir
  - [ ] Storage costs and performance
- [ ] Document Thanos operational patterns
- [ ] **ADR:** Document long-term metrics strategy

---

## Phase 5: Traffic Management

*Control how traffic flows before learning deployment strategies that depend on it.*

### 5.1 Gateway API Deep Dive

**Goal:** Master Kubernetes Gateway API for ingress and traffic routing

**Learning objectives:**
- Understand Gateway API resources (Gateway, HTTPRoute, GRPCRoute)
- Implement advanced routing patterns
- Compare with legacy Ingress

**Tasks:**
- [ ] Create `experiments/gateway-api-tutorial/`
- [ ] Deploy Gateway API implementation:
  - [ ] **Contour** (work requirement - Envoy-based)
  - [ ] Envoy Gateway (alternative)
  - [ ] Cilium Gateway (if using Cilium CNI)
- [ ] Configure Gateway resource
- [ ] Implement HTTPRoute patterns:
  - [ ] Path-based routing
  - [ ] Host-based routing (virtual hosts)
  - [ ] Header matching
  - [ ] Query parameter routing
  - [ ] Method matching (GET vs POST)
- [ ] Traffic manipulation:
  - [ ] Weight-based splitting (A/B)
  - [ ] Request mirroring
  - [ ] URL rewriting
  - [ ] Header modification (add/remove/set)
  - [ ] Redirects
- [ ] Advanced features:
  - [ ] Timeouts and retries
  - [ ] Rate limiting (via policy attachment)
  - [ ] CORS configuration
- [ ] TLS configuration:
  - [ ] TLS termination (with cert-manager certs)
  - [ ] TLS passthrough
  - [ ] mTLS with client certificates
- [ ] Multi-gateway setup:
  - [ ] Internal vs external gateways
  - [ ] Namespace isolation (ReferenceGrant)
- [ ] Document Gateway API vs Ingress migration
- [ ] **ADR:** Document Gateway API implementation choice

---

### 5.2 Ingress Controllers Comparison

**Goal:** Understand trade-offs between ingress implementations

**Learning objectives:**
- Compare nginx, Traefik, and Envoy-based controllers
- Understand feature/performance trade-offs
- Make informed controller selection

**Tasks:**
- [ ] Create `experiments/ingress-comparison/`
- [ ] Deploy and configure:
  - [ ] **Contour** (work requirement - Envoy-based, Gateway API native)
  - [ ] Nginx Ingress Controller
  - [ ] Traefik
  - [ ] Envoy Gateway
- [ ] Implement equivalent routing on each
- [ ] Compare:
  - [ ] Configuration complexity
  - [ ] Feature availability
  - [ ] Resource consumption
  - [ ] Custom resource patterns
- [ ] Test advanced features:
  - [ ] Rate limiting implementation
  - [ ] Authentication integration
  - [ ] Custom error pages
- [ ] Document selection criteria

---

### 5.3 API Gateway Patterns

**Goal:** Implement API management patterns beyond basic routing

**Learning objectives:**
- Understand API gateway responsibilities
- Implement authentication, rate limiting, and API versioning
- Evaluate managed vs self-hosted options

**Tasks:**
- [ ] Create `experiments/api-gateway-tutorial/`
- [ ] Deploy Kong or Ambassador (or use Envoy Gateway)
- [ ] Implement API management features:
  - [ ] API key authentication
  - [ ] JWT validation (integrate with **Auth0** from Phase 3.7)
  - [ ] OAuth2/OIDC integration
- [ ] Rate limiting patterns:
  - [ ] Per-client rate limits
  - [ ] Global rate limits
  - [ ] Quota management
- [ ] API versioning strategies:
  - [ ] Path-based (/v1/, /v2/)
  - [ ] Header-based (Accept-Version)
  - [ ] Request transformation between versions
- [ ] API analytics:
  - [ ] Request logging
  - [ ] Usage metrics per consumer
  - [ ] Error rate by endpoint
- [ ] Developer portal (optional):
  - [ ] API documentation (OpenAPI)
  - [ ] Self-service key provisioning
- [ ] Document API gateway patterns
- [ ] **ADR:** Document API versioning strategy

---

## Phase 6: Service Mesh

*Service mesh builds on traffic management with mTLS, observability, and advanced traffic control.*

### 6.0 Service Mesh Decision Framework

**Goal:** Understand when and which service mesh to use before diving into implementations

**Learning objectives:**
- Compare service mesh architectures (sidecar vs sidecar-free)
- Understand feature trade-offs between options
- Make informed mesh selection decisions

**Tasks:**
- [ ] Create `docs/service-mesh-comparison.md`
- [ ] Architecture comparison:
  - [ ] Sidecar proxy model (Istio, Linkerd)
  - [ ] Sidecar-free/eBPF model (Cilium)
  - [ ] Control plane architectures
- [ ] Feature matrix:
  - [ ] mTLS implementation differences
  - [ ] Traffic management capabilities
  - [ ] Observability integration
  - [ ] Multi-cluster support
- [ ] Operational considerations:
  - [ ] Resource overhead (CPU, memory per pod)
  - [ ] Upgrade complexity
  - [ ] Debugging difficulty
  - [ ] Learning curve
- [ ] When to use each:
  - [ ] Istio: Full feature set, complex policies needed
  - [ ] Linkerd: Simplicity, lower resource overhead
  - [ ] Cilium: Already using Cilium CNI, performance critical
  - [ ] No mesh: Simple applications, overhead not justified
- [ ] When NOT to use service mesh:
  - [ ] Small number of services
  - [ ] No mTLS requirement
  - [ ] Team lacks operational capacity
- [ ] Document decision criteria
- [ ] **ADR:** Document service mesh decision for this lab

*Note: Phase 12.3 provides detailed performance benchmarks after you've learned each mesh.*

---

### 6.1 Istio Deep Dive

**Goal:** Master Istio service mesh fundamentals

**Learning objectives:**
- Understand Istio architecture (control plane, data plane, sidecars)
- Configure traffic management policies
- Implement security with mTLS

**Tasks:**
- [ ] Create `experiments/istio-tutorial/`
- [ ] Install Istio (istioctl or Helm)
- [ ] Enable sidecar injection (namespace label)
- [ ] Deploy sample microservices app
- [ ] Traffic management:
  - [ ] VirtualService routing rules
  - [ ] DestinationRule load balancing
  - [ ] Traffic splitting (canary)
  - [ ] Fault injection (delays, aborts)
  - [ ] Circuit breaking
  - [ ] Retries and timeouts
- [ ] Security:
  - [ ] Automatic mTLS (PeerAuthentication)
  - [ ] Authorization policies (allow/deny)
  - [ ] JWT validation (RequestAuthentication)
- [ ] Observability:
  - [ ] Kiali service graph
  - [ ] Distributed tracing (Jaeger integration)
  - [ ] Metrics (Prometheus integration)
- [ ] Gateway:
  - [ ] Istio Gateway (vs Gateway API)
  - [ ] External traffic management
- [ ] Document Istio patterns and gotchas

---

### 6.2 Linkerd Tutorial

**Goal:** Learn lightweight service mesh alternative

**Learning objectives:**
- Understand Linkerd architecture (simpler than Istio)
- Compare operational complexity
- Evaluate for different use cases

**Tasks:**
- [ ] Create `experiments/linkerd-tutorial/`
- [ ] Install Linkerd (CLI + control plane)
- [ ] Inject proxies into workloads
- [ ] Deploy same sample app as Istio experiment
- [ ] Configure:
  - [ ] Automatic mTLS
  - [ ] Traffic splitting (TrafficSplit CRD)
  - [ ] Retries and timeouts (ServiceProfile)
  - [ ] Authorization policies
- [ ] Observability:
  - [ ] Linkerd dashboard
  - [ ] Tap for live traffic inspection
  - [ ] Metrics and golden signals
- [ ] Compare with Istio:
  - [ ] Resource consumption
  - [ ] Configuration complexity
  - [ ] Feature coverage
- [ ] Document when to choose Linkerd vs Istio

---

### 6.3 Cilium Service Mesh (eBPF)

**Goal:** Explore sidecar-free service mesh with eBPF

**Learning objectives:**
- Understand eBPF-based networking
- Compare sidecar vs sidecar-free architectures
- Evaluate Cilium for CNI + service mesh

**Tasks:**
- [ ] Create `experiments/cilium-tutorial/`
- [ ] Install Cilium as CNI with service mesh features
- [ ] Deploy sample app (no sidecars needed)
- [ ] Configure:
  - [ ] L7 traffic policies (CiliumNetworkPolicy)
  - [ ] mTLS (Cilium encryption)
  - [ ] Load balancing
  - [ ] Ingress (Cilium Ingress or Gateway API)
- [ ] Observability:
  - [ ] Hubble for network visibility
  - [ ] Hubble UI
  - [ ] Prometheus metrics
- [ ] Compare with sidecar meshes:
  - [ ] Performance overhead
  - [ ] Resource consumption
  - [ ] Operational complexity
- [ ] Document eBPF advantages and limitations
- [ ] **ADR:** Document service mesh selection criteria

---

### 6.4 Cross-Cluster Networking

**Goal:** Enable service discovery and communication across multiple clusters

**Learning objectives:**
- Understand multi-cluster networking patterns
- Implement cross-cluster service discovery
- Design for geographic distribution

**Tasks:**
- [ ] Create `experiments/cross-cluster-networking/`
- [ ] Evaluate and implement option:
  - [ ] **Cilium ClusterMesh** (if using Cilium CNI) OR
  - [ ] **Submariner** (CNI-agnostic)
- [ ] Cilium ClusterMesh path:
  - [ ] Enable ClusterMesh on multiple Kind clusters
  - [ ] Configure cluster peering
  - [ ] Global services (service available in all clusters)
  - [ ] Service affinity (prefer local, failover to remote)
- [ ] Submariner path:
  - [ ] Deploy Submariner broker
  - [ ] Join clusters to broker
  - [ ] ServiceExport/ServiceImport resources
  - [ ] Lighthouse DNS for service discovery
- [ ] Cross-cluster patterns:
  - [ ] Active-active service deployment
  - [ ] Failover scenarios
  - [ ] Latency-aware routing
- [ ] Security:
  - [ ] Encrypted tunnel between clusters
  - [ ] NetworkPolicy across clusters
  - [ ] Identity federation
- [ ] Testing:
  - [ ] Cross-cluster service call latency
  - [ ] Failover time measurement
  - [ ] Partition tolerance testing
- [ ] Document cross-cluster patterns
- [ ] **ADR:** Document multi-cluster networking decision

---

## Phase 7: Messaging & Event Streaming

*Asynchronous communication patterns for event-driven architectures.*

### 7.0 Messaging Decision Framework

**Goal:** Understand messaging patterns and when to use each technology

**Learning objectives:**
- Compare messaging paradigms (queues vs streams vs pub/sub)
- Understand delivery guarantees and trade-offs
- Make informed messaging technology decisions

**Tasks:**
- [ ] Create `docs/messaging-comparison.md`
- [ ] Messaging paradigms:
  - [ ] Message queues (point-to-point, competing consumers)
  - [ ] Event streaming (log-based, replay capability)
  - [ ] Pub/sub (topic-based, fan-out)
  - [ ] Request/reply (synchronous over async)
- [ ] Technology comparison:
  - [ ] **Kafka:** Event streaming, high throughput, log retention
  - [ ] **RabbitMQ:** Traditional queuing, routing flexibility, protocols
  - [ ] **NATS:** Lightweight, low latency, cloud-native
  - [ ] **Cloud queues (SQS/Service Bus):** Managed, serverless integration
- [ ] Decision criteria:
  - [ ] Message ordering requirements
  - [ ] Delivery guarantees (at-most-once, at-least-once, exactly-once)
  - [ ] Throughput and latency requirements
  - [ ] Message replay needs
  - [ ] Operational complexity tolerance
- [ ] Use case mapping:
  - [ ] Event sourcing / CQRS → Kafka
  - [ ] Task queues / work distribution → RabbitMQ
  - [ ] Real-time microservice communication → NATS
  - [ ] Cloud-native serverless → SQS/Service Bus
- [ ] Anti-patterns:
  - [ ] Using Kafka for simple task queues
  - [ ] Using RabbitMQ for event sourcing
  - [ ] Over-engineering with messaging when HTTP suffices
- [ ] Document decision criteria
- [ ] **ADR:** Document messaging technology selection for this lab

*Note: Phase 12.2 provides detailed performance benchmarks after you've learned each system.*

---

### 7.1 Kafka with Strimzi

**Goal:** Deploy and operate Kafka on Kubernetes

**Learning objectives:**
- Understand Kafka architecture (brokers, topics, partitions, consumers)
- Use Strimzi operator for Kafka lifecycle
- Implement common messaging patterns

**Tasks:**
- [ ] Create `experiments/kafka-tutorial/`
- [ ] Deploy Strimzi operator via ArgoCD
- [ ] Create Kafka cluster (KafkaCluster CRD)
- [ ] Configure:
  - [ ] Topics (KafkaTopic CRD)
  - [ ] Users and ACLs (KafkaUser CRD)
  - [ ] Replication and partitions
- [ ] Build producer/consumer apps:
  - [ ] Simple pub/sub
  - [ ] Consumer groups
  - [ ] Exactly-once semantics
- [ ] Implement patterns:
  - [ ] Event sourcing
  - [ ] CQRS with Kafka
  - [ ] Dead letter queue
- [ ] Monitoring:
  - [ ] Kafka metrics in Prometheus
  - [ ] Consumer lag monitoring
  - [ ] Grafana dashboards
- [ ] Connect (optional):
  - [ ] Kafka Connect for integrations
  - [ ] Source/sink connectors
- [ ] Document Kafka operational patterns

---

### 7.2 RabbitMQ with Operator

**Goal:** Deploy and operate RabbitMQ for task queues

**Learning objectives:**
- Understand RabbitMQ architecture (exchanges, queues, bindings)
- Use RabbitMQ Cluster Operator
- Compare with Kafka use cases

**Tasks:**
- [ ] Create `experiments/rabbitmq-tutorial/`
- [ ] Deploy RabbitMQ Cluster Operator
- [ ] Create RabbitMQ cluster (RabbitmqCluster CRD)
- [ ] Configure:
  - [ ] Exchanges (direct, fanout, topic, headers)
  - [ ] Queues and bindings
  - [ ] Users and permissions
- [ ] Build producer/consumer apps:
  - [ ] Work queues (competing consumers)
  - [ ] Pub/sub (fanout)
  - [ ] Routing (topic exchange)
  - [ ] RPC pattern
- [ ] Implement reliability:
  - [ ] Publisher confirms
  - [ ] Consumer acknowledgments
  - [ ] Dead letter exchanges
  - [ ] Message TTL
- [ ] Monitoring:
  - [ ] RabbitMQ management UI
  - [ ] Prometheus metrics
  - [ ] Queue depth alerting
- [ ] Document RabbitMQ vs Kafka decision guide
- [ ] **ADR:** Document messaging technology selection

---

### 7.3 NATS & JetStream

**Goal:** Learn lightweight, high-performance messaging

**Learning objectives:**
- Understand NATS core vs JetStream
- Implement request-reply patterns
- Compare with Kafka and RabbitMQ

**Tasks:**
- [ ] Create `experiments/nats-tutorial/`
- [ ] Deploy NATS with JetStream enabled
- [ ] Core NATS patterns:
  - [ ] Pub/sub (fire and forget)
  - [ ] Request/reply
  - [ ] Queue groups (load balancing)
- [ ] JetStream (persistence):
  - [ ] Streams and consumers
  - [ ] At-least-once delivery
  - [ ] Message replay
  - [ ] Key-value store
  - [ ] Object store
- [ ] Build demo apps showcasing each pattern
- [ ] Compare with Kafka/RabbitMQ:
  - [ ] Latency
  - [ ] Throughput
  - [ ] Operational complexity
  - [ ] Use case fit
- [ ] Document NATS patterns and when to use

---

### 7.4 Cloud Messaging with Crossplane

**Goal:** Abstract cloud message queues with Crossplane XRDs

**Learning objectives:**
- Use Crossplane for managed messaging services
- Create portable queue abstractions
- Compare managed vs self-hosted

**Tasks:**
- [ ] Create `experiments/cloud-messaging/`
- [ ] Create XRD: SimpleQueue
  - [ ] Abstracts AWS SQS and Azure Service Bus
  - [ ] Common interface for both clouds
- [ ] Deploy same producer/consumer app to both clouds
- [ ] Compare:
  - [ ] Message visibility handling
  - [ ] Dead letter queue behavior
  - [ ] FIFO vs standard queues
  - [ ] Pricing models
- [ ] Test failover (queue in different region)
- [ ] Document cloud queue patterns

---

### 7.5 Distributed Coordination & ZooKeeper

**Goal:** Understand distributed coordination primitives and when to use them

**Learning objectives:**
- Understand ZooKeeper architecture and use cases
- Compare coordination systems (ZooKeeper vs etcd vs Consul)
- Implement common coordination patterns

**Tasks:**
- [ ] Create `experiments/distributed-coordination/`
- [ ] ZooKeeper deep dive:
  - [ ] Deploy ZooKeeper ensemble (3+ nodes)
  - [ ] Understand znodes, watches, ephemeral nodes
  - [ ] Leader election pattern
  - [ ] Distributed locks
  - [ ] Configuration management
  - [ ] ZooKeeper with Kafka (legacy mode)
- [ ] etcd comparison:
  - [ ] Deploy etcd cluster
  - [ ] Key-value operations
  - [ ] Watch API
  - [ ] etcd as Kubernetes backing store
  - [ ] Compare with ZooKeeper use cases
- [ ] Consul comparison:
  - [ ] Deploy Consul cluster
  - [ ] Service discovery features
  - [ ] Key-value store
  - [ ] Connect (service mesh features)
  - [ ] Multi-datacenter capabilities
- [ ] Modern alternatives:
  - [ ] Kafka KRaft (ZooKeeper-less Kafka)
  - [ ] When to use coordination services vs embedded consensus
- [ ] Use case mapping:
  - [ ] Leader election → ZooKeeper/etcd
  - [ ] Service discovery → Consul/Kubernetes DNS
  - [ ] Configuration → etcd/Consul KV
  - [ ] Distributed locks → ZooKeeper/etcd
- [ ] Operational considerations:
  - [ ] Quorum and failure tolerance
  - [ ] Performance characteristics
  - [ ] Backup and recovery
  - [ ] Monitoring and alerting
- [ ] Document coordination patterns and selection criteria
- [ ] **ADR:** Document coordination service selection

---

## Phase 8: Deployment Strategies

*Progressive complexity: rolling → blue-green → canary → GitOps patterns → feature flags.*

### 8.1 Rolling Updates Optimization

**Goal:** Master Kubernetes native rolling deployments

**Learning objectives:**
- Understand rolling update parameters
- Optimize for zero-downtime deployments
- Handle graceful shutdown correctly

**Tasks:**
- [ ] Create `experiments/rolling-update-tutorial/`
- [ ] Build app with slow startup and graceful shutdown
- [ ] Test parameter combinations:
  - [ ] maxSurge/maxUnavailable variations
  - [ ] minReadySeconds impact
  - [ ] progressDeadlineSeconds
- [ ] Implement graceful shutdown:
  - [ ] preStop hooks
  - [ ] terminationGracePeriodSeconds
  - [ ] Connection draining
- [ ] Readiness probe tuning:
  - [ ] initialDelaySeconds
  - [ ] periodSeconds
  - [ ] failureThreshold
- [ ] Load test during rollout (measure errors)
- [ ] Document recommended configurations

---

### 8.2 Blue-Green Deployments

**Goal:** Implement instant cutover deployments

**Learning objectives:**
- Understand blue-green pattern
- Implement with different tools
- Handle rollback scenarios

**Tasks:**
- [ ] Create `experiments/blue-green-tutorial/`
- [ ] Implement blue-green with:
  - [ ] Kubernetes Services (label selector swap)
  - [ ] Gateway API traffic switching
  - [ ] Argo Rollouts BlueGreen strategy
- [ ] Test scenarios:
  - [ ] Successful deployment
  - [ ] Failed health check (no switch)
  - [ ] Rollback after deployment
- [ ] Measure:
  - [ ] Cutover time
  - [ ] Request failures during switch
  - [ ] Resource overhead (2x replicas)
- [ ] Handle stateful considerations:
  - [ ] Database compatibility
  - [ ] Session handling
- [ ] Document blue-green patterns

---

### 8.3 Canary Deployments with Argo Rollouts

**Goal:** Implement gradual traffic shifting with automated analysis

*Requires: Phase 4.2 (SLOs) for analysis metrics, Phase 5.1 (Gateway API) for traffic splitting*

**Learning objectives:**
- Understand canary deployment pattern
- Configure Argo Rollouts
- Implement metric-based promotion/rollback

**Tasks:**
- [ ] Create `experiments/canary-tutorial/`
- [ ] Install Argo Rollouts
- [ ] Configure Rollout resource:
  - [ ] Traffic splitting steps (5% → 25% → 50% → 100%)
  - [ ] Pause durations
  - [ ] Manual gates
- [ ] Implement AnalysisTemplate:
  - [ ] Success rate query (Prometheus)
  - [ ] Latency threshold query
  - [ ] Custom business metrics
- [ ] Create "bad" versions to test:
  - [ ] High error rate version
  - [ ] High latency version
- [ ] Test automated rollback on failure
- [ ] Integrate with:
  - [ ] Gateway API (traffic splitting)
  - [ ] Istio (if mesh deployed)
- [ ] Document canary analysis patterns

---

### 8.4 GitOps Patterns with ArgoCD

**Goal:** Master ArgoCD for GitOps deployments

**Learning objectives:**
- Understand ArgoCD sync strategies
- Implement progressive delivery via Git
- Use ApplicationSets for multi-cluster

**Tasks:**
- [ ] Create `experiments/argocd-patterns/`
- [ ] Sync strategies:
  - [ ] Auto-sync vs manual
  - [ ] Self-heal behavior
  - [ ] Prune policies
- [ ] Sync waves and hooks:
  - [ ] Pre-sync hooks (DB migration job)
  - [ ] Sync waves (ordering)
  - [ ] Post-sync hooks (smoke tests)
  - [ ] SyncFail hooks (notifications)
- [ ] ApplicationSet patterns:
  - [ ] Git generator (directory/file)
  - [ ] Cluster generator (multi-cluster)
  - [ ] Matrix generator (combinations)
  - [ ] Progressive rollout across clusters
- [ ] App-of-apps pattern
- [ ] Document GitOps workflow patterns

---

### 8.5 Feature Flags & Progressive Delivery

**Goal:** Decouple deployment from release with feature flags

**Learning objectives:**
- Understand feature flag patterns
- Implement OpenFeature standard
- Combine with deployment strategies

**Tasks:**
- [ ] Create `experiments/feature-flags-tutorial/`
- [ ] Deploy feature flag service:
  - [ ] Flagsmith (self-hosted) OR
  - [ ] OpenFeature with flagd
- [ ] Implement flag patterns:
  - [ ] Boolean flags (feature on/off)
  - [ ] Percentage rollout
  - [ ] User segment targeting
  - [ ] A/B testing variants
- [ ] Integrate with application:
  - [ ] OpenFeature SDK integration
  - [ ] Server-side evaluation
  - [ ] Client-side evaluation
- [ ] Operational patterns:
  - [ ] Flag lifecycle (create → test → release → cleanup)
  - [ ] Kill switches for incidents
  - [ ] Gradual rollout with monitoring
- [ ] Combine with canary:
  - [ ] Deploy code → enable flag → monitor → full release
- [ ] Document feature flag patterns
- [ ] **ADR:** Document deployment vs release strategy

---

## Phase 9: Autoscaling & Resource Management

*Scale applications and infrastructure efficiently based on various signals.*

### 9.1 Horizontal Pod Autoscaler Deep Dive

**Goal:** Master HPA configuration for different workloads

**Learning objectives:**
- Understand HPA algorithm and behavior
- Configure for various metric types
- Tune for responsiveness vs stability

**Tasks:**
- [ ] Create `experiments/hpa-tutorial/`
- [ ] Build test app with configurable CPU/memory load
- [ ] Configure HPA scenarios:
  - [ ] CPU-based scaling
  - [ ] Memory-based scaling
  - [ ] Custom metrics (Prometheus adapter)
  - [ ] External metrics
- [ ] Tune parameters:
  - [ ] Target utilization thresholds
  - [ ] Stabilization windows (scale up/down)
  - [ ] Scaling policies (pods vs percent)
- [ ] Test workload patterns:
  - [ ] Gradual ramp-up
  - [ ] Sudden spike
  - [ ] Oscillating load
- [ ] Measure:
  - [ ] Time to scale
  - [ ] Over/under provisioning
  - [ ] Request latency during scaling
- [ ] Document HPA tuning guide

---

### 9.2 KEDA Event-Driven Autoscaling

**Goal:** Scale based on external event sources

*Note: Kafka/RabbitMQ scalers require Phase 7 (Messaging) knowledge*

**Learning objectives:**
- Understand KEDA architecture
- Configure various scalers
- Implement scale-to-zero

**Tasks:**
- [ ] Create `experiments/keda-tutorial/`
- [ ] Install KEDA
- [ ] Implement scalers:
  - [ ] Prometheus scaler (custom metrics)
  - [ ] Kafka scaler (consumer lag)
  - [ ] RabbitMQ scaler (queue depth)
  - [ ] Cron scaler (scheduled scaling)
  - [ ] Azure Service Bus / AWS SQS (via Crossplane)
- [ ] Configure ScaledObject:
  - [ ] Triggers and thresholds
  - [ ] Cooldown periods
  - [ ] Min/max replicas
  - [ ] Scale-to-zero behavior
- [ ] Test ScaledJob for batch workloads
- [ ] Compare KEDA vs HPA:
  - [ ] Configuration complexity
  - [ ] Supported triggers
  - [ ] Scale-to-zero capability
- [ ] Document KEDA patterns

---

### 9.3 Vertical Pod Autoscaler

**Goal:** Right-size pod resource requests automatically

**Learning objectives:**
- Understand VPA modes and recommendations
- Combine VPA with HPA
- Implement resource optimization workflow

**Tasks:**
- [ ] Create `experiments/vpa-tutorial/`
- [ ] Install VPA
- [ ] Configure VPA modes:
  - [ ] Off (recommendations only)
  - [ ] Initial (set on pod creation)
  - [ ] Auto (update running pods)
- [ ] Test with various workloads:
  - [ ] CPU-bound application
  - [ ] Memory-bound application
  - [ ] Variable workload
- [ ] Analyze recommendations:
  - [ ] Lower bound, target, upper bound
  - [ ] Uncapped vs capped
- [ ] Combine with HPA (mutually exclusive metrics)
- [ ] Document resource optimization workflow

---

### 9.4 Cluster Autoscaling

**Goal:** Automatically scale cluster nodes based on workload demand

**Learning objectives:**
- Understand Cluster Autoscaler vs Karpenter
- Configure node pools and scaling policies
- Optimize for cost and performance

**Tasks:**
- [ ] Create `experiments/cluster-autoscaler-tutorial/`
- [ ] Implement Cluster Autoscaler (AKS/EKS):
  - [ ] Node pool configuration
  - [ ] Scale-down policies
  - [ ] Pod disruption budgets interaction
- [ ] Implement Karpenter (EKS):
  - [ ] Provisioner configuration
  - [ ] Instance type selection
  - [ ] Spot vs on-demand
  - [ ] Consolidation policies
- [ ] Test scenarios:
  - [ ] Scale-up on pending pods
  - [ ] Scale-down on low utilization
  - [ ] Node replacement (spot interruption)
- [ ] Cost optimization:
  - [ ] Right-sizing node types
  - [ ] Spot instance integration
  - [ ] Reserved capacity planning
- [ ] Measure:
  - [ ] Time to provision new node
  - [ ] Scale-down delay
  - [ ] Cost per workload
- [ ] Document cluster autoscaling patterns
- [ ] **ADR:** Document Cluster Autoscaler vs Karpenter decision

---

### 9.5 Production Multi-Tenancy

**Goal:** Scale multi-tenant patterns for production with resource management

*Requires: Phase 3.8 (Multi-Tenancy Security) for isolation foundations*

**Learning objectives:**
- Implement resource fairness and quotas at scale
- Design blast radius boundaries
- Automate tenant lifecycle

**Tasks:**
- [ ] Create `experiments/multi-tenancy-production/`
- [ ] Build on Phase 3.8 security foundations:
  - [ ] Verify isolation from Phase 3.8 still holds
  - [ ] Add resource management layer
- [ ] Hierarchical namespaces (HNC):
  - [ ] Deploy Hierarchical Namespace Controller
  - [ ] Parent/child namespace inheritance
  - [ ] Propagated resources (secrets, configmaps)
  - [ ] Quota inheritance across hierarchy
- [ ] Resource quotas and limits:
  - [ ] ResourceQuotas per tenant namespace
  - [ ] LimitRanges for default pod resources
  - [ ] Aggregate quotas across tenant hierarchy
- [ ] Resource fairness:
  - [ ] PriorityClasses for tenant workloads
  - [ ] Pod priority and preemption rules
  - [ ] Fair-share scheduling concepts
- [ ] Noisy neighbor mitigation:
  - [ ] CPU/memory limits enforcement
  - [ ] I/O throttling patterns (if supported)
  - [ ] Network bandwidth limits (Cilium bandwidth manager)
- [ ] Tenant onboarding automation:
  - [ ] GitOps-driven tenant provisioning
  - [ ] Crossplane XRD for tenant creation
  - [ ] Automatic policy/quota application
- [ ] Tenant observability:
  - [ ] Per-tenant dashboards
  - [ ] Tenant-scoped alerting
  - [ ] Resource usage reporting
- [ ] Document production multi-tenancy patterns
- [ ] **ADR:** Document tenancy scaling decisions

---

### 9.6 FinOps Implementation & Chargeback

**Goal:** Full cost management with multi-tenant attribution

*Requires: Phase 1.4 (FinOps Foundation), Phase 4.1 (Prometheus), Phase 9.5 (Multi-Tenancy)*

**Learning objectives:**
- Implement per-tenant cost tracking
- Build chargeback/showback workflows
- Create cost optimization automation

**Tasks:**
- [ ] Create `experiments/finops-implementation/`
- [ ] Deploy full Kubecost or OpenCost:
  - [ ] Integration with cloud billing APIs
  - [ ] Azure Cost Management connection
  - [ ] AWS Cost Explorer connection
- [ ] Per-tenant cost attribution:
  - [ ] Cost by namespace (tenant)
  - [ ] Cost by label (team, project, cost-center)
  - [ ] Shared cost distribution (control plane, monitoring)
- [ ] Cost dashboards:
  - [ ] Daily/weekly/monthly trends
  - [ ] Tenant cost comparison
  - [ ] Idle resource identification
  - [ ] Right-sizing recommendations
- [ ] Chargeback workflows:
  - [ ] Automated cost reports per tenant
  - [ ] Budget allocation per tenant
  - [ ] Overage notifications
- [ ] Cost optimization:
  - [ ] Spot instance savings analysis
  - [ ] Reserved capacity recommendations
  - [ ] Resource right-sizing automation
- [ ] Alerts and governance:
  - [ ] Budget threshold alerts
  - [ ] Anomaly detection
  - [ ] Cost forecasting
- [ ] Document FinOps implementation patterns

---

## Phase 10: Data & Storage

*Stateful workloads: databases, caching, persistent storage, and disaster recovery.*

### 10.1 PostgreSQL with CloudNativePG

**Goal:** Operate PostgreSQL on Kubernetes with CloudNativePG

**Learning objectives:**
- Understand CloudNativePG operator
- Configure HA PostgreSQL clusters
- Implement backup and recovery

**Tasks:**
- [ ] Create `experiments/postgres-tutorial/`
- [ ] Deploy CloudNativePG operator
- [ ] Create PostgreSQL cluster:
  - [ ] Primary + replicas
  - [ ] Synchronous replication
  - [ ] Connection pooling (PgBouncer)
- [ ] Configure storage:
  - [ ] PVC sizing and storage class
  - [ ] WAL archiving to object storage
- [ ] Backup and recovery:
  - [ ] Scheduled backups (to S3/Azure via Crossplane)
  - [ ] Point-in-time recovery (PITR)
  - [ ] Restore to new cluster
- [ ] Monitoring:
  - [ ] pg_stat metrics in Prometheus
  - [ ] Grafana dashboards
  - [ ] Alerting on replication lag
- [ ] Failover testing:
  - [ ] Kill primary, verify promotion
  - [ ] Measure failover time
- [ ] Document PostgreSQL operational patterns

---

### 10.2 Redis with Spotahome Operator

**Goal:** Operate Redis on Kubernetes for caching

**Learning objectives:**
- Understand Redis sentinel vs cluster mode
- Configure persistence and HA
- Implement caching patterns

**Tasks:**
- [ ] Create `experiments/redis-tutorial/`
- [ ] Deploy Redis operator (Spotahome or similar)
- [ ] Create Redis deployments:
  - [ ] Standalone (development)
  - [ ] Sentinel (HA failover)
  - [ ] Cluster (horizontal scaling)
- [ ] Configure:
  - [ ] Persistence (RDB/AOF)
  - [ ] Memory limits and eviction
  - [ ] Password authentication
- [ ] Implement caching patterns:
  - [ ] Cache-aside
  - [ ] Write-through
  - [ ] Session storage
- [ ] Monitoring:
  - [ ] Redis metrics in Prometheus
  - [ ] Memory usage tracking
  - [ ] Hit/miss ratio
- [ ] Document Redis patterns for Kubernetes

---

### 10.3 Backup & Disaster Recovery

**Goal:** Implement comprehensive backup and DR strategy

*Requires: Phase 4.3 (MinIO) for backup storage*

**Learning objectives:**
- Understand Velero for cluster backup
- Implement cross-region DR patterns
- Design RTO/RPO strategies

**Tasks:**
- [ ] Create `experiments/backup-dr-tutorial/`
- [ ] Deploy Velero:
  - [ ] Configure backup storage (S3/Azure Blob)
  - [ ] Install plugins (AWS, Azure, CSI)
- [ ] Implement backup strategies:
  - [ ] Full cluster backup
  - [ ] Namespace-scoped backup
  - [ ] Label-selected backup
  - [ ] Scheduled backups (hourly/daily)
- [ ] Test restore scenarios:
  - [ ] Restore to same cluster
  - [ ] Restore to different cluster
  - [ ] Partial restore (specific resources)
- [ ] Volume backup:
  - [ ] CSI snapshot integration
  - [ ] Restic for non-CSI volumes
- [ ] Cross-region DR:
  - [ ] Backup replication to secondary region
  - [ ] DR cluster provisioning (Crossplane)
  - [ ] Application failover procedure
- [ ] Document RTO/RPO for different scenarios
- [ ] Create DR runbook
- [ ] **ADR:** Document backup and DR strategy

---

### 10.4 Schema Migration Patterns

**Goal:** Manage database schema changes in Kubernetes deployments

**Learning objectives:**
- Understand schema migration tools
- Implement zero-downtime migrations
- Integrate with GitOps workflows

**Tasks:**
- [ ] Create `experiments/schema-migration-tutorial/`
- [ ] Deploy migration tool:
  - [ ] Flyway OR Liquibase
- [ ] Implement migration patterns:
  - [ ] Init container migrations
  - [ ] Kubernetes Job migrations
  - [ ] ArgoCD pre-sync hook migrations
- [ ] Zero-downtime strategies:
  - [ ] Expand-contract pattern
  - [ ] Backward compatible changes
  - [ ] Blue-green database migrations
- [ ] Version management:
  - [ ] Migration versioning
  - [ ] Rollback strategies
  - [ ] Baseline migrations
- [ ] GitOps integration:
  - [ ] Migrations in Git
  - [ ] Sync wave ordering
  - [ ] Migration verification
- [ ] Document migration patterns
- [ ] **ADR:** Document schema migration strategy

---

## Phase 11: AI/ML Platform & Experiment Automation

*Use AI to conduct experiments, analyze results, and learn modern MLOps patterns.*

### 11.1 AI-Assisted Experiment Analysis

**Goal:** Use LLMs to analyze experiment results and generate insights

**Learning objectives:**
- Integrate AI into observability workflows
- Automate experiment analysis and reporting
- Build AI-powered operational tools

**Tasks:**
- [ ] Create `experiments/ai-analysis-tutorial/`
- [ ] Deploy AI infrastructure:
  - [ ] Ollama or vLLM for local inference
  - [ ] Model serving (Llama 3, Mistral, or similar)
  - [ ] GPU node pool (if cloud) or CPU inference
- [ ] Build analysis tools:
  - [ ] Prometheus metrics analyzer (anomaly detection)
  - [ ] Log summarization from Loki
  - [ ] Trace analysis for performance bottlenecks
- [ ] Experiment automation:
  - [ ] AI-generated experiment reports
  - [ ] Automated comparison analysis
  - [ ] Natural language query interface for metrics
- [ ] Integrate with workflows:
  - [ ] Post-experiment analysis step
  - [ ] Slack/notification summaries
  - [ ] Recommendation engine for next experiments
- [ ] Document AI-assisted operations patterns

---

### 11.2 Kubeflow Pipelines & MLOps

**Goal:** Implement ML training and serving pipelines on Kubernetes

**Learning objectives:**
- Understand Kubeflow components and architecture
- Build end-to-end ML pipelines
- Implement model versioning and serving

**Tasks:**
- [ ] Create `experiments/kubeflow-tutorial/`
- [ ] Deploy Kubeflow components:
  - [ ] Kubeflow Pipelines
  - [ ] Katib (hyperparameter tuning)
  - [ ] KServe (model serving)
- [ ] Build ML pipeline:
  - [ ] Data preprocessing step
  - [ ] Model training step
  - [ ] Model evaluation step
  - [ ] Model registration
- [ ] Implement MLOps patterns:
  - [ ] Experiment tracking (MLflow or Kubeflow native)
  - [ ] Model versioning and lineage
  - [ ] A/B model serving
  - [ ] Canary model deployments
- [ ] Integrate with platform:
  - [ ] Artifact storage (MinIO)
  - [ ] Metrics to Prometheus
  - [ ] Pipeline triggers from Argo Events
- [ ] Document MLOps architecture
- [ ] **ADR:** Document ML platform selection

---

### 11.3 KServe Model Serving

**Goal:** Deploy and manage ML models in production

**Learning objectives:**
- Understand KServe architecture
- Implement inference autoscaling
- Configure model monitoring

**Tasks:**
- [ ] Create `experiments/kserve-tutorial/`
- [ ] Deploy KServe:
  - [ ] Serverless inference
  - [ ] RawDeployment mode comparison
- [ ] Model serving patterns:
  - [ ] Single model deployment
  - [ ] Multi-model serving
  - [ ] Model transformers (pre/post processing)
- [ ] Autoscaling:
  - [ ] Scale-to-zero configuration
  - [ ] GPU autoscaling
  - [ ] Request-based scaling
- [ ] Traffic management:
  - [ ] Canary rollouts for models
  - [ ] A/B testing
  - [ ] Shadow deployments
- [ ] Monitoring:
  - [ ] Inference latency metrics
  - [ ] Model drift detection
  - [ ] Request logging
- [ ] Document model serving patterns

---

### 11.4 Vector Databases & RAG Infrastructure

**Goal:** Deploy vector search infrastructure for AI applications

**Learning objectives:**
- Understand vector database architectures
- Implement RAG (Retrieval Augmented Generation) patterns
- Evaluate different vector DB options

**Tasks:**
- [ ] Create `experiments/vector-db-tutorial/`
- [ ] Deploy vector databases:
  - [ ] Qdrant (Kubernetes-native)
  - [ ] Weaviate OR Milvus (comparison)
- [ ] Implement RAG pipeline:
  - [ ] Document ingestion and chunking
  - [ ] Embedding generation
  - [ ] Vector storage and indexing
  - [ ] Semantic search queries
  - [ ] LLM integration for generation
- [ ] Operational patterns:
  - [ ] Index management
  - [ ] Backup and restore
  - [ ] Horizontal scaling
- [ ] Build practical application:
  - [ ] Documentation search for this lab
  - [ ] Experiment results Q&A
- [ ] Compare vector DBs:
  - [ ] Query performance
  - [ ] Resource consumption
  - [ ] Ease of operation
- [ ] Document RAG architecture patterns
- [ ] **ADR:** Document vector database selection

---

## Phase 12: Advanced Topics & Benchmarks

*Deep dives and performance comparisons - now that fundamentals are solid.*

### 12.1 Database Performance Comparison

**Goal:** Compare relational databases for Kubernetes workloads

**Learning objectives:**
- Benchmark database performance objectively
- Understand trade-offs between options
- Make data-driven database selection

**Tasks:**
- [ ] Create `experiments/database-benchmark/`
- [ ] Deploy databases via Crossplane/operators:
  - [ ] PostgreSQL (CloudNativePG)
  - [ ] MySQL (via operator)
  - [ ] Cloud-managed (Azure SQL, RDS via Crossplane)
- [ ] Create benchmark schema and data
- [ ] Run benchmarks:
  - [ ] pgbench / sysbench
  - [ ] OLTP workloads (TPC-C style)
  - [ ] Read-heavy vs write-heavy
- [ ] Compare:
  - [ ] Throughput (TPS)
  - [ ] Latency percentiles
  - [ ] Resource consumption
  - [ ] Operational complexity
- [ ] Document findings and recommendations

---

### 12.2 Message Queue Performance Comparison

**Goal:** Compare messaging systems under load

**Learning objectives:**
- Benchmark throughput and latency
- Understand performance characteristics
- Inform technology selection

**Tasks:**
- [ ] Create `experiments/messaging-benchmark/`
- [ ] Deploy all three brokers (from Phase 7)
- [ ] Build benchmarking clients
- [ ] Test scenarios:
  - [ ] High throughput (max messages/sec)
  - [ ] Low latency (p99 measurement)
  - [ ] Fan-out (1 → N consumers)
  - [ ] Persistence impact
- [ ] Compare:
  - [ ] Messages per second
  - [ ] End-to-end latency
  - [ ] Resource consumption
  - [ ] Recovery time after failure
- [ ] Document performance comparison

---

### 12.3 Service Mesh Performance Comparison

**Goal:** Measure service mesh overhead

**Learning objectives:**
- Quantify latency overhead
- Compare resource consumption
- Inform mesh selection

**Tasks:**
- [ ] Create `experiments/mesh-benchmark/`
- [ ] Deploy baseline app (no mesh)
- [ ] Deploy same app with:
  - [ ] Istio
  - [ ] Linkerd
  - [ ] Cilium
- [ ] Measure:
  - [ ] Latency overhead (p50, p95, p99)
  - [ ] CPU per pod (sidecar cost)
  - [ ] Memory per pod
  - [ ] Control plane resources
- [ ] Test at scale:
  - [ ] 10, 50, 100 services
  - [ ] High RPS scenarios
- [ ] Document mesh comparison

---

### 12.4 Runtime Performance Comparison

**Goal:** Compare web server runtimes for API workloads

**Learning objectives:**
- Benchmark different language runtimes
- Understand performance characteristics
- Portfolio piece for runtime expertise

**Tasks:**
- [ ] Create `experiments/runtime-benchmark/`
- [ ] Build identical API in:
  - [ ] Go (net/http)
  - [ ] Rust (Axum)
  - [ ] .NET (ASP.NET Core)
  - [ ] Node.js (Fastify)
  - [ ] Bun
- [ ] Implement endpoints:
  - [ ] GET /health
  - [ ] GET /json (serialize response)
  - [ ] POST /echo (deserialize + serialize)
  - [ ] GET /compute (CPU-bound work)
- [ ] Benchmark with k6:
  - [ ] Throughput (RPS)
  - [ ] Latency distribution
  - [ ] Memory footprint
  - [ ] Container image size
  - [ ] Cold start time
- [ ] Document runtime comparison

---

## Phase 13: Chaos Engineering

*Validate resilience - capstone experiments after everything else is solid.*

### 13.1 Pod Failure & Recovery

**Goal:** Measure application resilience to pod failures

**Tasks:**
- [ ] Create `experiments/chaos-pod-failure/`
- [ ] Deploy Chaos Mesh
- [ ] Test scenarios:
  - [ ] Single pod kill
  - [ ] Multiple pod kill (50%)
  - [ ] Container crash loop
- [ ] Measure recovery time and error rates
- [ ] Document resilience findings

---

### 13.2 Network Chaos

**Goal:** Test application behavior under network issues

**Tasks:**
- [ ] Create `experiments/chaos-network/`
- [ ] Test with Chaos Mesh NetworkChaos:
  - [ ] Latency injection (50ms, 200ms, 500ms)
  - [ ] Packet loss (1%, 5%, 20%)
  - [ ] Network partition
- [ ] Measure:
  - [ ] Timeout behavior
  - [ ] Circuit breaker activation
  - [ ] Retry storms
- [ ] Document network resilience patterns

---

### 13.3 Node Drain & Zone Failure

**Goal:** Test infrastructure-level failures

**Tasks:**
- [ ] Create `experiments/chaos-infrastructure/`
- [ ] Test scenarios:
  - [ ] Graceful node drain
  - [ ] Sudden node failure
  - [ ] Zone failure (multi-zone cluster)
- [ ] Measure:
  - [ ] Workload redistribution time
  - [ ] Request failures during event
  - [ ] PVC reattachment time
- [ ] Document infrastructure resilience

---

## Phase 14: Workflow Orchestration & Automation

*Build automation that ties experiments together - this phase uses learnings from all previous phases.*

### 14.1 Argo Workflows Deep Dive

**Goal:** Master workflow orchestration patterns (informed by running experiments)

**Learning objectives:**
- Understand Argo Workflows concepts
- Build complex multi-step workflows
- Handle artifacts and parameters

**Tasks:**
- [ ] Create `experiments/argo-workflows-tutorial/`
- [ ] Workflow patterns:
  - [ ] Sequential steps
  - [ ] Parallel execution
  - [ ] DAG dependencies
  - [ ] Conditional execution (when)
  - [ ] Loops (withItems, withParam)
- [ ] Parameters and artifacts:
  - [ ] Input/output parameters
  - [ ] Artifact passing between steps
  - [ ] S3/MinIO artifact storage
- [ ] Templates:
  - [ ] Container templates
  - [ ] Script templates
  - [ ] WorkflowTemplate (reusable)
  - [ ] ClusterWorkflowTemplate
- [ ] Error handling:
  - [ ] Retry strategies
  - [ ] Timeout configuration
  - [ ] Exit handlers
  - [ ] ContinueOn failure
- [ ] Build practical workflows from experiments:
  - [ ] Experiment runner (deploy → test → analyze → cleanup)
  - [ ] Benchmark suite (run all Phase 12 benchmarks)
  - [ ] Chaos test pipeline (Phase 13 automation)
- [ ] Document workflow patterns

---

### 14.2 Argo Events

**Goal:** Event-driven workflow triggering

**Learning objectives:**
- Understand Argo Events architecture
- Configure event sources and sensors
- Integrate with Argo Workflows

**Tasks:**
- [ ] Create `experiments/argo-events-tutorial/`
- [ ] Deploy Argo Events
- [ ] Configure EventSources:
  - [ ] Webhook (HTTP triggers)
  - [ ] GitHub (push, PR events)
  - [ ] Kafka (message triggers)
  - [ ] Cron (scheduled triggers)
  - [ ] S3/MinIO (object events)
- [ ] Configure Sensors:
  - [ ] Event filtering
  - [ ] Parameter extraction
  - [ ] Trigger templates
- [ ] Integrate triggers:
  - [ ] Trigger Argo Workflow
  - [ ] Trigger Kubernetes resource
  - [ ] Trigger HTTP endpoint
- [ ] Build event-driven pipelines:
  - [ ] GitHub push → experiment workflow
  - [ ] Scheduled benchmark runs
  - [ ] Alert → chaos test trigger
- [ ] Document event-driven patterns

---

### 14.3 Advanced CI/CD Patterns

**Goal:** Advanced CI/CD orchestration building on Phase 2 foundations

*Builds on Phase 2 (CI/CD & Supply Chain Security) with advanced patterns*

**Learning objectives:**
- Compare advanced CI/CD orchestration options
- Implement complex multi-environment pipelines
- Design hybrid CI/CD architectures

**Tasks:**
- [ ] Create `experiments/advanced-cicd/`
- [ ] Argo Workflows for CI:
  - [ ] Build pipelines as workflows
  - [ ] Parallel test execution
  - [ ] Artifact management
- [ ] Tekton Pipelines comparison:
  - [ ] Pipeline and Task resources
  - [ ] Tekton vs Argo Workflows trade-offs
- [ ] Multi-environment promotion:
  - [ ] Dev → Staging → Production
  - [ ] Environment-specific configs
  - [ ] Promotion gates and approvals
  - [ ] Automated rollback on failure
- [ ] GitLab CI advanced patterns:
  - [ ] GitLab Kubernetes Agent
  - [ ] Auto DevOps vs custom pipelines
  - [ ] Review environments
- [ ] Hybrid CI/CD architecture:
  - [ ] CI (GitHub Actions/GitLab) + CD (ArgoCD)
  - [ ] Image Updater for GitOps
  - [ ] Notification integration
- [ ] Document advanced CI/CD patterns

---

## Phase 15: Developer Experience & Internal Platform

*Build an Internal Developer Platform (IDP) that ties together all the platform pieces.*

### 15.1 Backstage Developer Portal

**Goal:** Deploy Backstage as a unified developer portal

**Learning objectives:**
- Understand Internal Developer Platform (IDP) concepts
- Configure Backstage catalog and plugins
- Integrate with existing platform components

**Tasks:**
- [ ] Create `experiments/backstage-tutorial/`
- [ ] Deploy Backstage:
  - [ ] Helm chart or ArgoCD Application
  - [ ] PostgreSQL backend (via CloudNativePG from Phase 10)
  - [ ] Authentication (integrate with Auth0 from Phase 3.7)
- [ ] Software Catalog:
  - [ ] Define catalog-info.yaml for services
  - [ ] Component types (service, website, library)
  - [ ] System and domain groupings
  - [ ] API definitions (OpenAPI, AsyncAPI, gRPC)
- [ ] Integrations:
  - [ ] Kubernetes plugin (show deployments, pods)
  - [ ] ArgoCD plugin (deployment status)
  - [ ] GitHub/GitLab integration (repo info, CI status)
  - [ ] Prometheus/Grafana plugin (metrics links)
  - [ ] PagerDuty or Opsgenie plugin (on-call info)
- [ ] Software Templates:
  - [ ] Create scaffolder template for new services
  - [ ] Include CI/CD pipeline, Dockerfile, Helm chart
  - [ ] Integrate with Crossplane for infrastructure
- [ ] TechDocs:
  - [ ] Enable TechDocs plugin
  - [ ] Generate docs from markdown in repos
- [ ] Document Backstage patterns
- [ ] **ADR:** Document IDP strategy (Backstage vs Port vs Cortex)

---

### 15.2 Self-Service Infrastructure

**Goal:** Enable developer self-service through the platform

**Learning objectives:**
- Design golden paths for common workflows
- Balance flexibility with guardrails
- Measure developer productivity

**Tasks:**
- [ ] Create `experiments/self-service-infra/`
- [ ] Golden paths:
  - [ ] New service creation (Backstage template → repo → CI/CD → deployed)
  - [ ] Database provisioning (Backstage → Crossplane claim → ready)
  - [ ] Environment creation (dev/staging/prod namespaces)
- [ ] Guardrails integration:
  - [ ] Policies from Phase 3.5 enforced automatically
  - [ ] Cost controls from Phase 9.6 applied
  - [ ] Security scanning from Phase 2 in templates
- [ ] Developer metrics:
  - [ ] Lead time for changes
  - [ ] Deployment frequency
  - [ ] Time to onboard new service
- [ ] Document self-service patterns

---

## Phase 16: Architecture Artifacts

*Documentation accumulated from all experiments - ADRs, runbooks, and capacity planning.*

### 16.1 Architecture Decision Records

**Goal:** Consolidate ADRs created throughout experiments

**Tasks:**
- [x] Create `docs/adrs/` directory
- [ ] Write ADR template (based on Michael Nygard format)
- [ ] Consolidate and polish ADRs from experiments:
  - [x] ADR-001: Spacelift for IaC orchestration
  - [ ] ADR-002: Secrets management approach (ESO + Vault)
  - [ ] ADR-003: CI/CD platform selection
  - [ ] ADR-004: Supply chain security strategy
  - [ ] ADR-005: Policy engine selection (Kyverno vs OPA)
  - [ ] ADR-006: Service mesh selection
  - [ ] ADR-007: Messaging technology selection
  - [ ] ADR-008: Database strategy (managed vs self-hosted)
  - [ ] ADR-009: Observability stack choices
  - [ ] ADR-010: Identity federation approach
  - [ ] ADR-011: Cost management strategy
  - [ ] ADR-012: Long-term metrics (Thanos)
  - [ ] ADR-013: ML platform selection
  - [ ] ADR-014: Vector database selection
  - [ ] ADR-015: Multi-cloud strategy (Crossplane vs Terraform)
  - [ ] ADR-016: Config management (Ansible vs Puppet vs Chef)
  - [ ] ADR-017: IDP strategy (Backstage vs Port vs Cortex)

---

### 16.2 Runbook Library

**Goal:** Consolidate operational runbooks developed during experiments

**Tasks:**
- [ ] Create `docs/runbooks/` directory
- [ ] Consolidate runbooks from experiments:
  - [ ] Cluster upgrade procedure (from Platform Bootstrap)
  - [ ] Certificate expiry remediation (from cert-manager)
  - [ ] Vault seal/unseal recovery (from Vault)
  - [ ] Database failover recovery (from PostgreSQL)
  - [ ] Kafka partition rebalancing (from Kafka)
  - [ ] Redis failover procedure (from Redis)
  - [ ] Velero restore procedure (from Backup/DR)
  - [ ] Chaos test execution guide (from Chaos Engineering)
  - [ ] Model rollback procedure (from KServe)
  - [ ] Image vulnerability response (from Supply Chain Security)
- [ ] Standardize format:
  - [ ] Prerequisites and warnings
  - [ ] Step-by-step procedures
  - [ ] Verification steps
  - [ ] Rollback procedures
- [ ] Incident response template

---

### 16.3 Capacity Planning Guide

**Goal:** Document capacity planning methodology from experiment data

**Tasks:**
- [ ] Create `docs/capacity-planning.md`
- [ ] Document sizing methodology:
  - [ ] Pod resource estimation (from VPA data)
  - [ ] Node pool sizing (from Cluster Autoscaler)
  - [ ] Storage IOPS requirements (from database experiments)
  - [ ] Network bandwidth planning (from mesh benchmarks)
- [ ] Create capacity models:
  - [ ] Small (dev/test): specifications
  - [ ] Medium (staging): specifications
  - [ ] Large (production): specifications
- [ ] Growth planning:
  - [ ] 10x traffic scenario
  - [ ] Multi-region expansion
  - [ ] Cost projections (from FinOps data)

---

### 16.4 Reference Architecture Document

**Goal:** Create comprehensive reference architecture from all learnings

**Tasks:**
- [ ] Create `docs/reference-architecture.md`
- [ ] Document architecture layers:
  - [ ] Infrastructure layer (Crossplane, Spacelift)
  - [ ] Platform layer (Kubernetes, service mesh, observability)
  - [ ] Application layer (deployments, services, ingress)
  - [ ] Data layer (databases, caching, messaging)
  - [ ] AI/ML layer (model serving, pipelines)
- [ ] Include diagrams:
  - [ ] High-level architecture
  - [ ] Network topology
  - [ ] Data flow diagrams
  - [ ] CI/CD pipeline flow
  - [ ] Supply chain security flow
- [ ] Security architecture:
  - [ ] Identity and access
  - [ ] Network segmentation
  - [ ] Secrets management
  - [ ] Certificate management
  - [ ] Supply chain security
- [ ] Operational model:
  - [ ] Day 1 vs Day 2 operations
  - [ ] SLO/SLI definitions
  - [ ] On-call responsibilities

---

## Learning Path Summary

| Phase | Focus | Experiments | Key Skills |
|-------|-------|-------------|------------|
| 1 | Platform Bootstrap & GitOps | 5 | GitOps, Raspberry Pi/Ansible, Spacelift, Crossplane, FinOps |
| 2 | CI/CD & Supply Chain Security | 4 | Image building, scanning, signing, SBOM, registries |
| 3 | Security Foundations | 8 | ESO, Vault, cert-manager, policies, identity, multi-tenancy |
| 4 | Observability | 6 | Prometheus, SLOs, MinIO, Loki, OpenTelemetry, Thanos |
| 5 | Traffic Management | 3 | Gateway API, Ingress, API Gateway |
| 6 | Service Mesh | 5 | Decision framework, Istio, Linkerd, Cilium, Cross-cluster |
| 7 | Messaging & Coordination | 6 | Kafka, RabbitMQ, NATS, Cloud queues, ZooKeeper/etcd/Consul |
| 8 | Deployment Strategies | 5 | Rolling, Blue-Green, Canary, GitOps, Feature Flags |
| 9 | Autoscaling & Resources | 6 | HPA, KEDA, VPA, Cluster Autoscaler, Multi-tenancy, FinOps |
| 10 | Data & Storage | 4 | PostgreSQL, Redis, Backup/DR, Migrations |
| 11 | AI/ML Platform | 4 | AI Analysis, Kubeflow, KServe, Vector DBs |
| 12 | Benchmarks | 4 | DB, Messaging, Mesh, Runtime comparisons |
| 13 | Chaos Engineering | 3 | Pod, Network, Infrastructure chaos |
| 14 | Workflow Orchestration | 3 | Argo Workflows, Events, Advanced CI/CD |
| 15 | Developer Experience | 2 | Backstage, Self-Service Infrastructure, IDP |
| 16 | Architecture Artifacts | 4 | ADRs, Runbooks, Capacity Planning, Reference Arch |

**Total: ~72 experiments**

---

## Notes

- All experiments follow `experiments/_template/` structure
- **Environment progression**: Kind (dev) → Raspberry Pi (home lab) → Cloud (prod)
- Ansible for bare-metal/Pi, Spacelift+Terraform for cloud, Crossplane for K8s-native
- Each experiment should have a portfolio-ready README
