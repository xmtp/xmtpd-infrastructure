################################################################################
# VPC
################################################################################

output "vpc_ids" {
  description = "The IDs of the vpc"
  value       = [module.xmtp_node_us_east_2.vpc_id, module.xmtp_node_eu_north_1.vpc_id]
}
