# EKS Module Variables
# Interface mirrors AKS module for consistency

variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
}

variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "kubernetes_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.29"
}

# =============================================================================
# Node Configuration
# =============================================================================

variable "node_count" {
  description = "Desired number of nodes"
  type        = number
  default     = 3
}

variable "instance_type" {
  description = "EC2 instance type for nodes"
  type        = string
  default     = "t3.large"
}

variable "disk_size_gb" {
  description = "Disk size in GB for nodes"
  type        = number
  default     = 50
}

variable "min_nodes" {
  description = "Minimum nodes for autoscaling"
  type        = number
  default     = 1
}

variable "max_nodes" {
  description = "Maximum nodes for autoscaling"
  type        = number
  default     = 10
}

# =============================================================================
# Networking
# =============================================================================

variable "vpc_id" {
  description = "Existing VPC ID (creates new VPC if null)"
  type        = string
  default     = null
}

variable "subnet_ids" {
  description = "Existing subnet IDs (creates new subnets if null)"
  type        = list(string)
  default     = null
}

variable "vpc_cidr" {
  description = "VPC CIDR block (if creating new VPC)"
  type        = string
  default     = "10.0.0.0/16"
}

variable "private_subnet_cidrs" {
  description = "Private subnet CIDR blocks"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
}

variable "public_subnet_cidrs" {
  description = "Public subnet CIDR blocks"
  type        = list(string)
  default     = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
}

variable "single_nat_gateway" {
  description = "Use single NAT gateway (cost saving for non-prod)"
  type        = bool
  default     = true
}

# =============================================================================
# Features
# =============================================================================

variable "enable_monitoring" {
  description = "Enable CloudWatch monitoring"
  type        = bool
  default     = true
}

variable "enable_crossplane_irsa" {
  description = "Create IRSA role for Crossplane AWS provider"
  type        = bool
  default     = true
}

variable "crossplane_policy_arn" {
  description = "IAM policy ARN for Crossplane (defaults to AdministratorAccess)"
  type        = string
  default     = null
}

variable "create_ecr" {
  description = "Create an ECR repository for container images"
  type        = bool
  default     = false
}

# =============================================================================
# Tags
# =============================================================================

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default = {
    environment = "lab"
    managed_by  = "terraform"
  }
}
