## Phase 3: Security Foundations

*Security first - TLS, certificates, identity, secrets, and policies are prerequisites for everything else.*

*Secrets management is learned progressively: Sealed Secrets (simplest) → SOPS+age (cluster-independent) → ESO+OpenBao (full platform). This builds understanding of trade-offs before committing to a production pattern.*

### 3.1 Sealed Secrets

**Goal:** Learn the simplest GitOps-friendly secrets approach

*Sealed Secrets encrypts secrets so they can be stored in Git. The controller decrypts them in-cluster. Simple but cluster-bound.*

**Learning objectives:**
- Understand asymmetric encryption for secrets
- Learn the trade-offs of cluster-bound encryption
- Practice GitOps workflow with encrypted secrets

**Characteristics:**
| Aspect | Sealed Secrets |
|--------|----------------|
| **GitOps-friendly** | Yes - encrypted secrets in Git |
| **Cluster-independent** | No - keys tied to controller |
| **Central management** | No - each cluster has own keys |
| **Rotation** | Manual re-seal required |
| **Complexity** | Low |

**Tasks:**
- [ ] Create `experiments/scenarios/sealed-secrets/`
- [ ] Deploy Sealed Secrets controller to hub cluster
- [ ] Install `kubeseal` CLI
- [ ] Create and seal a secret:
  - [ ] Create plain Kubernetes Secret
  - [ ] Seal with `kubeseal`
  - [ ] Commit SealedSecret to Git
  - [ ] Verify unsealing in cluster
- [ ] Backup and restore:
  - [ ] Backup controller private key
  - [ ] Test restore to new cluster
  - [ ] Understand disaster recovery implications
- [ ] Limitations exercise:
  - [ ] Try to unseal on different cluster (should fail)
  - [ ] Document when this is a problem
- [ ] Document Sealed Secrets patterns and limitations

---

### 3.2 SOPS + age Encryption

**Goal:** Learn cluster-independent secret encryption

*SOPS encrypts files with age keys you control. Secrets can be decrypted anywhere you have the key - not tied to any cluster.*

**Learning objectives:**
- Understand SOPS and age encryption
- Learn key management fundamentals
- Compare with Sealed Secrets trade-offs

**Characteristics:**
| Aspect | SOPS + age |
|--------|------------|
| **GitOps-friendly** | Yes - encrypted files in Git |
| **Cluster-independent** | Yes - decrypt anywhere with key |
| **Central management** | Partial - key distribution needed |
| **Rotation** | Re-encrypt with new key |
| **Complexity** | Medium |

**Tasks:**
- [ ] Create `experiments/scenarios/sops-age/`
- [ ] Install SOPS and age CLIs
- [ ] Generate age keypair:
  - [ ] Understand public/private key model
  - [ ] Secure storage of private key
- [ ] Encrypt secrets with SOPS:
  - [ ] Create `.sops.yaml` configuration
  - [ ] Encrypt a secrets file
  - [ ] Commit encrypted file to Git
  - [ ] Decrypt locally to verify
- [ ] Integrate with ArgoCD:
  - [ ] Deploy KSOPS or SOPS plugin
  - [ ] Configure ArgoCD to decrypt on apply
  - [ ] Test GitOps workflow with encrypted secrets
- [ ] Key management patterns:
  - [ ] Multiple recipients (team access)
  - [ ] Key rotation procedure
- [ ] Compare with Sealed Secrets:
  - [ ] Portability advantages
  - [ ] Key management overhead
- [ ] Document SOPS patterns

---

### 3.3 OpenBao & External Secrets Operator

**Goal:** Implement centralized secrets management with OpenBao on the hub cluster

*OpenBao (Vault fork) on the hub cluster becomes the central secrets store. ESO syncs secrets to experiment clusters. This is the production pattern.*

**Learning objectives:**
- Understand OpenBao/Vault architecture (secrets engines, auth methods, policies)
- Configure External Secrets Operator
- Establish secrets management patterns for experiments

**Characteristics:**
| Aspect | ESO + OpenBao |
|--------|---------------|
| **GitOps-friendly** | Yes - ExternalSecret CRs in Git |
| **Cluster-independent** | Yes - OpenBao is external |
| **Central management** | Yes - single source of truth |
| **Rotation** | Yes - automatic via refreshInterval |
| **Complexity** | Higher |

**Architecture:**
```
┌─────────────────────────────────────┐
│  Hub Cluster                        │
│  ┌─────────────────────────────┐   │
│  │  OpenBao                     │   │
│  │  - KV secrets engine        │   │
│  │  - Kubernetes auth          │   │
│  │  - Per-experiment policies  │   │
│  └─────────────────────────────┘   │
└─────────────────────────────────────┘
          │ secrets API
          ▼
┌─────────────────────────────────────┐
│  Experiment Cluster                 │
│  ┌─────────────────────────────┐   │
│  │  External Secrets Operator   │   │
│  │  - ClusterSecretStore → Bao │   │
│  │  - ExternalSecret CRs       │   │
│  └─────────────────────────────┘   │
│             │                       │
│             ▼                       │
│  ┌─────────────────────────────┐   │
│  │  Kubernetes Secrets          │   │
│  │  (synced from OpenBao)      │   │
│  └─────────────────────────────┘   │
└─────────────────────────────────────┘
```

**Tasks:**
- [ ] Create `experiments/scenarios/eso-openbao/`
- [ ] Deploy OpenBao to hub cluster:
  - [ ] Helm chart deployment
  - [ ] Initialize and unseal
  - [ ] Enable KV secrets engine
- [ ] Configure authentication:
  - [ ] Kubernetes auth method
  - [ ] Per-namespace policies
  - [ ] Service account bindings
- [ ] Deploy ESO to experiment cluster:
  - [ ] Install ESO operator
  - [ ] Create ClusterSecretStore pointing to OpenBao
- [ ] Create ExternalSecret resources:
  - [ ] Sync a secret from OpenBao
  - [ ] Test refresh interval
  - [ ] Verify secret updates propagate
- [ ] Production patterns:
  - [ ] Secret rotation workflow
  - [ ] Audit logging
  - [ ] Backup and recovery
- [ ] Compare all three approaches:
  - [ ] When to use each
  - [ ] Migration paths
- [ ] Document ESO + OpenBao patterns
- [ ] **ADR:** Document secrets management progression (see `docs/adrs/ADR-002-secrets-management.md`)

---

### 3.4 cert-manager & TLS Automation

**Goal:** Automate TLS certificate lifecycle in Kubernetes

**Learning objectives:**
- Understand PKI fundamentals and certificate lifecycle
- Configure cert-manager issuers (self-signed, ACME, private CA)
- Automate certificate renewal and monitoring

**Tasks:**
- [ ] Create `experiments/scenarios/cert-manager-tutorial/`
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

### 3.5 Advanced OpenBao Patterns

**Goal:** Dynamic credentials, PKI, and advanced secret injection patterns

*Builds on Phase 3.3 (OpenBao + ESO basics) with production-grade patterns.*

**Learning objectives:**
- Implement dynamic database credentials
- Configure PKI secrets engine for certificate generation
- Compare secret injection methods (Agent vs CSI vs ESO)

**Tasks:**
- [ ] Create `experiments/scenarios/openbao-advanced/`
- [ ] Configure advanced auth methods:
  - [ ] AppRole (for CI/CD pipelines)
  - [ ] JWT/OIDC (for external identity providers)
- [ ] Set up dynamic secrets engines:
  - [ ] Database engine (dynamic PostgreSQL creds with TTL)
  - [ ] PKI engine (dynamic certificates for mTLS)
- [ ] Implement secret injection comparison:
  - [ ] OpenBao Agent Sidecar (file-based injection)
  - [ ] OpenBao CSI Provider (volume-based injection)
  - [ ] External Secrets Operator (Kubernetes Secret sync)
- [ ] Test secret rotation workflows:
  - [ ] Automatic credential rotation
  - [ ] Application restart-free rotation
- [ ] Implement audit logging and monitoring
- [ ] Configure OpenBao HA (Raft storage) for production
- [ ] Document injection method trade-offs
- [ ] **ADR:** Compare OpenBao injection methods (Agent vs CSI vs ESO)

---

### 3.6 Policy & Governance

**Goal:** Implement policy-as-code for compliance and operational guardrails

**Learning objectives:**
- Understand admission controllers and policy engines
- Implement organizational policies at scale
- Create audit trails for compliance

**Tasks:**
- [ ] Create `experiments/scenarios/policy-governance-tutorial/`
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

### 3.7 Network Policies & Pod Security

**Goal:** Implement defense-in-depth with network segmentation and pod security

**Learning objectives:**
- Write effective NetworkPolicy resources
- Understand Pod Security Standards (PSS)
- Implement least-privilege pod configurations

**Tasks:**
- [ ] Create `experiments/scenarios/network-security-tutorial/`
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

### 3.8 Identity & Access Management

**Goal:** Integrate external identity providers with Kubernetes RBAC

**Learning objectives:**
- Understand OIDC authentication flow
- Configure Kubernetes API server for external IdP
- Implement just-in-time access patterns

**Tasks:**
- [ ] Create `experiments/scenarios/identity-tutorial/`
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

### 3.9 Multi-Tenancy Security Foundations

**Goal:** Establish tenant isolation patterns using security primitives

*This covers security isolation - production-scale multi-tenancy (quotas, scheduling) comes in Phase 9.5*

**Learning objectives:**
- Implement namespace-based tenant isolation
- Apply policies for tenant boundaries
- Integrate RBAC with tenant model

**Tasks:**
- [ ] Create `experiments/scenarios/multi-tenancy-security/`
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

