locals {
  azs = var.public_subnets == null ? var.azs : length(var.public_subnets)
  vpc_cidr_size = tonumber(split("/", var.vpc_cidr)[1])
  vpc_split_ranges = local.vpc_cidr_size > 19 ? 1 : 2 #When the vpc_cicr is smaller than /19, then we only divide it into two(1),otherwise into four(2).
  
  #Calculations for "spoke"
  spoke_tgw_nacl      = cidrsubnet(cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 0), 2, 0), (23-local.vpc_cidr_size), 0)
  spoke_control_nacl  = cidrsubnet(cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 0), 2, 0), ((local.vpc_cidr_size <= 19 ? 22 : 23)-local.vpc_cidr_size), 1)
  spoke_public_nacl   =            cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 0), 2, 1)
  spoke_service_nacl  =            cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 0), 2, 2)
  spoke_database_nacl =            cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 0), 2, 3)
  spoke_compute_nacl  =                       cidrsubnet(var.vpc_cidr, 1, 1)
  
  spoke_tgw      =  [for i in range(local.azs) : cidrsubnet(local.spoke_tgw_nacl, 2, i)]
  spoke_control  =  [for i in range(local.azs) : cidrsubnet(local.spoke_control_nacl, 2, i)]
  spoke_public   =  [for i in range(local.azs) : cidrsubnet(local.spoke_public_nacl, 2, i)]
  spoke_service  =  [for i in range(local.azs) : cidrsubnet(local.spoke_service_nacl, 2, i)]
  spoke_database =  [for i in range(local.azs) : cidrsubnet(local.spoke_database_nacl, 2, i)]
  spoke_compute  =  [for i in range(local.azs) : cidrsubnet(local.spoke_compute_nacl, 2, i)]
  
  #Naming for "spoke"
  spoke_tgw_names =  [for i in range(local.azs) : format("${var.prefix}-tgw-%s", element(data.aws_availability_zones.available.names, i))]
  spoke_control_names =  [for i in range(local.azs) : format("${var.prefix}-control-%s", element(data.aws_availability_zones.available.names, i))]
  spoke_service_names =  [for i in range(local.azs) : format("${var.prefix}-service-%s", element(data.aws_availability_zones.available.names, i))]
  spoke_compute_names =  [for i in range(local.azs) : format("${var.prefix}-compute-%s", element(data.aws_availability_zones.available.names, i))]
  
  #Calculations for "default"

  default_public_nacl      =                                 cidrsubnet(cidrsubnet(var.vpc_cidr, local.vpc_split_ranges, 0), 1, 0)
  default_intra_nacl       =                                 cidrsubnet(cidrsubnet(var.vpc_cidr, local.vpc_split_ranges, 0), 1, 1)
  default_private_nacl     =                                            cidrsubnet(var.vpc_cidr, local.vpc_split_ranges, 1)
  default_database_nacl    = local.vpc_cidr_size > 19 ? "" :            cidrsubnet(var.vpc_cidr, local.vpc_split_ranges, 2)
  default_elasticache_nacl = local.vpc_cidr_size > 19 ? "" : cidrsubnet(cidrsubnet(var.vpc_cidr, local.vpc_split_ranges, 3), 1, 0)
  
  default_public      = [for i in range(local.azs) : cidrsubnet(local.default_public_nacl, 2, i)]
  default_intra       = [for i in range(local.azs) : cidrsubnet(local.default_intra_nacl, 2, i)]
  default_private     = [for i in range(local.azs) : cidrsubnet(local.default_private_nacl, 2, i)]
  default_database    = local.vpc_cidr_size > 19 ? [] : [for i in range(local.azs) : cidrsubnet(local.default_database_nacl, 2, i)]
  default_elasticache = local.vpc_cidr_size > 19 ? [] : [for i in range(local.azs) : cidrsubnet(local.default_elasticache_nacl, 2, i)] 
  
  #Determine the subnets to be passed to the vpc module depending on weahter users specify the range themselves or depending on the subnet split mode.
  public_subnets      = var.public_subnets == null      ? var.subnet_split_mode == "default" ? local.default_public      : local.spoke_public : var.public_subnets
  intra_subnets       = var.intra_subnets == null       ? var.subnet_split_mode == "default" ? local.default_intra       : local.spoke_tgw :    var.intra_subnets
  private_subnets     = var.private_subnets == null     ? var.subnet_split_mode == "default" ? local.default_private     : concat(local.spoke_control, local.spoke_service, local.spoke_compute) : var.private_subnets
  database_subnets    = var.database_subnets == null    ? var.subnet_split_mode == "default" ? local.default_database    : local.spoke_database : var.database_subnets
  elasticache_subnets = var.elasticache_subnets == null ? var.subnet_split_mode == "default" ? local.default_elasticache : [] : var.elasticache_subnets
}


#https://registry.terraform.io/modules/terraform-aws-modules/vpc/aws/latest
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "6.5.0"

  name = var.prefix
  cidr = var.vpc_cidr

  secondary_cidr_blocks = var.secondary_cidr_blocks

  azs                 = [for i in range(local.azs) : data.aws_availability_zones.available.names[i]]
  private_subnets     = local.private_subnets
  public_subnets      = local.public_subnets
  database_subnets    = local.database_subnets
  elasticache_subnets = local.elasticache_subnets
  intra_subnets       = local.intra_subnets

  private_subnet_names     = var.subnet_split_mode == "default" ? var.private_subnet_names : concat(local.spoke_control_names, local.spoke_service_names, local.spoke_compute_names)
  public_subnet_names      = var.public_subnet_names
  database_subnet_names    = var.database_subnet_names
  elasticache_subnet_names = var.elasticache_subnet_names
  intra_subnet_names       = var.subnet_split_mode == "default" ? var.intra_subnet_names : local.spoke_tgw_names

  database_subnet_group_name    = var.database_subnet_group_name
  elasticache_subnet_group_name = var.elasticache_subnet_group_name

  create_database_subnet_group    = length(local.database_subnets) > 0 ? true : false
  create_elasticache_subnet_group = length(local.elasticache_subnets) > 0 ? true : false
  create_multiple_intra_route_tables = var.create_multiple_intra_route_tables
  create_multiple_public_route_tables = var.create_multiple_public_route_tables
  

  enable_nat_gateway     = var.enable_nat_gateway
  single_nat_gateway     = var.one_nat_gateway_per_az ? false : true
  one_nat_gateway_per_az = var.one_nat_gateway_per_az

  reuse_nat_ips        = false
  enable_dns_hostnames = true
  enable_dns_support   = true

  manage_default_network_acl = var.manage_default_network_acl
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
    created-by = "entigo-infralib"
  }
}

resource "aws_security_group" "pipeline_security_group" {
  name        = "${var.prefix}-pipeline"
  description = "${var.prefix} Security group used by pipelines that run terraform"
  vpc_id      = module.vpc.vpc_id
  tags = {
    Terraform  = "true"
    Prefix     = var.prefix
    created-by = "entigo-infralib"
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
