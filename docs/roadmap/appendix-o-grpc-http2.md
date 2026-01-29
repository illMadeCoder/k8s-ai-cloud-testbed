# Appendix O: gRPC & HTTP/2 Patterns

*Deep dive into gRPC, HTTP/2, and high-performance API patterns. Use after completing Phase 4 (Traffic Management).*

---

## Overview

This appendix covers advanced gRPC and HTTP/2 patterns that go beyond basic traffic management:

- gRPC fundamentals and Protocol Buffers
- HTTP/2 multiplexing and performance characteristics
- gRPC-Web for browser clients
- Load balancing strategies for gRPC
- Observability for gRPC services

---

## O.1 gRPC Fundamentals

**Goal:** Understand gRPC architecture and Protocol Buffers

**Learning objectives:**
- Protocol Buffers schema design
- gRPC service definitions
- Unary vs streaming patterns
- Code generation workflow

**Tasks:**
- [ ] Create `experiments/scenarios/grpc-fundamentals/`
- [ ] Protocol Buffers:
  - [ ] Define `.proto` schema
  - [ ] Compile with `protoc`
  - [ ] Understand wire format efficiency
- [ ] gRPC patterns:
  - [ ] Unary RPC (request/response)
  - [ ] Server streaming
  - [ ] Client streaming
  - [ ] Bidirectional streaming
- [ ] Error handling:
  - [ ] gRPC status codes
  - [ ] Rich error details
  - [ ] Deadline propagation
- [ ] Document gRPC patterns

---

## O.2 HTTP/2 Deep Dive

**Goal:** Understand HTTP/2 protocol characteristics

**Learning objectives:**
- HTTP/2 multiplexing and streams
- Header compression (HPACK)
- Flow control and backpressure
- HTTP/2 vs HTTP/1.1 performance

**Tasks:**
- [ ] Create `experiments/scenarios/http2-analysis/`
- [ ] Protocol analysis:
  - [ ] Frame types (DATA, HEADERS, SETTINGS)
  - [ ] Stream multiplexing
  - [ ] Head-of-line blocking differences
- [ ] Performance characteristics:
  - [ ] Connection reuse
  - [ ] Header compression savings
  - [ ] Latency improvements
- [ ] Debugging:
  - [ ] h2c (cleartext HTTP/2)
  - [ ] Wireshark HTTP/2 analysis
- [ ] Document HTTP/2 vs HTTP/1.1 trade-offs

---

## O.3 gRPC-Web

**Goal:** Enable browser clients to call gRPC services

**Learning objectives:**
- gRPC-Web protocol differences
- Envoy as gRPC-Web proxy
- Browser client implementation

**Tasks:**
- [ ] Create `experiments/scenarios/grpc-web/`
- [ ] Deploy gRPC-Web proxy:
  - [ ] Envoy gRPC-Web filter
  - [ ] grpc-web standalone proxy
- [ ] Browser client:
  - [ ] JavaScript/TypeScript client
  - [ ] Proto compilation for web
- [ ] Limitations:
  - [ ] No bidirectional streaming (browser)
  - [ ] Server streaming workarounds
- [ ] Document gRPC-Web patterns

---

## O.4 gRPC Load Balancing

**Goal:** Implement effective load balancing for gRPC

**Learning objectives:**
- L4 vs L7 load balancing for gRPC
- Client-side vs proxy-side balancing
- Connection pooling strategies

**Challenges:**
- gRPC uses long-lived HTTP/2 connections
- L4 load balancing doesn't distribute requests evenly
- Need L7 awareness for proper distribution

**Tasks:**
- [ ] Create `experiments/scenarios/grpc-loadbalancing/`
- [ ] L7 load balancing:
  - [ ] Envoy gRPC routing
  - [ ] NGINX with grpc_pass
  - [ ] Linkerd/Istio gRPC balancing
- [ ] Client-side balancing:
  - [ ] gRPC name resolution
  - [ ] Round-robin picker
  - [ ] Health checking integration
- [ ] Benchmark:
  - [ ] L4 vs L7 distribution fairness
  - [ ] Connection overhead
- [ ] Document gRPC LB patterns
- [ ] **ADR:** gRPC load balancing strategy

---

## O.5 gRPC Observability

**Goal:** Implement comprehensive gRPC monitoring

**Learning objectives:**
- gRPC-specific metrics
- Distributed tracing for gRPC
- gRPC logging patterns

**Tasks:**
- [ ] Create `experiments/scenarios/grpc-observability/`
- [ ] Metrics:
  - [ ] Request count by method/status
  - [ ] Latency histograms
  - [ ] Stream message counts
  - [ ] Connection pool metrics
- [ ] Tracing:
  - [ ] OpenTelemetry gRPC interceptor
  - [ ] Trace context propagation
  - [ ] Span per RPC call
- [ ] Logging:
  - [ ] Request/response logging
  - [ ] Payload sampling
  - [ ] Error details capture
- [ ] Dashboard:
  - [ ] gRPC RED metrics
  - [ ] Per-method breakdown
- [ ] Document gRPC observability patterns

---

## O.6 gRPC Security

**Goal:** Secure gRPC communications

**Tasks:**
- [ ] TLS for gRPC:
  - [ ] Server-side TLS
  - [ ] mTLS (mutual TLS)
  - [ ] cert-manager integration
- [ ] Authentication:
  - [ ] Token-based (JWT in metadata)
  - [ ] Per-RPC credentials
  - [ ] Interceptor patterns
- [ ] Authorization:
  - [ ] Method-level policies
  - [ ] OPA integration
- [ ] Document gRPC security patterns

---

## O.7 gRPC Performance Tuning

**Goal:** Optimize gRPC for high throughput

**Tasks:**
- [ ] Connection tuning:
  - [ ] Keep-alive settings
  - [ ] Max concurrent streams
  - [ ] Flow control window
- [ ] Message optimization:
  - [ ] Compression (gzip)
  - [ ] Message size limits
  - [ ] Streaming vs batching
- [ ] Benchmarking:
  - [ ] ghz (gRPC benchmarking tool)
  - [ ] Latency vs throughput curves
- [ ] Document performance tuning guide

---

## O.8 gRPC Transport Layer Benchmarking

**Goal:** Measure gRPC latency and throughput across four transport layers to quantify the cost of each network hop

**Learning objectives:**
- Understand how transport layer choice (UDS, CNI, LAN, internet) impacts gRPC latency and throughput
- Design reproducible transport benchmarks that isolate network overhead from application overhead
- Produce data-driven guidance on when to use each transport for gRPC services on Kubernetes

**Tasks:**
- [ ] Create `experiments/scenarios/grpc-transport-benchmark/`
- [ ] Unix Domain Socket (UDS):
  - [ ] gRPC over UDS: shared `hostPath` or `emptyDir` volume between containers
  - [ ] Sidecar pattern: app + gRPC server in same pod sharing socket
  - [ ] Zero network stack overhead: no TCP, no IP, kernel-only IPC
  - [ ] Use cases: Envoy ↔ app, OTel collector sidecar, local AI inference
- [ ] Kubernetes CNI (pod-to-pod, same node and cross-node):
  - [ ] Same-node pod-to-pod: veth pairs through bridge/eBPF (CNI-dependent)
  - [ ] Cross-node pod-to-pod: VXLAN encapsulation, WireGuard, or direct routing
  - [ ] CNI comparison impact: Cilium (eBPF) vs Calico (iptables) vs Flannel (VXLAN)
  - [ ] Service mesh overhead: measure with and without Istio/Linkerd sidecar proxy
- [ ] LAN (cross-host, same datacenter/rack):
  - [ ] Bare-metal or VM-to-VM within same network segment
  - [ ] Factors: MTU size (1500 vs 9000 jumbo frames), switch hops, NIC offloading
  - [ ] Kubernetes context: cross-node traffic on Talos home lab or bare-metal cluster
  - [ ] Comparison: NodePort, LoadBalancer, and headless Service routing
- [ ] Internet (cross-region, WAN):
  - [ ] TLS handshake overhead on every connection establishment
  - [ ] Latency floor: speed-of-light RTT between regions
  - [ ] Connection multiplexing benefit: HTTP/2 amortizes handshake cost
  - [ ] Real-world: cloud region-to-region, edge-to-cloud, client-to-API
- [ ] Benchmark experiment design:
  - [ ] Test app: simple gRPC service (echo, key-value lookup, small payload, large payload)
  - [ ] Payload sizes: 64B, 1KB, 64KB, 1MB (measure serialization + transport)
  - [ ] RPC patterns: unary, server-streaming (1000 messages), bidirectional streaming
  - [ ] Concurrency levels: 1, 10, 50, 100 concurrent streams
- [ ] Benchmark tooling:
  - [ ] ghz for gRPC load generation (supports UDS, TLS, concurrency control)
  - [ ] Containerized benchmark client as Kubernetes Job
  - [ ] Prometheus metrics collection during benchmark runs
  - [ ] Grafana dashboard for latency heatmaps across transport layers
- [ ] Measurement methodology:
  - [ ] Warm-up: 10s discard period before recording
  - [ ] Steady-state: 60s measurement window per configuration
  - [ ] Report p50, p95, p99, p99.9 latency and throughput (msg/s, MB/s)
  - [ ] Multiple runs (5+) with coefficient of variation
- [ ] Isolation and controls:
  - [ ] Pin benchmark pods to specific nodes (nodeSelector/affinity)
  - [ ] Dedicated resource requests to prevent CPU throttling artifacts
  - [ ] Same payload and proto definition across all transport tests
  - [ ] Disable service mesh sidecars for baseline, then re-enable for comparison
- [ ] Expected outcome matrix:
  - [ ] UDS: lowest latency (~10-50μs), highest throughput, single-pod only
  - [ ] CNI same-node: ~100-300μs, near-UDS throughput, pod-to-pod flexibility
  - [ ] CNI cross-node: ~200-800μs, CNI and MTU dependent
  - [ ] LAN: ~0.5-2ms, switch/NIC dependent, jumbo frames help large payloads
  - [ ] Internet: ~20-200ms+, TLS + RTT dominated, multiplexing essential
- [ ] Analysis and decision framework:
  - [ ] Latency budget allocation: what percentage is transport vs application?
  - [ ] When UDS is worth the coupling (sidecar architectures, <100μs requirement)
  - [ ] When cross-node CNI is good enough (most microservices)
  - [ ] When internet latency demands connection pooling, retries, hedged requests
  - [ ] Cost of service mesh proxy hop per transport layer
- [ ] Produce benchmark report:
  - [ ] Latency CDF charts per transport layer and payload size
  - [ ] Throughput scaling charts (concurrency vs messages/second)
  - [ ] Resource consumption (CPU, memory) of gRPC server under each transport
  - [ ] Recommendations table: transport layer × workload pattern → guidance
- [ ] **ADR:** Document gRPC transport layer selection criteria and benchmark results

---

## Cross-References

| Topic | Location |
|-------|----------|
| Traffic routing basics | Phase 4: Traffic Management |
| Service mesh gRPC | Phase 7: Service Mesh |
| API design principles | Appendix F: API Design & Contracts |
| Benchmarking methodology | Appendix J.17: Database Benchmarking Methodology |
| Web serving & proxy internals | Appendix S: Web Serving Internals |

---

## When to Use This Appendix

- Building microservices with gRPC
- Migrating from REST to gRPC
- Optimizing high-throughput APIs
- Debugging gRPC performance issues
