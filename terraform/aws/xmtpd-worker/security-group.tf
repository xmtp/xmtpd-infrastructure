# The security group used for our xmtpd ECS services
resource "aws_security_group" "ecs_service" {
  name_prefix = "wrkr"
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
