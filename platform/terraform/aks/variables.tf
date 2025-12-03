# AKS Module Variables

variable "cluster_name" {
  description = "Name of the AKS cluster"
  type        = string
}

variable "resource_group_name" {
  description = "Name of the Azure resource group"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "eastus"
}

variable "dns_prefix" {
  description = "DNS prefix for the cluster"
  type        = string
  default     = ""
}

variable "kubernetes_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.29"
}

# Node pool configuration
variable "node_count" {
  description = "Initial number of nodes"
  type        = number
  default     = 3
}

variable "vm_size" {
  description = "VM size for nodes"
  type        = string
  default     = "Standard_D2s_v3"
}

variable "os_disk_size_gb" {
  description = "OS disk size in GB"
  type        = number
  default     = 50
}

variable "enable_auto_scaling" {
  description = "Enable cluster autoscaling"
  type        = bool
  default     = true
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

# Networking
variable "network_plugin" {
  description = "Network plugin (azure or kubenet)"
  type        = string
  default     = "azure"
}

variable "subnet_id" {
  description = "Subnet ID for nodes (optional)"
  type        = string
  default     = null
}

# Monitoring
variable "enable_monitoring" {
  description = "Enable Azure Monitor for containers"
  type        = bool
  default     = true
}

# Container Registry
variable "create_acr" {
  description = "Create an Azure Container Registry"
  type        = bool
  default     = false
}

# Tags
variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default = {
    environment = "lab"
    managed_by  = "terraform"
  }
}
