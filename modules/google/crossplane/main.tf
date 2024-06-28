resource "google_service_account_iam_member" "crossplane_editor" {
  service_account_id = google_service_account.crossplane.id
  member  = "serviceAccount:${data.google_client_config.this.project}.svc.id.goog[${var.kns_name}/${var.ksa_name}]"
  role    = "roles/editor"
}

resource "google_service_account_iam_member" "crossplane_workload_identity_user" {
  service_account_id = google_service_account.crossplane.id
  member  = "serviceAccount:${data.google_client_config.this.project}.svc.id.goog[${var.kns_name}/${var.ksa_name}]"
  role    = "roles/iam.workloadIdentityUser"
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

resource "google_service_account" "crossplane" {
  account_id   = "${substr(local.hname, 0, 25)}-cp"
  display_name = "${local.hname}-cp"
}