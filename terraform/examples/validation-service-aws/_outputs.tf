################################################################################
# VPC
################################################################################

output "vpc_ids" {
  description = "The IDs of the vpc"
  value       = [module.xmtp_node_us_east_2.vpc_id, module.xmtp_node_eu_north_1.vpc_id]
}

output "grpc_dns_address" {
  description = "Publicly accessible DNS records"
  value       = [cloudflare_record.grpc_gateway.hostname, cloudflare_record.grpc_gateway_2.hostname]
}

output "database_strings" {
  value     = [module.xmtp_node_us_east_2.database_string, module.xmtp_node_eu_north_1.database_string]
  sensitive = true
}
