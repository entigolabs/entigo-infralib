module "vpc_endpoints" {
  count = var.create_endpoint_ecr || var.create_gateway_s3 || var.create_endpoint_s3 ? 1 : 0
  source  = "terraform-aws-modules/vpc/aws//modules/vpc-endpoints"
  version = "5.21.0"
  
  vpc_id = module.vpc.vpc_id
  subnet_ids = var.subnet_split_mode == "default" ? module.vpc.private_subnets : [for i in range(local.azs) : module.vpc.private_subnets[i+(2*local.azs)]]
  create_security_group      = true
  security_group_name_prefix = "${var.prefix}-endpoint"
  security_group_description = "${var.prefix} VPC endpoint SG"

  endpoints = merge(var.create_gateway_s3 ? {
      s3 = {
        service             = "s3"
        service_type    = "Gateway"
        route_table_ids = module.vpc.private_route_table_ids
        tags                = { Name = "${var.prefix}-s3" }
      }
    } : {} ,var.create_endpoint_s3 ? {
      s3e = {
        service             = "s3"
        private_dns_enabled = true
        tags                = { Name = "${var.prefix}-s3e" }
      }
    } : {} , var.create_endpoint_ecr ? {
      ecr_api = {
        service             = "ecr.api"
        private_dns_enabled = true
        tags                = { Name = "${var.prefix}-ecr.api-vpc-endpoint" }
      },
      ecr_dkr = {
        service             = "ecr.dkr"
        private_dns_enabled = true
        tags                = { Name = "${var.prefix}-ecr.dkr-vpc-endpoint" }
      }
    } : {})

  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    created-by = "entigo-infralib"
  }
}
