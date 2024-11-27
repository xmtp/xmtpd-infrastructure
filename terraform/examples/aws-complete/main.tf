
module "mls_validation_service" {
  source     = "./aws/xmtp-validation-service" # TODO: Replace with git URL once merged to main
  depends_on = [module.vpc, aws_service_discovery_private_dns_namespace.xmtp]

  env                              = terraform.workspace
  cluster_id                       = aws_ecs_cluster.this.id
  vpc_id                           = module.vpc.vpc_id
  private_subnets                  = module.vpc.private_subnets
  allowed_ingress_cidr_blocks      = [for k, v in local.azs : cidrsubnet(local.vpc_cidr, 4, k)]
  docker_image                     = var.mls_validation_service_docker_image
  service_discovery_namespace_name = aws_service_discovery_private_dns_namespace.xmtp.name
  chain_rpc_urls                   = var.verifier_chain_rpc_urls

  providers = {
    aws = aws
  }
}

module "xmtpd_api" {
  source = "./aws/xmtpd-api" # TODO: Replace with git URL once merged to main

  vpc_id          = module.vpc.vpc_id
  public_subnets  = module.vpc.public_subnets
  private_subnets = module.vpc.private_subnets
  docker_image    = var.xmtpd_docker_image
  cluster_id      = aws_ecs_cluster.this.id

  service_config = {
    validation_service_grpc_address   = module.mls_validation_service.grpc_service_address
    chain_id                          = var.chain_id
    nodes_contract_address            = var.nodes_contract_address
    messages_contract_address         = var.messages_contract_address
    identity_updates_contract_address = var.identity_updates_contract_address
  }
  service_secrets = {
    signer_private_key = var.signer_private_key
    chain_rpc_url      = var.chain_rpc_url
    database_url       = "CHANGE_ME" # TODO:nm add database
  }
  enable_debug_logs = false

  providers = {
    aws = aws
  }
}

module "xmtpd_worker" {
  source = "./aws/xmtpd-worker" # TODO: Replace with git URL once merged to main

  vpc_id         = module.vpc.vpc_id
  public_subnets = module.vpc.public_subnets
  docker_image   = var.xmtpd_docker_image
  cluster_id     = aws_ecs_cluster.this.id
  service_config = {
    validation_service_grpc_address   = module.mls_validation_service.grpc_service_address
    chain_id                          = var.chain_id
    nodes_contract_address            = var.nodes_contract_address
    messages_contract_address         = var.messages_contract_address
    identity_updates_contract_address = var.identity_updates_contract_address
  }
  service_secrets = {
    signer_private_key = var.signer_private_key
    chain_rpc_url      = var.chain_rpc_url
    database_url       = "CHANGE_ME" # TODO:nm add database
  }
  enable_debug_logs = false

  providers = {
    aws = aws
  }
}
