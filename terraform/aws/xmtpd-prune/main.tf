locals {
  prune_container_name = "xmtpd-prune"

  xmtp_env_vars = {
    "GOLOG_LOG_FMT"               = "json"
    "XMTPD_LOG_ENCODING"          = "json"
    "ENV"                         = var.env
    "XMTPD_CONTRACTS_CONFIG_JSON" = var.service_config.contracts_config
  }

  xmtp_secrets = {
    "XMTPD_DB_WRITER_CONNECTION_STRING" = var.service_secrets.database_url
    "XMTPD_SIGNER_PRIVATE_KEY"          = var.service_secrets.signer_private_key
  }

  prune_command = var.enable_debug_logs ? ["--log.log-level=debug"] : []
}


module "task_definition_prune" {
  name   = local.prune_container_name
  source = "../fargate-task-definition"
  cpu    = var.cpu
  memory = var.memory

  env_vars = local.xmtp_env_vars
  secrets  = local.xmtp_secrets

  image           = var.docker_image

  command = local.prune_command

  providers = {
    aws = aws
  }
}

resource "aws_cloudwatch_event_rule" "hourly_schedule" {
  name_prefix         = "run-hourly-ecs-task"
  schedule_expression = "rate(1 hour)"
}

resource "aws_cloudwatch_event_target" "ecs_target" {
  rule     = aws_cloudwatch_event_rule.hourly_schedule.name
  role_arn = aws_iam_role.eventbridge_invoke_ecs.arn
  arn      = var.cluster_id

  ecs_target {
    task_count          = 1
    task_definition_arn = module.task_definition_prune.task_definition_arn
    launch_type         = "FARGATE"

    network_configuration {
      subnets         = var.private_subnets
      security_groups = [aws_security_group.ecs_service.id]
    }

    platform_version = "LATEST"
  }
}