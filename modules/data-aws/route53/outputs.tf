output "pub_zone_id" {
  value = data.aws_route53_zone.this.zone_id
}

output "pub_domain" {
  value = trimsuffix(data.aws_route53_zone.this.name, ".")
}

output "pub_cert_arn" {
  value = null
}

output "int_zone_id" {
  value = data.aws_route53_zone.this.zone_id
}

output "int_domain" {
  value = trimsuffix(data.aws_route53_zone.this.name, ".")
}

output "int_cert_arn" {
  value = null
}
