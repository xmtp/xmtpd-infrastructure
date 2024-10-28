variable "name" {
  description = "The service name"
  type        = string
}

variable "cpu" {
  description = "Number of CPU units"
  type        = number
}

variable "memory" {
  description = "Memory MB"
  type        = number
}

variable "env_vars" {
  description = "Key/value map of environment variables for the task"
  type        = map(string)
  default     = {}
}

variable "secrets" {
  description = "Key/value map of the secrets"
  sensitive   = true
  default     = {}
  validation {
    condition     = length([for k, v in var.secrets : k if v == ""]) == 0
    error_message = "Values for secrets cannot be empty. To pass an empty string, set it as an env var."
  }
}

variable "ports" {
  description = "The ports to open on the container"
  type        = list(number)
  default     = []
}

variable "image" {
  description = "The Docker image to pull"
}

variable "docker_labels" {
  description = "Key value map of docker labels"
  type        = map(string)
  default     = {}
}

variable "command" {
  description = "The command to run"
  default     = null
  type        = list(string)
}

variable "datadog_api_key" {
  description = "DataDog API Key"
  sensitive   = true
}

variable "env" {
  description = "Environment name"
}

variable "dockerhub_username" {
  default = ""
}

// These are account access tokens found in https://hub.docker.com/settings/security
// This should not be the account password.
variable "dockerhub_token" {
  sensitive = true
  default   = ""
}

variable "additional_task_role_statements" {
  type = list(object({
    Effect   = string
    Action   = list(string)
    Resource = list(string)
  }))
  default = []
}

variable "efs_mounts" {
  description = "EFS mount"
  type = list(object({
    name           = string
    file_system_id = string
    root_directory = string
  }))
  default = []
}

variable "health_check_config" {
  description = "ECS task definition health check config"
  nullable    = true
  type = object({
    command     = list(string)
    interval    = optional(number, 30)
    timeout     = optional(number, 5)
    retries     = optional(number, 3)
    startPeriod = optional(number, 10)
  })
  default = null
}
