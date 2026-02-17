---
description: Apply an experiment and monitor it through completion and PR merge
allowed-tools: Bash, Read, Glob, Grep, AskUserQuestion
argument-hint: [experiment-name]
---

# Run Experiment

Apply an experiment, monitor it through all lifecycle phases, track the AI analyzer, and manage the PR review/merge flow.

## Step 1: Parse Arguments & Read Experiment YAML

Extract the experiment name from `$ARGUMENTS`. If empty, use `AskUserQuestion` to ask the user which experiment to run.

Validate that `experiments/{name}/experiment.yaml` exists using the Glob tool. If it doesn't exist, tell the user and stop (suggest `/create-experiment` to scaffold one).

Read `experiments/{name}/experiment.yaml` and extract these key fields for use throughout:
- `metadata.generateName` — prefix used to identify instance names
- `spec.publish` — whether this experiment publishes to the benchmark site (controls Steps 6-7)
- `spec.workflow.template` — the WorkflowTemplate name
- `spec.workflow.completion.mode` — `workflow` or `manual`
- `spec.targets[*].name` — target names (e.g., `app`, `loadgen`)

Store these in your working memory — you'll reference them in every subsequent step.

## Toil Tracking

Throughout this skill, log friction and failures as beads for post-mortem analysis. This captures recurring patterns across experiment runs so they can be prioritized and fixed.

**When to create a bead:** any time something goes wrong or requires manual intervention — pre-flight failures, phase timeouts, experiment failures, analyzer failures, PR issues. Essentially: if the happy path didn't happen, log it.

**How to create a bead:**
```bash
bd create --title="{short description}" --type=bug --priority={0-4} -l toil -l {experiment-name}
```

Use the experiment name (directory name, not instance name) as a label so toil can be filtered per experiment. Add additional labels from: `flaky`, `config`, `timing`, `resources`, `networking`, `crossplane`, `argocd`, `operator`, `analyzer`, `workflow`.

**Priority guide for experiment toil:**
- P1: Experiment can't run at all (operator down, CRD missing, apply fails)
- P2: Experiment ran but failed or required manual intervention (timeout, phase stuck, workflow failure)
- P3: Experiment completed but with friction (analyzer failed, PR issues, slow phases)
- P4: Minor annoyance (warnings, slow but within timeout)

**Do NOT prompt the user** before creating toil beads — just create them and report what was logged. The whole point is zero-friction capture.

## Step 2: Pre-flight Checks

Run three checks before applying. Report results as a checklist.

### 2a. Validate experiment YAML

Run the `experiment-validator` agent (via the Task tool with `subagent_type: "experiment-validator"`) passing the experiment name. Report the results:
- All PASS: proceed
- Any WARN: show warnings, proceed
- Any FAIL: show failures, **stop** — tell the user to fix issues before running. Log toil:
  ```bash
  bd create --title="Validation failed: {name} - {first failure reason}" --type=bug --priority=2 -l toil -l {name} -l config
  ```

### 2b. WorkflowTemplate check

Check if the experiment's WorkflowTemplate exists in the cluster:

```bash
kubectl get workflowtemplate {workflow-template-name} -n argo-workflows -o name 2>&1
```

Also check if a local template file exists: `experiments/{name}/workflow/template.yaml`

**If template is NOT in cluster but local file exists**: ask the user whether to apply it:
```bash
kubectl apply -f experiments/{name}/workflow/template.yaml
```

**If template is NOT in cluster and no local file**: warn the user that the workflow will fail without it. Log toil:
```bash
bd create --title="Missing WorkflowTemplate: {template-name} for {name}" --type=bug --priority=2 -l toil -l {name} -l workflow -l config
```
Use `AskUserQuestion` to ask whether to continue anyway or stop.

**If template IS in cluster**: report OK.

### 2c. Operator health

```bash
kubectl get deployment experiment-operator-controller-manager -n experiment-operator-system -o jsonpath='{.status.readyReplicas}' 2>&1
```

- If readyReplicas >= 1: OK
- If 0 or error: **warn** (don't block) — the operator may recover, but the experiment might not progress. Log toil:
  ```bash
  bd create --title="Operator unhealthy at experiment start: {name}" --type=bug --priority=1 -l toil -l {name} -l operator
  ```

Report the pre-flight summary, e.g.:
```
Pre-flight:
  [PASS] Experiment YAML valid
  [PASS] WorkflowTemplate gateway-comparison-validation found in cluster
  [PASS] Operator healthy (1 ready replica)
```

## Step 3: Apply Experiment

Apply the experiment:

```bash
kubectl create -f experiments/{name}/experiment.yaml
```

Parse the generated instance name from the output. The output will look like:
```
experiment.experiments.illm.io/gw-comp-b5jbs created
```

Extract the instance name (e.g., `gw-comp-b5jbs`). This is the **instance name** used for ALL subsequent kubectl commands. Store it for the rest of the session.

Report: `Applied experiment: {instance-name} (from experiments/{name}/)`

## Step 4: Monitor Lifecycle Phases

Monitor the experiment through its phase transitions. Use a polling approach with phase-specific intervals and timeouts.

### Primary status query

Use a single kubectl call per poll to minimize API calls:

```bash
kubectl get experiment {instance-name} -n experiments -o jsonpath='{.status.phase}|{.status.published}|{.status.analysisPhase}|{.status.publishPRURL}|{.status.resultsURL}' 2>&1
```

This returns: `{phase}|{published}|{analysisPhase}|{prURL}|{resultsURL}`

### Phase transition table

| Waiting for | Poll interval | Timeout | On timeout |
|-------------|--------------|---------|------------|
| Pending → Provisioning | 30s | 20 min | Crossplane claims, operator logs |
| Provisioning → Ready | 30s | 20 min | Crossplane claims, operator logs |
| Ready → Running | 15s | 10 min | ArgoCD app health/sync |
| Running → Complete/Failed | 15s | No timeout | Workflow status (manual mode: remind user) |

### Poll loop

For each phase transition:

1. Report the current phase and what we're waiting for
2. Sleep for the poll interval using: `sleep {interval}`
3. Query status using the primary status query above
4. If phase unchanged and timeout not reached: continue polling
5. If phase changed: report transition and move to next phase
6. If terminal phase (Complete/Failed): break to Step 5

### Phase-specific supplementary info

Show these **once when entering a phase**, not every poll:

**Provisioning**: Show target phases and cluster names:
```bash
kubectl get experiment {instance-name} -n experiments -o jsonpath='{range .status.targets[*]}{.name}: {.phase} (cluster: {.clusterName}){"\n"}{end}'
```

**Ready**: Show ArgoCD app health:
```bash
kubectl get applications -l experiment={instance-name} -n argocd -o custom-columns='NAME:.metadata.name,HEALTH:.status.health.status,SYNC:.status.sync.status' --no-headers
```

**Running**: Show workflow status:
```bash
kubectl get experiment {instance-name} -n experiments -o jsonpath='Workflow: {.status.workflowStatus.name} ({.status.workflowStatus.phase})'
```

For **manual** completion mode: when entering Running, remind the user:
```
This is a manual-completion experiment. It will stay Running until you delete it.
Use: kubectl delete experiment {instance-name} -n experiments
```

### Timeout handling

When a phase exceeds its timeout, **log toil immediately** (before asking the user):
```bash
bd create --title="{phase} timeout ({timeout}m) for {instance-name}" --type=bug --priority=2 -l toil -l {name} -l timing -l {phase-label}
```
Where `{phase-label}` is `crossplane` for Pending/Provisioning, `argocd` for Ready, `workflow` for Running.

Then use `AskUserQuestion` with these options:
1. **Continue waiting** — extend timeout by the same duration
2. **Show diagnostics** — run phase-specific diagnostic commands (see below), then ask again
3. **Abort** — delete the experiment: `kubectl delete experiment {instance-name} -n experiments`

**Diagnostic commands by phase:**

Pending/Provisioning:
```bash
kubectl get gkecluster -l experiment={instance-name} -o wide
kubectl logs deployment/experiment-operator-controller-manager -n experiment-operator-system --tail=30 | grep -i "{instance-name}"
```

Ready:
```bash
kubectl get applications -l experiment={instance-name} -n argocd -o custom-columns='NAME:.metadata.name,HEALTH:.status.health.status,SYNC:.status.sync.status,MESSAGE:.status.conditions[0].message'
```

Running:
```bash
kubectl get workflow -l experiment={instance-name} -n argo-workflows -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,MESSAGE:.status.message'
```

## Step 5: Handle Completion

### On Complete

Report success with key info:
```
Experiment {instance-name} completed successfully!

Results: {resultsURL}
Hypothesis: {hypothesisResult or "N/A (no criteria)"}
Duration: {calculate from creation to completedAt if available}
```

Get the hypothesis result:
```bash
kubectl get experiment {instance-name} -n experiments -o jsonpath='{.status.hypothesisResult}'
```

Then proceed to Step 5.5 if `publish: true`, otherwise skip to Step 8.

## Step 5.5: Results Quality Assessment (publish: true only)

Skip this step entirely if `spec.publish` is not true.

Immediately after the experiment reaches Complete, fetch the raw results from S3 and assess data quality before waiting for the analyzer.

### Fetch summary.json from S3

```bash
kubectl run -n seaweedfs s3fetch-{instance-name} --rm -it --restart=Never \
  --image=curlimages/curl:8.5.0 -- \
  curl -s http://seaweedfs-s3.seaweedfs.svc.cluster.local:8333/experiment-results/{instance-name}/summary.json
```

Parse the JSON output. If the fetch fails or returns empty, log toil and skip quality assessment (proceed to Step 6).

### Quality Checks

Run all of these checks and track their results (PASS/WARN/FAIL):

**Check 1: Custom Metric Completeness**
- Read the experiment YAML's `spec.metrics[]` to get the list of expected custom metrics
- For each, check if it appears in `summary.metrics.queries` with:
  - `error` is null/absent (query didn't fail)
  - `data` array is non-empty (query returned results)
- Report: `Custom metrics: X/Y collected` with per-metric pass/fail
- If 0/Y collected → FAIL
- If partial (>0 but <Y) → WARN
- If Y/Y → PASS
- If no custom metrics defined in spec → PASS (skip this check)

**Check 2: Metrics Source**
- Check `summary.metrics.source`
- `"target:cadvisor"` means only cadvisor infrastructure metrics were collected (the operator fell back because it couldn't reach target Prometheus/VM)
- If source is `"target:cadvisor"` AND the experiment has custom metrics defined → WARN
- Otherwise → PASS

**Check 3: Hypothesis Machine Verdict**
- Check `summary.hypothesis.machineVerdict` (if success criteria are defined in the experiment YAML)
- `"insufficient"` → FAIL (metrics missing, criteria couldn't be evaluated)
- `"confirmed"` or `"rejected"` → PASS (criteria were evaluable)
- No success criteria defined → N/A (skip this check)

**Check 4: Workload Pod Visibility**
- Check if any non-infrastructure pods appear in `metrics.queries` data (e.g., `cpu_by_pod.data`)
- Infrastructure pod prefixes: `prometheus-*`, `alertmanager-*`, `grafana-*`, `kube-state-metrics-*`, `node-exporter-*`, `kube-prometheus-stack-*`, `alloy-*`, `ts-vm-hub-*`, `tailscale-operator-*`, `operator-*`
- Classify each pod in the data as infrastructure or workload
- Report: "Pod visibility: X infrastructure, Y workload (pod-names...)"
- If ALL pods are infrastructure → FAIL ("No workload pods visible in metrics")
- If at least one non-infrastructure pod appears → PASS
- If no pod-level metrics exist → N/A

### Quality Scorecard

Report the results as a scorecard:
```
Results Quality:
  [PASS] Experiment completed (38m, $0.017)
  [FAIL] Custom metrics: 0/10 collected
  [WARN] Metrics source: target:cadvisor (no custom metric pipeline)
  [FAIL] Workload pods: 0 visible in cadvisor
  [N/A]  Machine verdict: no success criteria defined
```

Determine the overall quality gate status:
- Any FAIL → `qualityGate = "fail"`
- Any WARN (no FAILs) → `qualityGate = "warn"`
- All PASS/N/A → `qualityGate = "pass"`

Store the quality gate status and scorecard for use in Steps 6.5, 7, and 8.

Proceed to Step 6 (Track Analyzer).

### On Failed

**Log toil immediately** with the failure reason from conditions:
```bash
bd create --title="Experiment failed: {instance-name} - {first condition message}" --type=bug --priority=2 -l toil -l {name} -l {category}
```
Choose `{category}` based on which phase the failure occurred in: `crossplane` (Provisioning), `argocd` (Ready), `workflow` (Running), or `operator` (unknown).

Show layered diagnostics:

1. **Experiment conditions:**
```bash
kubectl get experiment {instance-name} -n experiments -o jsonpath='{range .status.conditions[*]}{.type}: {.status} - {.message}{"\n"}{end}'
```

2. **Operator logs (filtered):**
```bash
kubectl logs deployment/experiment-operator-controller-manager -n experiment-operator-system --tail=50 2>&1 | grep "{instance-name}"
```

3. **Workflow status:**
```bash
kubectl get workflow -l experiment={instance-name} -n argo-workflows -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,MESSAGE:.status.message' --no-headers
```

4. **Pod failures in experiment namespace:**
```bash
kubectl get pods -l experiment={instance-name} -n experiments --field-selector=status.phase=Failed -o wide 2>/dev/null
```

5. **Crossplane claim status:**
```bash
kubectl get gkecluster -l experiment={instance-name} -o custom-columns='NAME:.metadata.name,READY:.status.conditions[?(@.type=="Ready")].status,MESSAGE:.status.conditions[?(@.type=="Ready")].message' --no-headers 2>/dev/null
```

After diagnostics, use `AskUserQuestion`:
1. **Delete experiment** — `kubectl delete experiment {instance-name} -n experiments`
2. **Keep for debugging** — leave it and stop

## Step 6: Track Analyzer (publish: true only)

Skip this step entirely if `spec.publish` is not true.

The operator launches an AI analyzer job after publishing results. Track its progress:

Poll `analysisPhase` every 30 seconds (15 minute timeout) using the primary status query from Step 4.

Report transitions:
- **Pending**: "Analyzer job created, waiting for pod scheduling..."
- **Running**: "Analyzer running..." — optionally show analyzer pod logs:
  ```bash
  kubectl logs job/{instance-name}-analyzer -n experiment-operator-system --tail=10 2>/dev/null
  ```
- **Succeeded**: "Analysis complete! Results enriched with AI insights." Then run the AI verdict check (see below).
- **Failed**: Log toil, show analyzer job logs, and ask user how to proceed:
  ```bash
  bd create --title="Analyzer failed for {instance-name}" --type=bug --priority=3 -l toil -l {name} -l analyzer
  kubectl logs job/{instance-name}-analyzer -n experiment-operator-system --tail=50 2>/dev/null
  kubectl describe job {instance-name}-analyzer -n experiment-operator-system 2>/dev/null | tail -20
  ```
  Use `AskUserQuestion`: continue to PR step anyway, or stop.
- **Skipped**: "Analysis skipped (no analyzerConfig or empty sections)."

If the analysis phase field is empty after the experiment completes, wait a few polls — the operator may not have created the job yet. If still empty after 2 minutes, treat as Skipped.

If the analyzer times out (15 minutes), log toil:
```bash
bd create --title="Analyzer timeout (15m) for {instance-name}" --type=bug --priority=3 -l toil -l {name} -l analyzer -l timing
```

### AI Verdict Check (after analyzer succeeds)

After the analyzer succeeds, re-fetch the enriched summary.json from S3 (the analyzer updates it with AI analysis):

```bash
kubectl run -n seaweedfs s3fetch2-{instance-name} --rm -it --restart=Never \
  --image=curlimages/curl:8.5.0 -- \
  curl -s http://seaweedfs-s3.seaweedfs.svc.cluster.local:8333/experiment-results/{instance-name}/summary.json
```

**Check 5: AI Hypothesis Verdict**
- Check `summary.analysis.hypothesisVerdict`
- `"insufficient"` → FAIL (AI confirms data quality is too low for conclusions)
- `"confirmed"` or `"rejected"` → PASS
- Absent → N/A
- Extract the first 1-2 sentences of `summary.analysis.abstract` as a concise explanation

Update the quality scorecard from Step 5.5 with this new check:
```
  [FAIL] AI verdict: insufficient — "no application-level latency, throughput, or I/O metrics..."
```

Recalculate the overall quality gate status (a new FAIL here can change `"warn"` → `"fail"`).

If the analyzer failed, was skipped, or timed out, mark the AI verdict check as N/A and don't change the quality gate status.

Proceed to Step 6.5.

## Step 6.5: Quality Gate Decision (publish: true only)

Skip this step entirely if `spec.publish` is not true.

This is the critical decision point. Behavior depends on the quality gate status determined in Steps 5.5 and 6.

### If `qualityGate == "pass"`

Proceed directly to Step 7 (PR Management) as normal.

### If `qualityGate == "warn"`

Report the quality scorecard with all warnings. Proceed to Step 7 but include the warnings in the PR summary.

### If `qualityGate == "fail"`

1. **Report the full quality scorecard** with all failures highlighted.

2. **Run root cause diagnostics** based on which checks failed:

   **If custom metrics empty + source is cadvisor:**
   - Check if the experiment has `kube-prometheus-stack`, `metrics-agent`, and `metrics-egress` components:
     ```bash
     grep -c 'kube-prometheus-stack\|metrics-agent\|metrics-egress' experiments/{name}/experiment.yaml
     ```
   - If `kube-prometheus-stack` missing: report root cause: "experiment lacks kube-prometheus-stack — the operator queries target-local Prometheus for custom metrics via ServiceMonitor scrapes"
   - If `metrics-agent`/`metrics-egress` missing: report root cause: "experiment lacks metrics-agent + metrics-egress — needed for hub VM fallback path"
   - Check for `$NAMESPACE` vs `$EXPERIMENT` in queries:
     ```bash
     grep '\$NAMESPACE' experiments/{name}/experiment.yaml
     ```
   - If `$NAMESPACE` found: report root cause: "`$NAMESPACE` resolves to the Experiment CR namespace ('experiments'), not the target namespace where pods run. Use `$EXPERIMENT` instead."
   - If all components present and queries correct, check operator logs:
     ```bash
     kubectl logs deployment/experiment-operator-controller-manager \
       -n experiment-operator-system --tail=100 2>&1 | grep -i "metric\|query\|custom\|prometheus\|local"
     ```
   - Note: the operator now tries local Prometheus first (via `DiscoverMonitoringServices`), then falls back to hub VM. Check logs for "Discovered local Prometheus" or "Local Prometheus discovery failed".

   **If workload pod not visible in cadvisor:**
   - Check operator logs for pod deployment info (target cluster is likely cleaned up by now):
     ```bash
     kubectl logs deployment/experiment-operator-controller-manager \
       -n experiment-operator-system --tail=100 2>&1 | grep -i "{instance-name}"
     ```
   - Common causes: monitoring stack pods (prometheus, grafana, etc.) are now filtered as infrastructure but may have been the only pods visible. Check if the workload pod (e.g., naive-db-0) was present and running during metrics collection.
   - Infrastructure pods are filtered by prefix: `prometheus-`, `alertmanager-`, `grafana-`, `kube-state-metrics-`, `node-exporter-`, `kube-prometheus-stack-`, `alloy-`, `ts-vm-hub-`, `tailscale-operator-`, `operator-`

   **If machine verdict insufficient:**
   - Parse `summary.hypothesis.criteria` and list which criteria have `passed: null` and why

3. **Log toil:**
   ```bash
   bd create --title="Quality gate failed: {instance-name} - {primary failure reason}" \
     --type=bug --priority=2 -l toil -l {name} -l quality
   ```

4. **Present remediation options** via `AskUserQuestion`:

   **Option 1: "Fix and re-run" (Recommended)**
   - Explain what needs to change based on the root cause diagnosis, with specific fixes:
     - Missing `kube-prometheus-stack`: add `- app: kube-prometheus-stack` to target components
     - Missing `metrics-agent`/`metrics-egress`: add `- app: metrics-agent` and `- app: metrics-egress` to target components
     - `$NAMESPACE` in queries: replace all `namespace=~"$NAMESPACE"` with `namespace=~"$EXPERIMENT"` in experiment.yaml
     - Missing load generation: add load-test step to workflow template
   - If user picks this: make the fix to the experiment YAML, delete the current experiment (`kubectl delete experiment {instance-name} -n experiments`), close the PR (`gh pr close {pr-number} --delete-branch`), then loop back to Step 3 (Apply Experiment) with the fixed YAML

   **Option 2: "Merge anyway"**
   - Warn: "Results will be published with insufficient data quality"
   - Proceed to Step 7 (PR Management) with a warning banner in the PR summary

   **Option 3: "Skip (keep PR open)"**
   - Leave PR open for manual review later
   - Report the PR URL and stop (proceed to Step 8 summary)

   **Option 4: "Abort (delete experiment + close PR)"**
   - Clean up: `kubectl delete experiment {instance-name} -n experiments` and `gh pr close {pr-number} --delete-branch`
   - Proceed to Step 8 summary with abort status

## Step 7: PR Management (publish: true only)

Skip this step entirely if `spec.publish` is not true.

**Important**: Wait for the analyzer to finish (Step 6) before offering merge — the analyzer commits to the same branch.

### Get PR details

Extract PR info from the experiment status:
```bash
kubectl get experiment {instance-name} -n experiments -o jsonpath='{.status.publishBranch}|{.status.publishPRNumber}|{.status.publishPRURL}'
```

If no PR URL is set yet, report this and skip PR management. Log toil:
```bash
bd create --title="No PR created for published experiment {instance-name}" --type=bug --priority=3 -l toil -l {name} -l operator
```

### Show PR summary

```bash
gh pr view {pr-number} --json title,state,additions,deletions,url
```

Report with quality context:

If `qualityGate == "pass"`:
```
PR #{number}: "{title}" (+{additions}, -{deletions})
Quality: PASS ({X}/{Y} metrics, hypothesis {verdict})
```

If `qualityGate == "warn"`:
```
PR #{number}: "{title}" (+{additions}, -{deletions})
Quality: WARN ({warnings summary})
```

If `qualityGate == "fail"` (user chose "Merge anyway" in Step 6.5):
```
PR #{number}: "{title}" (+{additions}, -{deletions})
Quality: FAIL ({X}/{Y} custom metrics, AI verdict: {verdict})
```

### Ask user what to do

Use `AskUserQuestion` with these options:

1. **Merge PR (squash)** — merge and clean up:
   ```bash
   gh pr merge {pr-number} --squash --delete-branch
   ```
   After merge, check for the deploy-site workflow:
   ```bash
   gh run list -w "Deploy Benchmark Site" -L 1
   ```
   Report the workflow status/URL.

2. **View PR diff** — show the diff:
   ```bash
   gh pr diff {pr-number}
   ```
   Then ask again (loop back to the options).

3. **Local preview instructions** — print these commands for the user:
   ```
   To preview locally:
     git fetch origin
     git checkout {branch}
     cd site && npm run dev
   ```
   Then ask again (loop back to the options).

4. **Skip** — leave PR open for later review. Report the PR URL for reference.

## Step 8: Final Summary

Print a compact summary:

```
--- Experiment Complete ---
Experiment:  {instance-name}
Phase:       Complete (or Failed)
Duration:    Xm (if calculable)
Results:     {resultsURL or "N/A"}
Hypothesis:  {hypothesisResult or "N/A"}
Analysis:    {analysisPhase or "N/A"}
Quality:     {PASS|WARN|FAIL} ({summary, e.g. "10/10 metrics, hypothesis validated" or "0/10 metrics, verdict: insufficient"})
PR:          {publishPRURL} ({Merged/Open/Blocked by quality gate/Closed}) or "N/A"
```

For non-publish experiments, omit the Analysis, Quality, and PR lines.

### Toil sync

At the very end, sync any beads created during this run:
```bash
bd sync
```

If any toil beads were created during this run, add a brief note listing them:
```
Toil logged: {count} issue(s) — run `bd list -l {name}` to review
```
