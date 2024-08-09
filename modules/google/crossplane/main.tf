// crossplane-core service account
resource "google_service_account_iam_member" "crossplane_editor" {
  service_account_id = google_service_account.crossplane.id
  member  = "serviceAccount:${data.google_client_config.this.project}.svc.id.goog[${var.kubernetes_namespace}/${var.kubernetes_service_account}]"
  role    = "roles/editor"
}

resource "google_service_account_iam_member" "crossplane_workload_identity_user" {
  service_account_id = google_service_account.crossplane.id
  member  = "serviceAccount:${data.google_client_config.this.project}.svc.id.goog[${var.kubernetes_namespace}/${var.kubernetes_service_account}]"
  role    = "roles/iam.workloadIdentityUser"
}

resource "google_service_account_iam_member" "crossplane_security_admin" {
  service_account_id = google_service_account.crossplane.id
  member  = "serviceAccount:${data.google_client_config.this.project}.svc.id.goog[${var.kubernetes_namespace}/${var.kubernetes_service_account}]"
  role    = "roles/iam.securityAdmin"
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
  account_id   = "${substr(local.hname, 0, 25)}-cp"
  display_name = "${local.hname}-cp"
}

module "service_account_email_crossplane_google" {
  source                             = "./secret"
  prefix = var.prefix
  key = "service_account_email_crossplane_google"
  value = google_service_account.crossplane.email
}

module "project_id" {
  source                             = "./secret"
  prefix = var.prefix
  key = "project_id"
  value = data.google_client_config.this.project
}
