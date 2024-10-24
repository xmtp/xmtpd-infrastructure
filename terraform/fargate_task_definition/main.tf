locals {
  # Convert env vars from plain map to the format ECS expects
  env_vars = [for k, v in var.env_vars : {
    name  = k
    value = v
  }]

  # Create port mappings. Does assume all ports are TCP. Will need to revise if we want UDP support
  port_mappings = [for port in var.ports : {
    containerPort = port
    hostPort      = port
    protocol      = "tcp"
  }]

  # Secrets is a sensitive input. Have to do this hack to use in a for_each
  # https://www.terraform.io/language/meta-arguments/for_each#limitations-on-values-used-in-for_each
  secret_keys = nonsensitive(toset([for k, v in var.secrets : k]))
  efs_mounts  = toset(var.efs_mounts)

  labels = merge(tomap({
    "com.datadoghq.tags.service" = var.name,
    "com.datadoghq.tags.env" = var.env, }),
    var.docker_labels,
  )
}

resource "aws_ecs_task_definition" "task" {
  family                   = var.name
  requires_compatibilities = ["FARGATE"]
  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "X86_64"
  }
  # Required for Fargate
  network_mode       = "awsvpc"
  cpu                = var.cpu
  memory             = var.memory
  execution_role_arn = aws_iam_role.execution_role.arn
  task_role_arn      = aws_iam_role.task_role.arn

  depends_on = [
    aws_secretsmanager_secret_version.secrets,
    aws_secretsmanager_secret_version.dockerhub_credentials
  ]

  dynamic "volume" {
    for_each = local.efs_mounts
    iterator = mount
    content {
      name = mount.value["name"]
      efs_volume_configuration {
        file_system_id = mount.value["file_system_id"]
      }
    }
  }

  container_definitions = jsonencode([
    merge(
      {
        name         = var.name
        image        = var.image
        essential    = true
        dockerLabels = local.labels
        command      = var.command
        logConfiguration = {
          # https://docs.datadoghq.com/integrations/ecs_fargate/?tab=fluentbitandfirelens
          logDriver = "awsfirelens"
          options = {
            dd_message_key = "log"
            apikey         = var.datadog_api_key
            provider       = "ecs"
            dd_service     = var.name
            dd_source      = var.name
            host           = "http-intake.logs.datadoghq.com"
            dd_tags        = "env:${var.env},infra_provider:aws"
            TLS            = "on"
            Name           = "datadog"
          }
        }
        environment = local.env_vars
        # Create a mapping of secret names to the ARN of the secret manager secrets
        secrets = [for name, secret in aws_secretsmanager_secret.secrets : {
          name      = name
          valueFrom = secret.arn
        }]
        dependsOn = [
          {
            containerName = "datadog-agent"
            condition     = "START"
          },
          {
            containerName = "log_router"
            condition     = "START"
          }
        ]
        portMappings = local.port_mappings
        # Default values that must be set to avoid re-creation
        cpu = 0
        mountPoints = length(var.efs_mounts) > 0 ? [{
          containerPath = var.efs_mounts[0].root_directory
          sourceVolume  = var.efs_mounts[0].name
          readOnly      = false
        }] : []
        volumesFrom = []
        ulimits = [
          {
            name      = "nofile"
            softLimit = 65535
            hardLimit = 65535
          }
        ]
      },
      var.dockerhub_username == "" || var.dockerhub_token == "" ? {} : {
        repositoryCredentials = {
          credentialsParameter = one(aws_secretsmanager_secret.dockerhub_credentials[*]).arn
        }
      },
      var.health_check_config != null ? {
        healthCheck = var.health_check_config
      } : {}
    ),
    # This comes basically direct from DataDog's support
    # https://docs.datadoghq.com/integrations/ecs_fargate/?tab=fluentbitandfirelens
    {
      name  = "log_router"
      image = "public.ecr.aws/aws-observability/aws-for-fluent-bit:stable"
      fireLensConfiguration = {
        type = "fluentbit",
        options = {
          enable-ecs-log-metadata = "true"
          config-file-type        = "file"
          config-file-value       = "/fluent-bit/configs/parse-json.conf"
        }
      }
      logConfiguration = null
      # Default values that must be set to avoid re-creation
      cpu          = 0
      environment  = []
      essential    = true
      mountPoints  = []
      portMappings = []
      user         = "0"
      volumesFrom  = []
    },
    {
      name      = "datadog-agent"
      image     = "public.ecr.aws/datadog/agent:latest"
      essential = true
      portMappings = [
        {
          containerPort = 8125
          hostPort      = 8125
          protocol      = "udp"
        },
        {
          containerPort = 8126
          hostPort      = 8126
          protocol      = "tcp"
        }
      ]
      environment = [
        {
          name  = "DD_API_KEY"
          value = var.datadog_api_key
        },
        {
          name  = "DD_APM_ENABLED"
          value = "true"
        },
        {
          name  = "DD_APM_NON_LOCAL_TRAFFIC"
          value = "true"
        },
        {
          name  = "ECS_FARGATE"
          value = "true"
        }
      ]
      # Default values that must be set to avoid re-creation
      cpu         = 0
      mountPoints = []
      volumesFrom = []
    }
  ])
}

# Transform the secrets map into AWS Secrets Manager secrets
resource "aws_secretsmanager_secret" "secrets" {
  for_each                = local.secret_keys
  description             = "Secret ${each.key}"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "secrets" {
  for_each      = aws_secretsmanager_secret.secrets
  secret_id     = each.value.id
  secret_string = var.secrets[each.key]
}

# Build dockerhub api token secret
resource "aws_secretsmanager_secret" "dockerhub_credentials" {
  count                   = var.dockerhub_username == "" || var.dockerhub_token == "" ? 0 : 1
  description             = "DockerHub API Token"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "dockerhub_credentials" {
  count     = var.dockerhub_username == "" || var.dockerhub_token == "" ? 0 : 1
  secret_id = aws_secretsmanager_secret.dockerhub_credentials[count.index].id
  secret_string = jsonencode({
    username = var.dockerhub_username
    password = var.dockerhub_token
  })
}

# Allow the ECS agent to assume the role created below
data "aws_iam_policy_document" "task_role_assume_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

# Create an execution role so that the ECS agent can access secrets
resource "aws_iam_role" "execution_role" {
  name_prefix        = "${var.name}-execution-role"
  assume_role_policy = data.aws_iam_policy_document.task_role_assume_policy.json
}

resource "aws_iam_role_policy" "secrets_policy" {
  count       = length(local.secret_keys) > 0 || (var.dockerhub_username != "" && var.dockerhub_token != "") ? 1 : 0
  name_prefix = var.name
  policy      = data.aws_iam_policy_document.secrets_policy.json
  role        = aws_iam_role.execution_role.id
}

# Allow the ECS agent to only access the secrets specified for this task
data "aws_iam_policy_document" "secrets_policy" {
  statement {
    actions = ["secretsmanager:GetSecretValue"]
    resources = concat(
      [for k, v in aws_secretsmanager_secret.secrets : v.arn],
      aws_secretsmanager_secret.dockerhub_credentials[*].arn
    )
  }
}

# This is used for ECS Exec Command
resource "aws_iam_role" "task_role" {
  name_prefix        = "${var.name}-task-role"
  assume_role_policy = data.aws_iam_policy_document.task_role_assume_policy.json
  inline_policy {
    name = "${var.name}-policy"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = concat(
        [
          {
            Effect = "Allow",
            Action = [
              "ssmmessages:CreateControlChannel",
              "ssmmessages:CreateDataChannel",
              "ssmmessages:OpenControlChannel",
              "ssmmessages:OpenDataChannel"
            ],
            Resource = "*"
          },
        ],
        var.additional_task_role_statements
      )
    })
  }
}
