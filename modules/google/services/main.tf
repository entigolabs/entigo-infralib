locals {
  default_services = [
    "artifactregistry.googleapis.com",
    "certificatemanager.googleapis.com",
    "clouddeploy.googleapis.com",
    "cloudkms.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "cloudscheduler.googleapis.com",
    "compute.googleapis.com",
    "container.googleapis.com",
    "dns.googleapis.com",
    "file.googleapis.com",
    "memorystore.googleapis.com",
    "pubsub.googleapis.com",
    "redis.googleapis.com",
    "run.googleapis.com",
    "secretmanager.googleapis.com",
    "servicenetworking.googleapis.com",
    "sqladmin.googleapis.com",
    "storage.googleapis.com"
  ]

  services = toset(concat(
    local.default_services,
    var.additional_services,
    var.additional_services_without_identity
  ))

  # Services that don't have identity
  services_without_identity = toset(concat(
    ["cloudresourcemanager.googleapis.com"],
    var.additional_services_without_identity
  ))

  # Services that don't return email from identity resource
  predefined_service_agent_emails = {
    "compute.googleapis.com" = "service-${data.google_project.this.number}@compute-system.iam.gserviceaccount.com"
    "storage.googleapis.com" = "service-${data.google_project.this.number}@gs-project-accounts.iam.gserviceaccount.com"
  }
}

resource "google_project_service" "this" {
  for_each = local.services
  service  = each.value
}

resource "google_project_service_identity" "this" {
  for_each = setsubtract(local.services, local.services_without_identity)
  provider = google-beta
  service  = each.value

  depends_on = [google_project_service.this]
}