# Spacelift Root Stack Outputs

output "azure_foundation_stack_id" {
  description = "Spacelift stack ID for Azure foundation"
  value       = spacelift_stack.azure_foundation.id
}

output "aws_foundation_stack_id" {
  description = "Spacelift stack ID for AWS foundation"
  value       = spacelift_stack.aws_foundation.id
}

output "experiment_azure_stacks" {
  description = "Map of Azure experiment stack IDs"
  value = {
    for name, stack in spacelift_stack.experiment_azure :
    name => {
      id   = stack.id
      name = stack.name
    }
  }
}

output "experiment_aws_stacks" {
  description = "Map of AWS experiment stack IDs"
  value = {
    for name, stack in spacelift_stack.experiment_aws :
    name => {
      id   = stack.id
      name = stack.name
    }
  }
}

output "all_experiment_stacks" {
  description = "Combined list of all experiment stacks"
  value = concat(
    [for name, stack in spacelift_stack.experiment_azure : stack.name],
    [for name, stack in spacelift_stack.experiment_aws : stack.name]
  )
}
