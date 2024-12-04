################################################################################
# VPC
################################################################################

output "vpc_id" {
  description = "The ID of the vpc"
  value       = module.vpc.vpc_id
}


#############################################
###############      API      ###############
#############################################

output "api_load_balancer_address" {
  description = "The full address for the API load balancer"
  value       = module.xmtpd_api.load_balancer_address
}

output "api_domain_name" {
  description = "The domain name for the API"
  value       = var.domain_name
}
