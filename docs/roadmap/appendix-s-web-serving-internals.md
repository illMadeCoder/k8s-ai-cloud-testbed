## Phase 16: Web Serving Architecture - The Capstone

*The crown jewel. After mastering infrastructure, security, observability, and operations - we finally examine what actually serves the traffic. This phase synthesizes everything into understanding complete distributed web systems.*

### 16.0 Performance Engineering Fundamentals

**Goal:** Establish the mental models and vocabulary for measuring and reasoning about system performance

*Before benchmarking anything, understand how to measure, what to measure, and how to interpret results. This is the foundation for everything else in this phase.*

**Learning objectives:**
- Master SLI/SLO/SLA hierarchy and how to apply it to web services
- Understand latency distributions, not just averages
- Learn the key metrics every SRE must know
- Build intuition for performance analysis

**SLI/SLO/SLA for Web Services:**

| Concept | Definition | Example |
|---------|------------|---------|
| **SLI** (Indicator) | The metric you measure | p99 latency, error rate, throughput |
| **SLO** (Objective) | Internal target for the SLI | p99 latency < 200ms, 99.9% success |
| **SLA** (Agreement) | Contractual commitment + consequences | 99.9% uptime or credits issued |

**Key Performance Metrics:**

| Metric | What It Tells You | Why It Matters |
|--------|-------------------|----------------|
| **Latency** | Time to complete a request | User experience, timeout risk |
| **Throughput** | Requests per second | Capacity, cost efficiency |
| **Saturation** | How "full" the service is | Headroom, scaling triggers |
| **Errors** | Failure rate | Reliability, user trust |

**Latency Deep Dive:**

| Latency Type | Definition | When It Matters |
|--------------|------------|-----------------|
| **Average (mean)** | Sum / count | Almost never - hides problems |
| **Median (p50)** | 50th percentile | Typical user experience |
| **p95** | 95th percentile | Most users' worst case |
| **p99** | 99th percentile | SLO boundary, tail starts |
| **p99.9** | 99.9th percentile | True tail latency |
| **Max** | Worst observed | Timeout tuning, outliers |

*"If you're not measuring percentiles, you're not measuring latency."*

**Why Tail Latency Matters:**

```
At 1000 req/s with p99 = 500ms:
  - 10 requests per second take > 500ms
  - Over a day: 864,000 slow requests
  - If each user makes 10 requests: 1 in 10 users hit the tail per session
```

**The RED Method (for request-driven services):**
- **R**ate - requests per second
- **E**rrors - failed requests per second
- **D**uration - latency distribution

**The USE Method (for resources):**
- **U**tilization - % time resource is busy
- **S**aturation - queue depth, backpressure
- **E**rrors - error events

**Foundational Laws:**

| Law | Formula | Implication |
|-----|---------|-------------|
| **Little's Law** | L = λW (queue = rate × time) | More latency = more in-flight requests = more memory |
| **Amdahl's Law** | Speedup limited by serial portion | Know your bottleneck before optimizing |
| **Universal Scalability Law** | Contention + coherence costs | Why systems don't scale linearly |

**Tasks:**
- [ ] Create `experiments/scenarios/perf-fundamentals/`
- [ ] SLI selection exercise:
  - [ ] Define SLIs for a typical web API
  - [ ] Choose appropriate percentiles for SLOs
  - [ ] Calculate error budgets from SLOs
- [ ] Latency distribution analysis:
  - [ ] Generate synthetic latency data (normal, bimodal, long-tail)
  - [ ] Visualize percentile distributions
  - [ ] Understand histogram vs heatmap representations
- [ ] Saturation experiments:
  - [ ] Load until saturation (find the knee)
  - [ ] Measure queue depth growth
  - [ ] Observe latency explosion at saturation
- [ ] Little's Law validation:
  - [ ] Measure concurrent connections at various loads
  - [ ] Verify L = λW relationship
- [ ] Create SLO dashboard template:
  - [ ] SLI current value
  - [ ] SLO threshold
  - [ ] Error budget remaining
  - [ ] Burn rate
- [ ] Document performance measurement methodology
- [ ] **ADR:** SLO philosophy and percentile selection

---

### 16.1 Concurrency & Threading Models

**Goal:** Deeply understand how different concurrency models handle load

**Learning objectives:**
- Compare thread-per-request vs event loop vs async/await models
- Understand OS threads vs green threads vs coroutines
- Measure how concurrency models behave under various load patterns

**Tasks:**
- [ ] Create `experiments/scenarios/concurrency-models/`
- [ ] Build equivalent servers demonstrating:
  - [ ] **Thread-per-request** (Java servlet, Go goroutines)
  - [ ] **Event loop** (Node.js, Nginx)
  - [ ] **Async/await** (Rust Tokio, .NET async, Python asyncio)
  - [ ] **Actor model** (Erlang/Elixir, Akka)
  - [ ] **Work-stealing** (Go, Rust Tokio, Java ForkJoinPool)
- [ ] Benchmark scenarios:
  - [ ] High concurrency, low compute (10K concurrent connections)
  - [ ] CPU-bound work under load
  - [ ] I/O-bound work (database calls, external APIs)
  - [ ] Mixed workloads
  - [ ] Connection storms (sudden spike)
- [ ] Measure:
  - [ ] Throughput vs concurrency curves
  - [ ] Latency distribution (p50, p95, p99, p99.9)
  - [ ] Memory per connection
  - [ ] Thread/goroutine/task count under load
  - [ ] Tail latency behavior
- [ ] Document mental models for each approach
- [ ] **ADR:** When to choose which concurrency model

---

### 16.2 Protocol Deep Dive: HTTP/1.1 vs HTTP/2 vs HTTP/3

**Goal:** Understand protocol-level performance characteristics

**Learning objectives:**
- Understand multiplexing, head-of-line blocking, connection reuse
- Compare TCP vs QUIC transport layer impact
- Measure real-world performance differences

**Tasks:**
- [ ] Create `experiments/scenarios/http-protocols/`
- [ ] Deploy identical apps with different protocols:
  - [ ] HTTP/1.1 (baseline)
  - [ ] HTTP/2 (multiplexing, header compression)
  - [ ] HTTP/3 (QUIC, 0-RTT)
- [ ] Test scenarios:
  - [ ] Single large response
  - [ ] Many small requests (API pattern)
  - [ ] Multiplexed streams
  - [ ] High latency networks (simulated)
  - [ ] Packet loss scenarios (simulated)
  - [ ] Mobile/unreliable networks
- [ ] Measure:
  - [ ] Time to first byte (TTFB)
  - [ ] Total transfer time
  - [ ] Connection establishment overhead
  - [ ] Head-of-line blocking impact
  - [ ] 0-RTT resumption benefits
- [ ] Tools: curl timing, h2load, quiche, Wireshark analysis
- [ ] Document protocol selection guidelines
- [ ] **ADR:** HTTP protocol selection by use case

---

### 16.3 API Protocols: REST vs GraphQL vs gRPC

**Goal:** Compare API protocol performance and characteristics

**Learning objectives:**
- Understand serialization overhead (JSON vs Protobuf)
- Compare request/response patterns
- Measure streaming and bidirectional communication

**Tasks:**
- [ ] Create `experiments/scenarios/api-protocols/`
- [ ] Build equivalent APIs:
  - [ ] REST/JSON (OpenAPI spec)
  - [ ] GraphQL (with DataLoader batching)
  - [ ] gRPC (unary, server streaming, client streaming, bidirectional)
  - [ ] gRPC-Web (browser compatibility)
  - [ ] Connect (Buf's gRPC alternative)
- [ ] Test patterns:
  - [ ] Simple CRUD operations
  - [ ] Nested data fetching (N+1 problem)
  - [ ] Large payload transfer
  - [ ] Streaming updates
  - [ ] Bidirectional communication
- [ ] Measure:
  - [ ] Serialization/deserialization time
  - [ ] Payload sizes (wire format)
  - [ ] Latency by operation type
  - [ ] Client complexity
  - [ ] Schema evolution handling
- [ ] Document trade-offs and selection criteria
- [ ] **ADR:** API protocol selection framework

---

### 16.4 Reverse Proxy & Load Balancer Shootout

**Goal:** Compare reverse proxies under realistic conditions

**Learning objectives:**
- Understand proxy architectures and performance characteristics
- Compare configuration complexity vs performance
- Measure proxy overhead in the request path

**Tasks:**
- [ ] Create `experiments/scenarios/proxy-benchmark/`
- [ ] Deploy and benchmark:
  - [ ] **NGINX** (C, event-driven, battle-tested)
  - [ ] **HAProxy** (C, purpose-built LB)
  - [ ] **Envoy** (C++, modern, extensible)
  - [ ] **Caddy** (Go, automatic HTTPS)
  - [ ] **Traefik** (Go, cloud-native)
  - [ ] **Pingora** (Rust, Cloudflare's proxy)
- [ ] Test scenarios:
  - [ ] Raw HTTP proxying (passthrough)
  - [ ] TLS termination
  - [ ] Load balancing algorithms (round-robin, least-conn, consistent hash)
  - [ ] Health checking overhead
  - [ ] Header manipulation
  - [ ] Rate limiting
  - [ ] Connection pooling efficiency
- [ ] Measure:
  - [ ] Requests per second (saturation point)
  - [ ] Added latency (proxy overhead)
  - [ ] Memory per connection
  - [ ] CPU utilization
  - [ ] Connection reuse efficiency
  - [ ] Failure handling (upstream down)
- [ ] Document proxy selection criteria
- [ ] **ADR:** Reverse proxy selection by use case

---

### 16.5 Static File Serving

**Goal:** Understand static content serving at scale

**Learning objectives:**
- Compare static file servers
- Understand caching, compression, and optimization
- Measure serving efficiency for different content types

**Tasks:**
- [ ] Create `experiments/scenarios/static-serving/`
- [ ] Deploy and compare:
  - [ ] NGINX (static serving config)
  - [ ] Caddy
  - [ ] Apache (for comparison)
  - [ ] Lighttpd
  - [ ] Go embedded (embed.FS)
  - [ ] Rust (Actix-files, Axum)
  - [ ] CDN simulation (Varnish)
- [ ] Test scenarios:
  - [ ] Small files (JS, CSS, icons)
  - [ ] Large files (images, videos)
  - [ ] Many concurrent downloads
  - [ ] Range requests (video seeking)
  - [ ] Brotli vs Gzip compression
  - [ ] Cache hit/miss patterns
- [ ] Measure:
  - [ ] Throughput (GB/s)
  - [ ] Requests per second
  - [ ] Memory efficiency
  - [ ] Sendfile/zero-copy utilization
- [ ] Document static serving best practices
- [ ] **ADR:** Static file serving architecture

---

### 16.6 Language Runtime Deep Dive

**Goal:** Understand runtime characteristics beyond simple benchmarks

**Learning objectives:**
- Understand GC impact vs manual memory management
- Compare JIT warmup vs AOT compilation
- Measure under sustained load (not just peak)

**Tasks:**
- [ ] Create `experiments/scenarios/runtime-deepdive/`
- [ ] Build identical HTTP servers in:
  - [ ] **Go** - goroutines, GC, fast compile
  - [ ] **Rust** - zero-cost abstractions, no GC
  - [ ] **Java** (GraalVM native) - JIT/AOT comparison
  - [ ] **.NET** - async, tiered JIT
  - [ ] **Node.js** - V8, single-threaded + workers
  - [ ] **Bun** - JavaScriptCore, all-in-one
  - [ ] **Python** (uvloop + uvicorn) - async Python
  - [ ] **Elixir/Phoenix** - BEAM VM, fault tolerance
- [ ] Advanced scenarios:
  - [ ] Sustained load (1 hour continuous)
  - [ ] GC pause measurement
  - [ ] Memory growth over time
  - [ ] Cold start vs warm performance
  - [ ] Multi-core scaling
  - [ ] Graceful degradation under overload
- [ ] Measure:
  - [ ] p99.9 latency (tail latency)
  - [ ] GC pause times and frequency
  - [ ] Memory footprint over time
  - [ ] CPU efficiency (requests per CPU-second)
  - [ ] Container resource limits behavior
- [ ] Document runtime selection guide
- [ ] **ADR:** Runtime selection framework

---

### 16.7 WebSocket & Real-time Communication

**Goal:** Benchmark persistent connection technologies

**Learning objectives:**
- Understand connection scaling challenges
- Compare WebSocket, SSE, and long-polling
- Measure broadcast and pub/sub patterns

**Tasks:**
- [ ] Create `experiments/scenarios/realtime-benchmark/`
- [ ] Implement servers supporting:
  - [ ] WebSocket
  - [ ] Server-Sent Events (SSE)
  - [ ] Long-polling (baseline)
  - [ ] Socket.IO (with fallbacks)
  - [ ] gRPC streaming
- [ ] Test patterns:
  - [ ] Connection scaling (10K, 100K connections)
  - [ ] Message broadcast (1 → N)
  - [ ] Pub/sub with topics
  - [ ] Bidirectional messaging
  - [ ] Reconnection storms
- [ ] Measure:
  - [ ] Connections per server
  - [ ] Memory per connection
  - [ ] Message delivery latency
  - [ ] Broadcast time (fan-out)
  - [ ] Reconnection handling
- [ ] Document real-time architecture patterns
- [ ] **ADR:** Real-time communication selection

---

### 16.8 The Complete Stack: End-to-End Benchmark

**Goal:** Synthesize everything - benchmark complete distributed systems

**Learning objectives:**
- Understand how all layers interact
- Identify bottlenecks in complete systems
- Make informed architecture decisions

**Tasks:**
- [ ] Create `experiments/scenarios/complete-stack/`
- [ ] Build reference architectures:
  - [ ] **Stack A:** NGINX → Go → PostgreSQL
  - [ ] **Stack B:** Envoy → Rust → PostgreSQL
  - [ ] **Stack C:** Caddy → Node.js → MongoDB
  - [ ] **Stack D:** Traefik → .NET → SQL Server
  - [ ] **Stack E:** HAProxy → Java (GraalVM) → PostgreSQL
- [ ] Realistic workload patterns:
  - [ ] E-commerce (read-heavy, sessions, cart)
  - [ ] Social feed (fan-out, real-time)
  - [ ] API gateway (mixed protocols, auth)
  - [ ] Content platform (static + dynamic)
- [ ] Full observability:
  - [ ] Distributed tracing (where is time spent?)
  - [ ] Flame graphs (CPU profiling)
  - [ ] Memory profiling
  - [ ] Network analysis
- [ ] Measure end-to-end:
  - [ ] User-perceived latency
  - [ ] System throughput
  - [ ] Resource efficiency (cost per request)
  - [ ] Scaling characteristics
  - [ ] Failure behavior
- [ ] Document architecture comparison
- [ ] **ADR:** Reference architecture selection

---

### 16.9 Production Patterns & Anti-Patterns

**Goal:** Document learnings from all benchmarks into actionable guidance

**Learning objectives:**
- Synthesize benchmarking insights
- Create decision frameworks
- Build portfolio-quality documentation

**Tasks:**
- [ ] Create comprehensive documentation:
  - [ ] **Decision trees** for technology selection
  - [ ] **Performance profiles** for each technology
  - [ ] **Anti-patterns** discovered through benchmarking
  - [ ] **Optimization playbook** by bottleneck type
- [ ] Architecture diagrams:
  - [ ] Request flow through complete systems
  - [ ] Scaling patterns for each architecture
  - [ ] Failure modes and mitigations
- [ ] Create reusable benchmarking toolkit:
  - [ ] k6 scripts for all scenarios
  - [ ] Grafana dashboards for comparison
  - [ ] Automated benchmark runners
- [ ] Write blog-quality posts on key findings
- [ ] **Final ADR:** Personal technology radar based on evidence

---
