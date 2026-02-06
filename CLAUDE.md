# CLAUDE.md

Track progress in `docs/roadmap.md`. Use `/compact` at ~70% context.

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
- **Infrastructure**: Crossplane for cloud resource provisioning; credentials synced from OpenBao via ExternalSecrets
