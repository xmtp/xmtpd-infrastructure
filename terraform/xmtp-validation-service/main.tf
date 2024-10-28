locals {
  name              = "xmtp-validation-service"
  service_port      = 50051
  health_check_port = 50052
  health_check_path = "/health"
}

/*
* TODO:
* - How do we test this thing?
*    - terraform fmt, terraform init+validate, terraform plan, terraform apply, terraform destroy
*    - Set up CI
*    - terraform apply onto an AWS instance
* - AWS is required provider. Default tag can be set on the provider, so any AWS resource created will have that tag for the environment
* - Need fargate_task_definition module -> get rid of data dog stuff, logs go to cloudwatch, don’t need docker hub access token
* - Docker image can be specified as input variable, otherwise default to latest
* - Do we need environment input variable? Default to 'testnet'
* - Do we need service_discovery_namespace_name? -> use loadbalancer
*/

module "service_sg" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "5.2.0"

  name        = "validation"
  description = "Security group for the validation service allowing inbound traffic from inside the VPC"
  vpc_id      = var.vpc_id

  ingress_with_cidr_blocks = [
    {
      from_port   = local.service_port
      to_port     = local.service_port
      protocol    = "tcp"
      description = "Service port"
      # Only allow access from private subnets
      cidr_blocks = join(",", var.allowed_ingress_cidr_blocks)
    },
  ]
  egress_with_cidr_blocks = [
    {
      from_port   = 0
      to_port     = 0
      protocol    = "-1"
      description = "Allow all traffic to egress from ECS services"
      cidr_blocks = "0.0.0.0/0"
    }
  ]
}

module "task_definition" {
  source = "../fargate-task-definition"
  name   = local.name
  cpu    = var.cpu
  memory = var.memory
  env_vars = {
    "ENV" = var.env
  }

  secrets = { for key, val in var.chain_rpc_urls : upper(key) => val }

  ports           = [local.service_port, local.health_check_port]
  image           = var.docker_image
  datadog_api_key = var.datadog_api_key
  env             = var.env
  health_check_config = {
    # CMD-SHELL tells ECS to use the container's default shell to run the command
    # https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_HealthCheck.html
    command = ["CMD-SHELL", "curl -f http://localhost:${local.health_check_port}${local.health_check_path} || exit 1"]
  }

  providers = {
    aws = aws
  }
}

resource "aws_service_discovery_service" "validation" {
  name = local.name

  dns_config {
    namespace_id = data.aws_service_discovery_dns_namespace.xmtp.id

    dns_records {
      type = "A"
      ttl  = 10
    }
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}

resource "aws_ecs_service" "validation" {
  name                               = local.name
  cluster                            = var.cluster_id
  task_definition                    = module.task_definition.task_definition_arn
  enable_execute_command             = false
  desired_count                      = 2
  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 100
  wait_for_steady_state              = true

  network_configuration {
    subnets         = var.private_subnets
    security_groups = [module.service_sg.security_group_id]
  }

  capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 100
  }

  service_registries {
    registry_arn   = aws_service_discovery_service.validation.arn
    container_name = local.name
  }
}
