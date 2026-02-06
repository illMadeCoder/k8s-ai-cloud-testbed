# CI/CD Fundamentals Experiment

Phase 2 of the learning roadmap - CI/CD pipeline with supply chain security for Kubernetes.

## Overview

This experiment demonstrates a production-grade CI/CD pipeline using:
- **GitHub Actions** - Auto-detection build workflow
- **GHCR** - Container registry
- **Cosign** - Keyless image signing (Sigstore)
- **Syft** - SBOM generation
- **Trivy** - Vulnerability scanning with build gating
- **ArgoCD Image Updater** - Continuous deployment

## Architecture

```
Push to main (Dockerfile changes)
         │
         ▼
┌─────────────────────────────────────────────────────────────────┐
│  GitHub Actions (build-components.yaml)                          │
│                                                                   │
│  ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐          │
│  │ Detect  │ → │  Build  │ → │  SLSA   │ → │ Cosign  │          │
│  │ Changed │   │  Image  │   │ Provnce │   │  Sign   │          │
│  └─────────┘   └─────────┘   └─────────┘   └─────────┘          │
│                                     │                             │
│                     ┌───────────────┼───────────────┐            │
│                     ▼               ▼               ▼            │
│              ┌─────────┐     ┌─────────┐     ┌─────────┐        │
│              │  SBOM   │     │  Trivy  │     │  Push   │        │
│              │  Attest │     │  Gate   │     │  GHCR   │        │
│              └─────────┘     └─────────┘     └─────────┘        │
│                                     │                             │
│                              (fails on CRITICAL)                  │
└─────────────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────────┐
│  ArgoCD Image Updater (in-cluster)                               │
│  - Polls GHCR for new SHA tags                                   │
│  - Updates ArgoCD Application                                    │
│  - ArgoCD syncs to cluster                                       │
└─────────────────────────────────────────────────────────────────┘
```

## Components

### Sample Application

Located at `components/apps/hello-app/`:

| File | Purpose |
|------|---------|
| `Dockerfile` | Multi-stage build (golang → distroless) |
| `src/main.go` | Go HTTP server with health, ready, version endpoints |
| `hello-app.yaml` | ArgoCD Application with Image Updater annotations |
| `k8s/kustomization.yaml` | Kustomize base (required for Image Updater) |
| `k8s/deployment.yaml` | Kubernetes Deployment |
| `k8s/service.yaml` | ClusterIP Service |

### Endpoints

| Path | Purpose |
|------|---------|
| `/` | Hello response with pod name |
| `/health` | Liveness probe |
| `/ready` | Readiness probe |
| `/version` | JSON with git SHA and build time |

### GitHub Actions Workflow

Located at `.github/workflows/build-components.yaml`:

**Key features:**
- Auto-detects changed apps (no per-app workflow needed)
- Matrix builds for parallel execution
- Supply chain security: SLSA provenance → Cosign → SBOM → Trivy

### ArgoCD Image Updater

Configured via Application annotations:

```yaml
annotations:
  argocd-image-updater.argoproj.io/image-list: hello-app=ghcr.io/illmadecoder/hello-app
  argocd-image-updater.argoproj.io/hello-app.update-strategy: newest-build
  argocd-image-updater.argoproj.io/hello-app.allow-tags: regexp:^[a-f0-9]{7}$
  argocd-image-updater.argoproj.io/write-back-method: argocd
```

## Running the Experiment

### Prerequisites

- GitHub repository with Actions enabled
- ArgoCD with Image Updater installed
- Kyverno for signature verification (optional)

### Steps

1. **Make a change** to any file in `components/apps/hello-app/`
2. **Push to main** branch
3. **Watch the workflow** at GitHub Actions
4. **Check GHCR** for the new image: `ghcr.io/illmadecoder/hello-app`
5. **Check Security tab** for Trivy scan results
6. **Watch ArgoCD** - Image Updater will detect and deploy

### Verify Supply Chain

```bash
# Verify Cosign signature
cosign verify \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  --certificate-identity-regexp="github.com/illMadeCoder/illm-k8s-ai-labs" \
  ghcr.io/illmadecoder/hello-app:latest

# Verify SLSA provenance
gh attestation verify ghcr.io/illmadecoder/hello-app:latest \
  --owner illMadeCoder

# View SBOM attestation
cosign verify-attestation \
  --type spdxjson \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  --certificate-identity-regexp="github.com/illMadeCoder/illm-k8s-ai-labs" \
  ghcr.io/illmadecoder/hello-app:latest
```

### Deploy Manually (without Image Updater)

```bash
# Apply ArgoCD Application
kubectl apply -f components/apps/hello-app/hello-app.yaml

# Or deploy directly
kubectl apply -k components/apps/hello-app/k8s/

# Verify
kubectl get pods -n hello-app
kubectl port-forward svc/hello-app -n hello-app 8080:80
curl localhost:8080/version
```

## Learning Objectives

After completing this experiment, you should understand:

1. **Auto-detection CI** - Single workflow for multiple apps
2. **Supply chain security** - SLSA, Cosign, SBOM, Trivy
3. **Keyless signing** - OIDC, Fulcio, Rekor transparency log
4. **GitOps deployment** - Image Updater vs git-push patterns
5. **Build gating** - Failing builds on critical vulnerabilities

## Related Documentation

- [ADR-006: CI/CD Pipeline](../../../docs/adrs/ADR-006-cicd-pipeline.md)
- [ADR-007: Supply Chain Security](../../../docs/adrs/ADR-007-supply-chain-security.md)
- [Appendix M: SLSA Deep Dive](../../../docs/roadmap/appendix-slsa.md)
