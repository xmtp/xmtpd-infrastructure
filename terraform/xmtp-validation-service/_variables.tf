variable "vpc_id" {
  description = "The VPC ID"
}

variable "cluster_id" {
  description = "The ID of the cluster to deploy into"
}

variable "private_subnets" {
  description = "The (private) subnets used to run the service"
}

variable "datadog_api_key" {
  description = "API Key for DataDog agent"
  sensitive   = true
}

variable "cpu" {
  description = "Available CPU for the node container"
  default     = 2048
}

variable "memory" {
  description = "Available memory for the node container"
  default     = 4096
}

variable "env" {
  description = "Environment"
  default     = "testnet"
}

variable "docker_image" {
  description = "Docker image for service"
}

variable "service_discovery_namespace_name" {
  description = "The name of the service discovery namespace"
}

variable "allowed_ingress_cidr_blocks" {
  description = "The CIDR blocks allowed to connect to the service"
  type        = list(string)
}

variable "chain_rpc_urls" {
  description = "RPC URLs for the smart contract verifier"
  sensitive   = true
  type = object({
    chain_rpc_1     = string
    chain_rpc_8453  = string
    chain_rpc_42161 = string
    chain_rpc_10    = string
    chain_rpc_137   = string
    chain_rpc_324   = string
    chain_rpc_59144 = string
  })
}
