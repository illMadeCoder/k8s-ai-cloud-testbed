## Phase 3: Observability Stack

*You need to see what's happening before you can improve it. These skills are used in every subsequent experiment.*

### 4.1 Prometheus & Grafana Deep Dive

**Goal:** Master metrics collection, PromQL, alerting, and dashboards

**Learning objectives:**
- Understand Prometheus architecture (scraping, TSDB, federation)
- Write effective PromQL queries
- Build actionable Grafana dashboards
- Configure alerting pipelines

**Tasks:**
- [ ] Create `experiments/scenarios/prometheus-tutorial/`
- [ ] Deploy kube-prometheus-stack via ArgoCD
- [ ] Build sample app with custom metrics:
  - [ ] Counter (http_requests_total)
  - [ ] Gauge (active_connections)
  - [ ] Histogram (request_duration_seconds)
  - [ ] Summary (response_size_bytes)
- [ ] Create ServiceMonitor for scrape discovery
- [ ] Write PromQL tutorial queries:
  - [ ] rate() and irate() for counters
  - [ ] Aggregations (sum, avg, max by labels)
  - [ ] histogram_quantile() for percentiles
  - [ ] absent() for missing metric alerts
  - [ ] predict_linear() for capacity planning
- [ ] Build Grafana dashboards:
  - [ ] RED metrics (Rate, Errors, Duration)
  - [ ] USE metrics (Utilization, Saturation, Errors)
  - [ ] Dashboard variables and templating
- [ ] Configure alerting:
  - [ ] PrometheusRule CRDs
  - [ ] Alertmanager routing and silences
  - [ ] Alert grouping and inhibition
- [ ] Document PromQL patterns and anti-patterns

---

### 4.2 SLOs & Error Budgets

**Goal:** Implement Service Level Objectives for reliability-driven operations

*SLOs are taught early because they're used throughout: canary analysis, deployment decisions, capacity planning.*

**Learning objectives:**
- Understand SLI/SLO/SLA hierarchy
- Implement error budget tracking
- Use SLOs to drive architectural decisions

**Tasks:**
- [ ] Create `experiments/scenarios/slo-tutorial/`
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

### 4.3 MinIO Object Storage

**Goal:** Deploy S3-compatible object storage as foundation for observability backends

*MinIO is taught here because Loki, Thanos, Tempo, Velero, and Argo Workflows all need object storage.*

**Learning objectives:**
- Understand MinIO architecture
- Configure for observability use cases
- Establish storage foundation for later phases

**Tasks:**
- [ ] Create `experiments/scenarios/minio-tutorial/`
- [ ] Deploy MinIO operator
- [ ] Create MinIO tenant:
  - [ ] Single node (development)
  - [ ] Multi-node distributed (HA)
- [ ] Configure:
  - [ ] Buckets and policies
  - [ ] Access keys and IAM
  - [ ] Lifecycle rules
  - [ ] Versioning
- [ ] Create buckets for observability:
  - [ ] `loki-chunks` - for log storage
  - [ ] `thanos-blocks` - for metrics long-term storage
  - [ ] `tempo-traces` - for trace storage
  - [ ] `velero-backups` - for cluster backups (Phase 10)
  - [ ] `argo-artifacts` - for workflow artifacts (Phase 13)
- [ ] Monitoring:
  - [ ] MinIO metrics in Prometheus
  - [ ] Storage capacity dashboards
  - [ ] Alert on bucket growth
- [ ] Document object storage patterns

---

### 4.4 Loki & Log Aggregation

**Goal:** Centralized logging with Loki and LogQL

*Requires: Phase 4.3 (MinIO) for log chunk storage*

**Learning objectives:**
- Understand Loki's label-based architecture (vs full-text indexing)
- Write effective LogQL queries
- Correlate logs with metrics in Grafana

**Tasks:**
- [ ] Create `experiments/scenarios/loki-tutorial/`
- [ ] Deploy Loki stack (Loki + Promtail)
- [ ] Configure Loki storage:
  - [ ] Point to MinIO bucket from Phase 4.3
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

### 4.4b ELK Stack (Alternative to Loki)

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
- [ ] Create `experiments/scenarios/elk-tutorial/`
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
- [ ] Compare with Loki:
  - [ ] Resource usage comparison
  - [ ] Query performance comparison
  - [ ] Operational complexity comparison
  - [ ] Cost analysis (storage, compute)
- [ ] Document ELK patterns and anti-patterns
- [ ] **ADR:** Document logging stack selection criteria

---

### 4.5 OpenTelemetry & Distributed Tracing

**Goal:** End-to-end observability with traces, connecting metrics and logs

*Requires: Phase 4.1 (Prometheus), Phase 4.3 (MinIO), Phase 4.4 (Loki)*

**Learning objectives:**
- Understand OpenTelemetry architecture (SDK, Collector, backends)
- Instrument applications for distributed tracing
- Correlate traces ↔ metrics ↔ logs

**Tasks:**
- [ ] Create `experiments/scenarios/opentelemetry-tutorial/`
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

### 4.6 Thanos for Multi-Cluster Metrics

**Goal:** Long-term metrics storage and global query view across clusters

*Requires: Phase 4.1 (Prometheus), Phase 4.3 (MinIO)*

**Learning objectives:**
- Understand Thanos architecture (Sidecar, Store, Query, Compactor)
- Implement multi-cluster metrics aggregation
- Configure long-term retention with object storage

**Tasks:**
- [ ] Create `experiments/scenarios/thanos-tutorial/`
- [ ] Deploy Thanos components:
  - [ ] Sidecar (alongside Prometheus)
  - [ ] Store Gateway (for object storage queries)
  - [ ] Query (global query layer)
  - [ ] Compactor (downsampling and retention)
- [ ] Configure object storage:
  - [ ] Use MinIO bucket from Phase 4.3
  - [ ] Retention policies (raw, 5m, 1h downsampling)
- [ ] Multi-cluster setup:
  - [ ] Prometheus + Sidecar per cluster
  - [ ] Central Query component
  - [ ] External labels for cluster identification
- [ ] Query patterns:
  - [ ] Cross-cluster queries
  - [ ] Historical data queries
  - [ ] Deduplication strategies
- [ ] Grafana integration:
  - [ ] Thanos Query as datasource
  - [ ] Multi-cluster dashboards
- [ ] Compare with alternatives:
  - [ ] Thanos vs Cortex vs Mimir
  - [ ] Storage costs and performance
- [ ] Document Thanos operational patterns
- [ ] **ADR:** Document long-term metrics strategy

---

### 4.7 Observability Cost Management

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

