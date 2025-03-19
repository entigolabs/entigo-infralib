output "tgw_id" {
  value = data.aws_ec2_transit_gateway.tgw.id
}

output "tgw_attatchment_id" {
  value = aws_ec2_transit_gateway_vpc_attachment.tgw.id
}
