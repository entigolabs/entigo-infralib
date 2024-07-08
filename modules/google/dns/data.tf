data "google_dns_managed_zone" "parent_zone" {
  count = var.parent_zone_id != "" ? 1 : 0
  name  = var.parent_zone_id
}
