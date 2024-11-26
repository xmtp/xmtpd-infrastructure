output "private_subnets" {
  value = module.vpc.private_subnets
}

output "public_subnets" {
  value = module.vpc.public_subnets
}

output "azs" {
  value = local.azs
}

output "vpc_id" {
  value = module.vpc.vpc_id
}

output "vpc_cidr" {
  value = var.vpc_cidr
}