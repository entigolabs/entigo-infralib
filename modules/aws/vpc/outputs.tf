output "vpc_id" {
  value = module.vpc.vpc_id
}

output "private_subnets" {
  value = module.vpc.private_subnets
}

output "public_subnets" {
  value = module.vpc.public_subnets
}

output "intra_subnets" {
  value = module.vpc.intra_subnets
}

output "database_subnets" {
  value = module.vpc.database_subnets
}

output "database_subnet_group" {
  value = module.vpc.database_subnet_group
}

output "elasticache_subnets" {
  value = module.vpc.elasticache_subnets
}

output "elasticache_subnet_group" {
  value = module.vpc.elasticache_subnet_group
}

output "private_subnet_cidrs" {
  value = module.vpc.private_subnets_cidr_blocks
}

output "public_subnet_cidrs" {
  value = module.vpc.public_subnets_cidr_blocks
}

output "database_subnet_cidrs" {
  value = module.vpc.database_subnets_cidr_blocks
}

output "elasticache_subnet_cidrs" {
  value = module.vpc.elasticache_subnets_cidr_blocks
}

output "intra_subnet_cidrs" {
  value = module.vpc.intra_subnets_cidr_blocks
}


