resource "aws_service_discovery_private_dns_namespace" "xmtp" {
  name        = "xmtp.private"
  description = "The AWS service discovery namespace"
  vpc         = module.vpc.vpc_id
}
