variable "cpu" {
  description = "Available CPU for the replication server container"
  default     = 1024
}

variable "memory" {
  description = "Available memory for the replication server container"
  default     = 2048
}

variable "docker_image" {
  description = "Docker image for the service"
}

variable "service_config" {
  description = "Environment variables to pass to the service that are not sensitive"
  type = object({
    contracts_config = string
  })

}

variable "service_secrets" {
  description = "Environment variables to pass to the service"
  sensitive   = true
  type = object({
    database_url       = string
    signer_private_key = string
  })
}

variable "enable_debug_logs" {
  description = "Enable debug logs for pruner"
}

variable "cluster_id" {
  description = "The ID of the cluster to deploy into"
}

variable "private_subnets" {
  description = "Private subnets to deploy service into"
  type        = list(string)
}

variable "vpc_id" {
  description = "VPC ID for the service"
}