---
description: Scaffold a new experiment directory with validated experiment.yaml, workflow template, and CI checks
allowed-tools: Bash, Read, Write, Glob, Grep, Edit, AskUserQuestion
argument-hint: [experiment-name]
---

# Create Experiment

Scaffold a new experiment with a validated `experiment.yaml`, generate the WorkflowTemplate, and verify component images are built. Guide the user through an interactive flow, then generate all files and validate.

## Step 1: Parse Arguments

Extract the experiment name from `$ARGUMENTS`. If empty, ask the user with `AskUserQuestion`.

The experiment name becomes the directory under `experiments/`. It must be lowercase, hyphenated, and descriptive (e.g., `metrics-comparison`, `tempo-tutorial`).

Check that `experiments/{name}/` does not already exist. If it does, tell the user and stop.

## Step 2: Gather Experiment Details

Use `AskUserQuestion` to collect inputs. You may combine questions into a single call (up to 4 questions per call).

### Question Set 1: Type & Domain

Ask these two questions together:

**Experiment type** (determines structure):
- `comparison` — Compares two or more tools/approaches. Gets `publish: true`, metrics, hypothesis, analyzerConfig. Workflow completion mode: `workflow`.
- `tutorial` — Interactive learning. Gets `tutorial:` section with `exposeKubeconfig: true` and service refs. No publish, no metrics. Workflow completion mode: `manual`.
- `demo` — Simple demonstration. Minimal config, no publish. Workflow completion mode: `workflow`.
- `baseline` — Establishes baseline measurements. Gets metrics but simpler than comparison. Workflow completion mode: `workflow`.

**Domain** (determines tags):
- `observability` — metrics, logging, tracing, slos, cost
- `networking` — gateways, ingress, service-mesh
- `storage` — object-storage, database
- `cicd` — pipelines, ci, cd, supply-chain

### Question Set 2: Target Configuration

Ask about target setup:

**Number of targets:**
- `Single target (app only)` — Most tutorials and simple experiments
- `Two targets (app + loadgen)` — Comparisons with load testing (loadgen depends on app)
- `Custom` — Let user specify target names

**Machine type for main target:**
- `e2-standard-2` — Lightweight (loadgen, simple demos)
- `e2-standard-4` — Default (most experiments)
- `e2-standard-8` — Heavy compute (large observability stacks)

All targets default to `preemptible: true` and observability enabled with `transport: tailscale`.

### Question Set 3: Components

Present the component catalog grouped by category. Ask the user which components to include for each target. You can suggest relevant components based on the domain:

**Component Catalog (28 components):**

**apps/** (8):
- `hello-app` — Simple hello world app for load testing
- `naive-db` — Naive fsync-per-write database (i32 column, baseline for storage benchmarks)
- `nginx` — NGINX web server
- `otel-demo` — Multi-service OTel demo (user-service, order-service, payment-service)
- `cardinality-generator` — High-cardinality Prometheus metrics for cost analysis
- `station-monitor` — Station monitoring app for Prometheus tutorial
- `log-generator` — Structured logs for logging pipeline tutorials
- `metrics-app` — Demo app exposing Prometheus metrics

**core/** (4):
- `nginx-ingress` — NGINX Ingress Controller
- `envoy-gateway` — Envoy Gateway (Gateway API)
- `traefik` — Traefik proxy with Gateway API support
- `tailscale-operator` — Tailscale K8s operator for mesh networking

**observability/** (15):
- `kube-prometheus-stack` — Prometheus + Grafana + Alertmanager
- `victoria-metrics` — VictoriaMetrics single-node TSDB
- `mimir` — Grafana Mimir TSDB (monolithic mode)
- `loki` — Loki log aggregation
- `promtail` — Promtail log collector (ships to Loki)
- `fluent-bit` — Fluent Bit log processor/forwarder
- `elasticsearch` — Elasticsearch for log storage
- `kibana` — Kibana dashboard for Elasticsearch
- `tempo` — Grafana Tempo distributed tracing
- `jaeger` — Jaeger distributed tracing (all-in-one)
- `otel-collector` — OpenTelemetry Collector
- `pyrra` — Pyrra SLO management
- `seaweedfs` — SeaweedFS S3-compatible object storage
- `metrics-agent` — Grafana Alloy agent (scrapes + remote-writes to hub VictoriaMetrics)
- `metrics-egress` — Tailscale egress service to hub VictoriaMetrics

**storage/** (1):
- `minio` — MinIO S3-compatible object storage

**testing/** (1):
- `k6-gateway-loadtest` — k6 load test for gateway comparison (runs on loadgen cluster)

Suggest relevant components based on domain and type. For example:
- Observability comparison → suggest the stacks being compared + `kube-prometheus-stack` + `metrics-agent` + `metrics-egress`
- Networking comparison → suggest gateway controllers + `hello-app` + `k6-gateway-loadtest`
- Tutorial → suggest the learning target + `kube-prometheus-stack` for dashboards

### Question Set 4: Publish & Analysis (comparison/baseline only)

If the experiment type is `comparison` or `baseline`, ask:

**Publish to benchmark site?**
- `Yes` — Sets `publish: true`, includes `analyzerConfig`
- `No` — Results stored in S3 only

If publishing, the standard analyzer sections are included automatically. Ask if they want to customize (most users won't):
- Core: `abstract`, `targetAnalysis`, `performanceAnalysis`, `metricInsights`
- FinOps: `finopsAnalysis` (include by default), `secopsAnalysis` (omit by default unless security-relevant)
- Deep dive: `body`, `capabilitiesMatrix` (comparisons only), `feedback`
- Diagram: `architectureDiagram`

## Step 3: Generate the Experiment YAML

### GKE Name Length Validation

Before writing, validate GKE cluster name lengths:

```
GKE name = "illm-" (5) + experimentName + "-" (1) + targetName + "-" (1) + xrSuffix (5) = 12 + len(experimentName) + len(targetName)
experimentName = generateNamePrefix + k8sSuffix (5 chars)
Total = 17 + len(generateNamePrefix) + len(targetName), must be <= 40
```

The `generateName` prefix should be an abbreviation of the experiment name followed by a hyphen:
- `gateway-comparison` → `gw-comp-`
- `logging-comparison` → `logging-comparison-` (if it fits)
- `prometheus-tutorial` → `prometheus-tutorial-`
- Long names → abbreviate: `observability-cost-tutorial` → `obs-cost-tut-`

**For each target, check:** `17 + len(prefix) + len(targetName) <= 40`

If any target exceeds 40 chars, shorten the `generateName` prefix and warn the user.

### Metrics Pipeline Requirements

For published experiments (`publish: true`) with custom metrics:
- **Always include these components** on targets that have application workloads:
  - `kube-prometheus-stack` — Prometheus scrapes ServiceMonitors and serves the Prometheus API
  - `metrics-agent` — Grafana Alloy scrapes cadvisor + annotated pods and remote-writes to hub VictoriaMetrics
  - `metrics-egress` — Tailscale egress service enabling remote-write to hub VM
- **How metrics collection works:**
  1. Prometheus scrapes application ServiceMonitors on the target cluster
  2. The operator queries the target's local Prometheus via K8s API proxy for custom PromQL queries
  3. Hub VictoriaMetrics is a fallback path (receives cadvisor + annotated pod metrics via Alloy remote-write)
- **Variable reference for PromQL queries:**
  - `$EXPERIMENT` — resolves to the experiment instance name (e.g., `db-baseline-fsync-b8twf`). Use this for namespace filters on target clusters where pods deploy to the experiment-named namespace.
  - `$NAMESPACE` — also resolves to the experiment instance name on target clusters. Prefer `$EXPERIMENT` for clarity.
  - `$DURATION` — resolves to the experiment duration as a Prometheus duration string (e.g., `1h30m`). Use in `rate()` or `histogram_quantile()` for full-window aggregation.
- **Pod annotations for Alloy discovery:** Application pods should have `prometheus.io/scrape: "true"`, `prometheus.io/port`, and `prometheus.io/path` annotations in their pod template for Alloy to discover and forward metrics to hub VM.

### YAML Structure

Generate `experiments/{name}/experiment.yaml` following these patterns:

**All experiments:**
```yaml
apiVersion: experiments.illm.io/v1alpha1
kind: Experiment
metadata:
  generateName: {prefix}-
  namespace: experiments
spec:
  description: "{description}"
  tags: ["{type}", "{domain}", ...additional-tags]
  # publish: true  — only for comparisons/baselines being published
  targets:
    - name: app
      cluster:
        type: gke
        machineType: {machineType}
        preemptible: true
      observability:
        enabled: true
        transport: tailscale
      components:
        - app: {component1}
        - app: {component2}
          params:
            key: "value"
  workflow:
    template: {name}-validation
    completion:
      mode: workflow  # or manual for tutorials
```

**Comparisons add** (when published):
```yaml
  publish: true

  analyzerConfig:
    sections:
      - abstract
      - targetAnalysis
      - performanceAnalysis
      - metricInsights
      - finopsAnalysis
      # - secopsAnalysis      # Uncomment for security-relevant experiments
      - body
      - capabilitiesMatrix    # Only for comparisons
      - feedback
      - architectureDiagram

  hypothesis:
    claim: "{user-provided or placeholder}"
    questions:
      - "{question1}"
    focus:
      - "{focus1}"

  metrics:
    # Generate PromQL queries for the specific app being measured.
    # Metric names must match ^[a-z][a-z0-9_]*$
    # Types: instant (bar charts) or range (time-series line charts)
    # Variables: $EXPERIMENT (preferred for namespace), $DURATION
    #
    # For naive-db: generate these 10 queries:
    #   write_p99_latency, write_p50_latency, fsync_p99_latency, fsync_p50_latency,
    #   read_p99_latency, write_throughput, read_throughput, total_rows,
    #   write_latency_over_time, fsync_latency_over_time
    #   Use: naivedb_write_duration_seconds_bucket, naivedb_fsync_duration_seconds_bucket,
    #        naivedb_read_duration_seconds_bucket, naivedb_operations_total, naivedb_rows_total
    #   Filter: namespace=~"$EXPERIMENT"
    #
    # For hello-app: generate HTTP latency queries using standard Go HTTP metrics
    #
    # For unknown apps: at minimum include workload CPU/memory queries:
    #   cpu_by_pod, memory_by_pod (using container_cpu_usage_seconds_total,
    #   container_memory_working_set_bytes with namespace=~"$EXPERIMENT")
    #   Add TODO for custom application-specific queries
    []
```

**Tutorials add:**
```yaml
  tutorial:
    exposeKubeconfig: true
    services:
      - name: grafana
        target: app
        service: kube-prometheus-stack-grafana
        namespace: {experiment-name}
```

**Two-target (app + loadgen) adds:**
```yaml
    - name: loadgen
      depends: [app]
      cluster:
        type: gke
        machineType: e2-standard-2
        preemptible: true
      components:
        - app: k6-gateway-loadtest
```

### Tag Conventions

Tags are used for site categorization. Always include:
1. The experiment type: `comparison`, `tutorial`, `demo`, or `baseline`
2. The domain: `observability`, `networking`, `storage`, `cicd`
3. Specific technology tags (lowercase): `prometheus`, `loki`, `gateway`, `envoy`, `tracing`, etc.

### Load Generation Validation

For `baseline` and `comparison` experiments:
- **Load generation is required** — baselines and comparisons need traffic to produce meaningful metrics
- Check if the app component has an embedded load generator (`k8s/loadgen.yaml` in the component directory):
  - If yes: the workflow just needs a warm-up + observation window (load runs alongside the app)
  - If no: the workflow MUST include an explicit load-test step (curl loops, k6, etc.)
- For `tutorial` and `demo` types, load generation is optional

### Hypothesis Guidance (comparisons)

Ask the user for a 1-2 sentence claim about expected outcome. If they don't have one, generate a reasonable placeholder based on the components being compared. Include 2-3 questions the experiment should answer and 2-3 focus areas.

### Metrics Guidance (comparisons/baselines)

For published experiments, generate actual PromQL queries based on the app components (see the metrics section in YAML Structure above). If the app has known instrumentation (naive-db, hello-app), generate the full query set. For unknown apps, generate workload CPU/memory queries and add TODOs for custom metrics.

**Important:** Use `$EXPERIMENT` (not `$NAMESPACE`) for namespace filters in queries. On target clusters, pods deploy to the experiment-named namespace, and `$EXPERIMENT` resolves correctly to this name.

## Step 4: Write the File

Use the `Write` tool to create `experiments/{name}/experiment.yaml`.

## Step 5: Validate

After writing, run the `experiment-validator` agent on the new experiment to catch any issues:

```
Run experiment-validator agent with the experiment name
```

Report the validation results. If there are failures, fix them and re-validate.

### Component Pipeline Validation

After the experiment-validator passes, run these additional checks:

1. **Metrics pipeline completeness:** If `publish: true` and custom `metrics:` are defined, verify these components are present on the target:
   - `kube-prometheus-stack` (required for ServiceMonitor scraping + local Prometheus API)
   - `metrics-agent` (required for cadvisor + annotated pod metrics forwarding)
   - `metrics-egress` (required for Tailscale egress to hub VM)
   - If any are missing, warn: "Published experiment with custom metrics requires kube-prometheus-stack, metrics-agent, and metrics-egress components"

2. **Load generation:** If type is `baseline` or `comparison`, check that load generation exists:
   - Check `experiments/{name}/workflow/template.yaml` for load-test steps, OR
   - Check if the app component has `k8s/loadgen.yaml`
   - If neither exists, warn: "Baseline/comparison experiments need load generation for meaningful metrics"

3. **Query variable check:** If any queries use `$NAMESPACE`, warn and suggest replacing with `$EXPERIMENT`:
   - `grep '$NAMESPACE' experiments/{name}/experiment.yaml`
   - "$NAMESPACE resolves to the Experiment CR namespace ('experiments'), not the target namespace. Use $EXPERIMENT instead."

## Step 6: Generate WorkflowTemplate

Create `experiments/{name}/workflow/template.yaml`. This is the Argo WorkflowTemplate the operator submits when the experiment reaches the Running phase.

### Template anatomy

All templates follow this structure:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: WorkflowTemplate
metadata:
  name: {workflow-template-name}   # Must match experiment's spec.workflow.template
  namespace: argo-workflows
  labels:
    app.kubernetes.io/managed-by: experiment-operator
spec:
  entrypoint: {entrypoint-name}
  serviceAccountName: argo-workflow
  arguments:
    parameters:
      - name: experiment-name
        value: ""
      - name: target-endpoint
        value: ""
      - name: target-name
        value: ""
  templates:
    # ... (type-specific templates below)
```

### Type-specific workflow patterns

**All types** share a smoke-test step:

```yaml
- name: check-target
  container:
    image: curlimages/curl:8.5.0
    command: [sh, -c]
    args:
      - |
        echo "=== Smoke-testing target cluster ==="
        for i in $(seq 1 30); do
          if curl -sk "https://{{workflow.parameters.target-endpoint}}/livez" 2>/dev/null; then
            echo "Target cluster reachable"
            exit 0
          fi
          echo "Attempt $i/30 — retrying in 10s..."
          sleep 10
        done
        echo "Target cluster not reachable after 5 minutes"
        exit 1
```

**baseline** — smoke test → wait for app → load test + observation

Generate a load-test step that exercises the app's API. Tailor the script to the specific app components:
- If the experiment uses `naive-db`: sequential writes, then random reads against the naive-db HTTP API
- If the experiment uses `hello-app`: simple HTTP GET load with curl loops
- Otherwise: a generic sleep-based observation window

The load-test step should use `curlimages/curl:8.5.0` or `busybox:1.36` (no special tooling). Write the load test inline as a shell script in the `args` field. Structure it in phases:
1. **Warm-up** (2-5 min): let the stack stabilize after deployment
2. **Load test** (10-20 min): drive traffic at a steady rate, print progress every N iterations
3. **Cooldown** (2-5 min): let metrics settle before collection

Total duration: 15-30 minutes for baselines (shorter than comparisons).

**comparison** — smoke test → wait for stacks → phased observation

```yaml
- name: validate-and-observe
  steps:
    - - name: smoke-test
        template: check-target
    - - name: wait-for-stacks
        template: wait-for-stacks
    - - name: observe
        template: phased-observation
```

Phases (typically 45-90 min total):
1. Wait for stacks (5 min): let ArgoCD sync all components
2. Idle baseline (10 min): observe resource usage with no load
3. Load test (20-45 min): main measurement window
4. Cooldown (10 min): post-load observation

For multi-target experiments (app + loadgen), add parameters for each target endpoint:
```yaml
arguments:
  parameters:
    - name: app-endpoint
      value: ""
    - name: loadgen-endpoint
      value: ""
```

**tutorial** — smoke test → suspend (manual completion)

```yaml
- name: validate-and-suspend
  steps:
    - - name: smoke-test
        template: check-target
    - - name: wait-for-user
        template: suspend-step

- name: suspend-step
  suspend: {}
```

**demo** — smoke test → short observation (5 min)

```yaml
- name: validate-and-observe
  steps:
    - - name: smoke-test
        template: check-target
    - - name: observe
        template: observe

- name: observe
  container:
    image: busybox:1.36
    command: [sh, -c]
    args: ["echo 'Observing for 5 minutes...'; sleep 300; echo 'Done'"]
```

### Common images

Only use these lightweight images in workflow templates:
- `curlimages/curl:8.5.0` — HTTP checks, curl-based load tests
- `busybox:1.36` — Sleep/wait phases, shell scripting

### Write the file

Use the `Write` tool to create `experiments/{name}/workflow/template.yaml`.

## Step 7: Apply Component CRs

The operator resolves component sources from Component CRs in the hub cluster. If a Component CR
doesn't exist, the operator falls back to a path without the `/k8s` suffix, which causes ArgoCD
to scan the entire component directory (including `component.yaml`, `Dockerfile`, `src/`) and fail.

For each component with a local `component.yaml` (under `components/`):

1. **Check if CR exists** in the cluster:
   ```bash
   kubectl get component {component-name} -o name 2>&1
   ```

2. **If not found**, apply it:
   ```bash
   kubectl apply -f components/{category}/{component-name}/component.yaml
   ```

3. **If found**, verify the source path matches the local file (in case it was updated):
   ```bash
   kubectl get component {component-name} -o jsonpath='{.spec.sources[0].path}'
   ```

Report status per component:
- `[OK] {component} — Component CR exists, path correct`
- `[APPLIED] {component} — Component CR applied to cluster`
- `[SKIPPED] {component} — Helm chart, no local component.yaml`

## Step 8: Check Component Images

For each component referenced in the experiment, determine whether it needs a custom-built image:

1. **Check for Dockerfile**: Use `Glob` to check if `components/*/{component-name}/Dockerfile` exists.
   - If no Dockerfile: it's a Helm chart component (e.g., `kube-prometheus-stack`). No image build needed.
   - If Dockerfile exists: it's a custom image that CI must build.

2. **For custom-image components**, check the latest CI build:
   ```bash
   gh run list -w "Build Components" -L 1 --json status,conclusion,headBranch,createdAt
   ```

   Report status per component:
   - CI passed recently: `[OK] {component} — image built`
   - CI running/queued: `[PENDING] {component} — CI in progress, wait before running experiment`
   - CI failed or no recent build: `[ACTION NEEDED] {component} — push code and wait for CI`

   If any component image isn't built yet, **warn the user** that the experiment will fail at the Ready phase because ArgoCD won't be able to pull the image. Suggest:
   ```
   Watch CI: gh run watch {run-id}
   ```

## Step 9: Summary

Tell the user:
1. **Files created:**
   - `experiments/{name}/experiment.yaml`
   - `experiments/{name}/workflow/template.yaml`
2. **GKE name length**: `generateName` prefix and validation result
3. **Component images**: status of custom-image builds (OK / pending / action needed)
4. **TODOs** (if any): metrics queries, hypothesis details
5. **Next steps**:
   - If images built: `kubectl apply -f experiments/{name}/workflow/template.yaml && kubectl create -f experiments/{name}/experiment.yaml`
   - If images pending: wait for CI, then apply
   - Or use `/run-experiment {name}` which handles pre-flight checks, applying the template, and monitoring
