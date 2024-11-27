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

variable "xmtpd_docker_image" {
  description = "Docker image for xmtpd"
  default     = "ghcr.io/xmtp/xmtpd:main"
}

variable "chain_id" {
  description = "The chain ID of the XMTP chain"
  default     = "241320161"
}

variable "nodes_contract_address" {
  description = "The address of the nodes contract"
  type        = string
}

variable "messages_contract_address" {
  description = "The address of the messages contract"
  type        = string
}

variable "identity_updates_contract_address" {
  description = "The address of the identity updates contract"
  type        = string
}

variable "chain_rpc_url" {
  description = "The RPC URL to connect to the XMTP chain"
  sensitive   = true
  type        = string
}

variable "signer_private_key" {
  description = "The private key of the node's signer"
  sensitive   = true
  type        = string
}
