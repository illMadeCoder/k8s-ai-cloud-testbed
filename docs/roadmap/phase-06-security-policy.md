# Phase 6: Security & Policy

*Secure the platform with TLS, secrets, policies, and network controls. Prerequisite for production-grade benchmarking.*

**Consolidated from:** Previous Phase 7 (Security Foundations) + Phase 8 (Network Security)

---

## 6.1 TLS & cert-manager

**Goal:** Automate TLS certificate lifecycle

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
- [ ] DNS-01 challenge (Azure DNS or Route53)
- [ ] Certificate expiry alerting
- [ ] Document certificate patterns

---

## 6.2 Secrets Management (ESO + OpenBao)

**Goal:** Centralized secrets management with External Secrets Operator

*OpenBao on hub cluster as central store, ESO syncs to experiment clusters.*

**Architecture:**
```
Hub Cluster                    Experiment Cluster
┌─────────────────┐           ┌─────────────────┐
│  OpenBao        │◄──────────│  ESO            │
│  - KV engine    │  secrets  │  - SecretStore  │
│  - K8s auth     │   API     │  - ExternalSecret│
└─────────────────┘           └────────┬────────┘
                                       ▼
                              ┌─────────────────┐
                              │ Kubernetes      │
                              │ Secrets         │
                              └─────────────────┘
```

**Tasks:**
- [ ] Create `experiments/eso-openbao/`
- [ ] Deploy OpenBao to hub:
  - [ ] Helm deployment
  - [ ] Initialize and unseal
  - [ ] Enable KV secrets engine
- [ ] Configure ESO on experiment cluster:
  - [ ] ClusterSecretStore → OpenBao
  - [ ] ExternalSecret resources
  - [ ] Test refresh interval
- [ ] Production patterns:
  - [ ] Secret rotation workflow
  - [ ] Audit logging
- [ ] **ADR:** Document secrets management (see ADR-002)

**Alternatives (for reference):**
- Sealed Secrets: Simple, cluster-bound encryption
- SOPS + age: Portable, Git-friendly, key management overhead
- Vault Agent/CSI: Direct injection, no K8s Secrets

---

## 6.3 Policy & Admission Control

**Goal:** Policy-as-code for guardrails and compliance

**Learning objectives:**
- Understand admission controllers and policy engines
- Implement organizational policies at scale

**Tasks:**
- [ ] Create `experiments/policy-tutorial/`
- [ ] Deploy Kyverno (or OPA Gatekeeper)
- [ ] Implement policies:
  - **Security:**
    - [ ] Require non-root containers
    - [ ] Disallow privileged pods
    - [ ] Enforce resource limits
  - **Supply chain (from Phase 2):**
    - [ ] Require signed images
    - [ ] Restrict registries (allowlist)
  - **Operational:**
    - [ ] Require probes
    - [ ] Require labels (owner, cost-center)
- [ ] Policy lifecycle:
  - [ ] Audit mode vs enforce mode
  - [ ] Policy exceptions
  - [ ] Policy testing in CI
- [ ] Document policy patterns
- [ ] **ADR:** Policy engine selection

---

## 6.4 Pod Security & RBAC

**Goal:** Least-privilege workloads and access control

**Tasks:**
- [ ] Create `experiments/pod-security-tutorial/`
- [ ] Pod Security Standards (PSS):
  - [ ] Configure PSA labels (privileged, baseline, restricted)
  - [ ] Restricted security context
  - [ ] Read-only root filesystem
  - [ ] Non-root containers
- [ ] RBAC patterns:
  - [ ] ClusterRole/Role definitions
  - [ ] RoleBinding to groups
  - [ ] Namespace-scoped access
  - [ ] Service account best practices
- [ ] Audit logging:
  - [ ] API server audit policy
  - [ ] Who did what, when

---

## 6.5 NetworkPolicy Basics

**Goal:** Network segmentation with Kubernetes-native controls

**Tasks:**
- [ ] Create `experiments/network-policy-tutorial/`
- [ ] Default deny policies:
  - [ ] Deny all ingress
  - [ ] Deny all egress (allow DNS)
- [ ] Allow patterns:
  - [ ] Service-to-service communication
  - [ ] Namespace isolation
  - [ ] Egress to external CIDRs
- [ ] CNI-specific features (Cilium):
  - [ ] L7 policies
  - [ ] DNS-based egress
- [ ] Test and validate (netshoot)
- [ ] Document network policy patterns

---

## 6.6 WAF & API Protection

**Goal:** Protect applications from common attacks

**Tasks:**
- [ ] Create `experiments/waf-tutorial/`
- [ ] ModSecurity with NGINX Ingress:
  - [ ] OWASP Core Rule Set
  - [ ] Custom rules
  - [ ] False positive tuning
- [ ] API protection:
  - [ ] Request size limits
  - [ ] Rate limiting
  - [ ] JSON schema validation
- [ ] Testing:
  - [ ] OWASP ZAP scanning
  - [ ] SQL injection/XSS test cases
- [ ] Document WAF patterns
- [ ] **ADR:** WAF placement (edge vs cluster)

---

## 6.7 Security Baseline for Benchmarking

**Goal:** Establish secure defaults for all future experiments

**Tasks:**
- [ ] Create security baseline template:
  - [ ] Default NetworkPolicies
  - [ ] Pod Security context template
  - [ ] Required labels policy
  - [ ] Resource limits policy
- [ ] Document secure experiment scaffold
- [ ] Validate Phase 1-5 experiments against baseline

---

## Cross-References

| Topic | Where to Learn More |
|-------|---------------------|
| OAuth/OIDC deep dive | Appendix B: Identity & Authentication |
| PKI/mTLS internals | Appendix C: PKI & Certificate Management |
| SOC/PCI/HIPAA compliance | Appendix D: Compliance & Security Operations |
| Zero trust architecture | Appendix D: Compliance & Security Operations |
| DDoS protection (cloud) | Appendix N: Multi-Cloud & DR |

---

## FinOps Integration

**Cost considerations:**
- OpenBao: Self-hosted vs managed (HashiCorp Cloud)
- WAF: Cloud WAF vs in-cluster ModSecurity
- Policy violations: Audit mode to find issues before enforce

**Metrics to track:**
- Policy violation rate
- Certificate renewal success rate
- Secrets sync latency (ESO)

---

## Experiments Summary

| Experiment | Focus |
|------------|-------|
| `cert-manager-tutorial` | TLS automation |
| `eso-openbao` | Centralized secrets |
| `policy-tutorial` | Kyverno/OPA policies |
| `pod-security-tutorial` | PSS, RBAC |
| `network-policy-tutorial` | Network segmentation |
| `waf-tutorial` | Application protection |

---

## Success Criteria

- [ ] All experiments use HTTPS with valid certificates
- [ ] Secrets never committed to Git (ESO pattern)
- [ ] Baseline policies enforced (audit mode minimum)
- [ ] NetworkPolicies on all namespaces
- [ ] Security baseline documented for future experiments
