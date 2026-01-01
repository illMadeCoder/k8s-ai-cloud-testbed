# ADR-007: Supply Chain Security Strategy

## Status

Accepted

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
│                         │                                         │
│                         ▼                                         │
│  6. Push to GHCR                                                 │
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
│  - Subject: https://github.com/illMadeCoder/illm-k8s-ai-labs/*   │
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
| Scanning | Trivy | SARIF output, GitHub Security integration |
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

## Implementation

### CI Pipeline Steps

```yaml
permissions:
  id-token: write  # OIDC for keyless signing

steps:
  - name: Sign image with Cosign (keyless)
    run: cosign sign --yes ${IMAGE}@${DIGEST}

  - name: Generate SBOM with Syft
    uses: anchore/sbom-action@v0
    with:
      image: ${IMAGE}
      format: spdx-json

  - name: Attest SBOM to image
    run: cosign attest --yes --predicate sbom.spdx.json --type spdxjson ${IMAGE}@${DIGEST}
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
                    subject: "https://github.com/illMadeCoder/illm-k8s-ai-labs/*"
                    issuer: "https://token.actions.githubusercontent.com"
```

### Verification Commands

```bash
# Verify signature
cosign verify \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  --certificate-identity-regexp="github.com/illMadeCoder/illm-k8s-ai-labs" \
  ghcr.io/illmadecoder/cicd-sample:latest

# Verify SBOM attestation
cosign verify-attestation \
  --type spdxjson \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  --certificate-identity-regexp="github.com/illMadeCoder/illm-k8s-ai-labs" \
  ghcr.io/illmadecoder/cicd-sample:latest
```

## Files

```
.github/workflows/
└── cicd-sample.yaml              # CI pipeline with signing

hub/app-of-apps/kind/
├── kyverno.yaml                  # Kyverno Helm deployment
├── kyverno-policies.yaml         # Policy deployment
├── values/
│   └── kyverno.yaml              # Kyverno Helm values
└── manifests/kyverno-policies/
    └── verify-image-signature.yaml  # ClusterPolicy
```

## Future Enhancements

1. **Enforce mode**: Switch policy to block unsigned images
2. **SLSA provenance**: Add build provenance attestations
3. **Vulnerability attestations**: Attest Trivy scan results
4. **Policy exceptions**: Allow system images (kube-system)
5. **Self-hosted Sigstore**: For air-gapped environments

## References

- [Sigstore Documentation](https://docs.sigstore.dev/)
- [Cosign Keyless Signing](https://docs.sigstore.dev/cosign/keyless/)
- [Kyverno Image Verification](https://kyverno.io/docs/writing-policies/verify-images/)
- [SLSA Framework](https://slsa.dev/)
- [SPDX Specification](https://spdx.dev/)

## Decision Date

2026-01-01
