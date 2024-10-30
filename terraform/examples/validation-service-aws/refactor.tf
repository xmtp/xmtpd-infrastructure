moved {
  from = module.vpc
  to   = module.xmtp_node_us_east_2.module.network.module.vpc
}

moved {
  from = aws_iam_group.developers
  to   = module.grant_user_access.aws_iam_group.developers
}

moved {
  from = aws_iam_group_policy_attachment.power_user
  to   = module.grant_user_access.aws_iam_group_policy_attachment.power_user
}


moved {
  from = aws_iam_user.developer
  to   = module.grant_user_access.aws_iam_user.developer
}


moved {
  from = aws_iam_user_group_membership.developers
  to   = module.grant_user_access.aws_iam_user_group_membership.developers
}

moved {
  from = module.enforce_mfa
  to   = module.grant_user_access.module.enforce_mfa
}

moved {
  from = module.xmtpd_server
  to   = module.xmtp_node_us_east_2.module.xmtpd_server
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
