# Spacelift Root Stack Variables

variable "experiments" {
  description = "Map of experiment configurations to create Spacelift stacks for"
  type = map(object({
    clouds     = list(string) # ["azure"], ["aws"], or ["both"]
    autodeploy = bool         # Auto-deploy on git push
    clusters = optional(map(object({
      provider   = string
      vm_size    = optional(string)
      node_count = optional(number)
    })), {})
  }))

  default = {
    # Example: http-baseline experiment on Azure only
    http-baseline = {
      clouds     = ["azure"]
      autodeploy = false
      clusters = {
        target = {
          provider   = "azure"
          vm_size    = "Standard_D4s_v3"
          node_count = 3
        }
        loadgen = {
          provider   = "azure"
          vm_size    = "Standard_D2s_v3"
          node_count = 2
        }
      }
    }

    # Example: multi-cloud-demo on both Azure and AWS
    multi-cloud-demo = {
      clouds     = ["both"]
      autodeploy = false
      clusters = {
        target-azure = {
          provider   = "azure"
          vm_size    = "Standard_D4s_v3"
          node_count = 3
        }
        target-aws = {
          provider   = "aws"
          vm_size    = "t3.large"
          node_count = 3
        }
      }
    }
  }

  validation {
    condition = alltrue([
      for name, config in var.experiments :
      alltrue([
        for cloud in config.clouds :
        contains(["azure", "aws", "both"], cloud)
      ])
    ])
    error_message = "Clouds must be 'azure', 'aws', or 'both'."
  }
}
