data "aws_route53_zone" "this" {
  name         = var.zone_name
  private_zone = var.private_zone
}
