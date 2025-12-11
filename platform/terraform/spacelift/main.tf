terraform {
  required_version = ">= 1.0"

  required_providers {
    spacelift = {
      source  = "spacelift-io/spacelift"
      version = "~> 1.0"
    }
  }
}

# Spacelift provider authenticates via SPACELIFT_API_KEY_ENDPOINT,
# SPACELIFT_API_KEY_ID, and SPACELIFT_API_KEY_SECRET environment variables
# (set automatically when running in Spacelift)

# -----------------------------------------------------------------------------
# Contexts (shared variables/secrets across stacks)
# -----------------------------------------------------------------------------

# Azure context created manually during bootstrap (contains ARM_* secrets)
# We reference it here rather than creating it
data "spacelift_context" "azure" {
  context_id = "azure-credentials"
}

# AWS context - will add credentials when we set up AWS
resource "spacelift_context" "aws" {
  name        = "aws-credentials"
  description = "AWS credentials for Terraform"
  labels      = ["aws"]
}

# -----------------------------------------------------------------------------
# Azure Stacks
# -----------------------------------------------------------------------------

resource "spacelift_stack" "azure_foundation" {
  name        = "azure-foundation"
  description = "Azure foundation: Resource Group, Key Vault, Service Principals"

  repository = "illm-k8s-lab"
  branch     = "main"
  project_root = "platform/terraform/azure/foundation"

  manage_state = true
  autodeploy   = true

  labels = ["azure", "foundation"]
}

resource "spacelift_context_attachment" "azure_foundation" {
  context_id = data.spacelift_context.azure.id
  stack_id   = spacelift_stack.azure_foundation.id
}

resource "spacelift_stack" "azure_aks" {
  name        = "azure-aks"
  description = "Azure Kubernetes Service cluster"

  repository = "illm-k8s-lab"
  branch     = "main"
  project_root = "platform/terraform/aks"

  manage_state = true
  autodeploy   = false # Manual deploy for cost control

  labels = ["azure", "kubernetes"]
}

resource "spacelift_context_attachment" "azure_aks" {
  context_id = data.spacelift_context.azure.id
  stack_id   = spacelift_stack.azure_aks.id
}

# AKS depends on foundation (needs resource group, potentially Key Vault)
resource "spacelift_stack_dependency" "aks_on_foundation" {
  stack_id            = spacelift_stack.azure_aks.id
  depends_on_stack_id = spacelift_stack.azure_foundation.id
}

# -----------------------------------------------------------------------------
# AWS Stacks (placeholder for Phase 3.1)
# -----------------------------------------------------------------------------

# resource "spacelift_stack" "aws_foundation" {
#   name        = "aws-foundation"
#   description = "AWS foundation: Secrets Manager, IAM"
#
#   repository = "illm-k8s-lab"
#   branch     = "main"
#   project_root = "platform/terraform/aws/foundation"
#
#   manage_state = true
#   autodeploy   = true
#
#   labels = ["aws", "foundation"]
# }

# resource "spacelift_stack" "aws_eks" {
#   name        = "aws-eks"
#   description = "AWS Elastic Kubernetes Service cluster"
#
#   repository = "illm-k8s-lab"
#   branch     = "main"
#   project_root = "platform/terraform/eks"
#
#   manage_state = true
#   autodeploy   = false
#
#   labels = ["aws", "kubernetes"]
# }
