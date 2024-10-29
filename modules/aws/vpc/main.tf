locals {
  azs = var.public_subnets == null ? var.azs : length(var.public_subnets)

  #First range
  public_subnets = var.public_subnets == null ? [for i in range(local.azs) : cidrsubnet(cidrsubnet(cidrsubnet(var.vpc_cidr, 2, 0), 1, 0), 2, i)] : var.public_subnets
  intra_subnets  = var.intra_subnets == null ? [for i in range(local.azs) : cidrsubnet(cidrsubnet(cidrsubnet(var.vpc_cidr, 2, 0), 1, 1), 2, i)] : var.intra_subnets

  #second range
  private_subnets = var.private_subnets == null ? [for i in range(local.azs) : cidrsubnet(cidrsubnet(var.vpc_cidr, 2, 1), 2, i)] : var.private_subnets

  #third range
  database_subnets = var.database_subnets == null ? [for i in range(local.azs) : cidrsubnet(cidrsubnet(var.vpc_cidr, 2, 2), 2, i)] : var.database_subnets

  #fourth range
  elasticache_subnets = var.elasticache_subnets == null ? [for i in range(local.azs) : cidrsubnet(cidrsubnet(cidrsubnet(var.vpc_cidr, 2, 3), 1, 0), 2, i)] : var.elasticache_subnets
}


#https://registry.terraform.io/modules/terraform-aws-modules/vpc/aws/latest

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.8.1"

  name = var.prefix
  cidr = var.vpc_cidr

  azs                 = [for i in range(local.azs) : data.aws_availability_zones.available.names[i]]
  private_subnets     = local.private_subnets
  public_subnets      = local.public_subnets
  database_subnets    = local.database_subnets
  elasticache_subnets = local.elasticache_subnets
  intra_subnets       = local.intra_subnets

  private_subnet_names     = var.private_subnet_names
  public_subnet_names      = var.public_subnet_names
  database_subnet_names    = var.database_subnet_names
  elasticache_subnet_names = var.elasticache_subnet_names
  intra_subnet_names       = var.intra_subnet_names

  create_database_subnet_group    = length(local.database_subnets) > 0 ? true : false
  create_elasticache_subnet_group = length(local.elasticache_subnets) > 0 ? true : false

  enable_nat_gateway     = var.enable_nat_gateway
  single_nat_gateway     = var.one_nat_gateway_per_az ? false : true
  one_nat_gateway_per_az = var.one_nat_gateway_per_az

  reuse_nat_ips        = false
  enable_dns_hostnames = true
  enable_dns_support   = true

  map_public_ip_on_launch = var.map_public_ip_on_launch

  enable_flow_log                                 = false
  create_flow_log_cloudwatch_log_group            = false
  create_flow_log_cloudwatch_iam_role             = false
  flow_log_max_aggregation_interval               = 60
  flow_log_cloudwatch_log_group_retention_in_days = 7

  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
    "karpenter.sh/discovery" = var.prefix
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
    "karpenter.sh/discovery" = var.prefix
  }

  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_security_group" "pipeline_security_group" {
  name        = "${var.prefix}-pipeline"
  description = "${var.prefix} Security group used by pipelines that run terraform"
  vpc_id      = module.vpc.vpc_id
  tags = {
    Name = "Allow pipeline access to ${var.prefix}"
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
  name  = "/entigo-infralib/${var.prefix}/pipeline_security_group"
  type  = "String"
  value = aws_security_group.pipeline_security_group.id
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "vpc_id" {
  name  = "/entigo-infralib/${var.prefix}/vpc_id"
  type  = "String"
  value = module.vpc.vpc_id
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "private_subnets" {
  name  = "/entigo-infralib/${var.prefix}/private_subnets"
  type  = "StringList"
  value = "\"${join("\",\"", module.vpc.private_subnets)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "public_subnets" {
  name  = "/entigo-infralib/${var.prefix}/public_subnets"
  type  = "StringList"
  value = "\"${join("\",\"", module.vpc.public_subnets)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "intra_subnets" {
  count = length(module.vpc.intra_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/intra_subnets"
  type  = "StringList"
  value = "\"${join("\",\"", module.vpc.intra_subnets)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "database_subnets" {
  count = length(local.database_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/database_subnets"
  type  = "StringList"
  value = "\"${join("\",\"", module.vpc.database_subnets)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "database_subnet_group" {
  count          = length(local.database_subnets) > 0 ? 1 : 0
  name           = "/entigo-infralib/${var.prefix}/database_subnet_group"
  type           = "String"
  insecure_value = module.vpc.database_subnet_group
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "elasticache_subnets" {
  count = length(local.elasticache_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/elasticache_subnets"
  type  = "StringList"
  value = "\"${join("\",\"", module.vpc.elasticache_subnets)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "elasticache_subnet_group" {
  count          = length(local.elasticache_subnets) > 0 ? 1 : 0
  name           = "/entigo-infralib/${var.prefix}/elasticache_subnet_group"
  type           = "String"
  insecure_value = module.vpc.elasticache_subnet_group
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "private_subnet_cidrs" {
  name  = "/entigo-infralib/${var.prefix}/private_subnet_cidrs"
  type  = "StringList"
  value = "\"${join("\",\"", module.vpc.private_subnets_cidr_blocks)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}


resource "aws_ssm_parameter" "public_subnet_cidrs" {
  name  = "/entigo-infralib/${var.prefix}/public_subnet_cidrs"
  type  = "StringList"
  value = "\"${join("\",\"", module.vpc.public_subnets_cidr_blocks)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}


resource "aws_ssm_parameter" "database_subnet_cidrs" {
  count = length(local.database_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/database_subnet_cidrs"
  type  = "StringList"
  value = "\"${join("\",\"", module.vpc.database_subnets_cidr_blocks)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "elasticache_subnet_cidrs" {
  count = length(local.elasticache_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/elasticache_subnet_cidrs"
  type  = "StringList"
  value = "\"${join("\",\"", module.vpc.elasticache_subnets_cidr_blocks)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "intra_subnet_cidrs" {
  count = length(local.intra_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/intra_subnet_cidrs"
  type  = "StringList"
  value = "\"${join("\",\"", module.vpc.intra_subnets_cidr_blocks)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "private_subnet_names" {
  name  = "/entigo-infralib/${var.prefix}/private_subnet_names"
  type  = "StringList"
  value = "\"${join("\",\"", var.private_subnet_names)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "public_subnet_names" {
  name  = "/entigo-infralib/${var.prefix}/public_subnet_names"
  type  = "StringList"
  value = "\"${join("\",\"", var.public_subnet_names)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "database_subnet_names" {
  name  = "/entigo-infralib/${var.prefix}/database_subnet_names"
  type  = "StringList"
  value = "\"${join("\",\"", var.database_subnet_names)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "elasticache_subnet_names" {
  name  = "/entigo-infralib/${var.prefix}/elasticache_subnet_names"
  type  = "StringList"
  value = "\"${join("\",\"", var.elasticache_subnet_names)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "intra_subnet_names" {
  name  = "/entigo-infralib/${var.prefix}/intra_subnet_names"
  type  = "StringList"
  value = "\"${join("\",\"", var.intra_subnet_names)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}


resource "aws_ssm_parameter" "intra_route_table_ids" {
  count = length(local.intra_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/intra_route_table_ids"
  type  = "StringList"
  value = "\"${join("\",\"", module.vpc.intra_route_table_ids)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "private_route_table_ids" {
  count = length(local.private_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/private_route_table_ids"
  type  = "StringList"
  value = "\"${join("\",\"", module.vpc.private_route_table_ids)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "public_route_table_ids" {
  count = length(local.public_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/public_route_table_ids"
  type  = "StringList"
  value = "\"${join("\",\"", module.vpc.public_route_table_ids)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "database_route_table_ids" {
  count = length(local.database_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/database_route_table_ids"
  type  = "StringList"
  value = "\"${join("\",\"", module.vpc.database_route_table_ids)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "elasticache_route_table_ids" {
  count = length(local.elasticache_subnets) > 0 ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/elasticache_route_table_ids"
  type  = "StringList"
  value = "\"${join("\",\"", module.vpc.elasticache_route_table_ids)}\""
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

