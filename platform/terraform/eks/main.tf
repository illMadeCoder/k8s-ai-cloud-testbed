# AWS Elastic Kubernetes Service (EKS) Module
# Provisions a managed Kubernetes cluster on AWS
# Interface mirrors AKS module for consistency

terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.region

  default_tags {
    tags = var.tags
  }
}

# =============================================================================
# Data Sources
# =============================================================================

data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_caller_identity" "current" {}

# =============================================================================
# VPC (optional - creates new if vpc_id not provided)
# =============================================================================

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  count = var.vpc_id == null ? 1 : 0

  name = "${var.cluster_name}-vpc"
  cidr = var.vpc_cidr

  azs             = slice(data.aws_availability_zones.available.names, 0, 3)
  private_subnets = var.private_subnet_cidrs
  public_subnets  = var.public_subnet_cidrs

  enable_nat_gateway   = true
  single_nat_gateway   = var.single_nat_gateway
  enable_dns_hostnames = true
  enable_dns_support   = true

  # Tags required for EKS
  public_subnet_tags = {
    "kubernetes.io/role/elb"                    = 1
    "kubernetes.io/cluster/${var.cluster_name}" = "shared"
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb"           = 1
    "kubernetes.io/cluster/${var.cluster_name}" = "shared"
  }

  tags = var.tags
}

# =============================================================================
# EKS Cluster
# =============================================================================

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  cluster_name    = var.cluster_name
  cluster_version = var.kubernetes_version

  # Networking
  vpc_id     = var.vpc_id != null ? var.vpc_id : module.vpc[0].vpc_id
  subnet_ids = var.subnet_ids != null ? var.subnet_ids : module.vpc[0].private_subnets

  # Cluster access
  cluster_endpoint_public_access  = true
  cluster_endpoint_private_access = true

  # Enable IRSA (required for Crossplane Workload Identity)
  enable_irsa = true

  # Cluster addons
  cluster_addons = {
    coredns = {
      most_recent = true
    }
    kube-proxy = {
      most_recent = true
    }
    vpc-cni = {
      most_recent = true
    }
    aws-ebs-csi-driver = {
      most_recent = true
    }
  }

  # Managed node groups
  eks_managed_node_groups = {
    default = {
      name           = "default"
      instance_types = [var.instance_type]

      min_size     = var.min_nodes
      max_size     = var.max_nodes
      desired_size = var.node_count

      disk_size = var.disk_size_gb

      # Use latest Amazon Linux 2 EKS-optimized AMI
      ami_type = "AL2_x86_64"

      labels = {
        role = "general"
      }

      tags = var.tags
    }
  }

  # Allow access from Spacelift runners
  cluster_security_group_additional_rules = {
    ingress_spacelift = {
      description = "Allow access from Spacelift runners"
      protocol    = "tcp"
      from_port   = 443
      to_port     = 443
      type        = "ingress"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  tags = var.tags
}

# =============================================================================
# IRSA for Crossplane AWS Provider
# =============================================================================

module "crossplane_irsa" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  count = var.enable_crossplane_irsa ? 1 : 0

  role_name = "${var.cluster_name}-crossplane"

  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["crossplane-system:provider-aws-*"]
    }
  }

  # Crossplane needs broad permissions to manage AWS resources
  # In production, scope this down to specific services
  role_policy_arns = {
    admin = var.crossplane_policy_arn != null ? var.crossplane_policy_arn : "arn:aws:iam::aws:policy/AdministratorAccess"
  }

  tags = var.tags
}

# =============================================================================
# CloudWatch Log Group (for cluster logging)
# =============================================================================

resource "aws_cloudwatch_log_group" "eks" {
  count = var.enable_monitoring ? 1 : 0

  name              = "/aws/eks/${var.cluster_name}/cluster"
  retention_in_days = 30

  tags = var.tags
}

# =============================================================================
# ECR Repository (optional - for container images)
# =============================================================================

resource "aws_ecr_repository" "app" {
  count = var.create_ecr ? 1 : 0

  name                 = var.cluster_name
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = var.tags
}
