locals {
  service_port = 5050
  public_port  = 443
  api_name     = "xmtpd-api"

  // empty for now, all variables are provided via env
  xmtp_node_command = concat([
    # Turn on the metrics server (also required for health checks)
    "--metrics.enable",
    # Expose the metrics server to the DataDog docker container
    "--metrics.metrics-address=0.0.0.0",
    # Enable gRPC server reflection
    "--reflection.enable",
    # Enable the replication API
    "--replication.enable",
    # Make sure the port is set correctly
    "--api.port=${local.service_port}"
    ],
    var.enable_debug_logs ? ["--log.log-level=debug"] : [],
  )
}

#################################
#########     API     ###########
#################################

module "api_task_definition" {
  source = "../fargate-task-definition"
  name   = local.api_name
  cpu    = var.cpu
  memory = var.memory

  env_vars = {
    "GOLOG_LOG_FMT"                            = "json"
    "XMTPD_MLS_VALIDATION_GRPC_ADDRESS"        = var.service_config.validation_service_grpc_address
    "XMTPD_CONTRACTS_CHAIN_ID"                 = var.service_config.chain_id
    "XMTPD_CONTRACTS_NODES_ADDRESS"            = var.service_config.nodes_contract_address
    "XMTPD_CONTRACTS_MESSAGES_ADDRESS"         = var.service_config.messages_contract_address
    "XMTPD_CONTRACTS_IDENTITY_UPDATES_ADDRESS" = var.service_config.identity_updates_contract_address
  }

  secrets = {
    "XMTPD_DB_WRITER_CONNECTION_STRING" = var.service_secrets.database_url
    "XMTPD_SIGNER_PRIVATE_KEY"          = var.service_secrets.signer_private_key
    "XMTPD_PAYER_PRIVATE_KEY"           = var.service_secrets.signer_private_key
    "XMTPD_CONTRACTS_RPC_URL"           = var.service_secrets.chain_rpc_url
  }

  ports = [local.service_port]
  image = var.docker_image
  env   = var.env

  command = local.xmtp_node_command

  providers = {
    aws = aws
  }
}

resource "aws_ecs_service" "api" {
  name                               = local.api_name
  cluster                            = var.cluster_id
  task_definition                    = module.api_task_definition.task_definition_arn
  enable_execute_command             = false
  desired_count                      = var.desired_instance_count
  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 100
  wait_for_steady_state              = true

  network_configuration {
    subnets         = var.private_subnets
    security_groups = [aws_security_group.ecs_service.id]
  }

  capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 100
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.public.arn
    container_name   = local.api_name
    container_port   = local.service_port
  }
}
