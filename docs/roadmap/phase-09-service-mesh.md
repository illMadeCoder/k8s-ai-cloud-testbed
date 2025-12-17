## Phase 9: Service Mesh

*Service mesh builds on traffic management with mTLS, observability, and advanced traffic control.*

### 9.0 Service Mesh Decision Framework

**Goal:** Understand when and which service mesh to use before diving into implementations

**Learning objectives:**
- Compare service mesh architectures (sidecar vs sidecar-free)
- Understand feature trade-offs between options
- Make informed mesh selection decisions

**Tasks:**
- [ ] Create `docs/service-mesh-comparison.md`
- [ ] Architecture comparison:
  - [ ] Sidecar proxy model (Istio, Linkerd)
  - [ ] Sidecar-free/eBPF model (Cilium)
  - [ ] Control plane architectures
- [ ] Feature matrix:
  - [ ] mTLS implementation differences
  - [ ] Traffic management capabilities
  - [ ] Observability integration
  - [ ] Multi-cluster support
- [ ] Operational considerations:
  - [ ] Resource overhead (CPU, memory per pod)
  - [ ] Upgrade complexity
  - [ ] Debugging difficulty
  - [ ] Learning curve
- [ ] When to use each:
  - [ ] Istio: Full feature set, complex policies needed
  - [ ] Linkerd: Simplicity, lower resource overhead
  - [ ] Cilium: Already using Cilium CNI, performance critical
  - [ ] No mesh: Simple applications, overhead not justified
- [ ] When NOT to use service mesh:
  - [ ] Small number of services
  - [ ] No mTLS requirement
  - [ ] Team lacks operational capacity
- [ ] Document decision criteria
- [ ] **ADR:** Document service mesh decision for this lab

*Note: Phase 12.3 provides detailed performance benchmarks after you've learned each mesh.*

---

### 9.1 Istio Deep Dive

**Goal:** Master Istio service mesh fundamentals

**Learning objectives:**
- Understand Istio architecture (control plane, data plane, sidecars)
- Configure traffic management policies
- Implement security with mTLS

**Tasks:**
- [ ] Create `experiments/scenarios/istio-tutorial/`
- [ ] Install Istio (istioctl or Helm)
- [ ] Enable sidecar injection (namespace label)
- [ ] Deploy sample microservices app
- [ ] Traffic management:
  - [ ] VirtualService routing rules
  - [ ] DestinationRule load balancing
  - [ ] Traffic splitting (canary)
  - [ ] Fault injection (delays, aborts)
  - [ ] Circuit breaking
  - [ ] Retries and timeouts
- [ ] Security:
  - [ ] Automatic mTLS (PeerAuthentication)
  - [ ] Authorization policies (allow/deny)
  - [ ] JWT validation (RequestAuthentication)
- [ ] Observability:
  - [ ] Kiali service graph
  - [ ] Distributed tracing (Jaeger integration)
  - [ ] Metrics (Prometheus integration)
- [ ] Gateway:
  - [ ] Istio Gateway (vs Gateway API)
  - [ ] External traffic management
- [ ] Document Istio patterns and gotchas

---

### 9.2 Linkerd Tutorial

**Goal:** Learn lightweight service mesh alternative

**Learning objectives:**
- Understand Linkerd architecture (simpler than Istio)
- Compare operational complexity
- Evaluate for different use cases

**Tasks:**
- [ ] Create `experiments/scenarios/linkerd-tutorial/`
- [ ] Install Linkerd (CLI + control plane)
- [ ] Inject proxies into workloads
- [ ] Deploy same sample app as Istio experiment
- [ ] Configure:
  - [ ] Automatic mTLS
  - [ ] Traffic splitting (TrafficSplit CRD)
  - [ ] Retries and timeouts (ServiceProfile)
  - [ ] Authorization policies
- [ ] Observability:
  - [ ] Linkerd dashboard
  - [ ] Tap for live traffic inspection
  - [ ] Metrics and golden signals
- [ ] Compare with Istio:
  - [ ] Resource consumption
  - [ ] Configuration complexity
  - [ ] Feature coverage
- [ ] Document when to choose Linkerd vs Istio

---

### 9.3 Cilium Service Mesh (eBPF)

**Goal:** Explore sidecar-free service mesh with eBPF

**Learning objectives:**
- Understand eBPF-based networking
- Compare sidecar vs sidecar-free architectures
- Evaluate Cilium for CNI + service mesh

**Tasks:**
- [ ] Create `experiments/scenarios/cilium-tutorial/`
- [ ] Install Cilium as CNI with service mesh features
- [ ] Deploy sample app (no sidecars needed)
- [ ] Configure:
  - [ ] L7 traffic policies (CiliumNetworkPolicy)
  - [ ] mTLS (Cilium encryption)
  - [ ] Load balancing
  - [ ] Ingress (Cilium Ingress or Gateway API)
- [ ] Observability:
  - [ ] Hubble for network visibility
  - [ ] Hubble UI
  - [ ] Prometheus metrics
- [ ] Compare with sidecar meshes:
  - [ ] Performance overhead
  - [ ] Resource consumption
  - [ ] Operational complexity
- [ ] Document eBPF advantages and limitations
- [ ] **ADR:** Document service mesh selection criteria

---

### 9.4 Cross-Cluster Networking

**Goal:** Enable service discovery and communication across multiple clusters

**Learning objectives:**
- Understand multi-cluster networking patterns
- Implement cross-cluster service discovery
- Design for geographic distribution

**Tasks:**
- [ ] Create `experiments/scenarios/cross-cluster-networking/`
- [ ] Evaluate and implement option:
  - [ ] **Cilium ClusterMesh** (if using Cilium CNI) OR
  - [ ] **Submariner** (CNI-agnostic)
- [ ] Cilium ClusterMesh path:
  - [ ] Enable ClusterMesh on multiple Kind clusters
  - [ ] Configure cluster peering
  - [ ] Global services (service available in all clusters)
  - [ ] Service affinity (prefer local, failover to remote)
- [ ] Submariner path:
  - [ ] Deploy Submariner broker
  - [ ] Join clusters to broker
  - [ ] ServiceExport/ServiceImport resources
  - [ ] Lighthouse DNS for service discovery
- [ ] Cross-cluster patterns:
  - [ ] Active-active service deployment
  - [ ] Failover scenarios
  - [ ] Latency-aware routing
- [ ] Security:
  - [ ] Encrypted tunnel between clusters
  - [ ] NetworkPolicy across clusters
  - [ ] Identity federation
- [ ] Testing:
  - [ ] Cross-cluster service call latency
  - [ ] Failover time measurement
  - [ ] Partition tolerance testing
- [ ] Document cross-cluster patterns
- [ ] **ADR:** Document multi-cluster networking decision

---

### 9.5 Multi-Cluster Federation & Orchestration

**Goal:** Orchestrate workloads across multiple clusters as a single logical platform

**Learning objectives:**
- Understand multi-cluster orchestration patterns
- Implement workload distribution across clusters
- Design for multi-cluster GitOps

**Tasks:**
- [ ] Create `experiments/scenarios/multi-cluster-federation/`
- [ ] Federation approaches:
  - [ ] **Liqo** - Virtual node approach, transparent offloading
  - [ ] **Admiralty** - Multi-cluster scheduling
  - [ ] **KubeFed** (if still relevant) - Federated resources
- [ ] Liqo deep dive:
  - [ ] Peer clusters
  - [ ] Virtual nodes and resource sharing
  - [ ] Namespace offloading
  - [ ] Pod scheduling across clusters
- [ ] Admiralty deep dive:
  - [ ] Source and target clusters
  - [ ] Multi-cluster scheduling policies
  - [ ] Follow/delegate pod model
- [ ] Multi-cluster GitOps:
  - [ ] ArgoCD multi-cluster management (hub model)
  - [ ] Fleet (Rancher) for GitOps at scale
  - [ ] ApplicationSets for cluster templating
- [ ] Workload placement strategies:
  - [ ] Cost-based placement
  - [ ] Latency-based placement
  - [ ] Compliance-based placement (data residency)
  - [ ] Capacity-based placement
- [ ] Federation challenges:
  - [ ] State management across clusters
  - [ ] Network connectivity requirements
  - [ ] Consistency vs availability trade-offs
- [ ] Document federation patterns
- [ ] **ADR:** Document multi-cluster orchestration approach

---

### 9.6 Service Mesh Cost Analysis

**Goal:** Understand and optimize the cost overhead of service mesh

*FinOps consideration: Service mesh adds resource overhead via sidecars. Quantify this cost and optimize where possible.*

**Learning objectives:**
- Measure sidecar resource consumption
- Compare mesh options by cost efficiency
- Optimize mesh configuration for cost

**Tasks:**
- [ ] Sidecar resource analysis:
  - [ ] CPU/memory per sidecar
  - [ ] Aggregate sidecar cost across cluster
  - [ ] Sidecar cost as percentage of workload cost
- [ ] Mesh comparison by cost:
  - [ ] Istio sidecar overhead
  - [ ] Linkerd sidecar overhead (typically lighter)
  - [ ] Cilium sidecar-less approach (eBPF)
  - [ ] No mesh baseline
- [ ] Cost optimization:
  - [ ] Sidecar resource limits tuning
  - [ ] Selective sidecar injection (not all namespaces)
  - [ ] Ambient mesh (Istio) for reduced overhead
- [ ] Control plane costs:
  - [ ] Control plane resource consumption
  - [ ] HA vs single-replica trade-offs
- [ ] Multi-cluster cost considerations:
  - [ ] Cross-cluster traffic costs (egress)
  - [ ] Gateway resource costs
- [ ] Document mesh cost analysis
- [ ] **ADR:** Document mesh selection with cost as factor

---

