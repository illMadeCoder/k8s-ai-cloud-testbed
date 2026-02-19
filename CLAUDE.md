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
site/                            Astro + Tailwind benchmark site (GitHub Pages, ADR-017)
site/data/                       Experiment result JSONs + _categories.json + _series.json (committed, not LFS)
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
| Secrets | OpenBao + ESO | External Secrets Operator syncs from OpenBao → K8s Secrets |
| Benchmark site | Astro + Tailwind + Vega-Lite | GitHub Pages at `illmadecoder.github.io/k8s-ai-cloud-testbed/` (ADR-017) |
| CI | GitHub Actions | Builds operator + component images, deploys site |

Operator image: `ghcr.io/illmadecoder/experiment-operator`

## Benchmark Site (`site/`)

Astro static site with Tailwind CSS, deployed to GitHub Pages via `deploy-site.yaml`.

### Information Architecture

```
/                              Landing (hero + stats + series cards + recent experiments)
/series/                       Series index (all research tracks)
/series/{id}/                  Series detail (ordered experiments in a track)
/categories/                   Domain index (all categories)
/categories/{domain}/          Domain detail (observability, networking, storage, cicd)
/comparisons/                  All comparison experiments index
/about/                        Portfolio + methodology + architecture + tech stack
/experiments/{slug}/           Experiment group (latest run + run history + series nav)
/experiments/{slug}/{run}/     Individual run detail
/tags/{tag}/                   Tag filter
```

### Key Files

```
site/astro.config.mjs          Astro config (Tailwind integration, GitHub Pages base)
site/tailwind.config.mjs       Tailwind theme (maps CSS custom properties)
site/data/_categories.json     Domain taxonomy (observability, networking, storage, cicd)
site/data/_series.json         Series definitions with optional experiment ordering
site/data/{name}.json          Experiment result JSONs (auto-committed by operator/analyzer)
site/src/types.ts              TypeScript interfaces mirroring Go structs
site/src/lib/experiments.ts    Data loading, grouping, domain/type derivation, series siblings
site/src/lib/series.ts         Series data loading and ordering
site/src/lib/categories.ts     Category data loading
site/src/lib/format.ts         Value formatting (duration, bytes, cost)
site/src/lib/vega-specs.ts     Vega-Lite chart builders
site/src/layouts/Base.astro    Root layout (nav, breadcrumb slot, footer)
site/src/components/           Hero, StatsBar, DomainCard, Breadcrumb, cards, charts
site/src/pages/                All routes matching IA above
```

### Domain Taxonomy

Experiments are categorized by tags into domains:

| Domain | Tags | Subdomains |
|--------|------|------------|
| `observability` | metrics, logging, tracing, prometheus, victoria-metrics, loki, tempo, grafana, slos, cost | metrics, logging, tracing, slos, cost |
| `networking` | gateways, ingress, service-mesh, gateway, envoy, nginx, traefik | gateways |
| `storage` | object-storage, database, s3, seaweedfs | object-storage |
| `cicd` | pipelines, ci, cd, supply-chain | pipelines |

Experiment types derived from tags: `comparison`, `tutorial`, `demo`, `baseline`.

### Publishing

Experiments must have `spec.publish: true` to be published to the site and receive AI analysis.
Without this field (or `publish: false`), results are stored in S3 only — no GitHub commit, no
analyzer job, no API cost. Only comparisons are published; tutorials and demos stay private.

### Design

- **Tailwind-first** with CSS custom properties for dark mode (`prefers-color-scheme: dark`)
- **Monospace typography** (JetBrains Mono) — engineering/terminal aesthetic
- **Empty-state friendly** — categories and pages show "Coming soon" with no data
- **Responsive** — mobile hamburger nav, stacking grids

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

#### Credential Persistence (PVC Strategy)

The analyzer authenticates via Claude Code OAuth credentials (access token + refresh token).
OAuth refresh tokens are **single-use and rotated** — each token refresh invalidates the
previous refresh token and issues a new one. This creates a token chain problem:

- If the local CLI and the analyzer both start from the same refresh token, whichever
  refreshes first invalidates the other's copy.
- Access tokens expire after ~24h, so the analyzer must be able to refresh independently.

**Solution: PVC-based independent token chain.** The analyzer stores credentials on a
PersistentVolumeClaim (`claude-credentials-pvc`, 1Mi, `local-path-retain` storage class).
An init container conditionally seeds from the K8s Secret only if the PVC is empty:

```
First run:  Secret → PVC (seed) → analyzer refreshes → new tokens written to PVC
Next runs:  PVC already has tokens → skip seed → analyzer refreshes from PVC tokens
```

After the first successful refresh, the analyzer maintains its own independent token chain
on the PVC, completely decoupled from the local CLI's token chain. Both share the same
subscription but hold different, independently-rotating credentials.

**Key files:**
- PVC manifest: `platform/manifests/experiment-operator-config/claude-credentials-pvc.yaml`
- Init container logic: `operators/experiment-operator/internal/controller/experiment_controller.go`
  (search for `claude-home` volume)
- Seed secret: `claude-auth` in `experiment-operator-system` (synced from OpenBao via ESO)

#### Initial Setup / Re-seeding

Seed credentials into OpenBao when setting up the analyzer for the first time, or when
re-seeding is needed (PVC deleted, tokens expired from prolonged inactivity >30 days):

```bash
# 1. Store current credentials in OpenBao (seeds the PVC on next analyzer run):
kubectl exec -n openbao openbao-0 -- sh -c \
  "BAO_TOKEN='<root_token>' bao kv put secret/experiment-operator/claude-auth \
  credentials='$(cat ~/.claude/.credentials.json)'"

# 2. Force ESO to sync immediately:
kubectl annotate externalsecret claude-auth -n experiment-operator-system \
  force-sync=$(date +%s) --overwrite

# 3. Clear the PVC so the init container re-seeds (only if re-seeding):
kubectl delete pvc claude-credentials-pvc -n experiment-operator-system
kubectl apply -f platform/manifests/experiment-operator-config/claude-credentials-pvc.yaml

# 4. Next analyzer Job run will seed from Secret → PVC, then refresh independently.
```

**Important:** After re-seeding, the local CLI's refresh token is shared with the analyzer
exactly once. The first analyzer run that successfully refreshes will fork the token chain.
After that point, do NOT re-seed again — the analyzer's PVC tokens are independent and
re-seeding would break the analyzer's chain (the local CLI will have already rotated past
the OpenBao copy).

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
- **Experiment YAML**: Use `generateName:` prefix, `namespace: experiments`, `publish: true` for site-bound experiments
- **Component refs**: `spec.targets[].components[].app` maps to `components/*/component.yaml` by `metadata.name`
- **Metrics query names**: Must match `^[a-z][a-z0-9_]*$`

## Secrets Management (OpenBao + ESO)

OpenBao runs in the `openbao` namespace (pod: `openbao-0`). External Secrets Operator (ESO) syncs
secrets from OpenBao to K8s Secrets via `ClusterSecretStore` named `openbao`.

**Root token:** `~/.illmlab/openbao-keys.json` → `root_token` field.

### Stored Secrets

| OpenBao Path | K8s Secret | Namespace | Purpose | ExternalSecret Manifest |
|-------------|------------|-----------|---------|------------------------|
| `secret/experiment-operator/claude-auth` | `claude-auth` | `experiment-operator-system` | Claude Code credentials (seed only — see PVC strategy in §6) | `platform/manifests/external-secrets-config/claude-auth-secret.yaml` |
| `secret/experiment-operator/github-api-token` | `github-api-token` | `experiment-operator-system` | GitHub PAT for site auto-publish + analyzer commits | `platform/manifests/external-secrets-config/github-api-token-secret.yaml` |

### Analyzer Credentials PVC

The `claude-credentials-pvc` PVC in `experiment-operator-system` persists the analyzer's
OAuth tokens across Job runs. See §6 for the full credential persistence strategy.

- Manifest: `platform/manifests/experiment-operator-config/claude-credentials-pvc.yaml`
- Storage class: `local-path-retain` (survives PVC recreation)
- Size: 1Mi (holds a single JSON credentials file)
- **Do NOT delete** unless re-seeding is required (breaks the analyzer's token chain)

### Managing Secrets

```bash
# Read a secret
kubectl exec -n openbao openbao-0 -- sh -c "BAO_TOKEN='<root_token>' bao kv get secret/experiment-operator/claude-auth"

# Write/update a secret (e.g. Claude Code credentials with refresh token)
kubectl exec -n openbao openbao-0 -- sh -c "BAO_TOKEN='<root_token>' bao kv put secret/experiment-operator/claude-auth credentials='$(cat ~/.claude/.credentials.json)'"

# Force ExternalSecret refresh (normally refreshes every 1h)
kubectl annotate externalsecret <name> -n <namespace> force-sync=$(date +%s) --overwrite

# Verify sync status
kubectl get externalsecret -A
```

### Adding a New Secret

1. Store in OpenBao: `bao kv put secret/<path> key=value`
2. Create ExternalSecret YAML in `platform/manifests/external-secrets-config/`
3. Reference `ClusterSecretStore: openbao`
4. Apply: `kubectl apply -f platform/manifests/external-secrets-config/<file>.yaml`

### SeaweedFS S3 Credentials

SeaweedFS S3 uses a separate config (not OpenBao): `accessKey: any`, `secretKey: any`.
Config stored in `seaweedfs-s3-config` secret in `seaweedfs` namespace.
Anonymous access: Read + List only. Writes require credentials.

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
