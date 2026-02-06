## Phase 3: Observability Stack

*You need to see what's happening before you can improve it. These skills are used in every subsequent experiment.*

*Section order follows the "three pillars" pattern: metrics → logs → traces, then SLOs once you understand what you're measuring.*

> **Deep Dive:** For eBPF-based observability (I/O tracing, Pixie, Parca, Tetragon), see [Appendix T: eBPF & Advanced Metrics](docs/roadmap/appendix-t-ebpf-metrics.md). This phase covers standard Prometheus/Loki/Tempo fundamentals.

### 3.1 Prometheus & Grafana Deep Dive

**Goal:** Master metrics collection, PromQL, alerting, and dashboards

**Learning objectives:**
- Understand Prometheus architecture (scraping, TSDB, federation)
- Write effective PromQL queries
- Build actionable Grafana dashboards
- Configure alerting pipelines

**Tasks:**
- [x] Create `experiments/prometheus-tutorial/`
- [x] Deploy kube-prometheus-stack via ArgoCD
- [x] Build sample app with custom metrics:
  - [x] Counter (http_requests_total)
  - [x] Gauge (active_connections)
  - [x] Histogram (request_duration_seconds)
  - [ ] Summary (response_size_bytes)
- [x] Create ServiceMonitor for scrape discovery
- [x] Write PromQL tutorial queries:
  - [x] rate() and irate() for counters
  - [x] Aggregations (sum, avg, max by labels)
  - [x] histogram_quantile() for percentiles
  - [ ] absent() for missing metric alerts
  - [ ] predict_linear() for capacity planning
- [x] Build Grafana dashboards:
  - [x] RED metrics (Rate, Errors, Duration)
  - [ ] USE metrics (Utilization, Saturation, Errors)
  - [ ] Dashboard variables and templating
- [ ] Configure alerting:
  - [x] PrometheusRule CRDs
  - [ ] Alertmanager routing and silences
  - [ ] Alert grouping and inhibition
- [ ] Document PromQL patterns and anti-patterns
- [x] **Explore TSDB alternatives:** (see ADR-009, `experiments/tsdb-comparison/`)
  - [x] Victoria Metrics (drop-in replacement, better compression)
  - [x] Compare resource usage (Prometheus vs Victoria Metrics vs Mimir)
  - [x] Mimir architecture overview (Grafana's distributed TSDB)
  - [ ] InfluxDB comparison (different query language, use cases)
  - [x] **ADR:** Document TSDB selection criteria (ADR-009)

---

### 3.2 SeaweedFS Object Storage

**Goal:** Deploy S3-compatible object storage using Facebook Haystack-inspired architecture

*SeaweedFS replaces MinIO (see ADR-008). Loki, Thanos, Tempo, Velero, and Argo Workflows all need S3-compatible storage.*

**Learning objectives:**
- Understand Haystack architecture (volumes, O(1) lookups)
- Compare to traditional filesystem performance
- Configure S3 gateway for observability backends

**Why SeaweedFS over MinIO:**
- MinIO entered maintenance mode (Dec 2025)
- Apache 2.0 license (MinIO is AGPL)
- O(1) disk seeks for small files
- Lower resource usage

**Tasks:**
- [x] Create `experiments/seaweedfs-tutorial/`
- [x] Deploy SeaweedFS (master + volume servers)
- [x] Understand the architecture:
  - [x] Master server (metadata, volume allocation)
  - [x] Volume servers (32GB volumes containing millions of files)
  - [x] Filer (optional filesystem layer)
  - [x] S3 Gateway (S3 API compatibility)
- [x] **Haystack demo: "Needle in a Haystack"**
  - [x] Store 10,000 small sensor readings (scaled down for Kind)
  - [x] Demonstrate O(1) lookup time (constant regardless of file count)
  - [x] Compare to filesystem degradation (LOSF benchmark)
  - [ ] Visualize volume packing
- [ ] Configure S3 gateway:
  - [ ] Create buckets for observability
  - [ ] `loki-chunks` - log storage
  - [ ] `thanos-blocks` - metrics long-term storage
  - [ ] `tempo-traces` - trace storage
- [x] Monitoring:
  - [x] SeaweedFS metrics in Prometheus (ServiceMonitors)
  - [x] Volume capacity dashboards (Grafana)
  - [x] Alert on disk usage (PrometheusRules)
- [x] Document object storage patterns (ADR-008)

---

### 3.3 Loki & Log Aggregation

**Goal:** Centralized logging with Loki and LogQL

*Requires: Section 3.2 (SeaweedFS) for log chunk storage*

**Learning objectives:**
- Understand Loki's label-based architecture (vs full-text indexing)
- Write effective LogQL queries
- Correlate logs with metrics in Grafana

**Tasks:**
- [ ] Create `experiments/loki-tutorial/`
- [ ] Deploy Loki stack (Loki + Promtail)
- [ ] Configure Loki storage:
  - [ ] Point to SeaweedFS bucket from Section 3.2
  - [ ] Configure retention policies
- [ ] Build app with structured JSON logging
- [ ] Configure Promtail pipelines:
  - [ ] Label extraction (namespace, pod, container)
  - [ ] JSON field parsing
  - [ ] Regex extraction
  - [ ] Drop/keep filtering
  - [ ] Multiline log handling
- [ ] Write LogQL tutorial:
  - [ ] Label matchers and line filters
  - [ ] Parser expressions (json, pattern, regexp)
  - [ ] Metric queries (rate, count_over_time)
  - [ ] Unwrap for numeric fields
- [ ] Build log dashboards:
  - [ ] Log panel with live tail
  - [ ] Log volume over time
  - [ ] Error log filtering
- [ ] Set up log-based alerts (error rate threshold)
- [ ] Correlate logs ↔ metrics in Grafana (split view)
- [ ] Document logging best practices

---

### 3.3b ELK Stack (Alternative to Loki)

**Goal:** Centralized logging with Elasticsearch, Logstash/Fluentd, and Kibana

*Alternative to Loki for teams needing full-text search or existing ELK expertise.*

**Learning objectives:**
- Understand ELK architecture vs Loki trade-offs
- Configure Elasticsearch for log storage
- Build Kibana dashboards and visualizations
- Implement log pipelines with Fluentd/Logstash

**When to choose ELK over Loki:**
- Need full-text search across log content
- Existing ELK expertise on team
- Complex log parsing requirements
- Need for Kibana's visualization capabilities
- Log analytics beyond simple filtering

**When to choose Loki over ELK:**
- Already using Grafana (single pane of glass)
- Resource-constrained environments
- Label-based queries are sufficient
- Simpler operational overhead
- Cost-sensitive deployments

**Tasks:**
- [ ] Create `experiments/elk-tutorial/`
- [ ] Deploy ELK stack via ECK (Elastic Cloud on Kubernetes):
  - [ ] Elasticsearch cluster (single node dev, 3-node prod)
  - [ ] Kibana for visualization
  - [ ] Elastic Agent or Fluentd for collection
- [ ] Configure Elasticsearch:
  - [ ] Index templates and mappings
  - [ ] Index lifecycle management (ILM)
  - [ ] Retention policies (hot/warm/cold/delete)
  - [ ] Shard sizing and replica configuration
- [ ] Build log pipelines:
  - [ ] Fluentd/Fluent Bit DaemonSet
  - [ ] Parse Kubernetes metadata
  - [ ] JSON log parsing
  - [ ] Multi-line log handling
  - [ ] Field extraction and enrichment
- [ ] Kibana dashboards:
  - [ ] Log explorer and search
  - [ ] Log volume visualizations
  - [ ] Error rate dashboards
  - [ ] Index pattern configuration
- [ ] Alerting:
  - [ ] Kibana alerting rules
  - [ ] Watcher for complex conditions
  - [ ] Integration with notification channels
- [ ] Performance tuning:
  - [ ] JVM heap sizing
  - [ ] Bulk indexing optimization
  - [ ] Query performance analysis
- [x] Compare with Loki: (see `experiments/logging-comparison/`)
  - [x] Resource usage comparison dashboard
  - [x] Query performance benchmark scripts
  - [ ] Operational complexity comparison
  - [ ] Cost analysis (storage, compute)
- [ ] Document ELK patterns and anti-patterns
- [x] **ADR:** Document logging stack selection criteria (ADR-010)

**Note:** The `logging-comparison` tutorial deploys both stacks side-by-side.
- Uses plain K8s manifests for ES/Kibana (Helm chart has single-node conflicts)
- Loki + Promtail via Helm charts
- Grafana with both datasources and comparison dashboard

---

### 3.4 OpenTelemetry & Distributed Tracing

**Goal:** End-to-end observability with traces, connecting metrics and logs

*Requires: Section 3.1 (Prometheus), Section 3.2 (MinIO). Optional: Section 3.3 (Loki) for log correlation.*

**Learning objectives:**
- Understand OpenTelemetry architecture (SDK, Collector, backends)
- Instrument applications for distributed tracing
- Correlate traces ↔ metrics ↔ logs

**Tasks:**
- [ ] Create `experiments/opentelemetry-tutorial/`
- [ ] Deploy OpenTelemetry Collector
- [ ] Deploy Tempo (using MinIO for storage) or Jaeger as trace backend
- [ ] Build multi-service demo app (3+ services):
  - [ ] Service A → Service B → Service C
  - [ ] Each service instrumented with OTel SDK
- [ ] Implement tracing:
  - [ ] Auto-instrumentation (HTTP, gRPC, DB)
  - [ ] Manual span creation
  - [ ] Span attributes and events
  - [ ] Context propagation (W3C Trace Context)
- [ ] Configure Collector:
  - [ ] OTLP receiver
  - [ ] Batch processor
  - [ ] Exporters (Tempo/Jaeger, Prometheus)
- [ ] Connect the three pillars:
  - [ ] Exemplars (metrics → traces)
  - [ ] Trace ID in logs (logs → traces)
  - [ ] Service graph from traces
- [ ] Build trace-aware dashboards:
  - [ ] Service dependency map
  - [ ] Latency breakdown by span
  - [ ] Error trace exploration
- [ ] Document sampling strategies (head vs tail)

---

### 3.5 SLOs & Error Budgets

**Goal:** Implement Service Level Objectives for reliability-driven operations

*SLOs come after the three pillars so you understand what you're measuring. Used throughout: canary analysis, deployment decisions, capacity planning.*

**Learning objectives:**
- Understand SLI/SLO/SLA hierarchy
- Implement error budget tracking
- Use SLOs to drive architectural decisions

**Tasks:**
- [ ] Create `experiments/slo-tutorial/`
- [ ] Deploy SLO tooling:
  - [ ] Sloth (SLO generator for Prometheus)
  - [ ] Pyrra (SLO dashboards and alerts)
- [ ] Define SLIs for demo application:
  - [ ] Availability SLI (successful requests / total requests)
  - [ ] Latency SLI (requests < threshold / total requests)
  - [ ] Throughput SLI (if applicable)
- [ ] Create SLO specifications:
  - [ ] 99.9% availability (43.8 min/month error budget)
  - [ ] 99% latency < 200ms
  - [ ] Multi-window, multi-burn-rate alerts
- [ ] Error budget tracking:
  - [ ] Error budget remaining dashboard
  - [ ] Burn rate visualization
  - [ ] Budget depletion forecasting
- [ ] SLO-driven alerting:
  - [ ] Fast burn alerts (immediate action)
  - [ ] Slow burn alerts (trending toward breach)
  - [ ] Error budget exhaustion alerts
- [ ] Operational integration:
  - [ ] SLO review process
  - [ ] Error budget policy (freeze deploys when exhausted)
  - [ ] SLO-based incident prioritization
- [ ] Document SLO patterns and anti-patterns
- [ ] **ADR:** Document SLO strategy and target selection

---

### 3.6 Observability Cost Management

**Goal:** Understand and optimize the cost of observability systems

*FinOps consideration: Observability can become one of the largest infrastructure costs. Plan for this from the start.*

**Learning objectives:**
- Understand observability cost drivers
- Implement retention and sampling strategies
- Balance visibility with cost

**Tasks:**
- [ ] Metrics cost analysis:
  - [ ] Cardinality explosion detection
  - [ ] High-cardinality label identification
  - [ ] Unused metrics identification
  - [ ] Retention policy cost modeling
- [ ] Logs cost analysis:
  - [ ] Log volume by source
  - [ ] Noisy log identification
  - [ ] Sampling strategies for high-volume logs
  - [ ] Retention tiers (hot/warm/cold)
- [ ] Traces cost analysis:
  - [ ] Sampling rate optimization
  - [ ] Head-based vs tail-based sampling
  - [ ] Trace retention policies
- [ ] Storage optimization:
  - [ ] Downsampling for historical data (5m, 1h resolution)
  - [ ] Compression analysis
  - [ ] Storage tier selection (SSD vs HDD vs object)
- [ ] Cost allocation:
  - [ ] Per-namespace observability costs
  - [ ] Chargeback for heavy consumers
  - [ ] Budget alerts for observability spend
- [ ] Document observability cost patterns
- [ ] **ADR:** Document observability retention and sampling strategy

---
