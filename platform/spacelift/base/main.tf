# Spacelift Root Stack
# This administrative stack manages all other Spacelift stacks declaratively
#
# Prerequisites:
# 1. Create Spacelift account at https://spacelift.io (free tier)
# 2. Connect this GitHub repository to Spacelift
# 3. Create this stack manually as "spacelift-root" with administrative=true
# 4. Configure cloud credentials in Spacelift contexts

terraform {
  required_version = ">= 1.0.0"

  required_providers {
    spacelift = {
      source  = "spacelift-io/spacelift"
      version = "~> 1.0"
    }
  }
}

# =============================================================================
# Local Variables
# =============================================================================

locals {
  repository = "illMadeCoder/illm-k8s-lab"
  branch     = "main"

  # Common labels for all stacks
  common_labels = ["illm-k8s-lab", "portfolio"]

  # Cloud provider configurations
  clouds = {
    azure = {
      context     = "azure-credentials"
      foundation  = "azure-foundation"
      module_path = "terraform-modules/aks"
    }
    aws = {
      context     = "aws-credentials"
      foundation  = "aws-foundation"
      module_path = "terraform-modules/eks"
    }
  }
}

# =============================================================================
# Foundation Stacks - Core Networking (Manual Deploy)
# =============================================================================

resource "spacelift_stack" "azure_foundation" {
  name        = "azure-foundation"
  description = "Azure networking foundation (VNet, subnets, NSGs) for AKS clusters"

  repository   = local.repository
  branch       = local.branch
  project_root = "terraform-modules/azure-networking"

  administrative = false
  autodeploy     = false # Manual trigger for foundation changes

  labels = concat(local.common_labels, ["foundation", "azure", "networking"])
}

resource "spacelift_stack" "aws_foundation" {
  name        = "aws-foundation"
  description = "AWS networking foundation (VPC, subnets, IGW) for EKS clusters"

  repository   = local.repository
  branch       = local.branch
  project_root = "terraform-modules/aws-networking"

  administrative = false
  autodeploy     = false

  labels = concat(local.common_labels, ["foundation", "aws", "networking"])
}

# Attach cloud credentials to foundation stacks
resource "spacelift_context_attachment" "azure_foundation" {
  context_id = spacelift_context.azure_credentials.id
  stack_id   = spacelift_stack.azure_foundation.id
  priority   = 0
}

resource "spacelift_context_attachment" "aws_foundation" {
  context_id = spacelift_context.aws_credentials.id
  stack_id   = spacelift_stack.aws_foundation.id
  priority   = 0
}

# =============================================================================
# Credential Contexts
# =============================================================================

resource "spacelift_context" "azure_credentials" {
  name        = "azure-credentials"
  description = "Azure service principal credentials for AKS provisioning"
  labels      = ["credentials", "azure"]

  # Environment variables are set manually in Spacelift UI for security:
  # - ARM_CLIENT_ID
  # - ARM_CLIENT_SECRET (write-only)
  # - ARM_SUBSCRIPTION_ID
  # - ARM_TENANT_ID
}

resource "spacelift_context" "aws_credentials" {
  name        = "aws-credentials"
  description = "AWS IAM credentials for EKS provisioning"
  labels      = ["credentials", "aws"]

  # Environment variables are set manually in Spacelift UI for security:
  # - AWS_ACCESS_KEY_ID
  # - AWS_SECRET_ACCESS_KEY (write-only)
  # - AWS_DEFAULT_REGION
}

# =============================================================================
# Experiment Stacks (Created Dynamically)
# =============================================================================

# Azure experiment stacks
resource "spacelift_stack" "experiment_azure" {
  for_each = {
    for name, config in var.experiments : name => config
    if contains(config.clouds, "azure") || contains(config.clouds, "both")
  }

  name        = "exp-${each.key}-aks"
  description = "Azure AKS cluster(s) for ${each.key} experiment"

  repository   = local.repository
  branch       = local.branch
  project_root = "experiments/${each.key}/terraform/azure"

  administrative = false
  autodeploy     = each.value.autodeploy

  labels = concat(
    local.common_labels,
    ["experiment", "azure", "aks", each.key]
  )
}

# AWS experiment stacks
resource "spacelift_stack" "experiment_aws" {
  for_each = {
    for name, config in var.experiments : name => config
    if contains(config.clouds, "aws") || contains(config.clouds, "both")
  }

  name        = "exp-${each.key}-eks"
  description = "AWS EKS cluster(s) for ${each.key} experiment"

  repository   = local.repository
  branch       = local.branch
  project_root = "experiments/${each.key}/terraform/aws"

  administrative = false
  autodeploy     = each.value.autodeploy

  labels = concat(
    local.common_labels,
    ["experiment", "aws", "eks", each.key]
  )
}

# Attach credentials to experiment stacks
resource "spacelift_context_attachment" "experiment_azure" {
  for_each = spacelift_stack.experiment_azure

  context_id = spacelift_context.azure_credentials.id
  stack_id   = each.value.id
  priority   = 0
}

resource "spacelift_context_attachment" "experiment_aws" {
  for_each = spacelift_stack.experiment_aws

  context_id = spacelift_context.aws_credentials.id
  stack_id   = each.value.id
  priority   = 0
}

# =============================================================================
# Stack Dependencies
# =============================================================================

# Azure experiments depend on Azure foundation
resource "spacelift_stack_dependency" "azure_experiments_on_foundation" {
  for_each = spacelift_stack.experiment_azure

  stack_id            = each.value.id
  depends_on_stack_id = spacelift_stack.azure_foundation.id
}

# AWS experiments depend on AWS foundation
resource "spacelift_stack_dependency" "aws_experiments_on_foundation" {
  for_each = spacelift_stack.experiment_aws

  stack_id            = each.value.id
  depends_on_stack_id = spacelift_stack.aws_foundation.id
}

# =============================================================================
# Policies
# =============================================================================

resource "spacelift_policy" "plan_approval" {
  name        = "plan-approval"
  type        = "APPROVAL"
  body        = file("${path.module}/../../.spacelift/policies/plan-approval.rego")
  description = "Controls auto-approval of Terraform plans"
  labels      = ["approval", "governance"]
}

# Attach policy to all stacks
resource "spacelift_policy_attachment" "azure_foundation" {
  policy_id = spacelift_policy.plan_approval.id
  stack_id  = spacelift_stack.azure_foundation.id
}

resource "spacelift_policy_attachment" "aws_foundation" {
  policy_id = spacelift_policy.plan_approval.id
  stack_id  = spacelift_stack.aws_foundation.id
}

resource "spacelift_policy_attachment" "experiment_azure" {
  for_each = spacelift_stack.experiment_azure

  policy_id = spacelift_policy.plan_approval.id
  stack_id  = each.value.id
}

resource "spacelift_policy_attachment" "experiment_aws" {
  for_each = spacelift_stack.experiment_aws

  policy_id = spacelift_policy.plan_approval.id
  stack_id  = each.value.id
}
