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

variable "private_subnets" {
  description = "Private subnets to deploy the service into"
  type        = list(string)
}

variable "public_subnets" {
  description = "Public subnets to deploy LB into"
  type        = list(string)
}

variable "vpc_id" {
  description = "VPC ID for the service"
}

variable "certificate_arn" {
  description = "The ARN of the certificate to attach to the listener"
  nullable    = true
  type        = string
  default     = null
}

variable "elb_logs_bucket_name" {
  description = "Where to store ELB logs"
  sensitive   = true
  nullable    = true
  type        = string
  default     = null
}

variable "service_config" {
  description = "Environment variables to pass to the service that are not sensitive"
  type = object({
    validation_service_grpc_address = string
    contracts_config                = string
  })
}

variable "service_secrets" {
  description = "Environment variables to pass to the service"
  sensitive   = true
  type = object({
    database_url             = string
    signer_private_key       = string
    app_chain_wss_url        = string
    settlement_chain_wss_url = string
  })
}

variable "enable_debug_logs" {
  description = "Enable debug logs for XMTPD server"
}

variable "desired_instance_count" {
  description = "Desired number of instances to run"
  default     = 2
}
