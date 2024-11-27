# The security group used for our xmtpd ECS services
resource "aws_security_group" "ecs_service" {
  name_prefix = "ecs-svc"
  description = "xmtpd ECS service security group"
  vpc_id      = var.vpc_id
}

resource "aws_security_group_rule" "egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  ipv6_cidr_blocks  = ["::/0"]
  description       = "Allow all traffic to egress from ECS services"
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.ecs_service.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "lb_to_ecs" {
  description              = "Allow traffic from load balancer to ECS service"
  from_port                = local.service_port
  protocol                 = "-1"
  security_group_id        = aws_security_group.ecs_service.id
  source_security_group_id = aws_security_group.load_balancer.id
  to_port                  = local.service_port
  type                     = "ingress"
}


resource "aws_security_group" "load_balancer" {
  name_prefix = "lb"
  description = "Load balancer security group"
  vpc_id      = var.vpc_id
}

resource "aws_security_group_rule" "lb_ingress" {
  description       = "Allow https"
  cidr_blocks       = ["0.0.0.0/0"]
  ipv6_cidr_blocks  = ["::/0"]
  from_port         = local.public_port
  protocol          = "-1"
  security_group_id = aws_security_group.ecs_service.id
  to_port           = local.public_port
  type              = "ingress"
}

resource "aws_security_group_rule" "lb_egress" {
  description              = "Allow traffic from load balancer to ECS service"
  from_port                = local.service_port
  protocol                 = "-1"
  security_group_id        = aws_security_group.load_balancer.id
  source_security_group_id = aws_security_group.ecs_service.id
  to_port                  = local.service_port
  type                     = "egress"
}
