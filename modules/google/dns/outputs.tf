output "pub_zone_id" {
  value = local.pub_zone
}

output "pub_domain" {
  value = trimsuffix(local.pub_domain, ".")
}

output "int_zone_id" {
  value = local.int_zone
}

output "int_domain" {
  value = trimsuffix(local.int_domain, ".")
}

output "parent_zone_id" {
  value = var.parent_zone_id != "" ? var.parent_zone_id : null
}
