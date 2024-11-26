# Our ECS cluster
resource "aws_ecs_cluster" "this" {
  name = terraform.workspace

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# This is required for Fargate
resource "aws_ecs_cluster_capacity_providers" "this" {
  cluster_name = aws_ecs_cluster.this.name

  capacity_providers = ["FARGATE"]

  # Route 100% of traffic to Fargate instances (which also happen to be the only kind of instances available)
  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 100
  }
}
