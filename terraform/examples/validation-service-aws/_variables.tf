variable "xmtpd_server_docker_image" {
  description = "Docker image for xmtpd server"
  default     = "ghcr.io/xmtp/xmtpd:main"
}

variable "signer_private_key_ohio" {
  description = "Private key used to sign messages in Ohio"
  sensitive   = true
}

variable "signer_private_key_sweden" {
  description = "Private key used to sign messages in Sweden"
  sensitive   = true
}

variable "cloudflare_api_token" {
  type      = string
  sensitive = true
}

variable "nodes_domain" {
  description = "Domain that will house the nodes subdomain"
  default     = "xmtp.network"
}

variable "mls_validation_service_docker_image" {
  description = "Docker image for mls validation service"
  default     = "ghcr.io/xmtp/mls-validation-service:main"
}

variable "chain_rpc_url" {
  description = "RPC URL for the blockchain"
  type        = string
  sensitive   = true
}

variable "contracts" {
  description = "Contracts already deployed on the testnet"
  type = object({
    chain_id                          = string
    nodes_contract_address            = string
    messages_contract_address         = string
    identity_updates_contract_address = string
  })
}

variable "enable_debug_logs" {
  description = "Enable debug logs for XMTPD server"
  default     = false
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
