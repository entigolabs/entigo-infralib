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

output "ssl_policy_restricted" {
  description = "RESTRICTED profile, minimum TLS 1.2"
  value       = google_compute_ssl_policy.this["restricted"].name
}

output "ssl_policy_modern" {
  description = "MODERN profile, minimum TLS 1.2"
  value       = google_compute_ssl_policy.this["modern"].name
}

output "ssl_policy_compatible" {
  description = "COMPATIBLE profile, minimum TLS 1.0, the google default"
  value       = google_compute_ssl_policy.this["compatible"].name
}
