data "aws_ssm_parameter" "vpc_id" {
  name = "/entigo-infralib/${var.vpc_prefix}-${terraform.workspace}/vpc/vpc_id"
}

data "aws_route53_zone" "parent_zone" {
  count = var.parent_zone_id != "" ? 1 : 0
  zone_id         = var.parent_zone_id
}
