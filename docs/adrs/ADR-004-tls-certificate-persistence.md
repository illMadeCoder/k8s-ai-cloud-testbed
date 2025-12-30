# ADR-004: TLS Certificate Persistence via OpenBao

## Status

Accepted

## Context

Let's Encrypt has strict rate limits: 5 certificates per week for the same domain set. In a development environment where clusters are frequently recreated (Kind clusters for testing), each recreation triggers a new certificate request, quickly exhausting rate limits.

Additionally, zero-trust internal communication requires certificates that include both external domains (for browser access) and internal service names (for pod-to-pod TLS).

**Problems to solve:**
1. Avoid Let's Encrypt rate limits on cluster recreation
2. Enable zero-trust internal TLS with proper certificate validation
3. Automate certificate lifecycle without manual intervention

## Decision

**Persist TLS certificates in OpenBao and restore them on cluster recreation using External Secrets Operator (ESO).**

### Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                    CERTIFICATE ISSUANCE / RENEWAL                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  cert-manager ──────▶ argocd-server-tls-letsencrypt                  │
│  (Let's Encrypt)                 │                                    │
│                                  │ PushSecret (automatic)             │
│                                  ▼                                    │
│                              OpenBao ◀─── ~/.illmlab/openbao-data    │
│                         secret/tls/argocd    (persistent storage)    │
│                                  │                                    │
│                                  │ ExternalSecret (automatic)         │
│                                  ▼                                    │
│                         argocd-server-tls ──▶ ArgoCD                 │
│                                                                       │
├─────────────────────────────────────────────────────────────────────┤
│                      CLUSTER RECREATION                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│                              OpenBao                                 │
│                    (data persists on host)                           │
│                                  │                                    │
│              ┌───────────────────┼───────────────────┐                │
│              │ ExternalSecret    │ ExternalSecret    │                │
│              ▼                   ▼                   │                │
│  argocd-server-tls-letsencrypt  argocd-server-tls    │                │
│  (cert-manager sees valid       (ArgoCD uses this)   │                │
│   cert → no new request!)                            │                │
│                                                                       │
└─────────────────────────────────────────────────────────────────────┘
```

### Components

| Component | Purpose | File |
|-----------|---------|------|
| **Certificate** | Requests cert from Let's Encrypt | `cert-manager-config/argocd-certificate.yaml` |
| **PushSecret** | Pushes cert to OpenBao on issuance/renewal | `external-secrets-config/argocd-tls-push-secret.yaml` |
| **ExternalSecret (letsencrypt)** | Restores cert-manager's secret from OpenBao | `external-secrets-config/argocd-tls-letsencrypt-external-secret.yaml` |
| **ExternalSecret (argocd)** | Syncs production secret from OpenBao | `external-secrets-config/argocd-tls-external-secret.yaml` |

### Key Implementation Details

1. **Separate secrets**: cert-manager writes to `argocd-server-tls-letsencrypt`, ArgoCD uses `argocd-server-tls`. This decoupling allows ESO to manage the production secret.

2. **cert-manager annotations**: The ExternalSecret that restores `argocd-server-tls-letsencrypt` includes cert-manager annotations so cert-manager recognizes it as a valid, already-issued certificate:
   ```yaml
   template:
     metadata:
       annotations:
         cert-manager.io/issuer-name: letsencrypt-prod
         cert-manager.io/issuer-kind: ClusterIssuer
   ```

3. **Persistent OpenBao storage**: OpenBao data is mounted from `~/.illmlab/openbao-data`, surviving cluster deletion.

4. **Automatic sync**: No manual commands needed. PushSecret and ExternalSecret handle the bidirectional sync automatically.

### Internal TLS (Zero-Trust)

For internal service-to-service communication, OpenBao PKI can issue certificates with all required SANs:

```yaml
dnsNames:
  - argocd.illmlab.xyz              # external domain
  - argocd-server                    # short name
  - argocd-server.argocd             # namespace
  - argocd-server.argocd.svc         # svc
  - argocd-server.argocd.svc.cluster.local  # FQDN
```

This allows internal services (like webhook-relay) to connect via HTTPS using internal service names with proper TLS verification.

## Consequences

### Positive

- **No rate limits**: Cluster recreation restores existing cert, no new Let's Encrypt request
- **Zero-trust internal TLS**: Certificates include internal service names as SANs
- **Fully automated**: PushSecret/ExternalSecret handle sync without manual intervention
- **Separation of concerns**: cert-manager handles issuance, ESO handles distribution

### Negative

- **Complexity**: Multiple ESO resources to manage the flow
- **Bootstrap chicken-egg**: First cluster creation needs Let's Encrypt (OpenBao empty), subsequent recreations use cached cert
- **cert-manager status**: Certificate resource shows `False` after restore (cosmetic - cert is valid, renewal works)

### Trade-offs

| Approach | Rate Limit Safe | Automation | Complexity |
|----------|-----------------|------------|------------|
| cert-manager only | No | Full | Low |
| Manual backup to OpenBao | Yes | Manual | Medium |
| **PushSecret + ExternalSecret** | Yes | Full | Higher |

## Alternatives Considered

1. **Use OpenBao PKI for everything**: Simpler, but browsers show certificate warnings (not publicly trusted CA).

2. **Manual backup command**: Added `task kind:cert-backup` but requires human intervention.

3. **Use Let's Encrypt staging**: No rate limits but not browser-trusted.

## Files

```
hub/app-of-apps/kind/manifests/
├── cert-manager-config/
│   ├── argocd-certificate.yaml      # Let's Encrypt Certificate
│   ├── cluster-issuer.yaml          # Let's Encrypt ClusterIssuers
│   └── openbao-issuer.yaml          # OpenBao PKI ClusterIssuer
└── external-secrets-config/
    ├── argocd-tls-external-secret.yaml           # OpenBao → argocd-server-tls
    ├── argocd-tls-letsencrypt-external-secret.yaml  # OpenBao → argocd-server-tls-letsencrypt
    └── argocd-tls-push-secret.yaml               # argocd-server-tls-letsencrypt → OpenBao
```

## References

- [cert-manager Certificate resources](https://cert-manager.io/docs/usage/certificate/)
- [ESO PushSecret](https://external-secrets.io/latest/api/pushsecret/)
- [Let's Encrypt Rate Limits](https://letsencrypt.org/docs/rate-limits/)
- [ADR-002: Secrets Management](./ADR-002-secrets-management.md)

## Decision Date

2025-12-30
