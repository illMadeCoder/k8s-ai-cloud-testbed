# EKS Module Outputs
# Interface mirrors AKS module for consistency

output "cluster_id" {
  description = "The EKS cluster ID"
  value       = module.eks.cluster_id
}

output "cluster_name" {
  description = "The EKS cluster name"
  value       = module.eks.cluster_name
}

output "cluster_endpoint" {
  description = "The EKS cluster API endpoint"
  value       = module.eks.cluster_endpoint
}

# Alias for consistency with AKS module
output "cluster_fqdn" {
  description = "The cluster endpoint (alias for AKS compatibility)"
  value       = module.eks.cluster_endpoint
}

output "kube_config" {
  description = "Kubeconfig for the cluster (uses aws eks get-token for auth)"
  value       = <<-EOT
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: ${module.eks.cluster_endpoint}
    certificate-authority-data: ${module.eks.cluster_certificate_authority_data}
  name: ${module.eks.cluster_name}
contexts:
- context:
    cluster: ${module.eks.cluster_name}
    user: ${module.eks.cluster_name}
  name: ${module.eks.cluster_name}
current-context: ${module.eks.cluster_name}
users:
- name: ${module.eks.cluster_name}
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: aws
      args:
        - eks
        - get-token
        - --cluster-name
        - ${module.eks.cluster_name}
        - --region
        - ${var.region}
EOT
  sensitive   = true
}

output "kube_config_host" {
  description = "Kubernetes API server host"
  value       = module.eks.cluster_endpoint
}

output "cluster_ca_certificate" {
  description = "Cluster CA certificate (base64 encoded)"
  value       = module.eks.cluster_certificate_authority_data
  sensitive   = true
}

# =============================================================================
# IRSA / Workload Identity Outputs
# =============================================================================

output "oidc_provider_arn" {
  description = "OIDC provider ARN for IRSA"
  value       = module.eks.oidc_provider_arn
}

output "oidc_provider_url" {
  description = "OIDC provider URL (without https://)"
  value       = module.eks.oidc_provider
}

output "crossplane_role_arn" {
  description = "IAM role ARN for Crossplane AWS provider (if created)"
  value       = var.enable_crossplane_irsa ? module.crossplane_irsa[0].iam_role_arn : null
}

# =============================================================================
# Networking Outputs
# =============================================================================

output "vpc_id" {
  description = "VPC ID"
  value       = var.vpc_id != null ? var.vpc_id : module.vpc[0].vpc_id
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = var.subnet_ids != null ? var.subnet_ids : module.vpc[0].private_subnets
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = var.vpc_id != null ? [] : module.vpc[0].public_subnets
}

# =============================================================================
# Node Group Outputs
# =============================================================================

output "node_security_group_id" {
  description = "Security group ID attached to the EKS nodes"
  value       = module.eks.node_security_group_id
}

output "cluster_security_group_id" {
  description = "Security group ID attached to the EKS cluster"
  value       = module.eks.cluster_security_group_id
}

# =============================================================================
# ECR Outputs (optional)
# =============================================================================

output "ecr_repository_url" {
  description = "ECR repository URL (if created)"
  value       = var.create_ecr ? aws_ecr_repository.app[0].repository_url : null
}

# =============================================================================
# Resource Group Equivalent (for AKS parity)
# =============================================================================

output "resource_group_name" {
  description = "Resource group equivalent (returns cluster name for AWS)"
  value       = var.cluster_name
}

output "node_resource_group" {
  description = "Node resource group equivalent (returns cluster name for AWS)"
  value       = var.cluster_name
}
