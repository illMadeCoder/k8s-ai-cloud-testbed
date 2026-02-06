# ADR-001: GitLab CI for IaC Orchestration

## Status

**Superseded** by [ADR-012: Crossplane Experiment Abstraction](ADR-012-crossplane-experiment-abstraction.md)

Cloud infrastructure provisioning is now handled entirely by Crossplane running on the hub cluster. GitLab CI and Terraform are no longer used in this project.

## Context

Need a platform to manage Terraform deployments with state management, credential handling, and PR-based workflows. Key requirements:
- 100% CLI/API driven (agentic workflow)
- Free tier for learning
- Recognized on job postings

## Decision

**Use GitLab CI** with GitLab-managed Terraform state.

## Comparison

| Factor | GitLab CI | Spacelift | Terraform Cloud | GitHub Actions |
|--------|-----------|-----------|-----------------|----------------|
| **100% CLI/API** | Yes | No (UI bootstrap) | Yes | Yes |
| **Free tier** | 400 mins/mo | 1 worker | 500 resources | 2000 mins/mo |
| **State backend** | Built-in | Built-in | Built-in | DIY (S3/Azure) |
| **Stack dependencies** | DIY | Native | Limited | DIY |
| **OPA policies** | DIY | Native | Sentinel (paid) | DIY |
| **Job postings** | High | Rare | Medium | High |

## Why GitLab CI

- **Fully agentic** - `glab` CLI for everything, no UI required
- **Built-in state** - GitLab-managed Terraform state, no extra backend
- **Resume value** - GitLab widely recognized in job postings
- **Learning value** - Building workflows yourself teaches more than TACOS

## Why Not Others

**Spacelift**: Requires UI bootstrap, not fully agentic

**GitHub Actions**: Already using GitHub; GitLab adds skill diversity

**Terraform Cloud**: RUM pricing, Sentinel requires paid tier

## Trade-offs

**Losing**: Native stack dependencies, drift detection, OPA integration

**Gaining**: Transferable skills, resume recognition, full CLI control

Stack dependencies and drift detection can be built in GitLab CI pipelines - more work, but demonstrates deeper understanding.

## References

- [GitLab IaC with Terraform](https://docs.gitlab.com/user/infrastructure/iac/)
- [GitLab-managed Terraform State](https://docs.gitlab.com/ee/user/infrastructure/iac/terraform_state.html)

## Decision Date

2025-12-10
