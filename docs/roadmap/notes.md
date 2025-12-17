## Notes

**Three-Tier Cluster Architecture:**
- **Hub Cluster**: Persistent, hosts ArgoCD (root), OpenBao, Registry - portable across Kind/K3s/Talos/Cloud
- **Orchestrator Cluster**: Per-experiment, hosts ArgoCD + Argo Workflows, ephemeral
- **Target Cluster(s)**: Per-experiment, hosts workloads under test, ephemeral

**Experiment Structure:**
```
experiments/scenarios/<name>/
├── experiment.yaml              # Metadata (cluster providers, overlays)
├── orchestrator/
│   ├── cluster.yaml            # Orchestrator cluster config
│   └── gitops/
│       ├── root-app.yaml       # Orchestrator's app-of-apps
│       ├── argocd/             # ArgoCD config for this experiment
│       ├── argo-workflows/     # Workflow engine config
│       └── targets/            # ArgoCD apps for target cluster(s)
├── target/
│   ├── cluster.yaml            # Target cluster config
│   └── workloads/              # What gets deployed to target
├── loadgen/                     # Optional: separate load generator cluster
│   ├── cluster.yaml
│   └── workloads/
└── workflow/
    └── experiment.yaml         # Argo Workflow (the actual test)
```

**Experiment Lifecycle:**
```bash
task exp:run experiment=http-baseline    # Create clusters, deploy, run workflow
task exp:status experiment=http-baseline # Check status
task exp:teardown experiment=http-baseline # Clean up
```

**Other Notes:**
- **CapEx over OpEx**: Home lab infrastructure is self-hosted; cloud resources only for experiments that require them
- GitLab CI + Terraform for cloud IaC when experiments need cloud resources
- Crossplane for K8s-native cloud resource provisioning
- Ansible for initial Talos provisioning, not ongoing management
- Each experiment should have a portfolio-ready README
