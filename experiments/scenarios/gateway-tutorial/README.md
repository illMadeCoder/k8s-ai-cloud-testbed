# Gateway Tutorial: From Ingress to Gateway API

Master Kubernetes L7 traffic management, from legacy Ingress through modern Gateway API.

## Overview

This tutorial teaches the evolution of Kubernetes traffic management:

| Part | Topic | Key Learnings |
|------|-------|---------------|
| 1 | Ingress Basics | Path routing, host routing, TLS termination |
| 2 | Ingress Limitations | Annotation hell, portability problems |
| 3 | Gateway API Migration | Clean configs, portable patterns |
| 4 | Gateway API Deep Dive | Advanced routing, policies, multi-gateway |
| 5 | gRPC Traffic Management | HTTP/2 challenges, GRPCRoute |

## Quick Start

```bash
# Start the tutorial
task hub:experiment -- gateway-tutorial

# Or run interactively
task hub:conduct -- gateway-tutorial
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  Client                                                          │
└─────────────────────────────────────────────────────────────────┘
                            │
            ┌───────────────┴───────────────┐
            ▼                               ▼
┌───────────────────────┐     ┌───────────────────────┐
│   nginx-ingress       │     │   Envoy Gateway       │
│   (Ingress)           │     │   (Gateway API)       │
└───────────────────────┘     └───────────────────────┘
            │                               │
    ┌───────┴───────┐               ┌───────┴───────┐
    ▼               ▼               ▼               ▼
┌───────┐     ┌───────┐       ┌───────┐       ┌───────┐
│echo-v1│     │echo-v2│       │  api  │       │  web  │
└───────┘     └───────┘       └───────┘       └───────┘
```

## Demo Services

| Service | Purpose | Endpoints |
|---------|---------|-----------|
| `echo-v1` | Version routing demos | Returns "Hello from echo-v1" |
| `echo-v2` | Canary/splitting demos | Returns "Hello from echo-v2" |
| `api-service` | API routing demos | httpbin endpoints (/get, /post, /headers, etc.) |
| `web-service` | Web routing demos | Static nginx content |
| `slow-service` | Timeout demos | httpbin /delay endpoint |

## Part-by-Part Guide

### Part 1: Ingress Basics

**Path-Based Routing**
```yaml
spec:
  rules:
    - http:
        paths:
          - path: /api
            backend:
              service:
                name: api-service
```

**Host-Based Routing**
```yaml
spec:
  rules:
    - host: api.example.com
      http:
        paths:
          - path: /
            backend:
              service:
                name: api-service
```

### Part 2: Ingress Limitations

**The Annotation Problem**

Each feature requires controller-specific annotations:

| Feature | nginx | traefik | kong |
|---------|-------|---------|------|
| Rate Limit | `nginx.ingress.kubernetes.io/limit-rps` | `traefik.ingress.kubernetes.io/rate-limit` | `konghq.com/plugins` |
| Backend Protocol | `backend-protocol: "GRPC"` | `service.serversscheme: h2c` | `protocols: grpc` |
| Canary | `canary: "true"`, `canary-weight` | `traefik.ingress.kubernetes.io/service-weights` | `konghq.com/override` |

**Traffic Splitting (Awkward)**

Requires TWO Ingress resources:
```yaml
# Primary Ingress
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app-primary
spec:
  rules: [...]

---
# Canary Ingress (separate resource!)
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app-canary
  annotations:
    nginx.ingress.kubernetes.io/canary: "true"
    nginx.ingress.kubernetes.io/canary-weight: "20"
spec:
  rules: [...]  # Same rules, different backend
```

### Part 3: Gateway API Migration

**Gateway API Resource Hierarchy**
```
GatewayClass (infra admin)
    │
    ▼
Gateway (cluster operator)
    │
    ▼
HTTPRoute / GRPCRoute (app developer)
```

**Traffic Splitting (Clean)**
```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
spec:
  rules:
    - backendRefs:
        - name: app-v1
          weight: 80
        - name: app-v2
          weight: 20
```

### Part 4: Gateway API Deep Dive

**Matching Options**
- Path: Exact, PathPrefix, RegularExpression
- Headers: Name/value matching
- Query Parameters: Name/value matching
- Method: GET, POST, PUT, DELETE, etc.

**Traffic Manipulation**
- Request/Response header modification
- URL rewriting
- Redirects
- Request mirroring

**Policies (Envoy Gateway Extensions)**
- BackendTrafficPolicy: Timeouts, retries, circuit breaking
- SecurityPolicy: CORS, JWT auth, mTLS
- RateLimitPolicy: Request rate limiting

### Part 5: gRPC Traffic Management

**The HTTP/2 Load Balancing Problem**

HTTP/2 multiplexes requests over a single connection. Traditional L4 load balancers send ALL requests to ONE backend.

```
L4 LB (ClusterIP):
Client ──[connection]──► Pod A (gets 100%)
                         Pod B (gets 0%)
                         Pod C (gets 0%)

L7 LB (Gateway):
Client ──[connection]──► Gateway ──► Pod A (33%)
                                 ──► Pod B (33%)
                                 ──► Pod C (33%)
```

**GRPCRoute**
```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
spec:
  rules:
    - matches:
        - method:
            service: myapp.UserService
            method: GetUser
      backendRefs:
        - name: user-service
```

**gRPC-Web for Browsers**

Browsers can't use native gRPC (no HTTP/2 trailer support). gRPC-Web provides a workaround:

```
Browser ──[gRPC-Web]──► Gateway ──[native gRPC]──► Backend
          (HTTP/1.1)              (HTTP/2)
```

## Files Reference

```
gateway-tutorial/
├── target/
│   ├── cluster.yaml                    # Cluster config
│   └── argocd/
│       ├── app.yaml                    # ArgoCD Application
│       ├── nginx-ingress-values.yaml   # nginx-ingress config
│       └── envoy-gateway-values.yaml   # Envoy Gateway config
├── manifests/
│   ├── demo-apps.yaml                  # Demo services
│   ├── part1-ingress-basics.yaml       # Part 1 resources
│   ├── part2-ingress-limitations.yaml  # Part 2 resources
│   ├── part3-gateway-api.yaml          # Part 3 resources
│   ├── part4-gateway-deep-dive.yaml    # Part 4 resources
│   ├── part4-tls-multi-gateway.yaml    # Part 4 TLS
│   ├── part5-grpc-services.yaml        # Part 5 gRPC services
│   ├── part5-grpc-ingress.yaml         # Part 5 gRPC Ingress
│   ├── part5-grpc-gateway-api.yaml     # Part 5 GRPCRoute
│   └── part5-grpc-advanced.yaml        # Part 5 advanced
├── workflow/
│   └── experiment.yaml                 # Argo Workflow
├── tutorial.yaml                       # Interactive tutorial
└── README.md                           # This file
```

## Key Takeaways

1. **Ingress is limited** - Relies on controller-specific annotations
2. **Gateway API is portable** - Same config works across implementations
3. **Role separation** - GatewayClass/Gateway/Route hierarchy
4. **gRPC needs L7** - HTTP/2 breaks L4 load balancing
5. **GRPCRoute is native** - First-class gRPC routing support

## Related Resources

- [Gateway API Documentation](https://gateway-api.sigs.k8s.io/)
- [Envoy Gateway](https://gateway.envoyproxy.io/)
- [nginx-ingress](https://kubernetes.github.io/ingress-nginx/)
- [gRPC Load Balancing](https://grpc.io/blog/grpc-load-balancing/)

## Next Steps

After completing this tutorial:
1. **gateway-comparison** - Compare nginx, Traefik, Envoy Gateway
2. **cloud-gateway-comparison** - Compare cloud-native gateways (ALB, AGIC)
