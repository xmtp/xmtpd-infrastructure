output "vpc_id" {
  description = "The ID of the vpc"
  value       = module.network.vpc_id
}

output "grpc_service_address" {
  description = "The full address for the XMTPD GRPC service"
  value       = module.xmtpd_server.grpc_service_address
}

output "grpc_service_dns_name" {
  value = module.xmtpd_server.grpc_service_dns_name
}

output "database_string" {
  value     = module.xmtpd_rds.writer_connection_string
  sensitive = true
}
