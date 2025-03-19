data "aws_ec2_transit_gateway" "tgw" {
  filter {
    name   = "state"
    values = ["available"]
  }
}
