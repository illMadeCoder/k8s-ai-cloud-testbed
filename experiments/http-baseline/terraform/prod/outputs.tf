# HTTP Baseline Prod - Outputs

output "cluster_name" {
  description = "AKS cluster name"
  value       = module.aks.cluster_name
}

output "cluster_fqdn" {
  description = "AKS cluster FQDN"
  value       = module.aks.cluster_fqdn
}

output "resource_group" {
  description = "Resource group name"
  value       = module.aks.resource_group_name
}

output "kube_config_command" {
  description = "Command to get kubeconfig"
  value       = "az aks get-credentials --resource-group ${module.aks.resource_group_name} --name ${module.aks.cluster_name}"
}

output "acr_login_server" {
  description = "ACR login server (if created)"
  value       = module.aks.acr_login_server
}
