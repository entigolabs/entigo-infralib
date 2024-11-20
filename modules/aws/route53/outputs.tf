output "pub_zone_id" {
  value = local.pub_zone
}

output "pub_domain" {
  value = local.pub_domain
}

output "pub_cert_arn" {
  value = var.create_cert && ( var.create_public || var.parent_zone_id != "" ) ? aws_acm_certificate.pub[0].arn : null
}

output "int_zone_id" {
  value = local.int_zone
}

output "int_domain" {
  value = local.int_domain
}

output "int_cert_arn" {
  value = var.create_cert && var.create_private ? aws_acm_certificate.int[0].arn : null
}
