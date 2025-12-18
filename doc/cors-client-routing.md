# CORS and client routing

The **XMTP node** (xmtpd) uses **ConnectRPC** which supports both native **gRPC** (HTTP/2) and **gRPC-Web** (HTTP/1.1) protocols. 

To properly serve all client types—including native mobile/desktop apps, browser-based WASM clients, and web applications—your load balancer must route traffic appropriately and allow CORS preflight requests (OPTIONS requests) to reach the backend.

## Protocol requirements

| Client type | Protocol | Content-Type | Notes |
| --- | --- | --- | --- |
| Native gRPC (mobile, CLI, server) | `HTTP/2` | `application/grpc` | Requires HTTP/2 end-to-end |
| gRPC-Web / WASM (browser JS) | `HTTP/1.1` | `application/grpc-web` `application/grpc-web+proto*` `application/grpc-web-text*` `application/grpc-web-text+proto*` | Uses `x-grpc-web: 1` header |

## Load balancer configuration

It’s possible to configure an L4 or L7 load balancer in front of xmtpd. The requisite is the LB has to terminate the TLS, as `xmtpd` currently does not inject certificates nor expects ciphered connections.

```bash
┌─────────┐     HTTPS      ┌─────────┐     Plain HTTP     ┌─────────┐
│  Client │ ──────────────►│   NLB   │ ──────────────────►│  xmtpd  │
└─────────┘   (TLS)        └─────────┘   (TCP forward)    └─────────┘
                              │                               │
                         TLS terminates              h2c auto-detects:
                              here                   - HTTP/2 → gRPC
                                                     - HTTP/1.1 → gRPC-Web
```

### Target groups

In the case of using an L7 load balancer, you need **two separate target groups**:

1. **gRPC target group** (native clients)

- Protocol: `HTTP`
- Protocol version: `HTTP2`
- Used for native gRPC clients

2. **gRPC-Web target group** (browser/WASM clients)

- Protocol: `HTTP`
- Protocol version: `HTTP1`
- Used for gRPC-Web and WASM clients

Along with these target groups, you also need routing rules, in this order of priority:

1. HTTP method `OPTIONS` for CORS preflight requests.
2. Header `x-grpc-web: 1` for explicit gRPC-Web header. Match these Content-Type patterns:
   - `application/grpc-web*`
   - `application/grpc-web+proto*`
   - `application/grpc-web-text*`
   - `application/grpc-web-text+proto*`
3. Content-Type matches `application/grpc-web*` for gRPC-Web content types.
4. Final rule routes all other traffic, matching native gRPC clients.

> [!IMPORTANT]
> OPTIONS requests **must** be routed to the HTTP/1.1 target group. CORS preflight requests cannot be filtered by Content-Type and must reach the backend to receive proper CORS headers.

## CORS headers (handled by xmtpd)

The xmtpd server handles CORS **automatically**. Node operator infrastructure should **not** add CORS headers at the load balancer level to avoid conflicts.

### Headers set by xmtpd

#### On all responses

```bash
Access-Control-Allow-Origin: *

Access-Control-Expose-Headers: grpc-status,grpc-message,grpc-status-details-bin

Access-Control-Max-Age: 86400
```

#### On OPTIONS preflight responses (204 No Content)

```bash
Access-Control-Allow-Headers: Content-Type,Accept,Authorization,X-Client-Version,

X-App-Version,Baggage,DNT,Sec-CH-UA,Sec-CH-UA-Mobile,Sec-CH-UA-Platform,

x-grpc-web,grpc-timeout,Sentry-Trace,User-Agent,x-libxmtp-version,x-app-version

Access-Control-Allow-Methods: GET,HEAD,POST,PUT,PATCH,DELETE
```

## TLS/SSL requirements

- **TLS 1.2 or 1.3** required
  - Recommended: `ELBSecurityPolicy-TLS13-1-2-2021-06` or equivalent if in AWS
- Valid SSL certificate for your domain
- HTTPS listener on port 443

## Verification checklist

- Two target groups configured (HTTP2 for gRPC, HTTP1 for gRPC-Web)
- OPTIONS requests routed to HTTP1 target group (highest priority)
- `x-grpc-web: 1` header routing to HTTP1 target group
- `application/grpc-web*` Content-Type routing to HTTP1 target group
- Default traffic routes to HTTP2 target group
- TLS 1.2+ enabled with valid certificate
- **Do not** configure CORS headers at load balancer (let xmtpd handle it)

## Example AWS ALB configuration (Terraform)

### Layer 7 LB (AWS ALB)

```jsx
resource "aws_lb" "grpc" {
  name_prefix        = "example-"
  internal           = false
  load_balancer_type = local.lb_type
  subnets            = var.public_subnets
  security_groups    = [aws_security_group.ecs_service.id]

  access_logs {
    bucket  = var.elb_logs_bucket_name
    prefix  = local.access_log_prefix
    enabled = true
  }
}

resource "aws_lb_listener" "grpc" {
  load_balancer_arn = aws_lb.grpc.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = var.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.grpc.arn
  }
}

/* Listener rules for gRPC-Web requests */

# Prio 10: OPTIONS requests can't be redirected by Content-Type filters, so we need to route them to a separate target group.
resource "aws_lb_listener_rule" "grpcweb_preflight" {
  listener_arn = aws_lb_listener.grpc.arn
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
  listener_arn = aws_lb_listener.grpc.arn
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

// Prio 30: gRPC-Web requests by Content-Type.
resource "aws_lb_listener_rule" "grpcweb_by_content_type" {
  listener_arn = aws_lb_listener.grpc.arn
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

/* Target groups */

// Backend is a ConnectRPC service that uses HTTP2, and supports gRPC and gRPC-Web.
resource "aws_lb_target_group" "grpc" {
  depends_on       = [aws_lb.grpc]
  name_prefix      = "grpc-"
  port             = local.port
  protocol         = "HTTP"
  protocol_version = "HTTP2"
  target_type      = "ip"
  vpc_id           = var.vpc_id

  // ALB doesn't allow to send POST requests with specific Content-Type headers, required by the backend.
  // Sending a GET request fails with 405 Method Not Allowed. Accept 200-499 as healthy, until implementing health check,
  // that works with ALB.
  health_check {
    protocol          = "HTTP"
    path              = "/grpc.health.v1.Health/Check"
    port              = local.port
    interval          = 10
    healthy_threshold = 2
    matcher           = "200-499"
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_lb_target_group" "grpcweb" {
  name_prefix      = "wasm-"
  port             = local.port
  protocol         = "HTTP"
  protocol_version = "HTTP1"
  target_type      = "ip"
  vpc_id           = var.vpc_id

  health_check {
    protocol          = "HTTP"
    path              = "/grpc.health.v1.Health/Check"
    port              = local.port
    interval          = 10
    healthy_threshold = 2
    matcher           = "200-499"
  }

  lifecycle {
    create_before_destroy = true
  }
}

```

### Layer 4 LB (AWS NLB)

```bash
resource "aws_lb" "xmtpd" {
  name_prefix        = "xmtpd-"
  internal           = false
  load_balancer_type = "network"
  subnets            = var.public_subnets
}

resource "aws_lb_listener" "xmtpd" {
  load_balancer_arn = aws_lb.xmtpd.arn
  port              = 443
  protocol          = "TLS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = var.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.xmtpd.arn
  }
}

# Single target group - h2c handles protocol detection!
resource "aws_lb_target_group" "xmtpd" {
  name_prefix = "xmtpd-"
  port        = var.service_port
  protocol    = "TCP"
  target_type = "ip"
  vpc_id      = var.vpc_id

  health_check {
    protocol = "TCP"
    port     = var.service_port
  }

  lifecycle {
    create_before_destroy = true
  }
}
```

## Related documentation

- [XMTP Node Networking](https://github.com/xmtp/xmtpd/blob/main/doc/networking.md)
