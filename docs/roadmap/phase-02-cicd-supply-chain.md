## Phase 2: CI/CD & Supply Chain Security

**Status: Complete** (2026-01-01)

*You need to build and deploy images before anything else. Supply chain security is baked in from day one.*

### Implementation Summary

| Section | Status | Notes |
|---------|--------|-------|
| 2.1 CI/CD Fundamentals | ✅ Done | Auto-detection workflow, Image Updater, ADR-006 |
| 2.2 Image Security | ✅ Done | Trivy (gated), Syft SBOM, Renovate |
| 2.3 Signing & Verification | ✅ Done | Cosign, SLSA L2, Kyverno, ADR-007 |
| 2.4 Registry Management | ⏭️ Skipped | GHCR sufficient for lab |
| 2.5 Testing Strategies | ⏭️ Skipped | Experiments are the tests; not needed for lab |

**Key artifacts:**
- `.github/workflows/build-components.yaml` - Auto-detection CI
- `.github/workflows/auto-merge.yaml` - Dependency auto-merge
- `renovate.json` - Dependency updates
- `docs/adrs/ADR-006-cicd-pipeline.md`
- `docs/adrs/ADR-007-supply-chain-security.md`
- `docs/roadmap/appendix-slsa.md` - SLSA deep dive

---

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

### 2.5 Testing Strategies for Kubernetes

**Goal:** Implement comprehensive testing for Kubernetes deployments

**Learning objectives:**
- Understand testing pyramid for cloud-native apps
- Implement integration and e2e testing in CI/CD
- Test Kubernetes manifests and configurations

**Tasks:**
- [ ] Create `experiments/testing-strategies/`
- [ ] Unit and integration testing:
  - [ ] Application unit tests in CI
  - [ ] Integration tests with testcontainers
  - [ ] Database/service mocking strategies
- [ ] Kubernetes manifest testing:
  - [ ] Dry-run validation (`kubectl --dry-run`)
  - [ ] Schema validation (kubeconform)
  - [ ] Policy testing (conftest/OPA)
  - [ ] Helm chart testing (helm unittest, helm test)
- [ ] End-to-end testing:
  - [ ] Deploy to ephemeral namespace
  - [ ] Run e2e tests (Cypress, Playwright, k6)
  - [ ] Cleanup after tests
- [ ] Contract testing:
  - [ ] API contract testing (Pact)
  - [ ] Schema compatibility checks
  - [ ] Consumer-driven contracts
- [ ] Infrastructure testing:
  - [ ] Terratest for Terraform
  - [ ] Test cluster provisioning
  - [ ] Smoke tests post-deployment
- [ ] Test environments:
  - [ ] Ephemeral preview environments per PR
  - [ ] Namespace-per-branch strategy
  - [ ] Test data management
- [ ] Document testing patterns
- [ ] **ADR:** Document testing strategy

---

