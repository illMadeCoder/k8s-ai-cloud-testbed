# CI/CD Fundamentals Experiment

Phase 2.1 of the learning roadmap - establishing CI/CD pipeline patterns for Kubernetes.

## Overview

This experiment demonstrates a complete CI/CD pipeline using GitHub Actions to:
- Build a Go application as a container image
- Push to GitHub Container Registry (GHCR)
- Scan for vulnerabilities with Trivy
- Auto-update Kubernetes manifests with new image tags

## Components

### Sample Application

Located at `experiments/components/apps/cicd-sample/`:

| File | Purpose |
|------|---------|
| `src/main.go` | Go HTTP server with health, ready, and version endpoints |
| `src/go.mod` | Go module definition |
| `Dockerfile` | Multi-stage build with distroless runtime |
| `k8s/deployment.yaml` | Kubernetes Deployment with probes |
| `k8s/service.yaml` | ClusterIP Service |

### Endpoints

| Path | Purpose |
|------|---------|
| `/` | Hello response with pod name |
| `/health` | Liveness probe |
| `/ready` | Readiness probe |
| `/version` | JSON with git SHA and build time |

### GitHub Actions Workflow

Located at `.github/workflows/cicd-sample.yaml`:

```
Push to main (cicd-sample changes)
         │
         ▼
┌─────────────────────┐
│  Build & Push       │
│  (docker buildx)    │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Trivy Scan         │
│  (vulnerability)    │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Update Manifest    │
│  (sed + git push)   │
└─────────────────────┘
```

## Running the Experiment

### Prerequisites

- GitHub repository with Actions enabled
- GHCR access (automatic with GITHUB_TOKEN)

### Steps

1. **Make a change** to any file in `experiments/components/apps/cicd-sample/`
2. **Push to main** branch
3. **Watch the workflow** at GitHub Actions
4. **Check GHCR** for the new image: `ghcr.io/<owner>/cicd-sample`
5. **Check Security tab** for Trivy scan results
6. **Verify manifest update** in `k8s/deployment.yaml`

### Deploying to Kubernetes

```bash
# Deploy the app
kubectl apply -f experiments/components/apps/cicd-sample/k8s/

# Verify pods are running
kubectl get pods -n cicd-sample

# Test the endpoints
kubectl port-forward svc/cicd-sample -n cicd-sample 8080:80
curl localhost:8080/version
```

## Learning Objectives

After completing this experiment, you should understand:

1. **GitHub Actions workflow structure** - jobs, steps, permissions
2. **Docker buildx** - multi-stage builds, build args
3. **GHCR authentication** - using GITHUB_TOKEN
4. **Image tagging** - SHA-based tags for traceability
5. **Trivy scanning** - vulnerability detection, SARIF format
6. **GitOps pattern** - auto-updating manifests with image tags

## Next Steps

- **Phase 2.2**: Add SBOM generation with Syft
- **Phase 2.3**: Add Cosign image signing
- **Phase 2.4**: Explore ArgoCD Image Updater for automatic manifest updates

## Related ADRs

- [ADR-006: CI/CD Pipeline](../../../docs/adrs/ADR-006-cicd-pipeline.md)
