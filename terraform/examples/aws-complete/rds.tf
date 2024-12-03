locals {
  db_engine_version            = "16"
  db_name                      = "xmtp"
  db_root_user                 = "xmtp"
  is_production_environment    = false
  db_num_instances             = 2
  db_instance_class            = "db.t4g.medium"
  db_ca_certificate_identifier = "rds-ca-rsa2048-g1"
  db_parameter_group_family    = "aurora-postgresql16"
}

resource "random_password" "password" {
  length  = 64
  special = false
}

resource "aws_rds_cluster" "cluster" {
  engine                          = "aurora-postgresql"
  engine_version                  = local.db_engine_version
  availability_zones              = module.vpc.azs
  database_name                   = local.db_name
  master_username                 = local.db_root_user
  master_password                 = random_password.password.result
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.default.name
  db_subnet_group_name            = aws_db_subnet_group.cluster.name
  vpc_security_group_ids          = [aws_security_group.rds.id]
  deletion_protection             = local.is_production_environment
  apply_immediately               = true
  # These will need to be turned off for production usage
  backup_retention_period = local.is_production_environment ? 30 : 1
  skip_final_snapshot     = local.is_production_environment ? false : true

  lifecycle {
    ignore_changes = [
      availability_zones
    ]
  }
}

resource "aws_rds_cluster_instance" "instances" {
  count = local.db_num_instances

  cluster_identifier           = aws_rds_cluster.cluster.id
  instance_class               = local.db_instance_class
  engine                       = aws_rds_cluster.cluster.engine
  engine_version               = aws_rds_cluster.cluster.engine_version
  auto_minor_version_upgrade   = false
  ca_cert_identifier           = local.db_ca_certificate_identifier
  publicly_accessible          = false
  performance_insights_enabled = true
  db_subnet_group_name         = aws_db_subnet_group.cluster.name
  apply_immediately            = true
}

resource "aws_db_subnet_group" "cluster" {
  subnet_ids = module.vpc.private_subnets
}

# Create a parameter group so that we can adjust parameters later without recreating the cluster
resource "aws_rds_cluster_parameter_group" "default" {
  family      = local.db_parameter_group_family
  description = "RDS cluster parameter group"

  parameter {
    name = "log_temp_files"
    # Log any temp files greater than 1MB
    value        = "1000"
    apply_method = "pending-reboot"
  }
}


resource "aws_security_group" "rds" {
  description = "RDS security group"
  vpc_id      = module.vpc.vpc_id
}

resource "aws_security_group_rule" "ingress" {
  description       = "Allow Postgres traffic from our VPC"
  cidr_blocks       = [module.vpc.vpc_cidr_block]
  from_port         = 5432
  protocol          = "tcp"
  security_group_id = aws_security_group.rds.id
  to_port           = 5432
  type              = "ingress"
}
