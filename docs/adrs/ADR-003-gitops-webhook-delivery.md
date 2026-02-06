# ADR-003: GitOps Webhook Delivery to Home Lab

## Status

Proposed

## Context

The hub cluster uses ArgoCD for GitOps. When experiments are added or modified in Git, ArgoCD needs to sync. Two approaches:

1. **Polling**: ArgoCD polls Git every N minutes (default 3 min)
2. **Webhook**: GitHub sends webhook on push, ArgoCD syncs immediately

We prefer webhook for immediate feedback, but the hub cluster may run:
- On a laptop (Kind) - no stable IP
- On home lab (N100) - behind NAT, not internet-exposed
- On cloud (AKS/EKS) - publicly reachable

**Challenge:** How does GitHub's webhook reach a hub behind NAT?

## Decision

**Use Cloudflare Tunnel** for webhook delivery to non-cloud hub clusters.

## Comparison

| Approach | Inbound Required | Stable URL | Free Tier | Complexity | Production Ready |
|----------|------------------|------------|-----------|------------|------------------|
| **Polling only** | No | N/A | Yes | Lowest | Yes, but delayed |
| **Cloudflare Tunnel** | No (outbound tunnel) | Yes | Yes | Low | Yes |
| **Tailscale Funnel** | No (outbound) | Yes | Yes (limited) | Low | Yes |
| **ngrok** | No (outbound) | Yes (random unless paid) | Limited | Low | Development |
| **smee.io** | No (client polls smee) | Via proxy | Yes | Low | Development |
| **Port forwarding** | Yes (router config) | If static IP | N/A | Medium | Fragile |
| **VPN + static IP** | Depends | Yes | Varies | High | Yes |

## Why Cloudflare Tunnel

- **No inbound connectivity** - tunnel is outbound from hub
- **Free** - no cost for personal use
- **Stable URL** - `hub.yourdomain.com` or `*.cfargotunnel.com`
- **Runs as pod** - `cloudflared` deploys to hub via GitOps
- **Production grade** - used in enterprise environments
- **Works everywhere** - Kind, K3s, Talos, cloud

## Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                           GitHub                                      │
│  1. git push                                                         │
│  2. Webhook fires to https://hub.example.com/api/webhook             │
└──────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌──────────────────────────────────────────────────────────────────────┐
│                      Cloudflare Edge                                  │
│  3. Routes request through tunnel                                    │
└──────────────────────────────────────────────────────────────────────┘
                                    │
                          (outbound tunnel)
                                    │
┌──────────────────────────────────────────────────────────────────────┐
│  Hub Cluster (behind NAT)                                            │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │  cloudflared (Deployment)                                       │ │
│  │  - Maintains outbound tunnel to Cloudflare                      │ │
│  │  - Routes incoming requests to ArgoCD service                   │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                              │                                       │
│                              ▼                                       │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │  ArgoCD                                                         │ │
│  │  4. Receives webhook                                            │ │
│  │  5. Triggers sync for affected Applications                     │ │
│  │  6. ApplicationSet generates/updates experiment orchestrators   │ │
│  └────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────┘
```

## Configuration

**Bootstrap order (dependencies matter):**
1. ArgoCD deployed (imperative bootstrap)
2. OpenBao deployed via ArgoCD (sync wave 1)
3. OpenBao initialized and unsealed
4. Cloudflare Tunnel token stored in OpenBao
5. cloudflared deployed via ArgoCD (sync wave 2, pulls token from OpenBao via ESO)
6. GitHub webhook configured

**Cloudflare Tunnel setup:**
1. Create tunnel in Cloudflare Zero Trust dashboard (or via API)
2. Get tunnel token
3. Store token in OpenBao: `bao kv put secret/cloudflared token=<token>`
4. Deploy `cloudflared` to hub via ArgoCD (ExternalSecret pulls token from OpenBao)

**ArgoCD webhook configuration:**
- GitHub webhook URL: `https://hub.example.com/api/webhook`
- Webhook secret: stored in OpenBao, synced via ESO

**For cloud-hosted hub:**
- Cloudflare Tunnel optional (LoadBalancer exposes ArgoCD directly)
- Can still use tunnel for consistent setup across environments

## Alternatives Considered

**Polling only:**
- Simpler, no tunnel needed
- 3-minute delay acceptable for learning lab
- Could be fallback if tunnel fails

**Tailscale Funnel:**
- Requires Tailscale account and setup
- Good if already using Tailscale for other purposes
- Less universal than Cloudflare

**smee.io:**
- Development-focused, not production grade
- Adds intermediary dependency
- Good for quick testing

## Consequences

**Positive:**
- Immediate GitOps sync on push
- No inbound firewall rules needed
- Consistent experience across all hub environments
- Production-grade solution

**Negative:**
- Cloudflare account required (free)
- Additional component to manage (cloudflared)
- Tunnel token is a secret to manage

## Fallback

If tunnel is unavailable, ArgoCD falls back to polling (3-minute default). Experiments still work, just with slight delay.

## References

- [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/)
- [ArgoCD Webhook Configuration](https://argo-cd.readthedocs.io/en/stable/operator-manual/webhook/)
- [cloudflared Helm Chart](https://github.com/cloudflare/helm-charts)

## Decision Date

2025-12-12
