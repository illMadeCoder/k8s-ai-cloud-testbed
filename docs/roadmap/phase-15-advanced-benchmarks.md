## Phase 15: Advanced Topics & Benchmarks

*Deep dives and performance comparisons - now that fundamentals are solid.*

### 15.1 Database Performance Comparison

**Goal:** Compare relational databases for Kubernetes workloads

**Learning objectives:**
- Benchmark database performance objectively
- Understand trade-offs between options
- Make data-driven database selection

**Tasks:**
- [ ] Create `experiments/scenarios/database-benchmark/`
- [ ] Deploy databases via Crossplane/operators:
  - [ ] PostgreSQL (CloudNativePG)
  - [ ] MySQL (via operator)
  - [ ] Cloud-managed (Azure SQL, RDS via Crossplane)
- [ ] Create benchmark schema and data
- [ ] Run benchmarks:
  - [ ] pgbench / sysbench
  - [ ] OLTP workloads (TPC-C style)
  - [ ] Read-heavy vs write-heavy
- [ ] Compare:
  - [ ] Throughput (TPS)
  - [ ] Latency percentiles
  - [ ] Resource consumption
  - [ ] Operational complexity
- [ ] Document findings and recommendations

---

### 15.2 Message Queue Performance Comparison

**Goal:** Compare messaging systems under load

**Learning objectives:**
- Benchmark throughput and latency
- Understand performance characteristics
- Inform technology selection

**Tasks:**
- [ ] Create `experiments/scenarios/messaging-benchmark/`
- [ ] Deploy all three brokers (from Phase 7)
- [ ] Build benchmarking clients
- [ ] Test scenarios:
  - [ ] High throughput (max messages/sec)
  - [ ] Low latency (p99 measurement)
  - [ ] Fan-out (1 → N consumers)
  - [ ] Persistence impact
- [ ] Compare:
  - [ ] Messages per second
  - [ ] End-to-end latency
  - [ ] Resource consumption
  - [ ] Recovery time after failure
- [ ] Document performance comparison

---

### 15.3 Service Mesh Performance Comparison

**Goal:** Measure service mesh overhead

**Learning objectives:**
- Quantify latency overhead
- Compare resource consumption
- Inform mesh selection

**Tasks:**
- [ ] Create `experiments/scenarios/mesh-benchmark/`
- [ ] Deploy baseline app (no mesh)
- [ ] Deploy same app with:
  - [ ] Istio
  - [ ] Linkerd
  - [ ] Cilium
- [ ] Measure:
  - [ ] Latency overhead (p50, p95, p99)
  - [ ] CPU per pod (sidecar cost)
  - [ ] Memory per pod
  - [ ] Control plane resources
- [ ] Test at scale:
  - [ ] 10, 50, 100 services
  - [ ] High RPS scenarios
- [ ] Document mesh comparison

---

### 15.4 Runtime Performance Comparison

**Goal:** Compare web server runtimes for API workloads

**Learning objectives:**
- Benchmark different language runtimes
- Understand performance characteristics
- Portfolio piece for runtime expertise

**Tasks:**
- [ ] Create `experiments/scenarios/runtime-benchmark/`
- [ ] Build identical API in:
  - [ ] Go (net/http)
  - [ ] Rust (Axum)
  - [ ] .NET (ASP.NET Core)
  - [ ] Node.js (Fastify)
  - [ ] Bun
- [ ] Implement endpoints:
  - [ ] GET /health
  - [ ] GET /json (serialize response)
  - [ ] POST /echo (deserialize + serialize)
  - [ ] GET /compute (CPU-bound work)
- [ ] Benchmark with k6:
  - [ ] Throughput (RPS)
  - [ ] Latency distribution
  - [ ] Memory footprint
  - [ ] Container image size
  - [ ] Cold start time
- [ ] Document runtime comparison

---

### 15.5 Cost-Efficiency Benchmarking

**Goal:** Add cost as a first-class benchmark dimension

*FinOps consideration: Performance without cost context is incomplete. Always measure cost per request/transaction.*

**Learning objectives:**
- Measure cost efficiency, not just raw performance
- Calculate cost per transaction for different options
- Make data-driven technology decisions including cost

**Tasks:**
- [ ] Cost metrics for all benchmarks:
  - [ ] Cost per 1M requests (database queries, API calls)
  - [ ] Cost per GB stored (databases, object storage)
  - [ ] Cost per GB transferred (messaging, cross-region)
- [ ] Database cost efficiency:
  - [ ] Cost per transaction by database type
  - [ ] Reserved vs on-demand cost modeling
  - [ ] Self-managed vs cloud-managed TCO
- [ ] Messaging cost efficiency:
  - [ ] Cost per million messages by broker type
  - [ ] Storage cost per GB retained
  - [ ] Cloud messaging vs self-managed breakeven
- [ ] Compute cost efficiency:
  - [ ] Cost per request by runtime
  - [ ] Spot vs on-demand for batch workloads
  - [ ] Serverless vs always-on cost comparison
- [ ] Service mesh cost efficiency:
  - [ ] Overhead cost as percentage of workload
  - [ ] Security benefit vs cost trade-off
- [ ] Create cost-efficiency dashboard:
  - [ ] Cost per transaction trending
  - [ ] Comparison across technologies
  - [ ] Anomaly detection for cost spikes
- [ ] Document cost-efficiency patterns
- [ ] **ADR:** Document cost as benchmark dimension

---

