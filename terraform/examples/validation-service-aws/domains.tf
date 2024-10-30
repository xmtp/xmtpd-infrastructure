resource "cloudflare_record" "grpc_gateway" {
  zone_id = data.cloudflare_zone.nodes_domain.id
  name    = "grpc.${terraform.workspace}"
  value   = module.xmtp_node_us_east_2.grpc_service_dns_name
  type    = "CNAME"
  ttl     = 60
}

resource "cloudflare_record" "grpc_gateway_2" {
  zone_id = data.cloudflare_zone.nodes_domain.id
  name    = "grpc2.${terraform.workspace}"
  value   = module.xmtp_node_us_east_2.grpc_service_dns_name
  type    = "CNAME"
  ttl     = 60
}