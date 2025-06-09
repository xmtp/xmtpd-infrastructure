locals {
  name         = "xmtpd-worker"
  metrics_port = 8008

  // empty for now, all variables are provided via env
  xmtp_node_command = concat([
    # Turn on the metrics server (also required for health checks)
    "--metrics.enable",
    # Expose the metrics server to the DataDog docker container
    "--metrics.metrics-address=0.0.0.0",
    # Expose the metrics server on the expected port
    "--metrics.metrics-port=${local.metrics_port}",
    # Enable the indexer 
    "--indexer.enable",
    # Enable the sync service
    "--sync.enable"
    ],
    var.enable_debug_logs ? ["--log.log-level=debug"] : [],
  )
}


module "task_definition" {
  source = "../fargate-task-definition"
  name   = local.name
  cpu    = var.cpu
  memory = var.memory

  env_vars = {
    "GOLOG_LOG_FMT"                            = "json"
    "XMTPD_MLS_VALIDATION_GRPC_ADDRESS"        = var.service_config.validation_service_grpc_address
    "XMTPD_CONTRACTS_CONFIG_JSON"       = var.service_config.contracts_config
  }

  secrets = {
    "XMTPD_DB_WRITER_CONNECTION_STRING" = var.service_secrets.database_url
    "XMTPD_SIGNER_PRIVATE_KEY"          = var.service_secrets.signer_private_key
    "XMTPD_APP_CHAIN_WSS_URL"           = var.service_secrets.app_chain_wss_url
    "XMTPD_SETTLEMENT_CHAIN_WSS_URL"    = var.service_secrets.settlement_chain_wss_url
  }

  ports = []
  image = var.docker_image

  command = local.xmtp_node_command

  health_check_config = {
    command = ["CMD-SHELL", "curl -f http://localhost:${local.metrics_port}"]
  }

  providers = {
    aws = aws
  }
}

resource "aws_ecs_service" "worker" {
  name                               = local.name
  cluster                            = var.cluster_id
  task_definition                    = module.task_definition.task_definition_arn
  enable_execute_command             = false
  desired_count                      = 1 # Set the worker to run on a single instance except during deployments
  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 100
  wait_for_steady_state              = true

  network_configuration {
    subnets          = var.public_subnets # To avoid the NAT gateway we deploy the worker into the public subnets. This increases available bandwidth and reduces costs.
    security_groups  = [aws_security_group.ecs_service.id]
    assign_public_ip = true
  }

  capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 100
  }
}
