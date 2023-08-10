#https://registry.terraform.io/modules/terraform-aws-modules/vpc/aws/latest
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.0.0"

  name = local.hname
  cidr = var.vpc_cidr

  azs                 = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1], data.aws_availability_zones.available.names[2]]
  private_subnets     = var.private_subnets
  public_subnets      = var.public_subnets
  database_subnets    = var.database_subnets
  elasticache_subnets = var.elasticache_subnets
  intra_subnets       = var.intra_subnets

  create_database_subnet_group    = length(var.database_subnets) > 0 ? true : false
  create_elasticache_subnet_group = length(var.elasticache_subnets) > 0 ? true : false
  
  enable_nat_gateway              = true
  single_nat_gateway              = var.one_nat_gateway_per_az ? false : true
  one_nat_gateway_per_az          = var.one_nat_gateway_per_az
  
  reuse_nat_ips                   = false
  enable_dns_hostnames            = true
  enable_dns_support              = true

  enable_flow_log                                 = false
  create_flow_log_cloudwatch_log_group            = false
  create_flow_log_cloudwatch_iam_role             = false
  flow_log_max_aggregation_interval               = 60
  flow_log_cloudwatch_log_group_retention_in_days = 7

  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
  }

  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
} 


resource "aws_ssm_parameter" "vpc_id" {
  name  = "/entigo-infralib/${local.hname}/vpc/vpc_id"
  type  = "String"
  value = module.vpc.vpc_id
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "private_subnets" {
  name  = "/entigo-infralib/${local.hname}/vpc/private_subnets"
  type  = "String"
  value = join(",", module.vpc.private_subnets)
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "public_subnets" {
  name  = "/entigo-infralib/${local.hname}/vpc/public_subnets"
  type  = "String"
  value = join(",", module.vpc.public_subnets)
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "intra_subnets" {
  count = length(module.vpc.intra_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${local.hname}/vpc/intra_subnets"
  type  = "String"
  value = join(",", module.vpc.intra_subnets)
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "database_subnets" {
  count = length(var.database_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${local.hname}/vpc/database_subnets"
  type  = "String"
  value = join(",", module.vpc.database_subnets)
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "database_subnet_group" {
  count = length(var.database_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${local.hname}/vpc/database_subnet_group"
  type  = "String"
  insecure_value = module.vpc.database_subnet_group
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "elasticache_subnets" {
  count = length(var.elasticache_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${local.hname}/vpc/elasticache_subnets"
  type  = "String"
  value = join(",", module.vpc.elasticache_subnets)
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "elasticache_subnet_group" {
  count = length(var.elasticache_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${local.hname}/vpc/elasticache_subnet_group"
  type  = "String"
  insecure_value = module.vpc.elasticache_subnet_group
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "private_subnet_cidrs" {
  name  = "/entigo-infralib/${local.hname}/vpc/private_subnet_cidrs"
  type  = "String"
  value = join(",", module.vpc.private_subnets_cidr_blocks)
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}


resource "aws_ssm_parameter" "public_subnet_cidrs" {
  name  = "/entigo-infralib/${local.hname}/vpc/public_subnet_cidrs"
  type  = "String"
  value = join(",", module.vpc.public_subnets_cidr_blocks)
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}


resource "aws_ssm_parameter" "database_subnet_cidrs" {
  count = length(var.database_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${local.hname}/vpc/database_subnet_cidrs"
  type  = "String"
  value = join(",", module.vpc.database_subnets_cidr_blocks)
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "elasticache_subnet_cidrs" {
  count = length(var.elasticache_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${local.hname}/vpc/elasticache_subnet_cidrs"
  type  = "String"
  value = join(",", module.vpc.elasticache_subnets_cidr_blocks)
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "intra_subnet_cidrs" {
  count = length(var.intra_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${local.hname}/vpc/intra_subnet_cidrs"
  type  = "String"
  value = join(",", module.vpc.intra_subnets_cidr_blocks)
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

