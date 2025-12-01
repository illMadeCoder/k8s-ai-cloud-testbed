# HTTP Baseline Prod - Variables

variable "cluster_name" {
  description = "Name of the AKS cluster"
  type        = string
  default     = "http-baseline-prod"
}

variable "resource_group_name" {
  description = "Azure resource group name"
  type        = string
  default     = "http-baseline-prod-rg"
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "eastus"
}

variable "kubernetes_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.29"
}

variable "node_count" {
  description = "Initial node count"
  type        = number
  default     = 3
}

variable "vm_size" {
  description = "VM size for nodes"
  type        = string
  default     = "Standard_D4s_v3" # 4 vCPU, 16 GB - good for load testing
}

variable "min_nodes" {
  description = "Minimum nodes for autoscaling"
  type        = number
  default     = 2
}

variable "max_nodes" {
  description = "Maximum nodes for autoscaling"
  type        = number
  default     = 10
}

variable "create_acr" {
  description = "Create Azure Container Registry"
  type        = bool
  default     = false
}
