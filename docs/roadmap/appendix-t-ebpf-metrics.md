# Appendix T: eBPF & Advanced Metrics

*Deep system visibility with eBPF. Go beyond CPU/RAM to understand I/O, network, and kernel behavior. Use after completing Phase 3 (Observability).*

---

## Overview

Traditional metrics (CPU, RAM) miss critical dimensions:

| Metric | What It Reveals |
|--------|-----------------|
| **Block I/O** | Disk latency, IOPS, saturation |
| **Network I/O** | TCP retransmits, connection churn |
| **File System** | Page cache efficiency, VFS operations |
| **System Calls** | Kernel overhead, syscall patterns |

**Why eBPF?**
- Zero overhead when not tracing
- Production-safe (kernel verifier)
- No code changes required
- Real-time visibility

---

## T.1 eBPF Fundamentals

**Goal:** Understand eBPF architecture and BCC tools

**Learning objectives:**
- eBPF program types and hooks
- BCC tools suite
- bpftrace for ad-hoc tracing

**Tasks:**
- [ ] Create `experiments/ebpf-tutorial/`
- [ ] Deploy BCC tools container:
  ```yaml
  spec:
    hostPID: true
    hostNetwork: true
    containers:
    - name: bcc-tools
      image: zlim/bcc
      securityContext:
        privileged: true
  ```
- [ ] Core tools:
  - [ ] `biosnoop` - Block I/O latency
  - [ ] `biotop` - Top by I/O
  - [ ] `tcptop` - TCP throughput
  - [ ] `tcpretrans` - Retransmit tracing
  - [ ] `cachestat` - Page cache stats
  - [ ] `execsnoop` - Process execution
- [ ] bpftrace basics:
  - [ ] One-liner tracing
  - [ ] Custom scripts
- [ ] Document eBPF tools and use cases

---

## T.2 Block I/O Analysis

**Goal:** Understand disk performance with eBPF

**The Problem:**
```
Scenario: PostgreSQL slow
- CPU: 15%  ✓
- RAM: 40%  ✓
- Disk p99: 500ms  ← PROBLEM!
- IOPS: 95% saturated

Traditional metrics miss this.
```

**Tasks:**
- [ ] Create `experiments/io-bottleneck/`
- [ ] Metrics to capture:
  - [ ] Latency (p50, p95, p99)
  - [ ] IOPS (read/write)
  - [ ] Throughput (MB/s)
  - [ ] Queue depth
  - [ ] Device saturation
- [ ] Tools:
  - [ ] `biosnoop` - Per-I/O latency
  - [ ] `biotop` - Top processes by I/O
  - [ ] `iostat` - Classic stats
- [ ] Experiment:
  - [ ] Create I/O-bound workload
  - [ ] Measure with biosnoop
  - [ ] Show CPU/RAM miss the problem
- [ ] Document I/O analysis patterns

---

## T.3 Network I/O Analysis

**Goal:** Understand network behavior with eBPF

**The Problem:**
```
Scenario: Service mesh latency spikes
- CPU: 30%  ✓
- RAM: 50%  ✓
- TCP retransmits: 5%  ← PROBLEM!
- Socket buffer: 90% full

Root cause: Congestion causing retransmits
```

**Tasks:**
- [ ] Create `experiments/network-analysis/`
- [ ] Metrics to capture:
  - [ ] TCP connections (active, rate)
  - [ ] Retransmits (% of packets)
  - [ ] Socket buffer saturation
  - [ ] Connection lifespan
- [ ] Tools:
  - [ ] `tcptop` - Throughput by connection
  - [ ] `tcpretrans` - Retransmit tracer
  - [ ] `tcplife` - Connection duration
  - [ ] `ss` - Socket statistics
- [ ] Experiment:
  - [ ] Deploy service mesh
  - [ ] Generate high throughput
  - [ ] Trace retransmits to services
- [ ] Document network analysis patterns

---

## T.4 Page Cache & Memory

**Goal:** Understand memory efficiency with eBPF

**The Problem:**
```
Scenario: Observability queries slow
- CPU: 25%  ✓
- RAM: 60%  ✓
- Page cache hit: 40%  ← PROBLEM! (should be >90%)
- VFS reads: 10k/sec hitting disk

Root cause: Insufficient RAM for index caching
```

**Tasks:**
- [ ] Create `experiments/cache-analysis/`
- [ ] Metrics to capture:
  - [ ] Page cache hit/miss rate
  - [ ] VFS operations (read/write/open)
  - [ ] Dirty pages
  - [ ] Writeback frequency
- [ ] Tools:
  - [ ] `cachestat` - Cache hit/miss
  - [ ] `vfsstat` - VFS operation rate
  - [ ] `filetop` - Files by I/O
- [ ] Experiment:
  - [ ] Database query workload
  - [ ] Measure cache hit rate
  - [ ] Show RAM impact on performance
- [ ] Document cache analysis patterns

---

## T.5 Pixie (Auto-Instrumented Observability)

**Goal:** Deploy Pixie for automatic eBPF-based observability

**What Pixie Provides:**
- Auto-instrumented (no code changes)
- Network map (service topology)
- Request tracing (HTTP/gRPC/DNS)
- Resource profiling (CPU flamegraphs)

**Tasks:**
- [ ] Create `experiments/pixie-tutorial/`
- [ ] Deploy Pixie:
  ```bash
  helm install pixie pixie-operator/pixie-operator-chart \
    --set deployKey=$PIXIE_DEPLOY_KEY
  ```
- [ ] Explore:
  - [ ] Service map
  - [ ] Request traces
  - [ ] Flamegraphs
- [ ] PxL queries:
  ```python
  # Find slow database queries
  px.DataFrame(table='mysql.query', start_time='-5m')
    .groupby('query')
    .agg(latency_p99=('latency', px.quantiles, 0.99))
    .filter(latency_p99 > 100)
  ```
- [ ] Document Pixie patterns

---

## T.6 Parca (Continuous Profiling)

**Goal:** Always-on CPU and memory profiling

**What Parca Provides:**
- Continuous profiling (always on)
- Flamegraphs
- Differential profiling (before/after)
- Historical analysis

**Tasks:**
- [ ] Create `experiments/parca-tutorial/`
- [ ] Deploy Parca:
  ```bash
  helm install parca parca/parca
  ```
- [ ] Use cases:
  - [ ] Profile Go/Rust/.NET runtimes
  - [ ] Identify hot code paths
  - [ ] Before/after optimization
- [ ] Document profiling patterns

---

## T.7 Tetragon (Runtime Security)

**Goal:** eBPF-based runtime security observability

**What Tetragon Provides:**
- Process execution tracking
- Network connections by pod
- File access monitoring
- Syscall filtering

**Tasks:**
- [ ] Create `experiments/tetragon-tutorial/`
- [ ] Deploy Tetragon:
  ```bash
  helm install tetragon cilium/tetragon
  ```
- [ ] Use cases:
  - [ ] Detect unexpected process spawns
  - [ ] Track outbound connections
  - [ ] File access auditing
- [ ] Document security observability patterns

---

## T.8 I/O-Aware Benchmarking

**Goal:** Add I/O metrics to database and storage benchmarks

**Full Stack I/O Attribution:**
```
End-to-end p99: 200ms

Breakdown with eBPF:
├─ Gateway: 5ms (network I/O: 3ms)
├─ Mesh sidecar: 10ms (network I/O: 7ms)
├─ App: 50ms (disk I/O: 5ms)
├─ Database: 130ms (disk I/O: 100ms!) ← 50% of total
└─ Messaging: 5ms (disk I/O: 2ms)

Action: Add RAM for database index caching
```

**Tasks:**
- [ ] Integrate eBPF into Phase 5 database benchmark
- [ ] Add I/O breakdown to Phase 10 capstone
- [ ] Cost analysis:
  - [ ] "We thought we needed CPU, but eBPF showed disk"
  - [ ] Right-size based on actual bottleneck
- [ ] Document I/O-aware benchmarking

---

## Tools Summary

| Tool | Purpose | Phase Integration |
|------|---------|-------------------|
| **biosnoop** | Block I/O latency | Database benchmarks |
| **tcptop** | TCP throughput | Service mesh analysis |
| **tcpretrans** | Retransmit tracing | Network debugging |
| **cachestat** | Page cache efficiency | Memory tuning |
| **Pixie** | Auto-instrumented observability | Full stack visibility |
| **Parca** | Continuous profiling | Performance optimization |
| **Tetragon** | Runtime security | Security observability |

---

## Cross-References

| Topic | Location |
|-------|----------|
| Basic observability | Phase 3: Observability |
| Database benchmarking | Phase 5: Data & Persistence |
| Service mesh overhead | Phase 7: Service Mesh |
| Full stack benchmark | Phase 10: Performance & Cost |

---

## When to Use This Appendix

- Debugging performance issues CPU/RAM metrics can't explain
- Understanding I/O bottlenecks in databases
- Analyzing service mesh network overhead
- Implementing continuous profiling
- Building security observability
