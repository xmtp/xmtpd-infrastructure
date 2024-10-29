output "grpc_service_address" {
  description = "The full address for the GRPC service"
  value       = "http://${local.name}.${var.service_discovery_namespace_name}:${local.service_port}"
}