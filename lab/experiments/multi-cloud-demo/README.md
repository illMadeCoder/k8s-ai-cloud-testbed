# Multi-Cloud Demo Experiment

This experiment demonstrates the multi-cloud deployment capabilities of the illm-k8s-lab platform using GitLab CI + Terraform + Crossplane.

## Overview

The multi-cloud demo deploys:
- **Infrastructure**: AKS (Azure) or EKS (AWS) cluster via GitLab CI + Terraform
- **Application**: Demo web app via ArgoCD
- **Cloud Resources**: Database, cache, storage, and queue via Crossplane

## Cloud Abstraction

The key feature of this demo is **cloud-agnostic resource claims**. The same Crossplane claims in `crossplane/claims.yaml` work on both Azure and AWS:

```yaml
apiVersion: illm.io/v1alpha1
kind: Database
spec:
  engine: postgresql
  size: small      # Automatically maps to cloud-specific SKUs
  storageGB: 20
```

On Azure, this creates an Azure PostgreSQL Flexible Server.
On AWS, this creates an RDS PostgreSQL instance.

## Directory Structure

```
multi-cloud-demo/
├── terraform/
│   ├── azure/          # AKS cluster configuration
│   │   └── main.tf
│   └── aws/            # EKS cluster configuration
│       └── main.tf
├── argocd/
│   └── target.yaml     # ArgoCD app (deploys demo + claims)
├── crossplane/
│   └── claims.yaml     # Cloud-agnostic resource claims
├── k6/
│   └── k6-scripts.yaml # Load test scripts
└── README.md
```

## Deployment

### Via GitLab CI (Recommended)

```bash
# Deploy to Azure
task exp:deploy NAME=multi-cloud-demo CLOUD=azure

# Deploy to AWS
task exp:deploy NAME=multi-cloud-demo CLOUD=aws

# Deploy to both (parallel)
task exp:deploy NAME=multi-cloud-demo CLOUD=both
```

### Via Terraform (Manual)

```bash
# Azure
cd experiments/multi-cloud-demo/terraform/azure
terraform init && terraform apply

# AWS
cd experiments/multi-cloud-demo/terraform/aws
terraform init && terraform apply
```

## Verification

Check Crossplane claims status:

```bash
task crossplane:claims
```

Expected output:
```
=== Databases ===
NAMESPACE   NAME       ENGINE       SIZE    ENDPOINT                              CLOUD   AGE
demo        demo-db    postgresql   small   demo-db.postgres.database.azure.com   azure   5m

=== Caches ===
NAMESPACE   NAME         ENGINE   SIZE    ENDPOINT                    PORT   CLOUD   AGE
demo        demo-cache   redis    small   demo-cache.redis.cache...   6380   azure   5m
```

## Cleanup

```bash
# Via GitLab CI
task exp:destroy NAME=multi-cloud-demo CLOUD=azure

# Via Terraform
cd experiments/multi-cloud-demo/terraform/azure
terraform destroy
```

## Cost Considerations

This experiment creates real cloud resources. Estimated costs:
- **Azure**: ~$50-100/day (AKS + PostgreSQL + Redis + Storage)
- **AWS**: ~$50-100/day (EKS + RDS + ElastiCache + S3)

Always destroy resources when done testing!
