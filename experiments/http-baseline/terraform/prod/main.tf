# HTTP Baseline Experiment - Production Environment
# Deploys to Azure AKS for production load testing

terraform {
  required_version = ">= 1.0.0"

  # Uncomment to use remote state
  # backend "azurerm" {
  #   resource_group_name  = "tfstate"
  #   storage_account_name = "tfstate"
  #   container_name       = "tfstate"
  #   key                  = "http-baseline/prod.tfstate"
  # }
}

module "aks" {
  source = "../../../../terraform-modules/aks"

  cluster_name        = var.cluster_name
  resource_group_name = var.resource_group_name
  location            = var.location
  kubernetes_version  = var.kubernetes_version

  # Node pool sizing for load testing
  node_count          = var.node_count
  vm_size             = var.vm_size
  enable_auto_scaling = true
  min_nodes           = var.min_nodes
  max_nodes           = var.max_nodes

  # Monitoring
  enable_monitoring = true

  # Container registry for custom images
  create_acr = var.create_acr

  tags = {
    environment = "prod"
    experiment  = "http-baseline"
    managed_by  = "terraform"
  }
}

# Output kubeconfig to file for ArgoCD bootstrap
resource "local_file" "kubeconfig" {
  content         = module.aks.kube_config
  filename        = "${path.module}/kubeconfig"
  file_permission = "0600"
}
