# CLAUDE.md

Benchmarking-focused K8s experiment lab. Current phase: 3-4 (observability + traffic management).
Docs: `docs/roadmap.md` | `docs/adrs/` (17 ADRs) | `docs/gitops-patterns.md`
Use `/compact` at ~70% context.

## Architecture

### Experiment Operator

Kubebuilder operator that orchestrates full experiment lifecycle:

```
Pending → Provisioning → Ready → Running → Complete/Failed
```

Provisions target clusters (Crossplane), deploys components (ArgoCD), runs workflows (Argo Workflows), collects metrics, stores results (SeaweedFS S3).

### CRDs

| Kind | Group | Scope | Purpose |
|------|-------|-------|---------|
| `Experiment` | `experiments.illm.io/v1alpha1` | Namespaced | Orchestrates multi-cluster deployments with targets, components, workflows |
| `Component` | `experiments.illm.io/v1alpha1` | Cluster | Reusable component definitions with sources, params, observability config |

### Operator Packages (`operators/experiment-operator/internal/`)

| Package | Purpose |
|---------|---------|
| `controller` | Reconciliation loop for Experiment CRD |
| `argocd` | ArgoCD client, application lifecycle, cluster registration |
| `components` | Resolves ComponentRef → actual Git/Helm sources |
| `crossplane` | Creates/manages GKECluster claims via Crossplane |
| `github` | GitHub Contents API client, auto-commits results to benchmark site |
| `metrics` | Collects experiment metrics for results |
| `storage` | SeaweedFS S3 client for experiment results |
| `workflow` | Creates/monitors Argo Workflows for validation/lifecycle |

### Directory Map

```
operators/experiment-operator/   Kubebuilder operator (Go, CI-built)
components/{apps,core,obs,...}/  42 components with component.yaml (8 categories)
experiments/{name}/              17 experiment scenarios (+ _template)
platform/{apps,manifests,values} Hub cluster config + ArgoCD apps
site/                            Astro benchmark results site (GitHub Pages, ADR-017)
site/data/                       Experiment result JSONs (committed, not LFS)
docs/{adrs,roadmap}              17 ADRs, phase docs
.github/workflows/               build-operator, build-components, deploy-site, auto-merge
```

## Infrastructure Stack

| Layer | Tool | Notes |
|-------|------|-------|
| Hub cluster | Talos Linux | `192.168.1.178` (node: `talos-23n-3ay`) |
| GitOps | ArgoCD | Multi-source apps, sync waves |
| Orchestration | Argo Workflows | Experiment lifecycle workflows |
| Cloud provisioning | Crossplane v2.1.3 | GKE cluster claims |
| Metrics | Prometheus + VictoriaMetrics | VictoriaMetrics Single for backhaul (ADR-016) |
| Logs | Loki | |
| Traces | Tempo | |
| Object storage | SeaweedFS | S3-compatible, experiment results |
| Policy | Kyverno + Cosign | Supply chain security |
| Secrets | OpenBao | |
| Benchmark site | Astro + Vega-Lite | GitHub Pages at `illmadecoder.github.io/k8s-ai-cloud-testbed/` (ADR-017) |
| CI | GitHub Actions | Builds operator + component images, deploys site |

Operator image: `ghcr.io/illmadecoder/experiment-operator`

## Environment Constraints

**WSL ~8GB RAM. No Docker, Kind, or Go builds locally.**

- No `docker build/run/compose` — kills WSL responsiveness
- No `go build` — CI does this
- Terminal only: `kubectl`, `talosctl`, `gh`, `git`, code editing
- Talos cluster on LAN at `192.168.1.178` (node: `talos-23n-3ay`)
- Operator images built by GitHub Actions CI → `ghcr.io/illmadecoder/experiment-operator`

## Common Workflows

### 1. Deploy operator changes

```bash
# Commit, push — CI builds & pushes image
git add <files> && git commit -m "feat: ..." && git push

# Watch CI
gh run list -w "Build Experiment Operator" -L 3

# Once CI passes, restart to pull new :latest
kubectl rollout restart deployment/experiment-operator-controller-manager -n experiment-operator-system
kubectl rollout status deployment/experiment-operator-controller-manager -n experiment-operator-system
```

### 2. Apply CRD / infra updates

```bash
# CRD updates (after make manifests)
kubectl apply -f operators/experiment-operator/config/crd/bases/

# SeaweedFS bucket creation (re-run job)
kubectl delete job seaweedfs-create-buckets -n seaweedfs --ignore-not-found
kubectl apply -f platform/manifests/seaweedfs-config/buckets.yaml

# S3 credentials for Argo Workflows
kubectl apply -f platform/manifests/seaweedfs-config/s3-credentials.yaml
```

### 3. Run an experiment

```bash
# Create (generateName gives unique name each time)
kubectl create -f experiments/hello-app/experiment.yaml

# Watch lifecycle
kubectl get experiments -n experiments -w

# Check results after completion
kubectl get experiments -n experiments -o wide   # Shows ResultsURL column
```

### 4. Check S3 results

```bash
kubectl run -n seaweedfs s3check --rm -it --restart=Never \
  --image=curlimages/curl:8.5.0 -- \
  curl -s http://seaweedfs-s3.seaweedfs.svc.cluster.local:8333/experiment-results/<name>/summary.json
```

### 5. Publish experiment results to benchmark site

**Automated:** The operator auto-commits `summary.json` to `site/data/{name}.json` via the GitHub
Contents API on experiment completion. This triggers the `deploy-site.yaml` workflow to rebuild
the benchmark site on GitHub Pages. Requires `GITHUB_TOKEN` secret (see below).

```bash
# Setup: Create GitHub token secret (fine-grained PAT with Contents: Read and write)
kubectl create secret generic github-api-token \
  -n experiment-operator-system \
  --from-literal=token=github_pat_xxx

# Restart operator to pick up new env
kubectl rollout restart deployment/experiment-operator-controller-manager \
  -n experiment-operator-system
```

**Manual fallback** (if token not configured or commit failed):

```bash
EXP_NAME=$(kubectl get experiments -n experiments -o jsonpath='{.items[-1].metadata.name}')
kubectl run -n seaweedfs s3fetch --rm -it --restart=Never \
  --image=curlimages/curl:8.5.0 -- \
  curl -s http://seaweedfs-s3.seaweedfs.svc.cluster.local:8333/experiment-results/${EXP_NAME}/summary.json \
  > site/data/${EXP_NAME}.json
git add site/data/${EXP_NAME}.json && git commit -m "data: Add ${EXP_NAME} results" && git push
```

The file must conform to the `ExperimentSummary` JSON shape (see `operators/experiment-operator/internal/metrics/collector.go`).
Site types mirror Go structs in `site/src/types.ts`.

### 6. AI experiment analysis

On experiment completion, the operator creates an analyzer Job (`experiment-analyzer-{name}`)
that uses Claude Code CLI to generate AI analysis (summary, per-metric insights, recommendations).
The analysis is merged into `summary.json` and committed to the benchmark site.

**Setup: Store Claude Code token in OpenBao**

```bash
# Generate a long-lived token (uses Max subscription, ~1 year lifespan)
claude setup-token

# Store in OpenBao
bao kv put secret/experiment-operator/claude-auth token="<token from setup-token>"

# ESO syncs it to K8s Secret automatically via:
#   platform/manifests/external-secrets-config/claude-auth-secret.yaml

# Restart operator to pick up new env (if needed)
kubectl rollout restart deployment/experiment-operator-controller-manager \
  -n experiment-operator-system
```

Analyzer image: `ghcr.io/illmadecoder/experiment-analyzer` (built by CI on changes to `operators/experiment-analyzer/`).

### 7. SeaweedFS bucket / credential updates

```bash
# Re-create bucket job
kubectl delete job seaweedfs-create-buckets -n seaweedfs --ignore-not-found
kubectl apply -f platform/manifests/seaweedfs-config/buckets.yaml

# Update S3 credentials
kubectl apply -f platform/manifests/seaweedfs-config/s3-credentials.yaml

# ArgoCD auto-syncs Argo Workflows config from git
```

## Conventions

- **ArgoCD apps**: Labels `experiment: {name}`, `cluster: target|loadgen`
- **ArgoCD patterns**: Multi-source, sync waves, `ignoreDifferences` for CRDs (see `docs/gitops-patterns.md`)
- **Experiment YAML**: Use `generateName:` prefix, `namespace: experiments`
- **Component refs**: `spec.targets[].components[].app` maps to `components/*/component.yaml` by `metadata.name`
- **Metrics query names**: Must match `^[a-z][a-z0-9_]*$`

## Beads / Toil Tracking

Track recurring issues with `bd`. Labels: `toil`, `flaky`, `config`, `timing`, `resources`, `networking`.
Lab labels: `loki-tutorial`, `slo-tutorial`, `logging-comparison`, etc.

| Priority | Meaning |
|----------|---------|
| P0 | Blocks lab completely |
| P1 | Workaround exists but painful |
| P2 | Annoying but manageable |
| P3 | Minor friction |

```bash
bd ready                           # Find available work
bd create "Title" -l toil -p 2     # Log issue
bd close <id>                      # Mark complete
bd sync                            # Sync with git
```

## Agents

| Agent | Purpose | When to use |
|-------|---------|-------------|
| `experiment-validator` | Validates experiment YAML structure + component cross-refs | Before applying any experiment, after editing experiment YAML |
| `cluster-health` | Hub cluster health sweep (ArgoCD, pods, Crossplane, operator) | Session start, after deployments, debugging failures |
