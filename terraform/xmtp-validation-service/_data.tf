data "aws_service_discovery_dns_namespace" "xmtp" {
  name = var.service_discovery_namespace_name
  type = "DNS_PRIVATE"
}