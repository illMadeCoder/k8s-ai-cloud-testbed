## Phase 11: Autoscaling & Resource Management

*Scale applications and infrastructure efficiently based on various signals.*

### 11.1 Horizontal Pod Autoscaler Deep Dive

**Goal:** Master HPA configuration for different workloads

**Learning objectives:**
- Understand HPA algorithm and behavior
- Configure for various metric types
- Tune for responsiveness vs stability

**Tasks:**
- [ ] Create `experiments/scenarios/hpa-tutorial/`
- [ ] Build test app with configurable CPU/memory load
- [ ] Configure HPA scenarios:
  - [ ] CPU-based scaling
  - [ ] Memory-based scaling
  - [ ] Custom metrics (Prometheus adapter)
  - [ ] External metrics
- [ ] Tune parameters:
  - [ ] Target utilization thresholds
  - [ ] Stabilization windows (scale up/down)
  - [ ] Scaling policies (pods vs percent)
- [ ] Test workload patterns:
  - [ ] Gradual ramp-up
  - [ ] Sudden spike
  - [ ] Oscillating load
- [ ] Measure:
  - [ ] Time to scale
  - [ ] Over/under provisioning
  - [ ] Request latency during scaling
- [ ] Document HPA tuning guide

---

### 11.2 KEDA Event-Driven Autoscaling

**Goal:** Scale based on external event sources

*Note: Kafka/RabbitMQ scalers require Phase 7 (Messaging) knowledge*

**Learning objectives:**
- Understand KEDA architecture
- Configure various scalers
- Implement scale-to-zero

**Tasks:**
- [ ] Create `experiments/scenarios/keda-tutorial/`
- [ ] Install KEDA
- [ ] Implement scalers:
  - [ ] Prometheus scaler (custom metrics)
  - [ ] Kafka scaler (consumer lag)
  - [ ] RabbitMQ scaler (queue depth)
  - [ ] Cron scaler (scheduled scaling)
  - [ ] Azure Service Bus / AWS SQS (via Crossplane)
- [ ] Configure ScaledObject:
  - [ ] Triggers and thresholds
  - [ ] Cooldown periods
  - [ ] Min/max replicas
  - [ ] Scale-to-zero behavior
- [ ] Test ScaledJob for batch workloads
- [ ] Compare KEDA vs HPA:
  - [ ] Configuration complexity
  - [ ] Supported triggers
  - [ ] Scale-to-zero capability
- [ ] Document KEDA patterns

---

### 11.3 Vertical Pod Autoscaler

**Goal:** Right-size pod resource requests automatically

**Learning objectives:**
- Understand VPA modes and recommendations
- Combine VPA with HPA
- Implement resource optimization workflow

**Tasks:**
- [ ] Create `experiments/scenarios/vpa-tutorial/`
- [ ] Install VPA
- [ ] Configure VPA modes:
  - [ ] Off (recommendations only)
  - [ ] Initial (set on pod creation)
  - [ ] Auto (update running pods)
- [ ] Test with various workloads:
  - [ ] CPU-bound application
  - [ ] Memory-bound application
  - [ ] Variable workload
- [ ] Analyze recommendations:
  - [ ] Lower bound, target, upper bound
  - [ ] Uncapped vs capped
- [ ] Combine with HPA (mutually exclusive metrics)
- [ ] Document resource optimization workflow

---

### 11.4 Cluster Autoscaling

**Goal:** Automatically scale cluster nodes based on workload demand

**Learning objectives:**
- Understand Cluster Autoscaler vs Karpenter
- Configure node pools and scaling policies
- Optimize for cost and performance

**Tasks:**
- [ ] Create `experiments/scenarios/cluster-autoscaler-tutorial/`
- [ ] Implement Cluster Autoscaler (AKS/EKS):
  - [ ] Node pool configuration
  - [ ] Scale-down policies
  - [ ] Pod disruption budgets interaction
- [ ] Implement Karpenter (EKS):
  - [ ] Provisioner configuration
  - [ ] Instance type selection
  - [ ] Spot vs on-demand
  - [ ] Consolidation policies
- [ ] Test scenarios:
  - [ ] Scale-up on pending pods
  - [ ] Scale-down on low utilization
  - [ ] Node replacement (spot interruption)
- [ ] Cost optimization:
  - [ ] Right-sizing node types
  - [ ] Spot instance integration
  - [ ] Reserved capacity planning
- [ ] Measure:
  - [ ] Time to provision new node
  - [ ] Scale-down delay
  - [ ] Cost per workload
- [ ] Document cluster autoscaling patterns
- [ ] **ADR:** Document Cluster Autoscaler vs Karpenter decision

---

### 11.5 Production Multi-Tenancy

**Goal:** Scale multi-tenant patterns for production with resource management

*Requires: Phase 3.8 (Multi-Tenancy Security) for isolation foundations*

**Learning objectives:**
- Implement resource fairness and quotas at scale
- Design blast radius boundaries
- Automate tenant lifecycle

**Tasks:**
- [ ] Create `experiments/scenarios/multi-tenancy-production/`
- [ ] Build on Phase 3.8 security foundations:
  - [ ] Verify isolation from Phase 3.8 still holds
  - [ ] Add resource management layer
- [ ] Hierarchical namespaces (HNC):
  - [ ] Deploy Hierarchical Namespace Controller
  - [ ] Parent/child namespace inheritance
  - [ ] Propagated resources (secrets, configmaps)
  - [ ] Quota inheritance across hierarchy
- [ ] Resource quotas and limits:
  - [ ] ResourceQuotas per tenant namespace
  - [ ] LimitRanges for default pod resources
  - [ ] Aggregate quotas across tenant hierarchy
- [ ] Resource fairness:
  - [ ] PriorityClasses for tenant workloads
  - [ ] Pod priority and preemption rules
  - [ ] Fair-share scheduling concepts
- [ ] Noisy neighbor mitigation:
  - [ ] CPU/memory limits enforcement
  - [ ] I/O throttling patterns (if supported)
  - [ ] Network bandwidth limits (Cilium bandwidth manager)
- [ ] Tenant onboarding automation:
  - [ ] GitOps-driven tenant provisioning
  - [ ] Crossplane XRD for tenant creation
  - [ ] Automatic policy/quota application
- [ ] Tenant observability:
  - [ ] Per-tenant dashboards
  - [ ] Tenant-scoped alerting
  - [ ] Resource usage reporting
- [ ] Document production multi-tenancy patterns
- [ ] **ADR:** Document tenancy scaling decisions

---

### 11.6 FinOps Implementation & Chargeback

**Goal:** Full cost management with multi-tenant attribution

*Requires: Phase 1.4 (FinOps Foundation), Phase 4.1 (Prometheus), Phase 9.5 (Multi-Tenancy)*

**Learning objectives:**
- Implement per-tenant cost tracking
- Build chargeback/showback workflows
- Create cost optimization automation

**Tasks:**
- [ ] Create `experiments/scenarios/finops-implementation/`
- [ ] Deploy full Kubecost or OpenCost:
  - [ ] Integration with cloud billing APIs
  - [ ] Azure Cost Management connection
  - [ ] AWS Cost Explorer connection
- [ ] Per-tenant cost attribution:
  - [ ] Cost by namespace (tenant)
  - [ ] Cost by label (team, project, cost-center)
  - [ ] Shared cost distribution (control plane, monitoring)
- [ ] Cost dashboards:
  - [ ] Daily/weekly/monthly trends
  - [ ] Tenant cost comparison
  - [ ] Idle resource identification
  - [ ] Right-sizing recommendations
- [ ] Chargeback workflows:
  - [ ] Automated cost reports per tenant
  - [ ] Budget allocation per tenant
  - [ ] Overage notifications
- [ ] Cost optimization:
  - [ ] Spot instance savings analysis
  - [ ] Reserved capacity recommendations
  - [ ] Resource right-sizing automation
- [ ] Alerts and governance:
  - [ ] Budget threshold alerts
  - [ ] Anomaly detection
  - [ ] Cost forecasting
- [ ] Document FinOps implementation patterns

---

### 11.7 Serverless & Knative

**Goal:** Implement serverless patterns with scale-to-zero

**Learning objectives:**
- Understand Knative Serving and Eventing
- Implement request-driven autoscaling
- Compare serverless vs traditional deployment

**Tasks:**
- [ ] Create `experiments/scenarios/serverless-knative/`
- [ ] Knative Serving:
  - [ ] Deploy Knative Serving
  - [ ] Create Knative Services
  - [ ] Scale-to-zero configuration
  - [ ] Concurrency-based autoscaling
  - [ ] Revision management and traffic splitting
- [ ] Cold start optimization:
  - [ ] Minimum scale settings
  - [ ] Container startup time optimization
  - [ ] Keep-alive strategies
- [ ] Knative Eventing:
  - [ ] Event sources (Kafka, GitHub, Cron)
  - [ ] Brokers and triggers
  - [ ] Event-driven function invocation
  - [ ] CloudEvents format
- [ ] Comparison with alternatives:
  - [ ] Knative vs KEDA scale-to-zero
  - [ ] Knative vs OpenFaaS
  - [ ] When to use serverless vs always-on
- [ ] Measure:
  - [ ] Cold start latency
  - [ ] Request latency under load
  - [ ] Cost savings from scale-to-zero
- [ ] Document serverless patterns
- [ ] **ADR:** Document serverless vs traditional deployment decision

---

