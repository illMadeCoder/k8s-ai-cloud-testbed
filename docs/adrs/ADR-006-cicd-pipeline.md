# ADR-006: CI/CD Pipeline Architecture

## Status

Accepted (Updated 2026-01-01)

## Context

The platform needs a CI/CD pipeline to build, test, and deploy containerized applications. Key requirements:

1. **Build container images** from source code
2. **Push to a container registry** accessible from Kubernetes
3. **Scan for vulnerabilities** before deployment
4. **Automatic deployment** without CI needing cluster access
5. **Scale to hundreds of apps** without per-app workflow configuration

**Constraints:**
- GitHub-hosted repository
- CI should not require cluster credentials (separation of concerns)
- Need to support both local (Kind) and cloud (AKS/EKS) clusters
- Supply chain security is a priority

## Decision

**Use GitHub Actions with GHCR for CI, with ArgoCD Image Updater for continuous deployment.**

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     CI/CD ARCHITECTURE                       │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  1. Developer pushes code to components/         │
│                     │                                         │
│                     ▼                                         │
│  2. GitHub Actions (build-components.yaml)                   │
│     ┌─────────────────────────────────────────┐              │
│     │  Detect → Build → Sign → SBOM → Scan    │              │
│     └─────────────────────────────────────────┘              │
│                     │                                         │
│                     ▼                                         │
│  3. Image pushed to GHCR with SHA tag                        │
│                     │                                         │
│        ┌───────────┴───────────┐                             │
│        ▼                       ▼                              │
│  ┌──────────────┐    ┌─────────────────────┐                 │
│  │ Rekor Log    │    │ ArgoCD Image        │                 │
│  │ (signature)  │    │ Updater (in-cluster)│                 │
│  └──────────────┘    └─────────────────────┘                 │
│                              │                                │
│                              ▼                                │
│  4. Image Updater detects new SHA tag                        │
│     Updates ArgoCD Application annotation                    │
│                              │                                │
│                              ▼                                │
│  5. ArgoCD syncs new image to cluster                        │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

### CI Platform: GitHub Actions

| Feature | GitHub Actions | GitLab CI |
|---------|----------------|-----------|
| Repository integration | Native | Requires mirror |
| Container registry | GHCR (free) | GitLab Registry |
| Secrets management | GitHub Secrets | CI Variables |
| Self-hosted runners | Supported | Supported |
| Cost | Free for public repos | Free tier limited |

**Decision**: GitHub Actions for tight integration with GitHub repository.

### Container Registry: GHCR

| Feature | GHCR | Docker Hub | ACR/ECR |
|---------|------|------------|---------|
| Authentication | GITHUB_TOKEN | Separate creds | Cloud IAM |
| Cost | Free for public | Rate limits | Cloud billing |
| Proximity | GitHub integration | External | Cloud-native |

**Decision**: GHCR for seamless GitHub integration and free tier.

### Image Tagging Strategy

| Tag Type | Format | Purpose |
|----------|--------|---------|
| SHA | `sha-abc1234` | Immutable, traceable to commit |
| Latest | `latest` | Current main branch |
| Semver | `v1.2.3` (future) | Release versioning |

**Decision**: SHA-based tags for immutability; `latest` for convenience.

### Deployment Pattern: ArgoCD Image Updater

**Why Image Updater over Git Push:**

| Aspect | Git Push from CI | Image Updater (Chosen) |
|--------|------------------|------------------------|
| CI cluster access | Requires kubeconfig or git creds | None - CI is decoupled |
| Audit trail | Git commits | ArgoCD Application annotations |
| Scaling | One workflow per app | Single workflow, auto-detect |
| Complexity | CI needs git operations | Image Updater handles updates |
| Real-time | Minutes (commit → sync) | Seconds (registry poll) |

**Decision**: ArgoCD Image Updater with `argocd` write-back method.
- CI pushes to GHCR with SHA tags
- Image Updater polls registry, detects new SHA tags
- Updates ArgoCD Application directly (no git push needed)
- ArgoCD syncs the new image

**Configuration:**
```yaml
# ArgoCD Application annotations
argocd-image-updater.argoproj.io/image-list: app=ghcr.io/illmadecoder/app
argocd-image-updater.argoproj.io/app.update-strategy: newest-build
argocd-image-updater.argoproj.io/app.allow-tags: regexp:^[a-f0-9]{7}$
argocd-image-updater.argoproj.io/write-back-method: argocd
```

### Vulnerability Scanning

| Tool | Integration | Output |
|------|-------------|--------|
| Trivy | GitHub Action | SARIF → GitHub Security |
| Grype | CLI | JSON/SARIF |
| Snyk | SaaS | Dashboard |

**Decision**: Trivy for open-source, SARIF integration with GitHub Security tab.

## Implementation

### Auto-Detection Workflow

The workflow automatically detects which apps changed and builds them in parallel:

```yaml
# .github/workflows/build-components.yaml
on:
  push:
    paths:
      - 'components/**/Dockerfile'
      - 'components/**/src/**'
      - 'components/**/go.mod'
      # ... other dependency files

jobs:
  detect:
    # Finds changed app directories with Dockerfiles
    outputs:
      apps: ${{ steps.detect.outputs.apps }}

  build:
    strategy:
      matrix:
        app: ${{ fromJson(needs.detect.outputs.apps) }}
    # Builds each app in parallel
```

### Workflow Permissions

```yaml
permissions:
  packages: write        # Push to GHCR
  security-events: write # Upload Trivy SARIF
  id-token: write        # OIDC for Cosign keyless signing
```

Note: `contents: write` is NOT needed - CI doesn't push to git.

### Build Args for Traceability

```dockerfile
ARG VERSION=dev
ARG BUILD_TIME=unknown

RUN go build -ldflags="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" ...
```

Application exposes `/version` endpoint returning build metadata.

## Consequences

### Positive

- **Scalable**: One workflow handles hundreds of apps via auto-detection
- **Decoupled**: CI has no cluster access, no git credentials needed
- **Fast**: Image Updater polls registry, deploys in seconds
- **Secure**: Supply chain security built-in (signing, SBOM, scanning)
- **Cost**: Free for open source

### Negative

- **Partial audit trail**: Image updates in ArgoCD annotations, not git commits
- **Polling delay**: Image Updater polls every 2 minutes by default
- **GitHub dependency**: Tightly coupled to GitHub ecosystem

### Trade-offs

| Approach | Audit Trail | CI Complexity | Deployment Speed |
|----------|-------------|---------------|------------------|
| Git Push from CI | Full (git) | High (git ops) | Minutes |
| **Image Updater (chosen)** | Partial (annotations) | Low | Seconds |
| Flux Image Automation | Full (git) | Medium | Seconds |

## Files

```
.github/workflows/
├── build-components.yaml     # Auto-detection build workflow
└── auto-merge.yaml           # Auto-merge dependency PRs

components/apps/
├── _template/                 # Template for new apps
│   ├── Dockerfile
│   └── k8s/
└── hello-app/                 # Example app
    ├── Dockerfile
    ├── hello-app.yaml         # ArgoCD Application with Image Updater annotations
    └── k8s/
        ├── kustomization.yaml # Required for Image Updater
        ├── deployment.yaml
        └── service.yaml

hub/app-of-apps/kind/
└── argocd-image-updater.yaml  # Image Updater deployment
```

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GHCR Documentation](https://docs.github.com/en/packages)
- [ArgoCD Image Updater](https://argocd-image-updater.readthedocs.io/)
- [Trivy GitHub Action](https://github.com/aquasecurity/trivy-action)
- [docker/build-push-action](https://github.com/docker/build-push-action)

## Decision Date

2025-12-31 (Updated 2026-01-01)
