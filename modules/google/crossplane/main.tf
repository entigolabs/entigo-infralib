// crossplane-core service account
resource "google_service_account_iam_member" "crossplane_editor" {
  service_account_id = google_service_account.crossplane.id
  member  = "serviceAccount:${data.google_client_config.this.project}.svc.id.goog[${var.kubernetes_namespace}/${var.kubernetes_service_account_crossplane}]"
  role    = "roles/editor"
}

resource "google_service_account_iam_member" "crossplane_workload_identity_user" {
  service_account_id = google_service_account.crossplane.id
  member  = "serviceAccount:${data.google_client_config.this.project}.svc.id.goog[${var.kubernetes_namespace}/${var.kubernetes_service_account_crossplane}]"
  role    = "roles/iam.workloadIdentityUser"
}

resource "google_service_account_iam_member" "crossplane_security_admin" {
  service_account_id = google_service_account.crossplane.id
  member  = "serviceAccount:${data.google_client_config.this.project}.svc.id.goog[${var.kubernetes_namespace}/${var.kubernetes_service_account_crossplane}]"
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

module "service_account_email_crossplane" {
  source                             = "./secret"
  prefix = var.prefix
  key = "service_account_email_crossplane"
  value = google_service_account.crossplane.email
}

// crossplane-google service account
resource "google_service_account_iam_member" "crossplane_google_editor" {
  service_account_id = google_service_account.crossplane.id
  member  = "serviceAccount:${data.google_client_config.this.project}.svc.id.goog[${var.kubernetes_namespace}/${var.kubernetes_service_account_crossplane_google}]"
  role    = "roles/editor"
}

resource "google_service_account_iam_member" "crossplane_google_workload_identity_user" {
  service_account_id = google_service_account.crossplane.id
  member  = "serviceAccount:${data.google_client_config.this.project}.svc.id.goog[${var.kubernetes_namespace}/${var.kubernetes_service_account_crossplane_google}]"
  role    = "roles/iam.workloadIdentityUser"
}

resource "google_service_account_iam_member" "crossplane_google_security_admin" {
  service_account_id = google_service_account.crossplane.id
  member  = "serviceAccount:${data.google_client_config.this.project}.svc.id.goog[${var.kubernetes_namespace}/${var.kubernetes_service_account_crossplane_google}]"
  role    = "roles/iam.securityAdmin"
}

resource "google_project_iam_member" "crossplane_google_editor" {
  project = data.google_client_config.this.project
  role    = "roles/editor"
  member  = "serviceAccount:${google_service_account.crossplane_google.email}"
}

resource "google_project_iam_member" "crossplane_google_workload_identity_user" {
  project = data.google_client_config.this.project
  role    = "roles/iam.workloadIdentityUser"
  member  = "serviceAccount:${google_service_account.crossplane_google.email}"
}

resource "google_project_iam_member" "crossplane_google_security_admin" {
  project = data.google_client_config.this.project
  role    = "roles/iam.securityAdmin"
  member  = "serviceAccount:${google_service_account.crossplane_google.email}"
}

resource "google_service_account" "crossplane_google" {
  account_id   = "${substr(local.hname, 0, 18)}-cp-google"
  display_name = "${local.hname}-cp-google"
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
