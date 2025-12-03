# Spacelift Stacks

This directory contains the Spacelift stack configuration that manages infrastructure provisioning for the illm-k8s-lab project.

## Architecture

```
spacelift-root (administrative)
        │
        ├── azure-foundation ─────────────────┐
        │       │                             │
        │       ├── exp-http-baseline-aks     │ Stack Dependencies
        │       └── exp-multi-cloud-demo-aks  │
        │                                     │
        └── aws-foundation ───────────────────┤
                │                             │
                └── exp-multi-cloud-demo-eks  │
                                              ▼
                                    Experiments depend on
                                    foundation networking
```

## Setup Instructions

### 1. Create Spacelift Account

1. Go to [spacelift.io](https://spacelift.io) and sign up (free tier available)
2. Connect your GitHub account
3. Add this repository to Spacelift

### 2. Create the Root Stack

1. In Spacelift, create a new stack named `spacelift-root`
2. Set it as **Administrative** (allows it to manage other stacks)
3. Point to: `spacelift-stacks/base`
4. Enable "Manage state" to use Spacelift's state backend

### 3. Configure Cloud Credentials

Create two contexts in Spacelift with environment variables:

**azure-credentials:**
```
ARM_CLIENT_ID=<service-principal-app-id>
ARM_CLIENT_SECRET=<service-principal-password>  # Mark as write-only
ARM_SUBSCRIPTION_ID=<azure-subscription-id>
ARM_TENANT_ID=<azure-tenant-id>
```

**aws-credentials:**
```
AWS_ACCESS_KEY_ID=<iam-access-key>
AWS_SECRET_ACCESS_KEY=<iam-secret-key>  # Mark as write-only
AWS_DEFAULT_REGION=us-east-1
```

### 4. Trigger Initial Run

1. Trigger a run on the `spacelift-root` stack
2. Review the plan (it will create foundation and experiment stacks)
3. Confirm to apply

## Adding New Experiments

To add a new experiment to Spacelift:

1. Edit `spacelift-stacks/base/variables.tf`
2. Add an entry to the `experiments` variable:

```hcl
my-new-experiment = {
  clouds     = ["azure"]  # or ["aws"] or ["both"]
  autodeploy = false
  clusters = {
    target = {
      provider   = "azure"
      vm_size    = "Standard_D4s_v3"
      node_count = 3
    }
  }
}
```

3. Commit and push
4. Spacelift will detect the change and create the new stack

## Stack Naming Convention

- Foundation stacks: `azure-foundation`, `aws-foundation`
- Experiment stacks: `exp-<experiment-name>-<provider>`
  - Example: `exp-http-baseline-aks`, `exp-multi-cloud-demo-eks`

## Triggering Runs

### Via Spacelift UI
Navigate to the stack and click "Trigger" → "Tracked run"

### Via Spacelift CLI
```bash
# Install spacectl
brew install spacelift-io/spacelift/spacectl

# Login
spacectl profile login illm-k8s-lab

# List stacks
spacectl stack list

# Trigger a run
spacectl stack run-manual --id exp-http-baseline-aks
```

### Via Taskfile (Recommended)
```bash
# Deploy experiment via Spacelift
task exp:deploy:spacelift NAME=http-baseline CLOUD=azure

# Destroy experiment
task exp:destroy:spacelift NAME=http-baseline CLOUD=azure
```

## Free Tier Limits

Spacelift free tier includes:
- 2 users
- 200 runs per month
- Unlimited stacks
- Public workers

To stay within limits:
- Use manual triggers for foundation stacks
- Foundation stacks rarely need updates
- Experiment runs are the main consumption
