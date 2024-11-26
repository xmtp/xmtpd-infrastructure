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

  labels = var.docker_labels
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
    aws_secretsmanager_secret_version.secrets
  ]

  container_definitions = jsonencode([
    merge(
      {
        name         = var.name
        image        = var.image
        essential    = true
        dockerLabels = local.labels
        command      = var.command
        logConfiguration = {
          logDriver = "awslogs"
          options = {
            awslogs-create-group  = "true"
            awslogs-group         = "awslogs-${var.name}"
            awslogs-region        = data.aws_region.current.name
            awslogs-stream-prefix = "awslogs-${var.name}"
          }
        }
        environment = local.env_vars
        # Create a mapping of secret names to the ARN of the secret manager secrets
        secrets = [for name, secret in aws_secretsmanager_secret.secrets : {
          name      = name
          valueFrom = secret.arn
        }]
        portMappings = local.port_mappings
        # Default values that must be set to avoid re-creation
        cpu = 0
        volumesFrom = []
        ulimits = [
          {
            name      = "nofile"
            softLimit = 65535
            hardLimit = 65535
          }
        ]
      },
      var.health_check_config != null ? {
        healthCheck = var.health_check_config
      } : {}
    )
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

data "aws_region" "current" {}

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
  count       = length(local.secret_keys) > 0 ? 1 : 0
  name_prefix = var.name
  policy      = data.aws_iam_policy_document.secrets_policy.json
  role        = aws_iam_role.execution_role.id
}

# Allow the ECS agent to only access the secrets specified for this task
data "aws_iam_policy_document" "secrets_policy" {
  statement {
    actions   = ["secretsmanager:GetSecretValue"]
    resources = [for k, v in aws_secretsmanager_secret.secrets : v.arn]
  }
}

resource "aws_iam_role_policy" "logs_policy" {
  name_prefix = var.name
  policy      = data.aws_iam_policy_document.logs_policy.json
  role        = aws_iam_role.execution_role.id
}

data "aws_iam_policy_document" "logs_policy" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["arn:aws:logs:*:*:*"]
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
