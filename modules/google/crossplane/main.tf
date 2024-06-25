resource "google_service_account_iam_member" "crossplane" {
  service_account_id = google_service_account.crossplane.id
  member  = "serviceAccount:${data.google_client_config.this.project}.svc.id.goog[${var.kns_name}/${var.ksa_name}]"
  role    = "roles/editor"
}

resource "google_project_iam_member" "crossplane" {
  project = data.google_client_config.this.project
  role    = "roles/editor"
  member  = "serviceAccount:${google_service_account.crossplane.email}"
}

resource "google_service_account" "crossplane" {
  account_id   = "${local.hname}-cp"
  display_name = "${local.hname}-cp"
}