module "xmtp_node_us_east_2" {
  source = "./xmtp_region_node"

  vpc_cidr = "10.1.0.0/16"

  mls_validation_service_docker_image = var.mls_validation_service_docker_image

  verifier_chain_rpc_urls = var.verifier_chain_rpc_urls

  providers = {
    aws = aws
  }
}

module "xmtp_node_eu_north_1" {
  source = "./xmtp_region_node"

  vpc_cidr = "10.2.0.0/16"

  mls_validation_service_docker_image = var.mls_validation_service_docker_image

  verifier_chain_rpc_urls = var.verifier_chain_rpc_urls

  providers = {
    aws = aws.eu-north-1
  }
}
