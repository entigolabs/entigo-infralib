terraform {
  backend "s3" {}
  required_version = ">= 1.4"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">4"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~>2.0"
    }
  }
}

provider "aws" {
}

# Test with default values
module "test_vpc" {
  source = "../"

  prefix = var.prefix
  vpc_cidr = var.vpc_cidr
  one_nat_gateway_per_az = var.one_nat_gateway_per_az
  private_subnets = var.private_subnets
  public_subnets = var.public_subnets
  database_subnets = var.database_subnets
  elasticache_subnets = var.elasticache_subnets
  intra_subnets = var.intra_subnets
}


output "vpc_id" {
  value = module.test_vpc.vpc_id
}

output "private_subnets" {
  value = module.test_vpc.private_subnets
}

output "public_subnets" {
  value = module.test_vpc.public_subnets
}

output "intra_subnets" {
  value = module.test_vpc.intra_subnets
}

output "database_subnets" {
  value = module.test_vpc.database_subnets
}

output "database_subnet_group" {
  value = module.test_vpc.database_subnet_group
}

output "elasticache_subnets" {
  value = module.test_vpc.elasticache_subnets
}

output "elasticache_subnet_group" {
  value = module.test_vpc.elasticache_subnet_group
}

output "private_subnet_cidrs" {
  value = module.test_vpc.private_subnet_cidrs
}

output "public_subnet_cidrs" {
  value = module.test_vpc.public_subnet_cidrs
}

output "database_subnet_cidrs" {
  value = module.test_vpc.database_subnet_cidrs
}

output "elasticache_subnet_cidrs" {
  value = module.test_vpc.elasticache_subnet_cidrs
}

output "intra_subnet_cidrs" {
  value = module.test_vpc.intra_subnet_cidrs
}


