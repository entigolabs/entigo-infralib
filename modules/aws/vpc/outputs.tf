output "vpc_id" {
  description = "The ID of the VPC"
  value = module.vpc.vpc_id
}

output "vpc_arn" {
  description = "The ARN of the VPC"
  value = module.vpc.vpc_arn
}

output "vpc_cidr_block" {
  description = "The CIDR block of the VPC"
  value = module.vpc.vpc_cidr_block
}

output "default_security_group_id" {
  description = "The ID of the security group created by default on VPC creation"
  value = module.vpc.default_security_group_id
}

output "default_network_acl_id" {
  description = "The ID of the default network ACL"
  value = module.vpc.default_network_acl_id
}

output "default_route_table_id" {
  description = "The ID of the default route table"
  value = module.vpc.default_route_table_id
}

output "vpc_instance_tenancy" {
  description = "Tenancy of instances spin up within VPC"
  value = module.vpc.vpc_instance_tenancy
}

output "vpc_enable_dns_support" {
  description = "Whether or not the VPC has DNS support"
  value = module.vpc.vpc_enable_dns_support
}

output "vpc_enable_dns_hostnames" {
  description = "Whether or not the VPC has DNS hostname support"
  value = module.vpc.vpc_enable_dns_hostnames
}

output "vpc_main_route_table_id" {
  description = "The ID of the main route table associated with this VPC"
  value = module.vpc.vpc_main_route_table_id
}

output "vpc_ipv6_association_id" {
  description = "The association ID for the IPv6 CIDR block"
  value = module.vpc.vpc_ipv6_association_id
}

output "vpc_ipv6_cidr_block" {
  description = "The IPv6 CIDR block"
  value = module.vpc.vpc_ipv6_cidr_block
}

output "vpc_secondary_cidr_blocks" {
  description = "List of secondary CIDR blocks of the VPC"
  value = module.vpc.vpc_secondary_cidr_blocks
}

output "vpc_owner_id" {
  description = "The ID of the AWS account that owns the VPC"
  value = module.vpc.vpc_owner_id
}

output "dhcp_options_id" {
  description = "The ID of the DHCP options"
  value = module.vpc.dhcp_options_id
}

output "igw_id" {
  description = "The ID of the Internet Gateway"
  value = module.vpc.igw_id
}

output "igw_arn" {
  description = "The ARN of the Internet Gateway"
  value = module.vpc.igw_arn
}

output "public_subnets" {
  description = "List of IDs of public subnets"
  value = module.vpc.public_subnets
}

output "public_subnet_arns" {
  description = "List of ARNs of public subnets"
  value = module.vpc.public_subnet_arns
}

output "public_subnets_cidr_blocks" {
  description = "List of cidr_blocks of public subnets"
  value = module.vpc.public_subnets_cidr_blocks
}

output "public_subnets_ipv6_cidr_blocks" {
  description = "List of IPv6 cidr_blocks of public subnets in an IPv6 enabled VPC"
  value = module.vpc.public_subnets_ipv6_cidr_blocks
}

output "public_route_table_ids" {
  description = "List of IDs of public route tables"
  value = module.vpc.public_route_table_ids
}

output "public_internet_gateway_route_id" {
  description = "ID of the internet gateway route"
  value = module.vpc.public_internet_gateway_route_id
}

output "public_internet_gateway_ipv6_route_id" {
  description = "ID of the IPv6 internet gateway route"
  value = module.vpc.public_internet_gateway_ipv6_route_id
}

output "public_route_table_association_ids" {
  description = "List of IDs of the public route table association"
  value = module.vpc.public_route_table_association_ids
}

output "public_network_acl_id" {
  description = "ID of the public network ACL"
  value = module.vpc.public_network_acl_id
}

output "public_network_acl_arn" {
  description = "ARN of the public network ACL"
  value = module.vpc.public_network_acl_arn
}

output "private_subnets" {
  description = "List of IDs of private subnets"
  value = module.vpc.private_subnets
}

output "private_subnet_arns" {
  description = "List of ARNs of private subnets"
  value = module.vpc.private_subnet_arns
}

output "private_subnets_cidr_blocks" {
  description = "List of cidr_blocks of private subnets"
  value = module.vpc.private_subnets_cidr_blocks
}

output "private_subnets_ipv6_cidr_blocks" {
  description = "List of IPv6 cidr_blocks of private subnets in an IPv6 enabled VPC"
  value = module.vpc.private_subnets_ipv6_cidr_blocks
}

output "private_route_table_ids" {
  description = "List of IDs of private route tables"
  value = module.vpc.private_route_table_ids
}

output "private_nat_gateway_route_ids" {
  description = "List of IDs of the private nat gateway route"
  value = module.vpc.private_nat_gateway_route_ids
}

output "private_ipv6_egress_route_ids" {
  description = "List of IDs of the ipv6 egress route"
  value = module.vpc.private_ipv6_egress_route_ids
}

output "private_route_table_association_ids" {
  description = "List of IDs of the private route table association"
  value = module.vpc.private_route_table_association_ids
}

output "private_network_acl_id" {
  description = "ID of the private network ACL"
  value = module.vpc.private_network_acl_id
}

output "private_network_acl_arn" {
  description = "ARN of the private network ACL"
  value = module.vpc.private_network_acl_arn
}

output "outpost_subnets" {
  description = "List of IDs of outpost subnets"
  value = module.vpc.outpost_subnets
}

output "outpost_subnet_arns" {
  description = "List of ARNs of outpost subnets"
  value = module.vpc.outpost_subnet_arns
}

output "outpost_subnets_cidr_blocks" {
  description = "List of cidr_blocks of outpost subnets"
  value = module.vpc.outpost_subnets_cidr_blocks
}

output "outpost_subnets_ipv6_cidr_blocks" {
  description = "List of IPv6 cidr_blocks of outpost subnets in an IPv6 enabled VPC"
  value = module.vpc.outpost_subnets_ipv6_cidr_blocks
}

output "outpost_network_acl_id" {
  description = "ID of the outpost network ACL"
  value = module.vpc.outpost_network_acl_id
}

output "outpost_network_acl_arn" {
  description = "ARN of the outpost network ACL"
  value = module.vpc.outpost_network_acl_arn
}

output "database_subnets" {
  description = "List of IDs of database subnets"
  value = module.vpc.database_subnets
}

output "database_subnet_arns" {
  description = "List of ARNs of database subnets"
  value = module.vpc.database_subnet_arns
}

output "database_subnets_cidr_blocks" {
  description = "List of cidr_blocks of database subnets"
  value = module.vpc.database_subnets_cidr_blocks
}

output "database_subnets_ipv6_cidr_blocks" {
  description = "List of IPv6 cidr_blocks of database subnets in an IPv6 enabled VPC"
  value = module.vpc.database_subnets_ipv6_cidr_blocks
}

output "database_subnet_group" {
  description = "ID of database subnet group"
  value = module.vpc.database_subnet_group
}

output "database_subnet_group_name" {
  description = "Name of database subnet group"
  value = module.vpc.database_subnet_group_name
}

output "database_route_table_ids" {
  description = "List of IDs of database route tables"
  value = module.vpc.database_route_table_ids
}

output "database_internet_gateway_route_id" {
  description = "ID of the database internet gateway route"
  value = module.vpc.database_internet_gateway_route_id
}

output "database_nat_gateway_route_ids" {
  description = "List of IDs of the database nat gateway route"
  value = module.vpc.database_nat_gateway_route_ids
}

output "database_ipv6_egress_route_id" {
  description = "ID of the database IPv6 egress route"
  value = module.vpc.database_ipv6_egress_route_id
}

output "database_route_table_association_ids" {
  description = "List of IDs of the database route table association"
  value = module.vpc.database_route_table_association_ids
}

output "database_network_acl_id" {
  description = "ID of the database network ACL"
  value = module.vpc.database_network_acl_id
}

output "database_network_acl_arn" {
  description = "ARN of the database network ACL"
  value = module.vpc.database_network_acl_arn
}

output "redshift_subnets" {
  description = "List of IDs of redshift subnets"
  value = module.vpc.redshift_subnets
}

output "redshift_subnet_arns" {
  description = "List of ARNs of redshift subnets"
  value = module.vpc.redshift_subnet_arns
}

output "redshift_subnets_cidr_blocks" {
  description = "List of cidr_blocks of redshift subnets"
  value = module.vpc.redshift_subnets_cidr_blocks
}

output "redshift_subnets_ipv6_cidr_blocks" {
  description = "List of IPv6 cidr_blocks of redshift subnets in an IPv6 enabled VPC"
  value = module.vpc.redshift_subnets_ipv6_cidr_blocks
}

output "redshift_subnet_group" {
  description = "ID of redshift subnet group"
  value = module.vpc.redshift_subnet_group
}

output "redshift_route_table_ids" {
  description = "List of IDs of redshift route tables"
  value = module.vpc.redshift_route_table_ids
}

output "redshift_route_table_association_ids" {
  description = "List of IDs of the redshift route table association"
  value = module.vpc.redshift_route_table_association_ids
}

output "redshift_public_route_table_association_ids" {
  description = "List of IDs of the public redshift route table association"
  value = module.vpc.redshift_public_route_table_association_ids
}

output "redshift_network_acl_id" {
  description = "ID of the redshift network ACL"
  value = module.vpc.redshift_network_acl_id
}

output "redshift_network_acl_arn" {
  description = "ARN of the redshift network ACL"
  value = module.vpc.redshift_network_acl_arn
}

output "elasticache_subnets" {
  description = "List of IDs of elasticache subnets"
  value = module.vpc.elasticache_subnets
}

output "elasticache_subnet_arns" {
  description = "List of ARNs of elasticache subnets"
  value = module.vpc.elasticache_subnet_arns
}

output "elasticache_subnets_cidr_blocks" {
  description = "List of cidr_blocks of elasticache subnets"
  value = module.vpc.elasticache_subnets_cidr_blocks
}

output "elasticache_subnets_ipv6_cidr_blocks" {
  description = "List of IPv6 cidr_blocks of elasticache subnets in an IPv6 enabled VPC"
  value = module.vpc.elasticache_subnets_ipv6_cidr_blocks
}

output "elasticache_subnet_group" {
  description = "ID of elasticache subnet group"
  value = module.vpc.elasticache_subnet_group
}

output "elasticache_subnet_group_name" {
  description = "Name of elasticache subnet group"
  value = module.vpc.elasticache_subnet_group_name
}

output "elasticache_route_table_ids" {
  description = "List of IDs of elasticache route tables"
  value = module.vpc.elasticache_route_table_ids
}

output "elasticache_route_table_association_ids" {
  description = "List of IDs of the elasticache route table association"
  value = module.vpc.elasticache_route_table_association_ids
}

output "elasticache_network_acl_id" {
  description = "ID of the elasticache network ACL"
  value = module.vpc.elasticache_network_acl_id
}

output "elasticache_network_acl_arn" {
  description = "ARN of the elasticache network ACL"
  value = module.vpc.elasticache_network_acl_arn
}

output "intra_subnets" {
  description = "List of IDs of intra subnets"
  value = module.vpc.intra_subnets
}

output "intra_subnet_arns" {
  description = "List of ARNs of intra subnets"
  value = module.vpc.intra_subnet_arns
}

output "intra_subnets_cidr_blocks" {
  description = "List of cidr_blocks of intra subnets"
  value = module.vpc.intra_subnets_cidr_blocks
}

output "intra_subnets_ipv6_cidr_blocks" {
  description = "List of IPv6 cidr_blocks of intra subnets in an IPv6 enabled VPC"
  value = module.vpc.intra_subnets_ipv6_cidr_blocks
}

output "intra_route_table_ids" {
  description = "List of IDs of intra route tables"
  value = module.vpc.intra_route_table_ids
}

output "intra_route_table_association_ids" {
  description = "List of IDs of the intra route table association"
  value = module.vpc.intra_route_table_association_ids
}

output "intra_network_acl_id" {
  description = "ID of the intra network ACL"
  value = module.vpc.intra_network_acl_id
}

output "intra_network_acl_arn" {
  description = "ARN of the intra network ACL"
  value = module.vpc.intra_network_acl_arn
}

output "nat_ids" {
  description = "List of allocation ID of Elastic IPs created for AWS NAT Gateway"
  value = module.vpc.nat_ids
}

output "nat_public_ips" {
  description = "List of public Elastic IPs created for AWS NAT Gateway"
  value = module.vpc.nat_public_ips
}

output "natgw_ids" {
  description = "List of NAT Gateway IDs"
  value = module.vpc.natgw_ids
}

output "natgw_interface_ids" {
  description = "List of Network Interface IDs assigned to NAT Gateways"
  value = module.vpc.natgw_interface_ids
}

output "egress_only_internet_gateway_id" {
  description = "The ID of the egress only Internet Gateway"
  value = module.vpc.egress_only_internet_gateway_id
}

output "cgw_ids" {
  description = "List of IDs of Customer Gateway"
  value = module.vpc.cgw_ids
}

output "cgw_arns" {
  description = "List of ARNs of Customer Gateway"
  value = module.vpc.cgw_arns
}

output "this_customer_gateway" {
  description = "Map of Customer Gateway attributes"
  value = module.vpc.this_customer_gateway
}

output "vgw_id" {
  description = "The ID of the VPN Gateway"
  value = module.vpc.vgw_id
}

output "vgw_arn" {
  description = "The ARN of the VPN Gateway"
  value = module.vpc.vgw_arn
}

output "default_vpc_id" {
  description = "The ID of the Default VPC"
  value = module.vpc.default_vpc_id
}

output "default_vpc_arn" {
  description = "The ARN of the Default VPC"
  value = module.vpc.default_vpc_arn
}

output "default_vpc_cidr_block" {
  description = "The CIDR block of the Default VPC"
  value = module.vpc.default_vpc_cidr_block
}

output "default_vpc_default_security_group_id" {
  description = "The ID of the security group created by default on Default VPC creation"
  value = module.vpc.default_vpc_default_security_group_id
}

output "default_vpc_default_network_acl_id" {
  description = "The ID of the default network ACL of the Default VPC"
  value = module.vpc.default_vpc_default_network_acl_id
}

output "default_vpc_default_route_table_id" {
  description = "The ID of the default route table of the Default VPC"
  value = module.vpc.default_vpc_default_route_table_id
}

output "default_vpc_instance_tenancy" {
  description = "Tenancy of instances spin up within Default VPC"
  value = module.vpc.default_vpc_instance_tenancy
}

output "default_vpc_enable_dns_support" {
  description = "Whether or not the Default VPC has DNS support"
  value = module.vpc.default_vpc_enable_dns_support
}

output "default_vpc_enable_dns_hostnames" {
  description = "Whether or not the Default VPC has DNS hostname support"
  value = module.vpc.default_vpc_enable_dns_hostnames
}

output "default_vpc_main_route_table_id" {
  description = "The ID of the main route table associated with the Default VPC"
  value = module.vpc.default_vpc_main_route_table_id
}

output "vpc_flow_log_id" {
  description = "The ID of the Flow Log resource"
  value = module.vpc.vpc_flow_log_id
}

output "vpc_flow_log_destination_arn" {
  description = "The ARN of the destination for VPC Flow Logs"
  value = module.vpc.vpc_flow_log_destination_arn
}

output "vpc_flow_log_destination_type" {
  description = "The type of the destination for VPC Flow Logs"
  value = module.vpc.vpc_flow_log_destination_type
}

output "vpc_flow_log_cloudwatch_iam_role_arn" {
  description = "The ARN of the IAM role used when pushing logs to Cloudwatch log group"
  value = module.vpc.vpc_flow_log_cloudwatch_iam_role_arn
}

output "vpc_flow_log_deliver_cross_account_role" {
  description = "The ARN of the IAM role used when pushing logs cross account"
  value = module.vpc.vpc_flow_log_deliver_cross_account_role
}

output "azs" {
  description = "A list of availability zones specified as argument to this module"
  value = module.vpc.azs
}

output "name" {
  description = "The name of the VPC specified as argument to this module"
  value = module.vpc.name
}
# Additional outputs

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

output "private_subnet_names" {
  value = var.private_subnet_names
}

output "public_subnet_names" {
  value = var.public_subnet_names
}

output "database_subnet_names" {
  value = var.database_subnet_names
}

output "elasticache_subnet_names" {
  value = var.elasticache_subnet_names
}

output "intra_subnet_names" {
  value = var.intra_subnet_names
}

output "pipeline_security_group" {
  value = aws_security_group.pipeline_security_group.id
}

#Ouputs for subnet_split_mode (default and spoke)

output "control_subnets" {
  description = "List of IDs of control subnets"
  value = var.subnet_split_mode == "default" ? module.vpc.private_subnets : [for i in range(local.azs) : module.vpc.private_subnets[i]]
}

output "service_subnets" {
  description = "List of IDs of service subnets"
  value = var.subnet_split_mode == "default" ? module.vpc.private_subnets : [for i in range(local.azs) : module.vpc.private_subnets[i+local.azs]]
}

output "compute_subnets" {
  description = "List of IDs of compute subnets"
  value = var.subnet_split_mode == "default" ? module.vpc.private_subnets : [for i in range(local.azs) : module.vpc.private_subnets[i+(2*local.azs)]]
}

output "control_subnets_cidr_blocks" {
  description = "List of IDs of control subnets"
  value = var.subnet_split_mode == "default" ? module.vpc.private_subnets_cidr_blocks : [for i in range(local.azs) : module.vpc.private_subnets_cidr_blocks[i]]
}

output "service_subnets_cidr_blocks" {
  description = "List of IDs of service subnets"
  value = var.subnet_split_mode == "default" ? module.vpc.private_subnets_cidr_blocks : [for i in range(local.azs) : module.vpc.private_subnets_cidr_blocks[i+local.azs]]
}

output "compute_subnets_cidr_blocks" {
  description = "List of IDs of compute subnets"
  value = var.subnet_split_mode == "default" ? module.vpc.private_subnets_cidr_blocks : [for i in range(local.azs) : module.vpc.private_subnets_cidr_blocks[i+(2*local.azs)]]
}

#Output to be used for Network ACL or SG rules covering all zones of a specifi type of subnet with one CIDR. Does not work when not using automatic subnet calculation.
output "acl_control_subnets_cidr_block" {
  description = "List of IDs of control subnets"
  value = var.subnet_split_mode == "default" ? local.default_private_nacl : local.spoke_control_nacl
}

output "acl_service_subnets_cidr_block" {
  description = "List of IDs of service subnets"
  value = var.subnet_split_mode == "default" ? local.default_private_nacl : local.spoke_service_nacl
}

output "acl_compute_subnets_cidr_block" {
  description = "List of IDs of compute subnets"
  value = var.subnet_split_mode == "default" ? local.default_private_nacl : local.spoke_compute_nacl
}

output "acl_intra_subnets_cidr_block" {
  description = "List of IDs of compute subnets"
  value = var.subnet_split_mode == "default" ? local.default_intra_nacl : local.spoke_tgw_nacl
}

output "acl_database_subnets_cidr_block" {
  description = "List of IDs of compute subnets"
  value = var.subnet_split_mode == "default" ? local.default_database_nacl : local.spoke_database_nacl
}

output "acl_public_subnets_cidr_block" {
  description = "List of IDs of compute subnets"
  value = var.subnet_split_mode == "default" ? local.default_public_nacl : local.spoke_public_nacl
}

#Zone based outputs
output "zoned_private_subnets" {
  description = "List of IDs of private subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_id in module.vpc.private_subnets :
      subnet_id
      if subnet_index % local.azs == az_index
    ]
  }
}

output "zoned_private_subnets_cidr_blocks" {
  description = "List of IDs of private subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_cidr in module.vpc.private_subnets_cidr_blocks :
      subnet_cidr
      if subnet_index % local.azs == az_index
    ]
  }
}

output "zoned_public_subnets" {
  description = "List of IDs of public subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_id in module.vpc.public_subnets :
      subnet_id
      if subnet_index % local.azs == az_index
    ]
  }
}

output "zoned_public_subnets_cidr_blocks" {
  description = "List of IDs of public subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_cidr in module.vpc.public_subnets_cidr_blocks :
      subnet_cidr
      if subnet_index % local.azs == az_index
    ]
  }
}

output "zoned_intra_subnets" {
  description = "List of IDs of intra subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_id in module.vpc.intra_subnets :
      subnet_id
      if subnet_index % local.azs == az_index
    ]
  }
}

output "zoned_intra_subnets_cidr_blocks" {
  description = "List of IDs of intra subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_cidr in module.vpc.intra_subnets_cidr_blocks :
      subnet_cidr
      if subnet_index % local.azs == az_index
    ]
  }
}

output "zoned_elasticache_subnets" {
  description = "List of IDs of elasticache subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_id in module.vpc.elasticache_subnets :
      subnet_id
      if subnet_index % local.azs == az_index
    ]
  }
}

output "zoned_elasticache_subnets_cidr_blocks" {
  description = "List of IDs of elasticache subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_cidr in module.vpc.elasticache_subnets_cidr_blocks :
      subnet_cidr
      if subnet_index % local.azs == az_index
    ]
  }
}

output "zoned_database_subnets" {
  description = "List of IDs of database subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_id in module.vpc.database_subnets :
      subnet_id
      if subnet_index % local.azs == az_index
    ]
  }
}

output "zoned_database_subnets_cidr_blocks" {
  description = "List of IDs of database subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_cidr in module.vpc.database_subnets_cidr_blocks :
      subnet_cidr
      if subnet_index % local.azs == az_index
    ]
  }
}

output "zoned_control_subnets" {
  description = "List of IDs of control subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_id in var.subnet_split_mode == "default" ? module.vpc.private_subnets : [for i in range(local.azs) : module.vpc.private_subnets[i]] :
      subnet_id
      if subnet_index % local.azs == az_index
    ]
  }
}

output "zoned_control_subnets_cidr_blocks" {
  description = "List of IDs of control subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_cidr in var.subnet_split_mode == "default" ? module.vpc.private_subnets_cidr_blocks : [for i in range(local.azs) : module.vpc.private_subnets_cidr_blocks[i]] :
      subnet_cidr
      if subnet_index % local.azs == az_index
    ]
  }
}

output "zoned_service_subnets" {
  description = "List of IDs of service subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_id in var.subnet_split_mode == "default" ? module.vpc.private_subnets : [for i in range(local.azs) : module.vpc.private_subnets[i+local.azs]] :
      subnet_id
      if subnet_index % local.azs == az_index
    ]
  }
}

output "zoned_service_subnets_cidr_blocks" {
  description = "List of IDs of service subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_cidr in var.subnet_split_mode == "default" ? module.vpc.private_subnets_cidr_blocks : [for i in range(local.azs) : module.vpc.private_subnets_cidr_blocks[i+local.azs]] :
      subnet_cidr
      if subnet_index % local.azs == az_index
    ]
  }
}

output "zoned_compute_subnets" {
  description = "List of IDs of compute subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_id in var.subnet_split_mode == "default" ? module.vpc.private_subnets : [for i in range(local.azs) : module.vpc.private_subnets[i+(2*local.azs)]] :
      subnet_id
      if subnet_index % local.azs == az_index
    ]
  }
}

output "zoned_compute_subnets_cidr_blocks" {
  description = "List of IDs of compute subnets by zone"
  value = {
    for az_index in range(local.azs) : 
    substr(data.aws_availability_zones.available.names[az_index], -2, 2) => [
      for subnet_index, subnet_cidr in var.subnet_split_mode == "default" ? module.vpc.private_subnets_cidr_blocks : [for i in range(local.azs) : module.vpc.private_subnets_cidr_blocks[i+(2*local.azs)]] :
      subnet_cidr
      if subnet_index % local.azs == az_index
    ]
  }
}
