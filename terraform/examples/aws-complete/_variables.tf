variable "mls_validation_service_docker_image" {
  description = "Docker image for mls validation service"
  default     = "ghcr.io/xmtp/mls-validation-service:latest"
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

variable "xmtpd_docker_image" {
  description = "Docker image for xmtpd"
  default     = "ghcr.io/xmtp/xmtpd:latest"
}

variable "xmtpd_prune_docker_image" {
  description = "Docker image for xmtpd prune"
  default     = "ghcr.io/xmtp/xmtpd-prune:latest"
}

variable "app_chain_rpc_url" {
  description = "RPC URL for the app blockchain"
  type        = string
  sensitive   = true
}

variable "settlement_chain_rpc_url" {
  description = "RPC URL for the settlement blockchain"
  type        = string
  sensitive   = true
}


variable "app_chain_wss_url" {
  description = "WSS URL for the app blockchain"
  type        = string
  sensitive   = true
}

variable "settlement_chain_wss_url" {
  description = "WSS URL for the settlement blockchain"
  type        = string
  sensitive   = true
}

variable "contracts" {
  description = "JSON Contracts already deployed on the testnet"
  type        = string
}

variable "signer_private_key" {
  description = "The private key of the node's signer"
  sensitive   = true
  type        = string
}

variable "domain_name" {
  description = "The domain name to use for public endpoints. This must have already been registered through AWS Route53 Domains"
  type        = string
}
