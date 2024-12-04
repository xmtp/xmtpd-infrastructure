
# This domain must ALREADY BE REGISTERED in the AWS account.
# Terraform will then take control of the registered domain
resource "aws_route53domains_registered_domain" "public" {
  domain_name = var.domain_name

  # Enable auto-renewal
  auto_renew = true

  # Enable privacy protection
  admin_privacy      = true
  registrant_privacy = true
  tech_privacy       = true

  name_server {
    name = aws_route53_zone.public.name_servers[0]
  }

  name_server {
    name = aws_route53_zone.public.name_servers[1]
  }
}

# Create a Route53 hosted zone for the domain
resource "aws_route53_zone" "public" {
  name = var.domain_name
}

# Create an ACM certificate for the domain
resource "aws_acm_certificate" "public" {
  depends_on        = [aws_route53domains_registered_domain.public]
  domain_name       = var.domain_name
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}

# Create DNS records for ACM certificate validation
resource "aws_route53_record" "cert_validation" {
  for_each = {
    for dvo in aws_acm_certificate.public.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = aws_route53_zone.public.zone_id
}

# Wait for certificate validation to complete
resource "aws_acm_certificate_validation" "public" {
  certificate_arn         = aws_acm_certificate.public.arn
  validation_record_fqdns = [for record in aws_route53_record.cert_validation : record.fqdn]
}

# Create DNS record for the load balancer
resource "aws_route53_record" "lb" {
  zone_id = aws_route53_zone.public.zone_id
  name    = var.domain_name
  type    = "A"

  alias {
    name                   = module.xmtpd_api.load_balancer_address
    zone_id                = module.xmtpd_api.load_balancer_zone_id
    evaluate_target_health = true
  }
}
