# CLAUDE.md

Track progress in `docs/roadmap.md`. Use `/compact` at ~70% context.

## Environment Constraints

**This WSL has ~8GB RAM. Do NOT run Docker, Kind, or builds locally.**

- No `docker build`, `docker run`, `docker compose` — kills WSL responsiveness
- No `go build` for the operator locally — CI does this
- This machine is a **terminal only**: `kubectl`, `talosctl`, `gh`, `git`, editing code
- The Talos cluster is on the LAN at `192.168.1.178` (node: `talos-23n-3ay`)
- Operator images are built by GitHub Actions CI and pushed to `ghcr.io/illmadecoder/experiment-operator`

## Running Experiments

### Deploy operator changes
```bash
# 1. Commit and push — CI builds & pushes image to ghcr.io
git add <files> && git commit -m "feat: ..." && git push

# 2. Watch CI: gh run watch  (or gh run list)
gh run list -w "Build Experiment Operator" -L 3

# 3. Once CI passes, restart operator to pull new :latest
kubectl rollout restart deployment/experiment-operator-controller-manager -n experiment-operator-system
kubectl rollout status deployment/experiment-operator-controller-manager -n experiment-operator-system
```

### Apply infra changes (CRDs, buckets, secrets)
```bash
# CRD updates (after make manifests)
kubectl apply -f operators/experiment-operator/config/crd/bases/

# SeaweedFS bucket creation (re-run job)
kubectl delete job seaweedfs-create-buckets -n seaweedfs --ignore-not-found
kubectl apply -f platform/manifests/seaweedfs-config/buckets.yaml

# S3 credentials for Argo Workflows
kubectl apply -f platform/manifests/seaweedfs-config/s3-credentials.yaml

# ArgoCD auto-syncs Argo Workflows config from git
```

### Run an experiment
```bash
# Create (generateName gives unique name each time)
kubectl create -f experiments/hello-app/experiment.yaml

# Watch lifecycle: Pending → Provisioning → Ready → Running → Complete
kubectl get experiments -n experiments -w

# Check results after completion
kubectl get experiments -n experiments -o wide   # Shows ResultsURL column

# Verify S3 results (port-forward SeaweedFS or use a curl pod)
kubectl run -n seaweedfs s3check --rm -it --restart=Never \
  --image=curlimages/curl:8.5.0 -- \
  curl -s http://seaweedfs-s3.seaweedfs.svc.cluster.local:8333/experiment-results/<name>/summary.json
```

## Operational Toil Tracking (beads)

Use `bd` to track recurring issues during lab simulations. This helps identify toil patterns and prioritize fixes.

**Labels for categorization:**
- `toil` - Recurring manual work that should be automated
- `flaky` - Intermittent failures requiring investigation
- `config` - Configuration issues or drift
- `timing` - Race conditions, startup ordering issues
- `resources` - Resource limits, OOM, CPU throttling
- `networking` - DNS, connectivity, service mesh issues

**Lab-specific labels:** `loki-tutorial`, `slo-tutorial`, `logging-comparison`, etc.

**Priority guide:**
- P0: Blocks lab completely
- P1: Workaround exists but painful
- P2: Annoying but manageable
- P3: Minor friction

**Workflow:**
```bash
# Log issue during lab simulation
bd create "Issue title" -l toil,loki-tutorial -p 2 -d "Description of what happened"

# Find recurring patterns
bd list -l toil --sort priority
bd count --by-label              # See which labs have most issues
bd duplicates                    # Find repeated issues

# After fixing
bd close <id>
bd sync && git push
```

## Conventions

- **ArgoCD apps**: Use labels `experiment: {name}`, `cluster: target|loadgen`
- **ArgoCD patterns**: Multi-source, sync waves, `ignoreDifferences` for CRDs (see `docs/gitops-patterns.md`)
- **Terraform**: GitLab CI pipelines; credentials in GitLab CI variables, state in GitLab-managed backend
