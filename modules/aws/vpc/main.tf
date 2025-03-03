locals {
  azs = var.public_subnets == null ? var.azs : length(var.public_subnets)
  vpc_cidr_size = tonumber(split("/", var.vpc_cidr)[1])
  vpc_split_ranges = local.vpc_cidr_size > 19 ? 1 : 2 #When the vpc_cicr is smaller than /19, then we only divide it into two(1),otherwise into four(2).
  
  #First range
  public_subnets = var.public_subnets == null ? [for i in range(local.azs) : cidrsubnet(cidrsubnet(cidrsubnet(var.vpc_cidr, local.vpc_split_ranges, 0), 1, 0), 2, i)] : var.public_subnets
  intra_subnets  = var.intra_subnets == null ? [for i in range(local.azs) : cidrsubnet(cidrsubnet(cidrsubnet(var.vpc_cidr, local.vpc_split_ranges, 0), 1, 1), 2, i)] : var.intra_subnets

  #second range
  private_subnets = var.private_subnets == null ? [for i in range(local.azs) : cidrsubnet(cidrsubnet(var.vpc_cidr, local.vpc_split_ranges, 1), 2, i)] : var.private_subnets

  #third range (only when vpc_cidr_size is large)
  database_subnets = var.database_subnets == null ? (
    local.vpc_cidr_size > 19 ? [] : [for i in range(local.azs) : cidrsubnet(cidrsubnet(var.vpc_cidr, local.vpc_split_ranges, 2), 2, i)]
  ) : var.database_subnets

  #fourth range (only when vpc_cidr_size is large)
  elasticache_subnets = var.elasticache_subnets == null ? (
  local.vpc_cidr_size > 19 ? [] : [for i in range(local.azs) : cidrsubnet(cidrsubnet(cidrsubnet(var.vpc_cidr, local.vpc_split_ranges, 3), 1, 0), 2, i)] 
  ) : var.elasticache_subnets
}


#https://registry.terraform.io/modules/terraform-aws-modules/vpc/aws/latest
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.19.0"

  name = var.prefix
  cidr = var.vpc_cidr

  secondary_cidr_blocks = var.secondary_cidr_blocks

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

  enable_flow_log                                 = var.enable_flow_log
  create_flow_log_cloudwatch_log_group            = var.enable_flow_log
  create_flow_log_cloudwatch_iam_role             = var.enable_flow_log
  flow_log_cloudwatch_log_group_name_prefix       = "${var.prefix}/vpc-flow-log/"
  flow_log_max_aggregation_interval               = var.flow_log_max_aggregation_interval
  flow_log_cloudwatch_log_group_retention_in_days = var.flow_log_cloudwatch_log_group_retention_in_days
  flow_log_cloudwatch_log_group_kms_key_id = var.flow_log_cloudwatch_log_group_kms_key_id != "" ? var.flow_log_cloudwatch_log_group_kms_key_id : null

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
