module "network" {
  source   = "../network"
  vpc_cidr = var.vpc_cidr

  providers = {
    aws = aws
  }
}

module "xmtpd_rds" {
  source     = "../rds_cluster"
  depends_on = [module.network]

  availability_zones = module.network.azs
  vpc_id             = module.network.vpc_id
  subnet_ids         = module.network.private_subnets
  instance_class     = "db.t4g.medium"
  num_instances      = 2
  database_name      = "xmtpddata"
  tags = {
    name = "xmtpd"
  }

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

module "xmtpd_server" {
  source     = "../xmtpd_server"
  depends_on = [module.network, module.xmtpd_rds]

  env                  = terraform.workspace
  cluster_id           = aws_ecs_cluster.this.id
  vpc_id               = module.network.vpc_id
  private_subnets      = module.network.private_subnets
  public_subnets       = module.network.public_subnets
  docker_image         = var.xmtpd_server_docker_image
  datadog_api_key      = var.datadog_api_key
  elb_logs_bucket_name = module.datadog_forwarder.bucket_name
  enable_debug_logs    = var.enable_debug_logs

  service_config = {
    validation_service_grpc_address   = module.mls_validation_service.grpc_service_address
    chain_id                          = var.contracts.chain_id
    nodes_contract_address            = var.contracts.nodes_contract_address
    messages_contract_address         = var.contracts.messages_contract_address
    identity_updates_contract_address = var.contracts.identity_updates_contract_address
  }
  service_secrets = {
    database_url       = module.xmtpd_rds.writer_connection_string
    signer_private_key = var.signer_private_key
    chain_rpc_url      = var.chain_rpc_url
  }

  certificate_arn = var.certificate_arn

  providers = {
    aws = aws
  }
}
