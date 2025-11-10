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

/* Listener rules for gRPC-Web requests */

// Prio 10: OPTIONS requests can't be redirected by content-type filters, so we need to route them to a separate target group.
resource "aws_lb_listener_rule" "grpcweb_preflight" {
  listener_arn = aws_lb_listener.public.arn
  priority     = 10

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.grpcweb.arn
  }

  condition {
    http_request_method {
      values = ["OPTIONS"]
    }
  }
}

// Prio 20: gRPC-Web requests by header.
resource "aws_lb_listener_rule" "grpcweb_by_header" {
  listener_arn = aws_lb_listener.public.arn
  priority     = 20

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.grpcweb.arn
  }

  condition {
    http_header {
      http_header_name = "x-grpc-web"
      values           = ["1"]
    }
  }
}

// Prio 30: gRPC-Web requests by content-type.
resource "aws_lb_listener_rule" "grpcweb_by_content_type" {
  listener_arn = aws_lb_listener.public.arn
  priority     = 30

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.grpcweb.arn
  }

  condition {
    http_header {
      http_header_name = "content-type"
      values = [
        "application/grpc-web*",
        "application/grpc-web+proto*",
        "application/grpc-web-text*",
        "application/grpc-web-text+proto*",
      ]
    }
  }
}

// Target group used for gRPC requests.
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

// Target group used for gRPC-Web requests.
resource "aws_lb_target_group" "grpcweb" {
  name_prefix      = "wasm-"
  port             = local.service_port
  protocol         = "HTTP"
  protocol_version = "HTTP1"
  target_type      = "ip"
  vpc_id           = var.vpc_id

  health_check {
    protocol          = "HTTP"
    path              = "/grpc.health.v1.Health/Check"
    port              = local.service_port
    interval          = 10
    healthy_threshold = 2
    matcher           = "200-499"
  }

  lifecycle {
    create_before_destroy = true
  }
}
