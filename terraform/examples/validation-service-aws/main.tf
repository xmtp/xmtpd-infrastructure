module "datadog_iam" {
  source = "./modules/datadog_iam"

  providers = {
    aws     = aws
    datadog = datadog
  }
}

module "xmtp_node_us_east_2" {
  source = "./modules/xmtp_region_node"

  vpc_cidr        = "10.1.0.0/16"
  datadog_api_key = var.datadog_api_key

  xmtpd_server_docker_image           = var.xmtpd_server_docker_image
  mls_validation_service_docker_image = var.mls_validation_service_docker_image
  enable_debug_logs                   = var.enable_debug_logs

  signer_private_key = var.signer_private_key_ohio
  chain_rpc_url      = var.chain_rpc_url

  contracts = var.contracts

  certificate_arn = module.public_lb_cert.certificate_arn

  verifier_chain_rpc_urls = var.verifier_chain_rpc_urls

  providers = {
    aws = aws
  }
}

module "xmtp_node_eu_north_1" {
  source = "./modules/xmtp_region_node"

  vpc_cidr        = "10.2.0.0/16"
  datadog_api_key = var.datadog_api_key

  xmtpd_server_docker_image           = var.xmtpd_server_docker_image
  mls_validation_service_docker_image = var.mls_validation_service_docker_image
  enable_debug_logs                   = var.enable_debug_logs

  signer_private_key = var.signer_private_key_sweden
  chain_rpc_url      = var.chain_rpc_url

  contracts = var.contracts

  certificate_arn = aws_acm_certificate.north.arn

  verifier_chain_rpc_urls = var.verifier_chain_rpc_urls

  providers = {
    aws = aws.eu-north-1
  }
}
