# ADR-007: Supply Chain Security Strategy

## Status

Accepted (Updated 2026-01-01)

## Context

Container supply chain attacks are increasing. We need a defense-in-depth strategy that:

1. Scans images for known vulnerabilities
2. Generates Software Bill of Materials (SBOM)
3. Signs images with verifiable provenance
4. Enforces signature verification at admission

**Constraints:**
- GitHub-hosted CI/CD (GitHub Actions)
- No key management overhead preferred
- Public transparency acceptable (not air-gapped)

## Decision

**Implement a layered supply chain security approach using Sigstore ecosystem tools.**

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    CI/CD PIPELINE (GitHub Actions)               │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  1. Build image (docker/build-push-action)                       │
│                         │                                         │
│                         ▼                                         │
│  2. Sign with Cosign (keyless OIDC)                              │
│     └─→ Fulcio issues short-lived certificate                    │
│     └─→ Rekor records signature in transparency log              │
│                         │                                         │
│                         ▼                                         │
│  3. Generate SBOM with Syft (SPDX format)                        │
│                         │                                         │
│                         ▼                                         │
│  4. Attest SBOM to image (cosign attest)                         │
│                         │                                         │
│                         ▼                                         │
│  5. Scan with Trivy → GitHub Security tab                        │
│     (Fails build on CRITICAL vulnerabilities)                    │
│                         │                                         │
│                         ▼                                         │
│  6. Generate SLSA provenance attestation                         │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                    KUBERNETES ADMISSION                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Kyverno ClusterPolicy: verify-image-signature                   │
│  - Matches: ghcr.io/illmadecoder/*                               │
│  - Verifies: Cosign keyless signature                            │
│  - Issuer: https://token.actions.githubusercontent.com           │
│  - Subject: https://github.com/illMadeCoder/k8s-ai-testbed/*   │
│  - Mode: Audit (logs warnings, future: Enforce)                  │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Tool Selection

| Layer | Tool | Why |
|-------|------|-----|
| Signing | Cosign (keyless) | No key management, OIDC identity, transparency log |
| CA | Fulcio (public) | Free, managed, short-lived certificates |
| Transparency | Rekor (public) | Immutable audit trail, free |
| SBOM | Syft | Open source, SPDX/CycloneDX support, good catalogers |
| Scanning | Trivy | SARIF output, GitHub Security integration, build gating |
| Provenance | GitHub Attestations | SLSA Level 2, native GitHub integration |
| Admission | Kyverno | Native K8s CRDs, image verification built-in |

### Keyless Signing Flow

```
GitHub Actions OIDC Token
         │
         ▼
┌─────────────────┐
│     Fulcio      │ ←─ Verifies OIDC token
│  (Certificate   │    Issues X.509 cert (~10 min)
│   Authority)    │    Cert contains identity claims
└─────────────────┘
         │
         ▼
   Sign image with ephemeral key
   Key thrown away immediately
         │
         ▼
┌─────────────────┐
│     Rekor       │ ←─ Records: signature + cert + timestamp
│  (Transparency  │    Provides proof signature was made
│      Log)       │    while certificate was valid
└─────────────────┘
```

### Trust Model

**What we trust:**
- Sigstore infrastructure (Fulcio CA, Rekor log)
- GitHub Actions OIDC issuer
- GitHub's workflow isolation

**What signatures prove:**
- Image was built by a specific workflow in a specific repo
- Signature was created at a specific time (Rekor timestamp)
- No offline signing possible (no persistent key)

**What signatures DON'T prove:**
- The code is safe or correct
- Dependencies are trustworthy
- The GitHub account wasn't compromised

### Policy Mode

Starting with `validationFailureAction: Audit`:
- Logs policy violations without blocking
- Allows gradual rollout
- Visible in Kyverno PolicyReports

Production recommendation: `Enforce` after validation.

## Consequences

### Positive

- **No key management**: Keyless signing eliminates key rotation, storage, access control
- **Audit trail**: Every signature logged publicly in Rekor
- **Bounded blast radius**: Attacker must trigger CI to sign (leaves evidence)
- **Identity-based trust**: Verify "this came from repo X, workflow Y"
- **SBOM attached**: Consumers can inspect image contents

### Negative

- **External dependency**: Relies on Sigstore public infrastructure uptime
- **Public transparency**: Signing events visible to anyone
- **Complexity**: Multiple tools (Cosign, Fulcio, Rekor, Kyverno)
- **Not air-gapped**: Requires internet access for verification

### Trade-offs

| Approach | Key Management | Offline Signing | Audit Trail |
|----------|----------------|-----------------|-------------|
| **Keyless (chosen)** | None | No | Public (Rekor) |
| Traditional keys | Yes | Yes | None |
| Notation | Yes | Yes | Optional |

### Why SLSA Level 2, Not Level 3

We implement SLSA Level 2 provenance rather than Level 3. Here's the reasoning:

**SLSA Level 3 requires `slsa-github-generator`:**
```yaml
# SLSA 3: Reusable workflow (isolated provenance generation)
provenance:
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@v2.0.0
```

**Problem: Incompatible with matrix builds.**

Our auto-detection workflow builds multiple apps in parallel via matrix strategy. Reusable workflows cannot be called within a matrix, requiring either:
- One workflow per app (defeats auto-detection)
- Complex dynamic workflow generation
- Separate release workflow for SLSA 3

**Cosign Keyless Already Provides Non-Falsifiable Provenance:**

| Property | SLSA 2 + Cosign/Rekor | SLSA 3 |
|----------|----------------------|--------|
| Identity proof (repo, workflow) | ✅ OIDC claims in Fulcio cert | ✅ |
| Immutable audit log | ✅ Rekor transparency log | ✅ |
| Can't falsify provenance | ✅ GitHub controls OIDC issuance | ✅ |
| Isolated provenance generation | ❌ Workflow controls | ✅ |
| Rich build metadata | ❌ Identity only | ✅ |
| Standardized format | ❌ Sigstore-specific | ✅ in-toto |

The core security property—non-falsifiable proof of origin—is already achieved via Cosign keyless signing. The OIDC token is issued by GitHub (not user-controlled), recorded immutably in Rekor, and contains repo/workflow/commit claims.

**When to upgrade to SLSA 3:**
- Tagged releases for production deployment
- Compliance requirements explicitly requiring "SLSA Level 3"
- Apps leaving the lab for external consumption

See [Appendix M: SLSA Deep Dive](../roadmap/appendix-slsa.md) for detailed exploration of SLSA Levels 1-4.

## Implementation

### CI Pipeline Steps

```yaml
permissions:
  id-token: write     # OIDC for keyless signing
  attestations: write # GitHub provenance attestations

steps:
  # 1. Build and push image
  - uses: docker/build-push-action@v6

  # 2. SLSA provenance attestation
  - name: Attest build provenance (SLSA)
    uses: actions/attest-build-provenance@v2
    with:
      subject-name: ${IMAGE}
      subject-digest: ${DIGEST}
      push-to-registry: true

  # 3. Sign image with Cosign (keyless)
  - run: cosign sign --yes ${IMAGE}@${DIGEST}

  # 4. Generate SBOM with Syft
  - uses: anchore/sbom-action@v0
    with:
      image: ${IMAGE}
      format: spdx-json

  # 5. Attest SBOM to image
  - run: cosign attest --yes --predicate sbom.spdx.json --type spdxjson ${IMAGE}@${DIGEST}

  # 6. Scan for vulnerabilities (fails on CRITICAL)
  - uses: aquasecurity/trivy-action@0.28.0
    with:
      exit-code: '1'
      severity: 'CRITICAL'
      ignore-unfixed: true
```

### Kyverno Policy

```yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: verify-image-signature
spec:
  validationFailureAction: Audit
  rules:
    - name: verify-signature
      match:
        any:
          - resources:
              kinds:
                - Pod
      verifyImages:
        - imageReferences:
            - "ghcr.io/illmadecoder/*"
          attestors:
            - entries:
                - keyless:
                    subject: "https://github.com/illMadeCoder/k8s-ai-testbed/*"
                    issuer: "https://token.actions.githubusercontent.com"
```

### Verification Commands

```bash
# Verify signature
cosign verify \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  --certificate-identity-regexp="github.com/illMadeCoder/k8s-ai-testbed" \
  ghcr.io/illmadecoder/hello-app:latest

# Verify SLSA provenance
gh attestation verify ghcr.io/illmadecoder/hello-app:latest --owner illMadeCoder

# Verify SBOM attestation
cosign verify-attestation \
  --type spdxjson \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  --certificate-identity-regexp="github.com/illMadeCoder/k8s-ai-testbed" \
  ghcr.io/illmadecoder/hello-app:latest
```

## Files

```
.github/workflows/
├── build-components.yaml         # Auto-detection CI with signing
└── auto-merge.yaml               # Auto-merge dependency PRs

components/apps/hello-app/
├── Dockerfile                    # Multi-stage build
├── hello-app.yaml                # ArgoCD Application with Image Updater
└── k8s/
    ├── kustomization.yaml        # Required for Image Updater
    ├── deployment.yaml
    └── service.yaml

hub/app-of-apps/kind/
├── kyverno.yaml                  # Kyverno Helm deployment
├── kyverno-policies.yaml         # Policy deployment
└── manifests/kyverno-policies/
    └── verify-image-signature.yaml  # ClusterPolicy
```

## Future Enhancements

1. **Enforce mode**: Switch Kyverno policy to block unsigned images
2. **Vulnerability attestations**: Attest Trivy scan results to image
3. **Policy exceptions**: Allow system images (kube-system)
4. **Self-hosted Sigstore**: For air-gapped environments
5. **SLSA Level 3**: Use slsa-github-generator for isolated builds

## References

- [Sigstore Documentation](https://docs.sigstore.dev/)
- [Cosign Keyless Signing](https://docs.sigstore.dev/cosign/keyless/)
- [Kyverno Image Verification](https://kyverno.io/docs/writing-policies/verify-images/)
- [SLSA Framework](https://slsa.dev/)
- [GitHub Artifact Attestations](https://docs.github.com/en/actions/security-guides/using-artifact-attestations-to-establish-provenance-for-builds)
- [SPDX Specification](https://spdx.dev/)

## Decision Date

2026-01-01 (Updated with SLSA provenance and Trivy gating)
