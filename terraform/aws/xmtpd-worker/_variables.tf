variable "cpu" {
  description = "Available CPU for the replication server container"
  default     = 1024
}

variable "memory" {
  description = "Available memory for the replication server container"
  default     = 2048
}

variable "docker_image" {
  description = "xmtpd docker image"
}

variable "cluster_id" {
  description = "The ID of the ECS cluster"
}

variable "public_subnets" {
  description = "Public subnets to deploy LB into"
  type        = list(string)
}

variable "vpc_id" {
  description = "VPC ID for the service"
}

variable "service_config" {
  description = "Environment variables to pass to the service that are not sensitive"
  type = object({
    validation_service_grpc_address   = string
    chain_id                          = string
    nodes_contract_address            = string
    messages_contract_address         = string
    identity_updates_contract_address = string
  })
}

variable "service_secrets" {
  description = "Environment variables to pass to the service"
  sensitive   = true
  type = object({
    database_url       = string
    signer_private_key = string
    chain_rpc_url      = string
  })
}

variable "enable_debug_logs" {
  description = "Enable debug logs for XMTPD server"
}
