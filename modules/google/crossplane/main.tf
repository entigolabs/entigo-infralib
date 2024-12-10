// crossplane-core service account
locals {
  crossplane_service_account_id = var.crossplane_service_account_id != "" ? substr(var.crossplane_service_account_id, 0, 28) : substr(var.prefix, 0, 28)
}

resource "google_service_account_iam_member" "crossplane_workload_identity_user" {
  service_account_id = google_service_account.crossplane.id
  member             = "serviceAccount:${data.google_client_config.this.project}.svc.id.goog[${var.kubernetes_namespace}/${var.kubernetes_service_account}]"
  role               = "roles/iam.workloadIdentityUser"
}

resource "google_project_iam_member" "crossplane_editor" {
  project = data.google_client_config.this.project
  role    = "roles/editor"
  member  = "serviceAccount:${google_service_account.crossplane.email}"
}

resource "google_project_iam_member" "crossplane_workload_identity_user" {
  project = data.google_client_config.this.project
  role    = "roles/iam.workloadIdentityUser"
  member  = "serviceAccount:${google_service_account.crossplane.email}"
}

resource "google_project_iam_member" "crossplane_security_admin" {
  project = data.google_client_config.this.project
  role    = "roles/iam.securityAdmin"
  member  = "serviceAccount:${google_service_account.crossplane.email}"
}

resource "google_service_account" "crossplane" {
  account_id   = local.crossplane_service_account_id
  display_name = "Crossplane service account"
}

module "service_account_email" {
  source = "./secret"
  prefix = var.prefix
  key    = "service_account_email"
  value  = google_service_account.crossplane.email
}

module "project_id" {
  source = "./secret"
  prefix = var.prefix
  key    = "project_id"
  value  = data.google_client_config.this.project
}
