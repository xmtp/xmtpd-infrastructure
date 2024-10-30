moved {
  from = module.vpc
  to   = module.xmtp_node_us_east_2.module.network.module.vpc
}

moved {
  from = module.mls_validation_service
  to   = module.xmtp_node_us_east_2.module.mls_validation_service
}

moved {
  from = aws_ecs_cluster.testnet
  to   = module.xmtp_node_us_east_2.aws_ecs_cluster.this
}

moved {
  from = aws_ecs_cluster_capacity_providers.testnet
  to   = module.xmtp_node_us_east_2.aws_ecs_cluster_capacity_providers.this
}

moved {
  from = aws_service_discovery_private_dns_namespace.xmtp
  to   = module.xmtp_node_eu_north_1.aws_service_discovery_private_dns_namespace.xmtp
}
