module "public_lb_cert" {
  source = "./modules/acm_certificate"

  tld        = var.nodes_domain
  subdomains = ["*.${terraform.workspace}", terraform.workspace]

  providers = {
    aws        = aws
    cloudflare = cloudflare
  }
}

# re-use the validation for us-east-2 and deploy it in the other region
resource "aws_acm_certificate" "north" {
  domain_name               = module.public_lb_cert.domain_name
  subject_alternative_names = module.public_lb_cert.subject_alternative_names
  validation_method         = "DNS"

  lifecycle {
    create_before_destroy = true
  }

  tags = {
    Name = var.nodes_domain
  }

  provider = aws.eu-north-1
}
