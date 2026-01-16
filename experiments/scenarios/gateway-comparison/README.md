# Gateway Comparison: nginx-ingress vs Traefik vs Envoy Gateway

Side-by-side comparison of three Kubernetes ingress/gateway controllers.

## Overview

This scenario deploys identical routing configurations on three gateway implementations:

| Gateway | Type | Gateway API Support | Best For |
|---------|------|---------------------|----------|
| nginx-ingress | Ingress Controller | Partial | Legacy, wide adoption |
| Traefik | Ingress + Gateway API | Yes (experimental) | Kubernetes-native, middleware |
| Envoy Gateway | Gateway API | Full (reference impl) | Modern, gRPC, policies |

## Quick Start

```bash
# Start the scenario
task hub:experiment -- gateway-comparison

# Or run interactively
task hub:conduct -- gateway-comparison
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Client                                                                      │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        ▼                           ▼                           ▼
┌───────────────────┐   ┌───────────────────┐   ┌───────────────────┐
│  nginx-ingress    │   │     Traefik       │   │  Envoy Gateway    │
│  (Ingress)        │   │  (Gateway API)    │   │  (Gateway API)    │
│  ingress-nginx    │   │  traefik-system   │   │  gateway-comparison│
└───────────────────┘   └───────────────────┘   └───────────────────┘
        │                           │                           │
        └───────────────────────────┼───────────────────────────┘
                                    ▼
                    ┌───────────────────────────────┐
                    │  Shared Backend Services      │
                    │  (gateway-comparison)         │
                    │  echo-v1, echo-v2, api, web   │
                    │  grpc-echo, grpc-echo-v2      │
                    └───────────────────────────────┘
```

## Features Compared

| Feature | nginx-ingress | Traefik | Envoy Gateway |
|---------|---------------|---------|---------------|
| Gateway API HTTPRoute | Partial | Yes | Yes |
| Gateway API GRPCRoute | No | Experimental | Yes |
| Native Ingress Support | Yes | Yes | No |
| Path Routing | Ingress | HTTPRoute | HTTPRoute |
| Host Routing | Ingress | HTTPRoute | HTTPRoute |
| Rate Limiting | Annotation | Middleware (CRD) | BackendTrafficPolicy |
| Header Manipulation | Snippet annotation | HTTPRoute filter | HTTPRoute filter |
| Traffic Splitting | Canary annotation | HTTPRoute weight | HTTPRoute weight |
| Request Mirroring | No | Yes | Yes |
| gRPC L7 Load Balancing | Limited | Yes | Yes |
| Config Portability | Low | Medium | High |

## Services

| Service | Purpose | Port |
|---------|---------|------|
| echo-v1 | HTTP echo (v1) | 80 |
| echo-v2 | HTTP echo (v2) | 80 |
| api-service | httpbin endpoints | 80 |
| web-service | Static nginx | 80 |
| grpc-echo | gRPC echo (v1) | 50051 |
| grpc-echo-v2 | gRPC echo (v2) | 50051 |

## Gateway Hostnames

Each gateway has its own set of virtual hosts:

| Route Type | nginx | Traefik | Envoy |
|------------|-------|---------|-------|
| API | api.nginx.local | api.traefik.local | api.envoy.local |
| Echo | echo.nginx.local | echo.traefik.local | echo.envoy.local |
| Web | web.nginx.local | web.traefik.local | web.envoy.local |
| Headers | headers.nginx.local | headers.traefik.local | headers.envoy.local |
| Canary | canary.nginx.local | canary.traefik.local | canary.envoy.local |
| Rate Limited | ratelimited.nginx.local | ratelimited.traefik.local | ratelimited.envoy.local |
| gRPC | grpc.nginx.local | grpc.traefik.local | grpc.envoy.local |

## Running Benchmarks

```bash
# Exec into benchmark runner pod
kubectl exec -it -n gateway-comparison deploy/benchmark-runner -- bash

# Run HTTP connectivity tests
/scripts/benchmark.sh

# Run gRPC comparison
/scripts/grpc-test.sh

# View feature matrix
/scripts/feature-matrix.sh
```

## Files Reference

```
gateway-comparison/
├── target/
│   ├── cluster.yaml                    # Cluster config (medium)
│   └── argocd/
│       ├── app.yaml                    # Main app (manifests)
│       ├── nginx-ingress-app.yaml      # nginx controller
│       ├── nginx-ingress-values.yaml
│       ├── traefik-app.yaml            # Traefik controller
│       ├── traefik-values.yaml
│       ├── envoy-gateway-app.yaml      # Envoy Gateway controller
│       └── envoy-gateway-values.yaml
├── manifests/
│   ├── namespace.yaml                  # gateway-comparison namespace
│   ├── demo-apps.yaml                  # Shared backend services
│   ├── grpc-services.yaml              # gRPC backends
│   ├── nginx-routes.yaml               # Ingress resources
│   ├── traefik-routes.yaml             # Gateway API + IngressRoute
│   ├── envoy-routes.yaml               # Gateway API resources
│   └── benchmark-runner.yaml           # Test pod + scripts
├── tutorial.yaml                       # Interactive tutorial
└── README.md                           # This file
```

## Key Takeaways

1. **nginx-ingress** - Most mature, widest adoption, but annotation-heavy
2. **Traefik** - Kubernetes-native, good middleware system, Gateway API support growing
3. **Envoy Gateway** - CNCF reference implementation, best Gateway API support, modern policies
4. **Gateway API is portable** - Same HTTPRoute works on Traefik and Envoy
5. **gRPC needs L7** - Only Envoy Gateway has full GRPCRoute support

## Related Resources

- [Gateway API Documentation](https://gateway-api.sigs.k8s.io/)
- [Envoy Gateway](https://gateway.envoyproxy.io/)
- [nginx-ingress](https://kubernetes.github.io/ingress-nginx/)
- [Traefik](https://doc.traefik.io/traefik/)

## Prerequisites

- After gateway-tutorial (4.1) is recommended
