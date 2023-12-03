
data "aws_route53_zone" "parent_zone" {
  count = var.parent_zone_id != "" ? 1 : 0
  zone_id         = var.parent_zone_id
}
