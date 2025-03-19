
resource "aws_ec2_transit_gateway_vpc_attachment" "tgw" {
  subnet_ids         = var.attachment_subnets
  transit_gateway_id = data.aws_ec2_transit_gateway.tgw.id
  vpc_id             =  var.vpc_id
  
  dns_support = var.dns_support
  ipv6_support = var.ipv6_support
  security_group_referencing_support = var.security_group_referencing_support
  transit_gateway_default_route_table_association = var.transit_gateway_default_route_table_association
  transit_gateway_default_route_table_propagation = var.transit_gateway_default_route_table_propagation
  
  tags = {
    Name = var.prefix
    Terraform = "true"
    Prefix    = var.prefix
  }
}

locals {
  route_combinations = flatten([
    for cidr, route_tables in var.routes : [
      for rt_id in route_tables : {
        cidr   = cidr
        rt_id  = rt_id
      }
    ]
  ])
  
  route_map = {
    for combo in local.route_combinations : 
    "${combo.rt_id}-${combo.cidr}" => combo
  }
}

resource "aws_route" "tgw" {
  for_each               = local.route_map
  route_table_id         = each.value.rt_id
  destination_cidr_block = each.value.cidr
  transit_gateway_id     = data.aws_ec2_transit_gateway.tgw.id
}
