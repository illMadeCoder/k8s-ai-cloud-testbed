## Phase 4: Traffic Management

*Control how traffic flows before learning deployment strategies that depend on it.*

### 4.1 Gateway Tutorial: Ingress → Gateway API Evolution

**Goal:** Understand L7 traffic management from legacy Ingress through modern Gateway API

**Learning objectives:**
- Master Kubernetes Ingress and its limitations
- Understand Gateway API resources (Gateway, HTTPRoute, GRPCRoute)
- Experience the migration path from Ingress to Gateway API
- Implement advanced routing and traffic manipulation patterns

**Scenario:** `experiments/scenarios/gateway-tutorial/`

**Part 1: Ingress Basics**
- [ ] Deploy nginx-ingress controller
- [ ] Configure basic Ingress resources:
  - [ ] Path-based routing
  - [ ] Host-based routing (virtual hosts)
  - [ ] TLS termination with cert-manager
- [ ] Understand Ingress annotations pattern

**Part 2: Hitting Ingress Limitations**
- [ ] Attempt rate limiting (annotation hell begins)
- [ ] Try header-based routing (limited support)
- [ ] Add authentication (nginx-specific annotations)
- [ ] Traffic splitting for canary (awkward with Ingress)
- [ ] Document the pain points

**Part 3: Migrate to Gateway API**
- [ ] Deploy Envoy Gateway (CNCF reference implementation)
- [ ] Create Gateway resource
- [ ] Migrate Ingress rules to HTTPRoute
- [ ] Compare configuration complexity
- [ ] Side-by-side: same routes, both approaches

**Part 4: Gateway API Deep Dive**
- [ ] HTTPRoute patterns:
  - [ ] Path/host/header/query/method matching
  - [ ] Weight-based traffic splitting
  - [ ] Request mirroring
  - [ ] URL rewriting
  - [ ] Header modification
  - [ ] Redirects
- [ ] Advanced features:
  - [ ] Timeouts and retries
  - [ ] Rate limiting (BackendTrafficPolicy)
  - [ ] CORS configuration
- [ ] TLS configuration:
  - [ ] TLS termination
  - [ ] TLS passthrough
  - [ ] mTLS with client certificates
- [ ] Multi-gateway patterns:
  - [ ] Internal vs external gateways
  - [ ] Namespace isolation (ReferenceGrant)

**Part 5: gRPC Traffic Management**

*gRPC has fundamentally different traffic management requirements than HTTP/1.1. HTTP/2 multiplexing, streaming RPCs, and binary protocols require specialized handling.*

**5-zero: Why gRPC? Understanding the Motivation**

*Before learning gRPC traffic patterns, understand why gRPC exists and when to use it.*

- [ ] The problems gRPC solves:
  - [ ] **Performance**: Binary serialization (protobuf) vs JSON text parsing
  - [ ] **Strong typing**: Schema-first with .proto files, compile-time safety
  - [ ] **Code generation**: Auto-generated clients/servers in any language
  - [ ] **Streaming**: Native support for server/client/bidirectional streaming
  - [ ] **HTTP/2**: Multiplexing, header compression, efficient connection use
- [ ] When to choose gRPC over REST:
  - [ ] Internal service-to-service communication (microservices)
  - [ ] High-throughput, low-latency requirements
  - [ ] Streaming data (real-time feeds, large file transfers)
  - [ ] Polyglot environments (consistent contracts across languages)
- [ ] When REST is still better:
  - [ ] Public APIs (human-readable, browser-native)
  - [ ] Simple CRUD operations
  - [ ] When debugging simplicity matters more than performance
  - [ ] Caching requirements (HTTP caching works naturally with REST)
- [ ] The browser problem (critical limitation):
  - [ ] Browsers use fetch API which doesn't support HTTP/2 trailers
  - [ ] gRPC uses trailers for status codes and metadata
  - [ ] Result: **Native gRPC doesn't work in browsers**
  - [ ] Solutions: gRPC-Web (transcoding) or REST gateway
- [ ] Trade-offs to understand:
  - [ ] Debugging: Can't curl a gRPC endpoint easily (binary, needs tools)
  - [ ] Tooling: Need grpcurl, Postman gRPC, or custom clients
  - [ ] Learning curve: Protobuf schema language, code generation pipeline
  - [ ] Ecosystem: REST has more middleware, documentation tools
- [ ] Demo: Compare same API in REST vs gRPC
  - [ ] Latency comparison under load
  - [ ] Message size comparison
  - [ ] Developer experience comparison

**5a: gRPC Fundamentals & The Load Balancing Problem**
- [ ] Deploy gRPC demo services (3-service chain)
- [ ] Demonstrate the HTTP/2 load balancing problem:
  - [ ] L4 load balancing fails (all requests go to one pod)
  - [ ] Connection pooling vs request distribution
  - [ ] Why HTTP/2 multiplexing breaks traditional LB
- [ ] Load balancing strategies:
  - [ ] Proxy-based L7 LB (gateway handles it)
  - [ ] Client-side LB (grpc-go, grpc-java built-in)
  - [ ] Look-aside LB with xDS (Envoy service mesh pattern)
- [ ] Extracting the root cause here helps understand why solutions differ from HTTP/1.1 patterns

**5b: gRPC with Ingress (The Pain)**
- [ ] nginx-ingress gRPC configuration:
  - [ ] `nginx.ingress.kubernetes.io/backend-protocol: "GRPC"`
  - [ ] `nginx.ingress.kubernetes.io/ssl-redirect: "true"` (gRPC needs TLS or h2c)
  - [ ] Health check configuration (gRPC health protocol)
- [ ] Traefik gRPC:
  - [ ] h2c (HTTP/2 cleartext) configuration
  - [ ] TLS termination with gRPC backends
- [ ] Document the annotation complexity and lack of standardization

**5c: GRPCRoute with Gateway API**
- [ ] Deploy GRPCRoute resources:
  - [ ] Service matching by gRPC service name
  - [ ] Method matching (package.Service/Method)
  - [ ] Header matching on gRPC metadata
- [ ] Traffic manipulation:
  - [ ] Weight-based splitting for gRPC canary
  - [ ] gRPC mirroring
  - [ ] Metadata/header injection
- [ ] Compare: Ingress annotations vs GRPCRoute clarity

**5d: gRPC Timeouts, Retries & Deadlines**
- [ ] Deadline propagation:
  - [ ] Client sets deadline
  - [ ] Gateway respects/modifies deadline
  - [ ] Deadline propagation across service chain
  - [ ] `grpc-timeout` header handling
- [ ] Retry configuration:
  - [ ] Retryable status codes (UNAVAILABLE, RESOURCE_EXHAUSTED)
  - [ ] Non-retryable (INVALID_ARGUMENT, NOT_FOUND)
  - [ ] Retry budgets and backoff
  - [ ] Hedged requests (send to multiple backends)
- [ ] Timeout patterns:
  - [ ] Per-call timeout
  - [ ] Per-stream timeout (for streaming RPCs)
  - [ ] Connection timeout

**5e: gRPC Streaming Patterns**
- [ ] Unary RPC (request-response) - standard pattern
- [ ] Server streaming:
  - [ ] Gateway timeout considerations (long-lived streams)
  - [ ] Flow control / backpressure
- [ ] Client streaming:
  - [ ] Buffering at gateway
  - [ ] Upload size limits
- [ ] Bidirectional streaming:
  - [ ] WebSocket-like long-lived connections
  - [ ] Gateway connection limits
  - [ ] Idle timeout handling
- [ ] Test each pattern through the gateway

**5f: gRPC Health Checking**
- [ ] gRPC Health Checking Protocol (grpc.health.v1.Health):
  - [ ] Implement health service in demo apps
  - [ ] Service-level health vs server-level health
- [ ] Gateway health checking:
  - [ ] Envoy Gateway gRPC health checks
  - [ ] nginx-ingress gRPC health probes
- [ ] Kubernetes integration:
  - [ ] gRPC liveness/readiness probes (K8s 1.24+)
  - [ ] `grpc` probe type configuration

**5g: gRPC-Web for Browser Clients**
- [ ] The browser problem (no HTTP/2 trailers in browser fetch)
- [ ] gRPC-Web protocol:
  - [ ] Base64 encoding for binary
  - [ ] Trailer handling via special headers
- [ ] Deploy Envoy gRPC-Web filter:
  - [ ] Transcode gRPC-Web → gRPC
  - [ ] CORS configuration for browser
- [ ] Alternative: grpc-web proxy sidecar
- [ ] Test from browser JavaScript client

**5h: gRPC Transcoding (REST ↔ gRPC)**
- [ ] Use case: REST clients calling gRPC services
- [ ] google.api.http annotations in proto:
  ```protobuf
  rpc GetUser(GetUserRequest) returns (User) {
    option (google.api.http) = {
      get: "/v1/users/{user_id}"
    };
  }
  ```
- [ ] Envoy gRPC-JSON transcoder filter
- [ ] Test: curl REST endpoint → gRPC backend
- [ ] Extracting the root cause here helps understand when to use transcoding vs native gRPC

**5i: gRPC Security**
- [ ] TLS modes:
  - [ ] TLS termination at gateway (gateway → backend cleartext)
  - [ ] TLS passthrough (end-to-end encryption)
  - [ ] mTLS (mutual TLS with client certs)
- [ ] Per-RPC credentials:
  - [ ] Bearer tokens in metadata
  - [ ] JWT validation at gateway
- [ ] Channel vs call credentials:
  - [ ] Channel: TLS for transport
  - [ ] Call: per-request auth tokens

**5j: gRPC Observability**
- [ ] Metrics:
  - [ ] gRPC method-level metrics (rate, errors, duration)
  - [ ] Prometheus gRPC interceptors
  - [ ] Gateway-level gRPC metrics
- [ ] Distributed tracing:
  - [ ] OpenTelemetry gRPC instrumentation
  - [ ] Trace context propagation via metadata
  - [ ] View traces in Tempo from Part 4 setup
- [ ] Access logging:
  - [ ] gRPC method in access logs
  - [ ] Status code mapping (gRPC → HTTP)

**5k: gRPC Reflection & Debugging**
- [ ] gRPC Server Reflection:
  - [ ] Enable reflection in demo services
  - [ ] Why reflection is useful (schema discovery)
  - [ ] Security considerations (disable in prod?)
- [ ] Debugging tools:
  - [ ] grpcurl for CLI testing
  - [ ] grpc-client-cli
  - [ ] Postman gRPC support
  - [ ] BloomRPC / Evans
- [ ] Gateway debugging:
  - [ ] Envoy admin interface for gRPC stats
  - [ ] Connection/stream inspection

**Deliverables:**
- [ ] Working tutorial with all five parts
- [ ] gRPC demo application (3-service chain with all RPC types)
- [ ] **ADR:** Gateway API implementation choice (Envoy Gateway)
- [ ] **ADR:** gRPC load balancing strategy
- [ ] Comparison table: Ingress vs Gateway API
- [ ] Comparison table: gRPC gateway support matrix

---

### 4.2 Gateway Comparison: In-Cluster Implementations

**Goal:** Compare in-cluster gateway implementations for informed selection

**Learning objectives:**
- Understand trade-offs between gateway implementations
- Compare configuration patterns and complexity
- Evaluate feature availability and resource consumption

**Scenario:** `experiments/scenarios/gateway-comparison/`

**Implementations to compare:**
- [ ] nginx-ingress (most widely deployed)
- [ ] Traefik (popular, good Gateway API support)
- [ ] Envoy Gateway (CNCF reference, pure Gateway API)

**Same demo app, same routes on each:**
- [ ] Path-based routing to multiple services
- [ ] Host-based virtual hosting
- [ ] TLS termination
- [ ] Rate limiting
- [ ] Header manipulation
- [ ] gRPC routing (GRPCRoute or equivalent)
- [ ] gRPC load balancing verification

**Comparison metrics:**
| Metric | nginx | Traefik | Envoy Gateway |
|--------|-------|---------|---------------|
| Config complexity | | | |
| Gateway API support | | | |
| GRPCRoute support | | | |
| gRPC LB effectiveness | | | |
| gRPC streaming support | | | |
| Resource usage (CPU/mem) | | | |
| Feature completeness | | | |
| Observability integration | | | |
| Community/docs quality | | | |

**Deliverables:**
- [ ] Side-by-side deployment of all three
- [ ] Comparison matrix with findings
- [ ] Recommendation criteria document

---

### 4.3 Cloud Gateway Comparison: Managed vs In-Cluster

**Goal:** Compare cloud-native application gateways with in-cluster solutions

**Learning objectives:**
- Understand cloud provider L7 gateway offerings
- Evaluate cost, performance, and operational trade-offs
- Make informed decisions for production architectures

**Scenario:** `experiments/scenarios/cloud-gateway-comparison/`

**Implementations to compare:**
- [ ] AWS ALB Ingress Controller (provisions AWS ALBs)
- [ ] Azure AGIC (provisions Azure Application Gateways)
- [ ] Envoy Gateway in-cluster (baseline comparison)

**Infrastructure:**
- [ ] AWS EKS cluster via Crossplane
- [ ] Azure AKS cluster via Crossplane
- [ ] Talos cluster (in-cluster baseline)

**Comparison Metrics:**

| Category | Metrics |
|----------|---------|
| **Latency** | p50, p95, p99 request latency |
| **Throughput** | Max RPS, RPS at 10/100/1000 concurrency |
| **Provisioning** | Time to create gateway, time to add route |
| **Cost** | Hourly gateway cost, per-request cost, data transfer |
| **Config propagation** | Time from kubectl apply → traffic flowing |
| **Reliability** | Multi-AZ behavior, failover time, SLA |
| **Features** | Rate limiting, auth (JWT/mTLS), WAF, WebSocket |
| **gRPC Support** | gRPC routing, LB effectiveness, streaming, health checks |
| **Observability** | Metrics export, access logs, tracing integration |
| **Blast radius** | Impact of misconfig, rollback ease |
| **Vendor lock-in** | Portability of configuration |

**Load testing:**
- [ ] Use k6 for consistent load generation
- [ ] Test at multiple concurrency levels
- [ ] Measure during route changes

**Deliverables:**
- [ ] Crossplane compositions for AWS ALB IC and Azure AGIC
- [ ] Automated metrics collection
- [ ] Scoring matrix with all metrics
- [ ] **ADR:** When to use cloud-native vs in-cluster gateways
- [ ] Cost calculator for different traffic levels

---

## Dependencies

```
gateway-tutorial
       ↓
gateway-comparison
       ↓
cloud-gateway-comparison
```

## Backlog

After completing Phase 4:
- [ ] Play through gateway-tutorial (Parts 1-5)
- [ ] Play through gateway-comparison
- [ ] Play through cloud-gateway-comparison

## ADRs

- [ ] ADR: Gateway API implementation choice (Envoy Gateway)
- [ ] ADR: gRPC load balancing strategy
- [ ] ADR: Cloud-native vs in-cluster gateways
