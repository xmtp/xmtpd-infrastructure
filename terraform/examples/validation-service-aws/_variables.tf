variable "mls_validation_service_docker_image" {
  description = "Docker image for mls validation service"
  default     = "ghcr.io/xmtp/mls-validation-service:main"
}

variable "verifier_chain_rpc_urls" {
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
