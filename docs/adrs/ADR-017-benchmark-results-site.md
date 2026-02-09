# ADR-017: Benchmark Results Site

## Status

Accepted (2026-02-09) — supersedes original Proposed (2026-02-08)

**Amendment:** Changed from separate `illm-benchmarks` repo to monorepo (`site/` directory). Removed GitHub API publisher from operator; data is committed directly to `site/data/`. Operator auto-publish deferred to a future phase.

## Context

The lab runs 15+ experiment scenarios (TSDB comparisons, logging comparisons, gateway benchmarks, etc.) that produce `ExperimentSummary` JSON stored in SeaweedFS S3. The operator's `collectAndStoreResults()` uploads `summary.json` and `metrics-snapshot.json` to `s3://experiment-results/<name>/` on completion (see ADR-015).

Currently the only way to view results is:

```bash
kubectl run -n seaweedfs s3check --rm -it --restart=Never \
  --image=curlimages/curl:8.5.0 -- \
  curl -s http://seaweedfs-s3.seaweedfs.svc.cluster.local:8333/experiment-results/<name>/summary.json
```

There is no public-facing way to display, analyze, or cross-reference benchmark results. The data exists but is locked behind `kubectl` and a Tailscale-only network.

### Requirements

| Requirement | Priority | Notes |
|-------------|----------|-------|
| Public access (no auth) | Must | Portfolio piece, shareable links |
| General-purpose metrics | Must | CPU, memory, latency, throughput, cost — not tied to specific experiment types |
| Cross-experiment comparison | High | Overlay metrics from different runs |
| Interactive charts | High | Filtering, selection, tooltips |
| Zero hosting cost | High | Lab budget is $0 for static hosting |
| Low operational complexity | Medium | No new runtime infrastructure on hub |
| Auto-publish on experiment completion | Future | Operator-driven, no manual steps — deferred until experiment cadence justifies it |

## Options Considered

### Axis 1: Hosting / Site Framework

| Option | How it works | Cost | Complexity |
|--------|-------------|------|------------|
| **GitHub Pages + Astro** | Static site in repo, GitHub Actions builds, Pages serves. Astro's island architecture ships minimal JS. | Free | Medium |
| **GitHub Pages + github-action-benchmark** | Purpose-built GH Action that ingests JSON benchmarks and generates Chart.js pages. | Free | Low |
| **Bencher.dev (SaaS)** | Continuous benchmarking platform. Free for public projects. API-driven data ingestion, hosted dashboard. | Free (public) | Low |
| **Grafana Cloud public dashboards** | Push metrics to Grafana Cloud free tier, share dashboards as public snapshots. | Free tier (10k series) | Low |
| **Self-hosted on hub (Tailscale + Cloudflare Tunnel)** | Host site on hub cluster, expose via Cloudflare Tunnel for public access. | Free | Medium-High |

### Axis 2: Visualization

| Approach | Interactivity | Cross-experiment | Custom metrics |
|----------|--------------|-----------------|----------------|
| **Vega-Lite** (declarative JSON specs) | Filtering, selection, tooltips | Yes (concat datasets) | Yes (any JSON shape) |
| **Observable Plot** (JS library) | Good interactivity | Yes (JS data joins) | Yes |
| **Chart.js** (via github-action-benchmark) | Basic (zoom, hover) | Limited (single page) | Rigid format |
| **Grafana dashboards** | Excellent (native) | Via variables/templating | Must be time-series |
| **D3.js** (low-level) | Unlimited | Yes | Yes |

### Axis 3: Repo Structure

| Approach | Pros | Cons |
|----------|------|------|
| **Monorepo (`site/` directory)** | One PR shows everything. No cross-repo coordination. GitHub Pages deploys from same repo. No PAT/API needed for data. | Site dependencies (node_modules) in same repo. Slightly larger clone. |
| **Separate `illm-benchmarks` repo** | Clean separation. Independent deploy cycle. | Cross-repo data push requires GitHub API + PAT. Two repos to manage. PRs split across repos. |

### Axis 4: Data Pipeline (operator → site)

| Approach | How data flows | Latency | Complexity |
|----------|---------------|---------|------------|
| **Commit JSON to `site/data/`** | Copy summary from S3 → commit to repo → push triggers build | Manual (for now) | Minimal |
| **Extend experiment-operator** | On completion: push summary JSON via GitHub API | ~2-5 min | Medium (PAT, API client) |
| **Argo Workflow post-step** | Final workflow step commits results to repo | ~1 min | Low (WorkflowTemplate change) |
| **Periodic GitHub Action** | Cron action pulls from S3 (via Tailscale) and rebuilds site | ~15 min | Low but requires S3 access from GH |

## Decision

**GitHub Pages + Astro + Vega-Lite**, monorepo in `site/` directory. Experiment result JSONs committed directly to `site/data/`. Operator auto-publish deferred.

### Why Monorepo Over Separate Repo

| Factor | Reasoning |
|--------|-----------|
| **Single PR** | Experiment YAML, operator changes, and site updates are reviewable together |
| **No cross-repo plumbing** | No GitHub API publisher in operator, no PAT to manage or rotate |
| **Simpler data flow** | Commit JSON to `site/data/`, push, GitHub Actions builds and deploys |
| **GitHub Pages supports it** | Deploy from a directory in the same repo — no second repo needed |
| **Split later if needed** | If experiment cadence grows to dozens per day, extract to a separate repo then. Not now. |

### Why Not Git LFS for `site/data/`

Experiment summary JSONs are small text files (1-50 KB each). Even at 200 runs, total data is under 10 MB. Git's built-in delta compression handles similar JSON structures efficiently. LFS would add bandwidth quotas (GitHub free tier: 1 GB/month), extra tooling for cloners, and complexity — all for a problem that doesn't exist at this scale.

### Why This Combination

| Factor | Reasoning |
|--------|-----------|
| **Free public hosting** | GitHub Pages — zero cost, global CDN |
| **General-purpose** | Vega-Lite specs handle any JSON shape — not locked to specific metric types |
| **Interactive** | Vega-Lite provides filtering, selection, tooltips, cross-experiment comparison |
| **Static** | No runtime infrastructure (no DB, no server process on hub) |
| **Portfolio value** | Public site doubles as a portfolio piece showing real benchmark methodology |

### Why Not the Others

| Option | Reason to reject |
|--------|-----------------|
| **github-action-benchmark** | Too rigid — expects a specific benchmark format (`name`, `unit`, `value`). Weak cross-experiment comparison. Cannot display arbitrary `ExperimentSummary` fields like `costEstimate` or `mimirMetrics`. |
| **Bencher.dev** | Excellent for CI regression tracking but less flexible for arbitrary system design metrics. Locked to their dashboard UX. |
| **Grafana Cloud** | Must structure everything as time-series. Public dashboard snapshots are ephemeral (auto-expire). Free tier limited to 10k series. |
| **Self-hosted** | Unnecessary complexity when GitHub Pages is free and always-on. Adds a process to monitor on the hub. |
| **D3.js** | Maximum flexibility but requires writing chart logic from scratch. Vega-Lite gets 90% of the value declaratively. |

### Pipeline Architecture

```
Phase 1 (now): Manual data commit
──────────────────────────────────
Experiment completes
        │
        ▼
S3: summary.json, metrics-snapshot.json
        │
        ▼ (manual: copy from S3, commit to repo)
site/data/<experiment>.json
        │
        ▼
GitHub Actions (on push to site/**)
        │
        ▼
Astro build → GitHub Pages (public, CDN-served)


Phase 2 (future): Operator auto-publish
────────────────────────────────────────
Experiment CR (phase: Complete)
        │
        ▼
experiment-operator collectAndStoreResults()
        │
        ├──► SeaweedFS S3 (existing)
        │
        └──► git commit + push to site/data/ (new)
                │
                ▼
        GitHub Actions → Astro build → GitHub Pages
```

### Site Structure

```
site/                                  # Monorepo: same repo as operator + experiments
├── src/
│   ├── pages/
│   │   ├── index.astro                # Landing: experiment list with status cards
│   │   └── experiments/
│   │       └── [slug].astro           # Per-experiment detail page
│   ├── components/
│   │   ├── VegaChart.astro            # Reusable Vega-Lite chart island
│   │   ├── ComparisonTable.astro      # Side-by-side metric comparison
│   │   └── MetricCard.astro           # Summary metric display
│   └── layouts/
│       └── Base.astro
├── data/                              # Experiment result JSONs (committed)
│   ├── hello-app-x7k2q.json
│   ├── tsdb-comparison-a1b2.json
│   └── ...
├── public/
├── astro.config.mjs
├── package.json
└── tsconfig.json
```

### Vega-Lite Example

A single experiment detail page renders specs like:

```json
{
  "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
  "data": {"url": "/data/tsdb-comparison-a1b2.json"},
  "mark": "bar",
  "encoding": {
    "x": {"field": "targets[].name", "type": "nominal"},
    "y": {"field": "targets[].metrics.memoryPeakMB", "type": "quantitative"},
    "color": {"field": "targets[].components[0]", "type": "nominal"}
  }
}
```

The spec is generated at build time from the JSON shape — any metric field in `ExperimentSummary` becomes a chartable dimension without code changes.

## Implementation Phases

### Phase 1: Static Site Scaffold

- Create `site/` directory with Astro + Vega-Lite
- GitHub Actions workflow: on push to `site/**` → `astro build` → deploy to Pages
- Landing page: reads `data/*.json`, renders experiment cards (name, phase, duration, date)
- Detail page: per-experiment Vega-Lite charts for all numeric fields
- Seed `site/data/` with sample experiment JSONs from S3

### Phase 2: Cross-Experiment Comparison

- Comparison view: select 2+ experiments, overlay metrics
- Filter by experiment type (TSDB, logging, gateway)
- Tag-based grouping from `ExperimentSummary` metadata

### Phase 3: Operator Auto-Publish (future)

- Extend `collectAndStoreResults()` to commit JSON to `site/data/` and push
- Requires Git credentials as K8s Secret (deploy key or PAT)
- Non-fatal: log error and continue if publish fails (S3 remains source of truth)

## Consequences

### Positive

- Public benchmark portfolio at zero hosting cost
- Monorepo keeps everything reviewable in one PR
- No cross-repo coordination, no PAT for data publishing (Phase 1)
- Vega-Lite specs are declarative JSON — add new chart types without code changes
- Cross-experiment comparison (overlay metrics from different runs)
- `ExperimentSummary` JSON is the only contract — site works with any experiment type
- Static site has no runtime to monitor or restart
- Data committed as plain JSON — Git delta compression handles it efficiently, no LFS needed

### Negative

- Site node_modules in same repo (mitigate: `.gitignore`, only `site/` touches npm)
- Manual data pipeline until Phase 3 (copy from S3, commit, push)
- Vega-Lite has a learning curve vs raw Chart.js for custom interactions
- Site design/layout requires frontend work upfront
- GitHub Pages build adds ~30s to deploy on every push to `site/`

### Future

- Phase 3: Operator auto-publish removes manual data copy step
- Add Bencher.dev integration for regression detection once experiment cadence increases
- Embed Grafana panel links for live experiment monitoring (VictoriaMetrics dashboards)
- Add experiment metadata tags to `ExperimentSummary` for richer filtering
- If data volume grows beyond ~500 files, evaluate extracting to a separate repo

## References

- [ADR-009: TSDB Selection](ADR-009-tsdb-selection.md) — Metrics comparison experiments that will populate the site
- [ADR-011: Observability Architecture](ADR-011-observability-architecture.md) — Stack producing the metrics
- [ADR-015: Experiment Operator](ADR-015-experiment-operator.md) — Operator lifecycle and `collectAndStoreResults()`
- [ADR-016: Hub Metrics Backend](ADR-016-hub-metrics-backend.md) — VictoriaMetrics as metrics source
- [Astro Documentation](https://docs.astro.build/)
- [Vega-Lite Documentation](https://vega.github.io/vega-lite/)
- [GitHub Pages Documentation](https://docs.github.com/en/pages)
- Operator source: `operators/experiment-operator/internal/controller/experiment_controller.go`
