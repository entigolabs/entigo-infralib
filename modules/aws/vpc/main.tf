locals {
  private_subnets     = var.private_subnets == null ? [cidrsubnet(var.vpc_cidr, 5, 0), cidrsubnet(var.vpc_cidr, 5, 1), cidrsubnet(var.vpc_cidr, 5, 2)] : var.private_subnets
  public_subnets      = var.public_subnets == null ? [cidrsubnet(var.vpc_cidr, 5, 3), cidrsubnet(var.vpc_cidr, 5, 4)] : var.public_subnets
  database_subnets    = var.database_subnets == null ? [cidrsubnet(var.vpc_cidr, 5, 6), cidrsubnet(var.vpc_cidr, 5, 7), cidrsubnet(var.vpc_cidr, 5, 8)] : var.database_subnets
  elasticache_subnets = var.elasticache_subnets == null ? [cidrsubnet(var.vpc_cidr, 5, 9), cidrsubnet(var.vpc_cidr, 5, 10), cidrsubnet(var.vpc_cidr, 5, 11)] : var.elasticache_subnets
  intra_subnets       = var.intra_subnets == null ? [cidrsubnet(var.vpc_cidr, 5, 12), cidrsubnet(var.vpc_cidr, 5, 13), cidrsubnet(var.vpc_cidr, 5, 14)] : var.intra_subnets
}


#https://registry.terraform.io/modules/terraform-aws-modules/vpc/aws/latest
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.0.0"

  name = local.hname
  cidr = var.vpc_cidr

  azs                 = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1], data.aws_availability_zones.available.names[2]]
  private_subnets     = local.private_subnets
  public_subnets      = local.public_subnets
  database_subnets    = local.database_subnets
  elasticache_subnets = local.elasticache_subnets
  intra_subnets       = local.intra_subnets

  create_database_subnet_group    = length(local.database_subnets) > 0 ? true : false
  create_elasticache_subnet_group = length(local.elasticache_subnets) > 0 ? true : false
  
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

resource "aws_security_group" "pipeline_security_group" {
  name        = "${local.hname}-pipeline"
  description = "${local.hname} Security group used by pipelines that run terraform"
  vpc_id      = module.vpc.vpc_id
  tags = {
    Name = "Allow pipeline access to ${local.hname}"
  }
}

resource "aws_security_group_rule" "pipeline_security_group" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = -1
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.pipeline_security_group.id
}

resource "aws_ssm_parameter" "pipeline_security_group" {
  name  = "/entigo-infralib/${local.hname}/vpc/pipeline_security_group"
  type  = "String"
  value = aws_security_group.pipeline_security_group.id
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
  count = length(local.database_subnets) > 0 ? 1 : 0
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
  count = length(local.database_subnets) > 0 ? 1 : 0
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
  count = length(local.elasticache_subnets) > 0 ? 1 : 0
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
  count = length(local.elasticache_subnets) > 0 ? 1 : 0
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
  count = length(local.database_subnets) > 0 ? 1 : 0
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
  count = length(local.elasticache_subnets) > 0 ? 1 : 0
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
  count = length(local.intra_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${local.hname}/vpc/intra_subnet_cidrs"
  type  = "String"
  value = join(",", module.vpc.intra_subnets_cidr_blocks)
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

