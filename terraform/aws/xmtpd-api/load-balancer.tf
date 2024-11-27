locals {
  lb_type           = "application"
  access_log_prefix = "xmtpd-access-logs"
}

resource "aws_lb" "public" {
  name_prefix        = "xmtpd-"
  internal           = false
  load_balancer_type = local.lb_type
  subnets            = var.public_subnets
  security_groups    = [aws_security_group.load_balancer.id]

  dynamic "access_logs" {
    for_each = toset(var.elb_logs_bucket_name != null ? ["enabled"] : [])
    content {
      bucket  = var.elb_logs_bucket_name
      prefix  = local.access_log_prefix
      enabled = true
    }
  }
}

resource "aws_lb_listener" "public" {
  load_balancer_arn = aws_lb.public.arn
  port              = local.public_port
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = var.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.public.arn
  }
}

resource "aws_lb_target_group" "public" {
  depends_on       = [aws_lb.public]
  name_prefix      = "xmtpd-"
  port             = local.service_port
  protocol         = "HTTP"
  protocol_version = "GRPC"
  target_type      = "ip"
  vpc_id           = var.vpc_id

  health_check {
    path              = "/"
    port              = local.service_port
    interval          = 10
    healthy_threshold = 2
    matcher           = 12 // Check for a GRPC Unimplemented status which ensures that the GRPC server is running
  }
}
