# CLAUDE.md

Track progress in `TODO.md`. Use `/compact` at ~70% context.

## Conventions

- **ArgoCD apps**: Use labels `experiment: {name}`, `cluster: target|loadgen`
- **ArgoCD patterns**: Multi-source, sync waves, `ignoreDifferences` for CRDs (see `docs/gitops-patterns.md`)
- **Terraform**: Spacelift manages stacks; credentials in Spacelift contexts, not in code
