module "network" {
  source   = "../network"
  vpc_cidr = var.vpc_cidr

  providers = {
    aws = aws
  }
}

module "mls_validation_service" {
  source     = "../../../aws/xmtp-validation-service"
  depends_on = [module.network, aws_service_discovery_private_dns_namespace.xmtp]

  env                              = terraform.workspace
  cluster_id                       = aws_ecs_cluster.this.id
  vpc_id                           = module.network.vpc_id
  private_subnets                  = module.network.private_subnets
  allowed_ingress_cidr_blocks      = [for k, v in module.network.azs : cidrsubnet(module.network.vpc_cidr, 4, k)]
  docker_image                     = "ghcr.io/xmtp/mls-validation-service:latest"
  service_discovery_namespace_name = aws_service_discovery_private_dns_namespace.xmtp.name
  chain_rpc_urls                   = var.verifier_chain_rpc_urls
  providers = {
    aws = aws
  }
}
