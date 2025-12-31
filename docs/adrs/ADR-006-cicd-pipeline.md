# ADR-006: CI/CD Pipeline Architecture

## Status

Accepted

## Context

The platform needs a CI/CD pipeline to build, test, and deploy containerized applications. Key requirements:

1. **Build container images** from source code
2. **Push to a container registry** accessible from Kubernetes
3. **Scan for vulnerabilities** before deployment
4. **Update Kubernetes manifests** with new image tags
5. **Support GitOps** workflow where ArgoCD syncs from Git

**Constraints:**
- GitHub-hosted repository
- Need to support both local (Kind) and cloud (AKS/EKS) clusters
- Supply chain security is a priority

## Decision

**Use GitHub Actions with GHCR for CI/CD, with a GitOps manifest update pattern.**

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     DEVELOPER WORKFLOW                       │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  1. Developer pushes code                                    │
│                     │                                         │
│                     ▼                                         │
│  2. GitHub Actions triggers                                  │
│     ┌─────────────────────────────────────────┐              │
│     │  Build → Scan → Push → Update Manifest  │              │
│     └─────────────────────────────────────────┘              │
│                     │                                         │
│                     ▼                                         │
│  3. Manifest committed to Git                                │
│                     │                                         │
│                     ▼                                         │
│  4. ArgoCD detects change → syncs to cluster                 │
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

### Manifest Update Pattern

**Option 1: Git Push from CI** (Chosen)
- CI updates manifest file and pushes
- ArgoCD syncs from Git
- Full audit trail in Git history

**Option 2: ArgoCD Image Updater**
- Watches registry for new tags
- Updates annotations (not Git)
- Less Git history

**Decision**: Git Push pattern for full GitOps audit trail. May add ArgoCD Image Updater later for high-frequency updates.

### Vulnerability Scanning

| Tool | Integration | Output |
|------|-------------|--------|
| Trivy | GitHub Action | SARIF → GitHub Security |
| Grype | CLI | JSON/SARIF |
| Snyk | SaaS | Dashboard |

**Decision**: Trivy for open-source, SARIF integration with GitHub Security tab.

## Implementation

### Workflow Structure

```yaml
permissions:
  contents: write        # Git push for manifest updates
  packages: write        # Push to GHCR
  security-events: write # Upload Trivy SARIF

jobs:
  build:
    steps:
      - Build with docker/build-push-action
      - Scan with aquasecurity/trivy-action
      - Upload SARIF with github/codeql-action
      - Update manifest with sed
      - Commit and push
```

### Build Args for Traceability

```dockerfile
ARG VERSION=dev
ARG BUILD_TIME=unknown

RUN go build -ldflags="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" ...
```

Application exposes `/version` endpoint returning build metadata.

## Consequences

### Positive

- **Audit trail**: Every image change tracked in Git
- **Simplicity**: No external services beyond GitHub
- **Cost**: Free for open source
- **Security**: Trivy scans visible in GitHub Security tab

### Negative

- **CI commits clutter history**: Manifest updates create extra commits
- **Race conditions**: Parallel pushes can conflict (mitigated by single-app workflow)
- **GitHub dependency**: Tightly coupled to GitHub ecosystem

### Trade-offs

| Approach | Audit Trail | Complexity | Real-time |
|----------|-------------|------------|-----------|
| **Git Push (chosen)** | Full | Low | Minutes |
| ArgoCD Image Updater | Partial | Medium | Seconds |
| Flux Image Automation | Full | Higher | Seconds |

## Future Enhancements

1. **Cosign signing** (Phase 2.3): Add `id-token: write` for keyless signing
2. **SBOM generation** (Phase 2.2): Add Syft step
3. **Semantic versioning**: Tag releases with semver
4. **ArgoCD Image Updater**: For high-frequency update scenarios

## Files

```
.github/workflows/
└── cicd-sample.yaml          # Sample CI/CD workflow

experiments/components/apps/cicd-sample/
├── Dockerfile                 # Multi-stage with build args
├── src/
│   ├── main.go               # App with /version endpoint
│   └── go.mod
└── k8s/
    ├── deployment.yaml       # Auto-updated by CI
    └── service.yaml
```

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GHCR Documentation](https://docs.github.com/en/packages)
- [Trivy GitHub Action](https://github.com/aquasecurity/trivy-action)
- [docker/build-push-action](https://github.com/docker/build-push-action)

## Decision Date

2025-12-31
