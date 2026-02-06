## Phase 10: Performance & Cost Engineering

*The capstone - synthesize everything into data-driven system engineering. Benchmark each technology, measure full stack composition, calculate cost per transaction.*

> **Deep Dive:** For eBPF-based I/O profiling during benchmarks, see [Appendix T: eBPF & Advanced Metrics](docs/roadmap/appendix-t-ebpf-metrics.md).

### 10.1 Database Performance Comparison

**Goal:** Compare relational databases for Kubernetes workloads

**Learning objectives:**
- Benchmark database performance objectively
- Understand trade-offs between options
- Make data-driven database selection

**Tasks:**
- [ ] Create `experiments/database-benchmark/`
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

### 10.2 Message Queue Performance Comparison

**Goal:** Compare messaging systems under load

**Learning objectives:**
- Benchmark throughput and latency
- Understand performance characteristics
- Inform technology selection

**Tasks:**
- [ ] Create `experiments/messaging-benchmark/`
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

### 10.3 Service Mesh Performance Comparison

**Goal:** Measure service mesh overhead

**Learning objectives:**
- Quantify latency overhead
- Compare resource consumption
- Inform mesh selection

**Tasks:**
- [ ] Create `experiments/mesh-benchmark/`
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

### 10.4 Runtime Performance Comparison

**Goal:** Compare web server runtimes for API workloads

**Learning objectives:**
- Benchmark different language runtimes
- Understand performance characteristics
- Portfolio piece for runtime expertise

**Tasks:**
- [ ] Create `experiments/runtime-benchmark/`
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

### 10.5 Cost-Efficiency Benchmarking

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

### 10.6 Cost Tooling (Kubecost / OpenCost)

**Goal:** Deploy cost allocation and visibility tooling

**Learning objectives:**
- Understand Kubernetes cost allocation models
- Deploy and configure cost visibility tools
- Integrate cost data with benchmarking

**Tasks:**
- [ ] Create `experiments/cost-tooling/`
- [ ] Deploy OpenCost (CNCF):
  - [ ] Helm deployment
  - [ ] Cloud provider pricing integration
  - [ ] Prometheus metrics export
- [ ] Deploy Kubecost (optional comparison):
  - [ ] Free tier deployment
  - [ ] Compare with OpenCost
- [ ] Cost allocation:
  - [ ] Namespace-level cost breakdown
  - [ ] Label-based cost attribution (team, project)
  - [ ] Idle resource identification
  - [ ] Right-sizing recommendations
- [ ] Integration with benchmarks:
  - [ ] Cost per experiment calculation
  - [ ] Before/after cost comparison
  - [ ] Cost anomaly alerting
- [ ] Grafana dashboards:
  - [ ] Cost by namespace
  - [ ] Cost trending over time
  - [ ] Cost efficiency metrics
- [ ] Document cost tooling patterns
- [ ] **ADR:** OpenCost vs Kubecost selection

---

### 10.7 Full Stack Composition Benchmark

**Goal:** Measure end-to-end system with all layers

**Learning objectives:**
- Understand how layers compound
- Isolate overhead by layer
- Make holistic optimization decisions

**Tasks:**
- [ ] Create `experiments/full-stack-benchmark/`
- [ ] Deploy full stack:
  ```
  Client → Gateway → Mesh → App → Database → Messaging
  ```
- [ ] Baseline measurements:
  - [ ] App only (no gateway, no mesh)
  - [ ] + Gateway overhead
  - [ ] + Mesh overhead
  - [ ] + Observability overhead
- [ ] End-to-end metrics:
  - [ ] Request latency breakdown by layer
  - [ ] Cost attribution by component
  - [ ] Resource consumption by layer
- [ ] Optimization experiments:
  - [ ] Remove mesh for internal traffic
  - [ ] Tune gateway connection pools
  - [ ] Optimize database queries
- [ ] Document full stack patterns
- [ ] **Portfolio piece:** Full stack architecture decision guide

---

### 10.8 Success Criteria

- [ ] All technology benchmarks documented with data
- [ ] Cost per transaction calculated for each option
- [ ] Full stack composition overhead understood
- [ ] Cost tooling deployed and integrated
- [ ] Data-driven ADRs for technology selection

